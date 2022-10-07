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

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
	"golang.org/x/net/proxy"
)

// DialerManager handles dialer for the reverse proxy.
type DialerManager struct {
	rtLock  sync.RWMutex
	dialers map[string]proxy.Dialer
	configs map[string]*dynamic.TCPServersTransport
}

// NewDialerManager creates a new DialerManager.
func NewDialerManager() *DialerManager {
	return &DialerManager{
		dialers: make(map[string]proxy.Dialer),
		configs: make(map[string]*dynamic.TCPServersTransport),
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
		d.dialers[configName], err = createDialer(newConfig)
		if err != nil {
			log.WithoutContext().Errorf("Could not configure TCP Dialer %s, fallback on default dialer: %v", configName, err)
			d.dialers[configName] = &net.Dialer{}
		}
	}

	for newConfigName, newConfig := range newConfigs {
		if _, ok := d.configs[newConfigName]; ok {
			continue
		}

		var err error
		d.dialers[newConfigName], err = createDialer(newConfig)
		if err != nil {
			log.WithoutContext().Errorf("Could not configure TCP Dialer %s, fallback on default dialer: %v", newConfigName, err)
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
func createDialer(cfg *dynamic.TCPServersTransport) (proxy.Dialer, error) {
	if cfg == nil {
		return nil, errors.New("no transport configuration given")
	}

	dialer := &net.Dialer{
		Timeout:   time.Duration(cfg.DialTimeout),
		KeepAlive: time.Duration(cfg.DialKeepAlive),
	}

	if !cfg.InsecureSkipVerify && len(cfg.RootCAs) == 0 && len(cfg.ServerName) == 0 && len(cfg.Certificates) == 0 && cfg.PeerCertURI == "" {
		return dialer, nil
	}

	tlsConfig := &tls.Config{
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
			log.WithoutContext().Error("Error while read RootCAs", err)
			continue
		}
		roots.AppendCertsFromPEM(certContent)
	}

	return roots
}
