package fast

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
)

// TransportManager manages transport used for backend communications.
type TransportManager interface {
	Get(name string) (*dynamic.ServersTransport, error)
	GetTLSConfig(name string) (*tls.Config, error)
}

// ProxyBuilder handles the connection pools for the FastProxy proxies.
type ProxyBuilder struct {
	debug            bool
	transportManager TransportManager

	// lock isn't needed because ProxyBuilder is not called concurrently.
	pools map[string]map[string]*connPool
	proxy func(*http.Request) (*url.URL, error)

	// not goroutine safe.
	configs map[string]*dynamic.ServersTransport
}

// NewProxyBuilder creates a new ProxyBuilder.
func NewProxyBuilder(transportManager TransportManager, config static.FastProxyConfig) *ProxyBuilder {
	return &ProxyBuilder{
		debug:            config.Debug,
		transportManager: transportManager,
		pools:            make(map[string]map[string]*connPool),
		proxy:            http.ProxyFromEnvironment,
		configs:          make(map[string]*dynamic.ServersTransport),
	}
}

// Update updates all the round-tripper corresponding to the given configs.
// This method must not be used concurrently.
func (r *ProxyBuilder) Update(newConfigs map[string]*dynamic.ServersTransport) {
	for configName := range r.configs {
		if _, ok := newConfigs[configName]; !ok {
			for _, c := range r.pools[configName] {
				c.Close()
			}
			delete(r.pools, configName)
		}
	}

	for newConfigName, newConfig := range newConfigs {
		if !reflect.DeepEqual(newConfig, r.configs[newConfigName]) {
			for _, c := range r.pools[newConfigName] {
				c.Close()
			}
			delete(r.pools, newConfigName)
		}
	}

	r.configs = newConfigs
}

// Build builds a new ReverseProxy with the given configuration.
func (r *ProxyBuilder) Build(cfgName string, targetURL *url.URL, passHostHeader, preservePath bool) (http.Handler, error) {
	proxyURL, err := r.proxy(&http.Request{URL: targetURL})
	if err != nil {
		return nil, fmt.Errorf("getting proxy: %w", err)
	}

	cfg, err := r.transportManager.Get(cfgName)
	if err != nil {
		return nil, fmt.Errorf("getting ServersTransport: %w", err)
	}

	tlsConfig, err := r.transportManager.GetTLSConfig(cfgName)
	if err != nil {
		return nil, fmt.Errorf("getting TLS config: %w", err)
	}

	pool := r.getPool(cfgName, cfg, tlsConfig, targetURL, proxyURL)
	return NewReverseProxy(targetURL, proxyURL, r.debug, passHostHeader, preservePath, pool)
}

func (r *ProxyBuilder) getPool(cfgName string, config *dynamic.ServersTransport, tlsConfig *tls.Config, targetURL *url.URL, proxyURL *url.URL) *connPool {
	pool, ok := r.pools[cfgName]
	if !ok {
		pool = make(map[string]*connPool)
		r.pools[cfgName] = pool
	}

	if connPool, ok := pool[targetURL.String()]; ok {
		return connPool
	}

	idleConnTimeout := 90 * time.Second
	dialTimeout := 30 * time.Second
	var responseHeaderTimeout time.Duration
	if config.ForwardingTimeouts != nil {
		idleConnTimeout = time.Duration(config.ForwardingTimeouts.IdleConnTimeout)
		dialTimeout = time.Duration(config.ForwardingTimeouts.DialTimeout)
		responseHeaderTimeout = time.Duration(config.ForwardingTimeouts.ResponseHeaderTimeout)
	}

	proxyDialer := newDialer(dialerConfig{
		DialKeepAlive: 0,
		DialTimeout:   dialTimeout,
		HTTP:          true,
		TLS:           targetURL.Scheme == "https",
		ProxyURL:      proxyURL,
	}, tlsConfig)

	connPool := newConnPool(config.MaxIdleConnsPerHost, idleConnTimeout, responseHeaderTimeout, func() (net.Conn, error) {
		return proxyDialer.Dial("tcp", addrFromURL(targetURL))
	})

	r.pools[cfgName][targetURL.String()] = connPool

	return connPool
}
