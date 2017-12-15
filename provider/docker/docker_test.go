package docker

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/containous/traefik/types"
	docker "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
)

func TestDockerGetFrontendName(t *testing.T) {
	testCases := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "Host-foo-docker-localhost-0",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelFrontendRule: "Headers:User-Agent,bat/0.1.0",
			})),
			expected: "Headers-User-Agent-bat-0-1-0-0",
		},
		{
			container: containerJSON(labels(map[string]string{
				"com.docker.compose.project": "foo",
				"com.docker.compose.service": "bar",
			})),
			expected: "Host-bar-foo-docker-localhost-0",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelFrontendRule: "Host:foo.bar",
			})),
			expected: "Host-foo-bar-0",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelFrontendRule: "Path:/test",
			})),
			expected: "Path-test-0",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelFrontendRule: "PathPrefix:/test2",
			})),
			expected: "PathPrefix-test2-0",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(test.container)
			provider := &Provider{
				Domain: "docker.localhost",
			}
			actual := provider.getFrontendName(dockerData, 0)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetFrontendRule(t *testing.T) {
	testCases := []struct {
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
				types.LabelFrontendRule: "Host:foo.bar",
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
				types.LabelFrontendRule: "Path:/test",
			})),
			expected: "Path:/test",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(test.container)
			provider := &Provider{
				Domain: "docker.localhost",
			}
			actual := provider.getFrontendRule(dockerData)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetBackend(t *testing.T) {
	testCases := []struct {
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
				types.LabelBackend: "foobar",
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

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(test.container)
			actual := getBackend(dockerData)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetIPAddress(t *testing.T) {
	testCases := []struct {
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
					labelDockerNetwork: "testnet",
				}),
				withNetwork("testnet", ipv4("10.11.12.13")),
			),
			expected: "10.11.12.13",
		},
		{
			container: containerJSON(
				labels(map[string]string{
					labelDockerNetwork: "testnet2",
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
		{
			container: containerJSON(
				networkMode("host"),
			),
			expected: "127.0.0.1",
		},
		{
			container: containerJSON(
				networkMode("host"),
				nodeIP("10.0.0.5"),
			),
			expected: "10.0.0.5",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(test.container)
			provider := &Provider{}
			actual := provider.getIPAddress(dockerData)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetPort(t *testing.T) {
	testCases := []struct {
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
		{
			container: containerJSON(ports(nat.PortMap{
				"80/tcp":  {},
				"443/tcp": {},
			})),
			expected: "80",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelPort: "8080",
			})),
			expected: "8080",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelPort: "8080",
			}), ports(nat.PortMap{
				"80/tcp": {},
			})),
			expected: "8080",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelPort: "8080",
			}), ports(nat.PortMap{
				"8080/tcp": {},
				"80/tcp":   {},
			})),
			expected: "8080",
		},
	}

	for containerID, e := range testCases {
		e := e
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(e.container)
			actual := getPort(dockerData)
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
			expected:  "label not found:",
		},
		{
			container: containerJSON(labels(map[string]string{
				"foo": "bar",
			})),
			expected: "",
		},
	}

	for containerID, test := range containers {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(test.container)
			label, err := getLabel(dockerData, "foo")
			if test.expected != "" {
				if err == nil || !strings.Contains(err.Error(), test.expected) {
					t.Errorf("expected an error with %q, got %v", test.expected, err)
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
	testCases := []struct {
		container      docker.ContainerJSON
		expectedLabels map[string]string
		expectedError  string
	}{
		{
			container:      containerJSON(),
			expectedLabels: map[string]string{},
			expectedError:  "label not found:",
		},
		{
			container: containerJSON(labels(map[string]string{
				"foo": "fooz",
			})),
			expectedLabels: map[string]string{
				"foo": "fooz",
			},
			expectedError: "label not found: bar",
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

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(test.container)
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

func TestDockerTraefikFilter(t *testing.T) {
	testCases := []struct {
		container docker.ContainerJSON
		expected  bool
		provider  *Provider
	}{
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config:          &container.Config{},
				NetworkSettings: &docker.NetworkSettings{},
			},
			expected: false,
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						types.LabelEnable: "false",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						types.LabelFrontendRule: "Host:foo.bar",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container-multi-ports",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp":  {},
							"443/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						types.LabelPort: "80",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp":  {},
							"443/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						types.LabelEnable: "true",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						types.LabelEnable: "anything",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						types.LabelFrontendRule: "Host:foo.bar",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: true,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: false,
			},
			expected: false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						types.LabelEnable: "true",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				Domain:           "test",
				ExposedByDefault: false,
			},
			expected: true,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						types.LabelEnable: "true",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				ExposedByDefault: false,
			},
			expected: false,
		},
		{
			container: docker.ContainerJSON{
				ContainerJSONBase: &docker.ContainerJSONBase{
					Name: "container",
				},
				Config: &container.Config{
					Labels: map[string]string{
						types.LabelEnable:       "true",
						types.LabelFrontendRule: "Host:i.love.this.host",
					},
				},
				NetworkSettings: &docker.NetworkSettings{
					NetworkSettingsBase: docker.NetworkSettingsBase{
						Ports: nat.PortMap{
							"80/tcp": {},
						},
					},
				},
			},
			provider: &Provider{
				ExposedByDefault: false,
			},
			expected: true,
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(test.container)
			actual := test.provider.containerFilter(dockerData)
			if actual != test.expected {
				t.Errorf("expected %v for %+v, got %+v", test.expected, test, actual)
			}
		})
	}
}

