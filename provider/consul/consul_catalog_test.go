package consul

import (
	"reflect"
	"sort"
	"testing"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
)

func TestConsulCatalogGetFrontendRule(t *testing.T) {
	provider := &CatalogProvider{
		Domain:               "localhost",
		Prefix:               "traefik",
		FrontEndRule:         "Host:{{.ServiceName}}.{{.Domain}}",
		frontEndRuleTemplate: template.New("consul catalog frontend rule"),
	}
	provider.setupFrontEndTemplate()

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
		{
			service: serviceUpdate{
				ServiceName: "foo",
				Attributes: []string{
					"traefik.frontend.rule=Host:{{.ServiceName}}.example.com",
				},
			},
			expected: "Host:foo.example.com",
		},
		{
			service: serviceUpdate{
				ServiceName: "foo",
				Attributes: []string{
					"traefik.frontend.rule=PathPrefix:{{getTag \"contextPath\" .Attributes \"/\"}}",
					"contextPath=/bar",
				},
			},
			expected: "PathPrefix:/bar",
		},
	}

	for _, e := range services {
		actual := provider.getFrontendRule(e.service)
		if actual != e.expected {
			t.Fatalf("expected %s, got %s", e.expected, actual)
		}
	}
}

func TestConsulCatalogGetTag(t *testing.T) {
	provider := &CatalogProvider{
		Domain: "localhost",
		Prefix: "traefik",
	}

	services := []struct {
		tags         []string
		key          string
		defaultValue string
		expected     string
	}{
		{
			tags: []string{
				"foo.bar=random",
				"traefik.backend.weight=42",
				"management",
			},
			key:          "foo.bar",
			defaultValue: "0",
			expected:     "random",
		},
	}

	actual := provider.hasTag("management", []string{"management"})
	if !actual {
		t.Fatalf("expected %v, got %v", true, actual)
	}

	actual = provider.hasTag("management", []string{"management=yes"})
	if !actual {
		t.Fatalf("expected %v, got %v", true, actual)
	}

	for _, e := range services {
		actual := provider.getTag(e.key, e.tags, e.defaultValue)
		if actual != e.expected {
			t.Fatalf("expected %s, got %s", e.expected, actual)
		}
	}
}

func TestConsulCatalogGetAttribute(t *testing.T) {
	provider := &CatalogProvider{
		Domain: "localhost",
		Prefix: "traefik",
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
			defaultValue: "0",
			expected:     "42",
		},
		{
			tags: []string{
				"foo.bar=ramdom",
				"traefik.backend.wei=42",
			},
			key:          "backend.weight",
			defaultValue: "0",
			expected:     "0",
		},
	}

	expected := provider.Prefix + ".foo"
	actual := provider.getPrefixedName("foo")
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}

	for _, e := range services {
		actual := provider.getAttribute(e.key, e.tags, e.defaultValue)
		if actual != e.expected {
			t.Fatalf("expected %s, got %s", e.expected, actual)
		}
	}
}

func TestConsulCatalogGetAttributeWithEmptyPrefix(t *testing.T) {
	provider := &CatalogProvider{
		Domain: "localhost",
		Prefix: "",
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
				"backend.weight=42",
			},
			key:          "backend.weight",
			defaultValue: "0",
			expected:     "42",
		},
		{
			tags: []string{
				"foo.bar=ramdom",
				"backend.wei=42",
			},
			key:          "backend.weight",
			defaultValue: "0",
			expected:     "0",
		},
		{
			tags: []string{
				"foo.bar=ramdom",
				"backend.wei=42",
			},
			key:          "foo.bar",
			defaultValue: "random",
			expected:     "ramdom",
		},
	}

	expected := "foo"
	actual := provider.getPrefixedName("foo")
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}

	for _, e := range services {
		actual := provider.getAttribute(e.key, e.tags, e.defaultValue)
		if actual != e.expected {
			t.Fatalf("expected %s, got %s", e.expected, actual)
		}
	}
}

func TestConsulCatalogGetBackendAddress(t *testing.T) {
	provider := &CatalogProvider{
		Domain: "localhost",
		Prefix: "traefik",
	}

	services := []struct {
		node     *api.ServiceEntry
		expected string
	}{
		{
			node: &api.ServiceEntry{
				Node: &api.Node{
					Address: "10.1.0.1",
				},
				Service: &api.AgentService{
					Address: "10.2.0.1",
				},
			},
			expected: "10.2.0.1",
		},
		{
			node: &api.ServiceEntry{
				Node: &api.Node{
					Address: "10.1.0.1",
				},
				Service: &api.AgentService{
					Address: "",
				},
			},
			expected: "10.1.0.1",
		},
	}

	for _, e := range services {
		actual := provider.getBackendAddress(e.node)
		if actual != e.expected {
			t.Fatalf("expected %s, got %s", e.expected, actual)
		}
	}
}

