package constraints

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchLabels(t *testing.T) {
	testCases := []struct {
		expr        string
		labels      map[string]string
		expected    bool
		expectedErr bool
	}{
		{
			expr: `Label("hello", "world")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: true,
		},
		{
			expr: `Label("hello", "worlds")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: false,
		},
		{
			expr: `Label("hi", "world")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: false,
		},
		{
			expr: `!Label("hello", "world")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: false,
		},
		{
			expr: `Label("hello", "world") && Label("foo", "bar")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: true,
		},
		{
			expr: `Label("hello", "worlds") && Label("foo", "bar")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: false,
		},
		{
			expr: `Label("hello", "world") && !Label("foo", "bar")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: false,
		},
		{
			expr: `Label("hello", "world") || Label("foo", "bar")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: true,
		},
		{
			expr: `Label("hello", "worlds") || Label("foo", "bar")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: true,
		},
		{
			expr: `Label("hello", "world") || !Label("foo", "bar")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: true,
		},
		{
			expr: `Label("hello")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expectedErr: true,
		},
		{
			expr: `Foo("hello")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expectedErr: true,
		},
		{
			expr:     `Label("hello", "bar")`,
			expected: false,
		},
		{
			expr:     ``,
			expected: true,
		},
		{
			expr: `MarathonConstraint("bar")`,
			labels: map[string]string{
				"hello":                         "world",
				MarathonConstraintPrefix + "-1": "bar",
				MarathonConstraintPrefix + "-2": "foo",
			},
			expected: true,
		},
		{
			expr: `MarathonConstraint("bur")`,
			labels: map[string]string{
				"hello":                         "world",
				MarathonConstraintPrefix + "-1": "bar",
				MarathonConstraintPrefix + "-2": "foo",
			},
			expected: false,
		},
		{
			expr: `Label("hello", "world") && MarathonConstraint("bar")`,
			labels: map[string]string{
				"hello":                         "world",
				MarathonConstraintPrefix + "-1": "bar",
				MarathonConstraintPrefix + "-2": "foo",
			},
			expected: true,
		},
		{
			expr: `LabelRegex("hello", "w\\w+")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: true,
		},
		{
			expr: `LabelRegex("hello", "w\\w+s")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: false,
		},
		{
			expr: `LabelRegex("hi", "w\\w+")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: false,
		},
		{
			expr: `!LabelRegex("hello", "w\\w+")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: false,
		},
		{
			expr: `LabelRegex("hello", "w(\\w+")`,
			labels: map[string]string{
				"hello": "world",
				"foo":   "bar",
			},
			expected: false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.expr, func(t *testing.T) {
			t.Parallel()

			matches, err := MatchLabels(test.labels, test.expr)
			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, test.expected, matches)
		})
	}
}
