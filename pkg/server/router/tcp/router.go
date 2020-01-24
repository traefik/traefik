package tcp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/rules"
	"github.com/containous/traefik/v2/pkg/server/provider"
	tcpservice "github.com/containous/traefik/v2/pkg/server/service/tcp"
	"github.com/containous/traefik/v2/pkg/tcp"
	traefiktls "github.com/containous/traefik/v2/pkg/tls"
)

const (
	defaultTLSConfigName = "default"
	defaultTLSStoreName  = "default"
)

// NewManager Creates a new Manager
func NewManager(conf *runtime.Configuration,
	serviceManager *tcpservice.Manager,
	httpHandlers map[string]http.Handler,
	httpsHandlers map[string]http.Handler,
	tlsManager *traefiktls.Manager,
) *Manager {
	return &Manager{
		serviceManager: serviceManager,
		httpHandlers:   httpHandlers,
		httpsHandlers:  httpsHandlers,
		tlsManager:     tlsManager,
		conf:           conf,
	}
}

// Manager is a route/router manager
type Manager struct {
	serviceManager *tcpservice.Manager
	httpHandlers   map[string]http.Handler
	httpsHandlers  map[string]http.Handler
	tlsManager     *traefiktls.Manager
	conf           *runtime.Configuration
}

func (m *Manager) getTCPRouters(ctx context.Context, entryPoints []string) map[string]map[string]*runtime.TCPRouterInfo {
	if m.conf != nil {
		return m.conf.GetTCPRoutersByEntryPoints(ctx, entryPoints)
	}

	return make(map[string]map[string]*runtime.TCPRouterInfo)
}

func (m *Manager) getHTTPRouters(ctx context.Context, entryPoints []string, tls bool) map[string]map[string]*runtime.RouterInfo {
	if m.conf != nil {
		return m.conf.GetRoutersByEntryPoints(ctx, entryPoints, tls)
	}

	return make(map[string]map[string]*runtime.RouterInfo)
}

// BuildHandlers builds the handlers for the given entrypoints
func (m *Manager) BuildHandlers(rootCtx context.Context, entryPoints []string) map[string]*tcp.Router {
	entryPointsRouters := m.getTCPRouters(rootCtx, entryPoints)
	entryPointsRoutersHTTP := m.getHTTPRouters(rootCtx, entryPoints, true)

	entryPointHandlers := make(map[string]*tcp.Router)
	for _, entryPointName := range entryPoints {
		entryPointName := entryPointName

		routers := entryPointsRouters[entryPointName]

		ctx := log.With(rootCtx, log.Str(log.EntryPointName, entryPointName))

		handler, err := m.buildEntryPointHandler(ctx, routers, entryPointsRoutersHTTP[entryPointName], m.httpHandlers[entryPointName], m.httpsHandlers[entryPointName])
		if err != nil {
			log.FromContext(ctx).Error(err)
			continue
		}
		entryPointHandlers[entryPointName] = handler
	}
	return entryPointHandlers
}

type nameAndConfig struct {
	routerName string // just so we have it as additional information when logging
	TLSConfig  *tls.Config
}

