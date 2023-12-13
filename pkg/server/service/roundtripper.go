package service

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"reflect"
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
	"golang.org/x/net/http2"
)

type h2cTransportWrapper struct {
	*http2.Transport
}

func (t *h2cTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	return t.Transport.RoundTrip(req)
}

// SpiffeX509Source allows to retrieve a x509 SVID and bundle.
type SpiffeX509Source interface {
	x509svid.Source
	x509bundle.Source
}

// NewRoundTripperManager creates a new RoundTripperManager.
func NewRoundTripperManager(spiffeX509Source SpiffeX509Source) *RoundTripperManager {
	return &RoundTripperManager{
		roundTrippers:    make(map[string]http.RoundTripper),
		configs:          make(map[string]*dynamic.ServersTransport),
		spiffeX509Source: spiffeX509Source,
	}
}

// RoundTripperManager handles roundtripper for the reverse proxy.
type RoundTripperManager struct {
	rtLock        sync.RWMutex
	roundTrippers map[string]http.RoundTripper
	configs       map[string]*dynamic.ServersTransport

	spiffeX509Source SpiffeX509Source
}

// Update updates the roundtrippers configurations.
func (r *RoundTripperManager) Update(newConfigs map[string]*dynamic.ServersTransport) {
	r.rtLock.Lock()
	defer r.rtLock.Unlock()

	for configName, config := range r.configs {
		newConfig, ok := newConfigs[configName]
		if !ok {
			delete(r.configs, configName)
			delete(r.roundTrippers, configName)
			continue
		}

		if reflect.DeepEqual(newConfig, config) {
			continue
		}

		var err error
		r.roundTrippers[configName], err = r.createRoundTripper(newConfig)
		if err != nil {
			log.Error().Err(err).Msgf("Could not configure HTTP Transport %s, fallback on default transport", configName)
			r.roundTrippers[configName] = http.DefaultTransport
		}
	}

	for newConfigName, newConfig := range newConfigs {
		if _, ok := r.configs[newConfigName]; ok {
			continue
		}

		var err error
		r.roundTrippers[newConfigName], err = r.createRoundTripper(newConfig)
		if err != nil {
			log.Error().Err(err).Msgf("Could not configure HTTP Transport %s, fallback on default transport", newConfigName)
			r.roundTrippers[newConfigName] = http.DefaultTransport
		}
	}

	r.configs = newConfigs
}

// Get gets a roundtripper by name.
func (r *RoundTripperManager) Get(name string) (http.RoundTripper, error) {
	if len(name) == 0 {
		name = "default@internal"
	}

	r.rtLock.RLock()
	defer r.rtLock.RUnlock()

	if rt, ok := r.roundTrippers[name]; ok {
		return rt, nil
	}

	return nil, fmt.Errorf("servers transport not found %s", name)
}

// createRoundTripper creates an http.RoundTripper configured with the Transport configuration settings.
// For the settings that can't be configured in Traefik it uses the default http.Transport settings.
// An exception to this is the MaxIdleConns setting as we only provide the option MaxIdleConnsPerHost in Traefik at this point in time.
// Setting this value to the default of 100 could lead to confusing behavior and backwards compatibility issues.
func (r *RoundTripperManager) createRoundTripper(cfg *dynamic.ServersTransport) (http.RoundTripper, error) {
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
	}

	if cfg.ForwardingTimeouts != nil {
		transport.ResponseHeaderTimeout = time.Duration(cfg.ForwardingTimeouts.ResponseHeaderTimeout)
		transport.IdleConnTimeout = time.Duration(cfg.ForwardingTimeouts.IdleConnTimeout)
	}

	if cfg.Spiffe != nil {
		if r.spiffeX509Source == nil {
			return nil, errors.New("SPIFFE is enabled for this transport, but not configured")
		}

		spiffeAuthorizer, err := buildSpiffeAuthorizer(cfg.Spiffe)
		if err != nil {
			return nil, fmt.Errorf("unable to build SPIFFE authorizer: %w", err)
		}

		transport.TLSClientConfig = tlsconfig.MTLSClientConfig(r.spiffeX509Source, r.spiffeX509Source, spiffeAuthorizer)
	}

	if cfg.InsecureSkipVerify || len(cfg.RootCAs) > 0 || len(cfg.ServerName) > 0 || len(cfg.Certificates) > 0 || cfg.PeerCertURI != "" {
		if transport.TLSClientConfig != nil {
			return nil, errors.New("TLS and SPIFFE configuration cannot be defined at the same time")
		}

		transport.TLSClientConfig = &tls.Config{
			ServerName:         cfg.ServerName,
			InsecureSkipVerify: cfg.InsecureSkipVerify,
			RootCAs:            createRootCACertPool(cfg.RootCAs),
			Certificates:       cfg.Certificates.GetCertificates(),
		}

		if cfg.PeerCertURI != "" {
			transport.TLSClientConfig.VerifyPeerCertificate = func(rawCerts [][]byte, _ [][]*x509.Certificate) error {
				return traefiktls.VerifyPeerCertificate(cfg.PeerCertURI, transport.TLSClientConfig, rawCerts)
			}
		}
	}

	// Return directly HTTP/1.1 transport when HTTP/2 is disabled
	if cfg.DisableHTTP2 {
		return transport, nil
	}

	return newSmartRoundTripper(transport, cfg.ForwardingTimeouts)
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
