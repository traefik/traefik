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
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/middlewares/compress"
	"github.com/traefik/traefik/v3/pkg/middlewares/respondingtimeout"
	"github.com/traefik/traefik/v3/pkg/proxy/httputil"
)

// wrap sets both timeouts to the same value, as an HTTPRoute request timeout does.
func wrap(t *testing.T, timeout time.Duration, next http.Handler) http.Handler {
	t.Helper()

	return wrapTimeouts(t, dynamic.RouterRespondingTimeouts{
		ReadTimeout:  new(ptypes.Duration(timeout)),
		WriteTimeout: new(ptypes.Duration(timeout)),
	}, next)
}

func wrapTimeouts(t *testing.T, timeouts dynamic.RouterRespondingTimeouts, next http.Handler) http.Handler {
	t.Helper()

	handler, err := respondingtimeout.WrapHandler(timeouts)(next)
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

// flushDeadlineListener reports, once per connection, the write deadline in force when the server first
// writes to it.
type flushDeadlineListener struct {
	net.Listener

	deadlines chan<- time.Time
}

func (l *flushDeadlineListener) Accept() (net.Conn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return &flushDeadlineConn{Conn: conn, deadlines: l.deadlines}, nil
}

type flushDeadlineConn struct {
	net.Conn

	deadlines chan<- time.Time

	mu       sync.Mutex
	deadline time.Time
	reported bool
}

func (c *flushDeadlineConn) SetWriteDeadline(deadline time.Time) error {
	c.mu.Lock()
	c.deadline = deadline
	c.mu.Unlock()

	return c.Conn.SetWriteDeadline(deadline)
}

func (c *flushDeadlineConn) Write(b []byte) (int, error) {
	c.mu.Lock()
	if !c.reported {
		c.reported = true
		c.deadlines <- c.deadline
	}
	c.mu.Unlock()

	return c.Conn.Write(b)
}

func TestStatusNormalization(t *testing.T) {
	testCases := []struct {
		desc       string
		status     int
		waitExpiry bool
		expected   int
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
		// The rewriter promotes only the statuses the timeout path itself yields: the proxy maps an expired
		// deadline to a 5xx or to 499 (ComputeStatusCode), never to a 2xx or a non-499 4xx. So a 2xx or 4xx past
		// the deadline can only come from a handler that ignored it — a completed success or a definitive
		// client-side verdict — and is passed through unchanged rather than masked as a 504.
		{
			desc:       "4xx past the deadline is preserved, not masked as 504",
			status:     http.StatusNotFound,
			waitExpiry: true,
			expected:   http.StatusNotFound,
		},
		{
			desc:       "2xx past the deadline is preserved, not masked as 504",
			status:     http.StatusOK,
			waitExpiry: true,
			expected:   http.StatusOK,
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

	// Upgrade requests get no router deadline: well past the timeout, the tunnel must still relay both ways.
	time.Sleep(3 * timeout)

	_, err = conn.Write([]byte("after-deadline"))
	require.NoError(t, err)

	echoed := make([]byte, len("after-deadline"))
	_, err = io.ReadFull(br, echoed)
	require.NoError(t, err)

	assert.Equal(t, "after-deadline", string(echoed))
}

// TestSlowBackend also runs over HTTP/2, where a write deadline is a timer that resets the stream when it fires,
// even before anything is written.
func TestSlowBackend(t *testing.T) {
	hung := func(t *testing.T) http.Handler {
		t.Helper()

		return reverseProxy(hungBackend(t))
	}

	slow := func(t *testing.T) http.Handler {
		t.Helper()

		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			time.Sleep(200 * time.Millisecond)
			rw.WriteHeader(http.StatusOK)
		})
	}

	testCases := []struct {
		desc         string
		timeout      time.Duration
		writeTimeout time.Duration
		http2        bool
		next         func(t *testing.T) http.Handler
		expected     int
	}{
		{
			desc:     "backend slower than the timeout gets 504",
			timeout:  50 * time.Millisecond,
			next:     hung,
			expected: http.StatusGatewayTimeout,
		},
		{
			desc:     "backend slower than the timeout gets 504 over HTTP/2",
			timeout:  50 * time.Millisecond,
			http2:    true,
			next:     hung,
			expected: http.StatusGatewayTimeout,
		},
		{
			desc:         "backend slower than the entrypoint writeTimeout but within the timeout gets 200 over HTTP/2",
			timeout:      time.Second,
			writeTimeout: 50 * time.Millisecond,
			http2:        true,
			next:         slow,
			expected:     http.StatusOK,
		},
		{
			desc:         "backend slower than the entrypoint writeTimeout gets 200 with a disabled timeout over HTTP/2",
			writeTimeout: 50 * time.Millisecond,
			http2:        true,
			next:         slow,
			expected:     http.StatusOK,
		},
		{
			desc:         "backend slower than the timeout gets 504 despite a shorter entrypoint writeTimeout over HTTP/2",
			timeout:      200 * time.Millisecond,
			writeTimeout: 50 * time.Millisecond,
			http2:        true,
			next:         hung,
			expected:     http.StatusGatewayTimeout,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewUnstartedServer(wrap(t, test.timeout, test.next(t)))
			ts.EnableHTTP2 = test.http2
			ts.Config.WriteTimeout = test.writeTimeout
			if test.http2 {
				ts.StartTLS()
			} else {
				ts.Start()
			}
			t.Cleanup(ts.Close)

			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL, http.NoBody)
			require.NoError(t, err)

			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			t.Cleanup(func() { _ = res.Body.Close() })

			if test.http2 {
				require.Equal(t, "HTTP/2.0", res.Proto)
			}
			assert.Equal(t, test.expected, res.StatusCode)
		})
	}
}

