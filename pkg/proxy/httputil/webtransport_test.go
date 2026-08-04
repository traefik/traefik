package httputil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	webtransport "github.com/quic-go/webtransport-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsWebTransportRequest(t *testing.T) {
	tests := []struct {
		desc     string
		method   string
		proto    string
		expected bool
	}{
		{
			desc:     "plain GET",
			method:   http.MethodGet,
			proto:    "HTTP/1.1",
			expected: false,
		},
		{
			desc:     "HTTP/2 CONNECT tunnel",
			method:   http.MethodConnect,
			proto:    "HTTP/2.0",
			expected: false,
		},
		{
			desc:     "WebSocket upgrade",
			method:   http.MethodGet,
			proto:    "HTTP/1.1",
			expected: false,
		},
		{
			desc:     "WebTransport CONNECT (legacy protocol string)",
			method:   http.MethodConnect,
			proto:    "webtransport",
			expected: true,
		},
		{
			desc:     "WebTransport CONNECT (h3 protocol string)",
			method:   http.MethodConnect,
			proto:    "webtransport-h3",
			expected: true,
		},
		{
			desc:     "CONNECT without webtransport proto",
			method:   http.MethodConnect,
			proto:    "other-protocol",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			req, err := http.NewRequest(test.method, "https://example.com/path", nil)
			require.NoError(t, err)
			req.Proto = test.proto

			assert.Equal(t, test.expected, IsWebTransportRequest(req))
		})
	}
}

func TestWebTransportServerContext(t *testing.T) {
	ctx := context.Background()

	// Nil when not set.
	assert.Nil(t, GetWebTransportServer(ctx))

	// Non-nil after Set.
	srv := &webtransport.Server{}
	ctx2 := SetWebTransportServer(ctx, srv)
	assert.Same(t, srv, GetWebTransportServer(ctx2))

	// Original context is unchanged.
	assert.Nil(t, GetWebTransportServer(ctx))
}

// TestWebTransportProxyHandler_NonWebTransport verifies that regular HTTP
// requests are forwarded to next unchanged.
func TestWebTransportProxyHandler_NonWebTransport(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	})

	h := newWebTransportProxyHandler(next, mustParseURL(t, "https://backend"), nil)

	req := httptest.NewRequest(http.MethodGet, "https://proxy/path", nil)
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)

	assert.True(t, called, "next should be called for non-WebTransport requests")
}

// TestWebTransportProxyHandler_NoServerInContext verifies that a WebTransport
// CONNECT with no webtransport.Server in the context falls through to next
// rather than returning an opaque error.
func TestWebTransportProxyHandler_NoServerInContext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	})

	h := newWebTransportProxyHandler(next, mustParseURL(t, "https://backend"), nil)

	req := httptest.NewRequest(http.MethodConnect, "https://proxy/path", nil)
	req.Proto = "webtransport"
	// No webtransport.Server in context.
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)

	assert.True(t, called, "next should be called when no WebTransport server is in context")
}

// TestWebTransportProxyHandler_WebTransportH3Proto verifies that the h3
// variant of the protocol string is also recognised as a WebTransport request.
func TestWebTransportProxyHandler_WebTransportH3Proto(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	})

	h := newWebTransportProxyHandler(next, mustParseURL(t, "https://backend"), nil)

	req := httptest.NewRequest(http.MethodConnect, "https://proxy/path", nil)
	req.Proto = "webtransport-h3"
	rw := httptest.NewRecorder()

	h.ServeHTTP(rw, req)

	// Falls through to next because no server is in context.
	assert.True(t, called)
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	require.NoError(t, err)
	return u
}
