package ipallowlist

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestNewIPAllowLister(t *testing.T) {
	testCases := []struct {
		desc          string
		allowList     dynamic.IPAllowList
		expectedError bool
	}{
		{
			desc: "invalid IP",
			allowList: dynamic.IPAllowList{
				SourceRange: []string{"foo"},
			},
			expectedError: true,
		},
		{
			desc: "valid IP",
			allowList: dynamic.IPAllowList{
				SourceRange: []string{"10.10.10.10"},
			},
		},
		{
			desc: "invalid HTTP status code",
			allowList: dynamic.IPAllowList{
				SourceRange:      []string{"10.10.10.10"},
				RejectStatusCode: 600,
			},
			expectedError: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			allowLister, err := New(t.Context(), next, test.allowList, "traefikTest")

			if test.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, allowLister)
			}
		})
	}
}

func TestIPAllowLister_ServeHTTP(t *testing.T) {
	testCases := []struct {
		desc       string
		allowList  dynamic.IPAllowList
		remoteAddr string
		expected   int
	}{
		{
			desc: "authorized with remote address",
			allowList: dynamic.IPAllowList{
				SourceRange: []string{"20.20.20.20"},
			},
			remoteAddr: "20.20.20.20:1234",
			expected:   200,
		},
		{
			desc: "non authorized with remote address",
			allowList: dynamic.IPAllowList{
				SourceRange: []string{"20.20.20.20"},
			},
			remoteAddr: "20.20.20.21:1234",
			expected:   403,
		},
		{
			desc: "authorized with remote address, reject 404",
			allowList: dynamic.IPAllowList{
				SourceRange:      []string{"20.20.20.20"},
				RejectStatusCode: 404,
			},
			remoteAddr: "20.20.20.20:1234",
			expected:   200,
		},
		{
			desc: "non authorized with remote address, reject 404",
			allowList: dynamic.IPAllowList{
				SourceRange:      []string{"20.20.20.20"},
				RejectStatusCode: 404,
			},
			remoteAddr: "20.20.20.21:1234",
			expected:   404,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			allowLister, err := New(t.Context(), next, test.allowList, "traefikTest")
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "http://10.10.10.10", nil)

			if len(test.remoteAddr) > 0 {
				req.RemoteAddr = test.remoteAddr
			}

			allowLister.ServeHTTP(recorder, req)

			assert.Equal(t, test.expected, recorder.Code)
		})
	}
}
