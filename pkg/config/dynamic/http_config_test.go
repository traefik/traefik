package dynamic

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncodedCharactersMap(t *testing.T) {
	tests := []struct {
		name     string
		config   RouterDeniedEncodedPathCharacters
		expected map[string]struct{}
	}{
		{
			name: "Handles empty configuration",
			expected: map[string]struct{}{
				"%2F": {},
				"%2f": {},
				"%5C": {},
				"%5c": {},
				"%00": {},
				"%3B": {},
				"%3b": {},
				"%25": {},
				"%3F": {},
				"%3f": {},
				"%23": {},
			},
		},
		{
			name: "Exclude encoded slash when allowed",
			config: RouterDeniedEncodedPathCharacters{
				AllowEncodedSlash: true,
			},
			expected: map[string]struct{}{
				"%5C": {},
				"%5c": {},
				"%00": {},
				"%3B": {},
				"%3b": {},
				"%25": {},
				"%3F": {},
				"%3f": {},
				"%23": {},
			},
		},

		{
			name: "Exclude encoded backslash when allowed",
			config: RouterDeniedEncodedPathCharacters{
				AllowEncodedBackSlash: true,
			},
			expected: map[string]struct{}{
				"%2F": {},
				"%2f": {},
				"%00": {},
				"%3B": {},
				"%3b": {},
				"%25": {},
				"%3F": {},
				"%3f": {},
				"%23": {},
			},
		},

		{
			name: "Exclude encoded null character when allowed",
			config: RouterDeniedEncodedPathCharacters{
				AllowEncodedNullCharacter: true,
			},
			expected: map[string]struct{}{
				"%2F": {},
				"%2f": {},
				"%5C": {},
				"%5c": {},
				"%3B": {},
				"%3b": {},
				"%25": {},
				"%3F": {},
				"%3f": {},
				"%23": {},
			},
		},
		{
			name: "Exclude encoded semicolon when allowed",
			config: RouterDeniedEncodedPathCharacters{
				AllowEncodedSemicolon: true,
			},
			expected: map[string]struct{}{
				"%2F": {},
				"%2f": {},
				"%5C": {},
				"%5c": {},
				"%00": {},
				"%25": {},
				"%3F": {},
				"%3f": {},
				"%23": {},
			},
		},
		{
			name: "Exclude encoded percent when allowed",
			config: RouterDeniedEncodedPathCharacters{
				AllowEncodedPercent: true,
			},
			expected: map[string]struct{}{
				"%2F": {},
				"%2f": {},
				"%5C": {},
				"%5c": {},
				"%00": {},
				"%3B": {},
				"%3b": {},
				"%3F": {},
				"%3f": {},
				"%23": {},
			},
		},
		{
			name: "Exclude encoded question mark when allowed",
			config: RouterDeniedEncodedPathCharacters{
				AllowEncodedQuestionMark: true,
			},
			expected: map[string]struct{}{
				"%2F": {},
				"%2f": {},
				"%5C": {},
				"%5c": {},
				"%00": {},
				"%3B": {},
				"%3b": {},
				"%25": {},
				"%23": {},
			},
		},
		{
			name: "Exclude encoded hash when allowed",
			config: RouterDeniedEncodedPathCharacters{
				AllowEncodedHash: true,
			},
			expected: map[string]struct{}{
				"%2F": {},
				"%2f": {},
				"%5C": {},
				"%5c": {},
				"%00": {},
				"%3B": {},
				"%3b": {},
				"%25": {},
				"%3F": {},
				"%3f": {},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result := test.config.Map()
			require.Equal(t, test.expected, result)
		})
	}
}
