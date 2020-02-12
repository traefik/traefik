package ipwhitelist

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type WhiteListBuilder struct { }
func (b WhiteListBuilder) GetConfigs() map[string]*runtime.MiddlewareInfo {
	configs := make(map[string]*runtime.MiddlewareInfo)

	whitelist := dynamic.Middleware{IPWhiteList: &dynamic.IPWhiteList{SourceRange: []string{"20.20.20.20"}}}
	configs["other-whitelist"] = &runtime.MiddlewareInfo{Middleware: &whitelist}
	configs["not-whitelist"] = &runtime.MiddlewareInfo{Middleware: &dynamic.Middleware{}}

	return configs
}

func TestNewIPWhiteLister(t *testing.T) {

	testCases := []struct {
		desc          string
		whiteList     dynamic.IPWhiteList
		expectedError bool
	}{
		{
			desc: "invalid IP",
			whiteList: dynamic.IPWhiteList{
				SourceRange: []string{"foo"},
			},
			expectedError: true,
		},
		{
			desc: "non-existent append whitelist",
			whiteList: dynamic.IPWhiteList{
				AppendWhiteLists: []string{"bad-whitelist"},
			},
			expectedError: true,
		},
		{
			desc: "invalid append whitelist",
			whiteList: dynamic.IPWhiteList{
				AppendWhiteLists: []string{"not-whitelist"},
			},
			expectedError: true,
		},
		{
			desc: "valid IP",
			whiteList: dynamic.IPWhiteList{
				SourceRange: []string{"10.10.10.10"},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			whiteLister, err := New(context.Background(), next, test.whiteList, WhiteListBuilder{}, "traefikTest")

			if test.expectedError {
				assert.Error(t, err)
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
		whiteList  dynamic.IPWhiteList
		remoteAddr string
		expected   int
	}{
		{
			desc: "authorized with remote address",
			whiteList: dynamic.IPWhiteList{
				SourceRange: []string{"20.20.20.20"},
			},
			remoteAddr: "20.20.20.20:1234",
			expected:   200,
		},
		{
			desc: "authorized with append whitelist",
			whiteList: dynamic.IPWhiteList{
				AppendWhiteLists: []string{"other-whitelist"},
			},
			remoteAddr: "20.20.20.20:1234",
			expected:   200,
		},
		{
			desc: "non authorized with remote address",
			whiteList: dynamic.IPWhiteList{
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
			whiteLister, err := New(context.Background(), next, test.whiteList, WhiteListBuilder{}, "traefikTest")
			require.NoError(t, err)

			recorder := httptest.NewRecorder()

			req := httptest.NewRequest(http.MethodGet, "http://10.10.10.10", nil)

			if len(test.remoteAddr) > 0 {
				req.RemoteAddr = test.remoteAddr
			}

			whiteLister.ServeHTTP(recorder, req)

			assert.Equal(t, test.expected, recorder.Code)
		})
	}
}
