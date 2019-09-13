package consulcatalog

import (
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/hashicorp/consul/api"
	"net"
	"strconv"
	"strings"
)

func (p *Provider) buildHTTPRouterConfiguration(name string, tags []string) (string, *dynamic.Router) {
	router := &dynamic.Router{
		EntryPoints: p.Entrypoints,
		Middlewares: nil,
		Service:     name,
		Rule:        p.RouterRule,
		Priority:    0,
		TLS:         nil,
	}

	for _, tag := range tags {
		if strings.HasPrefix(tag, p.prefixes.routerRule) {
			router.Rule = tag[len(p.prefixes.routerRule):]
			continue
		}
		if strings.HasPrefix(tag, p.prefixes.routerEntrypoints) {
			router.EntryPoints = strings.Split(tag[len(p.prefixes.routerEntrypoints):], ",")
			continue
		}
	}

	return name, router
}

func (p *Provider) buildHTTPServiceConfiguration(serviceName string, consulServices []*api.CatalogService) (string, *dynamic.Service) {
	loadBalancer := &dynamic.ServersLoadBalancer{
		Sticky:             nil,
		Servers:            []dynamic.Server{},
		HealthCheck:        nil,
		PassHostHeader:     false,
		ResponseForwarding: nil,
	}
	loadBalancer.SetDefaults()

	for _, consulService := range consulServices {
		server := dynamic.Server{}
		server.SetDefaults()

		server.URL = server.Scheme + "://" + net.JoinHostPort(consulService.ServiceAddress, strconv.Itoa(consulService.ServicePort))

		loadBalancer.Servers = append(loadBalancer.Servers, server)
	}

	service := &dynamic.Service{
		LoadBalancer: loadBalancer,
		Weighted:     nil,
		Mirroring:    nil,
	}

	return serviceName, service
}