func TestConsulCatalogGetBackendName(t *testing.T) {
	provider := &CatalogProvider{
		Domain: "localhost",
		Prefix: "traefik",
	}

	services := []struct {
		node     *api.ServiceEntry
		expected string
	}{
		{
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{},
				},
			},
			expected: "api--10-0-0-1--80--0",
		},
		{
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{"traefik.weight=42", "traefik.enable=true"},
				},
			},
			expected: "api--10-0-0-1--80--traefik-weight-42--traefik-enable-true--1",
		},
		{
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{"a funny looking tag"},
				},
			},
			expected: "api--10-0-0-1--80--a-funny-looking-tag--2",
		},
	}

	for i, e := range services {
		actual := provider.getBackendName(e.node, i)
		if actual != e.expected {
			t.Fatalf("expected %s, got %s", e.expected, actual)
		}
	}
}

func TestConsulCatalogBuildConfig(t *testing.T) {
	provider := &CatalogProvider{
		Domain:               "localhost",
		Prefix:               "traefik",
		ExposedByDefault:     false,
		FrontEndRule:         "Host:{{.ServiceName}}.{{.Domain}}",
		frontEndRuleTemplate: template.New("consul catalog frontend rule"),
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
							"traefik.backend.maxconn.amount=1000",
							"traefik.backend.maxconn.extractorfunc=client.ip",
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
					Backend:        "backend-test",
					PassHostHeader: true,
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
						"test--127-0-0-1--80--traefik-backend-weight-42--random-foo-bar--traefik-backend-passHostHeader-true--traefik-protocol-https--0": {
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
					MaxConn: &types.MaxConn{
						Amount:        1000,
						ExtractorFunc: "client.ip",
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

func TestConsulCatalogNodeSorter(t *testing.T) {
	cases := []struct {
		nodes    []*api.ServiceEntry
		expected []*api.ServiceEntry
	}{
		{
			nodes:    []*api.ServiceEntry{},
			expected: []*api.ServiceEntry{},
		},
		{
			nodes: []*api.ServiceEntry{
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "127.0.0.1",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.1",
					},
				},
			},
			expected: []*api.ServiceEntry{
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "127.0.0.1",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.1",
					},
				},
			},
		},
		{
			nodes: []*api.ServiceEntry{
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "127.0.0.2",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
				{
					Service: &api.AgentService{
						Service: "bar",
						Address: "127.0.0.2",
						Port:    81,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "127.0.0.1",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.1",
					},
				},
				{
					Service: &api.AgentService{
						Service: "bar",
						Address: "127.0.0.2",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
			},
			expected: []*api.ServiceEntry{
				{
					Service: &api.AgentService{
						Service: "bar",
						Address: "127.0.0.2",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
				{
					Service: &api.AgentService{
						Service: "bar",
						Address: "127.0.0.2",
						Port:    81,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "127.0.0.1",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.1",
					},
				},
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "127.0.0.2",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
			},
		},
		{
			nodes: []*api.ServiceEntry{
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.1",
					},
				},
			},
			expected: []*api.ServiceEntry{
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.1",
					},
				},
				{
					Service: &api.AgentService{
						Service: "foo",
						Address: "",
						Port:    80,
					},
					Node: &api.Node{
						Node:    "localhost",
						Address: "127.0.0.2",
					},
				},
			},
		},
	}

	for _, c := range cases {
		sort.Sort(nodeSorter(c.nodes))
		actual := c.nodes
		if !reflect.DeepEqual(actual, c.expected) {
			t.Fatalf("expected %q, got %q", c.expected, actual)
		}
	}
}

