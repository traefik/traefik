package docker

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	docker "github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerServiceBuildConfiguration(t *testing.T) {
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
							Weight: 0,
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

						label.Prefix + "service." + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus:  "404",
						label.Prefix + "service." + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageBackend: "foobar",
						label.Prefix + "service." + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageQuery:   "foo_query",
						label.Prefix + "service." + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageStatus:  "500,600",
						label.Prefix + "service." + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageBackend: "foobar",
						label.Prefix + "service." + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageQuery:   "bar_query",

						label.Prefix + "service." + label.SuffixFrontendRateLimitExtractorFunc:                          "client.ip",
						label.Prefix + "service." + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  "6",
						label.Prefix + "service." + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: "12",
						label.Prefix + "service." + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   "18",
						label.Prefix + "service." + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  "3",
						label.Prefix + "service." + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: "6",
						label.Prefix + "service." + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   "9",
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

					Errors: map[string]*types.ErrorPage{
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
					RateLimit: &types.RateLimit{
						ExtractorFunc: "client.ip",
						RateSet: map[string]*types.Rate{
							"foo": {
								Period:  flaeg.Duration(6 * time.Second),
								Average: 12,
								Burst:   18,
							},
							"bar": {
								Period:  flaeg.Duration(3 * time.Second),
								Average: 6,
								Burst:   9,
							},
						},
					},
					Redirect: &types.Redirect{
						EntryPoint:  "https",
						Regex:       "",
						Replacement: "",
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
						"traefik.service.frontend.backend":             "foobar",
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

			actualConfig := provider.buildConfiguration(dockerDataList)
			require.NotNil(t, actualConfig, "actualConfig")

			assert.EqualValues(t, test.expectedBackends, actualConfig.Backends)
			assert.EqualValues(t, test.expectedFrontends, actualConfig.Frontends)
		})
	}
}

func TestDockerGetFuncServiceStringLabel(t *testing.T) {
	testCases := []struct {
		container    docker.ContainerJSON
		suffixLabel  string
		defaultValue string
		expected     string
	}{
		{
			container:    containerJSON(),
			suffixLabel:  label.SuffixWeight,
			defaultValue: label.DefaultWeight,
			expected:     "0",
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikWeight: "200",
			})),
			suffixLabel:  label.SuffixWeight,
			defaultValue: label.DefaultWeight,
			expected:     "200",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.weight": "31337",
			})),
			suffixLabel:  label.SuffixWeight,
			defaultValue: label.DefaultWeight,
			expected:     "31337",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(test.suffixLabel+strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getFuncServiceStringLabel(test.suffixLabel, test.defaultValue)(dData, "myservice")
			if actual != test.expected {
				t.Errorf("got %q, expected %q", actual, test.expected)
			}
		})
	}
}

func TestDockerGetFuncServiceSliceStringLabel(t *testing.T) {
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

			actual := getFuncServiceSliceStringLabel(test.suffixLabel)(dData, "myservice")

			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("for container %q: got %q, expected %q", dData.Name, actual, test.expected)
			}
		})
	}
}

