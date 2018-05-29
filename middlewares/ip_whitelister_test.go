package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/whitelist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIPWhiteLister(t *testing.T) {
	testCases := []struct {
		desc             string
		whiteList        []string
		useXForwardedFor bool
		expectedError    string
	}{
		{
			desc:             "invalid IP",
			whiteList:        []string{"foo"},
			useXForwardedFor: false,
			expectedError:    "parsing CIDR whitelist [foo]: parsing CIDR white list <nil>: invalid CIDR address: foo",
		},
		{
			desc:             "valid IP",
			whiteList:        []string{"10.10.10.10"},
			useXForwardedFor: false,
			expectedError:    "",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			whiteLister, err := NewIPWhiteLister(test.whiteList, test.useXForwardedFor)

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
		desc             string
		whiteList        []string
		useXForwardedFor bool
		remoteAddr       string
		xForwardedFor    []string
		expected         int
	}{
		{
			desc:             "authorized with remote address",
			whiteList:        []string{"20.20.20.20"},
			useXForwardedFor: false,
			remoteAddr:       "20.20.20.20:1234",
			xForwardedFor:    nil,
			expected:         200,
		},
		{
			desc:             "non authorized with remote address",
			whiteList:        []string{"20.20.20.20"},
			useXForwardedFor: false,
			remoteAddr:       "20.20.20.21:1234",
			xForwardedFor:    nil,
			expected:         403,
		},
		{
			desc:             "non authorized with remote address (X-Forwarded-For possible)",
			whiteList:        []string{"20.20.20.20"},
			useXForwardedFor: false,
			remoteAddr:       "20.20.20.21:1234",
			xForwardedFor:    []string{"20.20.20.20", "40.40.40.40"},
			expected:         403,
		},
		{
			desc:             "authorized with X-Forwarded-For",
			whiteList:        []string{"30.30.30.30"},
			useXForwardedFor: true,
			xForwardedFor:    []string{"30.30.30.30", "40.40.40.40"},
			expected:         200,
		},
		{
			desc:             "authorized with only one X-Forwarded-For",
			whiteList:        []string{"30.30.30.30"},
			useXForwardedFor: true,
			xForwardedFor:    []string{"30.30.30.30"},
			expected:         200,
		},
		{
			desc:             "non authorized with X-Forwarded-For",
			whiteList:        []string{"30.30.30.30"},
			useXForwardedFor: true,
			xForwardedFor:    []string{"30.30.30.31", "40.40.40.40"},
			expected:         403,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			whiteLister, err := NewIPWhiteLister(test.whiteList, test.useXForwardedFor)
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "http://10.10.10.10", nil)

			if len(test.remoteAddr) > 0 {
				req.RemoteAddr = test.remoteAddr
			}

			if len(test.xForwardedFor) > 0 {
				for _, xff := range test.xForwardedFor {
					req.Header.Add(whitelist.XForwardedFor, xff)
				}
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			whiteLister.ServeHTTP(recorder, req, next)

			assert.Equal(t, test.expected, recorder.Code)
		})
	}
}
