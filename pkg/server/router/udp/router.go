package udp

import (
	"context"
	"errors"
	"net"
	"sort"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/server/provider"
	udpservice "github.com/traefik/traefik/v3/pkg/server/service/udp"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/udp"
)

// Manager is a route/router manager.
type Manager struct {
	serviceManager *udpservice.Manager
	conf           *runtime.Configuration
	tlsManager     *traefiktls.Manager
}

// NewManager Creates a new Manager.
func NewManager(
	conf *runtime.Configuration,
	serviceManager *udpservice.Manager,
	tlsManager *traefiktls.Manager,
) *Manager {
	return &Manager{
		serviceManager: serviceManager,
		conf:           conf,
		tlsManager:     tlsManager,
	}
}

// BuildHandlers builds the handlers for the given entrypoints.
func (m *Manager) BuildHandlers(rootCtx context.Context, entryPoints []string, localAddrs map[string]net.Addr) map[string]udp.Handler {
	entryPointsRouters := m.getUDPRouters(rootCtx, entryPoints)

	entryPointHandlers := make(map[string]udp.Handler)
	for _, entryPointName := range entryPoints {
		routers := entryPointsRouters[entryPointName]

		logger := log.Ctx(rootCtx).With().Str(logs.EntryPointName, entryPointName).Logger()
		ctx := logger.WithContext(rootCtx)

		if len(routers) > 1 {
			logger.Warn().Msg("Config has more than one udp router for a given entrypoint.")
		}

		handlers := m.buildEntryPointHandlers(ctx, routers, localAddrs[entryPointName])

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

func (m *Manager) buildEntryPointHandlers(ctx context.Context, configs map[string]*runtime.UDPRouterInfo, localAddr net.Addr) []udp.Handler {
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

		if routerConfig.TLS != nil {
			tlsOptionsName := routerConfig.TLS.Options
			if len(tlsOptionsName) == 0 {
				tlsOptionsName = traefiktls.DefaultTLSConfigName
			}

			if tlsOptionsName != traefiktls.DefaultTLSConfigName {
				tlsOptionsName = provider.GetQualifiedName(ctxRouter, tlsOptionsName)
			}

			tlsConf, err := m.tlsManager.Get(traefiktls.DefaultTLSStoreName, tlsOptionsName)
			if err != nil {
				routerConfig.AddError(err, true)
				logger.Error().Err(err).Send()
				continue
			}

			handler = &udp.DTLSHandler{
				Next:           handler,
				GetCertificate: tlsConf.GetCertificate,
				LocalAddr:      localAddr,
			}
		}

		handlers = append(handlers, handler)
	}

	return handlers
}
