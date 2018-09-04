package rancher

import (
	"testing"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderBuildConfigurationV1(t *testing.T) {
	provider := &Provider{
		Domain:           "rancher.localhost",
		ExposedByDefault: true,
	}

	testCases := []struct {
		desc              string
		services          []rancherData
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc:              "without services",
			services:          []rancherData{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			desc: "when all labels are set",
			services: []rancherData{
				{
					Labels: map[string]string{
						label.TraefikPort:     "666",
						label.TraefikProtocol: "https",
						label.TraefikWeight:   "12",

						label.TraefikBackend: "foobar",

						label.TraefikBackendCircuitBreakerExpression:         "NetworkErrorRatio() > 0.5",
						label.TraefikBackendLoadBalancerMethod:               "drr",
						label.TraefikBackendLoadBalancerSticky:               "true",
						label.TraefikBackendLoadBalancerStickiness:           "true",
						label.TraefikBackendLoadBalancerStickinessCookieName: "chocolate",
						label.TraefikBackendMaxConnAmount:                    "666",
						label.TraefikBackendMaxConnExtractorFunc:             "client.ip",

						label.TraefikFrontendAuthBasic:           "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendEntryPoints:         "http,https",
						label.TraefikFrontendPassHostHeader:      "true",
						label.TraefikFrontendPriority:            "666",
						label.TraefikFrontendRedirectEntryPoint:  "https",
						label.TraefikFrontendRedirectRegex:       "nope",
						label.TraefikFrontendRedirectReplacement: "nope",
						label.TraefikFrontendRule:                "Host:traefik.io",
					},
					Health:     "healthy",
					Containers: []string{"10.0.0.1", "10.0.0.2"},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-traefik-io": {
					EntryPoints: []string{
						"http",
						"https",
					},
					Backend: "backend-foobar",
					Routes: map[string]types.Route{
						"route-frontend-Host-traefik-io": {
							Rule: "Host:traefik.io",
						},
					},
					PassHostHeader: true,
					Priority:       666,
					BasicAuth: []string{
						"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
						"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					},
					Redirect: &types.Redirect{
						EntryPoint:  "https",
						Regex:       "nope",
						Replacement: "nope",
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"server-0": {
							URL:    "https://10.0.0.1:666",
							Weight: 12,
						},
						"server-1": {
							URL:    "https://10.0.0.2:666",
							Weight: 12,
						},
					},
					CircuitBreaker: &types.CircuitBreaker{
						Expression: "NetworkErrorRatio() > 0.5",
					},
					LoadBalancer: &types.LoadBalancer{
						Method: "drr",
						Sticky: true,
						Stickiness: &types.Stickiness{
							CookieName: "chocolate",
						},
					},
					MaxConn: &types.MaxConn{
						Amount:        666,
						ExtractorFunc: "client.ip",
					},
				},
			},
		},
		{
			desc: "with services",
			services: []rancherData{
				{
					Name: "test/service",
					Labels: map[string]string{
						label.TraefikPort:                       "80",
						label.TraefikFrontendAuthBasic:          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendRedirectEntryPoint: "https",
					},
					Health:     "healthy",
					Containers: []string{"127.0.0.1"},
				},
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-service-rancher-localhost": {
					Backend:        "backend-test-service",
					PassHostHeader: true,
					EntryPoints:    []string{},
					BasicAuth:      []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
					Priority:       0,
					Redirect: &types.Redirect{
						EntryPoint: "https",
					},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-service-rancher-localhost": {
							Rule: "Host:test.service.rancher.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test-service": {
					Servers: map[string]types.Server{
						"server-0": {
							URL:    "http://127.0.0.1:80",
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actualConfig := provider.buildConfigurationV1(test.services)
			require.NotNil(t, actualConfig)

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestProviderServiceFilterV1(t *testing.T) {
	provider := &Provider{
		Domain:                    "rancher.localhost",
		EnableServiceHealthFilter: true,
	}

	constraint, _ := types.NewConstraint("tag==ch*se")
	provider.Constraints = types.Constraints{constraint}

	testCases := []struct {
		desc     string
		service  rancherData
		expected bool
	}{
		{
			desc: "missing Port labels, don't respect constraint",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: false,
		},
		{
			desc: "don't respect constraint",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikPort:   "80",
					label.TraefikEnable: "false",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: false,
		},
		{
			desc: "unhealthy",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "cheese",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "unhealthy",
				State:  "active",
			},
			expected: false,
		},
		{
			desc: "inactive",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "not-cheesy",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "inactive",
			},
			expected: false,
		},
		{
			desc: "healthy & active, tag: cheese",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "cheese",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: true,
		},
		{
			desc: "healthy & active, tag: chose",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "chose",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "active",
			},
			expected: true,
		},
		{
			desc: "healthy & upgraded",
			service: rancherData{
				Labels: map[string]string{
					label.TraefikTags:   "cheeeeese",
					label.TraefikPort:   "80",
					label.TraefikEnable: "true",
				},
				Health: "healthy",
				State:  "upgraded",
			},
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.serviceFilterV1(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestContainerFilterV1(t *testing.T) {
	testCases := []struct {
		name        string
		healthState string
		state       string
		expected    bool
	}{
		{
			healthState: "unhealthy",
			state:       "running",
			expected:    false,
		},
		{
			healthState: "healthy",
			state:       "stopped",
			expected:    false,
		},
		{
			state:    "stopped",
			expected: false,
		},
		{
			healthState: "healthy",
			state:       "running",
			expected:    true,
		},
		{
			healthState: "updating-healthy",
			state:       "updating-running",
			expected:    true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.healthState+" "+test.state, func(t *testing.T) {
			t.Parallel()

			actual := containerFilter(test.name, test.healthState, test.state)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetFrontendNameV1(t *testing.T) {
	provider := &Provider{Domain: "rancher.localhost"}

	testCases := []struct {
		desc     string
		service  rancherData
		expected string
	}{
		{
			desc: "default",
			service: rancherData{
				Name: "foo",
			},
			expected: "Host-foo-rancher-localhost",
		},
		{
			desc: "with Headers label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Headers:User-Agent,bat/0.1.0",
				},
			},
			expected: "Headers-User-Agent-bat-0-1-0",
		},
		{
			desc: "with Host label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Host:foo.bar",
				},
			},
			expected: "Host-foo-bar",
		},
		{
			desc: "with Path label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Path:/test",
				},
			},
			expected: "Path-test",
		},
		{
			desc: "with PathPrefix label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "PathPrefix:/test2",
				},
			},
			expected: "PathPrefix-test2",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.getFrontendNameV1(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetFrontendRuleV1(t *testing.T) {
	provider := &Provider{Domain: "rancher.localhost"}

	testCases := []struct {
		desc     string
		service  rancherData
		expected string
	}{
		{
			desc: "host",
			service: rancherData{
				Name: "foo",
			},
			expected: "Host:foo.rancher.localhost",
		},
		{
			desc: "host with /",
			service: rancherData{
				Name: "foo/bar",
			},
			expected: "Host:foo.bar.rancher.localhost",
		},
		{
			desc: "with Host label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Host:foo.bar.com",
				},
			},
			expected: "Host:foo.bar.com",
		},
		{
			desc: "with Path label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "Path:/test",
				},
			},
			expected: "Path:/test",
		},
		{
			desc: "with PathPrefix label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRule: "PathPrefix:/test2",
				},
			},
			expected: "PathPrefix:/test2",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.getFrontendRule(test.service.Name, test.service.Labels)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetBackendNameV1(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected string
	}{
		{
			desc: "without label",
			service: rancherData{
				Name: "test-service",
			},
			expected: "test-service",
		},
		{
			desc: "with label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikBackend: "foobar",
				},
			},

			expected: "foobar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getBackendNameV1(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}
