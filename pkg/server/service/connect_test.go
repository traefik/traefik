package service

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnect_HTTP1_Rejected(t *testing.T) {
	var backendCalled bool
	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		backendCalled = true
	}))
	t.Cleanup(backend.Close)

	backendURL, err := url.Parse(backend.URL)
	require.NoError(t, err)

	addr := serveProxy(t, backendURL)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodConnect, "http://"+addr, nil)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = res.Body.Close() })

	assert.Equal(t, http.StatusNotImplemented, res.StatusCode)
	assert.False(t, backendCalled)
}

func TestConnect_HTTP2_TunnelEstablished(t *testing.T) {
	backend := newConnectBackend(t, true)
	addr := serveProxy(t, backend.url)

	pipeReader, pipeWriter := io.Pipe()
	t.Cleanup(func() { _ = pipeWriter.Close() })

	req, err := http.NewRequestWithContext(t.Context(), http.MethodConnect, "http://"+addr, pipeReader)
	require.NoError(t, err)

	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)
	transport := http.Transport{Protocols: protocols}

	res, err := transport.RoundTrip(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = res.Body.Close() })

	require.Equal(t, http.StatusOK, res.StatusCode)

	// Payload sent after the tunnel is established must reach the backend and echo back.
	_, err = io.WriteString(pipeWriter, "ping\n")
	require.NoError(t, err)

	echo, err := bufio.NewReader(res.Body).ReadString('\n')
	require.NoError(t, err)

	assert.Equal(t, "PING", strings.TrimSpace(echo))
}

func TestConnect_HTTP2_RefusedTunnelWithContentLength(t *testing.T) {
	backend := newConnectBackend(t, false)
	addr := serveProxy(t, backend.url)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodConnect, "http://"+addr, strings.NewReader("foo"))
	require.NoError(t, err)

	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)
	transport := http.Transport{Protocols: protocols}

	res, err := transport.RoundTrip(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = res.Body.Close() })

	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
	assert.Equal(t, "foo", *backend.payload.Load())
}

func TestConnect_HTTP2_RefusedTunnelDropsPayloadWithoutContentLength(t *testing.T) {
	backend := newConnectBackend(t, false)
	addr := serveProxy(t, backend.url)

	// Wrapping the reader hides its length, so the request is sent without a Content-Length header.
	req, err := http.NewRequestWithContext(t.Context(), http.MethodConnect, "http://"+addr, io.NopCloser(strings.NewReader("foo")))
	require.NoError(t, err)

	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)
	transport := http.Transport{Protocols: protocols}

	res, err := transport.RoundTrip(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = res.Body.Close() })

	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)
	assert.Empty(t, *backend.payload.Load())
}

// connectBackend is an HTTP/1 backend that either accepts CONNECT tunnels and echoes them back,
// or refuses them. It records every payload byte received after the CONNECT header section.
type connectBackend struct {
	url     *url.URL
	payload *atomic.Pointer[string]
}

func newConnectBackend(t *testing.T, accept bool) *connectBackend {
	t.Helper()

	backend := &connectBackend{payload: &atomic.Pointer[string]{}}
	empty := ""
	backend.payload.Store(&empty)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if !accept {
			if req.ContentLength > 0 {
				// A declared body means the proxy forwarded payload; read it to capture what leaked.
				body, _ := io.ReadAll(req.Body)
				got := string(body)
				backend.payload.Store(&got)

				rw.WriteHeader(http.StatusMethodNotAllowed)

				return
			}

			// No declared body: hijack to make sure the proxy did not push anything on the raw connection.
			conn, brw, err := rw.(http.Hijacker).Hijack()
			if err != nil {
				return
			}
			defer conn.Close()

			var payload strings.Builder
			buf := make([]byte, 1)
			for {
				_ = conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
				n, err := brw.Read(buf)
				if n > 0 {
					payload.Write(buf[:n])
				}
				if err != nil {
					break
				}
			}
			got := payload.String()
			backend.payload.Store(&got)

			_, _ = io.WriteString(conn, "HTTP/1.1 405 Method Not Allowed\r\nContent-Length: 0\r\n\r\n")

			return
		}

		// The tunnel is a raw byte stream, so hijack the connection to bypass the HTTP response machinery.
		// The returned reader already holds any payload buffered alongside the CONNECT header section.
		conn, brw, err := rw.(http.Hijacker).Hijack()
		if err != nil {
			return
		}
		defer conn.Close()

		_, _ = io.WriteString(conn, "HTTP/1.1 200 Connection Established\r\n\r\n")

		// Blind relay: echo every line back uppercased.
		var payload strings.Builder
		for {
			_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
			line, err := brw.ReadString('\n')
			if len(line) > 0 {
				payload.WriteString(line)
				got := payload.String()
				backend.payload.Store(&got)
				_, _ = io.WriteString(conn, strings.ToUpper(line))
			}
			if err != nil {
				return
			}
		}
	}))
	t.Cleanup(server.Close)

	backendURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	backend.url = backendURL

	return backend
}

// serveProxy exposes the Traefik proxy handler over a plain TCP listener, with h2c enabled so that
// both HTTP/1 and prior-knowledge HTTP/2 clients can reach it.
func serveProxy(t *testing.T, target *url.URL) string {
	t.Helper()

	proxy, err := buildProxy(new(true), nil, http.DefaultTransport, newBufferPool())
	require.NoError(t, err)

	// The load balancer sets the backend server on the request URL before the proxy runs.
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		proxy.ServeHTTP(rw, req)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	srv := &http.Server{Handler: handler, Protocols: protocols}
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(func() { _ = srv.Close() })

	return listener.Addr().String()
}
