package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_safePrefix(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected string
	}{
		{
			desc:     "host",
			value:    "https://example.com",
			expected: "",
		},
		{
			desc:     "host with path",
			value:    "https://example.com/foo/bar?test",
			expected: "",
		},
		{
			desc:     "path",
			value:    "/foo/bar",
			expected: "/foo/bar",
		},
		{
			desc:     "path without leading slash",
			value:    "foo/bar",
			expected: "foo/bar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
			require.NoError(t, err)

			req.Header.Set("X-Forwarded-Prefix", test.value)

			prefix := safePrefix(req)

			assert.Equal(t, test.expected, prefix)
		})
	}
}

func Test_ContentSecurityPolicy(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/toto.html", nil)
	rw := httptest.NewRecorder()

	DashboardHandler{
		Assets: &assetfs.AssetFS{
			Asset: func(path string) ([]byte, error) {
				return []byte{}, nil
			},
		},
	}.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, []string{"frame-src 'self' https://traefik.io https://*.traefik.io;"}, rw.Result().Header["Content-Security-Policy"])
}
