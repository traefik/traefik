package docker

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	docker "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwarmGetFrontendName(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(serviceName("foo")),
			expected: "Host-foo-docker-localhost-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Headers:User-Agent,bat/0.1.0",
			})),
			expected: "Headers-User-Agent-bat-0-1-0-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Host:foo.bar",
			})),
			expected: "Host-foo-bar-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Path:/test",
			})),
			expected: "Path-test-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(
				serviceName("test"),
				serviceLabels(map[string]string{
					label.TraefikFrontendRule: "PathPrefix:/test2",
				}),
			),
			expected: "PathPrefix-test2-0",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dData := parseService(test.service, test.networks)
			provider := &Provider{
				Domain:    "docker.localhost",
				SwarmMode: true,
			}
			actual := provider.getFrontendName(dData, 0)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestSwarmGetFrontendRule(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(serviceName("foo")),
			expected: "Host:foo.docker.localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service:  swarmService(serviceName("bar")),
			expected: "Host:bar.docker.localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Host:foo.bar",
			})),
			expected: "Host:foo.bar",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Path:/test",
			})),
			expected: "Path:/test",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dData := parseService(test.service, test.networks)
			provider := &Provider{
				Domain:    "docker.localhost",
				SwarmMode: true,
			}
			actual := provider.getFrontendRule(dData)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestSwarmGetBackend(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(serviceName("foo")),
			expected: "foo",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service:  swarmService(serviceName("bar")),
			expected: "bar",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikBackend: "foobar",
			})),
			expected: "foobar",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dData := parseService(test.service, test.networks)
			actual := getBackend(dData)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestSwarmGetIPAddress(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(withEndpointSpec(modeDNSSR)),
			expected: "",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(
				withEndpointSpec(modeVIP),
				withEndpoint(virtualIP("1", "10.11.12.13/24")),
			),
			expected: "10.11.12.13",
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			service: swarmService(
				serviceLabels(map[string]string{
					labelDockerNetwork: "barnet",
				}),
				withEndpointSpec(modeVIP),
				withEndpoint(
					virtualIP("1", "10.11.12.13/24"),
					virtualIP("2", "10.11.12.99/24"),
				),
			),
			expected: "10.11.12.99",
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foonet",
				},
				"2": {
					Name: "barnet",
				},
			},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dData := parseService(test.service, test.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getIPAddress(dData)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestSwarmGetPort(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarmService(
				serviceLabels(map[string]string{
					label.TraefikPort: "8080",
				}),
				withEndpointSpec(modeDNSSR),
			),
			expected: "8080",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dData := parseService(test.service, test.networks)
			actual := getPort(dData)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestSwarmTraefikFilter(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected bool
		networks map[string]*docker.NetworkResource
		provider *Provider
	}{
		{
			service:  swarmService(),
			expected: false,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikEnable: "false",
				label.TraefikPort:   "80",
			})),
			expected: false,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Host:foo.bar",
				label.TraefikPort:         "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikPort: "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikEnable: "true",
				label.TraefikPort:   "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikEnable: "anything",
				label.TraefikPort:   "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikFrontendRule: "Host:foo.bar",
				label.TraefikPort:         "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikPort: "80",
			})),
			expected: false,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: false,
			},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikEnable: "true",
				label.TraefikPort:   "80",
			})),
			expected: true,
			networks: map[string]*docker.NetworkResource{},
			provider: &Provider{
				SwarmMode:        true,
				Domain:           "test",
				ExposedByDefault: false,
			},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dData := parseService(test.service, test.networks)
			actual := test.provider.containerFilter(dData)
			if actual != test.expected {
				t.Errorf("expected %v for %+v, got %+v", test.expected, test, actual)
			}
		})
	}
}

