package docker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/label"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/provider/constraints"
)

type DynConfBuilder struct {
	Shared
	apiClient client.APIClient
	swarm     bool
}

func NewDynConfBuilder(configuration Shared, apiClient client.APIClient, swarm bool) *DynConfBuilder {
	return &DynConfBuilder{Shared: configuration, apiClient: apiClient, swarm: swarm}
}

func (p *DynConfBuilder) build(ctx context.Context, containersInspected []dockerData) *dynamic.Configuration {
	configurations := make(map[string]*dynamic.Configuration)

	for _, container := range containersInspected {
		containerName := getServiceName(container) + "-" + container.ID

		logger := log.Ctx(ctx).With().Str("container", containerName).Logger()
		ctxContainer := logger.WithContext(ctx)

		if !p.keepContainer(ctxContainer, container) {
			continue
		}

		confFromLabel, err := label.DecodeConfiguration(container.Labels)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		var tcpOrUDP bool
		if len(confFromLabel.TCP.Routers) > 0 || len(confFromLabel.TCP.Services) > 0 {
			tcpOrUDP = true

			err := p.buildTCPServiceConfiguration(ctxContainer, container, confFromLabel.TCP)
			if err != nil {
				logger.Error().Err(err).Send()
				continue
			}
			provider.BuildTCPRouterConfiguration(ctxContainer, confFromLabel.TCP)
		}

		if len(confFromLabel.UDP.Routers) > 0 || len(confFromLabel.UDP.Services) > 0 {
			tcpOrUDP = true

			err := p.buildUDPServiceConfiguration(ctxContainer, container, confFromLabel.UDP)
			if err != nil {
				logger.Error().Err(err).Send()
				continue
			}
			provider.BuildUDPRouterConfiguration(ctxContainer, confFromLabel.UDP)
		}

		if tcpOrUDP && len(confFromLabel.HTTP.Routers) == 0 &&
			len(confFromLabel.HTTP.Middlewares) == 0 &&
			len(confFromLabel.HTTP.Services) == 0 {
			configurations[containerName] = confFromLabel
			continue
		}

		err = p.buildServiceConfiguration(ctxContainer, container, confFromLabel.HTTP)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		serviceName := getServiceName(container)

		model := struct {
			Name          string
			ContainerName string
			Labels        map[string]string
		}{
			Name:          serviceName,
			ContainerName: strings.TrimPrefix(container.Name, "/"),
			Labels:        container.Labels,
		}

		provider.BuildRouterConfiguration(ctx, confFromLabel.HTTP, serviceName, p.defaultRuleTpl, model)

		configurations[containerName] = confFromLabel
	}

	return provider.Merge(ctx, configurations)
}

func (p *DynConfBuilder) buildTCPServiceConfiguration(ctx context.Context, container dockerData, configuration *dynamic.TCPConfiguration) error {
	serviceName := getServiceName(container)

	if len(configuration.Services) == 0 {
		configuration.Services = map[string]*dynamic.TCPService{
			serviceName: {
				LoadBalancer: new(dynamic.TCPServersLoadBalancer),
			},
		}
	}

	if container.Health != "" && container.Health != dockertypes.Healthy {
		return nil
	}

	for name, service := range configuration.Services {
		ctx := log.Ctx(ctx).With().Str(logs.ServiceName, name).Logger().WithContext(ctx)
		if err := p.addServerTCP(ctx, container, service.LoadBalancer); err != nil {
			return fmt.Errorf("service %q error: %w", name, err)
		}
	}

	return nil
}

func (p *DynConfBuilder) buildUDPServiceConfiguration(ctx context.Context, container dockerData, configuration *dynamic.UDPConfiguration) error {
	serviceName := getServiceName(container)

	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.UDPService)
		configuration.Services[serviceName] = &dynamic.UDPService{
			LoadBalancer: &dynamic.UDPServersLoadBalancer{},
		}
	}

	if container.Health != "" && container.Health != dockertypes.Healthy {
		return nil
	}

	for name, service := range configuration.Services {
		ctx := log.Ctx(ctx).With().Str(logs.ServiceName, name).Logger().WithContext(ctx)
		if err := p.addServerUDP(ctx, container, service.LoadBalancer); err != nil {
			return fmt.Errorf("service %q error: %w", name, err)
		}
	}

	return nil
}

