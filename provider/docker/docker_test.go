package docker

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/containous/traefik/types"
	docker "github.com/docker/engine-api/types"
	"github.com/docker/go-connections/nat"
)

func TestDockerGetFrontendName(t *testing.T) {
	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "Host-foo-docker-localhost",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.frontend.rule": "Headers:User-Agent,bat/0.1.0",
			})),
			expected: "Headers-User-Agent-bat-0-1-0",
		},
		{
			container: containerJSON(labels(map[string]string{
				"com.docker.compose.project": "foo",
				"com.docker.compose.service": "bar",
			})),
			expected: "Host-bar-foo-docker-localhost",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.frontend.rule": "Host:foo.bar",
			})),
			expected: "Host-foo-bar",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.frontend.rule": "Path:/test",
			})),
			expected: "Path-test",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.frontend.rule": "PathPrefix:/test2",
			})),
			expected: "PathPrefix-test2",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			provider := &Provider{
				Domain: "docker.localhost",
			}
			actual := provider.getFrontendName(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetFrontendRule(t *testing.T) {
	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "Host:foo.docker.localhost",
		},
		{
			container: containerJSON(name("bar")),
			expected:  "Host:bar.docker.localhost",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.frontend.rule": "Host:foo.bar",
			})),
			expected: "Host:foo.bar",
		}, {
			container: containerJSON(labels(map[string]string{
				"com.docker.compose.project": "foo",
				"com.docker.compose.service": "bar",
			})),
			expected: "Host:bar.foo.docker.localhost",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.frontend.rule": "Path:/test",
			})),
			expected: "Path:/test",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			provider := &Provider{
				Domain: "docker.localhost",
			}
			actual := provider.getFrontendRule(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetBackend(t *testing.T) {
	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "foo",
		},
		{
			container: containerJSON(name("bar")),
			expected:  "bar",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.backend": "foobar",
			})),
			expected: "foobar",
		},
		{
			container: containerJSON(labels(map[string]string{
				"com.docker.compose.project": "foo",
				"com.docker.compose.service": "bar",
			})),
			expected: "bar-foo",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			provider := &Provider{}
			actual := provider.getBackend(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetIPAddress(t *testing.T) {
	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(withNetwork("testnet", ipv4("10.11.12.13"))),
			expected:  "10.11.12.13",
		},
		{
			container: containerJSON(
				labels(map[string]string{
					"traefik.docker.network": "testnet",
				}),
				withNetwork("testnet", ipv4("10.11.12.13")),
			),
			expected: "10.11.12.13",
		},
		{
			container: containerJSON(
				labels(map[string]string{
					"traefik.docker.network": "testnet2",
				}),
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("testnet2", ipv4("10.11.12.14")),
			),
			expected: "10.11.12.14",
		},
		{
			container: containerJSON(
				networkMode("host"),
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("testnet2", ipv4("10.11.12.14")),
			),
			expected: "127.0.0.1",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			provider := &Provider{}
			actual := provider.getIPAddress(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetPort(t *testing.T) {
	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "",
		},
		{
			container: containerJSON(ports(nat.PortMap{
				"80/tcp": {},
			})),
			expected: "80",
		},
		// FIXME handle this better..
		//{
		//	container: containerJSON(ports(nat.PortMap{
		//		"80/tcp": {},
		//		"443/tcp": {},
		//	})),
		//	expected: "80",
		//},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.port": "8080",
			})),
			expected: "8080",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.port": "8080",
			}), ports(nat.PortMap{
				"80/tcp": {},
			})),
			expected: "8080",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.port": "8080",
			}), ports(nat.PortMap{
				"8080/tcp": {},
				"80/tcp":   {},
			})),
			expected: "8080",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			provider := &Provider{}
			actual := provider.getPort(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetWeight(t *testing.T) {
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
				"traefik.weight": "10",
			})),
			expected: "10",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			provider := &Provider{}
			actual := provider.getWeight(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetDomain(t *testing.T) {
	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(),
			expected:  "docker.localhost",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.domain": "foo.bar",
			})),
			expected: "foo.bar",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			provider := &Provider{
				Domain: "docker.localhost",
			}
			actual := provider.getDomain(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetProtocol(t *testing.T) {
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
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			provider := &Provider{}
			actual := provider.getProtocol(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetPassHostHeader(t *testing.T) {
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
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			provider := &Provider{}
			actual := provider.getPassHostHeader(dockerData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestDockerGetLabel(t *testing.T) {
	containers := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(),
			expected:  "Label not found:",
		},
		{
			container: containerJSON(labels(map[string]string{
				"foo": "bar",
			})),
			expected: "",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
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

func TestDockerGetLabels(t *testing.T) {
	containers := []struct {
		container      docker.ContainerJSON
		expectedLabels map[string]string
		expectedError  string
	}{
		{
			container:      containerJSON(),
			expectedLabels: map[string]string{},
			expectedError:  "Label not found:",
		},
		{
			container: containerJSON(labels(map[string]string{
				"foo": "fooz",
			})),
			expectedLabels: map[string]string{
				"foo": "fooz",
			},
			expectedError: "Label not found: bar",
		},
		{
			container: containerJSON(labels(map[string]string{
				"foo": "fooz",
				"bar": "barz",
			})),
			expectedLabels: map[string]string{
				"foo": "fooz",
				"bar": "barz",
			},
			expectedError: "",
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
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

func TestDockerTraefikFilter(t *testing.T) {
	containers := []struct {
		container        docker.ContainerJSON
		exposedByDefault bool
		expected         bool
	}{
		{
			container:        containerJSON(),
			exposedByDefault: true,
			expected:         false,
		},
		{
			container: containerJSON(
				labels(map[string]string{
					"traefik.enable": "false",
				}),
				ports(nat.PortMap{
					"80/tcp": {},
				}),
			),
			exposedByDefault: true,
			expected:         false,
		},
		{
			container: containerJSON(
				labels(map[string]string{
					"traefik.frontend.rule": "Host:foo.bar",
				}),
				ports(nat.PortMap{
					"80/tcp": {},
				}),
			),
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: containerJSON(
				ports(nat.PortMap{
					"80/tcp":  {},
					"443/tcp": {},
				}),
			),
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: containerJSON(
				ports(nat.PortMap{
					"80/tcp": {},
				}),
			),
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: containerJSON(
				labels(map[string]string{
					"traefik.port": "80",
				}),
				ports(nat.PortMap{
					"80/tcp":  {},
					"443/tcp": {},
				}),
			),
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: containerJSON(
				labels(map[string]string{
					"traefik.enable": "true",
				}),
				ports(nat.PortMap{
					"80/tcp": {},
				}),
			),
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: containerJSON(
				labels(map[string]string{
					"traefik.enable": "anything",
				}),
				ports(nat.PortMap{
					"80/tcp": {},
				}),
			),
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: containerJSON(
				labels(map[string]string{
					"traefik.frontend.rule": "Host:foo.bar",
				}),
				ports(nat.PortMap{
					"80/tcp": {},
				}),
			),
			exposedByDefault: true,
			expected:         true,
		},
		{
			container: containerJSON(
				ports(nat.PortMap{
					"80/tcp": {},
				}),
			),
			exposedByDefault: false,
			expected:         false,
		},
		{
			container: containerJSON(
				labels(map[string]string{
					"traefik.enable": "true",
				}),
				ports(nat.PortMap{
					"80/tcp": {},
				}),
			),
			exposedByDefault: false,
			expected:         true,
		},
	}

	for containerID, e := range containers {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			provider := Provider{}
			provider.ExposedByDefault = e.exposedByDefault
			dockerData := parseContainer(e.container)
			actual := provider.containerFilter(dockerData)
			if actual != e.expected {
				t.Errorf("expected %v for %+v (%+v, %+v), got %+v", e.expected, e.container, e.container.NetworkSettings, e.container.ContainerJSONBase, actual)
			}
		})
	}
}

func TestDockerLoadDockerConfig(t *testing.T) {
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
					name("test"),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
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
				},
			},
		},
		{
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test1"),
					labels(map[string]string{
						"traefik.backend":              "foobar",
						"traefik.frontend.entryPoints": "http,https",
						"traefik.frontend.auth.basic":  "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
				containerJSON(
					name("test2"),
					labels(map[string]string{
						"traefik.backend": "foobar",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
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
				},
			},
		},
		{
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test1"),
					labels(map[string]string{
						"traefik.backend":                           "foobar",
						"traefik.frontend.entryPoints":              "http,https",
						"traefik.backend.maxconn.amount":            "1000",
						"traefik.backend.maxconn.extractorfunc":     "somethingelse",
						"traefik.backend.loadbalancer.method":       "drr",
						"traefik.backend.circuitbreaker.expression": "NetworkErrorRatio() > 0.5",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test1-docker-localhost": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					BasicAuth:      []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test1-docker-localhost": {
							Rule: "Host:test1.docker.localhost",
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
					},
					CircuitBreaker: &types.CircuitBreaker{
						Expression: "NetworkErrorRatio() > 0.5",
					},
					LoadBalancer: &types.LoadBalancer{
						Method: "drr",
					},
					MaxConn: &types.MaxConn{
						Amount:        1000,
						ExtractorFunc: "somethingelse",
					},
				},
			},
		},
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

			provider := &Provider{
				Domain:           "docker.localhost",
				ExposedByDefault: true,
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
