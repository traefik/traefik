package replacepathregex

import (
	"context"
	"net/http"
	"testing"

	"github.com/containous/traefik/pkg/config/dynamic"
	"github.com/containous/traefik/pkg/middlewares/replacepath"
	"github.com/containous/traefik/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplacePathRegex(t *testing.T) {
	testCases := []struct {
		desc           string
		path           string
		config         dynamic.ReplacePathRegex
		expectedPath   string
		expectedHeader string
		expectsError   bool
	}{
		{
			desc: "simple regex",
			path: "/whoami/and/whoami",
			config: dynamic.ReplacePathRegex{
				Replacement: "/who-am-i/$1",
				Regex:       `^/whoami/(.*)`,
			},
			expectedPath:   "/who-am-i/and/whoami",
			expectedHeader: "/whoami/and/whoami",
		},
		{
			desc: "simple replace (no regex)",
			path: "/whoami/and/whoami",
			config: dynamic.ReplacePathRegex{
				Replacement: "/who-am-i",
				Regex:       `/whoami`,
			},
			expectedPath:   "/who-am-i/and/who-am-i",
			expectedHeader: "/whoami/and/whoami",
		},
		{
			desc: "no match",
			path: "/whoami/and/whoami",
			config: dynamic.ReplacePathRegex{
				Replacement: "/whoami",
				Regex:       `/no-match`,
			},
			expectedPath: "/whoami/and/whoami",
		},
		{
			desc: "multiple replacement",
			path: "/downloads/src/source.go",
			config: dynamic.ReplacePathRegex{
				Replacement: "/downloads/$1-$2",
				Regex:       `^(?i)/downloads/([^/]+)/([^/]+)$`,
			},
			expectedPath:   "/downloads/src-source.go",
			expectedHeader: "/downloads/src/source.go",
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {

			var actualPath, actualHeader, requestURI string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				actualPath = r.URL.Path
				actualHeader = r.Header.Get(replacepath.ReplacedPathHeader)
				requestURI = r.RequestURI
			})

			handler, err := New(context.Background(), next, test.config, "foo-replace-path-regexp")
			if test.expectsError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+test.path, nil)
				req.RequestURI = test.path

				handler.ServeHTTP(nil, req)

				assert.Equal(t, test.expectedPath, actualPath, "Unexpected path.")
				assert.Equal(t, actualPath, requestURI, "Unexpected request URI.")
				if test.expectedHeader != "" {
					assert.Equal(t, test.expectedHeader, actualHeader, "Unexpected '%s' header.", replacepath.ReplacedPathHeader)
				}
			}
		})
	}
}
