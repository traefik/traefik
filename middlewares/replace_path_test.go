package middlewares

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

			req, err := http.NewRequest(http.MethodGet, "http://localhost"+path, nil)
			require.NoError(t, err, "%s: unexpected error.", path)

			handler.ServeHTTP(nil, req)
			assert.Equal(t, expectedPath, replacementPath, "Unexpected path.")
			assert.Equal(t, path, actualHeader, "Unexpected '%s' header.", ReplacedPathHeader)
		})
	}
}
