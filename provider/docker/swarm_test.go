package docker

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/containous/traefik/types"
	"github.com/davecgh/go-spew/spew"
	docker "github.com/docker/docker/api/types"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	dockerclient "github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
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
				types.LabelFrontendRule: "Headers:User-Agent,bat/0.1.0",
			})),
			expected: "Headers-User-Agent-bat-0-1-0-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				types.LabelFrontendRule: "Host:foo.bar",
			})),
			expected: "Host-foo-bar-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				types.LabelFrontendRule: "Path:/test",
			})),
			expected: "Path-test-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(
				serviceName("test"),
				serviceLabels(map[string]string{
					types.LabelFrontendRule: "PathPrefix:/test2",
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
			dockerData := parseService(test.service, test.networks)
			provider := &Provider{
				Domain:    "docker.localhost",
				SwarmMode: true,
			}
			actual := provider.getFrontendName(dockerData, 0)
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
				types.LabelFrontendRule: "Host:foo.bar",
			})),
			expected: "Host:foo.bar",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				types.LabelFrontendRule: "Path:/test",
			})),
			expected: "Path:/test",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(test.service, test.networks)
			provider := &Provider{
				Domain:    "docker.localhost",
				SwarmMode: true,
			}
			actual := provider.getFrontendRule(dockerData)
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
				types.LabelBackend: "foobar",
			})),
			expected: "foobar",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(test.service, test.networks)
			actual := getBackend(dockerData)
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
			dockerData := parseService(test.service, test.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getIPAddress(dockerData)
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
					types.LabelPort: "8080",
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
			dockerData := parseService(test.service, test.networks)
			actual := getPort(dockerData)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestSwarmGetLabel(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(),
			expected: "label not found:",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"foo": "bar",
			})),
			expected: "",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(test.service, test.networks)
			label, err := getLabel(dockerData, "foo")
			if test.expected != "" {
				if err == nil || !strings.Contains(err.Error(), test.expected) {
					t.Errorf("expected an error with %q, got %v", test.expected, err)
				}
			} else {
				if label != "bar" {
					t.Errorf("expected label 'bar', got '%s'", label)
				}
			}
		})
	}
}

func TestSwarmGetLabels(t *testing.T) {
	testCases := []struct {
		service        swarm.Service
		expectedLabels map[string]string
		expectedError  string
		networks       map[string]*docker.NetworkResource
	}{
		{
			service:        swarmService(),
			expectedLabels: map[string]string{},
			expectedError:  "label not found:",
			networks:       map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"foo": "fooz",
			})),
			expectedLabels: map[string]string{
				"foo": "fooz",
			},
			expectedError: "label not found: bar",
			networks:      map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"foo": "fooz",
				"bar": "barz",
			})),
			expectedLabels: map[string]string{
				"foo": "fooz",
				"bar": "barz",
			},
			expectedError: "",
			networks:      map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(test.service, test.networks)
			labels, err := getLabels(dockerData, []string{"foo", "bar"})
			if !reflect.DeepEqual(labels, test.expectedLabels) {
				t.Errorf("expect %v, got %v", test.expectedLabels, labels)
			}
			if test.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), test.expectedError) {
					t.Errorf("expected an error with %q, got %v", test.expectedError, err)
				}
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
				types.LabelEnable: "false",
				types.LabelPort:   "80",
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
				types.LabelFrontendRule: "Host:foo.bar",
				types.LabelPort:         "80",
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
				types.LabelPort: "80",
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
				types.LabelEnable: "true",
				types.LabelPort:   "80",
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
				types.LabelEnable: "anything",
				types.LabelPort:   "80",
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
				types.LabelFrontendRule: "Host:foo.bar",
				types.LabelPort:         "80",
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
				types.LabelPort: "80",
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
				types.LabelEnable: "true",
				types.LabelPort:   "80",
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
			dockerData := parseService(test.service, test.networks)
			actual := test.provider.containerFilter(dockerData)
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
						types.LabelPort: "80",
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
						types.LabelPort:                       "80",
						types.LabelBackend:                    "foobar",
						types.LabelFrontendEntryPoints:        "http,https",
						types.LabelFrontendAuthBasic:          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						types.LabelFrontendRedirectEntryPoint: "https",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
				swarmService(
					serviceName("test2"),
					serviceLabels(map[string]string{
						types.LabelPort:    "80",
						types.LabelBackend: "foobar",
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
				dockerData := parseService(service, test.networks)
				dockerDataList = append(dockerDataList, dockerData)
			}

			provider := &Provider{
				Domain:           "docker.localhost",
				ExposedByDefault: true,
				SwarmMode:        true,
			}
			actualConfig := provider.loadDockerConfig(dockerDataList)
			// Compare backends
			if !reflect.DeepEqual(actualConfig.Backends, test.expectedBackends) {
				t.Errorf("expected %#v, got %#v", test.expectedBackends, actualConfig.Backends)
			}
			if !reflect.DeepEqual(actualConfig.Frontends, test.expectedFrontends) {
				t.Errorf("expected %#v, got %#v", test.expectedFrontends, actualConfig.Frontends)
			}
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
			dockerData := parseService(test.service, test.networks)

			for _, task := range test.tasks {
				taskDockerData := parseTasks(task, dockerData, map[string]*docker.NetworkResource{}, test.isGlobalSVC)
				if !reflect.DeepEqual(taskDockerData.Name, test.expectedNames[task.ID]) {
					t.Errorf("expect %v, got %v", test.expectedNames[task.ID], taskDockerData.Name)
				}
			}
		})
	}
}

