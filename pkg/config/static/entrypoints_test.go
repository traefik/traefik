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

func TestEntryPoints_HasEncodedCharactersRestriction(t *testing.T) {
	tests := []struct {
		name        string
		entryPoints EntryPoints
		expected    bool
	}{
		{
			name:        "no entry points",
			entryPoints: EntryPoints{},
			expected:    false,
		},
		{
			name: "entry point without encoded characters configuration",
			entryPoints: EntryPoints{
				"web": {HTTP: HTTPConfig{}},
			},
			expected: false,
		},
		{
			name: "entry point with default (permissive) encoded characters configuration",
			entryPoints: EntryPoints{
				"web": {HTTP: HTTPConfig{EncodedCharacters: &EncodedCharacters{
					AllowEncodedSlash:         true,
					AllowEncodedBackSlash:     true,
					AllowEncodedNullCharacter: true,
					AllowEncodedSemicolon:     true,
					AllowEncodedPercent:       true,
					AllowEncodedQuestionMark:  true,
					AllowEncodedHash:          true,
				}}},
			},
			expected: false,
		},
		{
			name: "entry point restricting one encoded character",
			entryPoints: EntryPoints{
				"web": {HTTP: HTTPConfig{EncodedCharacters: &EncodedCharacters{
					AllowEncodedSlash:         false,
					AllowEncodedBackSlash:     true,
					AllowEncodedNullCharacter: true,
					AllowEncodedSemicolon:     true,
					AllowEncodedPercent:       true,
					AllowEncodedQuestionMark:  true,
					AllowEncodedHash:          true,
				}}},
			},
			expected: true,
		},
		{
			name: "only one entry point out of many restricts an encoded character",
			entryPoints: EntryPoints{
				"web": {HTTP: HTTPConfig{}},
				"websecure": {HTTP: HTTPConfig{EncodedCharacters: &EncodedCharacters{
					AllowEncodedHash: false,
				}}},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.entryPoints.HasEncodedCharactersRestriction())
		})
	}
}
