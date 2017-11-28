package docker

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/containous/traefik/types"
	docker "github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
)

func TestDockerGetServicePort(t *testing.T) {
	testCases := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(),
			expected:  "",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelPort: "2500",
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

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(test.container)
			actual := getServicePort(dockerData, "myservice")
			if actual != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetServiceFrontendRule(t *testing.T) {
	provider := &Provider{}

	testCases := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelFrontendRule: "Path:/helloworld",
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

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(test.container)
			actual := provider.getServiceFrontendRule(dockerData, "myservice")
			if actual != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetServiceBackend(t *testing.T) {
	testCases := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "foo-foo-myservice",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelBackend: "another-backend",
			})),
			expected: "fake-another-backend-myservice",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.frontend.backend": "custom-backend",
			})),
			expected: "fake-custom-backend",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(test.container)
			actual := getServiceBackend(dockerData, "myservice")
			if actual != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerLoadDockerServiceConfig(t *testing.T) {
	testCases := []struct {
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
						"traefik.service.frontend.redirect":    "https",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-foo-foo-service": {
					Backend:        "backend-foo-foo-service",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					BasicAuth:      []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
					Redirect:       "https",
					Routes: map[string]types.Route{
						"service-service": {
							Rule: "Host:foo.docker.localhost",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foo-foo-service": {
					Servers: map[string]types.Server{
						"service-0": {
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
						"traefik.service.frontend.redirect":       "https",
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
				"frontend-test1-foobar": {
					Backend:        "backend-test1-foobar",
					PassHostHeader: false,
					Priority:       5000,
					EntryPoints:    []string{"http", "https", "ws"},
					BasicAuth:      []string{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
					Redirect:       "https",
					Routes: map[string]types.Route{
						"service-service": {
							Rule: "Path:/mypath",
						},
					},
				},
				"frontend-test2-test2-anotherservice": {
					Backend:        "backend-test2-test2-anotherservice",
					PassHostHeader: true,
					EntryPoints:    []string{},
					BasicAuth:      []string{},
					Redirect:       "",
					Routes: map[string]types.Route{
						"service-anotherservice": {
							Rule: "Path:/anotherpath",
						},
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-test1-foobar": {
					Servers: map[string]types.Server{
						"service-0": {
							URL:    "https://127.0.0.1:2503",
							Weight: 80,
						},
					},
					CircuitBreaker: nil,
				},
				"backend-test2-test2-anotherservice": {
					Servers: map[string]types.Server{
						"service-0": {
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

	for caseID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			var dockerDataList []dockerData
			for _, container := range test.containers {
				dockerData := parseContainer(container)
				dockerDataList = append(dockerDataList, dockerData)
			}

			actualConfig := provider.loadDockerConfig(dockerDataList)
			// Compare backends
			if !reflect.DeepEqual(actualConfig.Backends, test.expectedBackends) {
				t.Fatalf("expected %#v, got %#v", test.expectedBackends, actualConfig.Backends)
			}
			if !reflect.DeepEqual(actualConfig.Frontends, test.expectedFrontends) {
				t.Fatalf("expected %#v, got %#v", test.expectedFrontends, actualConfig.Frontends)
			}
		})
	}
}
