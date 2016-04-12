package provider

import (
	"reflect"
	"testing"

	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
)

func TestConsulCatalogGetFrontendRule(t *testing.T) {
	provider := &ConsulCatalog{
		Domain: "localhost",
	}

	services := []struct {
		service  serviceUpdate
		expected string
	}{
		{
			service: serviceUpdate{
				ServiceName: "foo",
				Attributes:  []string{},
			},
			expected: "Host:foo.localhost",
		},
		{
			service: serviceUpdate{
				ServiceName: "foo",
				Attributes: []string{
					"traefik.frontend.rule=Host:*.example.com",
				},
			},
			expected: "Host:*.example.com",
		},
	}

	for _, e := range services {
		actual := provider.getFrontendRule(e.service)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestConsulCatalogGetAttribute(t *testing.T) {
	provider := &ConsulCatalog{
		Domain: "localhost",
	}

	services := []struct {
		tags         []string
		key          string
		defaultValue string
		expected     string
	}{
		{
			tags: []string{
				"foo.bar=ramdom",
				"traefik.backend.weight=42",
			},
			key:          "backend.weight",
			defaultValue: "",
			expected:     "42",
		},
		{
			tags: []string{
				"foo.bar=ramdom",
				"traefik.backend.wei=42",
			},
			key:          "backend.weight",
			defaultValue: "",
			expected:     "",
		},
	}

	for _, e := range services {
		actual := provider.getAttribute(e.key, e.tags, e.defaultValue)
		if actual != e.expected {
			t.Fatalf("expected %q, got %q", e.expected, actual)
		}
	}
}

func TestConsulCatalogBuildConfig(t *testing.T) {
	provider := &ConsulCatalog{
		Domain: "localhost",
	}

	cases := []struct {
		nodes             []catalogUpdate
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			nodes:             []catalogUpdate{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			nodes: []catalogUpdate{
				{
					Service: &serviceUpdate{
						ServiceName: "test",
						Attributes:  []string{},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			nodes: []catalogUpdate{
				{
					Service: &serviceUpdate{
						ServiceName: "test",
						Attributes: []string{
							"traefik.backend.loadbalancer=drr",
							"traefik.backend.circuitbreaker=NetworkErrorRatio() > 0.5",
							"random.foo=bar",
						},
					},
					Nodes: []*api.ServiceEntry{
						{
							Service: &api.AgentService{
								Service: "test",
								Address: "127.0.0.1",
								Port:    80,
								Tags: []string{
									"traefik.backend.weight=42",
									"random.foo=bar",
									"traefik.backend.passHostHeader=true",
									"traefik.protocol=https",
								},
							},
							Node: &api.Node{
								Node:    "localhost",
								Address: "127.0.0.1",
							},
						},
					},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-test": {
					Backend: "backend-test",
					Routes: map[string]types.Route{
						"route-host-test": {
							Rule: "Host:test.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"test--127-0-0-1--80": {
							URL:    "https://127.0.0.1:80",
							Weight: 42,
						},
					},
					CircuitBreaker: &types.CircuitBreaker{
						Expression: "NetworkErrorRatio() > 0.5",
					},
					LoadBalancer: &types.LoadBalancer{
						Method: "drr",
					},
				},
			},
		},
	}

	for _, c := range cases {
		actualConfig := provider.buildConfig(c.nodes)
		if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
			t.Fatalf("expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
		}
		if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
			t.Fatalf("expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
		}
	}
}
