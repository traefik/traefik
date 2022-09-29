package ipblacklist

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

func TestNewIPBlackLister(t *testing.T) {
	testCases := []struct {
		desc          string
		blackList     dynamic.IPBlackList
		expectedError bool
	}{
		{
			desc: "invalid IP",
			blackList: dynamic.IPBlackList{
				SourceRange: []string{"foo"},
			},
			expectedError: true,
		},
		{
			desc: "valid IP",
			blackList: dynamic.IPBlackList{
				SourceRange: []string{"10.10.10.10"},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			blackLister, err := New(context.Background(), next, test.blackList, "traefikTest")

			if test.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, blackLister)
			}
		})
	}
}

func TestIPBlackLister_ServeHTTP(t *testing.T) {
	testCases := []struct {
		desc       string
		blackList  dynamic.IPBlackList
		remoteAddr string
		expected   int
	}{
		{
			desc: "non authorized with remote address",
			blackList: dynamic.IPBlackList{
				SourceRange: []string{"20.20.20.20"},
			},
			remoteAddr: "20.20.20.20:1234",
			expected:   200,
		},
		{
			desc: "authorized with remote address",
			blackList: dynamic.IPBlackList{
				SourceRange: []string{"20.20.20.20"},
			},
			remoteAddr: "20.20.20.21:1234",
			expected:   403,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			blackLister, err := New(context.Background(), next, test.blackList, "traefikTest")
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "http://10.10.10.10", nil)

			if len(test.remoteAddr) > 0 {
				req.RemoteAddr = test.remoteAddr
			}

			blackLister.ServeHTTP(recorder, req)

			assert.Equal(t, test.expected, recorder.Code)
		})
	}
}
