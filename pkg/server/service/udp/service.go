package udp

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/log"
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
	ctx := provider.AddInContext(rootCtx, serviceQualifiedName)
	ctx = log.With(ctx, log.Str(log.ServiceName, serviceName))

	conf, ok := m.configs[serviceQualifiedName]
	if !ok {
		return nil, fmt.Errorf("the udp service %q does not exist", serviceQualifiedName)
	}

	if conf.LoadBalancer != nil && conf.Weighted != nil {
		err := errors.New("cannot create service: multi-types service not supported, consider declaring two different pieces of service instead")
		conf.AddError(err, true)
		return nil, err
	}

	logger := log.FromContext(ctx)
	switch {
	case conf.LoadBalancer != nil:
		loadBalancer := udp.NewWRRLoadBalancer()

		for name, server := range shuffle(conf.LoadBalancer.Servers, m.rand) {
			if _, _, err := net.SplitHostPort(server.Address); err != nil {
				logger.Errorf("In udp service %q: %v", serviceQualifiedName, err)
				continue
			}

			handler, err := udp.NewProxy(server.Address)
			if err != nil {
				logger.Errorf("In udp service %q server %q: %v", serviceQualifiedName, server.Address, err)
				continue
			}

			loadBalancer.AddServer(handler)
			logger.WithField(log.ServerName, name).Debugf("Creating UDP server %d at %s", name, server.Address)
		}
		return loadBalancer, nil
	case conf.Weighted != nil:
		loadBalancer := udp.NewWRRLoadBalancer()

		for _, service := range shuffle(conf.Weighted.Services, m.rand) {
			handler, err := m.BuildUDP(rootCtx, service.Name)
			if err != nil {
				logger.Errorf("In udp service %q: %v", serviceQualifiedName, err)
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
