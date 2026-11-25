package replacepath

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func FuzzReplacePath(f *testing.F) {
	f.Add("/replacement")
	f.Add("/foo%2Fbar")
	f.Add("/path/%08bar")

	f.Fuzz(func(t *testing.T, replacement string) {
		expected, err := url.PathUnescape(replacement)
		if err != nil {
			return
		}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, expected, r.URL.Path)
			require.Equal(t, r.URL.RequestURI(), r.RequestURI)
			w.WriteHeader(http.StatusNoContent)
		})

		handler, err := New(t.Context(), next, dynamic.ReplacePath{Path: replacement}, "fuzz-replace-path")
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "http://example.com/original", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	})
}
