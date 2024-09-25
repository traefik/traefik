package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

// SpiffeX509Source allows to retrieve a x509 SVID and bundle.
type SpiffeX509Source interface {
	x509svid.Source
	x509bundle.Source
}

// TransportManager handles transports for backend communication.
type TransportManager struct {
	rtLock        sync.RWMutex
	roundTrippers map[string]http.RoundTripper
	configs       map[string]*dynamic.ServersTransport
	tlsConfigs    map[string]*tls.Config

	spiffeX509Source SpiffeX509Source
}

// NewTransportManager creates a new TransportManager.
func NewTransportManager(spiffeX509Source SpiffeX509Source) *TransportManager {
	return &TransportManager{
		roundTrippers:    make(map[string]http.RoundTripper),
		configs:          make(map[string]*dynamic.ServersTransport),
		tlsConfigs:       make(map[string]*tls.Config),
		spiffeX509Source: spiffeX509Source,
	}
}

// Update updates the transport configurations.
func (t *TransportManager) Update(newConfigs map[string]*dynamic.ServersTransport) {
	t.rtLock.Lock()
	defer t.rtLock.Unlock()

	for configName, config := range t.configs {
		newConfig, ok := newConfigs[configName]
		if !ok {
			delete(t.configs, configName)
			delete(t.roundTrippers, configName)
			delete(t.tlsConfigs, configName)
			continue
		}

		if reflect.DeepEqual(newConfig, config) {
			continue
		}

		var err error

		var tlsConfig *tls.Config
		if tlsConfig, err = t.createTLSConfig(newConfig); err != nil {
			log.Error().Err(err).Msgf("Could not configure HTTP Transport %s TLS configuration, fallback on default TLS config", configName)
		}
		t.tlsConfigs[configName] = tlsConfig

		t.roundTrippers[configName], err = t.createRoundTripper(newConfig, tlsConfig)
		if err != nil {
			log.Error().Err(err).Msgf("Could not configure HTTP Transport %s, fallback on default transport", configName)
			t.roundTrippers[configName] = http.DefaultTransport
		}
	}

	for newConfigName, newConfig := range newConfigs {
		if _, ok := t.configs[newConfigName]; ok {
			continue
		}

		var err error

		var tlsConfig *tls.Config
		if tlsConfig, err = t.createTLSConfig(newConfig); err != nil {
			log.Error().Err(err).Msgf("Could not configure HTTP Transport %s TLS configuration, fallback on default TLS config", newConfigName)
		}
		t.tlsConfigs[newConfigName] = tlsConfig

		t.roundTrippers[newConfigName], err = t.createRoundTripper(newConfig, tlsConfig)
		if err != nil {
			log.Error().Err(err).Msgf("Could not configure HTTP Transport %s, fallback on default transport", newConfigName)
			t.roundTrippers[newConfigName] = http.DefaultTransport
		}
	}

	t.configs = newConfigs
}

// GetRoundTripper gets a roundtripper corresponding to the given transport name.
func (t *TransportManager) GetRoundTripper(name string) (http.RoundTripper, error) {
	if len(name) == 0 {
		name = "default@internal"
	}

	t.rtLock.RLock()
	defer t.rtLock.RUnlock()

	if rt, ok := t.roundTrippers[name]; ok {
		return rt, nil
	}

	return nil, fmt.Errorf("servers transport not found %s", name)
}

// Get gets transport by name.
func (t *TransportManager) Get(name string) (*dynamic.ServersTransport, error) {
	if len(name) == 0 {
		name = "default@internal"
	}

	t.rtLock.RLock()
	defer t.rtLock.RUnlock()

	if rt, ok := t.configs[name]; ok {
		return rt, nil
	}

	return nil, fmt.Errorf("servers transport not found %s", name)
}

// GetTLSConfig gets a TLS config  corresponding to the given transport name.
func (t *TransportManager) GetTLSConfig(name string) (*tls.Config, error) {
	if len(name) == 0 {
		name = "default@internal"
	}

	t.rtLock.RLock()
	defer t.rtLock.RUnlock()

	if rt, ok := t.tlsConfigs[name]; ok {
		return rt, nil
	}

	return nil, fmt.Errorf("tls config not found %s", name)
}

func (t *TransportManager) createTLSConfig(cfg *dynamic.ServersTransport) (*tls.Config, error) {
	var config *tls.Config
	if cfg.Spiffe != nil {
		if t.spiffeX509Source == nil {
			return nil, errors.New("SPIFFE is enabled for this transport, but not configured")
		}

		spiffeAuthorizer, err := buildSpiffeAuthorizer(cfg.Spiffe)
		if err != nil {
			return nil, fmt.Errorf("unable to build SPIFFE authorizer: %w", err)
		}

		config = tlsconfig.MTLSClientConfig(t.spiffeX509Source, t.spiffeX509Source, spiffeAuthorizer)
	}

	if cfg.InsecureSkipVerify || len(cfg.RootCAs) > 0 || len(cfg.ServerName) > 0 || len(cfg.Certificates) > 0 || cfg.PeerCertURI != "" {
		if config != nil {
			return nil, errors.New("TLS and SPIFFE configuration cannot be defined at the same time")
		}

		config = &tls.Config{
			ServerName:         cfg.ServerName,
			InsecureSkipVerify: cfg.InsecureSkipVerify,
			RootCAs:            createRootCACertPool(cfg.RootCAs),
			Certificates:       cfg.Certificates.GetCertificates(),
		}

		if cfg.PeerCertURI != "" {
			config.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
				return traefiktls.VerifyPeerCertificate(cfg.PeerCertURI, config, rawCerts)
			}
		}
	}

	return config, nil
}

