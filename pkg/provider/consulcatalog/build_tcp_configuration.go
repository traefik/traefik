package consulcatalog

import (
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/hashicorp/consul/api"
	"net"
	"strconv"
	"strings"
)

func (p *Provider) buildTCPRouterConfiguration(name string, tags []string) (string, *dynamic.TCPRouter) {
	router := &dynamic.TCPRouter{
		EntryPoints: p.Entrypoints,
		Service:     name,
		Rule:        p.RouterRule,
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

func (p *Provider) buildTCPServiceConfiguration(serviceName string, consulServices []*api.CatalogService) (string, *dynamic.TCPService) {
	loadBalancer := &dynamic.TCPLoadBalancerService{
		Servers: []dynamic.TCPServer{},
	}

	for _, consulService := range consulServices {
		server := dynamic.TCPServer{}

		server.Address = net.JoinHostPort(consulService.ServiceAddress, strconv.Itoa(consulService.ServicePort))

		loadBalancer.Servers = append(loadBalancer.Servers, server)
	}

	service := &dynamic.TCPService{
		LoadBalancer: loadBalancer,
	}

	return serviceName, service
}