func TestConsulCatalogGetChangedKeys(t *testing.T) {
	type Input struct {
		currState map[string][]string
		prevState map[string][]string
	}

	type Output struct {
		addedKeys   []string
		removedKeys []string
	}

	cases := []struct {
		input  Input
		output Output
	}{
		{
			input: Input{
				currState: map[string][]string{
					"foo-service":    {"v1"},
					"bar-service":    {"v1"},
					"baz-service":    {"v1"},
					"qux-service":    {"v1"},
					"quux-service":   {"v1"},
					"quuz-service":   {"v1"},
					"corge-service":  {"v1"},
					"grault-service": {"v1"},
					"garply-service": {"v1"},
					"waldo-service":  {"v1"},
					"fred-service":   {"v1"},
					"plugh-service":  {"v1"},
					"xyzzy-service":  {"v1"},
					"thud-service":   {"v1"},
				},
				prevState: map[string][]string{
					"foo-service":    {"v1"},
					"bar-service":    {"v1"},
					"baz-service":    {"v1"},
					"qux-service":    {"v1"},
					"quux-service":   {"v1"},
					"quuz-service":   {"v1"},
					"corge-service":  {"v1"},
					"grault-service": {"v1"},
					"garply-service": {"v1"},
					"waldo-service":  {"v1"},
					"fred-service":   {"v1"},
					"plugh-service":  {"v1"},
					"xyzzy-service":  {"v1"},
					"thud-service":   {"v1"},
				},
			},
			output: Output{
				addedKeys:   []string{},
				removedKeys: []string{},
			},
		},
		{
			input: Input{
				currState: map[string][]string{
					"foo-service":    {"v1"},
					"bar-service":    {"v1"},
					"baz-service":    {"v1"},
					"qux-service":    {"v1"},
					"quux-service":   {"v1"},
					"quuz-service":   {"v1"},
					"corge-service":  {"v1"},
					"grault-service": {"v1"},
					"garply-service": {"v1"},
					"waldo-service":  {"v1"},
					"fred-service":   {"v1"},
					"plugh-service":  {"v1"},
					"xyzzy-service":  {"v1"},
					"thud-service":   {"v1"},
				},
				prevState: map[string][]string{
					"foo-service":    {"v1"},
					"bar-service":    {"v1"},
					"baz-service":    {"v1"},
					"corge-service":  {"v1"},
					"grault-service": {"v1"},
					"garply-service": {"v1"},
					"waldo-service":  {"v1"},
					"fred-service":   {"v1"},
					"plugh-service":  {"v1"},
					"xyzzy-service":  {"v1"},
					"thud-service":   {"v1"},
				},
			},
			output: Output{
				addedKeys:   []string{"qux-service", "quux-service", "quuz-service"},
				removedKeys: []string{},
			},
		},
		{
			input: Input{
				currState: map[string][]string{
					"foo-service":    {"v1"},
					"qux-service":    {"v1"},
					"quux-service":   {"v1"},
					"quuz-service":   {"v1"},
					"corge-service":  {"v1"},
					"grault-service": {"v1"},
					"garply-service": {"v1"},
					"waldo-service":  {"v1"},
					"fred-service":   {"v1"},
					"plugh-service":  {"v1"},
					"xyzzy-service":  {"v1"},
					"thud-service":   {"v1"},
				},
				prevState: map[string][]string{
					"foo-service":   {"v1"},
					"bar-service":   {"v1"},
					"baz-service":   {"v1"},
					"qux-service":   {"v1"},
					"quux-service":  {"v1"},
					"quuz-service":  {"v1"},
					"corge-service": {"v1"},
					"waldo-service": {"v1"},
					"fred-service":  {"v1"},
					"plugh-service": {"v1"},
					"xyzzy-service": {"v1"},
					"thud-service":  {"v1"},
				},
			},
			output: Output{
				addedKeys:   []string{"grault-service", "garply-service"},
				removedKeys: []string{"bar-service", "baz-service"},
			},
		},
	}

	for _, c := range cases {
		addedKeys, removedKeys := getChangedKeys(c.input.currState, c.input.prevState)

		if !reflect.DeepEqual(fun.Set(addedKeys), fun.Set(c.output.addedKeys)) {
			t.Fatalf("Added keys comparison results: got %q, want %q", addedKeys, c.output.addedKeys)
		}

		if !reflect.DeepEqual(fun.Set(removedKeys), fun.Set(c.output.removedKeys)) {
			t.Fatalf("Removed keys comparison results: got %q, want %q", removedKeys, c.output.removedKeys)
		}
	}
}

func TestConsulCatalogFilterEnabled(t *testing.T) {
	cases := []struct {
		desc             string
		exposedByDefault bool
		node             *api.ServiceEntry
		expected         bool
	}{
		{
			desc:             "exposed",
			exposedByDefault: true,
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{""},
				},
			},
			expected: true,
		},
		{
			desc:             "exposed and tolerated by valid label value",
			exposedByDefault: true,
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{"", "traefik.enable=true"},
				},
			},
			expected: true,
		},
		{
			desc:             "exposed and tolerated by invalid label value",
			exposedByDefault: true,
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{"", "traefik.enable=bad"},
				},
			},
			expected: true,
		},
		{
			desc:             "exposed but overridden by label",
			exposedByDefault: true,
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{"", "traefik.enable=false"},
				},
			},
			expected: false,
		},
		{
			desc:             "non-exposed",
			exposedByDefault: false,
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{""},
				},
			},
			expected: false,
		},
		{
			desc:             "non-exposed but overridden by label",
			exposedByDefault: false,
			node: &api.ServiceEntry{
				Service: &api.AgentService{
					Service: "api",
					Address: "10.0.0.1",
					Port:    80,
					Tags:    []string{"", "traefik.enable=true"},
				},
			},
			expected: true,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			provider := &CatalogProvider{
				Domain:           "localhost",
				Prefix:           "traefik",
				ExposedByDefault: c.exposedByDefault,
			}
			if provider.nodeFilter("test", c.node) != c.expected {
				t.Errorf("got unexpected filtering = %t", !c.expected)
			}
		})
	}
}
