package tcp

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/dialer"
	"golang.org/x/net/proxy"
)

type TLSConfigManager interface {
	GetTLSConfig(name string) (*tls.Config, error)
}

// DialerManager handles dialer for the reverse proxy.
type DialerManager struct {
	rtLock sync.RWMutex

	configs    map[string]*dynamic.TCPServersTransport
	tlsManager TLSConfigManager
}

// NewDialerManager creates a new DialerManager.
func NewDialerManager(tlsManager TLSConfigManager) *DialerManager {
	return &DialerManager{
		tlsManager: tlsManager,
	}
}

// Update updates the dialers configurations.
func (d *DialerManager) Update(stConfigs map[string]*dynamic.TCPServersTransport) {
	d.rtLock.Lock()
	defer d.rtLock.Unlock()

	d.configs = stConfigs
}

type Dialer interface {
	proxy.Dialer

	TerminationDelay() time.Duration
}

// Get gets a dialer by name.
func (d *DialerManager) Get(name string, isTls bool, proxyURL *url.URL) (Dialer, error) {
	if len(name) == 0 {
		name = "default"
	}

	d.rtLock.RLock()
	defer d.rtLock.RUnlock()
	transport, ok := d.configs[name]
	if !ok {
		return nil, fmt.Errorf("TCP dialer not found %s", name)
	}

	tlsConfig, err := d.tlsManager.GetTLSConfig(name)
	if err != nil {
		return nil, fmt.Errorf("error while creating dialer: %w", err)
	}

	dialer := dialer.NewDialer(dialer.Config{
		DialKeepAlive: time.Duration(transport.DialKeepAlive),
		DialTimeout:   time.Duration(transport.DialTimeout),
		TLS:           isTls,
		ProxyURL:      proxyURL,
	}, tlsConfig)

	return tcpDialer{dialer, time.Duration(transport.TerminationDelay)}, nil
}

type tcpDialer struct {
	proxy.Dialer
	terminationDelay time.Duration
}

func (d tcpDialer) TerminationDelay() time.Duration {
	return d.terminationDelay
}
