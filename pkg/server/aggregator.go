package server

import (
	"context"
	"fmt"
	"slices"

	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	httpmuxer "github.com/traefik/traefik/v2/pkg/muxer/http"
	"github.com/traefik/traefik/v2/pkg/server/provider"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
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
			Routers:     make(map[string]*dynamic.TCPRouter),
			Services:    make(map[string]*dynamic.TCPService),
			Middlewares: make(map[string]*dynamic.TCPMiddleware),
		},
		UDP: &dynamic.UDPConfiguration{
			Routers:  make(map[string]*dynamic.UDPRouter),
			Services: make(map[string]*dynamic.UDPService),
		},
		TLS: &dynamic.TLSConfiguration{
			Stores:  make(map[string]traefiktls.Store),
			Options: make(map[string]traefiktls.Options),
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
			for serversTransportName, serversTransport := range configuration.HTTP.ServersTransports {
				conf.HTTP.ServersTransports[provider.MakeQualifiedName(pvd, serversTransportName)] = serversTransport
			}
		}

		if configuration.TCP != nil {
			for routerName, router := range configuration.TCP.Routers {
				if len(router.EntryPoints) == 0 {
					log.WithoutContext().
						WithField(log.RouterName, routerName).
						Debugf("No entryPoint defined for this TCP router, using the default one(s) instead: %+v", defaultEntryPoints)
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
				if key != traefiktls.DefaultTLSStoreName {
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
		log.WithoutContext().Errorf("Default TLS Store defined in multiple providers: %v", defaultTLSStoreProviders)
		delete(conf.TLS.Stores, traefiktls.DefaultTLSStoreName)
	}

	if len(defaultTLSOptionProviders) == 0 {
		conf.TLS.Options[traefiktls.DefaultTLSConfigName] = traefiktls.DefaultTLSOptions
	} else if len(defaultTLSOptionProviders) > 1 {
		log.WithoutContext().Errorf("Default TLS Options defined in multiple providers %v", defaultTLSOptionProviders)
		// We do not set an empty tls.TLS{} as above so that we actually get a "cascading failure" later on,
		// i.e. routers depending on this missing TLS option will fail to initialize as well.
		delete(conf.TLS.Options, traefiktls.DefaultTLSConfigName)
	}

	return resolveTLSOptions(conf)
}

func resolveTLSOptions(cfg dynamic.Configuration) dynamic.Configuration {
	if cfg.HTTP == nil || len(cfg.HTTP.Routers) == 0 {
		return cfg
	}

	ctx := context.Background()
	rts := make(map[string]*dynamic.Router)

	// Keyed by domain, then by options reference.
	// The actual source of truth for what TLS options will actually be used for the connection.
	// As opposed to tlsOptionsForHost, it keeps track of all the (different) TLS
	// options that occur for a given host name, so that later on we can set relevant
	// errors and logging for all the routers concerned (i.e. wrongly configured).
	tlsOptionsForHostSNI := map[string]map[string][]string{}

	for routerHTTPName, routerHTTPConfig := range cfg.HTTP.Routers {
		rts[routerHTTPName] = routerHTTPConfig.DeepCopy()

		if routerHTTPConfig.TLS == nil {
			continue
		}

		ctxRouter := log.With(provider.AddInContext(ctx, routerHTTPName), log.Str(log.RouterName, routerHTTPName))
		logger := log.FromContext(ctxRouter)

		tlsOptionsName := traefiktls.DefaultTLSConfigName
		if len(routerHTTPConfig.TLS.Options) > 0 && routerHTTPConfig.TLS.Options != traefiktls.DefaultTLSConfigName {
			tlsOptionsName = provider.GetQualifiedName(ctxRouter, routerHTTPConfig.TLS.Options)
		}

		domains, err := httpmuxer.ParseDomains(routerHTTPConfig.Rule)
		if err != nil {
			routerErr := fmt.Errorf("invalid rule %s, error: %w", routerHTTPConfig.Rule, err)
			logger.Error(routerErr)
			continue
		}

		if len(domains) == 0 {
			rts[routerHTTPName].TLS.ResolvedOptions = "default"
			logger.Warnf("No domain found in rule %v, the TLS options applied for this router will depend on the SNI of each request", routerHTTPConfig.Rule)
		}

		for _, domain := range domains {
			// domain is already in lower case thanks to the domain parsing
			if tlsOptionsForHostSNI[domain] == nil {
				tlsOptionsForHostSNI[domain] = make(map[string][]string)
			}
			tlsOptionsForHostSNI[domain][tlsOptionsName] = append(tlsOptionsForHostSNI[domain][tlsOptionsName], routerHTTPName)
		}
	}

	for hostSNI, tlsConfigs := range tlsOptionsForHostSNI {
		if len(tlsConfigs) == 1 {
			var optionsName string
			for k, v := range tlsConfigs {
				optionsName = k
				log.WithoutContext().Debugf("Adding route for %s with TLS options %s", hostSNI, optionsName)
				for _, s := range v {
					rts[s].TLS.ResolvedOptions = optionsName
				}
			}
			continue
		}

		// multiple tlsConfigs
		routers := make([]string, 0, len(tlsConfigs))
		for _, v := range tlsConfigs {
			for _, s := range v {
				rts[s].TLS.ResolvedOptions = traefiktls.DefaultTLSConfigName
				routers = append(routers, s)
			}
		}

		log.WithoutContext().Warnf("Found different TLS options for routers on the same host %v, so using the default TLS options instead for these routers: %#v", hostSNI, routers)
	}

	cfg.HTTP.Routers = rts
	return cfg
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

				if m.DeniedEncodedPathCharacters != nil {
					// As the denied encoded path characters option is not configurable at the router level,
					// we can simply copy the whole structure to override the router's default config.
					cp.DeniedEncodedPathCharacters = m.DeniedEncodedPathCharacters
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

	return cfg
}
