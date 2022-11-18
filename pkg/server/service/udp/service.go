package udp

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/logs"
	"github.com/traefik/traefik/v2/pkg/server/provider"
	"github.com/traefik/traefik/v2/pkg/udp"
)

// Manager handles UDP services creation.
type Manager struct {
	configs map[string]*runtime.UDPServiceInfo
	rand    *rand.Rand // For the initial shuffling of load-balancers.
}

// NewManager creates a new manager.
func NewManager(conf *runtime.Configuration) *Manager {
	return &Manager{
		configs: conf.UDPServices,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// BuildUDP creates the UDP handler for the given service name.
func (m *Manager) BuildUDP(rootCtx context.Context, serviceName string) (udp.Handler, error) {
	serviceQualifiedName := provider.GetQualifiedName(rootCtx, serviceName)

	logger := log.Ctx(rootCtx).With().Str(logs.ServiceName, serviceName).Logger()
	ctx := logger.WithContext(provider.AddInContext(rootCtx, serviceQualifiedName))

	conf, ok := m.configs[serviceQualifiedName]
	if !ok {
		return nil, fmt.Errorf("the udp service %q does not exist", serviceQualifiedName)
	}

	if conf.LoadBalancer != nil && conf.Weighted != nil {
		err := errors.New("cannot create service: multi-types service not supported, consider declaring two different pieces of service instead")
		conf.AddError(err, true)
		return nil, err
	}

	switch {
	case conf.LoadBalancer != nil:
		loadBalancer := udp.NewWRRLoadBalancer()

		for index, server := range shuffle(conf.LoadBalancer.Servers, m.rand) {
			if _, _, err := net.SplitHostPort(server.Address); err != nil {
				logger.Error().Err(err).Msgf("In udp service %q", serviceQualifiedName)
				continue
			}

			handler, err := udp.NewProxy(server.Address)
			if err != nil {
				logger.Error().Err(err).Msgf("In udp service %q server %q", serviceQualifiedName, server.Address)
				continue
			}

			loadBalancer.AddServer(handler)
			logger.Debug().Int(logs.ServerIndex, index).Str("serverAddress", server.Address).
				Msg("Creating UDP server")
		}
		return loadBalancer, nil
	case conf.Weighted != nil:
		loadBalancer := udp.NewWRRLoadBalancer()

		for _, service := range shuffle(conf.Weighted.Services, m.rand) {
			handler, err := m.BuildUDP(ctx, service.Name)
			if err != nil {
				logger.Error().Err(err).Msgf("In udp service %q", serviceQualifiedName)
				return nil, err
			}
			loadBalancer.AddWeightedServer(handler, service.Weight)
		}
		return loadBalancer, nil
	default:
		err := fmt.Errorf("the udp service %q does not have any type defined", serviceQualifiedName)
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
