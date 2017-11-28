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
