package tls

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/safe"
)

const (
	defaultCRLReloadInterval = time.Hour
	defaultHTTPTimeout       = 30 * time.Second
	defaultMaxCRLBytes       = 1024 * 1024 // 1MB
	defaultCRLErrorBackoff   = 30 * time.Second
)

// crlOptionsState bundles everything tied to the lifetime of a single TLSOptions'
// CRL configuration. The store and clientHolder are long-lived: they survive dynamic
// configuration reloads as long as the TLSOptions itself still exists, regardless of
// how its CRL configuration changes. Only enforcer is rebuilt when the policy changes.
type crlOptionsState struct {
	store        *CRLStore
	clientHolder *crlHTTPClientHolder

	// clientHash fingerprints the subset of configuration affecting clientHolder
	// (currently: Timeout). Changing it only triggers an atomic client swap.
	clientHash string

	// policyHash fingerprints the full CRL configuration. Changing it triggers a cheap
	// enforcer rebuild (no network calls, no cache loss, since store is reused as-is).
	policyHash string
	enforcer   CRLEnforcer
}

// CRLManagerConfig holds the static-level settings required to build a CRLManager.
type CRLManagerConfig struct {
	// Transport is used to build per-TLSOptions HTTP clients for CRL fetching. If nil, a
	// default transport with reasonable pooling settings is used.
	Transport *http.Transport

	// DefaultHTTPTimeout is used for TLSOptions that do not explicitly configure
	// ClientAuth.Expiry.CRL.HTTP.Timeout.
	DefaultHTTPTimeout time.Duration

	// Default maximum CRL size exprimed in bytes when downloading them via HTTP calls
	DefaultMaxCRLBytes int64

	// FileCRLs optionally maps CRL distribution point URLs to local file paths, for a
	// globally shared, file based CRL store (e.g. offline/air-gapped CRLs). This is
	// expected to come from Traefik's static configuration, not from per-router dynamic
	// configuration.
	FileCRLs map[string]string

	// ReloadInterval is used both for the global file based store watcher and as the
	// default refresh interval for HTTP sourced CRLs.
	ReloadInterval time.Duration

	// DefaultErrorBackoff bounds retry frequency for a distribution point after a
	// failed resolution/download, used when ClientAuth.Expiry.CRL.HTTP.ErrorBackoff
	// is not explicitly configured.
	DefaultErrorBackoff time.Duration

	// routines pool for watchers
	SafePool *safe.Pool
}

// CRLManager builds and caches CRLEnforcer instances, one per TLS Options configuration.
//
// It is meant to be instantiated once at Traefik startup and kept alive for the whole
// process lifetime, so that CRL caches (in particular HTTP fetched CRLs) survive dynamic
// configuration reloads instead of being rebuilt from scratch on every reload.
type CRLManager struct {
	mu     sync.RWMutex
	states map[string]*crlOptionsState // keyed by TLS Options name

	transport           *http.Transport
	defaultHTTPTimeout  time.Duration
	defaultMaxCRLBytes  int64
	reloadInterval      time.Duration
	defaultErrorBackoff time.Duration

	// global is a file based CRL store shared across all TLS Options.
	global *CRLStore

	// routines pool for watchers
	safePool *safe.Pool
}

