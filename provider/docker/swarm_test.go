package docker

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/containous/traefik/types"
	"github.com/davecgh/go-spew/spew"
	dockerclient "github.com/docker/engine-api/client"
	docker "github.com/docker/engine-api/types"
	dockertypes "github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/swarm"
	"golang.org/x/net/context"
)

func TestSwarmGetFrontendName(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(serviceName("foo")),
			expected: "Host-foo-docker-localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.rule": "Headers:User-Agent,bat/0.1.0",
			})),
			expected: "Headers-User-Agent-bat-0-1-0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.rule": "Host:foo.bar",
			})),
			expected: "Host-foo-bar",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.rule": "Path:/test",
			})),
			expected: "Path-test",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(
				serviceName("test"),
				serviceLabels(map[string]string{
					"traefik.frontend.rule": "PathPrefix:/test2",
				}),
			),
			expected: "PathPrefix-test2",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				Domain:    "docker.localhost",
				SwarmMode: true,
			}
			actual := provider.getFrontendName(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetFrontendRule(t *testing.T) {
	services := []struct {
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
				"traefik.frontend.rule": "Host:foo.bar",
			})),
			expected: "Host:foo.bar",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.rule": "Path:/test",
			})),
			expected: "Path:/test",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				Domain:    "docker.localhost",
				SwarmMode: true,
			}
			actual := provider.getFrontendRule(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetBackend(t *testing.T) {
	services := []struct {
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
				"traefik.backend": "foobar",
			})),
			expected: "foobar",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getBackend(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetIPAddress(t *testing.T) {
	services := []struct {
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
					"traefik.docker.network": "barnet",
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

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getIPAddress(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetPort(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service: swarmService(
				serviceLabels(map[string]string{
					"traefik.port": "8080",
				}),
				withEndpointSpec(modeDNSSR),
			),
			expected: "8080",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getPort(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetWeight(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(),
			expected: "0",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.weight": "10",
			})),
			expected: "10",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getWeight(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetDomain(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(serviceName("foo")),
			expected: "docker.localhost",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.domain": "foo.bar",
			})),
			expected: "foo.bar",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				Domain:    "docker.localhost",
				SwarmMode: true,
			}
			actual := provider.getDomain(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetProtocol(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(),
			expected: "http",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.protocol": "https",
			})),
			expected: "https",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getProtocol(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetPassHostHeader(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(),
			expected: "true",
			networks: map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.passHostHeader": "false",
			})),
			expected: "false",
			networks: map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			actual := provider.getPassHostHeader(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestSwarmGetLabel(t *testing.T) {
	services := []struct {
		service  swarm.Service
		expected string
		networks map[string]*docker.NetworkResource
	}{
		{
			service:  swarmService(),
			expected: "Label not found:",
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

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			label, err := getLabel(dockerData, "foo")
			if e.expected != "" {
				if err == nil || !strings.Contains(err.Error(), e.expected) {
					t.Errorf("expected an error with %q, got %v", e.expected, err)
				}
			} else {
				if label != "bar" {
					t.Errorf("expected label 'bar', got %s", label)
				}
			}
		})
	}
}

func TestSwarmGetLabels(t *testing.T) {
	services := []struct {
		service        swarm.Service
		expectedLabels map[string]string
		expectedError  string
		networks       map[string]*docker.NetworkResource
	}{
		{
			service:        swarmService(),
			expectedLabels: map[string]string{},
			expectedError:  "Label not found:",
			networks:       map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"foo": "fooz",
			})),
			expectedLabels: map[string]string{
				"foo": "fooz",
			},
			expectedError: "Label not found: bar",
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

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			labels, err := getLabels(dockerData, []string{"foo", "bar"})
			if !reflect.DeepEqual(labels, e.expectedLabels) {
				t.Errorf("expect %v, got %v", e.expectedLabels, labels)
			}
			if e.expectedError != "" {
				if err == nil || !strings.Contains(err.Error(), e.expectedError) {
					t.Errorf("expected an error with %q, got %v", e.expectedError, err)
				}
			}
		})
	}
}

