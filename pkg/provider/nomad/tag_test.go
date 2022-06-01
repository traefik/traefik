package nomad

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_tagsToLabels(t *testing.T) {
	testCases := []struct {
		desc     string
		tags     []string
		prefix   string
		expected map[string]string
	}{
		{
			desc:     "no tags",
			tags:     []string{},
			prefix:   "traefik",
			expected: map[string]string{},
		},
		{
			desc:   "minimal global config",
			tags:   []string{"traefik.enable=false"},
			prefix: "traefik",
			expected: map[string]string{
				"traefik.enable": "false",
			},
		},
		{
			desc: "config with domain",
			tags: []string{
				"traefik.enable=true",
				"traefik.domain=example.com",
			},
			prefix: "traefik",
			expected: map[string]string{
				"traefik.enable": "true",
				"traefik.domain": "example.com",
			},
		},
		{
			desc: "config with custom prefix",
			tags: []string{
				"custom.enable=true",
				"custom.domain=example.com",
			},
			prefix: "custom",
			expected: map[string]string{
				"traefik.enable": "true",
				"traefik.domain": "example.com",
			},
		},
		{
			desc: "config with spaces in tags",
			tags: []string{
				"custom.enable = true",
				"custom.domain = example.com",
			},
			prefix: "custom",
			expected: map[string]string{
				"traefik.enable": "true",
				"traefik.domain": "example.com",
			},
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

			labels := tagsToLabels(test.tags, test.prefix)

			assert.Equal(t, test.expected, labels)
		})
	}
}
