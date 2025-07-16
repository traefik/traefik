package httputil

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/metrics"
)

// TransportManager manages transport used for backend communications.
type TransportManager interface {
	Get(name string) (*dynamic.ServersTransport, error)
	GetRoundTripper(name string) (http.RoundTripper, error)
	GetTLSConfig(name string) (*tls.Config, error)
}

// ProxyBuilder handles the http.RoundTripper for httputil reverse proxies.
type ProxyBuilder struct {
	bufferPool             *bufferPool
	transportManager       TransportManager
	semConvMetricsRegistry *metrics.SemConvMetricsRegistry
}

// NewProxyBuilder creates a new ProxyBuilder.
func NewProxyBuilder(transportManager TransportManager, semConvMetricsRegistry *metrics.SemConvMetricsRegistry) *ProxyBuilder {
	return &ProxyBuilder{
		bufferPool:             newBufferPool(),
		transportManager:       transportManager,
		semConvMetricsRegistry: semConvMetricsRegistry,
	}
}

// Update does nothing.
func (r *ProxyBuilder) Update(_ map[string]*dynamic.ServersTransport) {}

// Build builds a new httputil.ReverseProxy with the given configuration.
func (r *ProxyBuilder) Build(cfgName string, targetURL *url.URL, passHostHeader, preservePath bool, flushInterval time.Duration) (http.Handler, error) {
	roundTripper, err := r.transportManager.GetRoundTripper(cfgName)
	if err != nil {
		return nil, fmt.Errorf("getting RoundTripper: %w", err)
	}

	// Wrapping the roundTripper with the Tracing roundTripper,
	// to create, if necessary, the reverseProxy client span and the semConv client metric.
	roundTripper = newObservabilityRoundTripper(r.semConvMetricsRegistry, roundTripper)

	return buildSingleHostProxy(targetURL, passHostHeader, preservePath, flushInterval, roundTripper, r.bufferPool), nil
}