func TestSwarmTraefikFilter(t *testing.T) {
	services := []struct {
		service          swarm.Service
		exposedByDefault bool
		expected         bool
		networks         map[string]*docker.NetworkResource
	}{
		{
			service:          swarmService(),
			exposedByDefault: true,
			expected:         false,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.enable": "false",
				"traefik.port":   "80",
			})),
			exposedByDefault: true,
			expected:         false,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.rule": "Host:foo.bar",
				"traefik.port":          "80",
			})),
			exposedByDefault: true,
			expected:         true,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.port": "80",
			})),
			exposedByDefault: true,
			expected:         true,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.enable": "true",
				"traefik.port":   "80",
			})),
			exposedByDefault: true,
			expected:         true,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.enable": "anything",
				"traefik.port":   "80",
			})),
			exposedByDefault: true,
			expected:         true,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.frontend.rule": "Host:foo.bar",
				"traefik.port":          "80",
			})),
			exposedByDefault: true,
			expected:         true,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.port": "80",
			})),
			exposedByDefault: false,
			expected:         false,
			networks:         map[string]*docker.NetworkResource{},
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				"traefik.enable": "true",
				"traefik.port":   "80",
			})),
			exposedByDefault: false,
			expected:         true,
			networks:         map[string]*docker.NetworkResource{},
		},
	}

	for serviceID, e := range services {
		e := e
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			provider := &Provider{
				SwarmMode: true,
			}
			provider.ExposedByDefault = e.exposedByDefault
			actual := provider.containerFilter(dockerData)
			if actual != e.expected {
				t.Errorf("expected %v for %+v, got %+v", e.expected, e, actual)
			}
		})
	}
}

func TestSwarmLoadDockerConfig(t *testing.T) {
	cases := []struct {
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
						"traefik.port": "80",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test-docker-localhost": {
					Backend:        "backend-test",
					PassHostHeader: true,
					EntryPoints:    []string{},
					BasicAuth:      []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test-docker-localhost": {
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
						"traefik.port":                 "80",
						"traefik.backend":              "foobar",
						"traefik.frontend.entryPoints": "http,https",
						"traefik.frontend.auth.basic":  "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
				swarmService(
					serviceName("test2"),
					serviceLabels(map[string]string{
						"traefik.port":    "80",
						"traefik.backend": "foobar",
					}),
					withEndpointSpec(modeVIP),
					withEndpoint(virtualIP("1", "127.0.0.1/24")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test1-docker-localhost": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					BasicAuth:      []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
					Routes: map[string]types.Route{
						"route-frontend-Host-test1-docker-localhost": {
							Rule: "Host:test1.docker.localhost",
						},
					},
				},
				"frontend-Host-test2-docker-localhost": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{},
					BasicAuth:      []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test2-docker-localhost": {
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

	for caseID, c := range cases {
		c := c
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			var dockerDataList []dockerData
			for _, service := range c.services {
				dockerData := parseService(service, c.networks)
				dockerDataList = append(dockerDataList, dockerData)
			}

			provider := &Provider{
				Domain:           "docker.localhost",
				ExposedByDefault: true,
				SwarmMode:        true,
			}
			actualConfig := provider.loadDockerConfig(dockerDataList)
			// Compare backends
			if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
				t.Errorf("expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
			}
			if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
				t.Errorf("expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
			}
		})
	}
}

func TestSwarmTaskParsing(t *testing.T) {
	cases := []struct {
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

	for caseID, e := range cases {
		e := e
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)

			for _, task := range e.tasks {
				taskDockerData := parseTasks(task, dockerData, map[string]*docker.NetworkResource{}, e.isGlobalSVC)
				if !reflect.DeepEqual(taskDockerData.Name, e.expectedNames[task.ID]) {
					t.Errorf("expect %v, got %v", e.expectedNames[task.ID], taskDockerData.Name)
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
	cases := []struct {
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

	for caseID, e := range cases {
		e := e
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseService(e.service, e.networks)
			dockerClient := &fakeTasksClient{tasks: e.tasks}
			taskDockerData, _ := listTasks(context.Background(), dockerClient, e.service.ID, dockerData, map[string]*docker.NetworkResource{}, e.isGlobalSVC)

			if len(e.expectedTasks) != len(taskDockerData) {
				t.Errorf("expected tasks %v, got %v", spew.Sdump(e.expectedTasks), spew.Sdump(taskDockerData))
			}

			for i, taskID := range e.expectedTasks {
				if taskDockerData[i].Name != taskID {
					t.Errorf("expect task id %v, got %v", taskID, taskDockerData[i].Name)
				}
			}
		})
	}
}
