package static

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEntryPointProtocol(t *testing.T) {
	tests := []struct {
		name             string
		address          string
		expectedAddress  string
		expectedProtocol string
		expectedError    bool
	}{
		{
			name:             "Without protocol",
			address:          "127.0.0.1:8080",
			expectedAddress:  "127.0.0.1:8080",
			expectedProtocol: "tcp",
			expectedError:    false,
		},
		{
			name:             "With TCP protocol in upper case",
			address:          "127.0.0.1:8080/TCP",
			expectedAddress:  "127.0.0.1:8080",
			expectedProtocol: "tcp",
			expectedError:    false,
		},
		{
			name:             "With UDP protocol in upper case",
			address:          "127.0.0.1:8080/UDP",
			expectedAddress:  "127.0.0.1:8080",
			expectedProtocol: "udp",
			expectedError:    false,
		},
		{
			name:             "With UDP protocol in weird case",
			address:          "127.0.0.1:8080/uDp",
			expectedAddress:  "127.0.0.1:8080",
			expectedProtocol: "udp",
			expectedError:    false,
		},

		{
			name:          "With invalid protocol",
			address:       "127.0.0.1:8080/toto/tata",
			expectedError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ep := EntryPoint{
				Address: tt.address,
			}
			protocol, err := ep.GetProtocol()
			if tt.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedProtocol, protocol)
			require.Equal(t, tt.expectedAddress, ep.GetAddress())
		})
	}
}

func TestEncodedCharactersMap(t *testing.T) {
	tests := []struct {
		name     string
		config   EncodedCharacters
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
			config: EncodedCharacters{
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
			config: EncodedCharacters{
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
			config: EncodedCharacters{
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
			config: EncodedCharacters{
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
			config: EncodedCharacters{
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
			config: EncodedCharacters{
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
			config: EncodedCharacters{
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
