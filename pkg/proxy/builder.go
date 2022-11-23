package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"sync"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/proxy/fasthttp"
	"github.com/traefik/traefik/v2/pkg/proxy/httputil"
	"github.com/traefik/traefik/v2/pkg/tls/client"
)

// Builder is a proxy builder which returns a fasthttp or httputil proxy corresponding
// to the ServersTransport configuration.
type Builder struct {
	fasthttpBuilder *fasthttp.ProxyBuilder
	httputilBuilder *httputil.ProxyBuilder

	tlsClientConfigManager *client.TLSConfigManager

	configsLock sync.RWMutex
	configs     map[string]*dynamic.ServersTransport
}

// NewBuilder creates and returns a new Builder instance.
func NewBuilder(tlsClientConfigManager *client.TLSConfigManager) *Builder {
	return &Builder{
		fasthttpBuilder:        fasthttp.NewProxyBuilder(),
		httputilBuilder:        httputil.NewProxyBuilder(),
		tlsClientConfigManager: tlsClientConfigManager,
		configs:                make(map[string]*dynamic.ServersTransport),
	}
}

// Update is the handler called when the dynamic configuration is updated.
func (b *Builder) Update(newConfigs map[string]*dynamic.ServersTransport) {
	b.configsLock.Lock()
	defer b.configsLock.Unlock()

	for configName := range b.configs {
		if _, ok := newConfigs[configName]; !ok {
			b.httputilBuilder.Delete(configName)
			b.fasthttpBuilder.Delete(configName)
		}
	}

	for newConfigName, newConfig := range newConfigs {
		if !reflect.DeepEqual(newConfig, b.configs[newConfigName]) {
			// Delete previous builders cache because the configuration changed.
			b.httputilBuilder.Delete(newConfigName)
			b.fasthttpBuilder.Delete(newConfigName)
		}
	}

	b.configs = newConfigs
}

// Build builds an HTTP proxy for the given URL using the ServersTransport with the given name.
func (b *Builder) Build(configName string, targetURL *url.URL) (http.Handler, error) {
	if len(configName) == 0 {
		configName = "default"
	}

	b.configsLock.RLock()
	defer b.configsLock.RUnlock()

	config, ok := b.configs[configName]
	if !ok {
		return nil, fmt.Errorf("unknown ServersTransport:  %s", configName)
	}

	tlsConfig, err := b.tlsClientConfigManager.GetTLSConfig(configName)
	if err != nil {
		return nil, err
	}

	if config.HTTP != nil && config.HTTP.EnableHTTP2 && targetURL.Scheme == "https" || targetURL.Scheme == "h2c" {
		return b.httputilBuilder.Build(configName, config.HTTP, tlsConfig, targetURL)
	}

	return b.fasthttpBuilder.Build(configName, config.HTTP, tlsConfig, targetURL)
}
