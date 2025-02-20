package consulcatalog

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"net"
	"sort"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/label"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/provider/constraints"
)

func (p *Provider) buildConfiguration(ctx context.Context, items []itemData, certInfo *connectCert) *dynamic.Configuration {
	configurations := make(map[string]*dynamic.Configuration)

	for _, item := range items {
		svcName := provider.Normalize(item.Node + "-" + item.Name + "-" + item.ID)

		logger := log.Ctx(ctx).With().Str(logs.ServiceName, svcName).Logger()
		ctxSvc := logger.WithContext(ctx)

		if !p.keepContainer(ctxSvc, item) {
			continue
		}

		confFromLabel, err := label.DecodeConfiguration(item.Labels)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		var tcpOrUDP bool
		if len(confFromLabel.TCP.Routers) > 0 || len(confFromLabel.TCP.Services) > 0 {
			tcpOrUDP = true

			if item.ExtraConf.ConsulCatalog.Connect {
				if confFromLabel.TCP.ServersTransports == nil {
					confFromLabel.TCP.ServersTransports = make(map[string]*dynamic.TCPServersTransport)
				}

				serversTransportKey := itemServersTransportKey(item)
				if confFromLabel.TCP.ServersTransports[serversTransportKey] == nil {
					confFromLabel.TCP.ServersTransports[serversTransportKey] = certInfo.tcpServersTransport(item)
				}
			}

			if err := p.buildTCPServiceConfiguration(item, confFromLabel.TCP); err != nil {
				logger.Error().Err(err).Send()
				continue
			}

			provider.BuildTCPRouterConfiguration(ctxSvc, confFromLabel.TCP)
		}

		if len(confFromLabel.UDP.Routers) > 0 || len(confFromLabel.UDP.Services) > 0 {
			tcpOrUDP = true

			if err := p.buildUDPServiceConfiguration(item, confFromLabel.UDP); err != nil {
				logger.Error().Err(err).Send()
				continue
			}
			provider.BuildUDPRouterConfiguration(ctxSvc, confFromLabel.UDP)
		}

		if tcpOrUDP && len(confFromLabel.HTTP.Routers) == 0 &&
			len(confFromLabel.HTTP.Middlewares) == 0 &&
			len(confFromLabel.HTTP.Services) == 0 {
			configurations[svcName] = confFromLabel
			continue
		}

		if item.ExtraConf.ConsulCatalog.Connect {
			if confFromLabel.HTTP.ServersTransports == nil {
				confFromLabel.HTTP.ServersTransports = make(map[string]*dynamic.ServersTransport)
			}

			serversTransportKey := itemServersTransportKey(item)
			if confFromLabel.HTTP.ServersTransports[serversTransportKey] == nil {
				confFromLabel.HTTP.ServersTransports[serversTransportKey] = certInfo.serversTransport(item)
			}
		}

		if err = p.buildServiceConfiguration(item, confFromLabel.HTTP); err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		model := struct {
			Name   string
			Labels map[string]string
		}{
			Name:   item.Name,
			Labels: item.Labels,
		}

		provider.BuildRouterConfiguration(ctx, confFromLabel.HTTP, getName(item), p.defaultRuleTpl, model)

		configurations[svcName] = confFromLabel
	}

	return provider.Merge(ctx, configurations)
}

func (p *Provider) keepContainer(ctx context.Context, item itemData) bool {
	logger := log.Ctx(ctx)

	if !item.ExtraConf.Enable {
		logger.Debug().Msg("Filtering disabled item")
		return false
	}

	if !p.ConnectAware && item.ExtraConf.ConsulCatalog.Connect {
		logger.Debug().Msg("Filtering out Connect aware item, Connect support is not enabled")
		return false
	}

	matches, err := constraints.MatchTags(item.Tags, p.Constraints)
	if err != nil {
		logger.Error().Err(err).Msg("Error matching constraint expressions")
		return false
	}
	if !matches {
		logger.Debug().Msgf("Container pruned by constraint expressions: %q", p.Constraints)
		return false
	}

	if !p.includesHealthStatus(item.Status) {
		logger.Debug().Msgf("Status %q is not included in the configured strictChecks of %q", item.Status, strings.Join(p.StrictChecks, ","))
		return false
	}

	return true
}

func (p *Provider) buildTCPServiceConfiguration(item itemData, configuration *dynamic.TCPConfiguration) error {
	if len(configuration.Services) == 0 {
		configuration.Services = map[string]*dynamic.TCPService{
			getName(item): {
				LoadBalancer: new(dynamic.TCPServersLoadBalancer),
			},
		}
	}

	for name, service := range configuration.Services {
		if err := p.addServerTCP(item, service.LoadBalancer); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}

	return nil
}