// NewCRLManager creates a CRLManager and, if any file based CRLs are configured, the
// associated global CRLStore along with its file watcher.
func NewCRLManager(ctx context.Context, cfg CRLManagerConfig) (*CRLManager, error) {
	transport := cfg.Transport
	if transport == nil {
		transport = http.DefaultTransport.(*http.Transport).Clone()
	}

	reloadInterval := cfg.ReloadInterval
	if reloadInterval <= 0 {
		reloadInterval = defaultCRLReloadInterval
	}

	defaultErrorBackoff := cfg.DefaultErrorBackoff
	if defaultErrorBackoff <= 0 {
		defaultErrorBackoff = defaultCRLErrorBackoff
	}

	httpTimeout := cfg.DefaultHTTPTimeout
	if httpTimeout <= 0 {
		httpTimeout = defaultHTTPTimeout
	}

	maxCRLBytes := cfg.DefaultMaxCRLBytes
	if maxCRLBytes <= 0 {
		maxCRLBytes = defaultMaxCRLBytes
	}

	pool := cfg.SafePool
	if pool == nil {
		pool = safe.NewPool(ctx)
	}

	m := &CRLManager{
		states:              make(map[string]*crlOptionsState),
		transport:           transport,
		defaultHTTPTimeout:  httpTimeout,
		defaultMaxCRLBytes:  maxCRLBytes,
		reloadInterval:      reloadInterval,
		defaultErrorBackoff: defaultErrorBackoff,
		safePool:            pool,
	}

	if len(cfg.FileCRLs) > 0 {
		store, err := NewFileCRLStore(cfg.FileCRLs, reloadInterval, defaultErrorBackoff)
		if err != nil {
			return nil, fmt.Errorf("initializing global file based CRL store: %w", err)
		}
		m.global = store

		m.safePool.GoCtx(store.WatchEntries)

		if fileWatcher, err := store.WatchFiles(ctx, cfg.FileCRLs); err != nil {
			return nil, fmt.Errorf("watching global CRL files: %w", err)
		} else {
			m.safePool.GoCtx(fileWatcher)
		}
	} else {
		// Empty but non-nil store, so enforcers can unconditionally query it.
		m.global = &CRLStore{crlReloadInterval: reloadInterval, crlErrorBackoff: defaultErrorBackoff}
	}

	return m, nil
}

// GetEnforcer returns the CRLEnforcer associated with a TLS Options name, building (or
// rebuilding, if the configuration changed) it as needed.
//
// optionsName should uniquely identify the TLS Options block (its name in the dynamic
// configuration, which is unique process-wide for Traefik TLS Options).
func (m *CRLManager) GetEnforcer(optionsName string, cfg CRL) (CRLEnforcer, error) {
	if cfg.Mode == "" || cfg.Mode == CRLNOOP {
		return &crlEnforcerNOOP{}, nil
	}

	clientHash := hashClientConfig(cfg.HTTP)
	policyHash, err := hashCRLConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("hashing CRL configuration for %q: %w", optionsName, err)
	}

	// Fast path: configuration unchanged. Avoids taking the write lock on every call,
	// which matters if GetEnforcer ends up being called on the hot path (e.g. once per
	// TLS handshake rather than once per dynamic configuration reload).
	m.mu.RLock()
	if state, ok := m.states[optionsName]; ok && state.policyHash == policyHash && state.clientHash == clientHash {
		enforcer := state.enforcer
		m.mu.RUnlock()
		return enforcer, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	state, ok := m.states[optionsName]
	if !ok {
		timeout := time.Duration(cfg.HTTP.Timeout)
		if timeout <= 0 {
			timeout = m.defaultHTTPTimeout
		}
		maxCRLBytes := cfg.HTTP.MaxCRLBytes
		if cfg.HTTP.MaxCRLBytes <= 0 {
			maxCRLBytes = m.defaultMaxCRLBytes
		}
		errorBackoff := time.Duration(cfg.HTTP.ErrorBackoff)
		if errorBackoff <= 0 {
			errorBackoff = m.defaultErrorBackoff
		}
		state = &crlOptionsState{
			store:        &CRLStore{crlReloadInterval: time.Duration(cfg.ReloadInterval), crlErrorBackoff: errorBackoff},
			clientHolder: newCRLHTTPClientHolder(m.transport, timeout, maxCRLBytes),
		}
		m.states[optionsName] = state

		log.Debug().Str("tlsOptions", optionsName).Msg("initializing CRL store")
	}

	if state.clientHash != clientHash {
		timeout := time.Duration(cfg.HTTP.Timeout)
		if timeout <= 0 {
			timeout = m.defaultHTTPTimeout
		}
		maxCRLBytes := cfg.HTTP.MaxCRLBytes
		if cfg.HTTP.MaxCRLBytes <= 0 {
			maxCRLBytes = m.defaultMaxCRLBytes
		}
		state.clientHolder.set(m.transport, timeout, maxCRLBytes)
		state.clientHash = clientHash

		log.Debug().Str("tlsOptions", optionsName).Msg("CRL HTTP client settings changed, updated in place (cache preserved)")
		// will update remaining config options in next block if needed
	}

	if state.enforcer == nil || state.policyHash != policyHash {
		snapProvider, err := m.snapshotProvider(cfg.HTTP, state.clientHolder)
		if err != nil {
			return nil, fmt.Errorf("building CRL snapshot provider for %q: %w", optionsName, err)
		}

		state.enforcer = buildEnforcerFromStore(cfg, state.store, m.global, snapProvider)
		state.policyHash = policyHash

		log.Debug().Str("tlsOptions", optionsName).Msg("CRL policy changed, rebuilt enforcer (cache preserved)")
	}

	return state.enforcer, nil
}

