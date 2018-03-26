package consulcatalog

import (
	"testing"
	"text/template"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func TestProviderBuildConfiguration(t *testing.T) {
	provider := &Provider{
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

func TestProviderGetPrefixedName(t *testing.T) {
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

			pro := &Provider{Prefix: test.prefix}

			actual := pro.getPrefixedName(test.name)
			assert.Equal(t, test.expected, actual)
		})
	}

}

func TestProviderGetAttribute(t *testing.T) {
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

			p := &Provider{
				Domain: "localhost",
				Prefix: test.prefix,
			}

			actual := p.getAttribute(test.key, test.tags, test.defaultValue)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetIntAttribute(t *testing.T) {
	p := &Provider{
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
			desc:         "should return default value when empty name",
			name:         "",
			tags:         []string{"traefik.foo=10"},
			defaultValue: 666,
			expected:     666,
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

func TestProviderGetInt64Attribute(t *testing.T) {
	p := &Provider{
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
			desc:         "should return default value when empty name",
			name:         "",
			tags:         []string{"traefik.foo=10"},
			defaultValue: 666,
			expected:     666,
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

func TestProviderGetBoolAttribute(t *testing.T) {
	p := &Provider{
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
			desc:         "should return default value when empty name",
			name:         "",
			tags:         []string{"traefik.foo=true"},
			defaultValue: true,
			expected:     true,
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

func TestProviderGetSliceAttribute(t *testing.T) {
	p := &Provider{
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

func TestProviderGetFrontendRule(t *testing.T) {
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

			p := &Provider{
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

func TestProviderGetServerName(t *testing.T) {
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
	p := &Provider{
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

func TestProviderGetCircuitBreaker(t *testing.T) {
	p := &Provider{
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

func TestProviderGetLoadBalancer(t *testing.T) {
	p := &Provider{
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

func TestProviderGetMaxConn(t *testing.T) {
	p := &Provider{
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

func TestProviderGetHealthCheck(t *testing.T) {
	p := &Provider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc     string
		tags     []string
		expected *types.HealthCheck
	}{
		{
			desc:     "should return nil when no tags",
			tags:     []string{},
			expected: nil,
		},
		{
			desc: "should return nil when Path tag is missing",
			tags: []string{
				label.TraefikBackendHealthCheckPort + "=80",
				label.TraefikBackendHealthCheckInterval + "=7",
			},
			expected: nil,
		},
		{
			desc: "should return a struct when has tags",
			tags: []string{
				label.TraefikBackendHealthCheckPath + "=/health",
				label.TraefikBackendHealthCheckPort + "=80",
				label.TraefikBackendHealthCheckInterval + "=7",
			},
			expected: &types.HealthCheck{
				Path:     "/health",
				Port:     80,
				Interval: "7",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := p.getHealthCheck(test.tags)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderGetBuffering(t *testing.T) {
	p := &Provider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc     string
		tags     []string
		expected *types.Buffering
	}{
		{
			desc:     "should return nil when no tags",
			tags:     []string{},
			expected: nil,
		},
		{
			desc: "should return a struct when has proper tags",
			tags: []string{
				label.TraefikBackendBufferingMaxResponseBodyBytes + "=10485760",
				label.TraefikBackendBufferingMemResponseBodyBytes + "=2097152",
				label.TraefikBackendBufferingMaxRequestBodyBytes + "=10485760",
				label.TraefikBackendBufferingMemRequestBodyBytes + "=2097152",
				label.TraefikBackendBufferingRetryExpression + "=IsNetworkError() && Attempts() <= 2",
			},
			expected: &types.Buffering{
				MaxResponseBodyBytes: 10485760,
				MemResponseBodyBytes: 2097152,
				MaxRequestBodyBytes:  10485760,
				MemRequestBodyBytes:  2097152,
				RetryExpression:      "IsNetworkError() && Attempts() <= 2",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := p.getBuffering(test.tags)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderWhiteList(t *testing.T) {
	p := &Provider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc     string
		tags     []string
		expected *types.WhiteList
	}{
		{
			desc:     "should return nil when no white list labels",
			expected: nil,
		},
		{
			desc: "should return a struct when only range",
			tags: []string{
				label.TraefikFrontendWhiteListSourceRange + "=10.10.10.10",
			},
			expected: &types.WhiteList{
				SourceRange: []string{
					"10.10.10.10",
				},
				UseXForwardedFor: false,
			},
		},
		{
			desc: "should return a struct when range and UseXForwardedFor",
			tags: []string{
				label.TraefikFrontendWhiteListSourceRange + "=10.10.10.10",
				label.TraefikFrontendWhiteListUseXForwardedFor + "=true",
			},
			expected: &types.WhiteList{
				SourceRange: []string{
					"10.10.10.10",
				},
				UseXForwardedFor: true,
			},
		},
		{
			desc: "should return nil when only UseXForwardedFor",
			tags: []string{
				label.TraefikFrontendWhiteListUseXForwardedFor + "=true",
			},
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := p.getWhiteList(test.tags)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetRedirect(t *testing.T) {
	p := &Provider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc     string
		tags     []string
		expected *types.Redirect
	}{
		{
			desc:     "should return nil when no tags",
			tags:     []string{},
			expected: nil,
		},
		{
			desc: "should use only entry point tag when mix regex redirect and entry point redirect",
			tags: []string{
				label.TraefikFrontendRedirectEntryPoint + "=https",
				label.TraefikFrontendRedirectRegex + "=(.*)",
				label.TraefikFrontendRedirectReplacement + "=$1",
			},
			expected: &types.Redirect{
				EntryPoint: "https",
			},
		},
		{
			desc: "should return a struct when entry point redirect tag",
			tags: []string{
				label.TraefikFrontendRedirectEntryPoint + "=https",
			},
			expected: &types.Redirect{
				EntryPoint: "https",
			},
		},
		{
			desc: "should return a struct when entry point redirect tags (permanent)",
			tags: []string{
				label.TraefikFrontendRedirectEntryPoint + "=https",
				label.TraefikFrontendRedirectPermanent + "=true",
			},
			expected: &types.Redirect{
				EntryPoint: "https",
				Permanent:  true,
			},
		},
		{
			desc: "should return a struct when regex redirect tags",
			tags: []string{
				label.TraefikFrontendRedirectRegex + "=(.*)",
				label.TraefikFrontendRedirectReplacement + "=$1",
			},
			expected: &types.Redirect{
				Regex:       "(.*)",
				Replacement: "$1",
			},
		},
		{
			desc: "should return a struct when regex redirect tags (permanent)",
			tags: []string{
				label.TraefikFrontendRedirectRegex + "=(.*)",
				label.TraefikFrontendRedirectReplacement + "=$1",
				label.TraefikFrontendRedirectPermanent + "=true",
			},
			expected: &types.Redirect{
				Regex:       "(.*)",
				Replacement: "$1",
				Permanent:   true,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := p.getRedirect(test.tags)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderGetErrorPages(t *testing.T) {
	p := &Provider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc     string
		tags     []string
		expected map[string]*types.ErrorPage
	}{
		{
			desc:     "should return nil when no tags",
			tags:     []string{},
			expected: nil,
		},
		{
			desc: "should return a map when tags are present",
			tags: []string{
				label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus + "=404",
				label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageBackend + "=foo_backend",
				label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageQuery + "=foo_query",
				label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageStatus + "=500,600",
				label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageBackend + "=bar_backend",
				label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageQuery + "=bar_query",
			},
			expected: map[string]*types.ErrorPage{
				"foo": {
					Status:  []string{"404"},
					Query:   "foo_query",
					Backend: "foo_backend",
				},
				"bar": {
					Status:  []string{"500", "600"},
					Query:   "bar_query",
					Backend: "bar_backend",
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := p.getErrorPages(test.tags)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderGetRateLimit(t *testing.T) {
	p := &Provider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc     string
		tags     []string
		expected *types.RateLimit
	}{
		{
			desc:     "should return nil when no tags",
			tags:     []string{},
			expected: nil,
		},
		{
			desc: "should return a map when tags are present",
			tags: []string{
				label.TraefikFrontendRateLimitExtractorFunc + "=client.ip",
				label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod + "=6",
				label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage + "=12",
				label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst + "=18",
				label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod + "=3",
				label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage + "=6",
				label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst + "=9",
			},
			expected: &types.RateLimit{
				ExtractorFunc: "client.ip",
				RateSet: map[string]*types.Rate{
					"foo": {
						Period:  flaeg.Duration(6 * time.Second),
						Average: 12,
						Burst:   18,
					},
					"bar": {
						Period:  flaeg.Duration(3 * time.Second),
						Average: 6,
						Burst:   9,
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := p.getRateLimit(test.tags)

			assert.Equal(t, test.expected, result)
		})
	}
}

func TestProviderGetHeaders(t *testing.T) {
	p := &Provider{
		Prefix: "traefik",
	}

	testCases := []struct {
		desc     string
		tags     []string
		expected *types.Headers
	}{
		{
			desc:     "should return nil when no tags",
			tags:     []string{},
			expected: nil,
		},
		{
			desc: "should return a struct when has tags",
			tags: []string{
				label.TraefikFrontendRequestHeaders + "=Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
				label.TraefikFrontendResponseHeaders + "=Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
				label.TraefikFrontendSSLProxyHeaders + "=Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
				label.TraefikFrontendAllowedHosts + "=foo,bar,bor",
				label.TraefikFrontendHostsProxyHeaders + "=foo,bar,bor",
				label.TraefikFrontendSSLHost + "=foo",
				label.TraefikFrontendCustomFrameOptionsValue + "=foo",
				label.TraefikFrontendContentSecurityPolicy + "=foo",
				label.TraefikFrontendPublicKey + "=foo",
				label.TraefikFrontendReferrerPolicy + "=foo",
				label.TraefikFrontendCustomBrowserXSSValue + "=foo",
				label.TraefikFrontendSTSSeconds + "=666",
				label.TraefikFrontendSSLRedirect + "=true",
				label.TraefikFrontendSSLTemporaryRedirect + "=true",
				label.TraefikFrontendSTSIncludeSubdomains + "=true",
				label.TraefikFrontendSTSPreload + "=true",
				label.TraefikFrontendForceSTSHeader + "=true",
				label.TraefikFrontendFrameDeny + "=true",
				label.TraefikFrontendContentTypeNosniff + "=true",
				label.TraefikFrontendBrowserXSSFilter + "=true",
				label.TraefikFrontendIsDevelopment + "=true",
			},
			expected: &types.Headers{
				CustomRequestHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
				},
				CustomResponseHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
				},
				SSLProxyHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
				},
				AllowedHosts:            []string{"foo", "bar", "bor"},
				HostsProxyHeaders:       []string{"foo", "bar", "bor"},
				SSLHost:                 "foo",
				CustomFrameOptionsValue: "foo",
				ContentSecurityPolicy:   "foo",
				PublicKey:               "foo",
				ReferrerPolicy:          "foo",
				CustomBrowserXSSValue:   "foo",
				STSSeconds:              666,
				SSLRedirect:             true,
				SSLTemporaryRedirect:    true,
				STSIncludeSubdomains:    true,
				STSPreload:              true,
				ForceSTSHeader:          true,
				FrameDeny:               true,
				ContentTypeNosniff:      true,
				BrowserXSSFilter:        true,
				IsDevelopment:           true,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := p.getHeaders(test.tags)

			assert.Equal(t, test.expected, result)
		})
	}
}