func (p *DynConfBuilder) buildServiceConfiguration(ctx context.Context, container dockerData, configuration *dynamic.HTTPConfiguration) error {
	serviceName := getServiceName(container)

	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.Service)
		lb := &dynamic.ServersLoadBalancer{}
		lb.SetDefaults()
		configuration.Services[serviceName] = &dynamic.Service{
			LoadBalancer: lb,
		}
	}

	if container.Health != "" && container.Health != dockertypes.Healthy {
		return nil
	}

	for name, service := range configuration.Services {
		ctx := log.Ctx(ctx).With().Str(logs.ServiceName, name).Logger().WithContext(ctx)
		if err := p.addServer(ctx, container, service.LoadBalancer); err != nil {
			return fmt.Errorf("service %q error: %w", name, err)
		}
	}

	return nil
}

func (p *DynConfBuilder) keepContainer(ctx context.Context, container dockerData) bool {
	logger := log.Ctx(ctx)

	if !container.ExtraConf.Enable {
		logger.Debug().Msg("Filtering disabled container")
		return false
	}

	matches, err := constraints.MatchLabels(container.Labels, p.Constraints)
	if err != nil {
		logger.Error().Err(err).Msg("Error matching constraints expression")
		return false
	}
	if !matches {
		logger.Debug().Msgf("Container pruned by constraint expression: %q", p.Constraints)
		return false
	}

	if !p.AllowEmptyServices && container.Health != "" && container.Health != dockertypes.Healthy {
		logger.Debug().Msg("Filtering unhealthy or starting container")
		return false
	}

	return true
}

func (p *DynConfBuilder) addServerTCP(ctx context.Context, container dockerData, loadBalancer *dynamic.TCPServersLoadBalancer) error {
	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	if len(loadBalancer.Servers) == 0 {
		loadBalancer.Servers = []dynamic.TCPServer{{}}
	}

	serverPort := loadBalancer.Servers[0].Port
	loadBalancer.Servers[0].Port = ""

	ip, port, err := p.getIPPort(ctx, container, serverPort)
	if err != nil {
		return err
	}

	if port == "" {
		return errors.New("port is missing")
	}

	loadBalancer.Servers[0].Address = net.JoinHostPort(ip, port)

	return nil
}

func (p *DynConfBuilder) addServerUDP(ctx context.Context, container dockerData, loadBalancer *dynamic.UDPServersLoadBalancer) error {
	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	if len(loadBalancer.Servers) == 0 {
		loadBalancer.Servers = []dynamic.UDPServer{{}}
	}

	serverPort := loadBalancer.Servers[0].Port
	loadBalancer.Servers[0].Port = ""

	ip, port, err := p.getIPPort(ctx, container, serverPort)
	if err != nil {
		return err
	}

	if port == "" {
		return errors.New("port is missing")
	}

	loadBalancer.Servers[0].Address = net.JoinHostPort(ip, port)

	return nil
}

func (p *DynConfBuilder) addServer(ctx context.Context, container dockerData, loadBalancer *dynamic.ServersLoadBalancer) error {
	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	if len(loadBalancer.Servers) == 0 {
		loadBalancer.Servers = []dynamic.Server{{}}
	}

	if loadBalancer.Servers[0].URL != "" {
		if loadBalancer.Servers[0].Scheme != "" || loadBalancer.Servers[0].Port != "" {
			return errors.New("defining scheme or port is not allowed when URL is defined")
		}
		return nil
	}

	serverPort := loadBalancer.Servers[0].Port
	loadBalancer.Servers[0].Port = ""

	ip, port, err := p.getIPPort(ctx, container, serverPort)
	if err != nil {
		return err
	}

	if port == "" {
		return errors.New("port is missing")
	}

	scheme := loadBalancer.Servers[0].Scheme
	loadBalancer.Servers[0].Scheme = ""
	if scheme == "" {
		scheme = "http"
	}

	loadBalancer.Servers[0].URL = fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(ip, port))

	return nil
}

