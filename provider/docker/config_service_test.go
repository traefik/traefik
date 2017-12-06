package docker

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/containous/traefik/provider/label"
	docker "github.com/docker/docker/api/types"
)

func TestDockerGetFuncMapLabel(t *testing.T) {
	serviceName := "myservice"
	testCases := []struct {
		container   docker.ContainerJSON
		suffixLabel string
		expectedKey string
		expected    map[string]string
	}{
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendRequestHeaders: "X-Custom-Header: ContainerRequestHeader",
			})),
			suffixLabel: label.SuffixFrontendRequestHeaders,
			expected: map[string]string{
				"X-Custom-Header": "ContainerRequestHeader",
			},
		},
		{
			container: containerJSON(labels(map[string]string{
				label.GetServiceLabel(label.SuffixFrontendRequestHeaders, serviceName): "X-Custom-Header: ServiceRequestHeader",
			})),
			suffixLabel: label.SuffixFrontendRequestHeaders,
			expected: map[string]string{
				"X-Custom-Header": "ServiceRequestHeader",
			},
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendResponseHeaders: "X-Custom-Header: ServiceResponseHeader",
			})),
			suffixLabel: label.SuffixFrontendResponseHeaders,
			expected: map[string]string{
				"X-Custom-Header": "ServiceResponseHeader",
			},
		},
		{
			container: containerJSON(labels(map[string]string{
				label.GetServiceLabel(label.SuffixFrontendResponseHeaders, serviceName): "X-Custom-Header: ServiceResponseHeader",
			})),
			suffixLabel: label.SuffixFrontendResponseHeaders,
			expected: map[string]string{
				"X-Custom-Header": "ServiceResponseHeader",
			},
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendRequestHeaders: "X-Custom-Header: MutliRequestHeaders || Authorization: Basic YWRtaW46YWRtaW4=",
			})),
			suffixLabel: label.SuffixFrontendRequestHeaders,
			expected: map[string]string{
				"X-Custom-Header": "MutliRequestHeaders",
				"Authorization":   "Basic YWRtaW46YWRtaW4=",
			},
		},
		{
			container: containerJSON(labels(map[string]string{
				label.TraefikFrontendRequestHeaders: "X-Custom-Header: MutliResponseHeaders || Cache-Control: no-cache",
			})),
			suffixLabel: label.SuffixFrontendRequestHeaders,
			expected: map[string]string{
				"X-Custom-Header": "MutliResponseHeaders",
				"Cache-Control":   "no-cache",
			},
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(test.suffixLabel+strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			values := getFuncServiceMapLabel(test.suffixLabel)(dData, serviceName)
			for k, v := range values {
				if v != test.expected[k] {
					t.Fatalf("got %q, expected %q", v, test.expected[k])
				}
			}
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
				t.Fatalf("got %q, expected %q", actual, test.expected)
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
				t.Fatalf("for container %q: got %q, expected %q", dData.Name, actual, test.expected)
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
