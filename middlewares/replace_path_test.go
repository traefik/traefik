package middlewares

import (
	"net/http"
	"testing"

	"github.com/containous/traefik/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestReplacePath(t *testing.T) {
	const replacementPath = "/replacement-path"

	paths := []string{
		"/example",
		"/some/really/long/path",
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {

			var expectedPath, actualHeader string
			handler := &ReplacePath{
				Path: replacementPath,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					expectedPath = r.URL.Path
					actualHeader = r.Header.Get(ReplacedPathHeader)
				}),
			}

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost"+path, nil)

			handler.ServeHTTP(nil, req)
			assert.Equal(t, expectedPath, replacementPath, "Unexpected path.")
			assert.Equal(t, path, actualHeader, "Unexpected '%s' header.", ReplacedPathHeader)
		})
	}
}
