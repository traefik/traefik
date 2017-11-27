package docker

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/containous/traefik/types"
	docker "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
)

func TestDockerGetFuncStringLabel(t *testing.T) {
	testCases := []struct {
		container    docker.ContainerJSON
		labelName    string
		defaultValue string
		expected     string
	}{
		// weight
		{
			container:    containerJSON(),
			labelName:    types.LabelWeight,
			defaultValue: defaultWeight,
			expected:     "0",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelWeight: "10",
			})),
			labelName:    types.LabelWeight,
			defaultValue: defaultWeight,
			expected:     "10",
		},
		// Domain
		{
			container:    containerJSON(),
			expected:     "docker.localhost",
			labelName:    types.LabelDomain,
			defaultValue: "docker.localhost",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelDomain: "foo.bar",
			})),
			labelName:    types.LabelDomain,
			defaultValue: "docker.localhost",
			expected:     "foo.bar",
		},
		// Protocol
		{
			container:    containerJSON(),
			labelName:    types.LabelProtocol,
			defaultValue: defaultProtocol,
			expected:     "http",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelProtocol: "https",
			})),
			labelName:    types.LabelProtocol,
			defaultValue: defaultProtocol,
			expected:     "https",
		},
		// FrontendPassHostHeader
		{
			container:    containerJSON(),
			labelName:    types.LabelFrontendPassHostHeader,
			defaultValue: defaultPassHostHeader,
			expected:     "true",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelFrontendPassHostHeader: "false",
			})),
			labelName:    types.LabelFrontendPassHostHeader,
			defaultValue: defaultPassHostHeader,
			expected:     "false",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(test.labelName+strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dockerData := parseContainer(test.container)

			actual := getFuncStringLabel(test.labelName, test.defaultValue)(dockerData)

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
				types.LabelTraefikFrontendWhitelistSourceRange: "",
			})),
			labelName: types.LabelTraefikFrontendWhitelistSourceRange,
			expected:  nil,
		},
		{
			desc: "whitelist-label with IPv4 mask",
			container: containerJSON(labels(map[string]string{
				types.LabelTraefikFrontendWhitelistSourceRange: "1.2.3.4/16",
			})),
			labelName: types.LabelTraefikFrontendWhitelistSourceRange,
			expected: []string{
				"1.2.3.4/16",
			},
		},
		{
			desc: "whitelist-label with IPv6 mask",
			container: containerJSON(labels(map[string]string{
				types.LabelTraefikFrontendWhitelistSourceRange: "fe80::/16",
			})),
			labelName: types.LabelTraefikFrontendWhitelistSourceRange,
			expected: []string{
				"fe80::/16",
			},
		},
		{
			desc: "whitelist-label with multiple masks",
			container: containerJSON(labels(map[string]string{
				types.LabelTraefikFrontendWhitelistSourceRange: "1.1.1.1/24, 1234:abcd::42/32",
			})),
			labelName: types.LabelTraefikFrontendWhitelistSourceRange,
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
			dockerData := parseContainer(test.container)

			actual := getFuncSliceStringLabel(test.labelName)(dockerData)

			if !reflect.DeepEqual(actual, test.expected) {
				t.Errorf("expected %q, got %q", test.expected, actual)
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
		// Weight
		{
			container:    containerJSON(),
			suffixLabel:  types.SuffixWeight,
			defaultValue: defaultWeight,
			expected:     "0",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelWeight: "200",
			})),
			suffixLabel:  types.SuffixWeight,
			defaultValue: defaultWeight,
			expected:     "200",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.weight": "31337",
			})),
			suffixLabel:  types.SuffixWeight,
			defaultValue: defaultWeight,
			expected:     "31337",
		},
		// Protocol
		{
			container:    containerJSON(),
			suffixLabel:  types.SuffixProtocol,
			defaultValue: defaultProtocol,
			expected:     "http",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelProtocol: "https",
			})),
			suffixLabel:  types.SuffixProtocol,
			defaultValue: defaultProtocol,
			expected:     "https",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.protocol": "https",
			})),
			suffixLabel:  types.SuffixProtocol,
			defaultValue: defaultProtocol,
			expected:     "https",
		},
		// Priority
		{
			container:    containerJSON(),
			suffixLabel:  types.SuffixFrontendPriority,
			defaultValue: defaultFrontendPriority,
			expected:     "0",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelFrontendPriority: "33",
			})),
			suffixLabel:  types.SuffixFrontendPriority,
			defaultValue: defaultFrontendPriority,
			expected:     "33",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.frontend.priority": "2503",
			})),
			suffixLabel:  types.SuffixFrontendPriority,
			defaultValue: defaultFrontendPriority,
			expected:     "2503",
		},
		// FrontendPassHostHeader
		{
			container:    containerJSON(),
			suffixLabel:  types.SuffixFrontendPassHostHeader,
			defaultValue: defaultPassHostHeader,
			expected:     "true",
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelFrontendPassHostHeader: "false",
			})),
			suffixLabel:  types.SuffixFrontendPassHostHeader,
			defaultValue: defaultPassHostHeader,
			expected:     "false",
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.frontend.passHostHeader": "false",
			})),
			suffixLabel:  types.SuffixFrontendPassHostHeader,
			defaultValue: defaultPassHostHeader,
			expected:     "false",
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(test.suffixLabel+strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dockerData := parseContainer(test.container)

			actual := getFuncServiceStringLabel(test.suffixLabel, test.defaultValue)(dockerData, "myservice")
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
			suffixLabel: types.SuffixFrontendEntryPoints,
			expected:    nil,
		},
		{
			container: containerJSON(labels(map[string]string{
				types.LabelFrontendEntryPoints: "http,https",
			})),
			suffixLabel: types.SuffixFrontendEntryPoints,
			expected:    []string{"http", "https"},
		},
		{
			container: containerJSON(labels(map[string]string{
				"traefik.myservice.frontend.entryPoints": "http,https",
			})),
			suffixLabel: types.SuffixFrontendEntryPoints,
			expected:    []string{"http", "https"},
		},
	}

	for containerID, test := range testCases {
		test := test
		t.Run(test.suffixLabel+strconv.Itoa(containerID), func(t *testing.T) {
			t.Parallel()

			dockerData := parseContainer(test.container)

			actual := getFuncServiceSliceStringLabel(test.suffixLabel)(dockerData, "myservice")

			if !reflect.DeepEqual(actual, test.expected) {
				t.Fatalf("for container %q: got %q, expected %q", dockerData.Name, actual, test.expected)
			}
		})
	}
}

