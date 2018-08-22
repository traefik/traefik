package consulcatalog

import (
	"testing"
	"text/template"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func TestProviderBuildConfiguration(t *testing.T) {
	p := &Provider{
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
							label.TraefikBackendLoadBalancerMethod + "=drr",
							label.TraefikBackendCircuitBreakerExpression + "=NetworkErrorRatio() > 0.5",
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
									label.Prefix + "backend.weight=42", // Deprecated label
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
					LoadBalancer: &types.LoadBalancer{
						Method: "drr",
					},
					CircuitBreaker: &types.CircuitBreaker{
						Expression: "NetworkErrorRatio() > 0.5",
					},
					MaxConn: &types.MaxConn{
						Amount:        1000,
						ExtractorFunc: "client.ip",
					},
				},
			},
		},
		{
			desc: "Should build config containing one frontend, one IPv4 and one IPv6 backend",
			nodes: []catalogUpdate{
				{
					Service: &serviceUpdate{
						ServiceName: "test",
						Attributes: []string{
							"random.foo=bar",
							label.TraefikBackendLoadBalancerMethod + "=drr",
							label.TraefikBackendCircuitBreakerExpression + "=NetworkErrorRatio() > 0.5",
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
									label.Prefix + "backend.weight=42", // Deprecated label
									label.TraefikFrontendPassHostHeader + "=true",
									label.TraefikProtocol + "=https",
								},
							},
							Node: &api.Node{
								Node:    "localhost",
								Address: "127.0.0.1",
							},
						},
						{
							Service: &api.AgentService{
								Service: "test",
								Address: "::1",
								Port:    80,
								Tags: []string{
									"random.foo=bar",
									label.Prefix + "backend.weight=42", // Deprecated label
									label.TraefikFrontendPassHostHeader + "=true",
									label.TraefikProtocol + "=https",
								},
							},
							Node: &api.Node{
								Node:    "localhost",
								Address: "::1",
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
						"test-1-Gh4zrXo5flAAz1A8LAEHm1-TSnE": {
							URL:    "https://[::1]:80",
							Weight: 42,
						},
					},
					LoadBalancer: &types.LoadBalancer{
						Method: "drr",
					},
					CircuitBreaker: &types.CircuitBreaker{
						Expression: "NetworkErrorRatio() > 0.5",
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

			nodes := fakeLoadTraefikLabelsSlice(test.nodes, p.Prefix)

			actualConfig := p.buildConfigurationV2(nodes)
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

			p := &Provider{Prefix: test.prefix}

			actual := p.getPrefixedName(test.name)
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

			labels := tagsToNeutralLabels(test.service.Attributes, p.Prefix)
			test.service.TraefikLabels = labels

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

func TestGetServerName(t *testing.T) {
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

func fakeLoadTraefikLabelsSlice(nodes []catalogUpdate, prefix string) []catalogUpdate {
	var result []catalogUpdate

	for _, node := range nodes {
		labels := tagsToNeutralLabels(node.Service.Attributes, prefix)
		node.Service.TraefikLabels = labels
		result = append(result, node)
	}

	return result
}