// TestDisabledTimeoutLiftsEntryPointDeadlines guards the zero timeouts over HTTP/1, where the entrypoint
// readTimeout and writeTimeout are connection deadlines.
func TestDisabledTimeoutLiftsEntryPointDeadlines(t *testing.T) {
	testCases := []struct {
		desc         string
		timeouts     dynamic.RouterRespondingTimeouts
		readTimeout  time.Duration
		writeTimeout time.Duration
	}{
		{
			desc:        "zero readTimeout and upload slower than the entrypoint readTimeout",
			timeouts:    dynamic.RouterRespondingTimeouts{ReadTimeout: new(ptypes.Duration(0))},
			readTimeout: 50 * time.Millisecond,
		},
		{
			desc:         "zero writeTimeout and backend slower than the entrypoint writeTimeout",
			timeouts:     dynamic.RouterRespondingTimeouts{WriteTimeout: new(ptypes.Duration(0))},
			writeTimeout: 50 * time.Millisecond,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewUnstartedServer(wrapTimeouts(t, test.timeouts, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if _, err := io.Copy(io.Discard, req.Body); err != nil {
					rw.WriteHeader(http.StatusBadRequest)
					return
				}

				time.Sleep(200 * time.Millisecond)
				rw.WriteHeader(http.StatusOK)
			})))
			ts.Config.ReadTimeout = test.readTimeout
			ts.Config.WriteTimeout = test.writeTimeout
			ts.Start()
			t.Cleanup(ts.Close)

			conn, err := net.Dial("tcp", ts.Listener.Addr().String())
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })

			require.NoError(t, conn.SetDeadline(time.Now().Add(5*time.Second)))

			_, err = fmt.Fprintf(conn, "POST / HTTP/1.1\r\nHost: %s\r\nContent-Length: 4\r\n\r\nab", ts.Listener.Addr())
			require.NoError(t, err)

			time.Sleep(200 * time.Millisecond)

			_, err = fmt.Fprint(conn, "cd")
			require.NoError(t, err)

			res, err := http.ReadResponse(bufio.NewReader(conn), nil)
			require.NoError(t, err)
			t.Cleanup(func() { _ = res.Body.Close() })

			assert.Equal(t, http.StatusOK, res.StatusCode)
		})
	}
}

