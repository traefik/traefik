package constraints

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchTags(t *testing.T) {
	testCases := []struct {
		expr        string
		tags        []string
		expected    bool
		expectedErr bool
	}{
		{
			expr:     `Tag("world")`,
			tags:     []string{"hello", "world"},
			expected: true,
		},
		{
			expr:     `Tag("worlds")`,
			tags:     []string{"hello", "world"},
			expected: false,
		},
		{
			expr:     `!Tag("world")`,
			tags:     []string{"hello", "world"},
			expected: false,
		},
		{
			expr:     `Tag("hello") && Tag("world")`,
			tags:     []string{"hello", "world"},
			expected: true,
		},
		{
			expr:     `Tag("hello") && Tag("worlds")`,
			tags:     []string{"hello", "world"},
			expected: false,
		},
		{
			expr:     `Tag("hello") && !Tag("world")`,
			tags:     []string{"hello", "world"},
			expected: false,
		},
		{
			expr:     `Tag("hello") || Tag( "world")`,
			tags:     []string{"hello", "world"},
			expected: true,
		},
		{
			expr:     `Tag( "worlds") || Tag("hello")`,
			tags:     []string{"hello", "world"},
			expected: true,
		},
		{
			expr:     `Tag("hello") || !Tag("world")`,
			tags:     []string{"hello", "world"},
			expected: true,
		},
		{
			expr:        `Tag()`,
			tags:        []string{"hello", "world"},
			expectedErr: true,
		},
		{
			expr:        `Foo("hello")`,
			tags:        []string{"hello", "world"},
			expectedErr: true,
		},
		{
			expr:     `Tag("hello")`,
			expected: false,
		},
		{
			expr:     ``,
			expected: true,
		},
		{
			expr:     `TagRegex("hel\\w+")`,
			tags:     []string{"hello", "world"},
			expected: true,
		},
		{
			expr:     `TagRegex("hell\\w+s")`,
			tags:     []string{"hello", "world"},
			expected: false,
		},
		{
			expr:     `!TagRegex("hel\\w+")`,
			tags:     []string{"hello", "world"},
			expected: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.expr, func(t *testing.T) {
			t.Parallel()

			matches, err := MatchTags(test.tags, test.expr)
			if test.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, test.expected, matches)
		})
	}
}
