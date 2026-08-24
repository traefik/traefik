package fast

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

func TestProxyBuilder_ForwardingTimeouts(t *testing.T) {
	testCases := []struct {
		desc                 string
		forwardingTimeouts   *dynamic.ForwardingTimeouts
		expectedReadTimeout  time.Duration
		expectedWriteTimeout time.Duration
	}{
		{
			desc: "no forwarding timeouts",
		},
		{
			desc:               "no read/write timeout set",
			forwardingTimeouts: &dynamic.ForwardingTimeouts{},
		},
		{
			desc: "read timeout set",
			forwardingTimeouts: &dynamic.ForwardingTimeouts{
				ReadTimeout: ptypes.Duration(50 * time.Millisecond),
			},
			expectedReadTimeout: 50 * time.Millisecond,
		},
		{
			desc: "write timeout set",
			forwardingTimeouts: &dynamic.ForwardingTimeouts{
				WriteTimeout: ptypes.Duration(50 * time.Millisecond),
			},
			expectedWriteTimeout: 50 * time.Millisecond,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			builder := NewProxyBuilder(&transportManagerMock{forwardingTimeouts: test.forwardingTimeouts}, static.FastProxyConfig{})

			cfg, err := builder.transportManager.Get("test")
			require.NoError(t, err)

			pool := builder.getPool("test", cfg, nil, testhelpers.MustParseURL("http://127.0.0.1:0"), nil)
			t.Cleanup(pool.Close)

			assert.Equal(t, test.expectedReadTimeout, pool.readTimeout)
			assert.Equal(t, test.expectedWriteTimeout, pool.writeTimeout)
		})
	}
}

func TestReadTimeout_idlePooledConnection(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	var accepts atomic.Int32
	go func() {
		for {
			backendConn, err := ln.Accept()
			if err != nil {
				return
			}
			accepts.Add(1)

			go func() {
				defer backendConn.Close()

				br := bufio.NewReader(backendConn)
				for {
					if err := skipRequest(br); err != nil {
						return
					}
					if _, err := backendConn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n")); err != nil {
						return
					}
				}
			}()
		}
	}()

	readTimeout := 100 * time.Millisecond
	proxyHandler := buildProxyHandler(t, &dynamic.ForwardingTimeouts{ReadTimeout: ptypes.Duration(readTimeout)}, ln.Addr().String())

	rec := httptest.NewRecorder()
	proxyHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://"+ln.Addr().String(), http.NoBody))
	require.Equal(t, http.StatusOK, rec.Code)

	// The pooled keep-alive connection must survive an idle period longer than
	// the read timeout, and be reused for the next request.
	time.Sleep(3 * readTimeout)

	rec = httptest.NewRecorder()
	proxyHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://"+ln.Addr().String(), http.NoBody))
	require.Equal(t, http.StatusOK, rec.Code)

	assert.Equal(t, int32(1), accepts.Load())
}

func TestReadTimeout_silentBackend(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	go func() {
		for {
			backendConn, err := ln.Accept()
			if err != nil {
				return
			}

			go func() {
				defer backendConn.Close()

				br := bufio.NewReader(backendConn)
				if err := skipRequest(br); err != nil {
					return
				}
				// Never respond, and keep the connection open until the proxy
				// gives up and closes it.
				_, _ = br.ReadByte()
			}()
		}
	}()

	proxyHandler := buildProxyHandler(t, &dynamic.ForwardingTimeouts{ReadTimeout: ptypes.Duration(100 * time.Millisecond)}, ln.Addr().String())

	start := time.Now()
	rec := httptest.NewRecorder()
	proxyHandler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "http://"+ln.Addr().String(), http.NoBody))

	require.Equal(t, http.StatusGatewayTimeout, rec.Code)
	assert.Less(t, time.Since(start), 5*time.Second)
}

