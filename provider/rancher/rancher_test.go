package rancher

import (
	"strings"
	"testing"

	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderServiceFilter(t *testing.T) {
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
					types.LabelEnable: "true",
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
					types.LabelPort:   "80",
					types.LabelEnable: "false",
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
					types.LabelTags:   "cheese",
					types.LabelPort:   "80",
					types.LabelEnable: "true",
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
					types.LabelTags:   "not-cheesy",
					types.LabelPort:   "80",
					types.LabelEnable: "true",
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
					types.LabelTags:   "cheese",
					types.LabelPort:   "80",
					types.LabelEnable: "true",
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
					types.LabelTags:   "chose",
					types.LabelPort:   "80",
					types.LabelEnable: "true",
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
					types.LabelTags:   "cheeeeese",
					types.LabelPort:   "80",
					types.LabelEnable: "true",
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
					types.LabelFrontendRule: "Headers:User-Agent,bat/0.1.0",
				},
			},
			expected: "Headers-User-Agent-bat-0-1-0",
		},
		{
			desc: "with Host label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelFrontendRule: "Host:foo.bar",
				},
			},
			expected: "Host-foo-bar",
		},
		{
			desc: "with Path label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelFrontendRule: "Path:/test",
				},
			},
			expected: "Path-test",
		},
		{
			desc: "with PathPrefix label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelFrontendRule: "PathPrefix:/test2",
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
					types.LabelFrontendRule: "Host:foo.bar.com",
				},
			},
			expected: "Host:foo.bar.com",
		},
		{
			desc: "with Path label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelFrontendRule: "Path:/test",
				},
			},
			expected: "Path:/test",
		},
		{
			desc: "with PathPrefix label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelFrontendRule: "PathPrefix:/test2",
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

func TestProviderGetBackend(t *testing.T) {
	provider := &Provider{Domain: "rancher.localhost"}

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
					types.LabelBackend: "foobar",
				},
			},

			expected: "foobar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.getBackend(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetWeight(t *testing.T) {
	provider := &Provider{Domain: "rancher.localhost"}

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
			expected: "0",
		},
		{
			desc: "with label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelWeight: "5",
				},
			},
			expected: "5",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.getWeight(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetPort(t *testing.T) {
	provider := &Provider{Domain: "rancher.localhost"}

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
			expected: "",
		},
		{
			desc: "with label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelPort: "1337",
				},
			},
			expected: "1337",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.getPort(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetDomain(t *testing.T) {
	provider := &Provider{Domain: "rancher.localhost"}

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
			expected: "rancher.localhost",
		},
		{
			desc: "with label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelDomain: "foo.bar",
				},
			},
			expected: "foo.bar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.getDomain(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetProtocol(t *testing.T) {
	provider := &Provider{Domain: "rancher.localhost"}

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
			expected: "http",
		},
		{
			desc: "with label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelProtocol: "https",
				},
			},
			expected: "https",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.getProtocol(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetPassHostHeader(t *testing.T) {
	provider := &Provider{Domain: "rancher.localhost"}

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
			expected: "true",
		},
		{
			desc: "with label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelFrontendPassHostHeader: "false",
				},
			},
			expected: "false",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.getPassHostHeader(test.service)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestProviderGetLabel(t *testing.T) {
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
			expected: "label not found",
		},
		{
			desc: "with label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					"foo": "bar",
				},
			},
			expected: "",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run("", func(t *testing.T) {
			t.Parallel()

			label, err := getServiceLabel(test.service, "foo")

			if test.expected != "" {
				if err == nil || !strings.Contains(err.Error(), test.expected) {
					t.Fatalf("expected an error with %q, got %v", test.expected, err)
				}
			} else {
				assert.Equal(t, "bar", label)
			}
		})
	}
}

func TestProviderLoadRancherConfig(t *testing.T) {
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
						types.LabelPort:                       "80",
						types.LabelFrontendAuthBasic:          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						types.LabelFrontendRedirectEntryPoint: "https",
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

			actualConfig := provider.loadRancherConfig(test.services)

			require.NotNil(t, actualConfig)
			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestProviderHasStickinessLabel(t *testing.T) {
	provider := &Provider{Domain: "rancher.localhost"}

	testCases := []struct {
		desc     string
		service  rancherData
		expected bool
	}{
		{
			desc: "no labels",
			service: rancherData{
				Name: "test-service",
			},
			expected: false,
		},
		{
			desc: "stickiness=true",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelBackendLoadbalancerStickiness: "true",
				},
			},
			expected: true,
		},
		{
			desc: "stickiness=true",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelBackendLoadbalancerStickiness: "false",
				},
			},
			expected: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := provider.hasStickinessLabel(test.service)
			assert.Equal(t, actual, test.expected)
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
					types.LabelFrontendRedirectEntryPoint: "https",
				},
			},
			expected: true,
		},
		{
			desc: "with Redirect regex label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelFrontendRedirectRegex: `(.+)`,
				},
			},
			expected: false,
		},
		{
			desc: "with Redirect replacement label",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelFrontendRedirectReplacement: "$1",
				},
			},
			expected: false,
		},
		{
			desc: "with Redirect regex & replacement labels",
			service: rancherData{
				Name: "test-service",
				Labels: map[string]string{
					types.LabelFrontendRedirectRegex:       `(.+)`,
					types.LabelFrontendRedirectReplacement: "$1",
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
