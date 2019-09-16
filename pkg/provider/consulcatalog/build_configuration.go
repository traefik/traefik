package consulcatalog

import (
	"context"
	"fmt"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/config/parser"
	"github.com/containous/traefik/v2/pkg/provider"
	"net"
	"strconv"
)

func (p *Provider) buildConfiguration(ctx context.Context, data *consulCatalogData) (*dynamic.Configuration, error) {
	cfgs := dynamic.Configurations{}

	for _, item := range data.Items {
		cfg := &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{
				Routers:     map[string]*dynamic.Router{},
				Middlewares: map[string]*dynamic.Middleware{},
				Services:    map[string]*dynamic.Service{},
			},
			TCP: &dynamic.TCPConfiguration{
				Routers:  map[string]*dynamic.TCPRouter{},
				Services: map[string]*dynamic.TCPService{},
			},
			TLS: &dynamic.TLSConfiguration{},
		}

		err := parser.Decode(item.Labels, cfg, p.Prefix, p.Prefix+".http", p.Prefix+".tcp")
		if err != nil {
			return nil, fmt.Errorf("error decode labels, %v", err)
		}

		// If empty routers and/or services, build configuration with default values
		if len(cfg.HTTP.Routers) == 0 {
			switch p.Protocol {
			case "http":
				p.buildConfigurationHTTPRouter(ctx, cfg, item)
			case "tcp":
				p.buildConfigurationTCPRouter(ctx, cfg, item)
			}
		}
		if len(cfg.HTTP.Services) == 0 {
			switch p.Protocol {
			case "http":
				p.buildConfigurationHTTPService(ctx, cfg, item)
			case "tcp":
				p.buildConfigurationTCPService(ctx, cfg, item)
			}
		}

		cfgs[item.ID] = cfg
	}

	cfg := provider.Merge(ctx, cfgs)

	return cfg, nil
}

func (p *Provider) buildConfigurationHTTPService(ctx context.Context, cfg *dynamic.Configuration, item *consulCatalogItem) {

	server := dynamic.Server{}
	server.SetDefaults()
	server.URL = server.Scheme + "://" + net.JoinHostPort(item.Address, strconv.Itoa(item.Port))
	server.Scheme = ""

	cfg.HTTP.Services[item.Name] = &dynamic.Service{
		LoadBalancer: &dynamic.ServersLoadBalancer{
			Sticky:             nil,
			Servers:            []dynamic.Server{server},
			HealthCheck:        nil,
			PassHostHeader:     p.PassHostHeader,
			ResponseForwarding: nil,
		},
		Weighted:  nil,
		Mirroring: nil,
	}
}

func (p *Provider) buildConfigurationHTTPRouter(ctx context.Context, cfg *dynamic.Configuration, item *consulCatalogItem) {
	cfg.HTTP.Routers[item.Name+"-router"] = &dynamic.Router{
		EntryPoints: p.Entrypoints,
		Middlewares: p.Middlewares,
		Service:     item.Name,
		Rule:        p.RouterRule,
		Priority:    0,
		TLS:         nil,
	}
}

func (p *Provider) buildConfigurationTCPService(ctx context.Context, cfg *dynamic.Configuration, item *consulCatalogItem) {
	cfg.TCP.Services[item.Name] = &dynamic.TCPService{
		LoadBalancer: &dynamic.TCPLoadBalancerService{
			Servers: []dynamic.TCPServer{
				{Address: item.Address, Port: strconv.Itoa(item.Port)},
			},
		},
	}
}

func (p *Provider) buildConfigurationTCPRouter(ctx context.Context, cfg *dynamic.Configuration, item *consulCatalogItem) {
	cfg.TCP.Routers[item.Name+"-router"] = &dynamic.TCPRouter{
		EntryPoints: p.Entrypoints,
		Service:     item.Name,
		Rule:        p.RouterRule,
		TLS:         nil,
	}
}
