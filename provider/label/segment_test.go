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
