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
	configs map[string]*config.TCPService
}

// NewManager creates a new manager
func NewManager(configs map[string]*config.TCPService) *Manager {
	return &Manager{
		configs: configs,
	}
}

// BuildTCP Creates a tcp.Handler for a service configuration.
func (m *Manager) BuildTCP(rootCtx context.Context, serviceName string) (tcp.Handler, error) {
	serviceQualifiedName := internal.GetQualifiedName(rootCtx, serviceName)
	ctx := internal.AddProviderInContext(rootCtx, serviceQualifiedName)
	ctx = log.With(ctx, log.Str(log.ServiceName, serviceName))

	conf, ok := m.configs[serviceQualifiedName]
	if !ok {
		return nil, fmt.Errorf("the service %q does not exits", serviceQualifiedName)
	}

	if conf.LoadBalancer == nil {
		return nil, fmt.Errorf("the service %q doesn't have any TCP load balancer", serviceQualifiedName)
	}

	logger := log.FromContext(ctx)

	// FIXME Check if the service is declared multiple times with different types
	loadBalancer := tcp.NewRRLoadBalancer()

	for _, server := range conf.LoadBalancer.Servers {
		if _, err := parseIP(server.Address); err != nil {
			logger.Errorf("Invalid IP address for a %q server %q: %v", serviceQualifiedName, server.Address, err)
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

func parseIP(s string) (string, error) {
	ip, _, err := net.SplitHostPort(s)
	if err == nil {
		return ip, nil
	}

	ipNoPort := net.ParseIP(s)
	if ipNoPort == nil {
		return "", fmt.Errorf("invalid IP Address %s", ipNoPort)
	}

	return ipNoPort.String(), nil
}
