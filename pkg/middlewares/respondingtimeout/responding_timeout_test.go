package respondingtimeout_test

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	stdhttputil "net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/compress"
	"github.com/traefik/traefik/v3/pkg/middlewares/respondingtimeout"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
)

func wrap(t *testing.T, timeout time.Duration, next http.Handler) http.Handler {
	t.Helper()

	handler, err := respondingtimeout.WrapHandler(timeout)(next)
	require.NoError(t, err)

	return handler
}

// reverseProxy builds the same proxy Traefik puts below a router, so that the tests exercise the real
// error-to-status mapping instead of simulating it.
func reverseProxy(target *url.URL) http.Handler {
	return &stdhttputil.ReverseProxy{
		Rewrite:      func(pr *stdhttputil.ProxyRequest) { pr.SetURL(target) },
		ErrorHandler: httputil.ErrorHandler,
	}
}

// hungBackend accepts connections and never answers, holding them open until the test ends.
func hungBackend(t *testing.T) *url.URL {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	var mu sync.Mutex
	var conns []net.Conn

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}

			mu.Lock()
			conns = append(conns, conn)
			mu.Unlock()
		}
	}()

	t.Cleanup(func() {
		mu.Lock()
		defer mu.Unlock()

		for _, conn := range conns {
			_ = conn.Close()
		}
	})

	target, err := url.Parse("http://" + ln.Addr().String())
	require.NoError(t, err)

	return target
}

func healthyBackend(t *testing.T) *url.URL {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = io.Copy(io.Discard, req.Body)
		_, _ = rw.Write([]byte("backend"))
	}))
	t.Cleanup(ts.Close)

	target, err := url.Parse(ts.URL)
	require.NoError(t, err)

	return target
}

// upgradeBackend answers the handshake and then echoes back whatever the tunnel carries.
func upgradeBackend(t *testing.T) *url.URL {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		conn, brw, err := http.NewResponseController(rw).Hijack()
		if !assert.NoError(t, err) {
			return
		}
		defer conn.Close()

		_, err = brw.WriteString("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: spdy/3.1\r\n\r\n")
		if !assert.NoError(t, err) {
			return
		}
		if !assert.NoError(t, brw.Flush()) {
			return
		}

		_, _ = io.Copy(conn, brw)
	}))
	t.Cleanup(ts.Close)

	target, err := url.Parse(ts.URL)
	require.NoError(t, err)

	return target
}

func TestStatusNormalization(t *testing.T) {
	testCases := []struct {
		desc       string
		status     int
		waitExpiry bool
		expected   int
		upgradeReq bool
	}{
		{
			desc:       "5xx after expiry is normalized to 504",
			status:     http.StatusInternalServerError,
			waitExpiry: true,
			expected:   http.StatusGatewayTimeout,
		},
		{
			desc:       "502 after expiry is normalized to 504",
			status:     http.StatusBadGateway,
			waitExpiry: true,
			expected:   http.StatusGatewayTimeout,
		},
		{
			desc:       "499 after expiry is normalized to 504",
			status:     httputil.StatusClientClosedRequest,
			waitExpiry: true,
			expected:   http.StatusGatewayTimeout,
		},
		{
			desc:     "5xx before expiry is preserved",
			status:   http.StatusInternalServerError,
			expected: http.StatusInternalServerError,
		},
		{
			desc:     "499 before expiry is preserved",
			status:   httputil.StatusClientClosedRequest,
			expected: httputil.StatusClientClosedRequest,
		},
		{
			desc:       "4xx after expiry is preserved",
			status:     http.StatusNotFound,
			waitExpiry: true,
			expected:   http.StatusNotFound,
		},
		{
			desc:       "2xx after expiry is preserved",
			status:     http.StatusOK,
			waitExpiry: true,
			expected:   http.StatusOK,
		},
		{
			desc:       "499 after upgrade handshake expiry is normalized to 504",
			status:     httputil.StatusClientClosedRequest,
			waitExpiry: true,
			expected:   http.StatusGatewayTimeout,
			upgradeReq: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			timeout := time.Hour
			if test.waitExpiry {
				timeout = 20 * time.Millisecond
			}

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if test.waitExpiry {
					select {
					case <-req.Context().Done():
					case <-time.After(5 * time.Second):
						t.Error("deadline never fired")
					}
				}
				rw.WriteHeader(test.status)
			})

			req := httptest.NewRequest(http.MethodGet, "http://localhost/", http.NoBody)
			if test.upgradeReq {
				req.Header.Set("Connection", "Upgrade")
				req.Header.Set("Upgrade", "spdy/3.1")
			}

			rec := httptest.NewRecorder()
			wrap(t, timeout, next).ServeHTTP(rec, req)

			assert.Equal(t, test.expected, rec.Code)
		})
	}
}

