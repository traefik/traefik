package tcp

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/server/provider"
	"github.com/traefik/traefik/v3/pkg/tcp"
	"golang.org/x/net/proxy"
)

// Manager is the TCPHandlers factory.
type Manager struct {
	dialerManager *tcp.DialerManager
	configs       map[string]*runtime.TCPServiceInfo
	rand          *rand.Rand // For the initial shuffling of load-balancers.
}

// NewManager creates a new manager.
func NewManager(conf *runtime.Configuration, dialerManager *tcp.DialerManager) *Manager {
	return &Manager{
		dialerManager: dialerManager,
		configs:       conf.TCPServices,
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
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
		loadBalancer := tcp.NewWRRLoadBalancer()

		if conf.LoadBalancer.TerminationDelay != nil {
			log.Ctx(ctx).Warn().Msgf("Service %q load balancer uses `TerminationDelay`, but this option is deprecated, please use ServersTransport configuration instead.", serviceName)
		}

		if len(conf.LoadBalancer.ServersTransport) > 0 {
			conf.LoadBalancer.ServersTransport = provider.GetQualifiedName(ctx, conf.LoadBalancer.ServersTransport)
		}

		for index, server := range shuffle(conf.LoadBalancer.Servers, m.rand) {
			srvLogger := logger.With().
				Int(logs.ServerIndex, index).
				Str("serverAddress", server.Address).Logger()

			if _, _, err := net.SplitHostPort(server.Address); err != nil {
				srvLogger.Error().Err(err).Msg("Failed to split host port")
				continue
			}

			dialer, err := m.dialerManager.Get(conf.LoadBalancer.ServersTransport, server.TLS)
			if err != nil {
				return nil, err
			}

			// Handle TerminationDelay deprecated option.
			if conf.LoadBalancer.ServersTransport == "" && conf.LoadBalancer.TerminationDelay != nil {
				dialer = &dialerWrapper{
					Dialer:           dialer,
					terminationDelay: time.Duration(*conf.LoadBalancer.TerminationDelay),
				}
			}

			handler, err := tcp.NewProxy(server.Address, conf.LoadBalancer.ProxyProtocol, dialer)
			if err != nil {
				srvLogger.Error().Err(err).Msg("Failed to create server")
				continue
			}

			loadBalancer.AddServer(handler)
			logger.Debug().Msg("Creating TCP server")
		}

		return loadBalancer, nil

	case conf.Weighted != nil:
		loadBalancer := tcp.NewWRRLoadBalancer()

		for _, service := range shuffle(conf.Weighted.Services, m.rand) {
			handler, err := m.BuildTCP(ctx, service.Name)
			if err != nil {
				logger.Error().Err(err).Msg("Failed to build TCP handler")
				return nil, err
			}

			loadBalancer.AddWeightServer(handler, service.Weight)
		}

		return loadBalancer, nil

	default:
		err := fmt.Errorf("the service %q does not have any type defined", serviceQualifiedName)
		conf.AddError(err, true)
		return nil, err
	}
}

func shuffle[T any](values []T, r *rand.Rand) []T {
	shuffled := make([]T, len(values))
	copy(shuffled, values)
	r.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	return shuffled
}

// dialerWrapper is only used to handle TerminationDelay deprecated option on TCPServersLoadBalancer.
type dialerWrapper struct {
	proxy.Dialer
	terminationDelay time.Duration
}

func (d dialerWrapper) TerminationDelay() time.Duration {
	return d.terminationDelay
}
