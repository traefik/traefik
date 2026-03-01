package tcp

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"net"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/healthcheck"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
	"github.com/traefik/traefik/v3/pkg/server/provider"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

// Manager is the TCPHandlers factory.
type Manager struct {
	dialerManager  *tcp.DialerManager
	configs        map[string]*runtime.TCPServiceInfo
	rand           *rand.Rand // For the initial shuffling of load-balancers.
	healthCheckers map[string]*healthcheck.ServiceTCPHealthChecker
}

// NewManager creates a new manager.
func NewManager(conf *runtime.Configuration, dialerManager *tcp.DialerManager) *Manager {
	return &Manager{
		dialerManager:  dialerManager,
		healthCheckers: make(map[string]*healthcheck.ServiceTCPHealthChecker),
		configs:        conf.TCPServices,
		rand:           rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// BuildTCP Creates a tcp.Handler for a service configuration.
func (m *Manager) BuildTCP(rootCtx context.Context, serviceName string) (tcp.Handler, error) {
	serviceQualifiedName := provider.GetQualifiedName(rootCtx, serviceName)

	logger := log.Ctx(rootCtx).With().Str(logs.ServiceName, serviceQualifiedName).Logger()
	ctx := provider.AddInContext(rootCtx, serviceQualifiedName)

	conf, ok := m.configs[serviceQualifiedName]
	if !ok {
		return nil, fmt.Errorf("the service %q does not exist", serviceQualifiedName)
	}

	if conf.LoadBalancer != nil && conf.Weighted != nil {
		err := errors.New("cannot create service: multi-types service not supported, consider declaring two different pieces of service instead")
		conf.AddError(err, true)
		return nil, err
	}

	switch {
	case conf.LoadBalancer != nil:
		loadBalancer := tcp.NewWRRLoadBalancer(conf.LoadBalancer.HealthCheck != nil)

		if conf.LoadBalancer.TerminationDelay != nil {
			log.Ctx(ctx).Warn().Msgf("Service %q load balancer uses `TerminationDelay`, but this option is deprecated, please use ServersTransport configuration instead.", serviceName)
		}

		if conf.LoadBalancer.ProxyProtocol != nil {
			log.Ctx(ctx).Warn().Msgf("Service %q load balancer uses `ProxyProtocol`, but this option is deprecated, please use ServersTransport configuration instead.", serviceName)
		}

		if len(conf.LoadBalancer.ServersTransport) > 0 {
			conf.LoadBalancer.ServersTransport = provider.GetQualifiedName(ctx, conf.LoadBalancer.ServersTransport)
		}

		uniqHealthCheckTargets := make(map[string]healthcheck.TCPHealthCheckTarget, len(conf.LoadBalancer.Servers))

		for index, server := range shuffle(conf.LoadBalancer.Servers, m.rand) {
			srvLogger := logger.With().
				Int(logs.ServerIndex, index).
				Str("serverAddress", server.Address).Logger()

			if _, _, err := net.SplitHostPort(server.Address); err != nil {
				srvLogger.Error().Err(err).Msg("Failed to split host port")
				continue
			}

			dialer, err := m.dialerManager.Build(conf.LoadBalancer, server.TLS)
			if err != nil {
				return nil, err
			}

			handler, err := tcp.NewProxy(server.Address, dialer)
			if err != nil {
				srvLogger.Error().Err(err).Msg("Failed to create server")
				continue
			}

			loadBalancer.Add(server.Address, handler, nil)

			// Servers are considered UP by default.
			conf.UpdateServerStatus(server.Address, runtime.StatusUp)

			uniqHealthCheckTargets[server.Address] = healthcheck.TCPHealthCheckTarget{
				Address: server.Address,
				TLS:     server.TLS,
				Dialer:  dialer,
			}

			logger.Debug().Msg("Creating TCP server")
		}

		if conf.LoadBalancer.HealthCheck != nil {
			m.healthCheckers[serviceName] = healthcheck.NewServiceTCPHealthChecker(
				ctx,
				conf.LoadBalancer.HealthCheck,
				loadBalancer,
				conf,
				slices.Collect(maps.Values(uniqHealthCheckTargets)),
				serviceQualifiedName)
		}

		return loadBalancer, nil

	case conf.Weighted != nil:
		loadBalancer := tcp.NewWRRLoadBalancer(conf.Weighted.HealthCheck != nil)

		for _, service := range shuffle(conf.Weighted.Services, m.rand) {
			handler, err := m.BuildTCP(ctx, service.Name)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to build TCP handler")
				return nil, err
			}

			loadBalancer.Add(service.Name, handler, service.Weight)

			if conf.Weighted.HealthCheck == nil {
				continue
			}

			updater, ok := handler.(healthcheck.StatusUpdater)
			if !ok {
				return nil, fmt.Errorf("child service %v of %v not a healthcheck.StatusUpdater (%T)", service.Name, serviceName, handler)
			}

			if err := updater.RegisterStatusUpdater(func(up bool) {
				loadBalancer.SetStatus(ctx, service.Name, up)
			}); err != nil {
				return nil, fmt.Errorf("cannot register %v as updater for %v: %w", service.Name, serviceName, err)
			}

			log.Ctx(ctx).Debug().Str("parent", serviceName).Str("child", service.Name).
				Msg("Child service will update parent on status change")
		}

		return loadBalancer, nil

	default:
		err := fmt.Errorf("the service %q does not have any type defined", serviceQualifiedName)
		conf.AddError(err, true)
		return nil, err
	}
}

// LaunchHealthCheck launches the health checks.
func (m *Manager) LaunchHealthCheck(ctx context.Context) {
	for serviceName, hc := range m.healthCheckers {
		logger := log.Ctx(ctx).With().Str(logs.ServiceName, serviceName).Logger()
		go hc.Launch(logger.WithContext(ctx))
	}
}

func shuffle[T any](values []T, r *rand.Rand) []T {
	shuffled := make([]T, len(values))
	copy(shuffled, values)
	r.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	return shuffled
}
