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

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
	"golang.org/x/net/http2"
)

type h2cTransportWrapper struct {
	*http2.Transport
}

func (t *h2cTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	return t.Transport.RoundTrip(req)
}

// RoundTripperManager handles roundtripper for the reverse proxy.
type RoundTripperManager struct {
	rtLock        sync.RWMutex
	roundTrippers map[string]http.RoundTripper
	configs       map[string]*dynamic.ServersTransport
}

// NewRoundTripperManager creates a new RoundTripperManager.
func NewRoundTripperManager() *RoundTripperManager {
	return &RoundTripperManager{
		roundTrippers: make(map[string]http.RoundTripper),
		configs:       make(map[string]*dynamic.ServersTransport),
	}
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
		r.roundTrippers[configName], err = createRoundTripper(newConfig)
		if err != nil {
			log.WithoutContext().Errorf("Could not configure HTTP Transport %s, fallback on default transport: %v", configName, err)
			r.roundTrippers[configName] = http.DefaultTransport
		}
	}

	for newConfigName, newConfig := range newConfigs {
		if _, ok := r.configs[newConfigName]; ok {
			continue
		}

		var err error
		r.roundTrippers[newConfigName], err = createRoundTripper(newConfig)
		if err != nil {
			log.WithoutContext().Errorf("Could not configure HTTP Transport %s, fallback on default transport: %v", newConfigName, err)
			r.roundTrippers[newConfigName] = http.DefaultTransport
		}
	}

	r.configs = newConfigs
}

// Get get a roundtripper by name.
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
func createRoundTripper(cfg *dynamic.ServersTransport) (http.RoundTripper, error) {
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

	if cfg.InsecureSkipVerify || len(cfg.RootCAs) > 0 || len(cfg.ServerName) > 0 || len(cfg.Certificates) > 0 || cfg.PeerCertURI != "" {
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
		return &KerberosRoundTripper{
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
	return &KerberosRoundTripper{
		OriginalRoundTripper: rt,
		new: func() http.RoundTripper {
			return rt.Clone()
		},
	}, nil
}

type KerberosRoundTripper struct {
	new                  func() http.RoundTripper
	OriginalRoundTripper http.RoundTripper
}

// stickyRoundTrippers is keyed by owner, as there is one KerberosRoundTripper per ServersTransport and a dedicated
// round tripper carries the whole transport configuration (TLS client configuration, timeouts, connection pool) of
// the ServersTransport that created it, hence cannot be shared with another one.
// The map is lazily allocated, as most client connections never reach a Kerberos or NTLM backend.
type stickyRoundTrippers struct {
	mu            sync.Mutex
	roundTrippers map[*KerberosRoundTripper]http.RoundTripper
}

func (s *stickyRoundTrippers) get(owner *KerberosRoundTripper) http.RoundTripper {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.roundTrippers[owner]
}

// stick keeps the round tripper already stored for the given owner, as concurrent requests on the same client
// connection (HTTP/2 streams) can race on the same owner, and replacing an in-use round tripper would orphan the
// backend connections it holds.
func (s *stickyRoundTrippers) stick(owner *KerberosRoundTripper) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.roundTrippers[owner]; ok {
		return
	}

	if s.roundTrippers == nil {
		s.roundTrippers = make(map[*KerberosRoundTripper]http.RoundTripper)
	}

	s.roundTrippers[owner] = owner.new()
}

type transportKeyType string

var transportKey transportKeyType = "transport"

func AddTransportOnContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, transportKey, &stickyRoundTrippers{})
}

func (k *KerberosRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	connRoundTrippers, ok := request.Context().Value(transportKey).(*stickyRoundTrippers)
	if !ok {
		return k.OriginalRoundTripper.RoundTrip(request)
	}

	if roundTripper := connRoundTrippers.get(k); roundTripper != nil {
		return roundTripper.RoundTrip(request)
	}

	resp, err := k.OriginalRoundTripper.RoundTrip(request)

	// If we found that we are authenticating with Kerberos (Negotiate) or NTLM.
	// We put a dedicated roundTripper in the ConnContext.
	// This will stick the next calls to the same connection with the backend.
	if err == nil && containsNTLMorNegotiate(resp.Header.Values("WWW-Authenticate")) {
		connRoundTrippers.stick(k)
	}
	return resp, err
}

func containsNTLMorNegotiate(h []string) bool {
	return slices.ContainsFunc(h, func(s string) bool {
		// RFC 9110 section 11.1 defines the auth-scheme as case-insensitive,
		// hence a challenge is matched on its scheme token whatever its case.
		scheme, _, _ := strings.Cut(s, " ")
		return strings.EqualFold(scheme, "NTLM") || strings.EqualFold(scheme, "Negotiate")
	})
}

func createRootCACertPool(rootCAs []traefiktls.FileOrContent) *x509.CertPool {
	if len(rootCAs) == 0 {
		return nil
	}

	roots := x509.NewCertPool()

	for _, cert := range rootCAs {
		certContent, err := cert.Read()
		if err != nil {
			log.WithoutContext().Error("Error while read RootCAs", err)
			continue
		}
		roots.AppendCertsFromPEM(certContent)
	}

	return roots
}
