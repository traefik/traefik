package tcp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/rules"
	"github.com/containous/traefik/pkg/server/internal"
	tcpservice "github.com/containous/traefik/pkg/server/service/tcp"
	"github.com/containous/traefik/pkg/tcp"
)

// NewManager Creates a new Manager
func NewManager(conf *config.RuntimeConfiguration,
	serviceManager *tcpservice.Manager,
	httpHandlers map[string]http.Handler,
	httpsHandlers map[string]http.Handler,
	tlsConfig *tls.Config,
) *Manager {
	return &Manager{
		configs:        conf.TCPRouters,
		serviceManager: serviceManager,
		httpHandlers:   httpHandlers,
		httpsHandlers:  httpsHandlers,
		tlsConfig:      tlsConfig,
	}
}

// Manager is a route/router manager
type Manager struct {
	configs        map[string]*config.TCPRouterInfo
	serviceManager *tcpservice.Manager
	httpHandlers   map[string]http.Handler
	httpsHandlers  map[string]http.Handler
	tlsConfig      *tls.Config
}

// GetRuntimeConfiguration returns the configuration of all the current TCP routers.
func (m Manager) GetRuntimeConfiguration() map[string]*config.TCPRouterInfo {
	return m.configs
}

// BuildHandlers builds the handlers for the given entrypoints
func (m *Manager) BuildHandlers(rootCtx context.Context, entryPoints []string) map[string]*tcp.Router {
	entryPointsRouters := m.filteredRouters(rootCtx, entryPoints)

	entryPointHandlers := make(map[string]*tcp.Router)
	for _, entryPointName := range entryPoints {
		entryPointName := entryPointName

		routers := entryPointsRouters[entryPointName]

		ctx := log.With(rootCtx, log.Str(log.EntryPointName, entryPointName))

		handler, err := m.buildEntryPointHandler(ctx, routers, m.httpHandlers[entryPointName], m.httpsHandlers[entryPointName])
		if err != nil {
			log.FromContext(ctx).Error(err)
			continue
		}
		entryPointHandlers[entryPointName] = handler
	}
	return entryPointHandlers
}

func (m *Manager) buildEntryPointHandler(ctx context.Context, configs map[string]*config.TCPRouterInfo, handlerHTTP http.Handler, handlerHTTPS http.Handler) (*tcp.Router, error) {
	router := &tcp.Router{}
	router.HTTPHandler(handlerHTTP)
	router.HTTPSHandler(handlerHTTPS, m.tlsConfig)

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
					router.AddRouteTLS(domain, handler, m.tlsConfig)
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

func contains(entryPoints []string, entryPointName string) bool {
	for _, name := range entryPoints {
		if name == entryPointName {
			return true
		}
	}
	return false
}

func (m *Manager) filteredRouters(ctx context.Context, entryPoints []string) map[string]map[string]*config.TCPRouterInfo {
	entryPointsRouters := make(map[string]map[string]*config.TCPRouterInfo)

	for rtName, rt := range m.configs {
		eps := rt.EntryPoints
		if len(eps) == 0 {
			eps = entryPoints
		}

		for _, entryPointName := range eps {
			if !contains(entryPoints, entryPointName) {
				log.FromContext(log.With(ctx, log.Str(log.EntryPointName, entryPointName))).
					Errorf("entryPoint %q doesn't exist", entryPointName)
				continue
			}

			if _, ok := entryPointsRouters[entryPointName]; !ok {
				entryPointsRouters[entryPointName] = make(map[string]*config.TCPRouterInfo)
			}

			entryPointsRouters[entryPointName][rtName] = rt
		}
	}

	return entryPointsRouters
}