func TestNestedDeadlineClamp(t *testing.T) {
	t.Parallel()

	var deadline time.Time
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var ok bool
		deadline, ok = req.Context().Deadline()
		assert.True(t, ok)

		rw.WriteHeader(http.StatusOK)
	})

	// A child router cannot extend the budget set by its parent.
	handler := wrap(t, 30*time.Millisecond, wrap(t, time.Hour, next))

	start := time.Now()
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://localhost/", http.NoBody))

	assert.WithinDuration(t, start.Add(30*time.Millisecond), deadline, 10*time.Millisecond)
}

// TestUpgradeTunnelSurvivesDeadline relays the tunnel through the real proxy, the only way to cover both legs:
// the stdlib proxy closes the backend connection as soon as the request context is done
// (net/http/httputil handleUpgradeResponse), so an expiring deadline would tear a healthy tunnel down.
func TestUpgradeTunnelSurvivesDeadline(t *testing.T) {
	t.Parallel()

	timeout := 100 * time.Millisecond

	ts := httptest.NewServer(wrap(t, timeout, reverseProxy(upgradeBackend(t))))
	t.Cleanup(ts.Close)

	conn, err := net.Dial("tcp", ts.Listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	require.NoError(t, conn.SetDeadline(time.Now().Add(5*time.Second)))

	_, err = fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: %s\r\nConnection: Upgrade\r\nUpgrade: spdy/3.1\r\n\r\n", ts.Listener.Addr())
	require.NoError(t, err)

	br := bufio.NewReader(conn)
	res, err := http.ReadResponse(br, nil)
	require.NoError(t, err)
	require.Equal(t, http.StatusSwitchingProtocols, res.StatusCode)

	// The handshake succeeded, so the deadline is disarmed: well past it, the tunnel must still relay both ways.
	time.Sleep(3 * timeout)

	_, err = conn.Write([]byte("after-deadline"))
	require.NoError(t, err)

	echoed := make([]byte, len("after-deadline"))
	_, err = io.ReadFull(br, echoed)
	require.NoError(t, err)

	assert.Equal(t, "after-deadline", string(echoed))
}

func TestHungUpgradeHandshakeGets504(t *testing.T) {
	t.Parallel()

	// The deadline bounds the handshake: a backend that never switches protocol must not leave the
	// client hanging, and pre-101 the gateway has not responded yet, so a timeout can still be reported.
	ts := httptest.NewServer(wrap(t, 50*time.Millisecond, reverseProxy(hungBackend(t))))
	t.Cleanup(ts.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL, http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "spdy/3.1")

	res, err := ts.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = res.Body.Close() })

	assert.Equal(t, http.StatusGatewayTimeout, res.StatusCode)
}

func TestSlowBackendGets504(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(wrap(t, 50*time.Millisecond, reverseProxy(hungBackend(t))))
	t.Cleanup(ts.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL, http.NoBody)
	require.NoError(t, err)

	res, err := ts.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = res.Body.Close() })

	assert.Equal(t, http.StatusGatewayTimeout, res.StatusCode)
}

// TestExpiryDoesNotCancelNextKeepAliveRequest guards the read deadline arming: net/http cancels the whole
// connection context when a background read fails, so a request that reaches its deadline must not leave the
// next request on the same keep-alive connection with an already-canceled context.
func TestExpiryDoesNotCancelNextKeepAliveRequest(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.Handle("/timeout", wrap(t, 50*time.Millisecond, reverseProxy(hungBackend(t))))
	// No timeout on this router, and a backend that answers.
	mux.Handle("/plain", reverseProxy(healthyBackend(t)))

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	conn, err := net.Dial("tcp", ts.Listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	require.NoError(t, conn.SetDeadline(time.Now().Add(10*time.Second)))
	br := bufio.NewReader(conn)

	_, err = fmt.Fprintf(conn, "GET /timeout HTTP/1.1\r\nHost: %s\r\n\r\n", ts.Listener.Addr())
	require.NoError(t, err)

	res, err := http.ReadResponse(br, nil)
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, res.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusGatewayTimeout, res.StatusCode)

	_, err = fmt.Fprintf(conn, "GET /plain HTTP/1.1\r\nHost: %s\r\n\r\n", ts.Listener.Addr())
	require.NoError(t, err)

	res, err = http.ReadResponse(br, nil)
	require.NoError(t, err)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "backend", string(body))
}