func TestDockerGetServiceStringValue(t *testing.T) {
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

			actual := getServiceStringValue(dData, test.serviceLabels, test.labelSuffix, test.defaultValue)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerHasStrictServiceLabel(t *testing.T) {
	testCases := []struct {
		desc          string
		serviceLabels map[string]string
		labelSuffix   string
		expected      bool
	}{
		{
			desc:          "should return false when service don't have label",
			serviceLabels: map[string]string{},
			labelSuffix:   "",
			expected:      false,
		},
		{
			desc: "should return true when service have label",
			serviceLabels: map[string]string{
				"foo": "bar",
			},
			labelSuffix: "foo",
			expected:    true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := hasStrictServiceLabel(test.serviceLabels, test.labelSuffix)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetStrictServiceStringValue(t *testing.T) {
	testCases := []struct {
		desc          string
		serviceLabels map[string]string
		labelSuffix   string
		defaultValue  string
		expected      string
	}{
		{
			desc: "should return a string when the label exists",
			serviceLabels: map[string]string{
				"foo": "bar",
			},
			labelSuffix: "foo",
			expected:    "bar",
		},
		{
			desc: "should return a string when the label exists and value empty",
			serviceLabels: map[string]string{
				"foo": "",
			},
			labelSuffix:  "foo",
			defaultValue: "cube",
			expected:     "",
		},
		{
			desc:          "should return the default value when the label doesn't exist",
			serviceLabels: map[string]string{},
			labelSuffix:   "foo",
			defaultValue:  "cube",
			expected:      "cube",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := getStrictServiceStringValue(test.serviceLabels, test.labelSuffix, test.defaultValue)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetServiceMapValue(t *testing.T) {
	testCases := []struct {
		desc          string
		container     docker.ContainerJSON
		serviceLabels map[string]string
		serviceName   string
		labelSuffix   string
		expected      map[string]string
	}{
		{
			desc: "should return when no labels",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{})),
			serviceLabels: map[string]string{},
			serviceName:   "soo",
			labelSuffix:   "foo",
			expected:      nil,
		},
		{
			desc: "should return a map when label exists",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					"traefik.foo": "bir:fii",
				})),
			serviceLabels: map[string]string{
				"foo": "bar:foo",
			},
			serviceName: "soo",
			labelSuffix: "foo",
			expected: map[string]string{
				"Bar": "foo",
			},
		},
		{
			desc: "should return a map when label exists (fallback to container labels)",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					"traefik.foo": "bir:fii",
				})),
			serviceLabels: map[string]string{
				"fo": "bar:foo",
			},
			serviceName: "soo",
			labelSuffix: "foo",
			expected: map[string]string{
				"Bir": "fii",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getServiceMapValue(dData, test.serviceLabels, test.serviceName, test.labelSuffix)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetServiceSliceValue(t *testing.T) {
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

			actual := getServiceSliceValue(dData, test.serviceLabels, test.labelSuffix)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetServiceBoolValue(t *testing.T) {
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

			actual := getServiceBoolValue(dData, test.serviceLabels, test.labelSuffix, test.defaultValue)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetServiceInt64Value(t *testing.T) {
	testCases := []struct {
		desc          string
		container     docker.ContainerJSON
		serviceLabels map[string]string
		labelSuffix   string
		defaultValue  int64
		expected      int64
	}{
		{
			desc: "should return default value when no label",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{})),
			serviceLabels: map[string]string{},
			labelSuffix:   "foo",
			defaultValue:  666,
			expected:      666,
		},
		{
			desc: "should return a int64 when label",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					"traefik.foo": "20",
				})),
			serviceLabels: map[string]string{
				"foo": "10",
			},
			labelSuffix: "foo",
			expected:    10,
		},
		{
			desc: "should return a int64 when label (fallback to container labels)",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					"traefik.foo": "20",
				})),
			serviceLabels: map[string]string{
				"fo": "10",
			},
			labelSuffix: "foo",
			expected:    20,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getServiceInt64Value(dData, test.serviceLabels, test.labelSuffix, test.defaultValue)

			assert.Equal(t, test.expected, actual)
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
			err := checkServiceLabelPort(dData)

			if test.expectedError && err == nil {
				t.Error("expected an error but got nil")
			} else if !test.expectedError && err != nil {
				t.Errorf("expected no error, got %q", err)
			}
		})
	}
}

func TestDockerGetServiceBackendName(t *testing.T) {
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
				"traefik.myservice.frontend.backend": "custom-backend",
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
			actual := getServiceBackendName(dData, "myservice")
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
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
			actual := provider.getServiceFrontendRule(dData, "myservice")
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

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
			actual := getServicePort(dData, "myservice")
			if actual != test.expected {
				t.Errorf("expected %q, got %q", test.expected, actual)
			}
		})
	}
}

