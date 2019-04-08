package router

import (
	"context"
	"fmt"
	"net/http"

	"github.com/containous/alice"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/middlewares/accesslog"
	"github.com/containous/traefik/pkg/middlewares/recovery"
	"github.com/containous/traefik/pkg/middlewares/tracing"
	"github.com/containous/traefik/pkg/responsemodifiers"
	"github.com/containous/traefik/pkg/rules"
	"github.com/containous/traefik/pkg/server/internal"
	"github.com/containous/traefik/pkg/server/middleware"
	"github.com/containous/traefik/pkg/server/service"
)

const (
	recoveryMiddlewareName = "traefik-internal-recovery"
)

// NewManager Creates a new Manager
func NewManager(routers map[string]*config.RouterInfo,
	serviceManager *service.Manager, middlewaresBuilder *middleware.Builder, modifierBuilder *responsemodifiers.Builder,
) *Manager {
	return &Manager{
		routerHandlers:     make(map[string]http.Handler),
		configs:            routers,
		serviceManager:     serviceManager,
		middlewaresBuilder: middlewaresBuilder,
		modifierBuilder:    modifierBuilder,
	}
}

// Manager A route/router manager
type Manager struct {
	routerHandlers     map[string]http.Handler
	configs            map[string]*config.RouterInfo
	serviceManager     *service.Manager
	middlewaresBuilder *middleware.Builder
	modifierBuilder    *responsemodifiers.Builder
}

// GetRuntimeConfiguration returns the configuration of all the current HTTP routers.
func (m Manager) getRuntimeConfiguration() map[string]*config.RouterInfo {
	return m.configs
}

// BuildHandlers Builds handler for all entry points
func (m *Manager) BuildHandlers(rootCtx context.Context, entryPoints []string, tls bool) map[string]http.Handler {
	entryPointsRouters := m.filteredRouters(rootCtx, entryPoints, tls)

	entryPointHandlers := make(map[string]http.Handler)
	for entryPointName, routers := range entryPointsRouters {
		entryPointName := entryPointName
		ctx := log.With(rootCtx, log.Str(log.EntryPointName, entryPointName))

		// TODO: set routers.usedby
		handler, err := m.buildEntryPointHandler(ctx, routers)
		if err != nil {
			log.FromContext(ctx).Error(err)
			continue
		}

		handlerWithAccessLog, err := alice.New(func(next http.Handler) (http.Handler, error) {
			return accesslog.NewFieldHandler(next, log.EntryPointName, entryPointName, accesslog.AddOriginFields), nil
		}).Then(handler)
		if err != nil {
			log.FromContext(ctx).Error(err)
			entryPointHandlers[entryPointName] = handler
		} else {
			entryPointHandlers[entryPointName] = handlerWithAccessLog
		}
	}

	m.serviceManager.LaunchHealthCheck()

	return entryPointHandlers
}

func contains(entryPoints []string, entryPointName string) bool {
	for _, name := range entryPoints {
		if name == entryPointName {
			return true
		}
	}
	return false
}

func (m *Manager) filteredRouters(ctx context.Context, entryPoints []string, tls bool) map[string]map[string]*config.RouterInfo {
	entryPointsRouters := make(map[string]map[string]*config.RouterInfo)

	for rtName, rt := range m.configs {
		if (tls && rt.TLS == nil) || (!tls && rt.TLS != nil) {
			continue
		}

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
				entryPointsRouters[entryPointName] = make(map[string]*config.RouterInfo)
			}

			entryPointsRouters[entryPointName][rtName] = rt
		}
	}

	return entryPointsRouters
}

func (m *Manager) buildEntryPointHandler(ctx context.Context, configs map[string]*config.RouterInfo) (http.Handler, error) {
	router, err := rules.NewRouter()
	if err != nil {
		return nil, err
	}

	for routerName, routerConfig := range configs {
		ctxRouter := log.With(internal.AddProviderInContext(ctx, routerName), log.Str(log.RouterName, routerName))
		logger := log.FromContext(ctxRouter)

		handler, err := m.buildRouterHandler(ctxRouter, routerName)
		if err != nil {
			routerConfig.Err = err.Error()
			logger.Error(err)
			continue
		}

		err = router.AddRoute(routerConfig.Rule, routerConfig.Priority, handler)
		if err != nil {
			routerConfig.Err = err.Error()
			logger.Error(err)
			continue
		}
	}

	router.SortRoutes()

	chain := alice.New()
	chain = chain.Append(func(next http.Handler) (http.Handler, error) {
		return recovery.New(ctx, next, recoveryMiddlewareName)
	})

	return chain.Then(router)
}

func (m *Manager) buildRouterHandler(ctx context.Context, routerName string) (http.Handler, error) {
	if handler, ok := m.routerHandlers[routerName]; ok {
		return handler, nil
	}

	configRouter, ok := m.configs[routerName]
	if !ok {
		return nil, fmt.Errorf("no configuration for %s", routerName)
	}

	handler, err := m.buildHTTPHandler(ctx, configRouter, routerName)
	if err != nil {
		return nil, err
	}

	handlerWithAccessLog, err := alice.New(func(next http.Handler) (http.Handler, error) {
		return accesslog.NewFieldHandler(next, accesslog.RouterName, routerName, nil), nil
	}).Then(handler)
	if err != nil {
		log.FromContext(ctx).Error(err)
		m.routerHandlers[routerName] = handler
	} else {
		m.routerHandlers[routerName] = handlerWithAccessLog
	}

	return m.routerHandlers[routerName], nil
}

func (m *Manager) buildHTTPHandler(ctx context.Context, router *config.RouterInfo, routerName string) (http.Handler, error) {
	qualifiedNames := make([]string, len(router.Middlewares))
	for i, name := range router.Middlewares {
		qualifiedNames[i] = internal.GetQualifiedName(ctx, name)
	}
	rm := m.modifierBuilder.Build(ctx, qualifiedNames)

	sHandler, err := m.serviceManager.BuildHTTP(ctx, router.Service, rm)
	if err != nil {
		return nil, err
	}

	mHandler := m.middlewaresBuilder.BuildChain(ctx, router.Middlewares)

	tHandler := func(next http.Handler) (http.Handler, error) {
		return tracing.NewForwarder(ctx, routerName, router.Service, next), nil
	}

	return alice.New().Extend(*mHandler).Append(tHandler).Then(sHandler)
}
