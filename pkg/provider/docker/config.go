package docker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/config/label"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/provider"
	"github.com/containous/traefik/pkg/provider/constraints"
	"github.com/docker/go-connections/nat"
)

func (p *Provider) buildConfiguration(ctx context.Context, containersInspected []dockerData) *config.Configuration {
	configurations := make(map[string]*config.Configuration)

	for _, container := range containersInspected {
		containerName := getServiceName(container) + "-" + container.ID
		ctxContainer := log.With(ctx, log.Str("container", containerName))

		if !p.keepContainer(ctxContainer, container) {
			continue
		}

		logger := log.FromContext(ctxContainer)

		confFromLabel, err := label.DecodeConfiguration(container.Labels)
		if err != nil {
			logger.Error(err)
			continue
		}

		if len(confFromLabel.TCP.Routers) > 0 || len(confFromLabel.TCP.Services) > 0 {
			err := p.buildTCPServiceConfiguration(ctxContainer, container, confFromLabel.TCP)
			if err != nil {
				logger.Error(err)
				continue
			}
			provider.BuildTCPRouterConfiguration(ctxContainer, confFromLabel.TCP)
			if len(confFromLabel.HTTP.Routers) == 0 &&
				len(confFromLabel.HTTP.Middlewares) == 0 &&
				len(confFromLabel.HTTP.Services) == 0 {
				configurations[containerName] = confFromLabel
				continue
			}
		}

		err = p.buildServiceConfiguration(ctxContainer, container, confFromLabel.HTTP)
		if err != nil {
			logger.Error(err)
			continue
		}

		serviceName := getServiceName(container)

		model := struct {
			Name   string
			Labels map[string]string
		}{
			Name:   serviceName,
			Labels: container.Labels,
		}

		provider.BuildRouterConfiguration(ctx, confFromLabel.HTTP, serviceName, p.defaultRuleTpl, model)

		configurations[containerName] = confFromLabel
	}

	return provider.Merge(ctx, configurations)
}