func TestDockerGetServiceRedirect(t *testing.T) {
	service := "rubiks"

	testCases := []struct {
		desc      string
		container docker.ContainerJSON
		expected  *types.Redirect
	}{
		{
			desc: "should return nil when no redirect labels",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{})),
			expected: nil,
		},
		{
			desc: "should use only entry point tag when mix regex redirect and entry point redirect",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					label.Prefix + service + "." + label.SuffixFrontendRedirectEntryPoint:  "https",
					label.Prefix + service + "." + label.SuffixFrontendRedirectRegex:       "(.*)",
					label.Prefix + service + "." + label.SuffixFrontendRedirectReplacement: "$1",
				}),
			),
			expected: &types.Redirect{
				EntryPoint: "https",
			},
		},
		{
			desc: "should return a struct when entry point redirect label",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					label.Prefix + service + "." + label.SuffixFrontendRedirectEntryPoint: "https",
				}),
			),
			expected: &types.Redirect{
				EntryPoint: "https",
			},
		},
		{
			desc: "should return a struct when entry point redirect label (fallback to container labels)",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					label.TraefikFrontendRedirectEntryPoint: "https",
				}),
			),
			expected: &types.Redirect{
				EntryPoint: "https",
			},
		},
		{
			desc: "should return a struct when regex redirect labels",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					label.Prefix + service + "." + label.SuffixFrontendRedirectRegex:       "(.*)",
					label.Prefix + service + "." + label.SuffixFrontendRedirectReplacement: "$1",
				}),
			),
			expected: &types.Redirect{
				Regex:       "(.*)",
				Replacement: "$1",
			},
		},
		{
			desc: "should return a struct when regex redirect labels (fallback to container labels)",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					label.TraefikFrontendRedirectRegex:       "(.*)",
					label.TraefikFrontendRedirectReplacement: "$1",
				}),
			),
			expected: &types.Redirect{
				Regex:       "(.*)",
				Replacement: "$1",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getServiceRedirect(dData, service)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetServiceHeaders(t *testing.T) {
	service := "rubiks"

	testCases := []struct {
		desc      string
		container docker.ContainerJSON
		expected  *types.Headers
	}{
		{
			desc: "should return nil when no custom headers options are set",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{})),
			expected: nil,
		},
		{
			desc: "should return a struct when all custom headers options are set",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					label.Prefix + service + "." + label.SuffixFrontendRequestHeaders:                 "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
					label.Prefix + service + "." + label.SuffixFrontendResponseHeaders:                "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
					label.Prefix + service + "." + label.SuffixFrontendHeadersSSLProxyHeaders:         "Access-Control-Allow-Methods:POST,GET,OPTIONS || Content-type: application/json; charset=utf-8",
					label.Prefix + service + "." + label.SuffixFrontendHeadersAllowedHosts:            "foo,bar,bor",
					label.Prefix + service + "." + label.SuffixFrontendHeadersHostsProxyHeaders:       "foo,bar,bor",
					label.Prefix + service + "." + label.SuffixFrontendHeadersSSLHost:                 "foo",
					label.Prefix + service + "." + label.SuffixFrontendHeadersCustomFrameOptionsValue: "foo",
					label.Prefix + service + "." + label.SuffixFrontendHeadersContentSecurityPolicy:   "foo",
					label.Prefix + service + "." + label.SuffixFrontendHeadersPublicKey:               "foo",
					label.Prefix + service + "." + label.SuffixFrontendHeadersReferrerPolicy:          "foo",
					label.Prefix + service + "." + label.SuffixFrontendHeadersSTSSeconds:              "666",
					label.Prefix + service + "." + label.SuffixFrontendHeadersSSLRedirect:             "true",
					label.Prefix + service + "." + label.SuffixFrontendHeadersSSLTemporaryRedirect:    "true",
					label.Prefix + service + "." + label.SuffixFrontendHeadersSTSIncludeSubdomains:    "true",
					label.Prefix + service + "." + label.SuffixFrontendHeadersSTSPreload:              "true",
					label.Prefix + service + "." + label.SuffixFrontendHeadersForceSTSHeader:          "true",
					label.Prefix + service + "." + label.SuffixFrontendHeadersFrameDeny:               "true",
					label.Prefix + service + "." + label.SuffixFrontendHeadersContentTypeNosniff:      "true",
					label.Prefix + service + "." + label.SuffixFrontendHeadersBrowserXSSFilter:        "true",
					label.Prefix + service + "." + label.SuffixFrontendHeadersIsDevelopment:           "true",
				}),
			),
			expected: &types.Headers{
				CustomRequestHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
				},
				CustomResponseHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
				},
				SSLProxyHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
				},
				AllowedHosts:            []string{"foo", "bar", "bor"},
				HostsProxyHeaders:       []string{"foo", "bar", "bor"},
				SSLHost:                 "foo",
				CustomFrameOptionsValue: "foo",
				ContentSecurityPolicy:   "foo",
				PublicKey:               "foo",
				ReferrerPolicy:          "foo",
				STSSeconds:              666,
				SSLRedirect:             true,
				SSLTemporaryRedirect:    true,
				STSIncludeSubdomains:    true,
				STSPreload:              true,
				ForceSTSHeader:          true,
				FrameDeny:               true,
				ContentTypeNosniff:      true,
				BrowserXSSFilter:        true,
				IsDevelopment:           true,
			},
		},
		{
			desc: "should return a struct when all custom headers options are set (fallback to container labels)",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
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
			),
			expected: &types.Headers{
				CustomRequestHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
				},
				CustomResponseHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
				},
				SSLProxyHeaders: map[string]string{
					"Access-Control-Allow-Methods": "POST,GET,OPTIONS",
					"Content-Type":                 "application/json; charset=utf-8",
				},
				AllowedHosts:            []string{"foo", "bar", "bor"},
				HostsProxyHeaders:       []string{"foo", "bar", "bor"},
				SSLHost:                 "foo",
				CustomFrameOptionsValue: "foo",
				ContentSecurityPolicy:   "foo",
				PublicKey:               "foo",
				ReferrerPolicy:          "foo",
				STSSeconds:              666,
				SSLRedirect:             true,
				SSLTemporaryRedirect:    true,
				STSIncludeSubdomains:    true,
				STSPreload:              true,
				ForceSTSHeader:          true,
				FrameDeny:               true,
				ContentTypeNosniff:      true,
				BrowserXSSFilter:        true,
				IsDevelopment:           true,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getServiceHeaders(dData, service)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetServiceRateLimit(t *testing.T) {
	service := "rubiks"

	testCases := []struct {
		desc      string
		container docker.ContainerJSON
		expected  *types.RateLimit
	}{
		{
			desc: "should return nil when no rate limit labels",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{})),
			expected: nil,
		},
		{
			desc: "should return a struct when rate limit labels are defined",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					label.Prefix + service + "." + label.SuffixFrontendRateLimitExtractorFunc:                          "client.ip",
					label.Prefix + service + "." + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  "6",
					label.Prefix + service + "." + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: "12",
					label.Prefix + service + "." + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   "18",
					label.Prefix + service + "." + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  "3",
					label.Prefix + service + "." + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: "6",
					label.Prefix + service + "." + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   "9",
				})),
			expected: &types.RateLimit{
				ExtractorFunc: "client.ip",
				RateSet: map[string]*types.Rate{
					"foo": {
						Period:  flaeg.Duration(6 * time.Second),
						Average: 12,
						Burst:   18,
					},
					"bar": {
						Period:  flaeg.Duration(3 * time.Second),
						Average: 6,
						Burst:   9,
					},
				},
			},
		},
		{
			desc: "should return nil when ExtractorFunc is missing",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  "6",
					label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: "12",
					label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   "18",
					label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  "3",
					label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: "6",
					label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   "9",
				})),
			expected: nil,
		},
		{
			desc: "should return a struct when rate limit labels are defined (fallback to container labels)",
			container: containerJSON(
				name("test1"),
				labels(map[string]string{
					label.TraefikFrontendRateLimitExtractorFunc:                                        "client.ip",
					label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitPeriod:  "6",
					label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitAverage: "12",
					label.Prefix + label.BaseFrontendRateLimit + "foo." + label.SuffixRateLimitBurst:   "18",
					label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitPeriod:  "3",
					label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitAverage: "6",
					label.Prefix + label.BaseFrontendRateLimit + "bar." + label.SuffixRateLimitBurst:   "9",
				})),
			expected: &types.RateLimit{
				ExtractorFunc: "client.ip",
				RateSet: map[string]*types.Rate{
					"foo": {
						Period:  flaeg.Duration(6 * time.Second),
						Average: 12,
						Burst:   18,
					},
					"bar": {
						Period:  flaeg.Duration(3 * time.Second),
						Average: 6,
						Burst:   9,
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			actual := getServiceRateLimit(dData, service)

			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDockerGetServiceErrorPages(t *testing.T) {
	service := "courgette"
	testCases := []struct {
		desc     string
		data     dockerData
		expected map[string]*types.ErrorPage
	}{
		{
			desc: "2 errors pages",
			data: parseContainer(containerJSON(
				labels(map[string]string{
					label.Prefix + service + "." + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageStatus:  "404",
					label.Prefix + service + "." + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageBackend: "foo_backend",
					label.Prefix + service + "." + label.BaseFrontendErrorPage + "foo." + label.SuffixErrorPageQuery:   "foo_query",
					label.Prefix + service + "." + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageStatus:  "500,600",
					label.Prefix + service + "." + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageBackend: "bar_backend",
					label.Prefix + service + "." + label.BaseFrontendErrorPage + "bar." + label.SuffixErrorPageQuery:   "bar_query",
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
					label.Prefix + service + ".frontend.errors.foo.status": "404",
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

			pages := getServiceErrorPages(test.data, service)

			assert.EqualValues(t, test.expected, pages)
		})
	}
}
