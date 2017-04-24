package docker

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/containous/traefik/types"
	docker "github.com/docker/engine-api/types"
	"github.com/docker/go-connections/nat"
)

func TestDockerGetServiceProtocol(t *testing.T) {
	provider := &Provider{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(),
			expected:  "http",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.protocol": "https",
			})),
			expected: "https",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.protocol": "https",
			})),
			expected: "https",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getServiceProtocol(dockerData, "myservice")
			if actual != e.expected {
				t.Fatalf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetServiceWeight(t *testing.T) {
	provider := &Provider{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(),
			expected:  "0",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.weight": "200",
			})),
			expected: "200",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.weight": "31337",
			})),
			expected: "31337",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getServiceWeight(dockerData, "myservice")
			if actual != e.expected {
				t.Fatalf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetServicePort(t *testing.T) {
	provider := &Provider{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(),
			expected:  "",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.port": "2500",
			})),
			expected: "2500",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.port": "1234",
			})),
			expected: "1234",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getServicePort(dockerData, "myservice")
			if actual != e.expected {
				t.Fatalf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetServiceFrontendRule(t *testing.T) {
	provider := &Provider{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "Host:foo.",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.frontend.rule": "Path:/helloworld",
			})),
			expected: "Path:/helloworld",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.frontend.rule": "Path:/mycustomservicepath",
			})),
			expected: "Path:/mycustomservicepath",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getServiceFrontendRule(dockerData, "myservice")
			if actual != e.expected {
				t.Fatalf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetServiceBackend(t *testing.T) {
	provider := &Provider{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "foo-myservice",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.backend": "another-backend",
			})),
			expected: "another-backend-myservice",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.frontend.backend": "custom-backend",
			})),
			expected: "custom-backend",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getServiceBackend(dockerData, "myservice")
			if actual != e.expected {
				t.Fatalf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetServicePriority(t *testing.T) {
	provider := &Provider{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(),
			expected:  "0",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.frontend.priority": "33",
			})),
			expected: "33",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.frontend.priority": "2503",
			})),
			expected: "2503",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getServicePriority(dockerData, "myservice")
			if actual != e.expected {
				t.Fatalf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetServicePassHostHeader(t *testing.T) {
	provider := &Provider{}

	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(),
			expected:  "true",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.frontend.passHostHeader": "false",
			})),
			expected: "false",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.frontend.passHostHeader": "false",
			})),
			expected: "false",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getServicePassHostHeader(dockerData, "myservice")
			if actual != e.expected {
				t.Fatalf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetServiceEntryPoints(t *testing.T) {
	provider := &Provider{}

	containers := []struct {
		container docker.ContainerJSON
		expected  []string
	}{
		{
			container: containerJSON(),
			expected:  []string{},
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.frontend.entryPoints": "http,https",
			})),
			expected: []string{"http", "https"},
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.frontend.entryPoints": "http,https",
			})),
			expected: []string{"http", "https"},
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := provider.getServiceEntryPoints(dockerData, "myservice")
			if !reflect.DeepEqual(actual, e.expected) {
				t.Fatalf("expected %q, got %q for container %q", e.expected, actual, dockerData.Name)
			}
		})
	}
}

func TestDockerLoadDockerServiceConfig(t *testing.T) {
	cases := []struct {
		containers        []docker.ContainerJSON
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			containers:        []docker.ContainerJSON{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			containers: []docker.ContainerJSON{
				containerJSON(
					name("foo"),
					labels(map[string]string{
						"traefik.service.port":                 "2503",
						"traefik.service.frontend.entryPoints": "http,https",
						"traefik.service.frontend.auth.basic":  "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-foo-service": {
					Backend:        "backend-foo-service",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					BasicAuth:      []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
					Routes: map[string]types.Route{
						"service-service": {
							Rule: "Host:foo.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foo-service": {
					Servers: map[string]types.Server{
						"service": {
							URL:    "http://127.0.0.1:2503",
							Weight: 0,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test1"),
					labels(map[string]string{
						"traefik.service.port":                    "2503",
						"traefik.service.protocol":                "https",
						"traefik.service.weight":                  "80",
						"traefik.service.frontend.backend":        "foobar",
						"traefik.service.frontend.passHostHeader": "false",
						"traefik.service.frontend.rule":           "Path:/mypath",
						"traefik.service.frontend.priority":       "5000",
						"traefik.service.frontend.entryPoints":    "http,https,ws",
						"traefik.service.frontend.auth.basic":     "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
				containerJSON(
					name("test2"),
					labels(map[string]string{
						"traefik.anotherservice.port":          "8079",
						"traefik.anotherservice.weight":        "33",
						"traefik.anotherservice.frontend.rule": "Path:/anotherpath",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-foobar": {
					Backend:        "backend-foobar",
					PassHostHeader: false,
					Priority:       5000,
					EntryPoints:    []string{"http", "https", "ws"},
					BasicAuth:      []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
					Routes: map[string]types.Route{
						"service-service": {
							Rule: "Path:/mypath",
						},
					},
				},
				"frontend-test2-anotherservice": {
					Backend:        "backend-test2-anotherservice",
					PassHostHeader: true,
					EntryPoints:    []string{},
					BasicAuth:      []string{},
					Routes: map[string]types.Route{
						"service-anotherservice": {
							Rule: "Path:/anotherpath",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"service": {
							URL:    "https://127.0.0.1:2503",
							Weight: 80,
						},
					},
					CircuitBreaker: nil,
				},
				"backend-test2-anotherservice": {
					Servers: map[string]types.Server{
						"service": {
							URL:    "http://127.0.0.1:8079",
							Weight: 33,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
	}

	provider := &Provider{
		Domain:           "docker.localhost",
		ExposedByDefault: true,
	}

	for caseID, c := range cases {
		c := c
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			var dockerDataList []dockerData
			for _, container := range c.containers {
				dockerData := parseContainer(container)
				dockerDataList = append(dockerDataList, dockerData)
			}

			actualConfig := provider.loadDockerConfig(dockerDataList)
			// Compare backends
			if !reflect.DeepEqual(actualConfig.Backends, c.expectedBackends) {
				t.Fatalf("expected %#v, got %#v", c.expectedBackends, actualConfig.Backends)
			}
			if !reflect.DeepEqual(actualConfig.Frontends, c.expectedFrontends) {
				t.Fatalf("expected %#v, got %#v", c.expectedFrontends, actualConfig.Frontends)
			}
		})
	}
}