func TestSwarmLoadDockerConfig(t *testing.T) {
	testCases := []struct {
		services          []swarm.Service
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
		networks          map[string]*docker.NetworkResource
	}{
		{
			services:          []swarm.Service{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
			networks:          map[string]*docker.NetworkResource{},
		},
		{
			services: []swarm.Service{
				swarmService(
					serviceName("test"),
					serviceLabels(map[string]string{
						label.TraefikPort: "80",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-docker-localhost-0": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					BasicAuth:      []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-docker-localhost-0": {
							Rule: "Host:test.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test": {
					Servers: map[string]types.Server{
						"server-test": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
					LoadBalancer:   nil,
				},
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			services: []swarm.Service{
				swarmService(
					serviceName("test1"),
					serviceLabels(map[string]string{
						label.TraefikPort:                       "80",
						label.TraefikBackend:                    "foobar",
						label.TraefikFrontendEntryPoints:        "http,https",
						label.TraefikFrontendAuthBasic:          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendRedirectEntryPoint: "https",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
				swarmService(
					serviceName("test2"),
					serviceLabels(map[string]string{
						label.TraefikPort:    "80",
						label.TraefikBackend: "foobar",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test1-docker-localhost-0": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					BasicAuth:      []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
					Redirect: &types.Redirect{
						EntryPoint: "https",
					},
					Routes: map[string]types.Route{
						"route-frontend-Host-test1-docker-localhost-0": {
							Rule: "Host:test1.docker.localhost",
						},
					},
				},
				"frontend-Host-test2-docker-localhost-1": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{},
					BasicAuth:      []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test2-docker-localhost-1": {
							Rule: "Host:test2.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"server-test1": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
						"server-test2": {
							URL:    "http://127.0.0.1:80",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
					LoadBalancer:   nil,
				},
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
	}

	for caseID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			var dockerDataList []dockerData
			for _, service := range test.services {
				dData := parseService(service, test.networks)
				dockerDataList = append(dockerDataList, dData)
			}

			provider := &Provider{
				Domain:           "docker.localhost",
				ExposedByDefault: true,
				SwarmMode:        true,
			}

			actualConfig := provider.buildConfiguration(dockerDataList)
			require.NotNil(t, actualConfig, "actualConfig")

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestSwarmTaskParsing(t *testing.T) {
	testCases := []struct {
		service       swarm.Service
		tasks         []swarm.Task
		isGlobalSVC   bool
		expectedNames map[string]string
		networks      map[string]*docker.NetworkResource
	}{
		{
			service: swarmService(serviceName("container")),
			tasks: []swarm.Task{
				swarmTask("id1", taskSlot(1)),
				swarmTask("id2", taskSlot(2)),
				swarmTask("id3", taskSlot(3)),
			},
			isGlobalSVC: false,
			expectedNames: map[string]string{
				"id1": "container.1",
				"id2": "container.2",
				"id3": "container.3",
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			service: swarmService(serviceName("container")),
			tasks: []swarm.Task{
				swarmTask("id1"),
				swarmTask("id2"),
				swarmTask("id3"),
			},
			isGlobalSVC: true,
			expectedNames: map[string]string{
				"id1": "container.id1",
				"id2": "container.id2",
				"id3": "container.id3",
			},
			networks: map[string]*docker.NetworkResource{
				"1": {
					Name: "foo",
				},
			},
		},
	}

	for caseID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			dData := parseService(test.service, test.networks)

			for _, task := range test.tasks {
				taskDockerData := parseTasks(task, dData, map[string]*docker.NetworkResource{}, test.isGlobalSVC)
				if !reflect.DeepEqual(taskDockerData.Name, test.expectedNames[task.ID]) {
					t.Errorf("expect %v, got %v", test.expectedNames[task.ID], taskDockerData.Name)
				}
			}
		})
	}
}

func TestSwarmGetFuncStringLabel(t *testing.T) {
	testCases := []struct {
		service      swarm.Service
		labelName    string
		defaultValue string
		networks     map[string]*docker.NetworkResource
		expected     string
	}{
		{
			service:      swarmService(),
			labelName:    label.TraefikWeight,
			defaultValue: label.DefaultWeight,
			networks:     map[string]*docker.NetworkResource{},
			expected:     "0",
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				label.TraefikWeight: "10",
			})),
			labelName:    label.TraefikWeight,
			defaultValue: label.DefaultWeight,
			networks:     map[string]*docker.NetworkResource{},
			expected:     "10",
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(test.labelName+strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			dData := parseService(test.service, test.networks)

			actual := getFuncStringLabel(test.labelName, test.defaultValue)(dData)
			if actual != test.expected {
				t.Errorf("got %q, expected %q", actual, test.expected)
			}
		})
	}
}
