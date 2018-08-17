package forwardedheaders

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeHTTP(t *testing.T) {
	testCases := []struct {
		desc            string
		insecure        bool
		trustedIps      []string
		incomingHeaders map[string]string
		remoteAddr      string
		expectedHeaders map[string]string
	}{
		{
			desc:            "all Empty",
			insecure:        true,
			trustedIps:      nil,
			remoteAddr:      "",
			incomingHeaders: map[string]string{},
			expectedHeaders: map[string]string{
				"X-Forwarded-for": "",
			},
		},
		{
			desc:       "insecure true with incoming X-Forwarded-For",
			insecure:   true,
			trustedIps: nil,
			remoteAddr: "",
			incomingHeaders: map[string]string{
				"X-Forwarded-for": "10.0.1.0, 10.0.1.12",
			},
			expectedHeaders: map[string]string{
				"X-Forwarded-for": "10.0.1.0, 10.0.1.12",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded-For",
			insecure:   false,
			trustedIps: nil,
			remoteAddr: "",
			incomingHeaders: map[string]string{
				"X-Forwarded-for": "10.0.1.0, 10.0.1.12",
			},
			expectedHeaders: map[string]string{
				"X-Forwarded-for": "",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded-For and valid Trusted Ips",
			insecure:   false,
			trustedIps: []string{"10.0.1.100"},
			remoteAddr: "10.0.1.100:80",
			incomingHeaders: map[string]string{
				"X-Forwarded-for": "10.0.1.0, 10.0.1.12",
			},
			expectedHeaders: map[string]string{
				"X-Forwarded-for": "10.0.1.0, 10.0.1.12",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded-For and invalid Trusted Ips",
			insecure:   false,
			trustedIps: []string{"10.0.1.100"},
			remoteAddr: "10.0.1.101:80",
			incomingHeaders: map[string]string{
				"X-Forwarded-for": "10.0.1.0, 10.0.1.12",
			},
			expectedHeaders: map[string]string{
				"X-Forwarded-for": "",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded-For and valid Trusted Ips CIDR",
			insecure:   false,
			trustedIps: []string{"1.2.3.4/24"},
			remoteAddr: "1.2.3.156:80",
			incomingHeaders: map[string]string{
				"X-Forwarded-for": "10.0.1.0, 10.0.1.12",
			},
			expectedHeaders: map[string]string{
				"X-Forwarded-for": "10.0.1.0, 10.0.1.12",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded-For and invalid Trusted Ips CIDR",
			insecure:   false,
			trustedIps: []string{"1.2.3.4/24"},
			remoteAddr: "10.0.1.101:80",
			incomingHeaders: map[string]string{
				"X-Forwarded-for": "10.0.1.0, 10.0.1.12",
			},
			expectedHeaders: map[string]string{
				"X-Forwarded-for": "",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, "", nil)
			require.NoError(t, err)

			req.RemoteAddr = test.remoteAddr

			for k, v := range test.incomingHeaders {
				req.Header.Set(k, v)
			}

			m, err := NewXforwarded(test.insecure, test.trustedIps)
			require.NoError(t, err)

			m.ServeHTTP(nil, req, nil)

			for k, v := range test.expectedHeaders {
				assert.Equal(t, v, req.Header.Get(k))
			}
		})
	}
}
