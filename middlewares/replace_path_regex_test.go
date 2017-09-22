package middlewares

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestReplacePathRegex(t *testing.T) {
	path := "/whoami/and/whoami"

	tests := map[string]string{ // key is regexp and value is expected result.
		"^/whoami": "/who-am-i/and/whoami",
		"/whoami":  "/who-am-i/and/who-am-i",
	}

	for key, value := range tests {
		t.Run(key, func(t *testing.T) {

			var expectedPath, actualHeader, requestURI string
			handler := &ReplacePathRegex{
				Regexp: regexp.MustCompile(key),
				Repl:   "/who-am-i",
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					expectedPath = r.URL.Path
					actualHeader = r.Header.Get(ReplacedPathHeader)
					requestURI = r.RequestURI
				}),
			}

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+path, nil)

			handler.ServeHTTP(nil, req)

			assert.Equal(t, expectedPath, value, "Unexpected path.")
			assert.Equal(t, path, actualHeader, "Unexpected '%s' header.", ReplacedPathHeader)
			assert.Equal(t, expectedPath, requestURI, "Unexpected request URI.")
		})
	}
}
