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
