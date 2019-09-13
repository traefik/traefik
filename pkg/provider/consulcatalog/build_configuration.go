package consulcatalog

import (
	"context"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
)

func (p *Provider) buildConfiguration(ctx context.Context) (*dynamic.Configuration, error) {
	var err error

	httpConfiguration := &dynamic.HTTPConfiguration{
		Routers:     map[string]*dynamic.Router{},
		Services:    map[string]*dynamic.Service{},
		Middlewares: map[string]*dynamic.Middleware{},
	}

	tcpConfiguration := &dynamic.TCPConfiguration{
		Routers:  map[string]*dynamic.TCPRouter{},
		Services: map[string]*dynamic.TCPService{},
	}

	httpConfiguration.Routers, httpConfiguration.Services, tcpConfiguration.Routers, tcpConfiguration.Services, err = p.getConsulServicesData(ctx)
	if err != nil {
		return nil, err
	}

	cfg := &dynamic.Configuration{
		HTTP: httpConfiguration,
		TCP:  tcpConfiguration,
		TLS:  nil,
	}

	return cfg, nil
}
