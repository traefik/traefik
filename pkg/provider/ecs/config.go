package ecs

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/label"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/provider/constraints"
)

func (p *Provider) buildConfiguration(ctx context.Context, instances []ecsInstance) *dynamic.Configuration {
	configurations := make(map[string]*dynamic.Configuration)

	for _, instance := range instances {
		instanceName := getServiceName(instance) + "-" + instance.ID
		logger := log.Ctx(ctx).With().Str("ecs-instance", instanceName).Logger()
		ctxContainer := logger.WithContext(ctx)

		if !p.filterInstance(ctxContainer, instance) {
			continue
		}

		confFromLabel, err := label.DecodeConfiguration(instance.Labels)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		var tcpOrUDP bool
		if len(confFromLabel.TCP.Routers) > 0 || len(confFromLabel.TCP.Services) > 0 {
			tcpOrUDP = true

			err := p.buildTCPServiceConfiguration(instance, confFromLabel.TCP)
			if err != nil {
				logger.Error().Err(err).Send()
				continue
			}
			provider.BuildTCPRouterConfiguration(ctxContainer, confFromLabel.TCP)
		}

		if len(confFromLabel.UDP.Routers) > 0 || len(confFromLabel.UDP.Services) > 0 {
			tcpOrUDP = true

			err := p.buildUDPServiceConfiguration(instance, confFromLabel.UDP)
			if err != nil {
				logger.Error().Err(err).Send()
				continue
			}
			provider.BuildUDPRouterConfiguration(ctxContainer, confFromLabel.UDP)
		}

		if tcpOrUDP && len(confFromLabel.HTTP.Routers) == 0 &&
			len(confFromLabel.HTTP.Middlewares) == 0 &&
			len(confFromLabel.HTTP.Services) == 0 {
			configurations[instanceName] = confFromLabel
			continue
		}

		err = p.buildServiceConfiguration(ctxContainer, instance, confFromLabel.HTTP)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		serviceName := getServiceName(instance)

		model := struct {
			Name   string
			Labels map[string]string
		}{
			Name:   serviceName,
			Labels: instance.Labels,
		}

		provider.BuildRouterConfiguration(ctx, confFromLabel.HTTP, serviceName, p.defaultRuleTpl, model)

		configurations[instanceName] = confFromLabel
	}

	return provider.Merge(ctx, configurations)
}

func (p *Provider) buildTCPServiceConfiguration(instance ecsInstance, configuration *dynamic.TCPConfiguration) error {
	serviceName := getServiceName(instance)

	if len(configuration.Services) == 0 {
		configuration.Services = map[string]*dynamic.TCPService{
			serviceName: {
				LoadBalancer: new(dynamic.TCPServersLoadBalancer),
			},
		}
	}

	for name, service := range configuration.Services {
		err := p.addServerTCP(instance, service.LoadBalancer)
		if err != nil {
			return fmt.Errorf("service %q error: %w", name, err)
		}
	}

	return nil
}

func (p *Provider) buildUDPServiceConfiguration(instance ecsInstance, configuration *dynamic.UDPConfiguration) error {
	serviceName := getServiceName(instance)

	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.UDPService)
		lb := &dynamic.UDPServersLoadBalancer{}
		configuration.Services[serviceName] = &dynamic.UDPService{
			LoadBalancer: lb,
		}
	}

	for name, service := range configuration.Services {
		err := p.addServerUDP(instance, service.LoadBalancer)
		if err != nil {
			return fmt.Errorf("service %q error: %w", name, err)
		}
	}

	return nil
}

func (p *Provider) buildServiceConfiguration(_ context.Context, instance ecsInstance, configuration *dynamic.HTTPConfiguration) error {
	serviceName := getServiceName(instance)

	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.Service)
		lb := &dynamic.ServersLoadBalancer{}
		lb.SetDefaults()
		configuration.Services[serviceName] = &dynamic.Service{
			LoadBalancer: lb,
		}
	}

	for name, service := range configuration.Services {
		err := p.addServer(instance, service.LoadBalancer)
		if err != nil {
			return fmt.Errorf("service %q error: %w", name, err)
		}
	}

	return nil
}

