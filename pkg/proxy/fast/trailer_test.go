package fast

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/vulcand/oxy/v2/roundrobin"
)

// TestRequestTrailersNotForwardedToBackend ensures that request trailers are not
// forwarded to the backend by the fast reverse proxy.
//
// This is a deliberate behavior: trailers arrive after the body, once routing and
// security decisions have already been made, so forwarding them could raise security
// concerns in Traefik. Discarding them is allowed by
// https://www.rfc-editor.org/rfc/rfc9112#section-7.1.2
//
// The fast proxy builds its outgoing request from the incoming header section only, and
// never reads req.Trailer, so it forwards neither the values nor the declared names.
// Unlike the httputil proxy it does not keep the names as a hint: fasthttp turns a
// Trailer header into an actual trailer field with an empty value, which a backend
// merging trailers into its header namespace would act upon.
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

	backendURL, err := url.Parse(backend.URL)
	require.NoError(t, err)

	transportManager := &transportManagerMock{}

	p, err := NewProxyBuilder(transportManager, static.FastProxyConfig{}).Build("default", backendURL, true, false)
	require.NoError(t, err)

	lb, err := roundrobin.New(p)
	require.NoError(t, err)

	err = lb.UpsertServer(backendURL)
	require.NoError(t, err)

	var entryPointTrailer http.Header
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		entryPointTrailer = req.Trailer.Clone()
		lb.ServeHTTP(rw, req)
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

	// The entry point received the trailer, so the backend not seeing it is the result
	// of the proxy not forwarding it.
	require.Contains(t, entryPointTrailer, "X-Test-Trailer")

	// Assert the key is absent, and not merely valueless: Header.Get returns "" both for
	// an absent key and for a key declared with an empty value, so it cannot tell a
	// discarded trailer from one whose name was forwarded without its value.
	assert.NotContains(t, backendTrailer, "X-Test-Trailer")
}
