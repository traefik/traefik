package consulcatalog

import (
	"testing"
	"text/template"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func TestProviderBuildConfigurationV1(t *testing.T) {
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
							label.TraefikBackendLoadBalancer + "=drr",
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
					BasicAuth: []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
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

			actualConfig := p.buildConfigurationV1(test.nodes)
			assert.NotNil(t, actualConfig)
			assert.Equal(t, test.expectedBackends, actualConfig.Backends)
			assert.Equal(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestProviderGetIntAttributeV1(t *testing.T) {
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

func TestProviderGetInt64AttributeV1(t *testing.T) {
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

func TestProviderGetBoolAttributeV1(t *testing.T) {
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

func TestProviderGetSliceAttributeV1(t *testing.T) {
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

func TestProviderGetFrontendRuleV1(t *testing.T) {
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

			actual := p.getFrontendRuleV1(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestHasStickinessLabelV1(t *testing.T) {
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

			actual := p.hasStickinessLabelV1(test.tags)
			assert.Equal(t, test.expected, actual)
		})
	}
}
