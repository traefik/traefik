package nomad

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/label"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
	"github.com/traefik/traefik/v2/pkg/provider/constraints"
)

func (p *Provider) buildConfig(ctx context.Context, items []item) *dynamic.Configuration {
	configurations := make(map[string]*dynamic.Configuration)

	for _, i := range items {
		normalUnique := provider.Normalize(i.Node + "-" + i.Name + "-" + i.ID)
		ctxSvc := log.With(ctx, log.Str(log.ServiceName, normalUnique))

		if !p.keepItem(ctxSvc, i) {
			continue
		}

		logger := log.FromContext(ctx)
		labels := tagsToLabels(i.Tags, p.Prefix)

		config, decErr := label.DecodeConfiguration(labels)
		if decErr != nil {
			logger.Errorf("failed to decode configuration: %v", decErr)
			continue
		}

		var tcpOrUDP bool

		if len(config.TCP.Routers) > 0 || len(config.TCP.Services) > 0 {
			tcpOrUDP = true
			if buildErr := p.buildTCPConfig(i, config.TCP); buildErr != nil {
				logger.Errorf("failed to build tcp service configuration: %v", buildErr)
				continue
			}
			provider.BuildTCPRouterConfiguration(ctxSvc, config.TCP)
		}

		if len(config.UDP.Routers) > 0 || len(config.UDP.Services) > 0 {
			tcpOrUDP = true
			if buildErr := p.buildUDPConfig(i, config.UDP); buildErr != nil {
				logger.Errorf("failed to build udp service configuration: %v", buildErr)
				continue
			}
			provider.BuildUDPRouterConfiguration(ctxSvc, config.UDP)
		}

		// tcp/udp, skip configuring http service
		if tcpOrUDP && len(config.HTTP.Routers) == 0 &&
			len(config.HTTP.Middlewares) == 0 &&
			len(config.HTTP.Services) == 0 {
			configurations[normalUnique] = config
			continue
		}

		// configure http service
		if buildErr := p.buildServiceConfig(i, config.HTTP); buildErr != nil {
			logger.Errorf("failed to build http service configuration: %v", buildErr)
			continue
		}

		model := struct {
			Name   string
			Labels map[string]string
		}{
			Name:   i.Name,
			Labels: labels,
		}

		provider.BuildRouterConfiguration(ctx, config.HTTP, provider.Normalize(i.Name), p.defaultRuleTpl, model)
		configurations[normalUnique] = config
	}

	return provider.Merge(ctx, configurations)
}

func (p *Provider) buildTCPConfig(i item, configuration *dynamic.TCPConfiguration) error {
	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.TCPService)
		lb := new(dynamic.TCPServersLoadBalancer)
		lb.SetDefaults()
		configuration.Services[provider.Normalize(i.Name)] = &dynamic.TCPService{
			LoadBalancer: lb,
		}
	}

	for _, service := range configuration.Services {
		if err := p.addServerTCP(i, service.LoadBalancer); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) buildUDPConfig(i item, configuration *dynamic.UDPConfiguration) error {
	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.UDPService)
		lb := new(dynamic.UDPServersLoadBalancer)
		configuration.Services[provider.Normalize(i.Name)] = &dynamic.UDPService{
			LoadBalancer: lb,
		}
	}

	for _, service := range configuration.Services {
		if err := p.addServerUDP(i, service.LoadBalancer); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) buildServiceConfig(i item, configuration *dynamic.HTTPConfiguration) error {
	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.Service)
		lb := new(dynamic.ServersLoadBalancer)
		lb.SetDefaults()
		configuration.Services[provider.Normalize(i.Name)] = &dynamic.Service{
			LoadBalancer: lb,
		}
	}

	for _, service := range configuration.Services {
		if err := p.addServer(i, service.LoadBalancer); err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) keepItem(ctx context.Context, i item) bool {
	logger := log.FromContext(ctx)

	if !i.ExtraConf.Enable {
		logger.Debug("Filtering disabled item")
		return false
	}

	matches, err := constraints.MatchTags(i.Tags, p.Constraints)
	if err != nil {
		logger.Errorf("Error matching constraint expressions: %v", err)
		return false
	}
	if !matches {
		logger.Debugf("Filtering out item due to constraints: %q", p.Constraints)
		return false
	}

	// todo filter on health when that information exists (nomad 1.4+)

	return true
}

func (p *Provider) addServerTCP(i item, lb *dynamic.TCPServersLoadBalancer) error {
	if lb == nil {
		return errors.New("load-balancer is missing")
	}

	if len(lb.Servers) == 0 {
		lb.Servers = []dynamic.TCPServer{{}}
	}

	var port string
	if len(lb.Servers) > 0 {
		port = lb.Servers[0].Port
	}

	if i.Port != 0 && port == "" {
		port = strconv.Itoa(i.Port)
	}
	lb.Servers[0].Port = ""

	if port == "" {
		return errors.New("port is missing")
	}

	if i.Address == "" {
		return errors.New("address is missing")
	}

	lb.Servers[0].Address = net.JoinHostPort(i.Address, port)
	return nil
}

func (p *Provider) addServerUDP(i item, lb *dynamic.UDPServersLoadBalancer) error {
	if lb == nil {
		return errors.New("load-balancer is missing")
	}

	if len(lb.Servers) == 0 {
		lb.Servers = []dynamic.UDPServer{{}}
	}

	var port string
	if len(lb.Servers) > 0 {
		port = lb.Servers[0].Port
	}

	if i.Port != 0 && port == "" {
		port = strconv.Itoa(i.Port)
	}
	lb.Servers[0].Port = ""

	if port == "" {
		return errors.New("port is missing")
	}

	if i.Address == "" {
		return errors.New("address is missing")
	}

	lb.Servers[0].Address = net.JoinHostPort(i.Address, port)
	return nil
}

func (p *Provider) addServer(i item, lb *dynamic.ServersLoadBalancer) error {
	if lb == nil {
		return errors.New("load-balancer is missing")
	}

	if len(lb.Servers) == 0 {
		server := dynamic.Server{}
		server.SetDefaults()
		lb.Servers = []dynamic.Server{server}
	}

	var port string
	if len(lb.Servers) > 0 {
		port = lb.Servers[0].Port
	}

	if i.Port != 0 && port == "" {
		port = strconv.Itoa(i.Port)
	}
	lb.Servers[0].Port = ""

	if port == "" {
		return errors.New("port is missing")
	}

	if i.Address == "" {
		return errors.New("address is missing")
	}

	scheme := lb.Servers[0].Scheme
	lb.Servers[0].Scheme = ""
	lb.Servers[0].URL = fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(i.Address, port))

	return nil
}

func tagsToLabels(tags []string, prefix string) map[string]string {
	labels := make(map[string]string, len(tags))
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			if parts := strings.SplitN(tag, "=", 2); len(parts) == 2 {
				left, right := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
				key := "traefik" + strings.TrimPrefix(left, prefix)
				labels[key] = right
			}
		}
	}
	return labels
}
