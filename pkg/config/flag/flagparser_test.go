package flag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		desc     string
		args     []string
		element  interface{}
		expected map[string]string
	}{
		{
			desc:     "no args",
			args:     nil,
			expected: map[string]string{},
		},
		{
			desc: "bool value",
			args: []string{"--foo"},
			element: &struct {
				Foo bool
			}{},
			expected: map[string]string{
				"traefik.foo": "true",
			},
		},
		{
			desc: "equal",
			args: []string{"--foo=bar"},
			element: &struct {
				Foo string
			}{},
			expected: map[string]string{
				"traefik.foo": "bar",
			},
		},
		{
			desc: "space separated",
			args: []string{"--foo", "bar"},
			element: &struct {
				Foo string
			}{},
			expected: map[string]string{
				"traefik.foo": "bar",
			},
		},
		{
			desc: "space separated with end of parameter",
			args: []string{"--foo=bir", "--", "--bar"},
			element: &struct {
				Foo string
			}{},
			expected: map[string]string{
				"traefik.foo": "bir",
			},
		},
		{
			desc: "multiple bool flags without value",
			args: []string{"--foo", "--bar"},
			element: &struct {
				Foo bool
				Bar bool
			}{},
			expected: map[string]string{
				"traefik.foo": "true",
				"traefik.bar": "true",
			},
		},
		{
			desc: "slice with several flags",
			args: []string{"--foo=bar", "--foo=baz"},
			element: &struct {
				Foo []string
			}{},
			expected: map[string]string{
				"traefik.foo": "bar,baz",
			},
		},
		{
			desc: "map string",
			args: []string{"--foo.name=bar"},
			element: &struct {
				Foo map[string]string
			}{},
			expected: map[string]string{
				"traefik.foo.name": "bar",
			},
		},
		{
			desc: "map struct",
			args: []string{"--foo.name.value=bar"},
			element: &struct {
				Foo map[string]struct{ Value string }
			}{},
			expected: map[string]string{
				"traefik.foo.name.value": "bar",
			},
		},
		{
			desc: "map struct with sub-struct",
			args: []string{"--foo.name.bar.value=bar"},
			element: &struct {
				Foo map[string]struct {
					Bar *struct{ Value string }
				}
			}{},
			expected: map[string]string{
				"traefik.foo.name.bar.value": "bar",
			},
		},
		{
			desc: "map struct with sub-map",
			args: []string{"--foo.name1.bar.name2.value=bar"},
			element: &struct {
				Foo map[string]struct {
					Bar map[string]struct{ Value string }
				}
			}{},
			expected: map[string]string{
				"traefik.foo.name1.bar.name2.value": "bar",
			},
		},
		{
			desc: "slice with several flags 2",
			args: []string{"--foo", "bar", "--foo", "baz"},
			element: &struct {
				Foo []string
			}{},
			expected: map[string]string{
				"traefik.foo": "bar,baz",
			},
		},
		{
			desc: "slice with several flags 3",
			args: []string{"--foo", "bar", "--foo=", "--baz"},
			element: &struct {
				Foo []string
				Baz bool
			}{},
			expected: map[string]string{
				"traefik.foo": "bar,",
				"traefik.baz": "true",
			},
		},
		{
			desc: "slice with several flags 4",
			args: []string{"--foo", "bar", "--foo", "--baz"},
			element: &struct {
				Foo []string
				Baz bool
			}{},
			expected: map[string]string{
				"traefik.foo": "bar,--baz",
			},
		},
		{
			desc: "multiple string flag",
			element: &struct {
				Foo string
			}{},
			args: []string{"--foo=bar", "--foo=baz"},
			expected: map[string]string{
				"traefik.foo": "baz",
			},
		},
		{
			desc: "multiple string flag 2",
			element: &struct {
				Foo string
			}{},
			args: []string{"--foo", "bar", "--foo", "baz"},
			expected: map[string]string{
				"traefik.foo": "baz",
			},
		},
		{
			desc: "string without value",
			element: &struct {
				Foo string
				Bar bool
			}{},
			args: []string{"--foo", "--bar"},
			expected: map[string]string{
				"traefik.foo": "--bar",
			},
		},
		{
			desc: "struct pointer value",
			args: []string{"--foo"},
			element: &struct {
				Foo *struct{ Field string }
			}{},
			expected: map[string]string{
				"traefik.foo": "true",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			fl, err := Parse(test.args, test.element)
			require.NoError(t, err)
			assert.Equal(t, test.expected, fl)
		})
	}
}

func TestParse_Errors(t *testing.T) {
	testCases := []struct {
		desc    string
		args    []string
		element interface{}
	}{
		{
			desc: "triple hyphen",
			args: []string{"---foo"},
			element: &struct {
				Foo bool
			}{},
		},
		{
			desc: "equal",
			args: []string{"--=foo"},
			element: &struct {
				Foo bool
			}{},
		},
		{
			desc: "string without value",
			element: &struct {
				Foo string
				Bar bool
			}{},
			args: []string{"--foo"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			_, err := Parse(test.args, test.element)
			require.Error(t, err)
		})
	}
}
