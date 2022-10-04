package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
)

// SpiffeX509Source allows to retrieve a x509 SVID and bundle.
type SpiffeX509Source interface {
	x509svid.Source
	x509bundle.Source
}

// TLSConfigManager is the manager of TLS client configuration.
// The TLS client configuration is used when Traefik is forwarding requests to the backends.
type TLSConfigManager struct {
	spiffeX509Source SpiffeX509Source

	configsLock sync.RWMutex
	configs     map[string]*dynamic.ServersTransport
}

// NewTLSConfigManager returns a new TLSConfigManager.
func NewTLSConfigManager(spiffeX509Source SpiffeX509Source) *TLSConfigManager {
	return &TLSConfigManager{
		configs:          make(map[string]*dynamic.ServersTransport),
		spiffeX509Source: spiffeX509Source,
	}
}

// Update is the handler called when the dynamic configuration is updated.
func (t *TLSConfigManager) Update(configs map[string]*dynamic.ServersTransport) {
	t.configsLock.Lock()
	defer t.configsLock.Unlock()

	t.configs = configs
}

// GetTLSConfig returns the client TLS configuration corresponding to the given ServersTransport name.
// When name is empty, it defaults to "default".
func (t *TLSConfigManager) GetTLSConfig(name string) (*tls.Config, error) {
	if len(name) == 0 {
		name = "default"
	}

	t.configsLock.RLock()
	defer t.configsLock.RUnlock()

	if config, ok := t.configs[name]; ok {
		return t.createTLSConfig(config.TLS)
	}

	return nil, fmt.Errorf("unable to find client TLS configuration: %s", name)
}

// createTLSConfig returns a new tls.Config corresponding to the given TLSClientConfig.
func (t *TLSConfigManager) createTLSConfig(cfg *dynamic.TLSClientConfig) (*tls.Config, error) {
	if cfg == nil {
		return nil, nil
	}

	var tlsConfig *tls.Config

	if cfg.Spiffe != nil {
		if t.spiffeX509Source == nil {
			return nil, errors.New("SPIFFE is enabled for this transport, but not configured")
		}

		authorizer, err := buildSpiffeAuthorizer(cfg.Spiffe)
		if err != nil {
			return nil, fmt.Errorf("unable to build SPIFFE authorizer: %w", err)
		}

		tlsConfig = tlsconfig.MTLSClientConfig(t.spiffeX509Source, t.spiffeX509Source, authorizer)
	}

	if cfg.InsecureSkipVerify || len(cfg.RootCAs) > 0 || len(cfg.ServerName) > 0 || len(cfg.Certificates) > 0 || cfg.PeerCertURI != "" {
		if tlsConfig != nil {
			return nil, errors.New("TLS and SPIFFE configuration cannot be defined at the same time")
		}

		tlsConfig = &tls.Config{
			ServerName:         cfg.ServerName,
			InsecureSkipVerify: cfg.InsecureSkipVerify,
			RootCAs:            createRootCAPool(cfg.RootCAs),
			Certificates:       cfg.Certificates.GetCertificates(),
		}

		if cfg.PeerCertURI != "" {
			tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
				return traefiktls.VerifyPeerCertificate(cfg.PeerCertURI, tlsConfig, rawCerts)
			}
		}
	}

	return tlsConfig, nil
}

func createRootCAPool(rootCAs []traefiktls.FileOrContent) *x509.CertPool {
	if len(rootCAs) == 0 {
		return nil
	}

	roots := x509.NewCertPool()
	for _, cert := range rootCAs {
		certContent, err := cert.Read()
		if err != nil {
			log.Error().
				Err(err).
				Msg("Unable to read RootCA")

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