func (m *Manager) buildEntryPointHandler(ctx context.Context, configs map[string]*runtime.TCPRouterInfo, configsHTTP map[string]*runtime.RouterInfo, handlerHTTP http.Handler, handlerHTTPS http.Handler) (*tcp.Router, error) {
	router := &tcp.Router{}
	router.HTTPHandler(handlerHTTP)

	defaultTLSConf, err := m.tlsManager.Get(defaultTLSStoreName, defaultTLSConfigName)
	if err != nil {
		log.FromContext(ctx).Errorf("Error during the build of the default TLS configuration: %v", err)
	}

	router.HTTPSHandler(handlerHTTPS, defaultTLSConf)

	if len(configsHTTP) > 0 {
		router.AddRouteHTTPTLS("*", defaultTLSConf)
	}

	// Keyed by domain, then by options reference.
	tlsOptionsForHostSNI := map[string]map[string]nameAndConfig{}
	for routerHTTPName, routerHTTPConfig := range configsHTTP {
		if len(routerHTTPConfig.TLS.Options) == 0 || routerHTTPConfig.TLS.Options == defaultTLSConfigName {
			continue
		}

		ctxRouter := log.With(provider.AddInContext(ctx, routerHTTPName), log.Str(log.RouterName, routerHTTPName))
		logger := log.FromContext(ctxRouter)

		domains, err := rules.ParseDomains(routerHTTPConfig.Rule)
		if err != nil {
			routerErr := fmt.Errorf("invalid rule %s, error: %v", routerHTTPConfig.Rule, err)
			routerHTTPConfig.AddError(routerErr, true)
			logger.Debug(routerErr)
			continue
		}

		if len(domains) == 0 {
			logger.Warnf("No domain found in rule %v, the TLS options applied for this router will depend on the hostSNI of each request", routerHTTPConfig.Rule)
		}

		for _, domain := range domains {
			if routerHTTPConfig.TLS != nil {
				tlsOptionsName := routerHTTPConfig.TLS.Options
				if tlsOptionsName != defaultTLSConfigName {
					tlsOptionsName = provider.GetQualifiedName(ctxRouter, routerHTTPConfig.TLS.Options)
				}

				tlsConf, err := m.tlsManager.Get(defaultTLSStoreName, tlsOptionsName)
				if err != nil {
					routerHTTPConfig.AddError(err, true)
					logger.Debug(err)
					continue
				}

				if tlsOptionsForHostSNI[domain] == nil {
					tlsOptionsForHostSNI[domain] = make(map[string]nameAndConfig)
				}
				tlsOptionsForHostSNI[domain][routerHTTPConfig.TLS.Options] = nameAndConfig{
					routerName: routerHTTPName,
					TLSConfig:  tlsConf,
				}
			}
		}
	}

	logger := log.FromContext(ctx)
	for hostSNI, tlsConfigs := range tlsOptionsForHostSNI {
		if len(tlsConfigs) == 1 {
			var optionsName string
			var config *tls.Config
			for k, v := range tlsConfigs {
				optionsName = k
				config = v.TLSConfig
				break
			}

			logger.Debugf("Adding route for %s with TLS options %s", hostSNI, optionsName)

			router.AddRouteHTTPTLS(hostSNI, config)
		} else {
			routers := make([]string, 0, len(tlsConfigs))
			for _, v := range tlsConfigs {
				configsHTTP[v.routerName].AddError(fmt.Errorf("found different TLS options for routers on the same host %v, so using the default TLS options instead", hostSNI), false)
				routers = append(routers, v.routerName)
			}

			logger.Warnf("Found different TLS options for routers on the same host %v, so using the default TLS options instead for these routers: %#v", hostSNI, routers)

			router.AddRouteHTTPTLS(hostSNI, defaultTLSConf)
		}
	}

	for routerName, routerConfig := range configs {
		ctxRouter := log.With(provider.AddInContext(ctx, routerName), log.Str(log.RouterName, routerName))
		logger := log.FromContext(ctxRouter)

		if routerConfig.Service == "" {
			err := errors.New("the service is missing on the router")
			routerConfig.AddError(err, true)
			logger.Error(err)
			continue
		}

		if routerConfig.Rule == "" {
			err := errors.New("router has no rule")
			routerConfig.AddError(err, true)
			logger.Error(err)
			continue
		}

		handler, err := m.serviceManager.BuildTCP(ctxRouter, routerConfig.Service)
		if err != nil {
			routerConfig.AddError(err, true)
			logger.Error(err)
			continue
		}

		domains, err := rules.ParseHostSNI(routerConfig.Rule)
		if err != nil {
			routerErr := fmt.Errorf("unknown rule %s", routerConfig.Rule)
			routerConfig.AddError(routerErr, true)
			logger.Error(routerErr)
			continue
		}

		for _, domain := range domains {
			logger.Debugf("Adding route %s on TCP", domain)
			switch {
			case routerConfig.TLS != nil:
				if routerConfig.TLS.Passthrough {
					router.AddRoute(domain, handler)
				} else {
					tlsOptionsName := routerConfig.TLS.Options

					if len(tlsOptionsName) == 0 {
						tlsOptionsName = defaultTLSConfigName
					}

					if tlsOptionsName != defaultTLSConfigName {
						tlsOptionsName = provider.GetQualifiedName(ctxRouter, tlsOptionsName)
					}

					tlsConf, err := m.tlsManager.Get(defaultTLSStoreName, tlsOptionsName)
					if err != nil {
						routerConfig.AddError(err, true)
						logger.Debug(err)
						continue
					}

					router.AddRouteTLS(domain, handler, tlsConf)
				}
			case domain == "*":
				router.AddCatchAllNoTLS(handler)
			default:
				logger.Warn("TCP Router ignored, cannot specify a Host rule without TLS")
			}
		}
	}

	return router, nil
}
