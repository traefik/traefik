package proxy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/proxy/fast"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"github.com/traefik/traefik/v3/pkg/server/service"
)

// TransportManager manages transport used for backend communications.
type TransportManager interface {
	Get(name string) (*dynamic.ServersTransport, error)
	GetRoundTripper(name string) (http.RoundTripper, error)
	GetTLSConfig(name string) (*tls.Config, error)
}

// SmartBuilder is a proxy builder which returns a fast proxy or httputil proxy corresponding
// to the ServersTransport configuration.
type SmartBuilder struct {
	fastProxyBuilder *fast.ProxyBuilder
	proxyBuilder     service.ProxyBuilder

	transportManager httputil.TransportManager
}

// NewSmartBuilder creates and returns a new SmartBuilder instance.
func NewSmartBuilder(transportManager TransportManager, proxyBuilder service.ProxyBuilder, fastProxyConfig static.FastProxyConfig) *SmartBuilder {
	return &SmartBuilder{
		fastProxyBuilder: fast.NewProxyBuilder(transportManager, fastProxyConfig),
		proxyBuilder:     proxyBuilder,
		transportManager: transportManager,
	}
}

// Update is the handler called when the dynamic configuration is updated.
func (b *SmartBuilder) Update(newConfigs map[string]*dynamic.ServersTransport) {
	b.fastProxyBuilder.Update(newConfigs)
}

// Build builds an HTTP proxy for the given URL using the ServersTransport with the given name.
func (b *SmartBuilder) Build(configName string, targetURL *url.URL, passHostHeader, preservePath bool, flushInterval time.Duration) (http.Handler, error) {
	serversTransport, err := b.transportManager.Get(configName)
	if err != nil {
		return nil, fmt.Errorf("getting ServersTransport: %w", err)
	}

	// The fast proxy implementation cannot handle HTTP/2 requests for now.
	// For the https scheme we cannot guess if the backend communication will use HTTP2,
	// thus we check if HTTP/2 is disabled to use the fast proxy implementation when this is possible.
	if targetURL.Scheme == "h2c" || (targetURL.Scheme == "https" && !serversTransport.DisableHTTP2) {
		return b.proxyBuilder.Build(configName, targetURL, passHostHeader, preservePath, flushInterval)
	}
	return b.fastProxyBuilder.Build(configName, targetURL, passHostHeader, preservePath)
}