// TestNextKeepAliveRequest guards that the deadlines armed for one request never disturb the next request on the
// same keep-alive connection, served by a router without timeouts.
func TestNextKeepAliveRequest(t *testing.T) {
	testCases := []struct {
		desc          string
		first         func(t *testing.T) http.Handler
		expectedFirst int
		wait          time.Duration
	}{
		{
			// The timed-out router sits below capture, as in production: capture wraps req.Body, so body presence
			// must be read from ContentLength rather than an http.NoBody check, or the read deadline is armed on a
			// body-less request, and its expiry cancels the context of the next request.
			desc: "after a timed-out request",
			first: func(t *testing.T) http.Handler {
				t.Helper()

				captured, err := capture.Wrap(wrap(t, 50*time.Millisecond, reverseProxy(hungBackend(t))))
				require.NoError(t, err)

				return captured
			},
			expectedFirst: http.StatusGatewayTimeout,
		},
		{
			// Waiting past the first request's deadline makes any deadline left on the connection already expired.
			desc: "after a request answered within the timeout",
			first: func(t *testing.T) http.Handler {
				t.Helper()

				return wrap(t, 100*time.Millisecond, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					_, _ = rw.Write([]byte("fast"))
				}))
			},
			expectedFirst: http.StatusOK,
			wait:          200 * time.Millisecond,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			mux := http.NewServeMux()
			mux.Handle("/first", test.first(t))
			mux.Handle("/plain", reverseProxy(healthyBackend(t)))

			ts := httptest.NewServer(mux)
			t.Cleanup(ts.Close)

			conn, err := net.Dial("tcp", ts.Listener.Addr().String())
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })

			require.NoError(t, conn.SetDeadline(time.Now().Add(10*time.Second)))
			br := bufio.NewReader(conn)

			_, err = fmt.Fprintf(conn, "GET /first HTTP/1.1\r\nHost: %s\r\n\r\n", ts.Listener.Addr())
			require.NoError(t, err)

			res, err := http.ReadResponse(br, nil)
			require.NoError(t, err)
			_, err = io.Copy(io.Discard, res.Body)
			require.NoError(t, err)
			require.Equal(t, test.expectedFirst, res.StatusCode)

			time.Sleep(test.wait)

			_, err = fmt.Fprintf(conn, "GET /plain HTTP/1.1\r\nHost: %s\r\n\r\n", ts.Listener.Addr())
			require.NoError(t, err)

			res, err = http.ReadResponse(br, nil)
			require.NoError(t, err)
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, "backend", string(body))
		})
	}
}

func TestSlowClientBounded(t *testing.T) {
	timeout := 100 * time.Millisecond

	// The backend is healthy: only the read deadline can bound a client that stalls mid-upload.
	proxy := func(t *testing.T) http.Handler {
		t.Helper()

		return reverseProxy(healthyBackend(t))
	}

	// Without a writeTimeout, the status depends on whether the proxy sees the read timeout or the canceled
	// request context first, so those cases do not check it.
	testCases := []struct {
		desc     string
		handler  func(t *testing.T) http.Handler
		expected int
	}{
		{
			// Only a deadline comparison can normalize the outcome, since the connection deadline may cancel the
			// request context before the deadline context does.
			desc: "readTimeout sharing the writeTimeout deadline gets 504",
			handler: func(t *testing.T) http.Handler {
				t.Helper()

				return wrap(t, timeout, proxy(t))
			},
			expected: http.StatusGatewayTimeout,
		},
		{
			desc: "readTimeout alone",
			handler: func(t *testing.T) http.Handler {
				t.Helper()

				return wrapTimeouts(t, dynamic.RouterRespondingTimeouts{ReadTimeout: new(ptypes.Duration(timeout))}, proxy(t))
			},
		},
		{
			desc: "child router readTimeout cannot extend its parent one",
			handler: func(t *testing.T) http.Handler {
				t.Helper()

				child := wrapTimeouts(t, dynamic.RouterRespondingTimeouts{ReadTimeout: new(ptypes.Duration(time.Hour))}, proxy(t))

				return wrapTimeouts(t, dynamic.RouterRespondingTimeouts{ReadTimeout: new(ptypes.Duration(timeout))}, child)
			},
		},
		{
			// The handler answers without reading the body, as an authentication or a rate limit middleware would.
			// The server drains the unread body before flushing the response (net/http chunkWriter.writeHeader),
			// after the handler and its defers have run, so only a read deadline left armed on return bounds it.
			desc: "handler answering without reading the body",
			handler: func(t *testing.T) http.Handler {
				t.Helper()

				return wrap(t, timeout, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					rw.WriteHeader(http.StatusUnauthorized)
				}))
			},
			expected: http.StatusUnauthorized,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			ts := httptest.NewServer(test.handler(t))
			t.Cleanup(ts.Close)

			conn, err := net.Dial("tcp", ts.Listener.Addr().String())
			require.NoError(t, err)
			t.Cleanup(func() { _ = conn.Close() })

			// Announce a large body and stall after a few bytes.
			_, err = fmt.Fprintf(conn, "POST / HTTP/1.1\r\nHost: %s\r\nContent-Length: 1000\r\n\r\npartial", ts.Listener.Addr())
			require.NoError(t, err)

			require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))

			start := time.Now()
			res, err := http.ReadResponse(bufio.NewReader(conn), nil)
			require.NoError(t, err)

			if test.expected != 0 {
				assert.Equal(t, test.expected, res.StatusCode)
			}
			assert.Less(t, time.Since(start), 10*timeout)
		})
	}
}

