package label

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractTraefikLabels(t *testing.T) {
	testCases := []struct {
		desc         string
		prefix       string
		originLabels map[string]string
		expected     SegmentProperties
	}{
		{
			desc:         "nil labels map",
			prefix:       "traefik",
			originLabels: nil,
			expected:     SegmentProperties{"": {}},
		},
		{
			desc:   "container labels",
			prefix: "traefik",
			originLabels: map[string]string{
				"frontend.priority": "foo", // missing prefix: skip
				"traefik.port":      "bar",
			},
			expected: SegmentProperties{
				"": {
					"traefik.port": "bar",
				},
			},
		},
		{
			desc:   "segment labels: only segment no default",
			prefix: "traefik",
			originLabels: map[string]string{
				"traefik.goo.frontend.priority": "A",
				"traefik.goo.port":              "D",
				"traefik.port":                  "C",
			},
			expected: SegmentProperties{
				"goo": {
					"traefik.frontend.priority": "A",
					"traefik.port":              "D",
				},
			},
		},
		{
			desc:   "segment labels: use default",
			prefix: "traefik",
			originLabels: map[string]string{
				"traefik.guu.frontend.priority": "B",
				"traefik.port":                  "C",
			},
			expected: SegmentProperties{
				"guu": {
					"traefik.frontend.priority": "B",
					"traefik.port":              "C",
				},
			},
		},
		{
			desc:   "segment labels: several segments",
			prefix: "traefik",
			originLabels: map[string]string{
				"traefik.goo.frontend.priority": "A",
				"traefik.goo.port":              "D",
				"traefik.guu.frontend.priority": "B",
				"traefik.port":                  "C",
			},
			expected: SegmentProperties{
				"goo": {
					"traefik.frontend.priority": "A",
					"traefik.port":              "D",
				},
				"guu": {
					"traefik.frontend.priority": "B",
					"traefik.port":              "C",
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual := ExtractTraefikLabels(test.originLabels)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestExtractServiceProperties(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   map[string]string
		expected SegmentProperties
	}{
		{
			desc:     "empty labels map",
			expected: SegmentProperties{},
		},
		{
			desc: "valid label names",
			labels: map[string]string{
				"traefik.foo.port":         "bar",
				"traefik.foo.frontend.bar": "1bar",
				"traefik.foo.backend":      "3bar",
			},
			expected: SegmentProperties{
				"foo": SegmentPropertyValues{
					"port":         "bar",
					"frontend.bar": "1bar",
					"backend":      "3bar",
				},
			},
		},
		{
			desc: "invalid label names",
			labels: map[string]string{
				"foo.frontend.bar":             "1bar",
				"traefik.foo.frontend.":        "2bar",
				"traefik.foo.port.bar":         "barbar",
				"traefik.foo.frontend":         "0bar",
				"traefik.frontend.foo.backend": "0bar",
			},
			expected: SegmentProperties{},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := ExtractServiceProperties(test.labels)
			assert.EqualValues(t, test.expected, got)
		})
	}
}

func TestExtractServicePropertiesP(t *testing.T) {
	testCases := []struct {
		desc     string
		labels   *map[string]string
		expected SegmentProperties
	}{
		{
			desc:     "nil labels map",
			expected: SegmentProperties{},
		},
		{
			desc: "valid label names",
			labels: &map[string]string{
				"traefik.foo.port":         "bar",
				"traefik.foo.frontend.bar": "1bar",
				"traefik.foo.backend":      "3bar",
			},
			expected: SegmentProperties{
				"foo": SegmentPropertyValues{
					"port":         "bar",
					"frontend.bar": "1bar",
					"backend":      "3bar",
				},
			},
		},
		{
			desc: "invalid label names",
			labels: &map[string]string{
				"foo.frontend.bar":             "1bar",
				"traefik.foo.frontend.":        "2bar",
				"traefik.foo.port.bar":         "barbar",
				"traefik.foo.frontend":         "0bar",
				"traefik.frontend.foo.backend": "0bar",
			},
			expected: SegmentProperties{},
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := ExtractServicePropertiesP(test.labels)
			assert.EqualValues(t, test.expected, got)
		})
	}
}

func TestGetServiceLabel(t *testing.T) {
	testCases := []struct {
		desc        string
		labelName   string
		serviceName string
		expected    string
	}{
		{
			desc:      "without service name",
			labelName: TraefikPort,
			expected:  TraefikPort,
		},
		{
			desc:        "with service name",
			labelName:   TraefikPort,
			serviceName: "bar",
			expected:    "traefik.bar.port",
		},
	}
	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := GetServiceLabel(test.labelName, test.serviceName)
			assert.Equal(t, test.expected, got)
		})
	}
}
