package provider

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitAndTrimString(t *testing.T) {
	cases := []struct {
		desc     string
		input    string
		expected []string
	}{
		{
			desc:     "empty string",
			input:    "",
			expected: nil,
		}, {
			desc:     "one piece",
			input:    "foo",
			expected: []string{"foo"},
		}, {
			desc:     "two pieces",
			input:    "foo,bar",
			expected: []string{"foo", "bar"},
		}, {
			desc:     "three pieces",
			input:    "foo,bar,zoo",
			expected: []string{"foo", "bar", "zoo"},
		}, {
			desc:     "two pieces with whitespace",
			input:    " foo   ,  bar     ",
			expected: []string{"foo", "bar"},
		}, {
			desc:     "consecutive commas",
			input:    " foo   ,,  bar     ",
			expected: []string{"foo", "bar"},
		}, {
			desc:     "consecutive commas with witespace",
			input:    " foo   , ,  bar     ",
			expected: []string{"foo", "bar"},
		}, {
			desc:     "leading and trailing commas",
			input:    ",, foo   , ,  bar,, , ",
			expected: []string{"foo", "bar"},
		}, {
			desc:     "no valid pieces",
			input:    ",  , , ,, ,",
			expected: nil,
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := SplitAndTrimString(test.input)
			assert.Equal(t, test.expected, actual)
		})
	}
}
