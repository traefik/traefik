package docker

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	docker "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerBuildConfiguration(t *testing.T) {
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
						label.TraefikBackend:                    "foobar",
						label.TraefikFrontendEntryPoints:        "http,https",
						label.TraefikFrontendAuthBasic:          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendRedirectEntryPoint: "https",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
				containerJSON(
					name("test2"),
					labels(map[string]string{
						label.TraefikBackend: "foobar",
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
						label.TraefikBackend:                         "foobar",
						label.TraefikFrontendEntryPoints:             "http,https",
						label.TraefikBackendMaxConnAmount:            "1000",
						label.TraefikBackendMaxConnExtractorFunc:     "somethingelse",
						label.TraefikBackendLoadBalancerMethod:       "drr",
						label.TraefikBackendCircuitBreakerExpression: "NetworkErrorRatio() > 0.5",
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
		{
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test1"),
					labels(map[string]string{
						label.TraefikBackend: "foobar",
						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus:  "404",
						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageBackend: "foobar",
						label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageQuery:   "foo_query",
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageStatus:  "500,600",
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageBackend: "foobar",
						label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageQuery:   "bar_query",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-test1-docker-localhost-0": {
					EntryPoints:    []string{},
					BasicAuth:      []string{},
					PassHostHeader: true,
					Backend:        "backend-foobar",
					Routes: map[string]types.Route{
						"route-frontend-Host-test1-docker-localhost-0": {
							Rule: "Host:test1.docker.localhost",
						},
					},
					Errors: map[string]types.ErrorPage{
						"foo": {
							Status:  []string{"404"},
							Query:   "foo_query",
							Backend: "foobar",
						},
						"bar": {
							Status:  []string{"500", "600"},
							Query:   "bar_query",
							Backend: "foobar",
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
				dData := parseContainer(cont)
				dockerDataList = append(dockerDataList, dData)
			}

			provider := &Provider{
				Domain:           "docker.localhost",
				ExposedByDefault: true,
			}
			actualConfig := provider.buildConfiguration(dockerDataList)
			require.NotNil(t, actualConfig, "actualConfig")

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
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
						label.TraefikEnable: "false",
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
						label.TraefikFrontendRule: "Host:foo.bar",
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
						label.TraefikPort: "80",
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
						label.TraefikEnable: "true",
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
						label.TraefikEnable: "anything",
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
						label.TraefikFrontendRule: "Host:foo.bar",
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
						label.TraefikEnable: "true",
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
						label.TraefikEnable: "true",
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
						label.TraefikEnable:       "true",
						label.TraefikFrontendRule: "Host:i.love.this.host",
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
			dData := parseContainer(test.container)
			actual := test.provider.containerFilter(dData)
			if actual != test.expected {
				t.Errorf("expected %v for %+v, got %+v", test.expected, test, actual)
			}
		})
	}
}

func TestDockerGetFuncStringLabel(t *testing.T) {
	testCases := []struct {
		container    docker.ContainerJSON
		labelName    string
		defaultValue string
		expected     string
	}{
		{
			container:    containerJSON(),
			labelName:    label.TraefikWeight,
			defaultValue: label.DefaultWeight,
			expected:     "0",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikWeight: "10",
			})),
			labelName:    label.TraefikWeight,
			defaultValue: label.DefaultWeight,
			expected:     "10",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(test.labelName+strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getFuncStringLabel(test.labelName, test.defaultValue)(dData)

			if actual != test.expected {
				t.Errorf("got %q, expected %q", actual, test.expected)
			}
		})
	}
}

func TestDockerGetSliceStringLabel(t *testing.T) {
	testCases := []struct {
		desc      string
		container docker.ContainerJSON
		labelName string
		expected  []string
	}{
		{
			desc:      "no whitelist-label",
			container: containerJSON(),
			expected:  nil,
		},
		{
			desc: "whitelist-label with empty string",
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendWhitelistSourceRange: "",
			})),
			labelName: label.TraefikFrontendWhitelistSourceRange,
			expected:  nil,
		},
		{
			desc: "whitelist-label with IPv4 mask",
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendWhitelistSourceRange: "1.2.3.4/16",
			})),
			labelName: label.TraefikFrontendWhitelistSourceRange,
			expected: []string{
				"1.2.3.4/16",
			},
		},
		{
			desc: "whitelist-label with IPv6 mask",
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendWhitelistSourceRange: "fe80::/16",
			})),
			labelName: label.TraefikFrontendWhitelistSourceRange,
			expected: []string{
				"fe80::/16",
			},
		},
		{
			desc: "whitelist-label with multiple masks",
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendWhitelistSourceRange: "1.1.1.1/24, 1234:abcd::42/32",
			})),
			labelName: label.TraefikFrontendWhitelistSourceRange,
			expected: []string{
				"1.1.1.1/24",
				"1234:abcd::42/32",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			dData := parseContainer(test.container)

			actual := getFuncSliceStringLabel(test.labelName)(dData)

			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

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
				label.TraefikFrontendRule: "Headers:User-Agent,bat/0.1.0",
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
				label.TraefikFrontendRule: "Host:foo.bar",
			})),
			expected: "Host-foo-bar-0",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendRule: "Path:/test",
			})),
			expected: "Path-test-0",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendRule: "PathPrefix:/test2",
			})),
			expected: "PathPrefix-test2-0",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dData := parseContainer(test.container)
			provider := &Provider{
				Domain: "docker.localhost",
			}
			actual := provider.getFrontendName(dData, 0)
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
				label.TraefikFrontendRule: "Host:foo.bar",
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
				label.TraefikFrontendRule: "Path:/test",
			})),
			expected: "Path:/test",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dData := parseContainer(test.container)
			provider := &Provider{
				Domain: "docker.localhost",
			}
			actual := provider.getFrontendRule(dData)
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
				label.TraefikBackend: "foobar",
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
			dData := parseContainer(test.container)
			actual := getBackend(dData)
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
			dData := parseContainer(test.container)
			provider := &Provider{}
			actual := provider.getIPAddress(dData)
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
				label.TraefikPort: "8080",
			})),
			expected: "8080",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikPort: "8080",
			}), ports(nat.PortMap{
				"80/tcp": {},
			})),
			expected: "8080",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikPort: "8080",
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
			dData := parseContainer(e.container)
			actual := getPort(dData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}

func TestGetErrorPages(t *testing.T) {
	testCases := []struct {
		desc     string
		data     dockerData
		expected map[string]*types.ErrorPage
	}{
		{
			desc: "2 errors pages",
			data: parseContainer(containerJSON(
				labels(map[string]string{
					label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus:  "404",
					label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageBackend: "foo_backend",
					label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageQuery:   "foo_query",
					label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageStatus:  "500,600",
					label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageBackend: "bar_backend",
					label.Prefix + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageQuery:   "bar_query",
				}))),
			expected: map[string]*types.ErrorPage{
				"foo": {
					Status:  []string{"404"},
					Query:   "foo_query",
					Backend: "foo_backend",
				},
				"bar": {
					Status:  []string{"500", "600"},
					Query:   "bar_query",
					Backend: "bar_backend",
				},
			},
		},
		{
			desc: "only status field",
			data: parseContainer(containerJSON(
				labels(map[string]string{
					label.Prefix + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus: "404",
				}))),
			expected: map[string]*types.ErrorPage{
				"foo": {
					Status: []string{"404"},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			pages := getErrorPages(test.data)

			assert.EqualValues(t, test.expected, pages)
		})
	}
}
