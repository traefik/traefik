package provider

import (
	"reflect"
	"testing"

	"github.com/emilevauge/traefik/types"
	"github.com/hashicorp/consul/api"
	"fmt"
)

func TestConsulCatalogGetFrontendValue(t *testing.T) {
	provider := &ConsulCatalog{
		Domain: "localhost",
	}

	services := []struct {
		service  string
		expected string
	}{
		{
			service:  "foo",
			expected: "foo.localhost",
		},
	}

	for _, e := range services {
		actual := provider.getFrontendValue(e.service)
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
						Attributes: []string{},
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
							"traefik.loadbalancer=drr",
							"traefik.circuitbreaker=NetworkErrorRatio() > 0.5",
							"random.foo=bar",
						},
					},
					Nodes: []*api.ServiceEntry{
						{
							Service: &api.AgentService{
								Service: "test",
								Port:    80,
								Tags: []string{
									"traefik.weight=42",
									"random.foo=bar",
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
							Rule:  "Host",
							Value: "test.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"server-localhost-80": {
							URL: "http://127.0.0.1:80",
							Weight: 42,
						},
					},
					CircuitBreaker: &types.CircuitBreaker{
						Expression: "NetworkErrorRatio() > 0.5",
					},
					LoadBalancer:   &types.LoadBalancer{
						Method: "drr",
					},
				},
			},
		},
	}

	for _, c := range cases {
		actualConfig := provider.buildConfig(c.nodes)
		if len(actualConfig.Backends) > 0 {
			fmt.Println(actualConfig.Backends["backend-test"].LoadBalancer.Method)
			fmt.Println(actualConfig.Backends["backend-test"].CircuitBreaker.Expression)
			fmt.Println(actualConfig.Backends["backend-test"].Servers["server-localhost-80"].Weight)
		}
		if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
			t.Fatalf("expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
		}
		if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
			t.Fatalf("expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
		}
	}
}
