package nomad

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/label"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/provider/constraints"
)

func (p *Provider) buildConfig(ctx context.Context, items []item) *dynamic.Configuration {
	configurations := make(map[string]*dynamic.Configuration)

	for _, i := range items {
		svcName := provider.Normalize(i.Node + "-" + i.Name + "-" + i.ID)
		logger := log.Ctx(ctx).With().Str(logs.ServiceName, svcName).Logger()
		ctxSvc := logger.WithContext(ctx)

		if !p.keepItem(ctxSvc, i) {
			continue
		}

		labels := tagsToLabels(i.Tags, p.Prefix)

		config, err := label.DecodeConfiguration(labels)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to decode configuration")
			continue
		}

		var tcpOrUDP bool

		if len(config.TCP.Routers) > 0 || len(config.TCP.Services) > 0 {
			tcpOrUDP = true
			if err := p.buildTCPConfig(i, config.TCP); err != nil {
				logger.Error().Err(err).Msg("Failed to build TCP service configuration")
				continue
			}
			provider.BuildTCPRouterConfiguration(ctxSvc, config.TCP)
		}

		if len(config.UDP.Routers) > 0 || len(config.UDP.Services) > 0 {
			tcpOrUDP = true
			if err := p.buildUDPConfig(i, config.UDP); err != nil {
				logger.Error().Err(err).Msg("Failed to build UDP service configuration")
				continue
			}
			provider.BuildUDPRouterConfiguration(ctxSvc, config.UDP)
		}

		// tcp/udp, skip configuring http service
		if tcpOrUDP && len(config.HTTP.Routers) == 0 &&
			len(config.HTTP.Middlewares) == 0 &&
			len(config.HTTP.Services) == 0 {
			configurations[svcName] = config
			continue
		}

		// configure http service
		if err := p.buildServiceConfig(i, config.HTTP); err != nil {
			logger.Error().Err(err).Msg("Failed to build HTTP service configuration")
			continue
		}

		model := struct {
			Name   string
			Labels map[string]string
		}{
			Name:   i.Name,
			Labels: labels,
		}

		provider.BuildRouterConfiguration(ctx, config.HTTP, getName(i), p.defaultRuleTpl, model)
		configurations[svcName] = config
	}

	return provider.Merge(ctx, configurations)
}

func (p *Provider) buildTCPConfig(i item, configuration *dynamic.TCPConfiguration) error {
	if len(configuration.Services) == 0 {
		configuration.Services = map[string]*dynamic.TCPService{
			getName(i): {
				LoadBalancer: new(dynamic.TCPServersLoadBalancer),
			},
		}
	}

	for _, service := range configuration.Services {
		// Leave load balancer empty when no address and allowEmptyServices = true
		if !(i.Address == "" && p.AllowEmptyServices) {
			if err := p.addServerTCP(i, service.LoadBalancer); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Provider) buildUDPConfig(i item, configuration *dynamic.UDPConfiguration) error {
	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.UDPService)

		configuration.Services[getName(i)] = &dynamic.UDPService{
			LoadBalancer: new(dynamic.UDPServersLoadBalancer),
		}
	}

	for _, service := range configuration.Services {
		// Leave load balancer empty when no address and allowEmptyServices = true
		if !(i.Address == "" && p.AllowEmptyServices) {
			if err := p.addServerUDP(i, service.LoadBalancer); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Provider) buildServiceConfig(i item, configuration *dynamic.HTTPConfiguration) error {
	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.Service)

		lb := new(dynamic.ServersLoadBalancer)
		lb.SetDefaults()

		configuration.Services[getName(i)] = &dynamic.Service{
			LoadBalancer: lb,
		}
	}

	for _, service := range configuration.Services {
		// Leave load balancer empty when no address and allowEmptyServices = true
		if !(i.Address == "" && p.AllowEmptyServices) {
			if err := p.addServer(i, service.LoadBalancer); err != nil {
				return err
			}
		}
	}

	return nil
}

// TODO: check whether it is mandatory to filter again.
func (p *Provider) keepItem(ctx context.Context, i item) bool {
	logger := log.Ctx(ctx)

	if !i.ExtraConf.Enable {
		logger.Debug().Msg("Filtering disabled item")
		return false
	}

	matches, err := constraints.MatchTags(i.Tags, p.Constraints)
	if err != nil {
		logger.Error().Err(err).Msg("Error matching constraint expressions")
		return false
	}
	if !matches {
		logger.Debug().Msgf("Filtering out item due to constraints: %q", p.Constraints)
		return false
	}

	// TODO: filter on health when that information exists (nomad 1.4+)

	return true
}

func (p *Provider) addServerTCP(i item, lb *dynamic.TCPServersLoadBalancer) error {
	if lb == nil {
		return errors.New("load-balancer is missing")
	}

	if len(lb.Servers) == 0 {
		lb.Servers = []dynamic.TCPServer{{}}
	}

	if i.Address == "" {
		return errors.New("address is missing")
	}

	port := lb.Servers[0].Port
	lb.Servers[0].Port = ""

	if port == "" && i.Port > 0 {
		port = strconv.Itoa(i.Port)
	}

	if port == "" {
		return errors.New("port is missing")
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

	if i.Address == "" {
		return errors.New("address is missing")
	}

	port := lb.Servers[0].Port
	lb.Servers[0].Port = ""

	if port == "" && i.Port > 0 {
		port = strconv.Itoa(i.Port)
	}

	if port == "" {
		return errors.New("port is missing")
	}

	lb.Servers[0].Address = net.JoinHostPort(i.Address, port)

	return nil
}

func (p *Provider) addServer(i item, lb *dynamic.ServersLoadBalancer) error {
	if lb == nil {
		return errors.New("load-balancer is missing")
	}

	if len(lb.Servers) == 0 {
		lb.Servers = []dynamic.Server{{}}
	}

	if i.Address == "" {
		return errors.New("address is missing")
	}

	if lb.Servers[0].URL != "" {
		if lb.Servers[0].Scheme != "" || lb.Servers[0].Port != "" {
			return errors.New("defining scheme or port is not allowed when URL is defined")
		}
		return nil
	}

	port := lb.Servers[0].Port
	lb.Servers[0].Port = ""

	if port == "" && i.Port > 0 {
		port = strconv.Itoa(i.Port)
	}

	if port == "" {
		return errors.New("port is missing")
	}

	scheme := lb.Servers[0].Scheme
	lb.Servers[0].Scheme = ""
	if scheme == "" {
		scheme = "http"
	}

	lb.Servers[0].URL = fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(i.Address, port))

	return nil
}

func getName(i item) string {
	if !i.ExtraConf.Canary {
		return provider.Normalize(i.Name)
	}

	tags := make([]string, len(i.Tags))
	copy(tags, i.Tags)

	sort.Strings(tags)

	hasher := fnv.New64()
	hasher.Write([]byte(strings.Join(tags, "")))
	return provider.Normalize(fmt.Sprintf("%s-%d", i.Name, hasher.Sum64()))
}
