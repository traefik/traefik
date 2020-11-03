package tcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/server/provider"
	"github.com/traefik/traefik/v2/pkg/tcp"
)

// Manager is the TCPHandlers factory.
type Manager struct {
	configs map[string]*runtime.TCPServiceInfo
}

// NewManager creates a new manager.
func NewManager(conf *runtime.Configuration) *Manager {
	return &Manager{
		configs: conf.TCPServices,
	}
}

// BuildTCP Creates a tcp.Handler for a service configuration.
func (m *Manager) BuildTCP(rootCtx context.Context, serviceName string) (tcp.Handler, error) {
	serviceQualifiedName := provider.GetQualifiedName(rootCtx, serviceName)
	ctx := provider.AddInContext(rootCtx, serviceQualifiedName)
	ctx = log.With(ctx, log.Str(log.ServiceName, serviceName))

	conf, ok := m.configs[serviceQualifiedName]
	if !ok {
		return nil, fmt.Errorf("the service %q does not exist", serviceQualifiedName)
	}

	if conf.LoadBalancer != nil && conf.Weighted != nil {
		err := errors.New("cannot create service: multi-types service not supported, consider declaring two different pieces of service instead")
		conf.AddError(err, true)
		return nil, err
	}

	logger := log.FromContext(ctx)
	switch {
	case conf.LoadBalancer != nil:
		loadBalancer := tcp.NewWRRLoadBalancer()

		if conf.LoadBalancer.TerminationDelay == nil {
			defaultTerminationDelay := 100
			conf.LoadBalancer.TerminationDelay = &defaultTerminationDelay
		}
		duration := time.Duration(*conf.LoadBalancer.TerminationDelay) * time.Millisecond

		for name, server := range conf.LoadBalancer.Servers {
			if _, _, err := net.SplitHostPort(server.Address); err != nil {
				logger.Errorf("In service %q: %v", serviceQualifiedName, err)
				continue
			}

			handler, err := tcp.NewProxy(server.Address, duration)
			if err != nil {
				logger.Errorf("In service %q server %q: %v", serviceQualifiedName, server.Address, err)
				continue
			}

			loadBalancer.AddServer(handler)
			logger.WithField(log.ServerName, name).Debugf("Creating TCP server %d at %s", name, server.Address)
		}
		return loadBalancer, nil
	case conf.Weighted != nil:
		loadBalancer := tcp.NewWRRLoadBalancer()
		for _, service := range conf.Weighted.Services {
			handler, err := m.BuildTCP(rootCtx, service.Name)
			if err != nil {
				logger.Errorf("In service %q: %v", serviceQualifiedName, err)
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
