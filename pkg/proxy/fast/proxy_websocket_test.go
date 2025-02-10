package fast

import (
	"bufio"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/websocket"
)

const dialTimeout = time.Second

func TestWebSocketUpgradeCase(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		challengeKey := r.Header.Get("Sec-Websocket-Key")

		hijacker, ok := w.(http.Hijacker)
		require.True(t, ok)

		c, _, err := hijacker.Hijack()
		require.NoError(t, err)

		// Force answer with "Connection: upgrade" in lowercase.
		_, err = c.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: upgrade\r\nSec-WebSocket-Accept: " + computeAcceptKey(challengeKey) + "\r\n\n"))
		require.NoError(t, err)
	}))
	defer srv.Close()

	proxy := createProxyWithForwarder(t, srv.URL, createConnectionPool(srv.URL, nil))

	proxyAddr := proxy.Listener.Addr().String()
	_, conn, err := newWebsocketRequest(
		withServer(proxyAddr),
		withPath("/ws"),
	).open()
	require.NoError(t, err)

	conn.Close()
}

func TestCleanWebSocketHeaders(t *testing.T) {
	// Asserts that no headers are sent if the request contain anything.
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	cleanWebSocketHeaders(&req.Header)

	want := "GET / HTTP/1.1\r\n\r\n"
	assert.Equal(t, want, req.Header.String())

	// Asserts that the Sec-WebSocket-* is enforced.
	req.Reset()

	req.Header.Set("Sec-Websocket-Key", "key")
	req.Header.Set("Sec-Websocket-Extensions", "extensions")
	req.Header.Set("Sec-Websocket-Accept", "accept")
	req.Header.Set("Sec-Websocket-Protocol", "protocol")
	req.Header.Set("Sec-Websocket-Version", "version")

	cleanWebSocketHeaders(&req.Header)

	want = "GET / HTTP/1.1\r\nSec-WebSocket-Key: key\r\nSec-WebSocket-Extensions: extensions\r\nSec-WebSocket-Accept: accept\r\nSec-WebSocket-Protocol: protocol\r\nSec-WebSocket-Version: version\r\n\r\n"
	assert.Equal(t, want, req.Header.String())
}

func TestWebSocketTCPClose(t *testing.T) {
	errChan := make(chan error, 1)
	upgrader := gorillawebsocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		for {
			_, _, err := c.ReadMessage()
			if err != nil {
				errChan <- err
				break
			}
		}
	}))
	defer srv.Close()

	proxy := createProxyWithForwarder(t, srv.URL, createConnectionPool(srv.URL, nil))

	proxyAddr := proxy.Listener.Addr().String()
	_, conn, err := newWebsocketRequest(
		withServer(proxyAddr),
		withPath("/ws"),
	).open()
	require.NoError(t, err)

	conn.Close()

	serverErr := <-errChan

	var wsErr *gorillawebsocket.CloseError
	require.ErrorAs(t, serverErr, &wsErr)
	assert.Equal(t, 1006, wsErr.Code)
}

func TestWebSocketPingPong(t *testing.T) {
	upgrader := gorillawebsocket.Upgrader{
		HandshakeTimeout: 10 * time.Second,
		CheckOrigin: func(*http.Request) bool {
			return true
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		ws, err := upgrader.Upgrade(writer, request, nil)
		require.NoError(t, err)

		ws.SetPingHandler(func(appData string) error {
			err = ws.WriteMessage(gorillawebsocket.PongMessage, []byte(appData+"Pong"))
			require.NoError(t, err)

			return nil
		})

		_, _, _ = ws.ReadMessage()
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mux.ServeHTTP(w, req)
	}))
	defer srv.Close()

	proxy := createProxyWithForwarder(t, srv.URL, createConnectionPool(srv.URL, nil))
	serverAddr := proxy.Listener.Addr().String()

	headers := http.Header{}
	webSocketURL := "ws://" + serverAddr + "/ws"
	headers.Add("Origin", webSocketURL)

	conn, resp, err := gorillawebsocket.DefaultDialer.Dial(webSocketURL, headers)
	require.NoError(t, err, "Error during Dial with response: %+v", resp)
	defer conn.Close()

	goodErr := fmt.Errorf("signal: %s", "Good data")
	badErr := fmt.Errorf("signal: %s", "Bad data")
	conn.SetPongHandler(func(data string) error {
		if data == "PingPong" {
			return goodErr
		}

		return badErr
	})

	err = conn.WriteControl(gorillawebsocket.PingMessage, []byte("Ping"), time.Now().Add(time.Second))
	require.NoError(t, err)

	_, _, err = conn.ReadMessage()

	if !errors.Is(err, goodErr) {
		require.NoError(t, err)
	}
}

