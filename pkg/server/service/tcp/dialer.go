package tcp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spiffe/go-spiffe/v2/bundle/x509bundle"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
	"golang.org/x/net/proxy"
)

// SpiffeX509Source allows to retrieve a x509 SVID and bundle.
type SpiffeX509Source interface {
	x509svid.Source
	x509bundle.Source
}

// DialerManager handles dialer for the reverse proxy.
type DialerManager struct {
	rtLock           sync.RWMutex
	dialers          map[string]proxy.Dialer
	configs          map[string]*dynamic.TCPServersTransport
	spiffeX509Source SpiffeX509Source
}

// NewDialerManager creates a new DialerManager.
func NewDialerManager(spiffeX509Source SpiffeX509Source) *DialerManager {
	return &DialerManager{
		dialers:          make(map[string]proxy.Dialer),
		configs:          make(map[string]*dynamic.TCPServersTransport),
		spiffeX509Source: spiffeX509Source,
	}
}

// Update updates the dialers configurations.
func (d *DialerManager) Update(newConfigs map[string]*dynamic.TCPServersTransport) {
	d.rtLock.Lock()
	defer d.rtLock.Unlock()

	for configName, config := range d.configs {
		newConfig, ok := newConfigs[configName]
		if !ok {
			delete(d.configs, configName)
			delete(d.dialers, configName)
			continue
		}

		if reflect.DeepEqual(newConfig, config) {
			continue
		}

		var err error
		d.dialers[configName], err = d.createDialer(newConfig)
		if err != nil {
			log.Error().
				Str("dialer", configName).
				Err(err).
				Msg("Could not configure TCP Dialer, fallback on default dialer")
			d.dialers[configName] = &net.Dialer{}
		}
	}

	for newConfigName, newConfig := range newConfigs {
		if _, ok := d.configs[newConfigName]; ok {
			continue
		}

		var err error
		d.dialers[newConfigName], err = d.createDialer(newConfig)
		if err != nil {
			log.Error().
				Str("dialer", newConfigName).
				Err(err).
				Msg("Could not configure TCP Dialer, fallback on default dialer")
			d.dialers[newConfigName] = &net.Dialer{}
		}
	}

	d.configs = newConfigs
}

// Get gets a dialer by name.
func (d *DialerManager) Get(name string) (proxy.Dialer, error) {
	if len(name) == 0 {
		name = "default@internal"
	}

	d.rtLock.RLock()
	defer d.rtLock.RUnlock()

	if rt, ok := d.dialers[name]; ok {
		return rt, nil
	}

	return nil, fmt.Errorf("TCP dialer not found %s", name)
}

// createDialer creates a proxy.Dialer with the TLS configuration settings if needed.
func (d *DialerManager) createDialer(cfg *dynamic.TCPServersTransport) (proxy.Dialer, error) {
	if cfg == nil {
		return nil, errors.New("no transport configuration given")
	}

	dialer := &net.Dialer{
		Timeout:   time.Duration(cfg.DialTimeout),
		KeepAlive: time.Duration(cfg.DialKeepAlive),
	}

	var tlsConfig *tls.Config

	if cfg.Spiffe != nil {
		if d.spiffeX509Source == nil {
			return nil, errors.New("SPIFFE is enabled for this transport, but not configured")
		}

		spiffeAuthorizer, err := buildSpiffeAuthorizer(cfg.Spiffe)
		if err != nil {
			return nil, fmt.Errorf("unable to build SPIFFE authorizer: %w", err)
		}

		tlsConfig = tlsconfig.MTLSClientConfig(d.spiffeX509Source, d.spiffeX509Source, spiffeAuthorizer)
	}

	if cfg.InsecureSkipVerify || len(cfg.RootCAs) > 0 || len(cfg.ServerName) > 0 || len(cfg.Certificates) > 0 || cfg.PeerCertURI != "" {
		if tlsConfig != nil {
			return nil, errors.New("TLS and SPIFFE configuration cannot be defined at the same time")
		}

		tlsConfig = &tls.Config{
			ServerName:         cfg.ServerName,
			InsecureSkipVerify: cfg.InsecureSkipVerify,
			RootCAs:            createRootCACertPool(cfg.RootCAs),
			Certificates:       cfg.Certificates.GetCertificates(),
		}

		if cfg.PeerCertURI != "" {
			tlsConfig.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
				return traefiktls.VerifyPeerCertificate(cfg.PeerCertURI, tlsConfig, rawCerts)
			}
		}
	}

	if tlsConfig == nil {
		return dialer, nil
	}

	return &tls.Dialer{
		NetDialer: dialer,
		Config:    tlsConfig,
	}, nil
}

func createRootCACertPool(rootCAs []traefiktls.FileOrContent) *x509.CertPool {
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
