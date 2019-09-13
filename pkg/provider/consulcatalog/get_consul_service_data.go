package consulcatalog

import (
	"context"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/hashicorp/consul/api"
)

func (p *Provider) getConsulServicesData(ctx context.Context) (map[string]*dynamic.Router, map[string]*dynamic.Service, map[string]*dynamic.TCPRouter, map[string]*dynamic.TCPService, error) {
	httpRouters := make(map[string]*dynamic.Router)
	httpServices := make(map[string]*dynamic.Service)
	tcpRouters := make(map[string]*dynamic.TCPRouter)
	tcpServices := make(map[string]*dynamic.TCPService)

	var consulServiceNames map[string][]string
	var consulServices []*api.CatalogService
	var err error

	consulServiceNames, err = p.serviceNames(ctx)
	if err != nil {
		return httpRouters, httpServices, tcpRouters, tcpServices, err
	}

	for name, tags := range consulServiceNames {
		protocol := p.Protocol

		if value, ok := inArrayPrefix(p.prefixes.protocol, tags); ok {
			protocol = value
		}

		if err = validateProtocol(protocol); err != nil {
			return httpRouters, httpServices, tcpRouters, tcpServices, err
		}

		queryService := &api.QueryOptions{
			//Datacenter:        "",
			//AllowStale:        false,
			//RequireConsistent: false,
			//UseCache:          false,
			//MaxAge:            0,
			//StaleIfError:      0,
			//WaitIndex:         0,
			//WaitHash:          "",
			//WaitTime:          0,
			//Token:             "",
			//Near:              "",
			//NodeMeta:          nil,
			//RelayFactor:       0,
			//Connect:           false,
		}

		consulServices, _, err = p.clientCatalog.Service(name, "", queryService)
		if err != nil {
			return httpRouters, httpServices, tcpRouters, tcpServices, err
		}

		switch protocol {
		case "http":
			routerName, routerConfig := p.buildHTTPRouterConfiguration(name, tags)
			httpRouters[routerName] = routerConfig
			serviceName, serviceConfig := p.buildHTTPServiceConfiguration(name, consulServices)
			httpServices[serviceName] = serviceConfig
		case "tcp":
			routerName, routerConfig := p.buildTCPRouterConfiguration(name, tags)
			tcpRouters[routerName] = routerConfig
			serviceName, serviceConfig := p.buildTCPServiceConfiguration(name, consulServices)
			tcpServices[serviceName] = serviceConfig
		}
	}

	return httpRouters, httpServices, tcpRouters, tcpServices, err
}

func (p *Provider) serviceNames(ctx context.Context) (map[string][]string, error) {

	queryServiceNames := &api.QueryOptions{
		//Datacenter:        "",
		//AllowStale:        false,
		//RequireConsistent: false,
		//UseCache:          false,
		//MaxAge:            0,
		//StaleIfError:      0,
		//WaitIndex:         0,
		//WaitHash:          "",
		//WaitTime:          0,
		//Token:             "",
		//Near:              "",
		//NodeMeta:          nil,
		//RelayFactor:       0,
		//Connect:           false,
	}

	serviceNames, _, err := p.clientCatalog.Services(queryServiceNames)
	if err != nil {
		return nil, err
	}

	for name, tags := range serviceNames {
		if !inArray(p.prefixes.enabled, tags) {
			delete(serviceNames, name)
		}
	}

	return serviceNames, nil
}
