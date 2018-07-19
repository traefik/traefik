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

func TestDockerBuildConfigurationV1(t *testing.T) {
	testCases := []struct {
		desc              string
		containers        []docker.ContainerJSON
		expectedFrontends map[string]*types.Frontend
		expectedBackends  map[string]*types.Backend
	}{
		{
			desc:              "when no container",
			containers:        []docker.ContainerJSON{},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			desc: "when basic container configuration",
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
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when container has label 'enable' to false",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test"),
					labels(map[string]string{
						label.TraefikEnable:   "false",
						label.TraefikPort:     "666",
						label.TraefikProtocol: "https",
						label.TraefikWeight:   "12",
						label.TraefikBackend:  "foobar",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{},
			expectedBackends:  map[string]*types.Backend{},
		},
		{
			desc: "when all labels are set",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test1"),
					labels(map[string]string{
						label.TraefikPort:     "666",
						label.TraefikProtocol: "https",
						label.TraefikWeight:   "12",

						label.TraefikBackend: "foobar",

						label.TraefikBackendCircuitBreakerExpression:         "NetworkErrorRatio() > 0.5",
						label.TraefikBackendLoadBalancerMethod:               "drr",
						label.TraefikBackendLoadBalancerSticky:               "true",
						label.TraefikBackendLoadBalancerStickiness:           "true",
						label.TraefikBackendLoadBalancerStickinessCookieName: "chocolate",
						label.TraefikBackendMaxConnAmount:                    "666",
						label.TraefikBackendMaxConnExtractorFunc:             "client.ip",

						label.TraefikFrontendAuthBasic:            "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.TraefikFrontendEntryPoints:          "http,https",
						label.TraefikFrontendPassHostHeader:       "true",
						label.TraefikFrontendPassTLSCert:          "true",
						label.TraefikFrontendPriority:             "666",
						label.TraefikFrontendRedirectEntryPoint:   "https",
						label.TraefikFrontendRedirectRegex:        "nope",
						label.TraefikFrontendRedirectReplacement:  "nope",
						label.TraefikFrontendRule:                 "Host:traefik.io",
						label.TraefikFrontendWhitelistSourceRange: "10.10.10.10",

						label.TraefikFrontendRequestHeaders:          "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.TraefikFrontendResponseHeaders:         "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.TraefikFrontendSSLProxyHeaders:         "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.TraefikFrontendAllowedHosts:            "foo,bar,bor",
						label.TraefikFrontendHostsProxyHeaders:       "foo,bar,bor",
						label.TraefikFrontendSSLHost:                 "foo",
						label.TraefikFrontendCustomFrameOptionsValue: "foo",
						label.TraefikFrontendContentSecurityPolicy:   "foo",
						label.TraefikFrontendPublicKey:               "foo",
						label.TraefikFrontendReferrerPolicy:          "foo",
						label.TraefikFrontendSTSSeconds:              "666",
						label.TraefikFrontendSSLRedirect:             "true",
						label.TraefikFrontendSSLTemporaryRedirect:    "true",
						label.TraefikFrontendSTSIncludeSubdomains:    "true",
						label.TraefikFrontendSTSPreload:              "true",
						label.TraefikFrontendForceSTSHeader:          "true",
						label.TraefikFrontendFrameDeny:               "true",
						label.TraefikFrontendContentTypeNosniff:      "true",
						label.TraefikFrontendBrowserXSSFilter:        "true",
						label.TraefikFrontendIsDevelopment:           "true",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-Host-traefik-io-0": {
					EntryPoints: []string{
						"http",
						"https",
					},
					Backend: "backend-foobar",
					Routes: map[string]types.Route{
						"route-frontend-Host-traefik-io-0": {
							Rule: "Host:traefik.io",
						},
					},
					PassHostHeader: true,
					PassTLSCert:    true,
					Priority:       666,
					BasicAuth: []string{
						"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
						"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					},
					WhitelistSourceRange: []string{
						"10.10.10.10",
					},
					Headers: &types.Headers{
						CustomRequestHeaders: map[string]string{
							"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
							"Content-Type":                 "application/json; charset=utf-8",
						},
						CustomResponseHeaders: map[string]string{
							"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
							"Content-Type":                 "application/json; charset=utf-8",
						},
						AllowedHosts: []string{
							"foo",
							"bar",
							"bor",
						},
						HostsProxyHeaders: []string{
							"foo",
							"bar",
							"bor",
						},
						SSLRedirect:          true,
						SSLTemporaryRedirect: true,
						SSLHost:              "foo",
						SSLProxyHeaders: map[string]string{
							"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
							"Content-Type":                 "application/json; charset=utf-8",
						},
						STSSeconds:              666,
						STSIncludeSubdomains:    true,
						STSPreload:              true,
						ForceSTSHeader:          true,
						FrameDeny:               true,
						CustomFrameOptionsValue: "foo",
						ContentTypeNosniff:      true,
						BrowserXSSFilter:        true,
						ContentSecurityPolicy:   "foo",
						PublicKey:               "foo",
						ReferrerPolicy:          "foo",
						IsDevelopment:           true,
					},
					Redirect: &types.Redirect{
						EntryPoint:  "https",
						Regex:       "nope",
						Replacement: "nope",
					},
				},
			},
			expectedBackends: map[string]*types.Backend{
				"backend-foobar": {
					Servers: map[string]types.Server{
						"server-test1": {
							URL:    "https://127.0.0.1:666",
							Weight: 12,
						},
					},
					CircuitBreaker: &types.CircuitBreaker{
						Expression: "NetworkErrorRatio() > 0.5",
					},
					LoadBalancer: &types.LoadBalancer{
						Method: "drr",
						Sticky: true,
						Stickiness: &types.Stickiness{
							CookieName: "chocolate",
						},
					},
					MaxConn: &types.MaxConn{
						Amount:        666,
						ExtractorFunc: "client.ip",
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
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
			actualConfig := provider.buildConfigurationV1(dockerDataList)
			require.NotNil(t, actualConfig, "actualConfig")

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestDockerTraefikFilterV1(t *testing.T) {
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

			actual := test.provider.containerFilterV1(dData)
			if actual != test.expected {
				t.Errorf("expected %v for %+v, got %+v", test.expected, test, actual)
			}
		})
	}
}

func TestDockerGetFuncStringLabelV1(t *testing.T) {
	testCases := []struct {
		container    docker.ContainerJSON
		labelName    string
		defaultValue string
		expected     string
	}{
		{
			container:    containerJSON(),
			labelName:    label.TraefikProtocol,
			defaultValue: label.DefaultProtocol,
			expected:     "http",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikProtocol: "https",
			})),
			labelName:    label.TraefikProtocol,
			defaultValue: label.DefaultProtocol,
			expected:     "https",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(test.labelName+strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getFuncStringLabelV1(test.labelName, test.defaultValue)(dData)
			if actual != test.expected {
				t.Errorf("got %q, expected %q", actual, test.expected)
			}
		})
	}
}

func TestDockerGetSliceStringLabelV1(t *testing.T) {
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

			actual := getFuncSliceStringLabelV1(test.labelName)(dData)
			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetFrontendNameV1(t *testing.T) {
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

			actual := provider.getFrontendNameV1(dData, 0)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetFrontendRuleV1(t *testing.T) {
	testCases := []struct {
		container docker.ContainerJSON
		expected  string
	}{
		{
			container: containerJSON(name("foo")),
			expected:  "Host:foo.docker.localhost",
		},
		{
			container: containerJSON(name("foo"),
				labels(map[string]string{
					label.TraefikDomain: "traefik.localhost",
				})),
			expected: "Host:foo.traefik.localhost",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendRule: "Host:foo.bar",
			})),
			expected: "Host:foo.bar",
		},
		{
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

			actual := provider.getFrontendRuleV1(dData)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetBackendNameV1(t *testing.T) {
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

			actual := getBackendNameV1(dData)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetIPAddressV1(t *testing.T) {
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
			actual := provider.getDeprecatedIPAddress(dData)
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetPortV1(t *testing.T) {
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

			actual := getPortV1(dData)
			if actual != e.expected {
				t.Errorf("expected %q, got %q", e.expected, actual)
			}
		})
	}
}
