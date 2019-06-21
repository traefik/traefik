package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindPrefixedEnvVars(t *testing.T) {
	testCases := []struct {
		desc     string
		environ  []string
		element  interface{}
		expected []string
	}{
		{
			desc:     "exact name",
			environ:  []string{"TRAEFIK_FOO"},
			element:  &Yo{},
			expected: []string{"TRAEFIK_FOO"},
		},
		{
			desc:     "prefixed name",
			environ:  []string{"TRAEFIK_FII01"},
			element:  &Yo{},
			expected: []string{"TRAEFIK_FII01"},
		},
		{
			desc:     "excluded env vars",
			environ:  []string{"TRAEFIK_NOPE", "TRAEFIK_NO"},
			element:  &Yo{},
			expected: nil,
		},
		{
			desc:     "filter",
			environ:  []string{"TRAEFIK_NOPE", "TRAEFIK_NO", "TRAEFIK_FOO", "TRAEFIK_FII01"},
			element:  &Yo{},
			expected: []string{"TRAEFIK_FOO", "TRAEFIK_FII01"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			vars := FindPrefixedEnvVars(test.environ, DefaultNamePrefix, test.element)

			assert.Equal(t, test.expected, vars)
		})
	}
}

func Test_getRootFieldNames(t *testing.T) {
	testCases := []struct {
		desc     string
		element  interface{}
		expected []string
	}{
		{
			desc:     "simple fields",
			element:  &Yo{},
			expected: []string{"TRAEFIK_FOO", "TRAEFIK_FII", "TRAEFIK_FUU", "TRAEFIK_YI", "TRAEFIK_YU"},
		},
		{
			desc:     "embedded struct",
			element:  &Yu{},
			expected: []string{"TRAEFIK_FOO", "TRAEFIK_FII", "TRAEFIK_FUU"},
		},
		{
			desc:     "embedded struct pointer",
			element:  &Ye{},
			expected: []string{"TRAEFIK_FOO", "TRAEFIK_FII", "TRAEFIK_FUU"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			names := getRootPrefixes(test.element, DefaultNamePrefix)

			assert.Equal(t, test.expected, names)
		})
	}
}
