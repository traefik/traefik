package server

import (
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/server/provider"
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
		UDP: &dynamic.UDPConfiguration{
			Routers:  make(map[string]*dynamic.UDPRouter),
			Services: make(map[string]*dynamic.UDPService),
		},
		TLS: &dynamic.TLSConfiguration{
			Stores:  make(map[string]tls.Store),
			Options: make(map[string]tls.Options),
		},
	}

	var defaultTLSOptionProviders []string
	var defaultTLSStoreProviders []string
	for pvd, configuration := range configurations {
		if configuration.HTTP != nil {
			for routerName, router := range configuration.HTTP.Routers {
				conf.HTTP.Routers[provider.MakeQualifiedName(pvd, routerName)] = router
			}
			for middlewareName, middleware := range configuration.HTTP.Middlewares {
				conf.HTTP.Middlewares[provider.MakeQualifiedName(pvd, middlewareName)] = middleware
			}
			for serviceName, service := range configuration.HTTP.Services {
				conf.HTTP.Services[provider.MakeQualifiedName(pvd, serviceName)] = service
			}
		}

		if configuration.TCP != nil {
			for routerName, router := range configuration.TCP.Routers {
				conf.TCP.Routers[provider.MakeQualifiedName(pvd, routerName)] = router
			}
			for serviceName, service := range configuration.TCP.Services {
				conf.TCP.Services[provider.MakeQualifiedName(pvd, serviceName)] = service
			}
		}

		if configuration.UDP != nil {
			for routerName, router := range configuration.UDP.Routers {
				conf.UDP.Routers[provider.MakeQualifiedName(pvd, routerName)] = router
			}
			for serviceName, service := range configuration.UDP.Services {
				conf.UDP.Services[provider.MakeQualifiedName(pvd, serviceName)] = service
			}
		}

		if configuration.TLS != nil {
			conf.TLS.Certificates = append(conf.TLS.Certificates, configuration.TLS.Certificates...)

			for key, store := range configuration.TLS.Stores {
				if key != "default" {
					key = provider.MakeQualifiedName(pvd, key)
				} else {
					defaultTLSStoreProviders = append(defaultTLSStoreProviders, pvd)
				}
				conf.TLS.Stores[key] = store
			}

			for tlsOptionsName, options := range configuration.TLS.Options {
				if tlsOptionsName != "default" {
					tlsOptionsName = provider.MakeQualifiedName(pvd, tlsOptionsName)
				} else {
					defaultTLSOptionProviders = append(defaultTLSOptionProviders, pvd)
				}

				conf.TLS.Options[tlsOptionsName] = options
			}
		}
	}

	if len(defaultTLSStoreProviders) > 1 {
		log.WithoutContext().Errorf("Default TLS Stores defined multiple times in %v", defaultTLSOptionProviders)
		delete(conf.TLS.Stores, "default")
	}

	if len(defaultTLSOptionProviders) == 0 {
		conf.TLS.Options["default"] = tls.DefaultTLSOptions
	} else if len(defaultTLSOptionProviders) > 1 {
		log.WithoutContext().Errorf("Default TLS Options defined multiple times in %v", defaultTLSOptionProviders)
		// We do not set an empty tls.TLS{} as above so that we actually get a "cascading failure" later on,
		// i.e. routers depending on this missing TLS option will fail to initialize as well.
		delete(conf.TLS.Options, "default")
	}

	return conf
}
