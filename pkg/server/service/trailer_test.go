package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vulcand/oxy/v2/roundrobin"
)

// TestRequestTrailersNotForwardedToBackend ensures that request trailers are not
// forwarded to the backend by the httputil reverse proxy.
//
// This is a deliberate behavior: trailers arrive after the body, once routing and
// security decisions have already been made, so forwarding them could raise security
// concerns in Traefik. This test locks the current behavior to catch any regression.
func TestRequestTrailersNotForwardedToBackend(t *testing.T) {
	var backendTrailer http.Header

	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Request trailers are only populated after the body has been fully read.
		_, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		backendTrailer = req.Trailer.Clone()

		rw.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(backend.Close)

	proxyHandler, err := buildProxy(pointer(true), nil, http.DefaultTransport, nil)
	require.NoError(t, err)

	lb, err := roundrobin.New(proxyHandler)
	require.NoError(t, err)

	backendURL, err := url.Parse(backend.URL)
	require.NoError(t, err)

	err = lb.UpsertServer(backendURL)
	require.NoError(t, err)

	proxy := httptest.NewServer(lb)
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
