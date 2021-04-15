package api

import (
	"fmt"
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
	testCases := []struct {
		desc     string
		handler  DashboardHandler
		expected int
	}{
		{
			desc: "OK",
			handler: DashboardHandler{
				Assets: &assetfs.AssetFS{
					Asset: func(path string) ([]byte, error) {
						return []byte{}, nil
					},
					AssetDir: func(path string) ([]string, error) {
						return []string{}, nil
					},
				},
			},
			expected: http.StatusOK,
		},
		{
			desc: "Not found",
			handler: DashboardHandler{
				Assets: &assetfs.AssetFS{
					Asset: func(path string) ([]byte, error) {
						return []byte{}, fmt.Errorf("not found")
					},
					AssetDir: func(path string) ([]string, error) {
						return []string{}, fmt.Errorf("not found")
					},
				},
			},
			expected: http.StatusNotFound,
		},
		{
			desc: "Internal server error",
			handler: DashboardHandler{
				Assets: &assetfs.AssetFS{
					Asset: func(path string) ([]byte, error) {
						return []byte{}, fmt.Errorf("oops")
					},
					AssetDir: func(path string) ([]string, error) {
						return []string{}, fmt.Errorf("oops")
					},
				},
			},
			expected: http.StatusInternalServerError,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/foobar.html", nil)

			rw := httptest.NewRecorder()

			test.handler.ServeHTTP(rw, req)

			assert.Equal(t, test.expected, rw.Code)
			assert.Equal(t, "frame-src 'self' https://traefik.io https://*.traefik.io;", rw.Result().Header.Get("Content-Security-Policy"))
		})
	}
}
