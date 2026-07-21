package service

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// smuggled is a complete pipelined HTTP/1 request an attacker hides in a CONNECT body,
// hoping a refusing backend parses it as a second request.
const smuggled = "GET /smuggled HTTP/1.1\r\nHost: victim\r\n\r\n"

func TestConnect_HTTP2_TunnelEstablished(t *testing.T) {
	backend := newConnectBackend(t, true)
	addr := serveProxy(t, backend.url)

	transport := http2Transport()

	pipeReader, pipeWriter := io.Pipe()
	t.Cleanup(func() { _ = pipeWriter.Close() })

	req, err := http.NewRequestWithContext(t.Context(), http.MethodConnect, "http://"+addr, pipeReader)
	require.NoError(t, err)
	req.Host = "tunnel.example.com:443"

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

func TestConnect_HTTP2_RefusedTunnelDropsPayload(t *testing.T) {
	backend := newConnectBackend(t, false)
	addr := serveProxy(t, backend.url)

	transport := http2Transport()

	// The attacker pushes the smuggled request immediately, without waiting for the tunnel.
	req, err := http.NewRequestWithContext(t.Context(), http.MethodConnect, "http://"+addr, strings.NewReader(smuggled))
	require.NoError(t, err)
	req.Host = "tunnel.example.com:443"

	res, err := transport.RoundTrip(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = res.Body.Close() })

	assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode)

	time.Sleep(500 * time.Millisecond)
	assert.Empty(t, backend.received())
}

// http2Transport returns a Transport speaking prior-knowledge HTTP/2 over plain TCP.
func http2Transport() *http.Transport {
	protocols := new(http.Protocols)
	protocols.SetUnencryptedHTTP2(true)

	return &http.Transport{Protocols: protocols}
}

// connectBackend is an HTTP/1 backend that either accepts CONNECT tunnels and echoes them back,
// or refuses them. It records every payload byte received after the CONNECT header section.
type connectBackend struct {
	url     *url.URL
	payload *atomic.Pointer[string]
}

func newConnectBackend(t *testing.T, accept bool) *connectBackend {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	backend := &connectBackend{payload: &atomic.Pointer[string]{}}
	empty := ""
	backend.payload.Store(&empty)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go func() {
				defer conn.Close()

				br := bufio.NewReader(conn)
				// Consume the CONNECT header section.
				for {
					line, err := br.ReadString('\n')
					if err != nil {
						return
					}
					if strings.TrimSpace(line) == "" {
						break
					}
				}

				if !accept {
					_, _ = io.WriteString(conn, "HTTP/1.1 405 Method Not Allowed\r\nContent-Length: 0\r\n\r\n")
					// Record anything the proxy pushed despite the refusal.
					var payload strings.Builder
					buf := make([]byte, 1)
					for {
						_ = conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
						n, err := br.Read(buf)
						if n > 0 {
							payload.Write(buf[:n])
						}
						if err != nil {
							break
						}
					}
					got := payload.String()
					backend.payload.Store(&got)

					return
				}

				_, _ = io.WriteString(conn, "HTTP/1.1 200 Connection Established\r\n\r\n")

				// Blind relay: echo every line back uppercased.
				var payload strings.Builder
				for {
					_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
					line, err := br.ReadString('\n')
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
			}()
		}
	}()

	backend.url, err = url.Parse("http://" + listener.Addr().String())
	require.NoError(t, err)

	return backend
}

func (b *connectBackend) received() string {
	return *b.payload.Load()
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
