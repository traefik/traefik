package router

import (
	"context"
	"fmt"
	"net/http"

	"github.com/containous/alice"
	"github.com/containous/mux"
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares/accesslog"
	"github.com/containous/traefik/middlewares/recovery"
	"github.com/containous/traefik/middlewares/tracing"
	"github.com/containous/traefik/responsemodifiers"
	"github.com/containous/traefik/server/internal"
	"github.com/containous/traefik/server/middleware"
	"github.com/containous/traefik/server/service"
)

const (
	recoveryMiddlewareName = "traefik-internal-recovery"
)

// NewManager Creates a new Manager
func NewManager(routers map[string]*config.Router,
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
	configs            map[string]*config.Router
	serviceManager     *service.Manager
	middlewaresBuilder *middleware.Builder
	modifierBuilder    *responsemodifiers.Builder
}

// BuildHandlers Builds handler for all entry points
func (m *Manager) BuildHandlers(rootCtx context.Context, entryPoints []string) map[string]http.Handler {
	entryPointsRouters := m.filteredRouters(rootCtx, entryPoints)

	entryPointHandlers := make(map[string]http.Handler)
	for entryPointName, routers := range entryPointsRouters {
		ctx := log.With(rootCtx, log.Str(log.EntryPointName, entryPointName))

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

func (m *Manager) filteredRouters(ctx context.Context, entryPoints []string) map[string]map[string]*config.Router {
	entryPointsRouters := make(map[string]map[string]*config.Router)

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
				entryPointsRouters[entryPointName] = make(map[string]*config.Router)
			}

			entryPointsRouters[entryPointName][rtName] = rt
		}
	}

	return entryPointsRouters
}

func (m *Manager) buildEntryPointHandler(ctx context.Context, configs map[string]*config.Router) (http.Handler, error) {
	router := mux.NewRouter().
		SkipClean(true)

	for routerName, routerConfig := range configs {
		ctx := log.With(ctx, log.Str(log.RouterName, routerName))
		logger := log.FromContext(ctx)

		ctx = internal.AddProviderInContext(ctx, routerName)

		handler, err := m.buildRouterHandler(ctx, routerName)
		if err != nil {
			logger.Error(err)
			continue
		}

		err = addRoute(ctx, router, routerConfig.Rule, routerConfig.Priority, handler)
		if err != nil {
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

	handler, err := m.buildHandler(ctx, configRouter, routerName)
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

func (m *Manager) buildHandler(ctx context.Context, router *config.Router, routerName string) (http.Handler, error) {
	rm := m.modifierBuilder.Build(ctx, router.Middlewares)

	sHandler, err := m.serviceManager.Build(ctx, router.Service, rm)
	if err != nil {
		return nil, err
	}

	mHandler := m.middlewaresBuilder.BuildChain(ctx, router.Middlewares)

	tHandler := func(next http.Handler) (http.Handler, error) {
		return tracing.NewForwarder(ctx, routerName, router.Service, next), nil
	}

	return alice.New().Extend(*mHandler).Append(tHandler).Then(sHandler)
}
