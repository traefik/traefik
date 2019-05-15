package tcp

import (
	"context"
	"fmt"
	"net"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/server/internal"
	"github.com/containous/traefik/pkg/tcp"
)

// Manager is the TCPHandlers factory
type Manager struct {
	configs map[string]*config.TCPServiceInfo
}

// NewManager creates a new manager
func NewManager(conf *config.RuntimeConfiguration) *Manager {
	return &Manager{
		configs: conf.TCPServices,
	}
}

// BuildTCP Creates a tcp.Handler for a service configuration.
func (m *Manager) BuildTCP(rootCtx context.Context, serviceName string) (tcp.Handler, error) {
	serviceQualifiedName := internal.GetQualifiedName(rootCtx, serviceName)
	ctx := internal.AddProviderInContext(rootCtx, serviceQualifiedName)
	ctx = log.With(ctx, log.Str(log.ServiceName, serviceName))

	// FIXME Check if the service is declared multiple times with different types
	conf, ok := m.configs[serviceQualifiedName]
	if !ok {
		return nil, fmt.Errorf("the service %q does not exist", serviceQualifiedName)
	}
	if conf.LoadBalancer == nil {
		conf.Err = fmt.Errorf("the service %q doesn't have any TCP load balancer", serviceQualifiedName)
		return nil, conf.Err
	}

	logger := log.FromContext(ctx)

	loadBalancer := tcp.NewRRLoadBalancer()

	for _, server := range conf.LoadBalancer.Servers {
		if _, _, err := net.SplitHostPort(server.Address); err != nil {
			logger.Errorf("In service %q: %v", serviceQualifiedName, err)
			continue
		}

		handler, err := tcp.NewProxy(server.Address)
		if err != nil {
			logger.Errorf("In service %q server %q: %v", serviceQualifiedName, server.Address, err)
			continue
		}

		loadBalancer.AddServer(handler)
	}
	return loadBalancer, nil
}
