package middlewares

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestReplacePathRegex(t *testing.T) {
	testCases := []struct {
		desc         string
		path         string
		replacement  string
		regex        string
		expectedPath string
	}{
		{
			desc:         "^/whoami $> /who-am-i",
			path:         "/whoami/and/whoami",
			replacement:  "/who-am-i",
			regex:        `^/whoami`,
			expectedPath: "/who-am-i/and/whoami",
		},
		{
			desc:         "/whoami $> /who-am-i",
			path:         "/whoami/and/whoami",
			replacement:  "/who-am-i",
			regex:        `/whoami`,
			expectedPath: "/who-am-i/and/who-am-i",
		},
		{
			desc:         "{scope} $> internal",
			path:         "/resources/{scope}",
			replacement:  "internal",
			regex:        `{scope}`,
			expectedPath: "/resources/internal",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var expectedPath, actualHeader, requestURI string
			handler := &ReplacePathRegex{
				Regexp: regexp.MustCompile(test.regex),
				Repl:   test.replacement,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					expectedPath = r.URL.Path
					actualHeader = r.Header.Get(ReplacedPathHeader)
					requestURI = r.RequestURI
				}),
			}

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+test.path, nil)

			handler.ServeHTTP(nil, req)

			assert.Equal(t, expectedPath, test.expectedPath, "Unexpected path.")
			assert.Equal(t, test.path, actualHeader, "Unexpected '%s' header.", ReplacedPathHeader)
			assert.Equal(t, expectedPath, requestURI, "Unexpected request URI.")
		})
	}
}
