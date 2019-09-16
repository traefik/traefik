package consulcatalog

import (
	"context"
	"github.com/hashicorp/consul/api"
)

type consulCatalogItem struct {
	ID      string
	Name    string
	Address string
	Port    int
	Labels  map[string]string
}
type consulCatalogData struct {
	Items []*consulCatalogItem
}

func (p *Provider) getConsulServicesData(ctx context.Context) (*consulCatalogData, error) {
	data := &consulCatalogData{}

	consulServiceNames, err := p.getConsulServicesNames(ctx)
	if err != nil {
		return nil, err
	}

	for name := range consulServiceNames {
		consulServices, err := p.getConsulServices(ctx, name)
		if err != nil {
			return nil, err
		}

		for _, consulService := range consulServices {
			labels, err := convertLabels(consulService.ServiceTags)
			if err != nil {
				return nil, err
			}
			item := &consulCatalogItem{
				ID:      consulService.ServiceID,
				Name:    consulService.ServiceName,
				Address: consulService.ServiceAddress,
				Port:    consulService.ServicePort,
				Labels:  labels,
			}

			data.Items = append(data.Items, item)
		}
	}
	return data, nil
}

func (p *Provider) getConsulServices(ctx context.Context, name string) ([]*api.CatalogService, error) {
	tagFilter := ""
	if !p.ExposedByDefault {
		tagFilter = p.labelEnabled
	}

	consulServices, _, err := p.clientCatalog.Service(name, tagFilter, p.getQueryOptions())

	return consulServices, err
}

func (p *Provider) getConsulServicesNames(ctx context.Context) (map[string][]string, error) {
	serviceNames, _, err := p.clientCatalog.Services(p.getQueryOptions())

	return serviceNames, err
}

func (p *Provider) getQueryOptions() *api.QueryOptions {
	return &api.QueryOptions{
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
}
