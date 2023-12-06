package fasthttp

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/dialer"
)

// ProxyBuilder handles the connection pools for the FastHTTP proxies.
type ProxyBuilder struct {
	// lock isn't needed because ProxyBuilder is not called concurrently.
	pools map[string]map[string]*ConnPool
	proxy func(*http.Request) (*url.URL, error)
}

// NewProxyBuilder creates a new ProxyBuilder.
func NewProxyBuilder() *ProxyBuilder {
	return &ProxyBuilder{
		pools: make(map[string]map[string]*ConnPool),
		proxy: http.ProxyFromEnvironment,
	}
}

// Delete deletes the round-tripper corresponding to the given dynamic.HTTPClientConfig.
func (r *ProxyBuilder) Delete(cfgName string) {
	delete(r.pools, cfgName)
}

// Build builds a new ReverseProxy with the given configuration.
func (r *ProxyBuilder) Build(cfgName string, cfg *dynamic.ServersTransport, tlsConfig *tls.Config, targetURL *url.URL) (http.Handler, error) {
	proxyURL, err := r.proxy(&http.Request{URL: targetURL})
	if err != nil {
		return nil, err
	}

	var responseHeaderTimeout time.Duration
	if cfg.ForwardingTimeouts != nil {
		responseHeaderTimeout = time.Duration(cfg.ForwardingTimeouts.ResponseHeaderTimeout)
	}
	pool := r.getPool(cfgName, cfg, tlsConfig, targetURL, proxyURL)
	return NewReverseProxy(targetURL, proxyURL, cfg.PassHostHeader, responseHeaderTimeout, pool)
}

func (r *ProxyBuilder) getPool(cfgName string, config *dynamic.ServersTransport, tlsConfig *tls.Config, targetURL *url.URL, proxyURL *url.URL) *ConnPool {
	pool, ok := r.pools[cfgName]
	if !ok {
		pool = make(map[string]*ConnPool)
		r.pools[cfgName] = pool
	}

	if connPool, ok := pool[targetURL.String()]; ok {
		return connPool
	}

	idleConnTimeout := time.Duration(dynamic.DefaultIdleConnTimeout)
	dialTimeout := 30 * time.Second
	if config.ForwardingTimeouts != nil {
		idleConnTimeout = time.Duration(config.ForwardingTimeouts.IdleConnTimeout)
		dialTimeout = time.Duration(config.ForwardingTimeouts.DialTimeout)
	}

	proxyDialer := dialer.NewDialer(dialer.Config{
		DialKeepAlive: 0,
		DialTimeout:   dialTimeout,
		HTTP:          true,
		TLS:           targetURL.Scheme == "https",
		ProxyURL:      proxyURL,
	}, tlsConfig)

	connPool := NewConnPool(config.MaxIdleConnsPerHost, idleConnTimeout, func() (net.Conn, error) {
		return proxyDialer.Dial("tcp", dialer.AddrFromURL(targetURL))
	})

	r.pools[cfgName][targetURL.String()] = connPool

	return connPool
}
