package tcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/rules"
	"github.com/containous/traefik/pkg/server/internal"
	tcpservice "github.com/containous/traefik/pkg/server/service/tcp"
	"github.com/containous/traefik/pkg/tcp"
	"github.com/containous/traefik/pkg/tls"
)

// NewManager Creates a new Manager
func NewManager(conf *config.RuntimeConfiguration,
	serviceManager *tcpservice.Manager,
	httpHandlers map[string]http.Handler,
	httpsHandlers map[string]http.Handler,
	tlsManager *tls.Manager,
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
	tlsManager     *tls.Manager
	conf           *config.RuntimeConfiguration
}

func (m *Manager) getTCPRouters(ctx context.Context, entryPoints []string) map[string]map[string]*config.TCPRouterInfo {
	if m.conf != nil {
		return m.conf.GetTCPRoutersByEntrypoints(ctx, entryPoints)
	}

	return make(map[string]map[string]*config.TCPRouterInfo)
}

func (m *Manager) getHTTPRouters(ctx context.Context, entryPoints []string, tls bool) map[string]map[string]*config.RouterInfo {
	if m.conf != nil {
		return m.conf.GetRoutersByEntrypoints(ctx, entryPoints, tls)
	}

	return make(map[string]map[string]*config.RouterInfo)
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

func (m *Manager) buildEntryPointHandler(ctx context.Context, configs map[string]*config.TCPRouterInfo, configsHTTP map[string]*config.RouterInfo, handlerHTTP http.Handler, handlerHTTPS http.Handler) (*tcp.Router, error) {
	router := &tcp.Router{}
	router.HTTPHandler(handlerHTTP)

	defaultTLSConf, err := m.tlsManager.Get("default", "default")
	if err != nil {
		return nil, err
	}

	router.HTTPSHandler(handlerHTTPS, defaultTLSConf)

	for routerHTTPName, routerHTTPConfig := range configsHTTP {
		if len(routerHTTPConfig.TLS.Options) == 0 || routerHTTPConfig.TLS.Options == "default" {
			continue
		}

		ctxRouter := log.With(internal.AddProviderInContext(ctx, routerHTTPName), log.Str(log.RouterName, routerHTTPName))
		logger := log.FromContext(ctxRouter)

		domains, err := rules.ParseDomains(routerHTTPConfig.Rule)
		if err != nil {
			routerErr := fmt.Errorf("invalid rule %s, error: %v", routerHTTPConfig.Rule, err)
			routerHTTPConfig.Err = routerErr.Error()
			logger.Debug(routerErr)
			continue
		}

		if len(domains) == 0 {
			logger.Warnf("The 'default' TLS options will be applied instead of %q as no domain has been found in the rule", routerHTTPConfig.TLS.Options)
		}

		for _, domain := range domains {
			if routerHTTPConfig.TLS != nil {
				tlsConf, err := m.tlsManager.Get("default", routerHTTPConfig.TLS.Options)
				if err != nil {
					routerHTTPConfig.Err = err.Error()
					logger.Debug(err)
					continue
				}

				router.AddRouteHTTPTLS(domain, tlsConf)
			}
		}
	}

	for routerName, routerConfig := range configs {
		ctxRouter := log.With(internal.AddProviderInContext(ctx, routerName), log.Str(log.RouterName, routerName))
		logger := log.FromContext(ctxRouter)

		handler, err := m.serviceManager.BuildTCP(ctxRouter, routerConfig.Service)
		if err != nil {
			routerConfig.Err = err.Error()
			logger.Error(err)
			continue
		}

		domains, err := rules.ParseHostSNI(routerConfig.Rule)
		if err != nil {
			routerErr := fmt.Errorf("unknown rule %s", routerConfig.Rule)
			routerConfig.Err = routerErr.Error()
			logger.Debug(routerErr)
			continue
		}

		for _, domain := range domains {
			logger.Debugf("Adding route %s on TCP", domain)
			switch {
			case routerConfig.TLS != nil:
				if routerConfig.TLS.Passthrough {
					router.AddRoute(domain, handler)
				} else {
					configName := "default"
					if len(routerConfig.TLS.Options) > 0 {
						configName = routerConfig.TLS.Options
					}

					tlsConf, err := m.tlsManager.Get("default", configName)
					if err != nil {
						routerConfig.Err = err.Error()
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