func TestDockerLoadDockerConfig(t *testing.T) {
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
					name("test"),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
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
				},
			},
		},
		{
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test1"),
					labels(map[string]string{
						types.LabelBackend:                    "foobar",
						types.LabelFrontendEntryPoints:        "http,https",
						types.LabelFrontendAuthBasic:          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						types.LabelFrontendRedirectEntryPoint: "https",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
				containerJSON(
					name("test2"),
					labels(map[string]string{
						types.LabelBackend: "foobar",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
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
				},
			},
		},
		{
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test1"),
					labels(map[string]string{
						types.LabelBackend:                         "foobar",
						types.LabelFrontendEntryPoints:             "http,https",
						types.LabelBackendMaxconnAmount:            "1000",
						types.LabelBackendMaxconnExtractorfunc:     "somethingelse",
						types.LabelBackendLoadbalancerMethod:       "drr",
						types.LabelBackendCircuitbreakerExpression: "NetworkErrorRatio() > 0.5",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test1-docker-localhost-0": {
					Backend:        "backend-foobar",
					PassHostHeader: true,
					EntryPoints:    []string{"http", "https"},
					BasicAuth:      []string{},
					Routes: map[string]types.Route{
						"route-frontend-Host-test1-docker-localhost-0": {
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

	for caseID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
			t.Parallel()
			var dockerDataList []dockerData
			for _, cont := range test.containers {
				dockerData := parseContainer(cont)
				dockerDataList = append(dockerDataList, dockerData)
			}

			provider := &Provider{
				Domain:           "docker.localhost",
				ExposedByDefault: true,
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

func TestDockerHasStickinessLabel(t *testing.T) {
	testCases := []struct {
		desc      string
		container docker.ContainerJSON
		expected  bool
	}{
		{
			desc:      "no stickiness-label",
			container: containerJSON(),
			expected:  false,
		},
		{
			desc: "stickiness true",
			container: containerJSON(labels(map[string]string{
				types.LabelBackendLoadbalancerStickiness: "true",
			})),
			expected: true,
		},
		{
			desc: "stickiness false",
			container: containerJSON(labels(map[string]string{
				types.LabelBackendLoadbalancerStickiness: "false",
			})),
			expected: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			dockerData := parseContainer(test.container)
			actual := hasStickinessLabel(dockerData)
			assert.Equal(t, actual, test.expected)
		})
	}
}

func TestDockerCheckPortLabels(t *testing.T) {
	testCases := []struct {
		container     docker.ContainerJSON
		expectedError bool
	}{
		{
			container: containerJSON(labels(map[string]string{
				types.LabelPort: "80",
			})),
			expectedError: false,
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelPrefix + "servicename.protocol": "http",
				types.LabelPrefix + "servicename.port":     "80",
			})),
			expectedError: false,
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelPrefix + "servicename.protocol": "http",
				types.LabelPort:                            "80",
			})),
			expectedError: false,
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelPrefix + "servicename.protocol": "http",
			})),
			expectedError: true,
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dockerData := parseContainer(test.container)
			err := checkServiceLabelPort(dockerData)

			if test.expectedError && err == nil {
				t.Error("expected an error but got nil")
			} else if !test.expectedError && err != nil {
				t.Errorf("expected no error, got %q", err)
			}
		})
	}
}
