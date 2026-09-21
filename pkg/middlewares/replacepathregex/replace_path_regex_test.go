package replacepathregex

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/replacepath"
)

func TestReplacePathRegex(t *testing.T) {
	testCases := []struct {
		desc               string
		path               string
		config             dynamic.ReplacePathRegex
		expectedPath       string
		expectedRawPath    string
		expectedHeader     string
		expectedRequestURI string
		expectedStatusCode int
		expectsError       bool
	}{
		{
			desc: "simple regex",
			path: "/whoami/and/whoami",
			config: dynamic.ReplacePathRegex{
				Replacement: "/who-am-i/$1",
				Regex:       `^/whoami/(.*)`,
			},
			expectedPath:       "/who-am-i/and/whoami",
			expectedHeader:     "/whoami/and/whoami",
			expectedRequestURI: "/who-am-i/and/whoami",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "simple replace (no regex)",
			path: "/whoami/and/whoami",
			config: dynamic.ReplacePathRegex{
				Replacement: "/who-am-i",
				Regex:       `/whoami`,
			},
			expectedPath:       "/who-am-i/and/who-am-i",
			expectedHeader:     "/whoami/and/whoami",
			expectedRequestURI: "/who-am-i/and/who-am-i",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "empty replacement",
			path: "/whoami/and/whoami",
			config: dynamic.ReplacePathRegex{
				Replacement: "",
				Regex:       `/whoami`,
			},
			expectedPath:       "/and",
			expectedHeader:     "/whoami/and/whoami",
			expectedRequestURI: "/and",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "empty trimmed replacement",
			path: "/whoami/and/whoami",
			config: dynamic.ReplacePathRegex{
				Replacement: " ",
				Regex:       `/whoami`,
			},
			expectedPath:       "/and",
			expectedHeader:     "/whoami/and/whoami",
			expectedRequestURI: "/and",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "no match",
			path: "/whoami/and/whoami",
			config: dynamic.ReplacePathRegex{
				Replacement: "/whoami",
				Regex:       `/no-match`,
			},
			expectedPath:       "/whoami/and/whoami",
			expectedRequestURI: "/whoami/and/whoami",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "multiple replacement",
			path: "/downloads/src/source.go",
			config: dynamic.ReplacePathRegex{
				Replacement: "/downloads/$1-$2",
				Regex:       `^(?i)/downloads/([^/]+)/([^/]+)$`,
			},
			expectedPath:       "/downloads/src-source.go",
			expectedHeader:     "/downloads/src/source.go",
			expectedRequestURI: "/downloads/src-source.go",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "invalid regular expression",
			path: "/invalid/regexp/test",
			config: dynamic.ReplacePathRegex{
				Replacement: "/valid/regexp/$1",
				Regex:       `^(?err)/invalid/regexp/([^/]+)$`,
			},
			expectedPath: "/invalid/regexp/test",
			expectsError: true,
		},
		{
			desc: "replacement with escaped char",
			path: "/aaa/bbb",
			config: dynamic.ReplacePathRegex{
				Replacement: "/foo%2Fbar",
				Regex:       `/aaa/bbb`,
			},
			expectedPath:       "/foo/bar",
			expectedRawPath:    "/foo%2Fbar",
			expectedHeader:     "/aaa/bbb",
			expectedRequestURI: "/foo%2Fbar",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "path and regex with escaped char",
			path: "/aaa%2Fbbb",
			config: dynamic.ReplacePathRegex{
				Replacement: "/foo/bar",
				Regex:       `/aaa%2Fbbb`,
			},
			expectedPath:       "/foo/bar",
			expectedHeader:     "/aaa%2Fbbb",
			expectedRequestURI: "/foo/bar",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "path with escaped char (no match)",
			path: "/aaa%2Fbbb",
			config: dynamic.ReplacePathRegex{
				Replacement: "/foo/bar",
				Regex:       `/aaa/bbb`,
			},
			expectedPath:       "/aaa/bbb",
			expectedRawPath:    "/aaa%2Fbbb",
			expectedRequestURI: "/aaa%2Fbbb",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "path with percent encoded backspace char",
			path: "/foo/%08bar",
			config: dynamic.ReplacePathRegex{
				Replacement: "/$1",
				Regex:       `^/foo/(.*)`,
			},
			expectedPath:       "/\bbar",
			expectedRequestURI: "/%08bar",
			expectedStatusCode: http.StatusOK,
		},
		{
			desc: "path with ..",
			path: "/foo../bar",
			config: dynamic.ReplacePathRegex{
				Replacement: "/$1",
				Regex:       `^/foo(.*)`,
			},
			expectedStatusCode: http.StatusBadRequest,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var actualPath, actualRawPath, actualHeader, actualRequestURI string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualPath = r.URL.Path
				actualRawPath = r.URL.RawPath
				actualHeader = r.Header.Get(replacepath.ReplacedPathHeader)
				actualRequestURI = r.RequestURI
			})

			handler, err := New(t.Context(), next, test.config, "foo-replace-path-regexp")
			if test.expectsError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			server := httptest.NewServer(handler)
			defer server.Close()

			resp, err := http.Get(server.URL + test.path)
			require.NoError(t, err, "Unexpected error while making test request")
			require.Equal(t, test.expectedStatusCode, resp.StatusCode)

			assert.Equal(t, test.expectedPath, actualPath, "Unexpected path.")
			assert.Equal(t, test.expectedRawPath, actualRawPath, "Unexpected raw path.")
			assert.Equal(t, test.expectedRequestURI, actualRequestURI, "Unexpected request URI.")

			if test.expectedHeader != "" {
				assert.Equal(t, test.expectedHeader, actualHeader, "Unexpected '%s' header.", replacepath.ReplacedPathHeader)
			}
		})
	}
}
