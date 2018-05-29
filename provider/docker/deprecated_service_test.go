package docker

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	docker "github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerServiceBuildConfigurationV1(t *testing.T) {
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
			desc: "simple configuration",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("foo"),
					labels(map[string]string{
						"traefik.service.port":                 "2503",
						"traefik.service.frontend.entryPoints": "http,https",
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
					BasicAuth:      []string{},
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
							Weight: label.DefaultWeight,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "when all labels are set",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("foo"),
					labels(map[string]string{
						label.Prefix + "service." + label.SuffixPort:     "666",
						label.Prefix + "service." + label.SuffixProtocol: "https",
						label.Prefix + "service." + label.SuffixWeight:   "12",

						label.Prefix + "service." + label.SuffixFrontendAuthBasic:            "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						label.Prefix + "service." + label.SuffixFrontendEntryPoints:          "http,https",
						label.Prefix + "service." + label.SuffixFrontendPassHostHeader:       "true",
						label.Prefix + "service." + label.SuffixFrontendPassTLSCert:          "true",
						label.Prefix + "service." + label.SuffixFrontendPriority:             "666",
						label.Prefix + "service." + label.SuffixFrontendRedirectEntryPoint:   "https",
						label.Prefix + "service." + label.SuffixFrontendRedirectRegex:        "nope",
						label.Prefix + "service." + label.SuffixFrontendRedirectReplacement:  "nope",
						label.Prefix + "service." + label.SuffixFrontendRedirectPermanent:    "true",
						label.Prefix + "service." + label.SuffixFrontendWhitelistSourceRange: "10.10.10.10",

						label.Prefix + "service." + label.SuffixFrontendRequestHeaders:                 "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.Prefix + "service." + label.SuffixFrontendResponseHeaders:                "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.Prefix + "service." + label.SuffixFrontendHeadersSSLProxyHeaders:         "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
						label.Prefix + "service." + label.SuffixFrontendHeadersAllowedHosts:            "foo,bar,bor",
						label.Prefix + "service." + label.SuffixFrontendHeadersHostsProxyHeaders:       "foo,bar,bor",
						label.Prefix + "service." + label.SuffixFrontendHeadersSSLHost:                 "foo",
						label.Prefix + "service." + label.SuffixFrontendHeadersCustomFrameOptionsValue: "foo",
						label.Prefix + "service." + label.SuffixFrontendHeadersContentSecurityPolicy:   "foo",
						label.Prefix + "service." + label.SuffixFrontendHeadersPublicKey:               "foo",
						label.Prefix + "service." + label.SuffixFrontendHeadersReferrerPolicy:          "foo",
						label.Prefix + "service." + label.SuffixFrontendHeadersCustomBrowserXSSValue:   "foo",
						label.Prefix + "service." + label.SuffixFrontendHeadersSTSSeconds:              "666",
						label.Prefix + "service." + label.SuffixFrontendHeadersSSLRedirect:             "true",
						label.Prefix + "service." + label.SuffixFrontendHeadersSSLTemporaryRedirect:    "true",
						label.Prefix + "service." + label.SuffixFrontendHeadersSTSIncludeSubdomains:    "true",
						label.Prefix + "service." + label.SuffixFrontendHeadersSTSPreload:              "true",
						label.Prefix + "service." + label.SuffixFrontendHeadersForceSTSHeader:          "true",
						label.Prefix + "service." + label.SuffixFrontendHeadersFrameDeny:               "true",
						label.Prefix + "service." + label.SuffixFrontendHeadersContentTypeNosniff:      "true",
						label.Prefix + "service." + label.SuffixFrontendHeadersBrowserXSSFilter:        "true",
						label.Prefix + "service." + label.SuffixFrontendHeadersIsDevelopment:           "true",
					}),
					ports(nat.PortMap{
						"80/tcp": {},
					}),
					withNetwork("bridge", ipv4("127.0.0.1")),
				),
			},
			expectedFrontends: map[string]*types.Frontend{
				"frontend-foo-foo-service": {
					Backend: "backend-foo-foo-service",
					EntryPoints: []string{
						"http",
						"https",
					},
					PassHostHeader: true,
					PassTLSCert:    true,
					Priority:       666,
					BasicAuth: []string{
						"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
						"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
					},
					Redirect: &types.Redirect{
						EntryPoint:  "https",
						Regex:       "nope",
						Replacement: "nope",
					},
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
							URL:    "https://127.0.0.1:666",
							Weight: 12,
						},
					},
					CircuitBreaker: nil,
				},
			},
		},
		{
			desc: "several containers",
			containers: []docker.ContainerJSON{
				containerJSON(
					name("test1"),
					labels(map[string]string{
						"traefik.service.port":                         "2503",
						"traefik.service.protocol":                     "https",
						"traefik.service.weight":                       "80",
						"traefik.service.backend":                      "foobar",
						"traefik.service.frontend.passHostHeader":      "false",
						"traefik.service.frontend.rule":                "Path:/mypath",
						"traefik.service.frontend.priority":            "5000",
						"traefik.service.frontend.entryPoints":         "http,https,ws",
						"traefik.service.frontend.auth.basic":          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						"traefik.service.frontend.redirect.entryPoint": "https",
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
					Redirect: &types.Redirect{
						EntryPoint: "https",
					},
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

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			var dockerDataList []dockerData
			for _, container := range test.containers {
				dData := parseContainer(container)
				dockerDataList = append(dockerDataList, dData)
			}

			actualConfig := provider.buildConfigurationV1(dockerDataList)
			require.NotNil(t, actualConfig, "actualConfig")

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestDockerGetFuncServiceStringLabelV1(t *testing.T) {
	testCases := []struct {
		container    docker.ContainerJSON
		suffixLabel  string
		defaultValue string
		expected     string
	}{
		{
			container:    containerJSON(),
			suffixLabel:  label.SuffixProtocol,
			defaultValue: label.DefaultProtocol,
			expected:     "http",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikProtocol: "https",
			})),
			suffixLabel:  label.SuffixProtocol,
			defaultValue: label.DefaultProtocol,
			expected:     "https",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.Prefix + "myservice." + label.SuffixProtocol: "https",
			})),
			suffixLabel:  label.SuffixProtocol,
			defaultValue: label.DefaultProtocol,
			expected:     "https",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(test.suffixLabel+strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getFuncServiceStringLabelV1(test.suffixLabel, test.defaultValue)(dData, "myservice")
			if actual != test.expected {
				t.Errorf("got %q, expected %q", actual, test.expected)
			}
		})
	}
}