// TestReadTimeoutDoesNotBoundBackend guards that the read deadline only bounds the body: net/http clears it once
// the body is read, before reading in the background to detect a client disconnection.
func TestReadTimeoutDoesNotBoundBackend(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(wrapTimeouts(t, dynamic.RouterRespondingTimeouts{ReadTimeout: new(ptypes.Duration(50 * time.Millisecond))},
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if _, err := io.Copy(io.Discard, req.Body); err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}

			select {
			case <-req.Context().Done():
				rw.WriteHeader(httputil.StatusClientClosedRequest)
			case <-time.After(200 * time.Millisecond):
				rw.WriteHeader(http.StatusOK)
			}
		})))
	t.Cleanup(ts.Close)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, ts.URL, strings.NewReader("body"))
	require.NoError(t, err)

	res, err := ts.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = res.Body.Close() })

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

// TestFlushKeepsEntryPointWriteTimeout guards the deferred restore. A handler writing a small response leaves
// it entirely buffered, so the first write on the connection is the post-handler flush (net/http
// finishRequest): the deadline it runs under is the one this middleware left behind.
func TestFlushKeepsEntryPointWriteTimeout(t *testing.T) {
	testCases := []struct {
		desc         string
		writeTimeout time.Duration
	}{
		{
			desc: "the entrypoint has no write timeout",
		},
		{
			desc:         "the entrypoint has a write timeout",
			writeTimeout: time.Second,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			// A timeout far beyond the entrypoint write timeout: the deadline the flush runs under tells
			// unambiguously which of the two the middleware left on the connection.
			ts := httptest.NewUnstartedServer(wrap(t, time.Hour, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				_, _ = rw.Write([]byte("ok"))
			})))

			deadlines := make(chan time.Time, 1)
			ts.Listener = &flushDeadlineListener{Listener: ts.Listener, deadlines: deadlines}
			ts.Config.WriteTimeout = test.writeTimeout

			ts.Start()
			t.Cleanup(ts.Close)

			req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL, http.NoBody)
			require.NoError(t, err)

			start := time.Now()

			res, err := ts.Client().Do(req)
			require.NoError(t, err)
			t.Cleanup(func() { _ = res.Body.Close() })
			require.Equal(t, http.StatusOK, res.StatusCode)

			var deadline time.Time
			select {
			case deadline = <-deadlines:
			case <-time.After(5 * time.Second):
				require.FailNow(t, "the response never reached the connection")
			}

			if test.writeTimeout == 0 {
				assert.True(t, deadline.IsZero())
				return
			}

			assert.False(t, deadline.IsZero())
			// The flush runs right after the handler, so the recorded deadline is start+writeTimeout within the
			// flush latency; a tolerance well under the 1h timeout proves it is the entrypoint writeTimeout, not the router one.
			assert.WithinDuration(t, start.Add(test.writeTimeout), deadline, 2*time.Second)
		})
	}
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
