package server

import (
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/server/provider"
	"github.com/traefik/traefik/v2/pkg/tls"
)

func mergeConfiguration(configurations dynamic.Configurations, defaultEntryPoints []string) dynamic.Configuration {
	conf := dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:     make(map[string]*dynamic.Router),
			Middlewares: make(map[string]*dynamic.Middleware),
			Services:    make(map[string]*dynamic.Service),
			Models:      make(map[string]*dynamic.Model),
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
				if len(router.EntryPoints) == 0 {
					log.WithoutContext().
						WithField(log.RouterName, routerName).
						Debugf("No entryPoint defined for this router, using the default one(s) instead: %+v", defaultEntryPoints)
					router.EntryPoints = defaultEntryPoints
				}

				conf.HTTP.Routers[provider.MakeQualifiedName(pvd, routerName)] = router
			}
			for middlewareName, middleware := range configuration.HTTP.Middlewares {
				conf.HTTP.Middlewares[provider.MakeQualifiedName(pvd, middlewareName)] = middleware
			}
			for serviceName, service := range configuration.HTTP.Services {
				conf.HTTP.Services[provider.MakeQualifiedName(pvd, serviceName)] = service
			}
			for modelName, model := range configuration.HTTP.Models {
				conf.HTTP.Models[provider.MakeQualifiedName(pvd, modelName)] = model
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

func applyModel(cfg dynamic.Configuration) dynamic.Configuration {
	if cfg.HTTP == nil || len(cfg.HTTP.Models) == 0 {
		return cfg
	}

	rts := make(map[string]*dynamic.Router)

	for name, rt := range cfg.HTTP.Routers {
		router := rt.DeepCopy()

		eps := router.EntryPoints
		router.EntryPoints = nil

		for _, epName := range eps {
			m, ok := cfg.HTTP.Models[epName+"@internal"]
			if ok {
				cp := router.DeepCopy()

				cp.EntryPoints = []string{epName}

				if cp.TLS == nil {
					cp.TLS = m.TLS
				}

				cp.Middlewares = append(m.Middlewares, cp.Middlewares...)

				rtName := name
				if len(eps) > 1 {
					rtName = epName + "-" + name
				}
				rts[rtName] = cp
			} else {
				router.EntryPoints = append(router.EntryPoints, epName)

				rts[name] = router
			}
		}
	}

	cfg.HTTP.Routers = rts

	return cfg
}
