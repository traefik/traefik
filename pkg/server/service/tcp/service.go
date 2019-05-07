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

// GetRuntimeConfiguration returns the configuration of all the current TCP services.
func (m Manager) GetRuntimeConfiguration() map[string]*config.TCPServiceInfo {
	return m.configs
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
		lbErr := fmt.Errorf("the service %q doesn't have any TCP load balancer", serviceQualifiedName)
		conf.Err = lbErr
		return nil, lbErr
	}

	loadBalancer := tcp.NewRRLoadBalancer()

	for _, server := range conf.LoadBalancer.Servers {
		if _, _, err := net.SplitHostPort(server.Address); err != nil {
			conf.Err = fmt.Errorf("in service %q: %v", serviceQualifiedName, err)
			return nil, conf.Err
		}

		handler, err := tcp.NewProxy(server.Address)
		if err != nil {
			conf.Err = fmt.Errorf("in service %q, server %q: %v", serviceQualifiedName, server.Address, err)
			return nil, conf.Err
		}

		loadBalancer.AddServer(handler)
	}
	return loadBalancer, nil
}