func (p *Provider) filterInstance(ctx context.Context, instance ecsInstance) bool {
	logger := log.Ctx(ctx)

	if instance.machine == nil {
		logger.Debug().Msg("Filtering ecs instance with nil machine")
		return false
	}

	if instance.machine.state != ec2types.InstanceStateNameRunning {
		logger.Debug().Msgf("Filtering ecs instance with an incorrect state %s (%s) (state = %s)", instance.Name, instance.ID, instance.machine.state)
		return false
	}

	if instance.machine.healthStatus == ecstypes.HealthStatusUnhealthy {
		logger.Debug().Msgf("Filtering unhealthy ecs instance %s (%s)", instance.Name, instance.ID)
		return false
	}

	if len(instance.machine.privateIP) == 0 {
		logger.Debug().Msgf("Filtering ecs instance without an ip address %s (%s)", instance.Name, instance.ID)
		return false
	}

	if !instance.ExtraConf.Enable {
		logger.Debug().Msgf("Filtering disabled ecs instance %s (%s)", instance.Name, instance.ID)
		return false
	}

	matches, err := constraints.MatchLabels(instance.Labels, p.Constraints)
	if err != nil {
		logger.Error().Err(err).Msg("Error matching constraint expression")
		return false
	}
	if !matches {
		logger.Debug().Msgf("Container pruned by constraint expression: %q", p.Constraints)
		return false
	}

	return true
}

func (p *Provider) addServerTCP(instance ecsInstance, loadBalancer *dynamic.TCPServersLoadBalancer) error {
	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	if len(loadBalancer.Servers) == 0 {
		loadBalancer.Servers = []dynamic.TCPServer{{}}
	}

	serverPort := loadBalancer.Servers[0].Port
	loadBalancer.Servers[0].Port = ""

	ip, port, err := p.getIPPort(instance, serverPort)
	if err != nil {
		return err
	}

	if port == "" {
		return errors.New("port is missing")
	}

	loadBalancer.Servers[0].Address = net.JoinHostPort(ip, port)

	return nil
}

func (p *Provider) addServerUDP(instance ecsInstance, loadBalancer *dynamic.UDPServersLoadBalancer) error {
	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	if len(loadBalancer.Servers) == 0 {
		loadBalancer.Servers = []dynamic.UDPServer{{}}
	}

	serverPort := loadBalancer.Servers[0].Port
	loadBalancer.Servers[0].Port = ""

	ip, port, err := p.getIPPort(instance, serverPort)
	if err != nil {
		return err
	}

	if port == "" {
		return errors.New("port is missing")
	}

	loadBalancer.Servers[0].Address = net.JoinHostPort(ip, port)

	return nil
}

func (p *Provider) addServer(instance ecsInstance, loadBalancer *dynamic.ServersLoadBalancer) error {
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

	ip, port, err := p.getIPPort(instance, serverPort)
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

func (p *Provider) getIPPort(instance ecsInstance, serverPort string) (string, string, error) {
	var ip, port string

	ip = instance.machine.privateIP
	port = getPort(instance, serverPort)
	if len(ip) == 0 {
		return "", "", fmt.Errorf("unable to find the IP address for the instance %q: the server is ignored", instance.Name)
	}

	return ip, port, nil
}

func getPort(instance ecsInstance, serverPort string) string {
	if len(serverPort) > 0 {
		for _, port := range instance.machine.ports {
			containerPort := strconv.FormatInt(int64(port.containerPort), 10)
			if serverPort == containerPort {
				return strconv.FormatInt(int64(port.hostPort), 10)
			}
		}

		return serverPort
	}

	var ports []nat.Port
	for _, port := range instance.machine.ports {
		natPort, err := nat.NewPort(string(port.protocol), strconv.FormatInt(int64(port.hostPort), 10))
		if err != nil {
			continue
		}

		ports = append(ports, natPort)
	}

	less := func(i, j nat.Port) bool {
		return i.Int() < j.Int()
	}
	nat.Sort(ports, less)

	if len(ports) > 0 {
		return ports[0].Port()
	}

	return ""
}

func getServiceName(instance ecsInstance) string {
	return provider.Normalize(instance.Name)
}
