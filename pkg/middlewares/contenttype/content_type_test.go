package contenttype

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func TestAutoDetection(t *testing.T) {
	testCases := []struct {
		desc            string
		autoDetect      bool
		contentType     string
		wantContentType string
	}{
		{
			desc:            "Keep the Content-Type returned by the server",
			autoDetect:      false,
			contentType:     "application/json",
			wantContentType: "application/json",
		},
		{
			desc:            "Don't auto-detect Content-Type header by default when not set by the server",
			autoDetect:      false,
			contentType:     "",
			wantContentType: "",
		},
		{
			desc:            "Keep the Content-Type returned by the server with auto-detection middleware",
			autoDetect:      true,
			contentType:     "application/json",
			wantContentType: "application/json",
		},
		{
			desc:            "Auto-detect when Content-Type header is not already set by the server with auto-detection middleware",
			autoDetect:      true,
			contentType:     "",
			wantContentType: "text/plain; charset=utf-8",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var next http.Handler
			next = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test.contentType != "" {
					w.Header().Set("Content-Type", test.contentType)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("Test"))
			})

			if test.autoDetect {
				var err error
				next, err = New(t.Context(), next, dynamic.ContentType{}, "foo-content-type")
				require.NoError(t, err)
			}

			server := httptest.NewServer(
				DisableAutoDetection(next),
			)
			t.Cleanup(server.Close)

			req := testhelpers.MustNewRequest(http.MethodGet, server.URL, nil)
			res, err := server.Client().Do(req)
			require.NoError(t, err)

			assert.Equal(t, test.wantContentType, res.Header.Get("Content-Type"))
		})
	}
}
