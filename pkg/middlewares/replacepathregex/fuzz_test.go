package replacepathregex

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func FuzzReplacePathRegex(f *testing.F) {
	f.Add(`^/api`, "/v1")
	f.Add(`(.*)`, "/replacement")
	f.Add(`^/api`, "/foo%2Fbar")
	f.Add(`^/api/users$`, "")

	f.Fuzz(func(t *testing.T, expression, replacement string) {
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, r.URL.RequestURI(), r.RequestURI)
			w.WriteHeader(http.StatusNoContent)
		})

		handler, err := New(t.Context(), next, dynamic.ReplacePathRegex{
			Regex:       expression,
			Replacement: replacement,
		}, "fuzz-replace-path-regex")
		if err != nil {
			return
		}

		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	})
}