func TestWebSocketEcho(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/ws", websocket.Handler(func(conn *websocket.Conn) {
		msg := make([]byte, 4)

		n, err := conn.Read(msg)
		require.NoError(t, err)

		_, err = conn.Write(msg[:n])
		require.NoError(t, err)

		err = conn.Close()
		require.NoError(t, err)
	}))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mux.ServeHTTP(w, req)
	}))
	defer srv.Close()

	proxy := createProxyWithForwarder(t, srv.URL, createConnectionPool(srv.URL, nil))
	serverAddr := proxy.Listener.Addr().String()

	headers := http.Header{}
	webSocketURL := "ws://" + serverAddr + "/ws"
	headers.Add("Origin", webSocketURL)

	conn, resp, err := gorillawebsocket.DefaultDialer.Dial(webSocketURL, headers)
	require.NoError(t, err, "Error during Dial with response: %+v", resp)

	err = conn.WriteMessage(gorillawebsocket.TextMessage, []byte("OK"))
	require.NoError(t, err)

	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)

	assert.Equal(t, "OK", string(msg))

	err = conn.Close()
	require.NoError(t, err)
}

func TestWebSocketPassHost(t *testing.T) {
	testCases := []struct {
		desc     string
		passHost bool
		expected string
	}{
		{
			desc:     "PassHost false",
			passHost: false,
		},
		{
			desc:     "PassHost true",
			passHost: true,
			expected: "example.com",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			mux := http.NewServeMux()
			mux.Handle("/ws", websocket.Handler(func(conn *websocket.Conn) {
				req := conn.Request()

				if test.passHost {
					require.Equal(t, test.expected, req.Host)
				} else {
					require.NotEqual(t, test.expected, req.Host)
				}

				msg := make([]byte, 4)

				n, err := conn.Read(msg)
				require.NoError(t, err)

				_, err = conn.Write(msg[:n])
				require.NoError(t, err)

				err = conn.Close()
				require.NoError(t, err)
			}))

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				mux.ServeHTTP(w, req)
			}))
			defer srv.Close()

			proxy := createProxyWithForwarder(t, srv.URL, createConnectionPool(srv.URL, nil))

			serverAddr := proxy.Listener.Addr().String()

			headers := http.Header{}
			webSocketURL := "ws://" + serverAddr + "/ws"
			headers.Add("Origin", webSocketURL)
			headers.Add("Host", "example.com")

			conn, resp, err := gorillawebsocket.DefaultDialer.Dial(webSocketURL, headers)
			require.NoError(t, err, "Error during Dial with response: %+v", resp)

			err = conn.WriteMessage(gorillawebsocket.TextMessage, []byte("OK"))
			require.NoError(t, err)

			_, msg, err := conn.ReadMessage()
			require.NoError(t, err)

			assert.Equal(t, "OK", string(msg))

			err = conn.Close()
			require.NoError(t, err)
		})
	}
}

