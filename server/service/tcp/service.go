package tcp

import (
	"context"
	"fmt"
	"net"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/server/internal"
	"github.com/containous/traefik/tcp"
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
	ctx := log.With(rootCtx, log.Str(log.ServiceName, serviceName))

	serviceName = internal.GetQualifiedName(ctx, serviceName)
	ctx = internal.AddProviderInContext(ctx, serviceName)

	if conf, ok := m.configs[serviceName]; ok {
		// FIXME Check if the service is declared multiple times with different types
		if conf.LoadBalancer != nil {
			loadBalancer := tcp.NewRRLoadBalancer()

			var handler tcp.Handler
			for _, server := range conf.LoadBalancer.Servers {
				_, err := parseIP(server.Address)
				if err == nil {
					handler, _ = tcp.NewProxy(server.Address)
					loadBalancer.AddServer(handler)
				} else {
					log.FromContext(ctx).Errorf("Invalid IP address for a %s server %s: %v", serviceName, server.Address, err)
				}
			}
			return loadBalancer, nil
		}
		return nil, fmt.Errorf("the service %q doesn't have any TCP load balancer", serviceName)
	}
	return nil, fmt.Errorf("the service %q does not exits", serviceName)
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
