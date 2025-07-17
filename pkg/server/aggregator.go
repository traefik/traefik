package server

import (
	"slices"
	"strings"

	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/server/provider"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

func mergeConfiguration(configurations dynamic.Configurations, defaultEntryPoints []string) dynamic.Configuration {
	// TODO: see if we can use DeepCopies inside, so that the given argument is left
	// untouched, and the modified copy is returned.
	conf := dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:           make(map[string]*dynamic.Router),
			Middlewares:       make(map[string]*dynamic.Middleware),
			Services:          make(map[string]*dynamic.Service),
			Models:            make(map[string]*dynamic.Model),
			ServersTransports: make(map[string]*dynamic.ServersTransport),
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:           make(map[string]*dynamic.TCPRouter),
			Services:          make(map[string]*dynamic.TCPService),
			Middlewares:       make(map[string]*dynamic.TCPMiddleware),
			Models:            make(map[string]*dynamic.TCPModel),
			ServersTransports: make(map[string]*dynamic.TCPServersTransport),
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
					log.Debug().
						Str(logs.RouterName, routerName).
						Strs(logs.EntryPointName, defaultEntryPoints).
						Msg("No entryPoint defined for this router, using the default one(s) instead")
					router.EntryPoints = defaultEntryPoints
				}

				// The `ruleSyntax` option is deprecated.
				// We exclude the "default" value to avoid logging it,
				// as it is the value used for internal models and computed rules.
				if router.RuleSyntax != "" && router.RuleSyntax != "default" {
					log.Warn().
						Str(logs.RouterName, routerName).
						Msg("Router's `ruleSyntax` option is deprecated, please remove any usage of this option.")
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
			for serversTransportName, serversTransport := range configuration.HTTP.ServersTransports {
				conf.HTTP.ServersTransports[provider.MakeQualifiedName(pvd, serversTransportName)] = serversTransport
			}
		}

		if configuration.TCP != nil {
			for routerName, router := range configuration.TCP.Routers {
				if len(router.EntryPoints) == 0 {
					log.Debug().
						Str(logs.RouterName, routerName).
						Msgf("No entryPoint defined for this TCP router, using the default one(s) instead: %+v", defaultEntryPoints)
					router.EntryPoints = defaultEntryPoints
				}
				conf.TCP.Routers[provider.MakeQualifiedName(pvd, routerName)] = router
			}
			for middlewareName, middleware := range configuration.TCP.Middlewares {
				conf.TCP.Middlewares[provider.MakeQualifiedName(pvd, middlewareName)] = middleware
			}
			for serviceName, service := range configuration.TCP.Services {
				conf.TCP.Services[provider.MakeQualifiedName(pvd, serviceName)] = service
			}
			for modelName, model := range configuration.TCP.Models {
				conf.TCP.Models[provider.MakeQualifiedName(pvd, modelName)] = model
			}
			for serversTransportName, serversTransport := range configuration.TCP.ServersTransports {
				conf.TCP.ServersTransports[provider.MakeQualifiedName(pvd, serversTransportName)] = serversTransport
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
			for _, cert := range configuration.TLS.Certificates {
				if slices.Contains(cert.Stores, tlsalpn01.ACMETLS1Protocol) && pvd != "tlsalpn.acme" {
					continue
				}

				conf.TLS.Certificates = append(conf.TLS.Certificates, cert)
			}

			for key, store := range configuration.TLS.Stores {
				if key != tls.DefaultTLSStoreName {
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
		log.Error().Msgf("Default TLS Store defined in multiple providers: %v", defaultTLSStoreProviders)
		delete(conf.TLS.Stores, tls.DefaultTLSStoreName)
	}

	if len(defaultTLSOptionProviders) == 0 {
		conf.TLS.Options[tls.DefaultTLSConfigName] = tls.DefaultTLSOptions
	} else if len(defaultTLSOptionProviders) > 1 {
		log.Error().Msgf("Default TLS Options defined in multiple providers %v", defaultTLSOptionProviders)
		// We do not set an empty tls.TLS{} as above so that we actually get a "cascading failure" later on,
		// i.e. routers depending on this missing TLS option will fail to initialize as well.
		delete(conf.TLS.Options, tls.DefaultTLSConfigName)
	}

	return conf
}

func applyModel(cfg dynamic.Configuration) dynamic.Configuration {
	if cfg.HTTP != nil && len(cfg.HTTP.Models) > 0 {
		rts := make(map[string]*dynamic.Router)

		for name, rt := range cfg.HTTP.Routers {
			router := rt.DeepCopy()

			if !router.DefaultRule && router.RuleSyntax == "" {
				for modelName, model := range cfg.HTTP.Models {
					// models cannot be provided by another provider than the internal one.
					if !strings.HasSuffix(modelName, "@internal") {
						continue
					}
					router.RuleSyntax = model.DefaultRuleSyntax
					break
				}
			}

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

					if cp.Observability == nil {
						cp.Observability = &dynamic.RouterObservabilityConfig{}
					}

					if cp.Observability.AccessLogs == nil {
						cp.Observability.AccessLogs = m.Observability.AccessLogs
					}

					if cp.Observability.Metrics == nil {
						cp.Observability.Metrics = m.Observability.Metrics
					}

					if cp.Observability.Tracing == nil {
						cp.Observability.Tracing = m.Observability.Tracing
					}

					if cp.Observability.TraceVerbosity == "" {
						cp.Observability.TraceVerbosity = m.Observability.TraceVerbosity
					}

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
	}

	// Apply the default observability model to HTTP routers.
	applyDefaultObservabilityModel(cfg)

	if cfg.TCP == nil || len(cfg.TCP.Models) == 0 {
		return cfg
	}

	tcpRouters := make(map[string]*dynamic.TCPRouter)

	for name, rt := range cfg.TCP.Routers {
		router := rt.DeepCopy()

		if router.RuleSyntax == "" {
			for _, model := range cfg.TCP.Models {
				router.RuleSyntax = model.DefaultRuleSyntax
				break
			}
		}

		tcpRouters[name] = router
	}

	cfg.TCP.Routers = tcpRouters

	return cfg
}

// applyDefaultObservabilityModel applies the default observability model to the configuration.
// This function is used to ensure that the observability configuration is set for all routers,
// and make sure it is serialized and available in the API.
// We could have introduced a "default" model, but it would have been more complex to manage for now.
// This could be generalized in the future.
// TODO: check if we can remove this and rely on the SetDefaults instead.
func applyDefaultObservabilityModel(cfg dynamic.Configuration) {
	if cfg.HTTP != nil {
		for _, router := range cfg.HTTP.Routers {
			if router.Observability == nil {
				router.Observability = &dynamic.RouterObservabilityConfig{
					AccessLogs:     pointer(true),
					Metrics:        pointer(true),
					Tracing:        pointer(true),
					TraceVerbosity: types.MinimalVerbosity,
				}

				continue
			}

			if router.Observability.AccessLogs == nil {
				router.Observability.AccessLogs = pointer(true)
			}

			if router.Observability.Metrics == nil {
				router.Observability.Metrics = pointer(true)
			}

			if router.Observability.Tracing == nil {
				router.Observability.Tracing = pointer(true)
			}

			if router.Observability.TraceVerbosity == "" {
				router.Observability.TraceVerbosity = types.MinimalVerbosity
			}
		}
	}
}

func pointer[T any](v T) *T { return &v }
