package server

import (
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/server/internal"
	"github.com/containous/traefik/v2/pkg/tls"
)

func mergeConfiguration(configurations dynamic.Configurations) dynamic.Configuration {
	conf := dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:     make(map[string]*dynamic.Router),
			Middlewares: make(map[string]*dynamic.Middleware),
			Services:    make(map[string]*dynamic.Service),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:  make(map[string]*dynamic.TCPRouter),
			Services: make(map[string]*dynamic.TCPService),
		},
		TLS: &dynamic.TLSConfiguration{
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
