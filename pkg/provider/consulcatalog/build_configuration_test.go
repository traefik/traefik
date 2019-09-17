package consulcatalog

import (
	"context"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBuildConfiguration_FromLabels(t *testing.T) {
	p := &Provider{
		Prefix: "traefik",
	}

	data := &consulCatalogData{
		Items: []*consulCatalogItem{
			{
				ID:      "e0008457-7b24-4cdf-a0e2-1caf46091c05",
				Name:    "service1",
				Address: "192.168.1.1",
				Port:    8001,
				Labels: map[string]string{
					"traefik.enable":                                         "true",
					"traefik.http.routers.router1.rule":                      "Host(`example.com`)",
					"traefik.http.routers.router1.entrypoints":               "web",
					"traefik.http.routers.router1.service":                   "service1",
					"traefik.http.services.service1.loadBalancer.server.url": "192.168.1.1:8001",
				},
			},
			{
				ID:      "0948c7c2-eb96-47d7-86b0-4108ac79967a",
				Name:    "service1",
				Address: "192.168.1.1",
				Port:    8002,
				Labels: map[string]string{
					"traefik.enable":                                         "true",
					"traefik.http.routers.router1.rule":                      "Host(`example.com`)",
					"traefik.http.routers.router1.entrypoints":               "web",
					"traefik.http.routers.router1.service":                   "service1",
					"traefik.http.services.service1.loadBalancer.server.url": "192.168.1.1:8002",
				},
			},
		},
	}

	expectedConfig := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers: map[string]*dynamic.Router{
				"router1": {
					EntryPoints: []string{"web"},
					Middlewares: nil,
					Service:     "service1",
					Rule:        "Host(`example.com`)",
					Priority:    0,
					TLS:         nil,
				},
			},
			Middlewares: map[string]*dynamic.Middleware{},
			Services: map[string]*dynamic.Service{
				"service1": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Sticky: nil,
						Servers: []dynamic.Server{
							{URL: "192.168.1.1:8002", Scheme: "http", Port: ""},
							{URL: "192.168.1.1:8001", Scheme: "http", Port: ""},
						},
						HealthCheck:        nil,
						PassHostHeader:     true,
						ResponseForwarding: nil,
					},
					Weighted:  nil,
					Mirroring: nil,
				},
			},
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:  map[string]*dynamic.TCPRouter{},
			Services: map[string]*dynamic.TCPService{},
		},
		TLS: nil,
	}

	cfg, err := p.buildConfiguration(context.Background(), data)
	require.NoError(t, err)
	assert.Equal(t, expectedConfig, cfg)
}

func TestBuildConfiguration_DefaultHTTP(t *testing.T) {
	p := &Provider{
		PassHostHeader: true,
		RouterRule:     "Host(`example.com`)",
		Entrypoints:    []string{"web"},
		Protocol:       "http",
	}

	data := &consulCatalogData{
		Items: []*consulCatalogItem{
			{
				ID:      "e0008457-7b24-4cdf-a0e2-1caf46091c05",
				Name:    "service1",
				Address: "192.168.1.1",
				Port:    8001,
				Labels: map[string]string{
					"traefik.enable": "true",
				},
			},
			{
				ID:      "0948c7c2-eb96-47d7-86b0-4108ac79967a",
				Name:    "service1",
				Address: "192.168.1.1",
				Port:    8002,
				Labels: map[string]string{
					"traefik.enable": "true",
				},
			},
		},
	}

	expectedConfig := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers: map[string]*dynamic.Router{
				"service1-router": {
					EntryPoints: []string{"web"},
					Middlewares: nil,
					Service:     "service1",
					Rule:        "Host(`example.com`)",
					Priority:    0,
					TLS:         nil,
				},
			},
			Middlewares: map[string]*dynamic.Middleware{},
			Services: map[string]*dynamic.Service{
				"service1": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Sticky: nil,
						Servers: []dynamic.Server{
							{URL: "http://192.168.1.1:8002", Scheme: "", Port: ""},
							{URL: "http://192.168.1.1:8001", Scheme: "", Port: ""},
						},
						HealthCheck:        nil,
						PassHostHeader:     true,
						ResponseForwarding: nil,
					},
					Weighted:  nil,
					Mirroring: nil,
				},
			},
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:  map[string]*dynamic.TCPRouter{},
			Services: map[string]*dynamic.TCPService{},
		},
		TLS: nil,
	}

	cfg, err := p.buildConfiguration(context.Background(), data)
	require.NoError(t, err)
	assert.Equal(t, expectedConfig, cfg)
}

func TestBuildConfiguration_DefaultTCP(t *testing.T) {
	p := &Provider{
		PassHostHeader: true,
		RouterRule:     "Host(`example.com`)",
		Entrypoints:    []string{"web"},
		Protocol:       "tcp",
	}

	data := &consulCatalogData{
		Items: []*consulCatalogItem{
			{
				ID:      "e0008457-7b24-4cdf-a0e2-1caf46091c05",
				Name:    "service1",
				Address: "192.168.1.1",
				Port:    8001,
				Labels: map[string]string{
					"traefik.enable": "true",
				},
			},
			{
				ID:      "0948c7c2-eb96-47d7-86b0-4108ac79967a",
				Name:    "service1",
				Address: "192.168.1.1",
				Port:    8002,
				Labels: map[string]string{
					"traefik.enable": "true",
				},
			},
		},
	}

	expectedConfig := &dynamic.Configuration{
		HTTP: &dynamic.HTTPConfiguration{
			Routers:     map[string]*dynamic.Router{},
			Middlewares: map[string]*dynamic.Middleware{},
			Services:    map[string]*dynamic.Service{},
		},
		TCP: &dynamic.TCPConfiguration{
			Routers: map[string]*dynamic.TCPRouter{
				"service1-router": {
					EntryPoints: []string{"web"},
					Service:     "service1",
					Rule:        "Host(`example.com`)",
					TLS:         nil,
				},
			},
			Services: map[string]*dynamic.TCPService{
				"service1": {
					LoadBalancer: &dynamic.TCPLoadBalancerService{
						Servers: []dynamic.TCPServer{
							{Address: "192.168.1.1", Port: "8002"},
							{Address: "192.168.1.1", Port: "8001"},
						},
					},
				},
			},
		},
		TLS: nil,
	}

	cfg, err := p.buildConfiguration(context.Background(), data)
	require.NoError(t, err)
	assert.Equal(t, expectedConfig, cfg)
}
