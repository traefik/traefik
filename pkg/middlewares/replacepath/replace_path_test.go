package replacepath

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestReplacePath(t *testing.T) {
	testCases := []struct {
		desc                  string
		path                  string
		config                dynamic.ReplacePath
		expectedPath          string
		expectedRawPath       string
		expectedHeader        string
		expectedServiceHeader string
	}{
		{
			desc: "simple path",
			path: "/example",
			config: dynamic.ReplacePath{
				Path: "/replacement-path",
			},
			expectedPath:          "/replacement-path",
			expectedRawPath:       "",
			expectedHeader:        "/example",
			expectedServiceHeader: "/replacement-path",
		},
		{
			desc: "long path",
			path: "/some/really/long/path",
			config: dynamic.ReplacePath{
				Path: "/replacement-path",
			},
			expectedPath:          "/replacement-path",
			expectedRawPath:       "",
			expectedHeader:        "/some/really/long/path",
			expectedServiceHeader: "/replacement-path",
		},
		{
			desc: "path with escaped value",
			path: "/foo%2Fbar",
			config: dynamic.ReplacePath{
				Path: "/replacement-path",
			},
			expectedPath:          "/replacement-path",
			expectedRawPath:       "",
			expectedHeader:        "/foo%2Fbar",
			expectedServiceHeader: "/replacement-path",
		},
		{
			desc: "replacement with escaped value",
			path: "/path",
			config: dynamic.ReplacePath{
				Path: "/foo%2Fbar",
			},
			expectedPath:          "/foo/bar",
			expectedRawPath:       "/foo%2Fbar",
			expectedHeader:        "/path",
			expectedServiceHeader: "/foo/bar",
		},
		{
			desc: "replacement with percent encoded backspace char",
			path: "/path/%08bar",
			config: dynamic.ReplacePath{
				Path: "/path/%08bar",
			},
			expectedPath:          "/path/\bbar",
			expectedRawPath:       "/path/%08bar",
			expectedHeader:        "/path/%08bar",
			expectedServiceHeader: "/path/\bbar",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var actualPath, actualRawPath, actualHeader, actualServiceHeader, requestURI string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualPath = r.URL.Path
				actualRawPath = r.URL.RawPath
				actualHeader = r.Header.Get(ReplacedPathHeader)
				actualServiceHeader = r.Header.Get(ServicePathHeader)
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
			assert.Equal(t, test.expectedServiceHeader, actualServiceHeader, "Unexpected '%s' service header.", ServicePathHeader)

			if actualRawPath == "" {
				assert.Equal(t, actualPath, requestURI, "Unexpected request URI.")
			} else {
				assert.Equal(t, actualRawPath, requestURI, "Unexpected request URI.")
			}
		})
	}
}
