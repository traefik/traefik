package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/ip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIPWhiteLister(t *testing.T) {
	testCases := []struct {
		desc          string
		whiteList     []string
		expectedError string
	}{
		{
			desc:          "invalid IP",
			whiteList:     []string{"foo"},
			expectedError: "parsing CIDR whitelist [foo]: parsing CIDR trusted IPs <nil>: invalid CIDR address: foo",
		},
		{
			desc:          "valid IP",
			whiteList:     []string{"10.10.10.10"},
			expectedError: "",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			whiteLister, err := NewIPWhiteLister(test.whiteList, &ip.RemoteAddrStrategy{})

			if len(test.expectedError) > 0 {
				assert.EqualError(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, whiteLister)
			}
		})
	}
}

func TestIPWhiteLister_ServeHTTP(t *testing.T) {
	testCases := []struct {
		desc       string
		whiteList  []string
		remoteAddr string
		expected   int
	}{
		{
			desc:       "authorized with remote address",
			whiteList:  []string{"20.20.20.20"},
			remoteAddr: "20.20.20.20:1234",
			expected:   200,
		},
		{
			desc:       "non authorized with remote address",
			whiteList:  []string{"20.20.20.20"},
			remoteAddr: "20.20.20.21:1234",
			expected:   403,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			whiteLister, err := NewIPWhiteLister(test.whiteList, &ip.RemoteAddrStrategy{})
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "http://10.10.10.10", nil)

			if len(test.remoteAddr) > 0 {
				req.RemoteAddr = test.remoteAddr
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			whiteLister.ServeHTTP(recorder, req, next)

			assert.Equal(t, test.expected, recorder.Code)
		})
	}
}
