package httputil

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/retry"
	"github.com/vulcand/oxy/v2/roundrobin"
)

// TestRequestTrailerValuesNotForwardedToBackend ensures that the httputil reverse proxy
// forwards the declared request trailer names but never their values.
//
// This is a deliberate behavior: trailer values arrive after the body, once routing and
// security decisions have already been made, so forwarding them could raise security
// concerns in Traefik. Discarding them is allowed by
// https://www.rfc-editor.org/rfc/rfc9112#section-7.1.2 while the names are kept as the
// hint described in https://www.rfc-editor.org/rfc/rfc9110#section-6.6.2
//
// The values only become available to the proxy when a body-buffering middleware runs
// first, hence the buffered case, which is the one that would regress.
func TestRequestTrailerValuesNotForwardedToBackend(t *testing.T) {
	testCases := []struct {
		desc     string
		buffered bool
	}{
		{
			desc: "without a body-buffering middleware",
		},
		{
			desc:     "with a body-buffering middleware",
			buffered: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var backendTrailer http.Header

			backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// Request trailers are only populated after the body has been fully read.
				_, err := io.ReadAll(req.Body)
				require.NoError(t, err)

				backendTrailer = req.Trailer.Clone()

				rw.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(backend.Close)

			transportManager := &transportManagerMock{
				roundTrippers: map[string]http.RoundTripper{"default": &http.Transport{}},
			}

			backendURL, err := url.Parse(backend.URL)
			require.NoError(t, err)

			p, err := NewProxyBuilder(transportManager, nil).Build("default", backendURL, true, false, 0)
			require.NoError(t, err)

			lb, err := roundrobin.New(p)
			require.NoError(t, err)

			err = lb.UpsertServer(backendURL)
			require.NoError(t, err)

			var next http.Handler = lb
			if test.buffered {
				next, err = retry.New(context.Background(), lb,
					dynamic.Retry{Attempts: 2, Status: []string{"500-599"}, RetryNonIdempotentMethod: true},
					retry.Listeners{}, "retry")
				require.NoError(t, err)
			}

			var entryPointTrailer http.Header
			handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				entryPointTrailer = req.Trailer.Clone()
				next.ServeHTTP(rw, req)
			})

			proxy := httptest.NewServer(handler)
			t.Cleanup(proxy.Close)

			pr, pw := io.Pipe()
			req, err := http.NewRequest(http.MethodPost, proxy.URL, pr)
			require.NoError(t, err)

			req.Trailer = http.Header{"X-Test-Trailer": nil}

			go func() {
				_, _ = pw.Write([]byte("body data"))
				req.Trailer.Set("X-Test-Trailer", "trailer-value")
				_ = pw.Close()
			}()

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			t.Cleanup(func() { _ = resp.Body.Close() })

			_, err = io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			// The entry point received the trailer, so what the backend sees is the
			// result of the proxy emptying it.
			require.Contains(t, entryPointTrailer, "X-Test-Trailer")

			// The declared name is forwarded as a hint of the discarded metadata, but
			// carries no value. Assert on the values rather than with Header.Get, which
			// returns "" both for an absent key and for a key with an empty value.
			assert.Contains(t, backendTrailer, "X-Test-Trailer")
			assert.Empty(t, backendTrailer.Values("X-Test-Trailer"))
		})
	}
}
