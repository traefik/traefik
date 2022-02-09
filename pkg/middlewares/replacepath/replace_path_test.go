package replacepath

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

func TestReplacePath(t *testing.T) {
	testCases := []struct {
		desc            string
		path            string
		config          dynamic.ReplacePath
		expectedPath    string
		expectedRawPath string
		expectedHeader  string
	}{
		{
			desc: "simple path",
			path: "/example",
			config: dynamic.ReplacePath{
				Path: "/replacement-path",
			},
			expectedPath:    "/replacement-path",
			expectedRawPath: "",
			expectedHeader:  "/example",
		},
		{
			desc: "long path",
			path: "/some/really/long/path",
			config: dynamic.ReplacePath{
				Path: "/replacement-path",
			},
			expectedPath:    "/replacement-path",
			expectedRawPath: "",
			expectedHeader:  "/some/really/long/path",
		},
		{
			desc: "path with escaped value",
			path: "/foo%2Fbar",
			config: dynamic.ReplacePath{
				Path: "/replacement-path",
			},
			expectedPath:    "/replacement-path",
			expectedRawPath: "",
			expectedHeader:  "/foo%2Fbar",
		},
		{
			desc: "replacement with escaped value",
			path: "/path",
			config: dynamic.ReplacePath{
				Path: "/foo%2Fbar",
			},
			expectedPath:    "/foo/bar",
			expectedRawPath: "/foo%2Fbar",
			expectedHeader:  "/path",
		},
		{
			desc: "replacement with percent encoded backspace char",
			path: "/path/%08bar",
			config: dynamic.ReplacePath{
				Path: "/path/%08bar",
			},
			expectedPath:    "/path/\bbar",
			expectedRawPath: "/path/%08bar",
			expectedHeader:  "/path/%08bar",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var actualPath, actualRawPath, actualHeader, requestURI string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualPath = r.URL.Path
				actualRawPath = r.URL.RawPath
				actualHeader = r.Header.Get(ReplacedPathHeader)
				requestURI = r.RequestURI
			})

			handler, err := New(context.Background(), next, test.config, "foo-replace-path")
			require.NoError(t, err)

			server := httptest.NewServer(handler)
			defer server.Close()

			resp, err := http.Get(server.URL + test.path)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			assert.Equal(t, test.expectedPath, actualPath, "Unexpected path.")
			assert.Equal(t, test.expectedHeader, actualHeader, "Unexpected '%s' header.", ReplacedPathHeader)

			if actualRawPath == "" {
				assert.Equal(t, actualPath, requestURI, "Unexpected request URI.")
			} else {
				assert.Equal(t, actualRawPath, requestURI, "Unexpected request URI.")
			}
		})
	}
}
