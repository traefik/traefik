package udp

import (
	"context"
	"errors"
	"slices"
	"sort"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/middlewares/accessgovernance"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/server/provider"
	udpservice "github.com/traefik/traefik/v3/pkg/server/service/udp"
	"github.com/traefik/traefik/v3/pkg/udp"
)

// Manager is a route/router manager.
type Manager struct {
	serviceManager     *udpservice.Manager
	middlewaresBuilder middlewareBuilder
	conf               *runtime.Configuration
	defaultMiddlewares map[string][]string
}

type middlewareBuilder interface {
	BuildChain(ctx context.Context, middlewares []string) *udp.Chain
}

// NewManager Creates a new Manager.
func NewManager(conf *runtime.Configuration,
	serviceManager *udpservice.Manager,
	middlewaresBuilders ...middlewareBuilder,
) *Manager {
	var middlewaresBuilder middlewareBuilder
	if len(middlewaresBuilders) > 0 {
		middlewaresBuilder = middlewaresBuilders[0]
	}
	return &Manager{
		serviceManager:     serviceManager,
		middlewaresBuilder: middlewaresBuilder,
		conf:               conf,
	}
}

// SetDefaultMiddlewares configures entry-point middlewares applied to every UDP router.
func (m *Manager) SetDefaultMiddlewares(middlewares map[string][]string) {
	m.defaultMiddlewares = middlewares
}

// BuildHandlers builds the handlers for the given entrypoints.
func (m *Manager) BuildHandlers(rootCtx context.Context, entryPoints []string) map[string]udp.Handler {
	entryPointsRouters := m.getUDPRouters(rootCtx, entryPoints)

	entryPointHandlers := make(map[string]udp.Handler)
	for _, entryPointName := range entryPoints {
		routers := entryPointsRouters[entryPointName]

		logger := log.Ctx(rootCtx).With().Str(logs.EntryPointName, entryPointName).Logger()
		ctx := accessgovernance.WithRouteMetadata(logger.WithContext(rootCtx), entryPointName, "", "")

		if len(routers) > 1 {
			logger.Warn().Msg("Config has more than one udp router for a given entrypoint.")
		}

		handlers := m.buildEntryPointHandlers(ctx, entryPointName, routers)

		if len(handlers) > 0 {
			// As UDP support only one router per entrypoint, we only take the first one.
			entryPointHandlers[entryPointName] = handlers[0]
		}
	}
	return entryPointHandlers
}

func (m *Manager) getUDPRouters(ctx context.Context, entryPoints []string) map[string]map[string]*runtime.UDPRouterInfo {
	if m.conf != nil {
		return m.conf.GetUDPRoutersByEntryPoints(ctx, entryPoints)
	}

	return make(map[string]map[string]*runtime.UDPRouterInfo)
}

func (m *Manager) buildEntryPointHandlers(ctx context.Context, entryPointName string, configs map[string]*runtime.UDPRouterInfo) []udp.Handler {
	var rtNames []string
	for routerName := range configs {
		rtNames = append(rtNames, routerName)
	}

	sort.Slice(rtNames, func(i, j int) bool {
		return rtNames[i] > rtNames[j]
	})

	var handlers []udp.Handler

	for _, routerName := range rtNames {
		routerConfig := configs[routerName]
		logger := log.Ctx(ctx).With().Str(logs.RouterName, routerName).Logger()
		ctxRouter := logger.WithContext(provider.AddInContext(ctx, routerName))
		ctxRouter = accessgovernance.WithRouteMetadata(ctxRouter, entryPointName, routerName, routerConfig.Service)

		if routerConfig.Service == "" {
			err := errors.New("the service is missing on the udp router")
			routerConfig.AddError(err, true)
			logger.Error().Err(err).Send()
			continue
		}

		handler, err := m.serviceManager.BuildUDP(ctxRouter, routerConfig.Service)
		if err != nil {
			routerConfig.AddError(err, true)
			logger.Error().Err(err).Send()
			continue
		}

		var qualifiedNames []string
		middlewareNames := append(slices.Clone(m.defaultMiddlewares[entryPointName]), routerConfig.Middlewares...)
		for _, name := range middlewareNames {
			qualifiedNames = append(qualifiedNames, provider.GetQualifiedName(ctxRouter, name))
		}
		if len(qualifiedNames) > 0 {
			if m.middlewaresBuilder == nil {
				err := errors.New("UDP middleware builder is not configured")
				routerConfig.AddError(err, true)
				logger.Error().Err(err).Send()
				continue
			}
			middlewareChain := m.middlewaresBuilder.BuildChain(ctxRouter, qualifiedNames)
			handler, err = udp.NewChain().Extend(*middlewareChain).Then(handler)
			if err != nil {
				routerConfig.AddError(err, true)
				logger.Error().Err(err).Send()
				continue
			}
		}

		handlers = append(handlers, handler)
	}

	return handlers
}
