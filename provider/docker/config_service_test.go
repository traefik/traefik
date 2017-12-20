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

func TestDockerServiceBuildConfiguration(t *testing.T) {
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
						"traefik.service.port":                         "2503",
						"traefik.service.frontend.entryPoints":         "http,https",
						"traefik.service.frontend.auth.basic":          "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						"traefik.service.frontend.redirect.entryPoint": "https",
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
					Redirect: &types.Redirect{
						EntryPoint: "https",
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

	for caseID, test := range testCases {
		test := test
		t.Run(strconv.Itoa(caseID), func(t *testing.T) {
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

func TestDockerGetFuncMapLabel(t *testing.T) {
	serviceName := "myservice"
	fakeSuffix := "frontend.foo"
	fakeLabel := label.Prefix + fakeSuffix

	testCases := []struct {
		desc        string
		container   docker.ContainerJSON
		suffixLabel string
		expectedKey string
		expected    map[string]string
	}{
		{
			desc: "fallback to container label value",
			container: containerJSON(labels(map[string]string{
				fakeLabel: "X-Custom-Header: ContainerRequestHeader",
			})),
			suffixLabel: fakeSuffix,
			expected: map[string]string{
				"X-Custom-Header": "ContainerRequestHeader",
			},
		},
		{
			desc: "use service label instead of container label",
			container: containerJSON(labels(map[string]string{
				fakeLabel: "X-Custom-Header: ContainerRequestHeader",
				label.GetServiceLabel(fakeLabel, serviceName): "X-Custom-Header: ServiceRequestHeader",
			})),
			suffixLabel: fakeSuffix,
			expected: map[string]string{
				"X-Custom-Header": "ServiceRequestHeader",
			},
		},
		{
			desc: "use service label with an empty value instead of container label",
			container: containerJSON(labels(map[string]string{
				fakeLabel: "X-Custom-Header: ContainerRequestHeader",
				label.GetServiceLabel(fakeLabel, serviceName): "X-Custom-Header: ",
			})),
			suffixLabel: fakeSuffix,
			expected: map[string]string{
				"X-Custom-Header": "",
			},
		},
		{
			desc: "multiple values",
			container: containerJSON(labels(map[string]string{
				fakeLabel: "X-Custom-Header: MultiHeaders || Authorization: Basic YWRtaW46YWRtaW4=",
			})),
			suffixLabel: fakeSuffix,
			expected: map[string]string{
				"X-Custom-Header": "MultiHeaders",
				"Authorization":   "Basic YWRtaW46YWRtaW4=",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			values := getFuncServiceMapLabel(test.suffixLabel)(dData, serviceName)

			assert.EqualValues(t, test.expected, values)
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
				label.TraefikBackend: "another-backend",
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
			dData := parseContainer(test.container)
			actual := getServiceBackend(dData, "myservice")
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

func TestGetServiceErrorPages(t *testing.T) {
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
