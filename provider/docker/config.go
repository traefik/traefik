package docker

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/docker/go-connections/nat"
)

func (p *Provider) buildConfiguration(ctx context.Context, containersInspected []dockerData) *config.Configuration {
	configuration := &config.Configuration{
		Routers:     make(map[string]*config.Router),
		Middlewares: make(map[string]*config.Middleware),
		Services:    make(map[string]*config.Service),
	}

	for _, container := range containersInspected {
		if !p.keepContainer(container) {
			continue
		}

		confFromLabel, err := label.DecodeConfiguration(container.Labels)
		if err != nil {
			log.Error(err)
			continue
		}

		err = p.buildServiceConfiguration(container, confFromLabel)
		if err != nil {
			log.Error(err)
			continue
		}

		provider.MergeServices(ctx, confFromLabel, configuration)

		p.buildRouterConfiguration(container, confFromLabel)
		provider.MergeRouters(ctx, confFromLabel, configuration)

		provider.MergeMiddlewares(ctx, confFromLabel, configuration)
	}

	return configuration
}

func (p *Provider) buildServiceConfiguration(container dockerData, configuration *config.Configuration) error {
	serviceName := getServiceName(container)

	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*config.Service)
		lb := &config.LoadBalancerService{}
		lb.DefaultsHook()
		configuration.Services[serviceName] = &config.Service{
			LoadBalancer: lb,
		}
	}

	for _, service := range configuration.Services {
		err := p.addServer(container, service.LoadBalancer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) buildRouterConfiguration(container dockerData, configuration *config.Configuration) {
	serviceName := getServiceName(container)

	if len(configuration.Routers) == 0 {
		if len(configuration.Services) > 1 {
			log.Infof("could not create a router for the container %s: too many services", serviceName)
		} else {
			configuration.Routers = make(map[string]*config.Router)
			configuration.Routers[serviceName] = &config.Router{}
		}
	}

	for routerName, router := range configuration.Routers {
		if router.Rule == "" {
			router.Rule = "Host:" + getSubDomain(serviceName) + "." + container.ExtraConf.Domain
		}

		if router.Service == "" {
			if len(configuration.Services) > 1 {
				delete(configuration.Routers, routerName)
				log.Errorf("could not define the service name for the router %s: too many services", routerName)
				continue
			}

			for serviceName := range configuration.Services {
				router.Service = serviceName
			}
		}
	}
}

func (p *Provider) keepContainer(container dockerData) bool {
	if !container.ExtraConf.Enable {
		log.Debugf("Filtering disabled container %s", container.Name)
		return false
	}

	if ok, failingConstraint := p.MatchConstraints(container.ExtraConf.Tags); !ok {
		if failingConstraint != nil {
			log.Debugf("Container %s pruned by %q constraint", container.Name, failingConstraint.String())
		}
		return false
	}

	if container.Health != "" && container.Health != "healthy" {
		log.Debugf("Filtering unhealthy or starting container %s", container.Name)
		return false
	}

	return true
}

func (p *Provider) addServer(container dockerData, loadBalancer *config.LoadBalancerService) error {
	serverPort := getLBServerPort(loadBalancer)
	ip, port, err := p.getIPPort(container, serverPort)
	if err != nil {
		return err
	}

	if len(loadBalancer.Servers) == 0 {
		server := config.Server{}
		server.DefaultsHook()

		loadBalancer.Servers = []config.Server{server}
	}

	if serverPort != "" {
		port = serverPort
		loadBalancer.Servers[0].Port = ""
	}

	if port == "" {
		return fmt.Errorf("port is missing on container %s", getServiceName(container))
	}

	loadBalancer.Servers[0].URL = fmt.Sprintf("%s://%s", loadBalancer.Servers[0].Scheme, net.JoinHostPort(ip, port))
	loadBalancer.Servers[0].Scheme = ""

	return nil
}

func (p *Provider) getIPPort(container dockerData, serverPort string) (string, string, error) {
	var ip, port string
	usedBound := false

	if p.UseBindPortIP {
		portBinding, err := p.getPortBinding(container, serverPort)
		if err != nil {
			log.Infof("Unable to find a binding for container %q, falling back on its internal IP/Port.", container.Name)
		} else if (portBinding.HostIP == "0.0.0.0") || (len(portBinding.HostIP) == 0) {
			log.Infof("Cannot determine the IP address (got %q) for %q's binding, falling back on its internal IP/Port.", portBinding.HostIP, container.Name)
		} else {
			ip = portBinding.HostIP
			port = portBinding.HostPort
			usedBound = true
		}
	}

	if !usedBound {
		ip = p.getIPAddress(container)
		port = getPort(container, serverPort)
	}

	if len(ip) == 0 {
		return "", "", fmt.Errorf("unable to find the IP address for the container %q: the server is ignored", container.Name)
	}

	return ip, port, nil
}

func (p Provider) getIPAddress(container dockerData) string {
	if container.ExtraConf.Docker.Network != "" {
		networkSettings := container.NetworkSettings
		if networkSettings.Networks != nil {
			network := networkSettings.Networks[container.ExtraConf.Docker.Network]
			if network != nil {
				return network.Addr
			}

			log.Warnf("Could not find network named '%s' for container '%s'! Maybe you're missing the project's prefix in the label? Defaulting to first available network.", container.ExtraConf.Docker.Network, container.Name)
		}
	}

	if container.NetworkSettings.NetworkMode.IsHost() {
		if container.Node != nil {
			if container.Node.IPAddress != "" {
				return container.Node.IPAddress
			}
		}
		return "127.0.0.1"
	}

	if container.NetworkSettings.NetworkMode.IsContainer() {
		dockerClient, err := p.createClient()
		if err != nil {
			log.Warnf("Unable to get IP address for container %s, error: %s", container.Name, err)
			return ""
		}

		connectedContainer := container.NetworkSettings.NetworkMode.ConnectedContainer()
		containerInspected, err := dockerClient.ContainerInspect(context.Background(), connectedContainer)
		if err != nil {
			log.Warnf("Unable to get IP address for container %s : Failed to inspect container ID %s, error: %s", container.Name, connectedContainer, err)
			return ""
		}
		return p.getIPAddress(parseContainer(containerInspected))
	}

	for _, network := range container.NetworkSettings.Networks {
		return network.Addr
	}

	log.Warnf("Unable to find the IP address for the container %q.", container.Name)
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

// Escape beginning slash "/", convert all others to dash "-", and convert underscores "_" to dash "-"
func getSubDomain(name string) string {
	return strings.NewReplacer("/", "-", "_", "-").Replace(strings.TrimPrefix(name, "/"))
}

func getServiceName(container dockerData) string {
	serviceName := container.ServiceName

	if values, err := getStringMultipleStrict(container.Labels, labelDockerComposeProject, labelDockerComposeService); err == nil {
		serviceName = values[labelDockerComposeService] + "_" + values[labelDockerComposeProject]
	}

	return serviceName
}