func TestReadTimeout_upgradedConnection(t *testing.T) {
	upgrader := gorillawebsocket.Upgrader{}
	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c, err := upgrader.Upgrade(rw, req, nil)
		if err != nil {
			return
		}
		defer c.Close()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				return
			}
			if err := c.WriteMessage(mt, message); err != nil {
				return
			}
		}
	}))
	t.Cleanup(backend.Close)

	readTimeout := 100 * time.Millisecond
	forwardingTimeouts := &dynamic.ForwardingTimeouts{
		ReadTimeout:  ptypes.Duration(readTimeout),
		WriteTimeout: ptypes.Duration(readTimeout),
	}
	proxyHandler := buildProxyHandler(t, forwardingTimeouts, testhelpers.MustParseURL(backend.URL).Host)

	proxy := httptest.NewServer(proxyHandler)
	t.Cleanup(proxy.Close)

	conn, _, err := gorillawebsocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(proxy.URL, "http"), nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	require.NoError(t, conn.WriteMessage(gorillawebsocket.TextMessage, []byte("ping")))

	_, message, err := conn.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, "ping", string(message))

	// The upgraded connection must survive an idle period longer than the read timeout.
	time.Sleep(3 * readTimeout)

	require.NoError(t, conn.WriteMessage(gorillawebsocket.TextMessage, []byte("pong")))

	_, message, err = conn.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, "pong", string(message))
}

func TestConnTimeouts(t *testing.T) {
	testCases := []struct {
		desc             string
		readTimeout      time.Duration
		writeTimeout     time.Duration
		expectedResponse bool
		responsePending  bool
		upgraded         bool
		reads            bool
		expectedTimeout  bool
	}{
		{
			desc:             "read deadline armed while a response is pending",
			readTimeout:      50 * time.Millisecond,
			expectedResponse: true,
			responsePending:  true,
			reads:            true,
			expectedTimeout:  true,
		},
		{
			desc:        "read deadline not armed on idle connection",
			readTimeout: 50 * time.Millisecond,
			reads:       true,
		},
		{
			// expectedResponse is set before the request is written: arming the
			// deadline this early breaks requests slower to write than readTimeout.
			desc:             "read deadline not armed while the request is being written",
			readTimeout:      50 * time.Millisecond,
			expectedResponse: true,
			reads:            true,
		},
		{
			desc:             "read deadline not armed on upgraded connection",
			readTimeout:      50 * time.Millisecond,
			expectedResponse: true,
			responsePending:  true,
			upgraded:         true,
			reads:            true,
		},
		{
			desc:            "write deadline armed",
			writeTimeout:    50 * time.Millisecond,
			expectedTimeout: true,
		},
		{
			desc:         "write deadline not armed on upgraded connection",
			writeTimeout: 50 * time.Millisecond,
			upgraded:     true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			// net.Pipe has no OS buffering: reads and writes block until the other
			// side is ready, which allows the deadlines to be exercised deterministically.
			client, server := net.Pipe()
			t.Cleanup(func() { _ = client.Close() })
			t.Cleanup(func() { _ = server.Close() })

			co := &conn{
				Conn:         client,
				readTimeout:  test.readTimeout,
				writeTimeout: test.writeTimeout,
			}
			co.br = bufio.NewReaderSize(timeoutReader{conn: co}, bufioSize)
			co.expectedResponse.Store(test.expectedResponse)
			co.responsePending.Store(test.responsePending)
			co.upgraded.Store(test.upgraded)

			// The peer stays silent for much longer than the configured timeouts,
			// so that an operation succeeds only when no deadline is armed.
			go func() {
				time.Sleep(300 * time.Millisecond)
				if test.reads {
					_, _ = server.Write([]byte("A"))
					return
				}
				buf := make([]byte, 1)
				_, _ = server.Read(buf)
			}()

			var err error
			if test.reads {
				buf := make([]byte, 1)
				_, err = co.Read(buf)
			} else {
				_, err = co.Write([]byte("A"))
			}

			if test.expectedTimeout {
				var netErr net.Error
				require.ErrorAs(t, err, &netErr)
				require.True(t, netErr.Timeout())
				return
			}
			require.NoError(t, err)
		})
	}
}

func buildProxyHandler(t *testing.T, forwardingTimeouts *dynamic.ForwardingTimeouts, backendAddr string) http.Handler {
	t.Helper()

	builder := NewProxyBuilder(&transportManagerMock{forwardingTimeouts: forwardingTimeouts, maxIdleConnsPerHost: 1}, static.FastProxyConfig{})
	t.Cleanup(func() {
		for _, pools := range builder.pools {
			for _, pool := range pools {
				pool.Close()
			}
		}
	})

	proxyHandler, err := builder.Build("default", testhelpers.MustParseURL("http://"+backendAddr), false, false)
	require.NoError(t, err)

	return proxyHandler
}

func skipRequest(br *bufio.Reader) error {
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return err
		}
		if line == "\r\n" {
			return nil
		}
	}
}
