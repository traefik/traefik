package provider

import (
	"reflect"
	"testing"

	"github.com/emilevauge/traefik/mocks"
	"github.com/emilevauge/traefik/types"
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
		kv:     &mocks.MockConsulKV{},
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

func TestKvGetProperties(t *testing.T) {
	var p *ConsulCatalog = &ConsulCatalog{
		Domain: "localhost",
		kv:     &mocks.MockConsulKV{},
	}

	actual := p.getKV("defaultValue", "foo", "bar", "anything")
	if actual != "defaultValue" {
		t.Fatalf("expected \"defaultValue\", got %v", actual)
	}
	actual = p.getKV("", "foo", "bar", "anything")
	if actual != "" {
		t.Fatalf("expected \"\" (empty string), got %v", actual)
	}

	cases := []struct {
		provider     *ConsulCatalog
		defaultValue string
		service_name string
		property     string
		expected     string
	}{
		{
			provider: &ConsulCatalog{
				Prefix: "/traefik",
				kv:     &mocks.MockConsulKV{},
			},
			defaultValue: "42",
			service_name: "foo",
			property:     "weight",
			expected:     "42",
		},
		{
			provider: &ConsulCatalog{
				Prefix: "/traefik",
				kv: &mocks.MockConsulKV{
					KVPairs: []*api.KVPair{
						{
							Key:   "traefik/foo/weight",
							Value: []byte("bar"),
						},
					},
				},
			},
			defaultValue: "",
			service_name: "bar",
			property:     "weight",
			expected:     "",
		},
		{
			provider: &ConsulCatalog{
				Prefix: "/traefik",
				kv: &mocks.MockConsulKV{
					KVPairs: []*api.KVPair{
						{
							Key:   "traefik/foo/weight",
							Value: []byte("bar"),
						},
					},
				},
			},
			defaultValue: "",
			service_name: "foo",
			property:     "weight",
			expected:     "bar",
		},
		{
			provider: &ConsulCatalog{
				Prefix: "/traefik",
				kv: &mocks.MockConsulKV{
					KVPairs: []*api.KVPair{
						{
							Key:   "traefik/foo/baz/1",
							Value: []byte("bar1"),
						},
						{
							Key:   "traefik/foo/baz/2",
							Value: []byte("bar2"),
						},
						{
							Key:   "traefik/foo/baz/biz/1",
							Value: []byte("bar3"),
						},
					},
				},
			},
			defaultValue: "",
			service_name: "foo/baz",
			property:     "2",
			expected:     "bar2",
		},
	}

	for _, c := range cases {
		actual := c.provider.getKV(c.defaultValue, c.service_name, "", c.property)
		if actual != c.expected {
			t.Fatalf("expected %v, got '%v' with default value == '%v', service name == '%v' and property == '%v'", c.expected, actual, c.defaultValue, c.service_name, c.property)
		}
	}
}
