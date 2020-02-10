package udp

import (
	"context"
	"errors"

	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/server/provider"
	udpservice "github.com/containous/traefik/v2/pkg/server/service/udp"
	"github.com/containous/traefik/v2/pkg/udp"
)

// NewManager Creates a new Manager
func NewManager(conf *runtime.Configuration,
	serviceManager *udpservice.Manager,
) *Manager {
	return &Manager{
		serviceManager: serviceManager,
		conf:           conf,
	}
}

// Manager is a route/router manager
type Manager struct {
	serviceManager *udpservice.Manager
	conf           *runtime.Configuration
}

func (m *Manager) getUDPRouters(ctx context.Context, entryPoints []string) map[string]map[string]*runtime.UDPRouterInfo {
	if m.conf != nil {
		return m.conf.GetUDPRoutersByEntryPoints(ctx, entryPoints)
	}

	return make(map[string]map[string]*runtime.UDPRouterInfo)
}

// BuildHandlers builds the handlers for the given entrypoints
func (m *Manager) BuildHandlers(rootCtx context.Context, entryPoints []string) map[string]udp.Handler {
	entryPointsRouters := m.getUDPRouters(rootCtx, entryPoints)

	entryPointHandlers := make(map[string]udp.Handler)
	for _, entryPointName := range entryPoints {
		entryPointName := entryPointName

		routers := entryPointsRouters[entryPointName]

		ctx := log.With(rootCtx, log.Str(log.EntryPointName, entryPointName))

		handler, err := m.buildEntryPointHandler(ctx, routers)
		if err != nil {
			log.FromContext(ctx).Error(err)
			continue
		}
		entryPointHandlers[entryPointName] = handler
	}
	return entryPointHandlers
}

func (m *Manager) buildEntryPointHandler(ctx context.Context, configs map[string]*runtime.UDPRouterInfo) (udp.Handler, error) {
	logger := log.FromContext(ctx)

	if len(configs) > 1 {
		logger.Warn("Warning: config has more than one udp router for a given entrypoint")
	}
	for routerName, routerConfig := range configs {
		ctxRouter := log.With(provider.AddInContext(ctx, routerName), log.Str(log.RouterName, routerName))
		logger := log.FromContext(ctxRouter)

		if routerConfig.Service == "" {
			err := errors.New("the service is missing on the udp router")
			routerConfig.AddError(err, true)
			logger.Error(err)
			continue
		}

		handler, err := m.serviceManager.BuildUDP(ctxRouter, routerConfig.Service)
		if err != nil {
			routerConfig.AddError(err, true)
			logger.Error(err)
			continue
		}
		return handler, nil
	}

	return nil, nil
}
