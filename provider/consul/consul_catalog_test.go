package consul

import (
	"sort"
	"testing"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func TestGetPrefixedName(t *testing.T) {
	testCases := []struct {
		desc     string
		name     string
		prefix   string
		expected string
	}{
		{
			desc:     "empty name with prefix",
			name:     "",
			prefix:   "foo",
			expected: "",
		},
		{
			desc:     "empty name without prefix",
			name:     "",
			prefix:   "",
			expected: "",
		},
		{
			desc:     "with prefix",
			name:     "bar",
			prefix:   "foo",
			expected: "foo.bar",
		},
		{
			desc:     "without prefix",
			name:     "bar",
			prefix:   "",
			expected: "bar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			pro := &CatalogProvider{Prefix: test.prefix}

			actual := pro.getPrefixedName(test.name)
			assert.Equal(t, test.expected, actual)
		})
	}

}

func TestGetFrontendRule(t *testing.T) {

	testCases := []struct {
		desc     string
		service  serviceUpdate
		expected string
	}{
		{
			desc: "Should return default host foo.localhost",
			service: serviceUpdate{
				ServiceName: "foo",
				Attributes:  []string{},
			},
			expected: "Host:foo.localhost",
		},
		{
			desc: "Should return host *.example.com",
			service: serviceUpdate{
				ServiceName: "foo",
				Attributes: []string{
					"traefik.frontend.rule=Host:*.example.com",
				},
			},
			expected: "Host:*.example.com",
		},
		{
			desc: "Should return host foo.example.com",
			service: serviceUpdate{
				ServiceName: "foo",
				Attributes: []string{
					"traefik.frontend.rule=Host:{{.ServiceName}}.example.com",
				},
			},
			expected: "Host:foo.example.com",
		},
		{
			desc: "Should return path prefix /bar",
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

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			provider := &CatalogProvider{
				Domain:               "localhost",
				Prefix:               "traefik",
				FrontEndRule:         "Host:{{.ServiceName}}.{{.Domain}}",
				frontEndRuleTemplate: template.New("consul catalog frontend rule"),
			}
			provider.setupFrontEndTemplate()

			actual := provider.getFrontendRule(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetTag(t *testing.T) {
	testCases := []struct {
		desc         string
		tags         []string
		key          string
		defaultValue string
		expected     string
	}{
		{
			desc: "Should return value of foo.bar key",
			tags: []string{
				"foo.bar=random",
				"traefik.backend.weight=42",
				"management",
			},
			key:          "foo.bar",
			defaultValue: "0",
			expected:     "random",
		},
		{
			desc: "Should return default value when nonexistent key",
			tags: []string{
				"foo.bar.foo.bar=random",
				"traefik.backend.weight=42",
				"management",
			},
			key:          "foo.bar",
			defaultValue: "0",
			expected:     "0",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getTag(test.key, test.tags, test.defaultValue)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestHasTag(t *testing.T) {
	testCases := []struct {
		desc     string
		name     string
		tags     []string
		expected bool
	}{
		{
			desc:     "tag without value",
			name:     "foo",
			tags:     []string{"foo"},
			expected: true,
		},
		{
			desc:     "tag with value",
			name:     "foo",
			tags:     []string{"foo=true"},
			expected: true,
		},
		{
			desc:     "missing tag",
			name:     "foo",
			tags:     []string{"foobar=true"},
			expected: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := hasTag(test.name, test.tags)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetAttribute(t *testing.T) {
	provider := &CatalogProvider{
		Domain: "localhost",
		Prefix: "traefik",
	}

	testCases := []struct {
		desc         string
		tags         []string
		key          string
		defaultValue string
		expected     string
	}{
		{
			desc: "Should return tag value 42",
			tags: []string{
				"foo.bar=ramdom",
				"traefik.backend.weight=42",
			},
			key:          "backend.weight",
			defaultValue: "0",
			expected:     "42",
		},
		{
			desc: "Should return tag default value 0",
			tags: []string{
				"foo.bar=ramdom",
				"traefik.backend.wei=42",
			},
			key:          "backend.weight",
			defaultValue: "0",
			expected:     "0",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.getAttribute(test.key, test.tags, test.defaultValue)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetAttributeWithEmptyPrefix(t *testing.T) {
	provider := &CatalogProvider{
		Domain: "localhost",
		Prefix: "",
	}

	testCases := []struct {
		desc         string
		tags         []string
		key          string
		defaultValue string
		expected     string
	}{
		{
			desc: "Should return tag value 42",
			tags: []string{
				"foo.bar=ramdom",
				"backend.weight=42",
			},
			key:          "backend.weight",
			defaultValue: "0",
			expected:     "42",
		},
		{
			desc: "Should return default value 0",
			tags: []string{
				"foo.bar=ramdom",
				"backend.wei=42",
			},
			key:          "backend.weight",
			defaultValue: "0",
			expected:     "0",
		},
		{
			desc: "Should return for.bar key value random",
			tags: []string{
				"foo.bar=ramdom",
				"backend.wei=42",
			},
			key:          "foo.bar",
			defaultValue: "random",
			expected:     "ramdom",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.getAttribute(test.key, test.tags, test.defaultValue)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetBackendAddress(t *testing.T) {
	testCases := []struct {
		desc     string
		node     *api.ServiceEntry
		expected string
	}{
		{
			desc: "Should return the address of the service",
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
			desc: "Should return the address of the node",
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

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getBackendAddress(test.node)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetBackendName(t *testing.T) {
	testCases := []struct {
		desc     string
		node     *api.ServiceEntry
		expected string
	}{
		{
			desc: "Should create backend name without tags",
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
			desc: "Should create backend name with multiple tags",
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
			desc: "Should create backend name with one tag",
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

	for i, test := range testCases {
		test := test
		i := i
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getBackendName(test.node, i)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestBuildConfiguration(t *testing.T) {
	provider := &CatalogProvider{
		Domain:               "localhost",
		Prefix:               "traefik",
		ExposedByDefault:     false,
		FrontEndRule:         "Host:{{.ServiceName}}.{{.Domain}}",
		frontEndRuleTemplate: template.New("consul catalog frontend rule"),
	}

	testCases := []struct {
		desc              string
		nodes             []catalogUpdate
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc:              "Should build config of nothing",
			nodes:             []catalogUpdate{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			desc: "Should build config with no frontend and backend",
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
			desc: "Should build config who contains one frontend and one backend",
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
							"traefik.frontend.auth.basic=test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
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
					BasicAuth: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
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

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actualConfig := provider.buildConfiguration(test.nodes)
			assert.Equal(t, test.expectedBackends, actualConfig.Backends)
			assert.Equal(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestNodeSorter(t *testing.T) {
	testCases := []struct {
		desc     string
		nodes    []*api.ServiceEntry
		expected []*api.ServiceEntry
	}{
		{
			desc:     "Should sort nothing",
			nodes:    []*api.ServiceEntry{},
			expected: []*api.ServiceEntry{},
		},
		{
			desc: "Should sort by node address",
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
			desc: "Should sort by service name",
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
			desc: "Should sort by node address",
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

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			sort.Sort(nodeSorter(test.nodes))
			actual := test.nodes
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetChangedKeys(t *testing.T) {
	type Input struct {
		currState map[string]Service
		prevState map[string]Service
	}

	type Output struct {
		addedKeys   []string
		removedKeys []string
	}

	testCases := []struct {
		desc   string
		input  Input
		output Output
	}{
		{
			desc: "Should add 0 services and removed 0",
			input: Input{
				currState: map[string]Service{
					"foo-service":    {Name: "v1"},
					"bar-service":    {Name: "v1"},
					"baz-service":    {Name: "v1"},
					"qux-service":    {Name: "v1"},
					"quux-service":   {Name: "v1"},
					"quuz-service":   {Name: "v1"},
					"corge-service":  {Name: "v1"},
					"grault-service": {Name: "v1"},
					"garply-service": {Name: "v1"},
					"waldo-service":  {Name: "v1"},
					"fred-service":   {Name: "v1"},
					"plugh-service":  {Name: "v1"},
					"xyzzy-service":  {Name: "v1"},
					"thud-service":   {Name: "v1"},
				},
				prevState: map[string]Service{
					"foo-service":    {Name: "v1"},
					"bar-service":    {Name: "v1"},
					"baz-service":    {Name: "v1"},
					"qux-service":    {Name: "v1"},
					"quux-service":   {Name: "v1"},
					"quuz-service":   {Name: "v1"},
					"corge-service":  {Name: "v1"},
					"grault-service": {Name: "v1"},
					"garply-service": {Name: "v1"},
					"waldo-service":  {Name: "v1"},
					"fred-service":   {Name: "v1"},
					"plugh-service":  {Name: "v1"},
					"xyzzy-service":  {Name: "v1"},
					"thud-service":   {Name: "v1"},
				},
			},
			output: Output{
				addedKeys:   []string{},
				removedKeys: []string{},
			},
		},
		{
			desc: "Should add 3 services and removed 0",
			input: Input{
				currState: map[string]Service{
					"foo-service":    {Name: "v1"},
					"bar-service":    {Name: "v1"},
					"baz-service":    {Name: "v1"},
					"qux-service":    {Name: "v1"},
					"quux-service":   {Name: "v1"},
					"quuz-service":   {Name: "v1"},
					"corge-service":  {Name: "v1"},
					"grault-service": {Name: "v1"},
					"garply-service": {Name: "v1"},
					"waldo-service":  {Name: "v1"},
					"fred-service":   {Name: "v1"},
					"plugh-service":  {Name: "v1"},
					"xyzzy-service":  {Name: "v1"},
					"thud-service":   {Name: "v1"},
				},
				prevState: map[string]Service{
					"foo-service":    {Name: "v1"},
					"bar-service":    {Name: "v1"},
					"baz-service":    {Name: "v1"},
					"corge-service":  {Name: "v1"},
					"grault-service": {Name: "v1"},
					"garply-service": {Name: "v1"},
					"waldo-service":  {Name: "v1"},
					"fred-service":   {Name: "v1"},
					"plugh-service":  {Name: "v1"},
					"xyzzy-service":  {Name: "v1"},
					"thud-service":   {Name: "v1"},
				},
			},
			output: Output{
				addedKeys:   []string{"qux-service", "quux-service", "quuz-service"},
				removedKeys: []string{},
			},
		},
		{
			desc: "Should add 2 services and removed 2",
			input: Input{
				currState: map[string]Service{
					"foo-service":    {Name: "v1"},
					"qux-service":    {Name: "v1"},
					"quux-service":   {Name: "v1"},
					"quuz-service":   {Name: "v1"},
					"corge-service":  {Name: "v1"},
					"grault-service": {Name: "v1"},
					"garply-service": {Name: "v1"},
					"waldo-service":  {Name: "v1"},
					"fred-service":   {Name: "v1"},
					"plugh-service":  {Name: "v1"},
					"xyzzy-service":  {Name: "v1"},
					"thud-service":   {Name: "v1"},
				},
				prevState: map[string]Service{
					"foo-service":   {Name: "v1"},
					"bar-service":   {Name: "v1"},
					"baz-service":   {Name: "v1"},
					"qux-service":   {Name: "v1"},
					"quux-service":  {Name: "v1"},
					"quuz-service":  {Name: "v1"},
					"corge-service": {Name: "v1"},
					"waldo-service": {Name: "v1"},
					"fred-service":  {Name: "v1"},
					"plugh-service": {Name: "v1"},
					"xyzzy-service": {Name: "v1"},
					"thud-service":  {Name: "v1"},
				},
			},
			output: Output{
				addedKeys:   []string{"grault-service", "garply-service"},
				removedKeys: []string{"bar-service", "baz-service"},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			addedKeys, removedKeys := getChangedServiceKeys(test.input.currState, test.input.prevState)
			assert.Equal(t, fun.Set(test.output.addedKeys), fun.Set(addedKeys), "Added keys comparison results: got %q, want %q", addedKeys, test.output.addedKeys)
			assert.Equal(t, fun.Set(test.output.removedKeys), fun.Set(removedKeys), "Removed keys comparison results: got %q, want %q", removedKeys, test.output.removedKeys)
		})
	}
}

func TestFilterEnabled(t *testing.T) {
	testCases := []struct {
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

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			provider := &CatalogProvider{
				Domain:           "localhost",
				Prefix:           "traefik",
				ExposedByDefault: test.exposedByDefault,
			}
			actual := provider.nodeFilter("test", test.node)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetBasicAuth(t *testing.T) {
	testCases := []struct {
		desc     string
		tags     []string
		expected []string
	}{
		{
			desc:     "label missing",
			tags:     []string{},
			expected: []string{},
		},
		{
			desc: "label existing",
			tags: []string{
				"traefik.frontend.auth.basic=user:password",
			},
			expected: []string{"user:password"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			provider := &CatalogProvider{
				Prefix: "traefik",
			}
			actual := provider.getBasicAuth(test.tags)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestHasStickinessLabel(t *testing.T) {
	testCases := []struct {
		desc     string
		tags     []string
		expected bool
	}{
		{
			desc:     "label missing",
			tags:     []string{},
			expected: false,
		},
		{
			desc: "stickiness=true",
			tags: []string{
				label.TraefikBackendLoadBalancerStickiness + "=true",
			},
			expected: true,
		},
		{
			desc: "stickiness=false",
			tags: []string{
				label.TraefikBackendLoadBalancerStickiness + "=false",
			},
			expected: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := hasStickinessLabel(test.tags)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetChangedStringKeys(t *testing.T) {
	testCases := []struct {
		desc            string
		current         []string
		previous        []string
		expectedAdded   []string
		expectedRemoved []string
	}{
		{
			desc:            "1 element added, 0 removed",
			current:         []string{"chou"},
			previous:        []string{},
			expectedAdded:   []string{"chou"},
			expectedRemoved: []string{},
		}, {
			desc:            "0 element added, 0 removed",
			current:         []string{"chou"},
			previous:        []string{"chou"},
			expectedAdded:   []string{},
			expectedRemoved: []string{},
		},
		{
			desc:            "0 element added, 1 removed",
			current:         []string{},
			previous:        []string{"chou"},
			expectedAdded:   []string{},
			expectedRemoved: []string{"chou"},
		},
		{
			desc:            "1 element added, 1 removed",
			current:         []string{"carotte"},
			previous:        []string{"chou"},
			expectedAdded:   []string{"carotte"},
			expectedRemoved: []string{"chou"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actualAdded, actualRemoved := getChangedStringKeys(test.current, test.previous)
			assert.Equal(t, test.expectedAdded, actualAdded)
			assert.Equal(t, test.expectedRemoved, actualRemoved)
		})
	}
}

func TestHasNodeOrTagschanged(t *testing.T) {
	testCases := []struct {
		desc     string
		current  map[string]Service
		previous map[string]Service
		expected bool
	}{
		{
			desc: "Change detected due to change of nodes",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node2"},
					Tags:  []string{},
				},
			},
			expected: true,
		},
		{
			desc:    "No change missing current service",
			current: make(map[string]Service),
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			expected: false,
		},
		{
			desc: "No change on nodes",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			expected: false,
		},
		{
			desc: "No change on nodes and tags",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
				},
			},
			expected: false,
		},
		{
			desc: "Change detected con tags",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo"},
				},
			},
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := hasNodeOrTagsChanged(test.current, test.previous)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestHasChanged(t *testing.T) {
	testCases := []struct {
		desc     string
		current  map[string]Service
		previous map[string]Service
		expected bool
	}{
		{
			desc: "Change detected due to change new service",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			previous: make(map[string]Service),
			expected: true,
		},
		{
			desc:    "Change detected due to change service removed",
			current: make(map[string]Service),
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			expected: true,
		},
		{
			desc: "Change detected due to change of nodes",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node2"},
					Tags:  []string{},
				},
			},
			expected: true,
		},
		{
			desc: "No change on nodes",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{},
				},
			},
			expected: false,
		},
		{
			desc: "No change on nodes and tags",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
				},
			},
			expected: false,
		},
		{
			desc: "Change detected on tags",
			current: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo=bar"},
				},
			},
			previous: map[string]Service{
				"foo-service": {
					Name:  "foo",
					Nodes: []string{"node1"},
					Tags:  []string{"foo"},
				},
			},
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := hasChanged(test.current, test.previous)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetConstraintTags(t *testing.T) {
	provider := &CatalogProvider{
		Domain: "localhost",
		Prefix: "traefik",
	}

	testCases := []struct {
		desc     string
		tags     []string
		expected []string
	}{
		{
			desc: "nil tags",
		},
		{
			desc:     "invalid tag",
			tags:     []string{"tags=foobar"},
			expected: nil,
		},
		{
			desc:     "wrong tag",
			tags:     []string{"traefik_tags=foobar"},
			expected: nil,
		},
		{
			desc:     "empty value",
			tags:     []string{"traefik.tags="},
			expected: nil,
		},
		{
			desc:     "simple tag",
			tags:     []string{"traefik.tags=foobar "},
			expected: []string{"foobar"},
		},
		{
			desc:     "multiple values tag",
			tags:     []string{"traefik.tags=foobar, fiibir"},
			expected: []string{"foobar", "fiibir"},
		},
		{
			desc:     "multiple tags",
			tags:     []string{"traefik.tags=foobar", "traefik.tags=foobor"},
			expected: []string{"foobar", "foobor"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			constraints := provider.getConstraintTags(test.tags)
			assert.EqualValues(t, test.expected, constraints)
		})
	}
}
