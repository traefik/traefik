package fast

import (
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
	"golang.org/x/net/http/httpguts"
)

const http2Settings = "AAMAAABkAAQAoAAAAAIAAAAA"

func TestH2CUpgradeNotForwarded(t *testing.T) {
	testCases := []struct {
		desc            string
		requestHeaders  http.Header
		expectedHeaders http.Header
	}{
		{
			desc: "h2c upgrade with HTTP2-Settings listed in Connection",
			requestHeaders: http.Header{
				"Connection":     {"Upgrade, HTTP2-Settings"},
				"Upgrade":        {"h2c"},
				"Http2-Settings": {http2Settings},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "h2c upgrade with HTTP2-Settings smuggled out of Connection",
			requestHeaders: http.Header{
				"Connection":     {"Upgrade"},
				"Upgrade":        {"h2c"},
				"Http2-Settings": {http2Settings},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "h2c upgrade without HTTP2-Settings",
			requestHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"h2c"},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "h2c upgrade with an uppercase token",
			requestHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"H2C"},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "h2c upgrade combined with a websocket upgrade",
			requestHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"websocket, h2c"},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "HTTP2-Settings without an upgrade",
			requestHeaders: http.Header{
				"Http2-Settings": {http2Settings},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "websocket upgrade is preserved",
			requestHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"websocket"},
			},
			expectedHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"websocket"},
			},
		},
		// Only the first Upgrade value is inspected, as the upstream net/http/httputil fix does, hence the two
		// outcomes below. Both are equivalent in the end: the forwarded Upgrade values are replaced with the first
		// one, so a later h2c value never reaches the backend either.
		{
			desc: "h2c as the first Upgrade value drops every upgrade",
			requestHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"h2c", "websocket"},
			},
			expectedHeaders: http.Header{},
		},
		{
			desc: "h2c as a later Upgrade value keeps the first one",
			requestHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"websocket", "h2c"},
			},
			expectedHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"websocket"},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				forwarded := http.Header{}
				for _, name := range []string{"Connection", "Upgrade", "Http2-Settings"} {
					if values := req.Header.Values(name); len(values) > 0 {
						forwarded[name] = values
					}
				}
				assert.Equal(t, test.expectedHeaders, forwarded)

				// The backend switches to h2c on any Upgrade: h2c, without the validations RFC 7540 section 3.2.1 requires.
				if !httpguts.HeaderValuesContainsToken([]string{req.Header.Get("Upgrade")}, "h2c") {
					return
				}

				rw.Header().Set("Connection", "Upgrade")
				rw.Header().Set("Upgrade", "h2c")
				rw.WriteHeader(http.StatusSwitchingProtocols)
			}))
			t.Cleanup(backend.Close)

			proxy := createProxyWithForwarder(t, backend.URL, createConnectionPool(backend.URL, nil))

			req, err := http.NewRequest(http.MethodGet, proxy.URL+"/public", http.NoBody)
			require.NoError(t, err)

			req.Header = test.requestHeaders

			resp, err := proxy.Client().Do(req)
			require.NoError(t, err)
			t.Cleanup(func() { _ = resp.Body.Close() })

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

// After a 101, the connection is handed over to the tunnel copiers, which read it through the same
// bufio.Reader: the pool readLoop must not read from it anymore.
func TestUpgradedConnNotReadByReadLoop(t *testing.T) {
	// After the handler returns, net/http keeps the connection open and waits for the next request.
	// The proxy never sends one, so the backend stays silent and copyFromBackend parks inside the shared reader.
	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Connection", "Upgrade")
		rw.Header().Set("Upgrade", "test")
		rw.WriteHeader(http.StatusSwitchingProtocols)
	}))
	t.Cleanup(backend.Close)

	var detector atomic.Pointer[concurrentReadDetector]

	u := testhelpers.MustParseURL(backend.URL)
	pool := newConnPool(200, 0, 0, func() (net.Conn, error) {
		c, err := net.Dial("tcp", u.Host)
		require.NoError(t, err)

		d := &concurrentReadDetector{Conn: c}
		detector.Store(d)

		return d, nil
	})
	t.Cleanup(pool.Close)

	proxy := createProxyWithForwarder(t, backend.URL, pool)

	req, err := http.NewRequest(http.MethodGet, proxy.URL+"/ws", http.NoBody)
	require.NoError(t, err)

	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "test")

	resp, err := proxy.Client().Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	// Close the switched connection so copyToBackend returns first and the upgrade handler returns while
	// copyFromBackend is still parked in the shared reader.
	require.NoError(t, resp.Body.Close())

	// Let the readLoop come back around to Peek on the upgraded connection.
	time.Sleep(500 * time.Millisecond)

	d := detector.Load()
	require.NotNil(t, d)
	assert.False(t, d.overlapped.Load())
}

type concurrentReadDetector struct {
	net.Conn

	inRead     atomic.Int32
	overlapped atomic.Bool
}

func (c *concurrentReadDetector) Read(b []byte) (int, error) {
	if c.inRead.Add(1) > 1 {
		c.overlapped.Store(true)
	}
	defer c.inRead.Add(-1)

	return c.Conn.Read(b)
}
