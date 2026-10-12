package httputil

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
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
		// Only the first Upgrade value is inspected, as the upstream fix does, hence the two outcomes below.
		// Both are equivalent in the end: the reverse proxy replaces the Upgrade values it forwards with the first
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

			proxy := createProxyWithForwarder(t, backend.URL, http.DefaultTransport)

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

func TestUpgradedTLSConnClosedOnBackendClose(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		conn, brw, err := http.NewResponseController(rw).Hijack()
		if !assert.NoError(t, err) {
			return
		}
		defer conn.Close()

		_, err = brw.WriteString("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\nhello")
		assert.NoError(t, err)
		assert.NoError(t, brw.Flush())
	}))
	t.Cleanup(backend.Close)

	proxy := newUpgradeProxyServer(t, backend.URL, true)

	rawConn, err := net.Dial("tcp", proxy.Listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = rawConn.Close() })

	conn := tls.Client(rawConn, &tls.Config{InsecureSkipVerify: true})
	require.NoError(t, conn.SetDeadline(time.Now().Add(2*time.Second)))

	br := sendUpgradeRequest(t, conn)

	data, err := io.ReadAll(br)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))

	// The TLS layer reports io.EOF as soon as it receives a close_notify alert,
	// so the underlying TCP connection is read to check that it has been closed too.
	_, err = rawConn.Read(make([]byte, 1))
	assert.ErrorIs(t, err, io.EOF)
}

func TestUpgradedTCPConnHalfClosedOnBackendCloseWrite(t *testing.T) {
	received := make(chan string, 1)

	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		conn, brw, err := http.NewResponseController(rw).Hijack()
		if !assert.NoError(t, err) {
			return
		}
		defer conn.Close()

		_, err = brw.WriteString("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\nhello")
		assert.NoError(t, err)
		assert.NoError(t, brw.Flush())

		tcpConn, ok := conn.(*net.TCPConn)
		if !assert.True(t, ok) {
			return
		}
		assert.NoError(t, tcpConn.CloseWrite())

		data, err := io.ReadAll(brw)
		assert.NoError(t, err)

		received <- string(data)
	}))
	t.Cleanup(backend.Close)

	proxy := newUpgradeProxyServer(t, backend.URL, false)

	conn, err := net.Dial("tcp", proxy.Listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	require.NoError(t, conn.SetDeadline(time.Now().Add(2*time.Second)))

	br := sendUpgradeRequest(t, conn)

	data, err := io.ReadAll(br)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(data))

	// The connection to the client is only half-closed, so it can still send data to the backend.
	_, err = conn.Write([]byte("bye"))
	require.NoError(t, err)

	tcpConn, ok := conn.(*net.TCPConn)
	require.True(t, ok)
	require.NoError(t, tcpConn.CloseWrite())

	select {
	case data := <-received:
		assert.Equal(t, "bye", data)
	case <-time.After(2 * time.Second):
		t.Fatal("backend did not receive the client data")
	}
}

func newUpgradeProxyServer(t *testing.T, backendURL string, withTLS bool) *httptest.Server {
	t.Helper()

	transportManager := &transportManagerMock{
		roundTrippers: map[string]http.RoundTripper{"fwd": http.DefaultTransport},
	}

	p, err := NewProxyBuilder(transportManager, nil).Build("fwd", testhelpers.MustParseURL(backendURL), true, false, 0)
	require.NoError(t, err)

	srv := httptest.NewUnstartedServer(p)
	if withTLS {
		srv.StartTLS()
	} else {
		srv.Start()
	}
	t.Cleanup(srv.Close)

	return srv
}

func sendUpgradeRequest(t *testing.T, conn net.Conn) *bufio.Reader {
	t.Helper()

	_, err := conn.Write([]byte("GET /ws HTTP/1.1\r\nHost: localhost\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n"))
	require.NoError(t, err)

	br := bufio.NewReader(conn)
	res, err := http.ReadResponse(br, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, res.StatusCode)

	return br
}