func (p *Provider) buildUDPServiceConfiguration(item itemData, configuration *dynamic.UDPConfiguration) error {
	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.UDPService)

		lb := &dynamic.UDPServersLoadBalancer{}

		configuration.Services[getName(item)] = &dynamic.UDPService{
			LoadBalancer: lb,
		}
	}

	for name, service := range configuration.Services {
		if err := p.addServerUDP(item, service.LoadBalancer); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}

	return nil
}

func (p *Provider) buildServiceConfiguration(item itemData, configuration *dynamic.HTTPConfiguration) error {
	if len(configuration.Services) == 0 {
		configuration.Services = make(map[string]*dynamic.Service)

		lb := &dynamic.ServersLoadBalancer{}
		lb.SetDefaults()

		configuration.Services[getName(item)] = &dynamic.Service{
			LoadBalancer: lb,
		}
	}

	for name, service := range configuration.Services {
		if err := p.addServer(item, service.LoadBalancer); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}

	return nil
}

func (p *Provider) addServerTCP(item itemData, loadBalancer *dynamic.TCPServersLoadBalancer) error {
	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	if len(loadBalancer.Servers) == 0 {
		loadBalancer.Servers = []dynamic.TCPServer{{}}
	}

	if item.Address == "" {
		return errors.New("address is missing")
	}

	port := loadBalancer.Servers[0].Port
	loadBalancer.Servers[0].Port = ""

	if port == "" {
		port = item.Port
	}

	if port == "" {
		return errors.New("port is missing")
	}

	if item.Address == "" {
		return errors.New("address is missing")
	}

	if item.ExtraConf.ConsulCatalog.Connect {
		loadBalancer.ServersTransport = itemServersTransportKey(item)
		loadBalancer.Servers[0].TLS = true
	}

	loadBalancer.Servers[0].Address = net.JoinHostPort(item.Address, port)

	return nil
}

func (p *Provider) addServerUDP(item itemData, loadBalancer *dynamic.UDPServersLoadBalancer) error {
	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	if len(loadBalancer.Servers) == 0 {
		loadBalancer.Servers = []dynamic.UDPServer{{}}
	}

	if item.Address == "" {
		return errors.New("address is missing")
	}

	port := loadBalancer.Servers[0].Port
	loadBalancer.Servers[0].Port = ""

	if port == "" {
		port = item.Port
	}

	if port == "" {
		return errors.New("port is missing")
	}

	loadBalancer.Servers[0].Address = net.JoinHostPort(item.Address, port)

	return nil
}

func (p *Provider) addServer(item itemData, loadBalancer *dynamic.ServersLoadBalancer) error {
	if loadBalancer == nil {
		return errors.New("load-balancer is not defined")
	}

	if len(loadBalancer.Servers) == 0 {
		loadBalancer.Servers = []dynamic.Server{{}}
	}

	if item.Address == "" {
		return errors.New("address is missing")
	}

	if loadBalancer.Servers[0].URL != "" {
		if loadBalancer.Servers[0].Scheme != "" || loadBalancer.Servers[0].Port != "" {
			return errors.New("defining scheme or port is not allowed when URL is defined")
		}
		return nil
	}

	port := loadBalancer.Servers[0].Port
	loadBalancer.Servers[0].Port = ""

	if port == "" {
		port = item.Port
	}

	if port == "" {
		return errors.New("port is missing")
	}

	scheme := loadBalancer.Servers[0].Scheme
	loadBalancer.Servers[0].Scheme = ""
	if scheme == "" {
		scheme = "http"
	}

	if item.ExtraConf.ConsulCatalog.Connect {
		loadBalancer.ServersTransport = itemServersTransportKey(item)
		scheme = "https"
	}

	loadBalancer.Servers[0].URL = fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(item.Address, port))

	return nil
}

func itemServersTransportKey(item itemData) string {
	return provider.Normalize("tls-" + item.Namespace + "-" + item.Datacenter + "-" + item.Name)
}

func getName(i itemData) string {
	if !i.ExtraConf.ConsulCatalog.Canary {
		return provider.Normalize(i.Name)
	}

	tags := make([]string, len(i.Tags))
	copy(tags, i.Tags)

	sort.Strings(tags)

	hasher := fnv.New64()
	hasher.Write([]byte(strings.Join(tags, "")))
	return provider.Normalize(fmt.Sprintf("%s-%d", i.Name, hasher.Sum64()))
}

// defaultStrictChecks returns the default healthchecks to allow an upstream to be registered a route for loadbalancers.
func defaultStrictChecks() []string {
	return []string{api.HealthPassing, api.HealthWarning}
}
