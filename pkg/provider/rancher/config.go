package rancher

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/label"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/constraints"
)

func (p *Provider) buildConfiguration(ctx context.Context, services []rancherData) *dynamic.Configuration {
	configurations := make(map[string]*dynamic.Configuration)

	for _, service := range services {
		logger := log.Ctx(ctx).With().Str("service", service.Name).Logger()
		ctxService := logger.WithContext(ctx)

		if !p.keepService(ctx, service) {
			continue
		}

		confFromLabel, err := label.DecodeConfiguration(service.Labels)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		var tcpOrUDP bool
		if len(confFromLabel.TCP.Routers) > 0 || len(confFromLabel.TCP.Services) > 0 {
			tcpOrUDP = true

			err := p.buildTCPServiceConfiguration(ctxService, service, confFromLabel.TCP)
			if err != nil {
				logger.Error().Err(err).Send()
				continue
			}
			provider.BuildTCPRouterConfiguration(ctxService, confFromLabel.TCP)
		}

		if len(confFromLabel.UDP.Routers) > 0 || len(confFromLabel.UDP.Services) > 0 {
			tcpOrUDP = true

			err := p.buildUDPServiceConfiguration(ctxService, service, confFromLabel.UDP)
			if err != nil {
				logger.Error().Err(err).Send()
				continue
			}
			provider.BuildUDPRouterConfiguration(ctxService, confFromLabel.UDP)
		}

		if tcpOrUDP && len(confFromLabel.HTTP.Routers) == 0 &&
			len(confFromLabel.HTTP.Middlewares) == 0 &&
			len(confFromLabel.HTTP.Services) == 0 {
			configurations[service.Name] = confFromLabel
			continue
		}

		err = p.buildServiceConfiguration(ctx, service, confFromLabel.HTTP)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		model := struct {
			Name   string
			Labels map[string]string
		}{
			Name:   service.Name,
			Labels: service.Labels,
		}

		provider.BuildRouterConfiguration(ctx, confFromLabel.HTTP, service.Name, p.defaultRuleTpl, model)

		configurations[service.Name] = confFromLabel
	}

	return provider.Merge(ctx, configurations)
}

func (p *Provider) buildTCPServiceConfiguration(ctx context.Context, service rancherData, configuration *dynamic.TCPConfiguration) error {
	serviceName := service.Name

	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.TCPService)
		lb := &dynamic.TCPServersLoadBalancer{}
		lb.SetDefaults()
		configuration.Services[serviceName] = &dynamic.TCPService{
			LoadBalancer: lb,
		}
	}

	for _, confService := range configuration.Services {
		err := p.addServerTCP(ctx, service, confService.LoadBalancer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) buildUDPServiceConfiguration(ctx context.Context, service rancherData, configuration *dynamic.UDPConfiguration) error {
	serviceName := service.Name

	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.UDPService)
		lb := &dynamic.UDPServersLoadBalancer{}

		configuration.Services[serviceName] = &dynamic.UDPService{
			LoadBalancer: lb,
		}
	}

	for _, confService := range configuration.Services {
		err := p.addServerUDP(ctx, service, confService.LoadBalancer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) buildServiceConfiguration(ctx context.Context, service rancherData, configuration *dynamic.HTTPConfiguration) error {
	serviceName := service.Name

	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.Service)
		lb := &dynamic.ServersLoadBalancer{}
		lb.SetDefaults()
		configuration.Services[serviceName] = &dynamic.Service{
			LoadBalancer: lb,
		}
	}

	for _, confService := range configuration.Services {
		err := p.addServers(ctx, service, confService.LoadBalancer)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) keepService(ctx context.Context, service rancherData) bool {
	logger := log.Ctx(ctx)

	if !service.ExtraConf.Enable {
		logger.Debug().Msg("Filtering disabled service")
		return false
	}

	matches, err := constraints.MatchLabels(service.Labels, p.Constraints)
	if err != nil {
		logger.Error().Err(err).Msg("Error matching constraint expression")
		return false
	}
	if !matches {
		logger.Debug().Msgf("Service pruned by constraint expression: %q", p.Constraints)
		return false
	}

	if p.EnableServiceHealthFilter {
		if service.Health != "" && service.Health != healthy && service.Health != updatingHealthy {
			logger.Debug().Msgf("Filtering service %s with healthState of %s", service.Name, service.Health)
			return false
		}
		if service.State != "" && service.State != active && service.State != updatingActive && service.State != upgraded && service.State != upgrading {
			logger.Debug().Msgf("Filtering service %s with state of %s", service.Name, service.State)
			return false
		}
	}

	return true
}

func (p *Provider) addServerTCP(ctx context.Context, service rancherData, loadBalancer *dynamic.TCPServersLoadBalancer) error {
	log.Ctx(ctx).Debug().Msgf("Trying to add servers for service %s", service.Name)

	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	if len(loadBalancer.Servers) == 0 {
		loadBalancer.Servers = []dynamic.TCPServer{{}}
	}

	port := loadBalancer.Servers[0].Port
	loadBalancer.Servers[0].Port = ""

	if port == "" {
		port = getServicePort(service)
	}

	if port == "" {
		return errors.New("port is missing")
	}

	var servers []dynamic.TCPServer
	for _, containerIP := range service.Containers {
		servers = append(servers, dynamic.TCPServer{
			Address: net.JoinHostPort(containerIP, port),
		})
	}

	loadBalancer.Servers = servers

	return nil
}

func (p *Provider) addServerUDP(ctx context.Context, service rancherData, loadBalancer *dynamic.UDPServersLoadBalancer) error {
	log.Ctx(ctx).Debug().Msgf("Trying to add servers for service %s", service.Name)

	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	if len(loadBalancer.Servers) == 0 {
		loadBalancer.Servers = []dynamic.UDPServer{{}}
	}

	port := loadBalancer.Servers[0].Port
	loadBalancer.Servers[0].Port = ""

	if port == "" {
		port = getServicePort(service)
	}

	if port == "" {
		return errors.New("port is missing")
	}

	var servers []dynamic.UDPServer
	for _, containerIP := range service.Containers {
		servers = append(servers, dynamic.UDPServer{
			Address: net.JoinHostPort(containerIP, port),
		})
	}

	loadBalancer.Servers = servers

	return nil
}

func (p *Provider) addServers(ctx context.Context, service rancherData, loadBalancer *dynamic.ServersLoadBalancer) error {
	log.Ctx(ctx).Debug().Msgf("Trying to add servers for service %s", service.Name)

	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	if len(loadBalancer.Servers) == 0 {
		server := dynamic.Server{}
		server.SetDefaults()

		loadBalancer.Servers = []dynamic.Server{server}
	}

	port := loadBalancer.Servers[0].Port
	loadBalancer.Servers[0].Port = ""

	if port == "" {
		port = getServicePort(service)
	}

	if port == "" {
		return errors.New("port is missing")
	}

	var servers []dynamic.Server
	for _, containerIP := range service.Containers {
		servers = append(servers, dynamic.Server{
			URL: fmt.Sprintf("%s://%s", loadBalancer.Servers[0].Scheme, net.JoinHostPort(containerIP, port)),
		})
	}

	loadBalancer.Servers = servers

	return nil
}

func getServicePort(data rancherData) string {
	rawPort := strings.Split(data.Port, "/")[0]
	hostPort := strings.Split(rawPort, ":")

	if len(hostPort) >= 2 {
		return hostPort[1]
	}
	if len(hostPort) > 0 && hostPort[0] != "" {
		return hostPort[0]
	}
	return rawPort
}