func (p *Provider) buildTCPServiceConfiguration(ctx context.Context, container dockerData, configuration *config.TCPConfiguration) error {
	serviceName := getServiceName(container)

	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*config.TCPService)
		lb := &config.TCPLoadBalancerService{}
		configuration.Services[serviceName] = &config.TCPService{
			LoadBalancer: lb,
		}
	}

	for _, service := range configuration.Services {
		err := p.addServerTCP(ctx, container, service.LoadBalancer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) buildServiceConfiguration(ctx context.Context, container dockerData, configuration *config.HTTPConfiguration) error {
	serviceName := getServiceName(container)

	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*config.Service)
		lb := &config.LoadBalancerService{}
		lb.SetDefaults()
		configuration.Services[serviceName] = &config.Service{
			LoadBalancer: lb,
		}
	}

	for _, service := range configuration.Services {
		err := p.addServer(ctx, container, service.LoadBalancer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) keepContainer(ctx context.Context, container dockerData) bool {
	logger := log.FromContext(ctx)

	if !container.ExtraConf.Enable {
		logger.Debug("Filtering disabled container")
		return false
	}

	matches, err := constraints.Match(container.Labels, p.Constraints)
	if err != nil {
		logger.Error("Error matching constraints expression: %v", err)
		return false
	}
	if !matches {
		logger.Debugf("Container pruned by constraint expression: %q", p.Constraints)
		return false
	}

	if container.Health != "" && container.Health != "healthy" {
		logger.Debug("Filtering unhealthy or starting container")
		return false
	}

	return true
}

func (p *Provider) addServerTCP(ctx context.Context, container dockerData, loadBalancer *config.TCPLoadBalancerService) error {
	serverPort := ""
	if loadBalancer != nil && len(loadBalancer.Servers) > 0 {
		serverPort = loadBalancer.Servers[0].Port
	}
	ip, port, err := p.getIPPort(ctx, container, serverPort)
	if err != nil {
		return err
	}

	if len(loadBalancer.Servers) == 0 {
		server := config.TCPServer{}

		loadBalancer.Servers = []config.TCPServer{server}
	}

	if serverPort != "" {
		port = serverPort
		loadBalancer.Servers[0].Port = ""
	}

	if port == "" {
		return errors.New("port is missing")
	}

	loadBalancer.Servers[0].Address = net.JoinHostPort(ip, port)
	return nil
}

func (p *Provider) addServer(ctx context.Context, container dockerData, loadBalancer *config.LoadBalancerService) error {
	serverPort := getLBServerPort(loadBalancer)
	ip, port, err := p.getIPPort(ctx, container, serverPort)
	if err != nil {
		return err
	}

	if len(loadBalancer.Servers) == 0 {
		server := config.Server{}
		server.SetDefaults()

		loadBalancer.Servers = []config.Server{server}
	}

	if serverPort != "" {
		port = serverPort
		loadBalancer.Servers[0].Port = ""
	}

	if port == "" {
		return errors.New("port is missing")
	}

	loadBalancer.Servers[0].URL = fmt.Sprintf("%s://%s", loadBalancer.Servers[0].Scheme, net.JoinHostPort(ip, port))
	loadBalancer.Servers[0].Scheme = ""

	return nil
}

func (p *Provider) getIPPort(ctx context.Context, container dockerData, serverPort string) (string, string, error) {
	logger := log.FromContext(ctx)

	var ip, port string
	usedBound := false

	if p.UseBindPortIP {
		portBinding, err := p.getPortBinding(container, serverPort)
		switch {
		case err != nil:
			logger.Infof("Unable to find a binding for container %q, falling back on its internal IP/Port.", container.Name)
		case portBinding.HostIP == "0.0.0.0" || len(portBinding.HostIP) == 0:
			logger.Infof("Cannot determine the IP address (got %q) for %q's binding, falling back on its internal IP/Port.", portBinding.HostIP, container.Name)
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

func (p Provider) getIPAddress(ctx context.Context, container dockerData) string {
	logger := log.FromContext(ctx)

	if container.ExtraConf.Docker.Network != "" {
		settings := container.NetworkSettings
		if settings.Networks != nil {
			network := settings.Networks[container.ExtraConf.Docker.Network]
			if network != nil {
				return network.Addr
			}

			logger.Warnf("Could not find network named '%s' for container '%s'! Maybe you're missing the project's prefix in the label? Defaulting to first available network.", container.ExtraConf.Docker.Network, container.Name)
		}
	}

	if container.NetworkSettings.NetworkMode.IsHost() {
		if container.Node != nil && container.Node.IPAddress != "" {
			return container.Node.IPAddress
		}
		return "127.0.0.1"
	}

	if container.NetworkSettings.NetworkMode.IsContainer() {
		dockerClient, err := p.createClient()
		if err != nil {
			logger.Warnf("Unable to get IP address: %s", err)
			return ""
		}

		connectedContainer := container.NetworkSettings.NetworkMode.ConnectedContainer()
		containerInspected, err := dockerClient.ContainerInspect(context.Background(), connectedContainer)
		if err != nil {
			logger.Warnf("Unable to get IP address for container %s : Failed to inspect container ID %s, error: %s", container.Name, connectedContainer, err)
			return ""
		}
		return p.getIPAddress(ctx, parseContainer(containerInspected))
	}

	for _, network := range container.NetworkSettings.Networks {
		return network.Addr
	}

	logger.Warn("Unable to find the IP address.")
	return ""
}

func (p *Provider) getPortBinding(container dockerData, serverPort string) (*nat.PortBinding, error) {
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

func getLBServerPort(loadBalancer *config.LoadBalancerService) string {
	if loadBalancer != nil && len(loadBalancer.Servers) > 0 {
		return loadBalancer.Servers[0].Port
	}
	return ""
}

func getPort(container dockerData, serverPort string) string {
	if len(serverPort) > 0 {
		return serverPort
	}

	var ports []nat.Port
	for port := range container.NetworkSettings.Ports {
		ports = append(ports, port)
	}

	less := func(i, j nat.Port) bool {
		return i.Int() < j.Int()
	}
	nat.Sort(ports, less)

	if len(ports) > 0 {
		min := ports[0]
		return min.Port()
	}

	return ""
}

func getServiceName(container dockerData) string {
	serviceName := container.ServiceName

	if values, err := getStringMultipleStrict(container.Labels, labelDockerComposeProject, labelDockerComposeService); err == nil {
		serviceName = values[labelDockerComposeService] + "_" + values[labelDockerComposeProject]
	}

	return serviceName
}