// createRoundTripper creates an http.RoundTripper configured with the Transport configuration settings.
// For the settings that can't be configured in Traefik it uses the default http.Transport settings.
// An exception to this is the MaxIdleConns setting as we only provide the option MaxIdleConnsPerHost in Traefik at this point in time.
// Setting this value to the default of 100 could lead to confusing behavior and backwards compatibility issues.
func (t *TransportManager) createRoundTripper(cfg *dynamic.ServersTransport, tlsConfig *tls.Config) (http.RoundTripper, error) {
	if cfg == nil {
		return nil, errors.New("no transport configuration given")
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	if cfg.ForwardingTimeouts != nil {
		dialer.Timeout = time.Duration(cfg.ForwardingTimeouts.DialTimeout)
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ReadBufferSize:        64 * 1024,
		WriteBufferSize:       64 * 1024,
		TLSClientConfig:       tlsConfig,
	}

	if cfg.ForwardingTimeouts != nil {
		transport.ResponseHeaderTimeout = time.Duration(cfg.ForwardingTimeouts.ResponseHeaderTimeout)
		transport.IdleConnTimeout = time.Duration(cfg.ForwardingTimeouts.IdleConnTimeout)
	}

	// Return directly HTTP/1.1 transport when HTTP/2 is disabled
	if cfg.DisableHTTP2 {
		return &kerberosRoundTripper{
			OriginalRoundTripper: transport,
			new: func() http.RoundTripper {
				return transport.Clone()
			},
		}, nil
	}

	rt, err := newSmartRoundTripper(transport, cfg.ForwardingTimeouts)
	if err != nil {
		return nil, err
	}
	return &kerberosRoundTripper{
		OriginalRoundTripper: rt,
		new: func() http.RoundTripper {
			return rt.Clone()
		},
	}, nil
}

type stickyRoundTripper struct {
	RoundTripper http.RoundTripper
}

type transportKeyType string

var transportKey transportKeyType = "transport"

func AddTransportOnContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, transportKey, &stickyRoundTripper{})
}

type kerberosRoundTripper struct {
	new                  func() http.RoundTripper
	OriginalRoundTripper http.RoundTripper
}

func (k *kerberosRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	value, ok := request.Context().Value(transportKey).(*stickyRoundTripper)
	if !ok {
		return k.OriginalRoundTripper.RoundTrip(request)
	}

	if value.RoundTripper != nil {
		return value.RoundTripper.RoundTrip(request)
	}

	resp, err := k.OriginalRoundTripper.RoundTrip(request)

	// If we found that we are authenticating with Kerberos (Negotiate) or NTLM.
	// We put a dedicated roundTripper in the ConnContext.
	// This will stick the next calls to the same connection with the backend.
	if err == nil && containsNTLMorNegotiate(resp.Header.Values("WWW-Authenticate")) {
		value.RoundTripper = k.new()
	}
	return resp, err
}

func containsNTLMorNegotiate(h []string) bool {
	return slices.ContainsFunc(h, func(s string) bool {
		return strings.HasPrefix(s, "NTLM") || strings.HasPrefix(s, "Negotiate")
	})
}

func createRootCACertPool(rootCAs []types.FileOrContent) *x509.CertPool {
	if len(rootCAs) == 0 {
		return nil
	}

	roots := x509.NewCertPool()

	for _, cert := range rootCAs {
		certContent, err := cert.Read()
		if err != nil {
			log.Error().Err(err).Msg("Error while read RootCAs")
			continue
		}
		roots.AppendCertsFromPEM(certContent)
	}

	return roots
}

func buildSpiffeAuthorizer(cfg *dynamic.Spiffe) (tlsconfig.Authorizer, error) {
	switch {
	case len(cfg.IDs) > 0:
		spiffeIDs := make([]spiffeid.ID, 0, len(cfg.IDs))
		for _, rawID := range cfg.IDs {
			id, err := spiffeid.FromString(rawID)
			if err != nil {
				return nil, fmt.Errorf("invalid SPIFFE ID: %w", err)
			}

			spiffeIDs = append(spiffeIDs, id)
		}

		return tlsconfig.AuthorizeOneOf(spiffeIDs...), nil

	case cfg.TrustDomain != "":
		trustDomain, err := spiffeid.TrustDomainFromString(cfg.TrustDomain)
		if err != nil {
			return nil, fmt.Errorf("invalid SPIFFE trust domain: %w", err)
		}

		return tlsconfig.AuthorizeMemberOf(trustDomain), nil

	default:
		return tlsconfig.AuthorizeAny(), nil
	}
}
