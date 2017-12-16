package rancher

import (
	"testing"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderBuildConfiguration(t *testing.T) {
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
							Weight: 0,
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

			actualConfig := provider.buildConfiguration(test.services)
			require.NotNil(t, actualConfig)

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestProviderServiceFilter(t *testing.T) {
	provider := &Provider{
		Domain: "rancher.localhost",
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

			actual := provider.serviceFilter(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestContainerFilter(t *testing.T) {
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

func TestProviderGetFrontendName(t *testing.T) {
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

			actual := provider.getFrontendName(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetFrontendRule(t *testing.T) {
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

			actual := provider.getFrontendRule(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetBackend(t *testing.T) {
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

			actual := getBackend(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestHasRedirect(t *testing.T) {
	testCases := []struct {
		desc     string
		service  rancherData
		expected bool
	}{
		{
			desc: "without redirect labels",
			service: rancherData{
				Name: "test-service",
			},
			expected: false,
		},
		{
			desc: "with Redirect EntryPoint label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRedirectEntryPoint: "https",
				},
			},
			expected: true,
		},
		{
			desc: "with Redirect regex label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRedirectRegex: `(.+)`,
				},
			},
			expected: false,
		},
		{
			desc: "with Redirect replacement label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRedirectReplacement: "$1",
				},
			},
			expected: false,
		},
		{
			desc: "with Redirect regex & replacement labels",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					label.TraefikFrontendRedirectRegex:       `(.+)`,
					label.TraefikFrontendRedirectReplacement: "$1",
				},
			},
			expected: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := hasRedirect(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}