func TestDockerGetFuncServiceSliceStringLabelV1(t *testing.T) {
	testCases := []struct {
		container   docker.ContainerJSON
		suffixLabel string
		expected    []string
	}{
		{
			container:   containerJSON(),
			suffixLabel: label.SuffixFrontendEntryPoints,
			expected:    nil,
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendEntryPoints: "http,https",
			})),
			suffixLabel: label.SuffixFrontendEntryPoints,
			expected:    []string{"http", "https"},
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.frontend.entryPoints": "http,https",
			})),
			suffixLabel: label.SuffixFrontendEntryPoints,
			expected:    []string{"http", "https"},
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(test.suffixLabel+strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getFuncServiceSliceStringLabelV1(test.suffixLabel)(dData, "myservice")

			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("for container %q: got %q, expected %q", dData.Name, actual, test.expected)
			}
		})
	}
}

func TestDockerGetServiceStringValueV1(t *testing.T) {
	testCases := []struct {
		desc          string
		container     docker.ContainerJSON
		serviceLabels map[string]string
		labelSuffix   string
		defaultValue  string
		expected      string
	}{
		{
			desc: "should use service label when label exists in service labels",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					"traefik.foo": "bir",
				})),
			serviceLabels: map[string]string{
				"foo": "bar",
			},
			labelSuffix:  "foo",
			defaultValue: "fail",
			expected:     "bar",
		},
		{
			desc: "should use container label when label doesn't exist in service labels",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					"traefik.foo": "bir",
				})),
			serviceLabels: map[string]string{
				"fo": "bar",
			},
			labelSuffix:  "foo",
			defaultValue: "fail",
			expected:     "bir",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getServiceStringValueV1(dData, test.serviceLabels, test.labelSuffix, test.defaultValue)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetServiceSliceValueV1(t *testing.T) {
	testCases := []struct {
		desc          string
		container     docker.ContainerJSON
		serviceLabels map[string]string
		labelSuffix   string
		expected      []string
	}{
		{
			desc: "should return nil when no label",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{})),
			serviceLabels: map[string]string{},
			expected:      nil,
		},
		{
			desc: "should return a slice when label",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					"traefik.foo": "bor, byr, ber",
				})),
			serviceLabels: map[string]string{
				"foo": "bar, bir, bur",
			},
			labelSuffix: "foo",
			expected:    []string{"bar", "bir", "bur"},
		},
		{
			desc: "should return a slice when label (fallback to container labels)",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					"traefik.foo": "bor, byr, ber",
				})),
			serviceLabels: map[string]string{
				"fo": "bar, bir, bur",
			},
			labelSuffix: "foo",
			expected:    []string{"bor", "byr", "ber"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getServiceSliceValueV1(dData, test.serviceLabels, test.labelSuffix)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetServiceBoolValueV1(t *testing.T) {
	testCases := []struct {
		desc          string
		container     docker.ContainerJSON
		serviceLabels map[string]string
		labelSuffix   string
		defaultValue  bool
		expected      bool
	}{
		{
			desc: "should return default value when no label",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{})),
			serviceLabels: map[string]string{},
			labelSuffix:   "foo",
			defaultValue:  true,
			expected:      true,
		},
		{
			desc: "should return a bool when label",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					"traefik.foo": "false",
				})),
			serviceLabels: map[string]string{
				"foo": "true",
			},
			labelSuffix: "foo",
			expected:    true,
		},
		{
			desc: "should return a bool when label (fallback to container labels)",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					"traefik.foo": "true",
				})),
			serviceLabels: map[string]string{
				"fo": "false",
			},
			labelSuffix: "foo",
			expected:    true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getServiceBoolValueV1(dData, test.serviceLabels, test.labelSuffix, test.defaultValue)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerCheckPortLabelsV1(t *testing.T) {
	testCases := []struct {
		container     docker.ContainerJSON
		expectedError bool
	}{
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikPort: "80",
			})),
			expectedError: false,
		},
		{
			container: containerJSON(labels(map[string]string{
				label.Prefix + "servicename.protocol": "http",
				label.Prefix + "servicename.port":     "80",
			})),
			expectedError: false,
		},
		{
			container: containerJSON(labels(map[string]string{
				label.Prefix + "servicename.protocol": "http",
				label.TraefikPort:                     "80",
			})),
			expectedError: false,
		},
		{
			container: containerJSON(labels(map[string]string{
				label.Prefix + "servicename.protocol": "http",
			})),
			expectedError: true,
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)
			err := checkServiceLabelPortV1(dData)

			if test.expectedError && err == nil {
				t.Error("expected an error but got nil")
			} else if !test.expectedError && err != nil {
				t.Errorf("expected no error, got %q", err)
			}
		})
	}
}

func TestDockerGetServiceBackendNameV1(t *testing.T) {
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
				label.TraefikBackend: "another-backend",
			})),
			expected: "fake-another-backend-myservice",
		},
		{
			container: containerJSON(name("foo.bar")),
			expected:  "foo-bar-foo-bar-myservice",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.backend": "custom-backend",
			})),
			expected: "fake-custom-backend",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikBackend: "another.backend",
			})),
			expected: "fake-another-backend-myservice",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()
			dData := parseContainer(test.container)
			actual := getServiceBackendNameV1(dData, "myservice")
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetServiceFrontendRuleV1(t *testing.T) {
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
				label.TraefikFrontendRule: "Path:/helloworld",
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
			dData := parseContainer(test.container)
			actual := provider.getServiceFrontendRuleV1(dData, "myservice")
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetServicePortV1(t *testing.T) {
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
				label.TraefikPort: "2500",
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
			dData := parseContainer(test.container)
			actual := getServicePortV1(dData, "myservice")
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}