func TestSwarmGetFuncStringLabel(t *testing.T) {
	testCases := []struct {
		service      swarm.Service
		labelName    string
		defaultValue string
		networks     map[string]*docker.NetworkResource
		expected     string
	}{
		// Weight
		{
			service:      swarmService(),
			labelName:    types.LabelWeight,
			defaultValue: defaultWeight,
			networks:     map[string]*docker.NetworkResource{},
			expected:     "0",
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				types.LabelWeight: "10",
			})),
			labelName:    types.LabelWeight,
			defaultValue: defaultWeight,
			networks:     map[string]*docker.NetworkResource{},
			expected:     "10",
		},
		// Domain
		{
			service:      swarmService(serviceName("foo")),
			networks:     map[string]*docker.NetworkResource{},
			labelName:    types.LabelDomain,
			defaultValue: "docker.localhost",
			expected:     "docker.localhost",
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				types.LabelDomain: "foo.bar",
			})),
			labelName:    types.LabelDomain,
			defaultValue: "docker.localhost",
			networks:     map[string]*docker.NetworkResource{},
			expected:     "foo.bar",
		},
		// Protocol
		{
			service:      swarmService(),
			labelName:    types.LabelProtocol,
			defaultValue: defaultProtocol,
			networks:     map[string]*docker.NetworkResource{},
			expected:     "http",
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				types.LabelProtocol: "https",
			})),
			labelName:    types.LabelProtocol,
			defaultValue: defaultProtocol,
			networks:     map[string]*docker.NetworkResource{},
			expected:     "https",
		},
		// FrontendPassHostHeader
		{
			service:      swarmService(),
			networks:     map[string]*docker.NetworkResource{},
			labelName:    types.LabelFrontendPassHostHeader,
			defaultValue: defaultPassHostHeader,
			expected:     "true",
		},
		{
			service: swarmService(serviceLabels(map[string]string{
				types.LabelFrontendPassHostHeader: "false",
			})),
			labelName:    types.LabelFrontendPassHostHeader,
			defaultValue: defaultPassHostHeader,
			networks:     map[string]*docker.NetworkResource{},
			expected:     "false",
		},
	}

	for serviceID, test := range testCases {
		test := test
		t.Run(test.labelName+strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			dockerData := parseService(test.service, test.networks)

			actual := getFuncStringLabel(test.labelName, test.defaultValue)(dockerData)
			if actual != test.expected {
				t.Errorf("got %q, expected %q", actual, test.expected)
			}
		})
	}
}
