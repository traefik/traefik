package tcp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
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
	"golang.org/x/net/proxy"
)

type Dialer interface {
	proxy.Dialer

	TerminationDelay() time.Duration
}

type tcpDialer struct {
	proxy.Dialer
	terminationDelay time.Duration
}

func (d tcpDialer) TerminationDelay() time.Duration {
	return d.terminationDelay
}

// SpiffeX509Source allows to retrieve a x509 SVID and bundle.
type SpiffeX509Source interface {
	x509svid.Source
	x509bundle.Source
}

// DialerManager handles dialer for the reverse proxy.
type DialerManager struct {
	rtLock           sync.RWMutex
	dialers          map[string]Dialer
	dialersTLS       map[string]Dialer
	spiffeX509Source SpiffeX509Source
}

// NewDialerManager creates a new DialerManager.
func NewDialerManager(spiffeX509Source SpiffeX509Source) *DialerManager {
	return &DialerManager{
		dialers:          make(map[string]Dialer),
		dialersTLS:       make(map[string]Dialer),
		spiffeX509Source: spiffeX509Source,
	}
}

// Update updates the dialers configurations.
func (d *DialerManager) Update(configs map[string]*dynamic.TCPServersTransport) {
	d.rtLock.Lock()
	defer d.rtLock.Unlock()

	d.dialers = make(map[string]Dialer)
	d.dialersTLS = make(map[string]Dialer)
	for configName, config := range configs {
		if err := d.createDialers(configName, config); err != nil {
			log.Debug().
				Str("dialer", configName).
				Err(err).
				Msg("Create TCP Dialer")
		}
	}
}

// Get gets a dialer by name.
func (d *DialerManager) Get(name string, tls bool) (Dialer, error) {
	if len(name) == 0 {
		name = "default@internal"
	}

	d.rtLock.RLock()
	defer d.rtLock.RUnlock()

	if tls {
		if rt, ok := d.dialersTLS[name]; ok {
			return rt, nil
		}

		return nil, fmt.Errorf("TCP dialer not found %s", name)
	}

	if rt, ok := d.dialers[name]; ok {
		return rt, nil
	}

	return nil, fmt.Errorf("TCP dialer not found %s", name)
}

// createDialers creates the dialers according to the TCPServersTransport configuration.
func (d *DialerManager) createDialers(name string, cfg *dynamic.TCPServersTransport) error {
	if cfg == nil {
		return errors.New("no transport configuration given")
	}

	dialer := &net.Dialer{
		Timeout:   time.Duration(cfg.DialTimeout),
		KeepAlive: time.Duration(cfg.DialKeepAlive),
	}

	var tlsConfig *tls.Config

	if cfg.TLS != nil {
		if cfg.TLS.Spiffe != nil {
			if d.spiffeX509Source == nil {
				return errors.New("SPIFFE is enabled for this transport, but not configured")
			}

			authorizer, err := buildSpiffeAuthorizer(cfg.TLS.Spiffe)
			if err != nil {
				return fmt.Errorf("unable to build SPIFFE authorizer: %w", err)
			}

			tlsConfig = tlsconfig.MTLSClientConfig(d.spiffeX509Source, d.spiffeX509Source, authorizer)
		}

		if cfg.TLS.InsecureSkipVerify || len(cfg.TLS.RootCAs) > 0 || len(cfg.TLS.ServerName) > 0 || len(cfg.TLS.Certificates) > 0 || cfg.TLS.PeerCertURI != "" {
			if tlsConfig != nil {
				return errors.New("TLS and SPIFFE configuration cannot be defined at the same time")
			}

			tlsConfig = &tls.Config{
				ServerName:         cfg.TLS.ServerName,
				InsecureSkipVerify: cfg.TLS.InsecureSkipVerify,
				RootCAs:            createRootCACertPool(cfg.TLS.RootCAs),
				Certificates:       cfg.TLS.Certificates.GetCertificates(),
			}

			if cfg.TLS.PeerCertURI != "" {
				tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
					return traefiktls.VerifyPeerCertificate(cfg.TLS.PeerCertURI, tlsConfig, rawCerts)
				}
			}
		}
	}

	tlsDialer := &tls.Dialer{
		NetDialer: dialer,
		Config:    tlsConfig,
	}

	d.dialers[name] = tcpDialer{dialer, time.Duration(cfg.TerminationDelay)}
	d.dialersTLS[name] = tcpDialer{tlsDialer, time.Duration(cfg.TerminationDelay)}

	return nil
}

func createRootCACertPool(rootCAs []types.FileOrContent) *x509.CertPool {
	if len(rootCAs) == 0 {
		return nil
	}

	roots := x509.NewCertPool()

	for _, cert := range rootCAs {
		certContent, err := cert.Read()
		if err != nil {
			log.Err(err).Msg("Error while read RootCAs")
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
