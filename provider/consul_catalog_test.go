package provider

import (
	"reflect"
	"testing"

	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
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
					Service: "test",
				},
			},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			nodes: []catalogUpdate{
				{
					Service: "test",
					Nodes: []*api.ServiceEntry{
						{
							Service: &api.AgentService{
								Service: "test",
								Port:    80,
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
						},
					},
					CircuitBreaker: nil,
					LoadBalancer:   nil,
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
