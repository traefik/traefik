package service

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http/httpguts"
)

func TestH2CUpgradeNotForwarded(t *testing.T) {
	testCases := []struct {
		desc            string
		requestHeaders  string
		expectedHeaders http.Header
		expectedStatus  string
	}{
		{
			desc:            "h2c upgrade with HTTP2-Settings listed in Connection",
			requestHeaders:  "Connection: Upgrade, HTTP2-Settings\r\nUpgrade: h2c\r\nHTTP2-Settings: AAMAAABkAAQAoAAAAAIAAAAA\r\n",
			expectedHeaders: http.Header{},
			expectedStatus:  "HTTP/1.1 200 OK",
		},
		{
			desc:            "h2c upgrade with HTTP2-Settings smuggled out of Connection",
			requestHeaders:  "Connection: Upgrade\r\nUpgrade: h2c\r\nHTTP2-Settings: AAMAAABkAAQAoAAAAAIAAAAA\r\n",
			expectedHeaders: http.Header{},
			expectedStatus:  "HTTP/1.1 200 OK",
		},
		{
			desc:            "h2c upgrade without HTTP2-Settings",
			requestHeaders:  "Connection: Upgrade\r\nUpgrade: h2c\r\n",
			expectedHeaders: http.Header{},
			expectedStatus:  "HTTP/1.1 200 OK",
		},
		{
			desc:            "h2c upgrade with an uppercase token",
			requestHeaders:  "Connection: Upgrade\r\nUpgrade: H2C\r\n",
			expectedHeaders: http.Header{},
			expectedStatus:  "HTTP/1.1 200 OK",
		},
		{
			desc:            "h2c upgrade combined with a websocket upgrade",
			requestHeaders:  "Connection: Upgrade\r\nUpgrade: websocket, h2c\r\n",
			expectedHeaders: http.Header{},
			expectedStatus:  "HTTP/1.1 200 OK",
		},
		{
			desc:            "HTTP2-Settings without an upgrade",
			requestHeaders:  "HTTP2-Settings: AAMAAABkAAQAoAAAAAIAAAAA\r\n",
			expectedHeaders: http.Header{},
			expectedStatus:  "HTTP/1.1 200 OK",
		},
		{
			desc:           "websocket upgrade is preserved",
			requestHeaders: "Connection: Upgrade\r\nUpgrade: websocket\r\n",
			expectedHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"websocket"},
			},
			expectedStatus: "HTTP/1.1 200 OK",
		},
		// Only the first Upgrade value is inspected, as the upstream fix does, hence the two outcomes below.
		// Both are equivalent in the end: the reverse proxy replaces the Upgrade values it forwards with the first
		// one, so a later h2c value never reaches the backend either.
		{
			desc:            "h2c as the first Upgrade value drops every upgrade",
			requestHeaders:  "Connection: Upgrade\r\nUpgrade: h2c\r\nUpgrade: websocket\r\n",
			expectedHeaders: http.Header{},
			expectedStatus:  "HTTP/1.1 200 OK",
		},
		{
			desc:           "h2c as a later Upgrade value keeps the first one",
			requestHeaders: "Connection: Upgrade\r\nUpgrade: websocket\r\nUpgrade: h2c\r\n",
			expectedHeaders: http.Header{
				"Connection": {"Upgrade"},
				"Upgrade":    {"websocket"},
			},
			expectedStatus: "HTTP/1.1 200 OK",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			backendAddr, received := startUpgradeBackend(t)

			proxyHandler, err := buildProxy(new(true), nil, http.DefaultTransport, nil)
			require.NoError(t, err)

			proxy := createProxyWithForwarder(t, proxyHandler, "http://"+backendAddr)
			t.Cleanup(proxy.Close)

			conn, err := net.Dial("tcp", proxy.Listener.Addr().String())
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })

			_, err = fmt.Fprintf(conn, "GET /public HTTP/1.1\r\nHost: backend\r\n%s\r\n", test.requestHeaders)
			require.NoError(t, err)

			var backendHeaders http.Header
			select {
			case backendHeaders = <-received:
			case <-time.After(5 * time.Second):
				t.Fatal("the backend did not receive the request")
			}

			forwarded := http.Header{}
			for _, name := range []string{"Connection", "Upgrade", "HTTP2-Settings"} {
				if values := backendHeaders.Values(name); len(values) > 0 {
					forwarded[http.CanonicalHeaderKey(name)] = values
				}
			}
			assert.Equal(t, test.expectedHeaders, forwarded)

			require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))

			statusLine, err := bufio.NewReader(conn).ReadString('\n')
			require.NoError(t, err)
			assert.Equal(t, test.expectedStatus, strings.TrimSpace(statusLine))
		})
	}
}

// startUpgradeBackend starts a backend switching to h2c on any Upgrade: h2c,
// without performing the validations required by RFC 7540 section 3.2.1.
// It reports the headers of the single request it receives.
func startUpgradeBackend(t *testing.T) (string, <-chan http.Header) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	received := make(chan http.Header, 1)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go func() {
				defer conn.Close()

				req, err := http.ReadRequest(bufio.NewReader(conn))
				if err != nil {
					return
				}

				received <- req.Header

				if httpguts.HeaderValuesContainsToken([]string{req.Header.Get("Upgrade")}, "h2c") {
					_, _ = conn.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: h2c\r\n\r\n"))
					return
				}

				_, _ = conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"))
			}()
		}
	}()

	return listener.Addr().String(), received
}
