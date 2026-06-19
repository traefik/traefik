package fast

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

// TestRequestTrailersNotForwardedToBackend ensures that request trailers are not
// forwarded to the backend by the fast reverse proxy.
//
// This is a deliberate behavior: trailers arrive after the body, once routing and
// security decisions have already been made, so forwarding them could raise security
// concerns in Traefik. This test locks the current behavior to catch any regression.
func TestRequestTrailersNotForwardedToBackend(t *testing.T) {
	var backendTrailer http.Header

	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		backendTrailer = req.Trailer.Clone()

		rw.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(backend.Close)

	builder := NewProxyBuilder(&transportManagerMock{}, static.FastProxyConfig{})
	proxyHandler, err := builder.Build("", testhelpers.MustParseURL(backend.URL), true, false)
	require.NoError(t, err)

	proxy := httptest.NewServer(proxyHandler)
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

	assert.Empty(t, backendTrailer.Get("X-Test-Trailer"))
}
