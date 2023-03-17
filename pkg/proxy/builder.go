package proxy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sync"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/proxy/fasthttp"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
)

type TLSConfigManager interface {
	GetTLSConfig(name string) (*tls.Config, error)
}

// Builder is a proxy builder which returns a fasthttp or httputil proxy corresponding
// to the ServersTransport configuration.
type Builder struct {
	fasthttpBuilder *fasthttp.ProxyBuilder
	httputilBuilder *httputil.ProxyBuilder

	tlsClientConfigManager TLSConfigManager

	configsLock sync.RWMutex
	configs     map[string]*dynamic.ServersTransport
}

// NewBuilder creates and returns a new Builder instance.
func NewBuilder(tlsClientConfigManager TLSConfigManager) *Builder {
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

	if config.EnableHTTP2 && targetURL.Scheme == "https" || targetURL.Scheme == "h2c" || os.Getenv("TRAEFIK_FASTHTTP_DISABLE") == "1" {
		return b.httputilBuilder.Build(configName, config, tlsConfig, targetURL)
	}

	return b.fasthttpBuilder.Build(configName, config, tlsConfig, targetURL)
}