func (p *DynConfBuilder) getIPPort(ctx context.Context, container dockerData, serverPort string) (string, string, error) {
	logger := log.Ctx(ctx)

	var ip, port string
	usedBound := false

	if p.UseBindPortIP {
		portBinding, err := p.getPortBinding(container, serverPort)
		switch {
		case err != nil:
			logger.Info().Msgf("Unable to find a binding for container %q, falling back on its internal IP/Port.", container.Name)
		case portBinding.HostIP == "0.0.0.0" || len(portBinding.HostIP) == 0:
			logger.Info().Msgf("Cannot determine the IP address (got %q) for %q's binding, falling back on its internal IP/Port.", portBinding.HostIP, container.Name)
		default:
			ip = portBinding.HostIP
			port = portBinding.HostPort
			usedBound = true
		}
	}

	if !usedBound {
		ip = p.getIPAddress(ctx, container)
		port = getPort(container, serverPort)
	}

	if len(ip) == 0 {
		return "", "", fmt.Errorf("unable to find the IP address for the container %q: the server is ignored", container.Name)
	}

	return ip, port, nil
}

func (p *DynConfBuilder) getIPAddress(ctx context.Context, container dockerData) string {
	logger := log.Ctx(ctx)

	netNotFound := false
	if container.ExtraConf.Network != "" {
		settings := container.NetworkSettings
		if settings.Networks != nil {
			network := settings.Networks[container.ExtraConf.Network]
			if network != nil {
				return network.Addr
			}

			netNotFound = true
		}
	}

	if container.NetworkSettings.NetworkMode.IsHost() {
		if container.Node != nil && container.Node.IPAddress != "" {
			return container.Node.IPAddress
		}
		if host, err := net.LookupHost("host.docker.internal"); err == nil {
			return host[0]
		}
		if host, err := net.LookupHost("host.containers.internal"); err == nil {
			return host[0]
		}
		return "127.0.0.1"
	}

	if container.NetworkSettings.NetworkMode.IsContainer() {
		connectedContainer := container.NetworkSettings.NetworkMode.ConnectedContainer()
		containerInspected, err := p.apiClient.ContainerInspect(context.Background(), connectedContainer)
		if err != nil {
			logger.Warn().Err(err).Msgf("Unable to get IP address for container %s: failed to inspect container ID %s", container.Name, connectedContainer)
			return ""
		}

		// Check connected container for traefik.docker.network,
		// falling back to the network specified on the current container.
		containerParsed := parseContainer(containerInspected)
		extraConf, err := p.extractLabels(containerParsed)
		if err != nil {
			logger.Warn().Err(err).Msgf("Unable to get IP address for container %s: failed to get extra configuration for container %s", container.Name, containerInspected.Name)
			return ""
		}

		if extraConf.Network == "" {
			extraConf.Network = container.ExtraConf.Network
		}

		containerParsed.ExtraConf = extraConf
		return p.getIPAddress(ctx, containerParsed)
	}

	if netNotFound {
		logger.Warn().Msgf("Could not find network named %q for container %q. Maybe you're missing the project's prefix in the label?", container.ExtraConf.Network, container.Name)
	}
	for _, network := range container.NetworkSettings.Networks {
		if netNotFound {
			logger.Warn().Msgf("Defaulting to first available network (%q) for container %q.", network, container.Name)
		}
		return network.Addr
	}

	logger.Warn().Msg("Unable to find the IP address.")
	return ""
}

func (p *DynConfBuilder) getPortBinding(container dockerData, serverPort string) (*nat.PortBinding, error) {
	port := getPort(container, serverPort)

	for netPort, portBindings := range container.NetworkSettings.Ports {
		if strings.EqualFold(string(netPort), port+"/TCP") || strings.EqualFold(string(netPort), port+"/UDP") {
			for _, p := range portBindings {
				return &p, nil
			}
		}
	}

	return nil, fmt.Errorf("unable to find the external IP:Port for the container %q", container.Name)
}

func (p *DynConfBuilder) extractLabels(container dockerData) (configuration, error) {
	if p.swarm {
		return p.Shared.extractSwarmLabels(container)
	}
	return p.Shared.extractDockerLabels(container)
}
