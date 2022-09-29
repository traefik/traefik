package tcp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	//"errors"
	//"fmt"
	"net"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	//	"github.com/traefik/traefik/v2/pkg/server/service"

	"github.com/traefik/traefik/v2/pkg/log"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
	// "golang.org/x/net/http2"
)

// type h2cTransportWrapper struct {
// 	*http2.Transport
// }

// func (t *h2cTransportWrapper) RoundTrip(req *http.Request) (*http.Response, error) {
// 	req.URL.Scheme = "http"
// 	return t.Transport.RoundTrip(req)
// }

// NewRoundTripperManager creates a new RoundTripperManager.
func NewTcpManager() *TcpManager {
	return &TcpManager{
		transport: make(map[string]http.RoundTripper),
		configs:   make(map[string]*dynamic.ServersTransport),
	}
}

// RoundTripperManager handles roundtripper for the reverse proxy.
type TcpManager struct {
	rtLock    sync.RWMutex
	configs   map[string]*dynamic.ServersTransport
	transport map[string]http.RoundTripper
}

// Update updates the roundtrippers configurations.
func (r *TcpManager) Update(newConfigs map[string]*dynamic.ServersTransport) {
	r.rtLock.Lock()
	defer r.rtLock.Unlock()
	for configName, config := range r.configs {
		newConfig, ok := newConfigs[configName]
		if !ok {
			delete(r.configs, configName)
			delete(r.transport, configName)

			continue
		}
		// manager := service.NewRoundTripperManager()
		if reflect.DeepEqual(newConfig, config) {
			continue
		}
		var err error
		r.transport[configName], err = createTcptransport(newConfig)
		if err != nil {
			log.WithoutContext().Errorf("Could not configure HTTP Transport %s, fallback on default transport: %v", configName, err)
			r.transport[configName] = http.DefaultTransport
		}

	}

	for newConfigName := range newConfigs {
		if _, ok := r.configs[newConfigName]; ok {
			continue

		}
	}
	for newConfigName, newConfig := range newConfigs {
		if _, ok := r.configs[newConfigName]; ok {
			continue
		}

		var err error
		r.transport[newConfigName], err = createTcptransport(newConfig)
		if err != nil {
			log.WithoutContext().Errorf("Could not configure HTTP Transport %s, fallback on default transport: %v", newConfigName, err)
			r.transport[newConfigName] = http.DefaultTransport
		}
	}

	r.configs = newConfigs
}

// createTcptransport creates an initial tcp configurations configured with the Transport configuration settings.
// For the settings that can't be configured in Traefik it uses the default http.Transport settings.
// An exception to this is the MaxIdleConns setting as we only provide the option MaxIdleConnsPerHost in Traefik at this point in time.
// Setting this value to the default of 100 could lead to confusing behavior and backwards compatibility issues.
func createTcptransport(cfg *dynamic.ServersTransport) (*http.Transport, error) {
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
		return transport, nil
	}
	return transport, nil
}

// Get get a roundtripper by name.
func (t *TcpManager) Get(name string) (http.RoundTripper, error) {
	if len(name) == 0 {
		name = "default@internal"
	}

	t.rtLock.RLock()
	defer t.rtLock.RUnlock()

	if rt, ok := t.transport[name]; ok {
		return rt, nil
	}
	return nil, fmt.Errorf("servers transport not found %s", name)
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