// Prune removes CRLManager state for TLSOptions that no longer exist in the dynamic
// configuration, releasing their associated CRL caches. It should be called after each
// successful dynamic configuration reload.
func (m *CRLManager) Prune(activeOptionsNames map[string]Options) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name := range m.states {
		if _, ok := activeOptionsNames[name]; !ok {
			delete(m.states, name)
			log.Debug().Str("tlsOptions", name).Msg("removed CRL state for removed TLS options")
		}
	}
}

// snapshotProvider builds the crlSnapshotProvider matching the configured expiration strategy.
func (m *CRLManager) snapshotProvider(cfg CRLHTTP, holder *crlHTTPClientHolder) (crlSnapshotProvider, error) {
	switch cfg.ExpirationStrategy {
	case "", CRLExpirationOpen:
		return &openSnaphotProvider{clientHolder: holder}, nil
	case CRLExpirationFailedClosed:
		return &failedCloseSnapshotProvider{clientHolder: holder}, nil
	default:
		return nil, fmt.Errorf("unknown CRL expiration strategy: %q", cfg.ExpirationStrategy)
	}
}

// buildEnforcerFromStore instantiates the CRLEnforcer implementation matching cfg.Mode,
// wiring it to an existing store/snapshot provider rather than owning its own.
func buildEnforcerFromStore(cfg CRL, store, global *CRLStore, snapProvider crlSnapshotProvider) CRLEnforcer {
	switch cfg.Mode {
	case CRLLax:
		return &crlEnforcer{
			strict:                       false,
			whitelistEnabled:             cfg.HTTP.Whitelist.Enabled,
			AllowedCRLDistributionPoints: cfg.HTTP.Whitelist.DistributionPoints,
			store:                        store,
			globalStore:                  global,
			snaphotProvider:              snapProvider,
		}
	case CRLStrict:
		return &crlEnforcer{
			strict:                       true,
			whitelistEnabled:             cfg.HTTP.Whitelist.Enabled,
			AllowedCRLDistributionPoints: cfg.HTTP.Whitelist.DistributionPoints,
			store:                        store,
			globalStore:                  global,
			snaphotProvider:              snapProvider,
		}
	default:
		return &crlEnforcerNOOP{}
	}
}

// hashCRLConfig produces a stable fingerprint of a CRL configuration block, used to
// detect changes across dynamic configuration reloads.
func hashCRLConfig(cfg CRL) (string, error) {
	data, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// hashClientConfig fingerprints only the subset of configuration affecting the shared
// HTTP client (currently: Timeout), used to detect changes requiring only an atomic
// client swap, with no enforcer rebuild and no cache loss.
func hashClientConfig(cfg CRLHTTP) string {
	data, _ := json.Marshal(struct {
		Timeout    ptypes.Duration
		MaxCrlSize int64
	}{cfg.Timeout, cfg.MaxCRLBytes})
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