func TestWebSocketServerWithoutCheckOrigin(t *testing.T) {
	upgrader := gorillawebsocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))
	defer srv.Close()

	proxy := createProxyWithForwarder(t, srv.URL, createConnectionPool(srv.URL, nil))
	defer proxy.Close()

	proxyAddr := proxy.Listener.Addr().String()
	resp, err := newWebsocketRequest(
		withServer(proxyAddr),
		withPath("/ws"),
		withData("ok"),
		withOrigin("http://127.0.0.2"),
	).send()

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestWebSocketRequestWithOrigin(t *testing.T) {
	upgrader := gorillawebsocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))
	defer srv.Close()

	proxy := createProxyWithForwarder(t, srv.URL, createConnectionPool(srv.URL, nil))
	defer proxy.Close()

	proxyAddr := proxy.Listener.Addr().String()
	_, err := newWebsocketRequest(
		withServer(proxyAddr),
		withPath("/ws"),
		withData("echo"),
		withOrigin("http://127.0.0.2"),
	).send()
	require.EqualError(t, err, "bad status")

	resp, err := newWebsocketRequest(
		withServer(proxyAddr),
		withPath("/ws"),
		withData("ok"),
	).send()

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestWebSocketRequestWithQueryParams(t *testing.T) {
	upgrader := gorillawebsocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		assert.Equal(t, "test", r.URL.Query().Get("query"))
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			err = conn.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))
	defer srv.Close()

	proxy := createProxyWithForwarder(t, srv.URL, createConnectionPool(srv.URL, nil))
	defer proxy.Close()

	proxyAddr := proxy.Listener.Addr().String()

	resp, err := newWebsocketRequest(
		withServer(proxyAddr),
		withPath("/ws?query=test"),
		withData("ok"),
	).send()

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestWebSocketRequestWithHeadersInResponseWriter(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/ws", websocket.Handler(func(conn *websocket.Conn) {
		_ = conn.Close()
	}))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mux.ServeHTTP(w, req)
	}))
	defer srv.Close()

	u := parseURI(t, srv.URL)

	f, err := NewReverseProxy(u, nil, true, false, false, newConnPool(1, 0, 0, func() (net.Conn, error) {
		return net.Dial("tcp", u.Host)
	}))
	require.NoError(t, err)

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.URL = parseURI(t, srv.URL)
		w.Header().Set("HEADER-KEY", "HEADER-VALUE")
		f.ServeHTTP(w, req)
	}))
	defer proxy.Close()

	serverAddr := proxy.Listener.Addr().String()

	headers := http.Header{}
	webSocketURL := "ws://" + serverAddr + "/ws"
	headers.Add("Origin", webSocketURL)
	conn, resp, err := gorillawebsocket.DefaultDialer.Dial(webSocketURL, headers)
	require.NoError(t, err, "Error during Dial with response: %+v", err, resp)
	defer conn.Close()

	assert.Equal(t, "HEADER-VALUE", resp.Header.Get("HEADER-KEY"))
}

func TestWebSocketRequestWithEncodedChar(t *testing.T) {
	upgrader := gorillawebsocket.Upgrader{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		assert.Equal(t, "/%3A%2F%2F", r.URL.EscapedPath())
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break
			}
			err = conn.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))
	defer srv.Close()

	proxy := createProxyWithForwarder(t, srv.URL, createConnectionPool(srv.URL, nil))
	defer proxy.Close()

	proxyAddr := proxy.Listener.Addr().String()

	resp, err := newWebsocketRequest(
		withServer(proxyAddr),
		withPath("/%3A%2F%2F"),
		withData("ok"),
	).send()

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestWebSocketUpgradeFailed(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mux.ServeHTTP(w, req)
	}))
	defer srv.Close()

	u := parseURI(t, srv.URL)
	f, err := NewReverseProxy(u, nil, true, false, false, newConnPool(1, 0, 0, func() (net.Conn, error) {
		return net.Dial("tcp", u.Host)
	}))
	require.NoError(t, err)

	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path // keep the original path

		if path != "/ws" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Set new backend URL
		req.URL = parseURI(t, srv.URL)
		req.URL.Path = path
		f.ServeHTTP(w, req)
	}))
	defer proxy.Close()

	proxyAddr := proxy.Listener.Addr().String()
	conn, err := net.DialTimeout("tcp", proxyAddr, dialTimeout)

	require.NoError(t, err)
	defer conn.Close()

	req, err := http.NewRequest(http.MethodGet, "ws://127.0.0.1/ws", nil)
	require.NoError(t, err)

	req.Header.Add("upgrade", "websocket")
	req.Header.Add("Connection", "upgrade")

	err = req.Write(conn)
	require.NoError(t, err)

	// First request works with 400
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, req)
	require.NoError(t, err)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestForwardsWebsocketTraffic(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/ws", websocket.Handler(func(conn *websocket.Conn) {
		_, err := conn.Write([]byte("ok"))
		require.NoError(t, err)

		err = conn.Close()
		require.NoError(t, err)
	}))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mux.ServeHTTP(w, req)
	}))
	defer srv.Close()

	proxy := createProxyWithForwarder(t, srv.URL, createConnectionPool(srv.URL, nil))
	defer proxy.Close()

	proxyAddr := proxy.Listener.Addr().String()
	resp, err := newWebsocketRequest(
		withServer(proxyAddr),
		withPath("/ws"),
		withData("echo"),
	).send()

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func TestWebSocketTransferTLSConfig(t *testing.T) {
	srv := createTLSWebsocketServer()
	defer srv.Close()

	proxyWithoutTLSConfig := createProxyWithForwarder(t, srv.URL, createConnectionPool(srv.URL, nil))
	defer proxyWithoutTLSConfig.Close()

	proxyAddr := proxyWithoutTLSConfig.Listener.Addr().String()

	_, err := newWebsocketRequest(
		withServer(proxyAddr),
		withPath("/ws"),
		withData("ok"),
	).send()

	require.EqualError(t, err, "bad status")

	pool := createConnectionPool(srv.URL, &tls.Config{InsecureSkipVerify: true})

	proxyWithTLSConfig := createProxyWithForwarder(t, srv.URL, pool)
	defer proxyWithTLSConfig.Close()

	proxyAddr = proxyWithTLSConfig.Listener.Addr().String()

	resp, err := newWebsocketRequest(
		withServer(proxyAddr),
		withPath("/ws"),
		withData("ok"),
	).send()

	require.NoError(t, err)
	assert.Equal(t, "ok", resp)
}

