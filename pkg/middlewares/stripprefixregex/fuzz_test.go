package stripprefixregex

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func FuzzStripPrefixRegex(f *testing.F) {
	f.Add(`^/api`, "/api/users")
	f.Add(`^/`, "/")
	f.Add(`^/api`, "/api/../admin")

	f.Fuzz(func(t *testing.T, expression, path string) {
		if !strings.HasPrefix(path, "/") {
			return
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, r.URL.RequestURI(), r.RequestURI)
			w.WriteHeader(http.StatusNoContent)
		})

		handler, err := New(t.Context(), next, dynamic.StripPrefixRegex{Regex: []string{expression}}, "fuzz-strip-prefix-regex")
		if err != nil {
			return
		}

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		req.URL.Path = path
		req.RequestURI = req.URL.RequestURI()
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	})
}
