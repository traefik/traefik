package docker

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
	docker "github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

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
