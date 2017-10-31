package middlewares

import (
	"net/http"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestReplacePathRegex(t *testing.T) {
	testCases := []struct {
		desc           string
		path           string
		replacement    string
		regex          string
		expectedPath   string
		expectedHeader string
	}{
		{
			desc:           "simple regex",
			path:           "/whoami/and/whoami",
			replacement:    "/who-am-i/$1",
			regex:          `^/whoami/(.*)`,
			expectedPath:   "/who-am-i/and/whoami",
			expectedHeader: "/whoami/and/whoami",
		},
		{
			desc:           "simple replace (no regex)",
			path:           "/whoami/and/whoami",
			replacement:    "/who-am-i",
			regex:          `/whoami`,
			expectedPath:   "/who-am-i/and/who-am-i",
			expectedHeader: "/whoami/and/whoami",
		},
		{
			desc:           "multiple replacement",
			path:           "/downloads/src/source.go",
			replacement:    "/downloads/$1-$2",
			regex:          `^(?i)/downloads/([^/]+)/([^/]+)$`,
			expectedPath:   "/downloads/src-source.go",
			expectedHeader: "/downloads/src/source.go",
		},
		{
			desc:         "invalid regular expression",
			path:         "/invalid/regexp/test",
			replacement:  "/valid/regexp/$1",
			regex:        `^(?err)/invalid/regexp/([^/]+)$`,
			expectedPath: "/invalid/regexp/test",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var actualPath, actualHeader, requestURI string
			handler := NewReplacePathRegexHandler(
				test.regex,
				test.replacement,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					actualPath = r.URL.Path
					actualHeader = r.Header.Get(ReplacedPathHeader)
					requestURI = r.RequestURI
				}),
			)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+test.path, nil)

			handler.ServeHTTP(nil, req)

			assert.Equal(t, test.expectedPath, actualPath, "Unexpected path.")
			assert.Equal(t, test.expectedHeader, actualHeader, "Unexpected '%s' header.", ReplacedPathHeader)
			if test.expectedHeader != "" {
				assert.Equal(t, actualPath, requestURI, "Unexpected request URI.")
			}
		})
	}
}
