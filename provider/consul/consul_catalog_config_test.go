package consul

import (
	"testing"
	"text/template"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func TestCatalogProviderBuildConfiguration(t *testing.T) {
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
							"random.foo=bar",
							label.Prefix + "backend.loadbalancer=drr",
							label.TraefikBackendCircuitBreaker + "=NetworkErrorRatio() > 0.5",
							label.TraefikBackendMaxConnAmount + "=1000",
							label.TraefikBackendMaxConnExtractorFunc + "=client.ip",
							label.TraefikFrontendAuthBasic + "=test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						},
					},
					Nodes: []*api.ServiceEntry{
						{
							Service: &api.AgentService{
								Service: "test",
								Address: "127.0.0.1",
								Port:    80,
								Tags: []string{
									"random.foo=bar",
									label.Prefix + "backend.weight=42",
									label.TraefikFrontendPassHostHeader + "=true",
									label.TraefikProtocol + "=https",
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
					EntryPoints: []string{},
					BasicAuth:   []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"test-0-us4-27hAOu2ARV7nNrmv6GoKlcA": {
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
			assert.NotNil(t, actualConfig)
			assert.Equal(t, test.expectedBackends, actualConfig.Backends)
			assert.Equal(t, test.expectedFrontends, actualConfig.Frontends)
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

func TestCatalogProviderGetPrefixedName(t *testing.T) {
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

func TestCatalogProviderGetAttribute(t *testing.T) {
	testCases := []struct {
		desc         string
		tags         []string
		key          string
		defaultValue string
		prefix       string
		expected     string
	}{
		{
			desc:   "Should return tag value 42",
			prefix: "traefik",
			tags: []string{
				"foo.bar=ramdom",
				"traefik.backend.weight=42",
			},
			key:          "backend.weight",
			defaultValue: "0",
			expected:     "42",
		},
		{
			desc:   "Should return tag default value 0",
			prefix: "traefik",
			tags: []string{
				"foo.bar=ramdom",
				"traefik.backend.wei=42",
			},
			key:          "backend.weight",
			defaultValue: "0",
			expected:     "0",
		},
		{
			desc: "Should return tag value 42 when empty prefix",
			tags: []string{
				"foo.bar=ramdom",
				"backend.weight=42",
			},
			key:          "backend.weight",
			defaultValue: "0",
			expected:     "42",
		},
		{
			desc: "Should return default value 0 when empty prefix",
			tags: []string{
				"foo.bar=ramdom",
				"backend.wei=42",
			},
			key:          "backend.weight",
			defaultValue: "0",
			expected:     "0",
		},
		{
			desc: "Should return for.bar key value random when empty prefix",
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

			p := &CatalogProvider{
				Domain: "localhost",
				Prefix: test.prefix,
			}

			actual := p.getAttribute(test.key, test.tags, test.defaultValue)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestCatalogProviderGetIntAttribute(t *testing.T) {
	p := &CatalogProvider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc         string
		name         string
		tags         []string
		defaultValue int
		expected     int
	}{
		{
			desc:     "should return default value when empty name",
			name:     "",
			tags:     []string{"traefik.foo=10"},
			expected: 0,
		},
		{
			desc:     "should return default value when empty tags",
			name:     "traefik.foo",
			tags:     nil,
			expected: 0,
		},
		{
			desc:     "should return default value when value is not a int",
			name:     "foo",
			tags:     []string{"traefik.foo=bar"},
			expected: 0,
		},
		{
			desc:     "should return a value when tag exist",
			name:     "foo",
			tags:     []string{"traefik.foo=10"},
			expected: 10,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := p.getIntAttribute(test.name, test.tags, test.defaultValue)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCatalogProviderGetInt64Attribute(t *testing.T) {
	p := &CatalogProvider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc         string
		name         string
		tags         []string
		defaultValue int64
		expected     int64
	}{
		{
			desc:     "should return default value when empty name",
			name:     "",
			tags:     []string{"traefik.foo=10"},
			expected: 0,
		},
		{
			desc:     "should return default value when empty tags",
			name:     "traefik.foo",
			tags:     nil,
			expected: 0,
		},
		{
			desc:     "should return default value when value is not a int",
			name:     "foo",
			tags:     []string{"traefik.foo=bar"},
			expected: 0,
		},
		{
			desc:     "should return a value when tag exist",
			name:     "foo",
			tags:     []string{"traefik.foo=10"},
			expected: 10,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := p.getInt64Attribute(test.name, test.tags, test.defaultValue)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCatalogProviderGetBoolAttribute(t *testing.T) {
	p := &CatalogProvider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc         string
		name         string
		tags         []string
		defaultValue bool
		expected     bool
	}{
		{
			desc:     "should return default value when empty name",
			name:     "",
			tags:     []string{"traefik.foo=10"},
			expected: false,
		},
		{
			desc:     "should return default value when empty tags",
			name:     "traefik.foo",
			tags:     nil,
			expected: false,
		},
		{
			desc:     "should return default value when value is not a bool",
			name:     "foo",
			tags:     []string{"traefik.foo=bar"},
			expected: false,
		},
		{
			desc:     "should return a value when tag exist",
			name:     "foo",
			tags:     []string{"traefik.foo=true"},
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := p.getBoolAttribute(test.name, test.tags, test.defaultValue)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCatalogProviderGetSliceAttribute(t *testing.T) {
	p := &CatalogProvider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc     string
		name     string
		tags     []string
		expected []string
	}{
		{
			desc:     "should return nil when empty name",
			name:     "",
			tags:     []string{"traefik.foo=bar,bor,bir"},
			expected: nil,
		},
		{
			desc:     "should return nil when empty tags",
			name:     "foo",
			tags:     nil,
			expected: nil,
		},
		{
			desc:     "should return nil when tag doesn't have value",
			name:     "",
			tags:     []string{"traefik.foo="},
			expected: nil,
		},
		{
			desc:     "should return a slice when tag contains comma separated values",
			name:     "foo",
			tags:     []string{"traefik.foo=bar,bor,bir"},
			expected: []string{"bar", "bor", "bir"},
		},
		{
			desc:     "should return a slice when tag contains one value",
			name:     "foo",
			tags:     []string{"traefik.foo=bar"},
			expected: []string{"bar"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := p.getSliceAttribute(test.name, test.tags)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCatalogProviderGetFrontendRule(t *testing.T) {
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

			p := &CatalogProvider{
				Domain:               "localhost",
				Prefix:               "traefik",
				FrontEndRule:         "Host:{{.ServiceName}}.{{.Domain}}",
				frontEndRuleTemplate: template.New("consul catalog frontend rule"),
			}
			p.setupFrontEndRuleTemplate()

			actual := p.getFrontendRule(test.service)
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

func TestCatalogProviderGetServerName(t *testing.T) {
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
			expected: "api-0-eUSiqD6uNvvh6zxsY-OeRi8ZbaE",
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
			expected: "api-1-eJ8MR2JxjXyZgs1bhurVa0-9OI8",
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
			expected: "api-2-lMCDCsG7sh0SCXOHo4oBOQB-9D4",
		},
	}

	for i, test := range testCases {
		test := test
		i := i
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getServerName(test.node, i)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestHasStickinessLabel(t *testing.T) {
	p := &CatalogProvider{
		Prefix: "traefik",
	}

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

			actual := p.hasStickinessLabel(test.tags)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestCatalogProviderGetCircuitBreaker(t *testing.T) {
	p := &CatalogProvider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc     string
		tags     []string
		expected *types.CircuitBreaker
	}{
		{
			desc:     "should return nil when no tags",
			tags:     []string{},
			expected: nil,
		},
		{
			desc: "should return a struct when has tag",
			tags: []string{label.Prefix + label.SuffixBackendCircuitBreaker + "=foo"},
			expected: &types.CircuitBreaker{
				Expression: "foo",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := p.getCircuitBreaker(test.tags)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCatalogProviderGetLoadBalancer(t *testing.T) {
	p := &CatalogProvider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc     string
		tags     []string
		expected *types.LoadBalancer
	}{
		{
			desc: "should return a default struct when no tags",
			tags: []string{},
			expected: &types.LoadBalancer{
				Method: "wrr",
			},
		},
		{
			desc: "should return a struct when has Method tag",
			tags: []string{label.Prefix + "backend.loadbalancer" + "=drr"},
			expected: &types.LoadBalancer{
				Method: "drr",
			},
		},
		{
			desc: "should return a struct when has Sticky tag",
			tags: []string{
				label.Prefix + label.SuffixBackendLoadBalancerSticky + "=true",
			},
			expected: &types.LoadBalancer{
				Method: "wrr",
				Sticky: true,
			},
		},
		{
			desc: "should skip Sticky when Sticky tag has invalid value",
			tags: []string{
				label.Prefix + label.SuffixBackendLoadBalancerSticky + "=goo",
			},
			expected: &types.LoadBalancer{
				Method: "wrr",
			},
		},
		{
			desc: "should return a struct when has Stickiness tag",
			tags: []string{
				label.Prefix + label.SuffixBackendLoadBalancerStickiness + "=true",
			},
			expected: &types.LoadBalancer{
				Method:     "wrr",
				Stickiness: &types.Stickiness{},
			},
		},
		{
			desc: "should skip Stickiness when Stickiness tag has invalid value",
			tags: []string{
				label.Prefix + label.SuffixBackendLoadBalancerStickiness + "=goo",
			},
			expected: &types.LoadBalancer{
				Method: "wrr",
			},
		},
		{
			desc: "should return a struct when has Stickiness tag",
			tags: []string{
				label.Prefix + label.SuffixBackendLoadBalancerStickiness + "=true",
				label.Prefix + label.SuffixBackendLoadBalancerStickinessCookieName + "=bar",
			},
			expected: &types.LoadBalancer{
				Method: "wrr",
				Stickiness: &types.Stickiness{
					CookieName: "bar",
				},
			},
		},
		{
			desc: "should skip Stickiness when Stickiness tag has false as value",
			tags: []string{
				label.Prefix + label.SuffixBackendLoadBalancerStickiness + "=false",
				label.Prefix + label.SuffixBackendLoadBalancerStickinessCookieName + "=bar",
			},
			expected: &types.LoadBalancer{
				Method: "wrr",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := p.getLoadBalancer(test.tags)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCatalogProviderGetMaxConn(t *testing.T) {
	p := &CatalogProvider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc     string
		tags     []string
		expected *types.MaxConn
	}{
		{
			desc:     "should return nil when no tags",
			tags:     []string{},
			expected: nil,
		},
		{
			desc: "should return a struct when Amount & ExtractorFunc tags",
			tags: []string{
				label.Prefix + label.SuffixBackendMaxConnAmount + "=10",
				label.Prefix + label.SuffixBackendMaxConnExtractorFunc + "=bar",
			},
			expected: &types.MaxConn{
				ExtractorFunc: "bar",
				Amount:        10,
			},
		},
		{
			desc: "should return nil when Amount tags is missing",
			tags: []string{
				label.Prefix + label.SuffixBackendMaxConnExtractorFunc + "=bar",
			},
			expected: nil,
		},
		{
			desc: "should return nil when ExtractorFunc tags is empty",
			tags: []string{
				label.Prefix + label.SuffixBackendMaxConnAmount + "=10",
				label.Prefix + label.SuffixBackendMaxConnExtractorFunc + "=",
			},
			expected: nil,
		},
		{
			desc: "should return a struct when ExtractorFunc tags is missing",
			tags: []string{
				label.Prefix + label.SuffixBackendMaxConnAmount + "=10",
			},
			expected: &types.MaxConn{
				ExtractorFunc: label.DefaultBackendMaxconnExtractorFunc,
				Amount:        10,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := p.getMaxConn(test.tags)

			assert.Equal(t, test.expected, result)
		})
	}
}
