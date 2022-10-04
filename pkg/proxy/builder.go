package proxy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/metrics"
	"github.com/traefik/traefik/v3/pkg/proxy/fast"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
)

// TransportManager manages transport used for backend communications.
// FIXME duplicate??
type TransportManager interface {
	Get(name string) (*dynamic.ServersTransport, error)
	GetRoundTripper(name string) (http.RoundTripper, error)
	GetTLSConfig(name string) (*tls.Config, error)
}

// Builder is a proxy builder which returns a fasthttp or httputil proxy corresponding
// to the ServersTransport configuration.
type Builder struct {
	fastProxyBuilder *fast.ProxyBuilder
	httputilBuilder  *httputil.ProxyBuilder

	transportManager httputil.TransportManager
	fastProxy        bool
}

// NewBuilder creates and returns a new Builder instance.
func NewBuilder(transportManager TransportManager, semConvMetricsRegistry *metrics.SemConvMetricsRegistry, fastProxy bool) *Builder {
	return &Builder{
		fastProxyBuilder: fast.NewProxyBuilder(transportManager),
		httputilBuilder:  httputil.NewProxyBuilder(transportManager, semConvMetricsRegistry),
		transportManager: transportManager,
		fastProxy:        fastProxy,
	}
}

// Update is the handler called when the dynamic configuration is updated.
func (b *Builder) Update(newConfigs map[string]*dynamic.ServersTransport) {
	b.fastProxyBuilder.Update(newConfigs)
}

// Build builds an HTTP proxy for the given URL using the ServersTransport with the given name.
func (b *Builder) Build(configName string, targetURL *url.URL, shouldObserve, passHostHeader bool, flushInterval time.Duration) (http.Handler, error) {
	serversTransport, err := b.transportManager.Get(configName)
	if err != nil {
		return nil, fmt.Errorf("getting ServersTransport: %w", err)
	}

	if !b.fastProxy || !serversTransport.DisableHTTP2 && (targetURL.Scheme == "https" || targetURL.Scheme == "h2c") {
		return b.httputilBuilder.Build(configName, targetURL, shouldObserve, passHostHeader, flushInterval)
	}

	return b.fastProxyBuilder.Build(configName, targetURL, passHostHeader)
}
