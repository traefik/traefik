package server

import (
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/server/internal"
	"github.com/containous/traefik/tls"
)

func mergeConfiguration(configurations config.Configurations) config.Configuration {
	conf := config.Configuration{
		HTTP: &config.HTTPConfiguration{
			Routers:     make(map[string]*config.Router),
			Middlewares: make(map[string]*config.Middleware),
			Services:    make(map[string]*config.Service),
		},
		TCP: &config.TCPConfiguration{
			Routers:  make(map[string]*config.TCPRouter),
			Services: make(map[string]*config.TCPService),
		},
		TLSOptions: make(map[string]tls.TLS),
		TLSStores:  make(map[string]tls.Store),
	}

	for provider, configuration := range configurations {
		if configuration.HTTP != nil {
			for routerName, router := range configuration.HTTP.Routers {
				conf.HTTP.Routers[internal.MakeQualifiedName(provider, routerName)] = router
			}
			for middlewareName, middleware := range configuration.HTTP.Middlewares {
				conf.HTTP.Middlewares[internal.MakeQualifiedName(provider, middlewareName)] = middleware
			}
			for serviceName, service := range configuration.HTTP.Services {
				conf.HTTP.Services[internal.MakeQualifiedName(provider, serviceName)] = service
			}
		}

		if configuration.TCP != nil {
			for routerName, router := range configuration.TCP.Routers {
				conf.TCP.Routers[internal.MakeQualifiedName(provider, routerName)] = router
			}
			for serviceName, service := range configuration.TCP.Services {
				conf.TCP.Services[internal.MakeQualifiedName(provider, serviceName)] = service
			}
		}
		conf.TLS = append(conf.TLS, configuration.TLS...)

		for key, store := range configuration.TLSStores {
			conf.TLSStores[key] = store
		}

		for key, config := range configuration.TLSOptions {
			conf.TLSOptions[key] = config
		}
	}

	return conf
}