func createTLSWebsocketServer() *httptest.Server {
	upgrader := gorillawebsocket.Upgrader{}
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			err = conn.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	}))
	return srv
}

type websocketRequestOpt func(w *websocketRequest)

func withServer(server string) websocketRequestOpt {
	return func(w *websocketRequest) {
		w.ServerAddr = server
	}
}

func withPath(path string) websocketRequestOpt {
	return func(w *websocketRequest) {
		w.Path = path
	}
}

func withData(data string) websocketRequestOpt {
	return func(w *websocketRequest) {
		w.Data = data
	}
}

func withOrigin(origin string) websocketRequestOpt {
	return func(w *websocketRequest) {
		w.Origin = origin
	}
}

func newWebsocketRequest(opts ...websocketRequestOpt) *websocketRequest {
	wsrequest := &websocketRequest{}
	for _, opt := range opts {
		opt(wsrequest)
	}

	if wsrequest.Origin == "" {
		wsrequest.Origin = "http://" + wsrequest.ServerAddr
	}

	if wsrequest.Config == nil {
		wsrequest.Config, _ = websocket.NewConfig(fmt.Sprintf("ws://%s%s", wsrequest.ServerAddr, wsrequest.Path), wsrequest.Origin)
	}

	return wsrequest
}

type websocketRequest struct {
	ServerAddr string
	Path       string
	Data       string
	Origin     string
	Config     *websocket.Config
}

func (w *websocketRequest) send() (string, error) {
	conn, _, err := w.open()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	if _, err := conn.Write([]byte(w.Data)); err != nil {
		return "", err
	}

	msg := make([]byte, 512)

	var n int
	n, err = conn.Read(msg)
	if err != nil {
		return "", err
	}

	received := string(msg[:n])
	return received, nil
}

func (w *websocketRequest) open() (*websocket.Conn, net.Conn, error) {
	client, err := net.DialTimeout("tcp", w.ServerAddr, dialTimeout)
	if err != nil {
		return nil, nil, err
	}

	conn, err := websocket.NewClient(w.Config, client)
	if err != nil {
		return nil, nil, err
	}

	return conn, client, err
}

func parseURI(t *testing.T, uri string) *url.URL {
	t.Helper()

	out, err := url.ParseRequestURI(uri)
	require.NoError(t, err)

	return out
}

func createConnectionPool(target string, tlsConfig *tls.Config) *connPool {
	u := testhelpers.MustParseURL(target)
	return newConnPool(200, 0, 0, func() (net.Conn, error) {
		if tlsConfig != nil {
			return tls.Dial("tcp", u.Host, tlsConfig)
		}

		return net.Dial("tcp", u.Host)
	})
}

func createProxyWithForwarder(t *testing.T, uri string, pool *connPool) *httptest.Server {
	t.Helper()

	u := parseURI(t, uri)
	proxy, err := NewReverseProxy(u, nil, false, true, false, pool)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path // keep the original path
		// Set new backend URL
		req.URL = u
		req.URL.Path = path

		proxy.ServeHTTP(w, req)
	}))
	t.Cleanup(srv.Close)

	return srv
}

func computeAcceptKey(challengeKey string) string {
	h := sha1.New() // #nosec G401 -- (CWE-326) https://datatracker.ietf.org/doc/html/rfc6455#page-54
	h.Write([]byte(challengeKey))
	h.Write([]byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