// TestWriteDeadlineNotLeakedToNextKeepAliveRequest guards the deferred clear: net/http only re-arms the write
// deadline when Server.WriteTimeout > 0, which Traefik entrypoints leave at 0, so a write deadline left behind
// by a request would still be live — and already expired — when the next request on the connection responds.
func TestWriteDeadlineNotLeakedToNextKeepAliveRequest(t *testing.T) {
	t.Parallel()

	timeout := 100 * time.Millisecond

	mux := http.NewServeMux()
	// This router has a timeout, but the response comes well within it.
	mux.Handle("/fast", wrap(t, timeout, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write([]byte("fast"))
	})))
	mux.HandleFunc("/plain", func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write([]byte("plain"))
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)

	conn, err := net.Dial("tcp", ts.Listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	br := bufio.NewReader(conn)

	_, err = fmt.Fprintf(conn, "GET /fast HTTP/1.1\r\nHost: %s\r\n\r\n", ts.Listener.Addr())
	require.NoError(t, err)

	res, err := http.ReadResponse(br, nil)
	require.NoError(t, err)
	_, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	// Wait past the first request's deadline: a leaked write deadline is now expired.
	time.Sleep(2 * timeout)

	_, err = fmt.Fprintf(conn, "GET /plain HTTP/1.1\r\nHost: %s\r\n\r\n", ts.Listener.Addr())
	require.NoError(t, err)

	res, err = http.ReadResponse(br, nil)
	require.NoError(t, err)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "plain", string(body))
}

func TestSlowClientBounded(t *testing.T) {
	t.Parallel()

	timeout := 100 * time.Millisecond

	// The backend is healthy: only the read deadline can bound a client that stalls mid-upload, and only a
	// deadline comparison can normalize the outcome, since the connection deadline may cancel the request
	// context before the deadline context does.
	ts := httptest.NewServer(wrap(t, timeout, reverseProxy(healthyBackend(t))))
	t.Cleanup(ts.Close)

	conn, err := net.Dial("tcp", ts.Listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	// Announce a large body and stall after a few bytes: only the read deadline can bound this.
	_, err = fmt.Fprintf(conn, "POST / HTTP/1.1\r\nHost: %s\r\nContent-Length: 1000\r\n\r\npartial", ts.Listener.Addr())
	require.NoError(t, err)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))

	start := time.Now()
	res, err := http.ReadResponse(bufio.NewReader(conn), nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusGatewayTimeout, res.StatusCode)
	assert.Less(t, time.Since(start), 10*timeout)
}

// TestEarlyResponseDoesNotWaitForStalledUpload guards the read deadline being left armed on return: the
// server drains the unread request body before flushing the response (net/http chunkWriter.writeHeader),
// after the handler and its defers have run, so only a deadline still armed on the connection bounds it.
func TestEarlyResponseDoesNotWaitForStalledUpload(t *testing.T) {
	t.Parallel()

	timeout := 100 * time.Millisecond

	// The handler answers without reading the body, as an authentication or a rate limit middleware would.
	ts := httptest.NewServer(wrap(t, timeout, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusUnauthorized)
	})))
	t.Cleanup(ts.Close)

	conn, err := net.Dial("tcp", ts.Listener.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	_, err = fmt.Fprintf(conn, "POST / HTTP/1.1\r\nHost: %s\r\nContent-Length: 1000\r\n\r\npartial", ts.Listener.Addr())
	require.NoError(t, err)

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))

	start := time.Now()
	res, err := http.ReadResponse(bufio.NewReader(conn), nil)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	assert.Less(t, time.Since(start), 10*timeout)
}

func TestFlushPropagatesThroughCompress(t *testing.T) {
	t.Parallel()

	firstChunk := strings.Repeat("a", 2048)

	release := make(chan struct{})
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, err := rw.Write([]byte(firstChunk))
		require.NoError(t, err)
		require.NoError(t, http.NewResponseController(rw).Flush())

		<-release

		_, err = rw.Write([]byte("tail"))
		require.NoError(t, err)
	})

	// Compress asserts http.Flusher directly on the ResponseWriter it receives:
	// the statusRewriter must expose the full writer surface for flushes to reach the client.
	compressHandler, err := compress.New(t.Context(), next, dynamic.Compress{Encodings: []string{"gzip"}}, "compress")
	require.NoError(t, err)

	ts := httptest.NewServer(wrap(t, time.Hour, compressHandler))
	t.Cleanup(ts.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL, http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Accept-Encoding", "gzip")

	res, err := ts.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = res.Body.Close() })

	require.Equal(t, "gzip", res.Header.Get("Content-Encoding"))

	// The first compressed bytes must arrive while the handler is still blocked.
	buf := make([]byte, 1)
	_, err = res.Body.Read(buf)
	assert.NoError(t, err)

	close(release)
	_, err = io.Copy(io.Discard, res.Body)
	require.NoError(t, err)
}
