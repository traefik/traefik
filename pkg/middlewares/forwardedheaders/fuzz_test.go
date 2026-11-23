package forwardedheaders

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func FuzzForwardedHeaders(f *testing.F) {
	f.Add("X-Forwarded-For", "spoofed", "203.0.113.1:1234")
	f.Add("Connection", "X-Secret", "203.0.113.1:1234")
	f.Add("X-Forwarded-Proto", "https", "203.0.113.1:1234")

	f.Fuzz(func(t *testing.T, header, value, remoteAddr string) {
		if header == "" {
			return
		}

		var got string
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got = r.Header.Get(header)
			w.WriteHeader(http.StatusNoContent)
		})

		handler, err := NewXForwarded(false, nil, nil, false, false, next)
		require.NoError(t, err)

		request := func(withHeader bool) *httptest.ResponseRecorder {
			req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
			req.RemoteAddr = remoteAddr
			if withHeader {
				req.Header.Set(header, value)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			return rec
		}

		request(false)
		baseline := got
		request(true)

		if isManagedXHeader(header) {
			require.Equal(t, baseline, got)
		}
	})
}
