package consulcatalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTagsToNeutralLabels(t *testing.T) {
	testCases := []struct {
		desc     string
		tags     []string
		prefix   string
		expected map[string]string
	}{
		{
			desc:     "without tags",
			expected: nil,
		},
		{
			desc:   "with a prefix",
			prefix: "test",
			tags: []string{
				"test.aaa=01",
				"test.bbb=02",
				"ccc=03",
				"test.ddd=04=to",
			},
			expected: map[string]string{
				"traefik.aaa": "01",
				"traefik.bbb": "02",
				"traefik.ddd": "04=to",
			},
		},

		{
			desc:   "with an empty prefix",
			prefix: "",
			tags: []string{
				"test.aaa=01",
				"test.bbb=02",
				"ccc=03",
				"test.ddd=04=to",
			},
			expected: map[string]string{
				"traefik.test.aaa": "01",
				"traefik.test.bbb": "02",
				"traefik.ccc":      "03",
				"traefik.test.ddd": "04=to",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			labels := tagsToNeutralLabels(test.tags, test.prefix)

			assert.Equal(t, test.expected, labels)
		})
	}
}