type fakeTasksClient struct {
	dockerclient.APIClient
	tasks []swarm.Task
	err   error
}

func (c *fakeTasksClient) TaskList(ctx context.Context, options dockertypes.TaskListOptions) ([]swarm.Task, error) {
	return c.tasks, c.err
}

func TestListTasks(t *testing.T) {
	testCases := []struct {
		service       swarm.Service
		tasks         []swarm.Task
		isGlobalSVC   bool
		expectedTasks []string
		networks      map[string]*docker.NetworkResource
	}{
		{
			service: swarmService(serviceName("container")),
			tasks: []swarm.Task{
				swarmTask("id1", taskSlot(1), taskStatus(taskState(swarm.TaskStateRunning))),
				swarmTask("id2", taskSlot(2), taskStatus(taskState(swarm.TaskStatePending))),
				swarmTask("id3", taskSlot(3)),
				swarmTask("id4", taskSlot(4), taskStatus(taskState(swarm.TaskStateRunning))),
				swarmTask("id5", taskSlot(5), taskStatus(taskState(swarm.TaskStateFailed))),
			},
			isGlobalSVC: false,
			expectedTasks: []string{
				"container.1",
				"container.4",
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
			dockerData := parseService(test.service, test.networks)
			dockerClient := &fakeTasksClient{tasks: test.tasks}
			taskDockerData, _ := listTasks(context.Background(), dockerClient, test.service.ID, dockerData, map[string]*docker.NetworkResource{}, test.isGlobalSVC)

			if len(test.expectedTasks) != len(taskDockerData) {
				t.Errorf("expected tasks %v, got %v", spew.Sdump(test.expectedTasks), spew.Sdump(taskDockerData))
			}

			for i, taskID := range test.expectedTasks {
				if taskDockerData[i].Name != taskID {
					t.Errorf("expect task id %v, got %v", taskID, taskDockerData[i].Name)
				}
			}
		})
	}
}

type fakeServicesClient struct {
	dockerclient.APIClient
	dockerVersion string
	networks      []dockertypes.NetworkResource
	services      []swarm.Service
	err           error
}

func (c *fakeServicesClient) ServiceList(ctx context.Context, options dockertypes.ServiceListOptions) ([]swarm.Service, error) {
	return c.services, c.err
}

func (c *fakeServicesClient) ServerVersion(ctx context.Context) (dockertypes.Version, error) {
	return dockertypes.Version{APIVersion: c.dockerVersion}, c.err
}

func (c *fakeServicesClient) NetworkList(ctx context.Context, options dockertypes.NetworkListOptions) ([]dockertypes.NetworkResource, error) {
	return c.networks, c.err
}

func TestListServices(t *testing.T) {
	testCases := []struct {
		desc             string
		services         []swarm.Service
		dockerVersion    string
		networks         []dockertypes.NetworkResource
		expectedServices []string
	}{
		{
			desc: "Should return no service due to no networks defined",
			services: []swarm.Service{
				swarmService(
					serviceName("service1"),
					serviceLabels(map[string]string{
						labelDockerNetwork:            "barnet",
						labelBackendLoadBalancerSwarm: "true",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(
						virtualIP("1", "10.11.12.13/24"),
						virtualIP("2", "10.11.12.99/24"),
					)),
				swarmService(
					serviceName("service2"),
					serviceLabels(map[string]string{
						labelDockerNetwork: "barnet",
					}),
					withEndpointSpec(modeDNSSR)),
			},
			dockerVersion:    "1.30",
			networks:         []dockertypes.NetworkResource{},
			expectedServices: []string{},
		},
		{
			desc: "Should return only service1",
			services: []swarm.Service{
				swarmService(
					serviceName("service1"),
					serviceLabels(map[string]string{
						labelDockerNetwork:            "barnet",
						labelBackendLoadBalancerSwarm: "true",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(
						virtualIP("yk6l57rfwizjzxxzftn4amaot", "10.11.12.13/24"),
						virtualIP("2", "10.11.12.99/24"),
					)),
				swarmService(
					serviceName("service2"),
					serviceLabels(map[string]string{
						labelDockerNetwork: "barnet",
					}),
					withEndpointSpec(modeDNSSR)),
			},
			dockerVersion: "1.30",
			networks: []dockertypes.NetworkResource{
				{
					Name:       "network_name",
					ID:         "yk6l57rfwizjzxxzftn4amaot",
					Created:    time.Now(),
					Scope:      "swarm",
					Driver:     "overlay",
					EnableIPv6: false,
					Internal:   true,
					Ingress:    false,
					ConfigOnly: false,
					Options: map[string]string{
						"com.docker.network.driver.overlay.vxlanid_list": "4098",
						"com.docker.network.enable_ipv6":                 "false",
					},
					Labels: map[string]string{
						"com.docker.stack.namespace": "test",
					},
				},
			},
			expectedServices: []string{
				"service1",
			},
		},
	}

	for caseID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			dockerClient := &fakeServicesClient{services: test.services, dockerVersion: test.dockerVersion, networks: test.networks}
			serviceDockerData, _ := listServices(context.Background(), dockerClient)

			assert.Equal(t, len(test.expectedServices), len(serviceDockerData))
			for i, serviceName := range test.expectedServices {
				assert.Equal(t, serviceName, serviceDockerData[i].Name)
			}
		})
	}
}
