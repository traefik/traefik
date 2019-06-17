package server

import (
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/server/internal"
	"github.com/containous/traefik/pkg/tls"
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

	var defaultTLSOptionProviders []string
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

		for tlsOptionsName, config := range configuration.TLSOptions {
			if tlsOptionsName != "default" {
				tlsOptionsName = internal.MakeQualifiedName(provider, tlsOptionsName)
			} else {
				defaultTLSOptionProviders = append(defaultTLSOptionProviders, provider)
			}

			conf.TLSOptions[tlsOptionsName] = config
		}
	}

	if len(defaultTLSOptionProviders) == 0 {
		conf.TLSOptions["default"] = tls.TLS{}
	} else if len(defaultTLSOptionProviders) > 1 {
		log.WithoutContext().Errorf("Default TLS Options defined multiple times in %v", defaultTLSOptionProviders)
		delete(conf.TLSOptions, "default")
	}

	return conf
}
