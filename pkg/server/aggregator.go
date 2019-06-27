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
		TLS: &config.TLSConfiguration{
			Stores:  make(map[string]tls.Store),
			Options: make(map[string]tls.Options),
		},
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

		if configuration.TLS != nil {
			conf.TLS.Certificates = append(conf.TLS.Certificates, configuration.TLS.Certificates...)

			for key, store := range configuration.TLS.Stores {
				conf.TLS.Stores[key] = store
			}

			for tlsOptionsName, options := range configuration.TLS.Options {
				if tlsOptionsName != "default" {
					tlsOptionsName = internal.MakeQualifiedName(provider, tlsOptionsName)
				} else {
					defaultTLSOptionProviders = append(defaultTLSOptionProviders, provider)
				}

				conf.TLS.Options[tlsOptionsName] = options
			}
		}
	}

	if len(defaultTLSOptionProviders) == 0 {
		conf.TLS.Options["default"] = tls.Options{}
	} else if len(defaultTLSOptionProviders) > 1 {
		log.WithoutContext().Errorf("Default TLS Options defined multiple times in %v", defaultTLSOptionProviders)
		// We do not set an empty tls.TLS{} as above so that we actually get a "cascading failure" later on,
		// i.e. routers depending on this missing TLS option will fail to initialize as well.
		delete(conf.TLS.Options, "default")
	}

	return conf
}
