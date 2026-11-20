package fast

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/armon/go-socks5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
	proxyhttputil "github.com/traefik/traefik/v3/pkg/proxy/httputil"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

const (
	proxyHTTP   = "http"
	proxyHTTPS  = "https"
	proxySocks5 = "socks"
)

type authCreds struct {
	user     string
	password string
}

func TestProxyFromEnvironment(t *testing.T) {
	testCases := []struct {
		desc      string
		proxyType string
		tls       bool
		auth      *authCreds
	}{
		{
			desc:      "Proxy HTTP with HTTP Backend",
			proxyType: proxyHTTP,
		},
		{
			desc:      "Proxy HTTP with HTTP backend and proxy auth",
			proxyType: proxyHTTP,
			tls:       false,
			auth: &authCreds{
				user:     "user",
				password: "password",
			},
		},
		{
			desc:      "Proxy HTTP with HTTPS backend",
			proxyType: proxyHTTP,
			tls:       true,
		},
		{
			desc:      "Proxy HTTP with HTTPS backend and proxy auth",
			proxyType: proxyHTTP,
			tls:       true,
			auth: &authCreds{
				user:     "user",
				password: "password",
			},
		},
		{
			desc:      "Proxy HTTPS with HTTP backend",
			proxyType: proxyHTTPS,
		},
		{
			desc:      "Proxy HTTPS with HTTP backend and proxy auth",
			proxyType: proxyHTTPS,
			tls:       false,
			auth: &authCreds{
				user:     "user",
				password: "password",
			},
		},
		{
			desc:      "Proxy HTTPS with HTTPS backend",
			proxyType: proxyHTTPS,
			tls:       true,
		},
		{
			desc:      "Proxy HTTPS with HTTPS backend and proxy auth",
			proxyType: proxyHTTPS,
			tls:       true,
			auth: &authCreds{
				user:     "user",
				password: "password",
			},
		},
		{
			desc:      "Proxy Socks5 with HTTP backend",
			proxyType: proxySocks5,
		},
		{
			desc:      "Proxy Socks5 with HTTP backend and proxy auth",
			proxyType: proxySocks5,
			auth: &authCreds{
				user:     "user",
				password: "password",
			},
		},
		{
			desc:      "Proxy Socks5 with HTTPS backend",
			proxyType: proxySocks5,
			tls:       true,
		},
		{
			desc:      "Proxy Socks5 with HTTPS backend and proxy auth",
			proxyType: proxySocks5,
			tls:       true,
			auth: &authCreds{
				user:     "user",
				password: "password",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var backendServer *httptest.Server
			if test.tls {
				backendServer = httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					_, _ = rw.Write([]byte("backendTLS"))
				}))
			} else {
				backendServer = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					_, _ = rw.Write([]byte("backend"))
				}))
			}
			t.Cleanup(backendServer.Close)

			var proxyCalled bool
			proxyHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				proxyCalled = true

				if test.auth != nil {
					proxyAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(test.auth.user+":"+test.auth.password))
					require.Equal(t, proxyAuth, req.Header.Get("Proxy-Authorization"))
				}

				if req.Method != http.MethodConnect {
					proxy := httputil.NewSingleHostReverseProxy(testhelpers.MustParseURL("http://" + req.Host))
					proxy.ServeHTTP(rw, req)
					return
				}

				// CONNECT method
				conn, err := net.Dial("tcp", req.Host)
				require.NoError(t, err)

				hj, ok := rw.(http.Hijacker)
				require.True(t, ok)

				rw.WriteHeader(http.StatusOK)
				connHj, _, err := hj.Hijack()
				require.NoError(t, err)

				defer func() {
					_ = connHj.Close()
					_ = conn.Close()
				}()

				errCh := make(chan error, 1)
				go func() {
					_, err = io.Copy(connHj, conn)
					errCh <- err
				}()
				go func() {
					_, err = io.Copy(conn, connHj)
					errCh <- err
				}()
				<-errCh // Wait for one of the copy operations to finish
			})

			var proxyURL string
			var proxyCert *x509.Certificate

			switch test.proxyType {
			case proxySocks5:
				ln, err := net.Listen("tcp", ":0")
				require.NoError(t, err)

				proxyURL = fmt.Sprintf("socks5://%s", ln.Addr())

				go func() {
					conn, err := ln.Accept()
					require.NoError(t, err)

					proxyCalled = true

					conf := &socks5.Config{}
					if test.auth != nil {
						conf.Credentials = socks5.StaticCredentials{test.auth.user: test.auth.password}
					}

					server, err := socks5.New(conf)
					require.NoError(t, err)

					// We are not checking the error, because ServeConn is blocked until the client or the backend
					// connection is closed which, in some cases, raises a connection reset by peer error.
					_ = server.ServeConn(conn)

					err = ln.Close()
					require.NoError(t, err)
				}()

			case proxyHTTP:
				proxyServer := httptest.NewServer(proxyHandler)
				t.Cleanup(proxyServer.Close)

				proxyURL = proxyServer.URL

			case proxyHTTPS:
				proxyServer := httptest.NewTLSServer(proxyHandler)
				t.Cleanup(proxyServer.Close)

				proxyURL = proxyServer.URL
				proxyCert = proxyServer.Certificate()
			}

			certPool := x509.NewCertPool()
			if proxyCert != nil {
				certPool.AddCert(proxyCert)
			}
			if backendServer.Certificate() != nil {
				certPool.AddCert(backendServer.Certificate())
			}

			builder := NewProxyBuilder(&transportManagerMock{tlsConfig: &tls.Config{RootCAs: certPool}}, static.FastProxyConfig{})
			builder.proxy = func(req *http.Request) (*url.URL, error) {
				u, err := url.Parse(proxyURL)
				if err != nil {
					return nil, err
				}

				if test.auth != nil {
					u.User = url.UserPassword(test.auth.user, test.auth.password)
				}

				return u, nil
			}

			reverseProxy, err := builder.Build("foo", testhelpers.MustParseURL(backendServer.URL), false, false)
			require.NoError(t, err)

			reverseProxyServer := httptest.NewServer(reverseProxy)
			t.Cleanup(reverseProxyServer.Close)

			client := http.Client{Timeout: 5 * time.Second}

			resp, err := client.Get(reverseProxyServer.URL)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if test.tls {
				assert.Equal(t, "backendTLS", string(body))
			} else {
				assert.Equal(t, "backend", string(body))
			}
			assert.True(t, proxyCalled)
		})
	}
}

func TestPreservePath(t *testing.T) {
	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		callCount++
		assert.Equal(t, "/base/foo/bar", req.URL.Path)
		assert.Equal(t, "/base/foo%2Fbar", req.URL.RawPath)
	}))
	t.Cleanup(server.Close)

	builder := NewProxyBuilder(&transportManagerMock{}, static.FastProxyConfig{})

	serverURL, err := url.JoinPath(server.URL, "base")
	require.NoError(t, err)

	proxyHandler, err := builder.Build("", testhelpers.MustParseURL(serverURL), true, true)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/foo%2Fbar", http.NoBody)
	res := httptest.NewRecorder()

	proxyHandler.ServeHTTP(res, req)

	assert.Equal(t, 1, callCount)
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestOpaqueRequestURL(t *testing.T) {
	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		callCount++
		assert.Equal(t, "/", req.RequestURI)
	}))
	t.Cleanup(server.Close)

	builder := NewProxyBuilder(&transportManagerMock{}, static.FastProxyConfig{})

	proxyHandler, err := builder.Build("", testhelpers.MustParseURL(server.URL), true, false)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "http:evil.example.com/admin", http.NoBody)
	res := httptest.NewRecorder()

	proxyHandler.ServeHTTP(res, req)

	assert.Equal(t, 1, callCount)
	assert.Equal(t, http.StatusOK, res.Code)
}

// TestRequestHeaderSanitization checks that CR and LF in request header values do not reach the upstream.
// fasthttp replaces them with a space since v1.73.0 (see https://github.com/traefik/traefik/issues/13624).
func TestRequestHeaderSanitization(t *testing.T) {
	testCases := []struct {
		desc     string
		values   []string
		expected []string
	}{
		{
			desc:     "carriage return and line feed",
			values:   []string{"benign\r\nX-Injected: admin"},
			expected: []string{"benign  X-Injected: admin"},
		},
		{
			desc:     "carriage return",
			values:   []string{"before\rafter"},
			expected: []string{"before after"},
		},
		{
			desc:     "line feed",
			values:   []string{"before\nafter"},
			expected: []string{"before after"},
		},
		{
			desc:     "multiple values",
			values:   []string{"first\r\nX-Injected: admin", "second"},
			expected: []string{"first  X-Injected: admin", "second"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var (
				receivedValues []string
				injectedValue  string
			)
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				receivedValues = req.Header.Values("X-Tenant-Note")
				injectedValue = req.Header.Get("X-Injected")
			}))
			t.Cleanup(server.Close)

			builder := NewProxyBuilder(&transportManagerMock{}, static.FastProxyConfig{})

			proxyHandler, err := builder.Build("", testhelpers.MustParseURL(server.URL), true, true)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			req.Header["X-Tenant-Note"] = test.values
			res := httptest.NewRecorder()

			proxyHandler.ServeHTTP(res, req)

			assert.Equal(t, http.StatusOK, res.Code)
			assert.Equal(t, test.expected, receivedValues)
			assert.Empty(t, injectedValue)
		})
	}
}

func TestHeadRequest(t *testing.T) {
	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		callCount++

		assert.Equal(t, http.MethodHead, req.Method)

		rw.Header().Set("Content-Length", "42")
	}))
	t.Cleanup(server.Close)

	builder := NewProxyBuilder(&transportManagerMock{}, static.FastProxyConfig{})

	serverURL, err := url.JoinPath(server.URL)
	require.NoError(t, err)

	proxyHandler, err := builder.Build("", testhelpers.MustParseURL(serverURL), true, true)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodHead, "/", http.NoBody)
	res := httptest.NewRecorder()

	proxyHandler.ServeHTTP(res, req)

	assert.Equal(t, 1, callCount)
	assert.Equal(t, http.StatusOK, res.Code)
}

func TestInvalidStatusCode(t *testing.T) {
	testCases := []struct {
		desc          string
		rawStatusLine string
	}{
		{
			desc:          "status code above 999",
			rawStatusLine: "HTTP/1.1 99999 X",
		},
		{
			desc:          "status code below 100",
			rawStatusLine: "HTTP/1.1 50 X",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			backendListener, err := net.Listen("tcp", ":0")
			require.NoError(t, err)

			t.Cleanup(func() {
				_ = backendListener.Close()
			})

			go func() {
				conn, err := backendListener.Accept()
				if err != nil {
					return
				}

				_, _ = conn.Write([]byte(test.rawStatusLine + "\r\n\r\n"))
			}()

			builder := NewProxyBuilder(&transportManagerMock{}, static.FastProxyConfig{})

			serverURL := "http://" + backendListener.Addr().String()

			proxyHandler, err := builder.Build("", testhelpers.MustParseURL(serverURL), true, true)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			res := httptest.NewRecorder()

			proxyHandler.ServeHTTP(res, req)

			assert.Equal(t, http.StatusInternalServerError, res.Code)
		})
	}
}

// TestReadLoopExitBeforeHandoff checks that a request does not block forever when the backend closes the connection
// after the request is written, but before the response writer is passed to the connection readLoop (handoff).
func TestReadLoopExitBeforeHandoff(t *testing.T) {
	testCases := []struct {
		desc             string
		method           string
		body             io.Reader
		expectedStatus   int
		expectedRequests int32
	}{
		{
			desc:             "replayable request is retried on a new connection",
			method:           http.MethodGet,
			body:             http.NoBody,
			expectedStatus:   http.StatusOK,
			expectedRequests: 2,
		},
		{
			desc:             "non-replayable request fails",
			method:           http.MethodPost,
			body:             strings.NewReader("data"),
			expectedStatus:   http.StatusInternalServerError,
			expectedRequests: 1,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			// The backend drops the first request by closing its connection without an answer, and answers the following ones.
			var requests atomic.Int32
			backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if requests.Add(1) > 1 {
					return
				}

				if conn, _, err := rw.(http.Hijacker).Hijack(); err == nil {
					_ = conn.Close()
				}
			}))
			t.Cleanup(backend.Close)

			pool := newConnPool(1, 0, 0, func() (net.Conn, error) {
				return net.Dial("tcp", backend.Listener.Addr().String())
			})
			t.Cleanup(pool.Close)

			// The first connection is dialed ahead, so the test can wait for its readLoop to exit.
			co, err := pool.AcquireConn()
			require.NoError(t, err)
			pool.ReleaseConn(co)

			proxy, err := NewReverseProxy(testhelpers.MustParseURL(backend.URL), nil, false, true, false, pool)
			require.NoError(t, err)

			// The request goroutine is held after writing the request until the readLoop has exited,
			// so the response writer is always passed to a connection whose readLoop is gone.
			trace := &httptrace.ClientTrace{
				WroteRequest: func(httptrace.WroteRequestInfo) { <-co.DoneCh },
			}
			req := httptest.NewRequest(test.method, "/", test.body).WithContext(httptrace.WithClientTrace(t.Context(), trace))
			res := httptest.NewRecorder()

			done := make(chan struct{})
			go func() {
				defer close(done)
				proxy.ServeHTTP(res, req)
			}()

			select {
			case <-done:
			case <-time.After(10 * time.Second):
				t.Fatal("request blocked waiting for the readLoop")
			}

			assert.Equal(t, test.expectedStatus, res.Code)
			assert.Equal(t, test.expectedRequests, requests.Load())
		})
	}
}

// TestReplayOnReusedConnOnly checks that a request is replayed once on a connection that comes from the pool,
// and not again on the fresh connection of the replay, so a backend that closes every connection does not get
// the request in a loop.
func TestReplayOnReusedConnOnly(t *testing.T) {
	// The backend drops every request by closing its connection without an answer.
	var requests atomic.Int32
	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		if conn, _, err := rw.(http.Hijacker).Hijack(); err == nil {
			_ = conn.Close()
		}
	}))
	t.Cleanup(backend.Close)

	// Every connection signals its close, so the test can wait for its readLoop to exit.
	closed := make(chan struct{}, 100)
	pool := newConnPool(1, 0, 0, func() (net.Conn, error) {
		conn, err := net.Dial("tcp", backend.Listener.Addr().String())
		if err != nil {
			return nil, err
		}
		return closeNotifyingConn{Conn: conn, closed: closed}, nil
	})
	t.Cleanup(pool.Close)

	// The first connection comes from the pool, the replay dials a fresh one.
	co, err := pool.AcquireConn()
	require.NoError(t, err)
	pool.ReleaseConn(co)

	proxy, err := NewReverseProxy(testhelpers.MustParseURL(backend.URL), nil, false, true, false, pool)
	require.NoError(t, err)

	// The request goroutine is held after writing the request until the readLoop has exited,
	// so the response writer is always passed to a connection whose readLoop is gone.
	trace := &httptrace.ClientTrace{
		WroteRequest: func(httptrace.WroteRequestInfo) { <-closed },
	}
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody).WithContext(httptrace.WithClientTrace(t.Context(), trace))
	res := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		defer close(done)
		proxy.ServeHTTP(res, req)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("request replayed in a loop")
	}

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, int32(2), requests.Load())
}

// TestReplayOnReusedConnOnlyOnWriteError checks that a request whose write fails is replayed once on a
// connection that comes from the pool, and not again on the fresh connection of the replay.
func TestReplayOnReusedConnOnlyOnWriteError(t *testing.T) {
	// The backend is never reached, every connection fails its writes.
	backend := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(backend.Close)

	var dials atomic.Int32
	pool := newConnPool(1, 0, 0, func() (net.Conn, error) {
		dials.Add(1)
		conn, err := net.Dial("tcp", backend.Listener.Addr().String())
		if err != nil {
			return nil, err
		}
		return failingWriteConn{Conn: conn}, nil
	})
	t.Cleanup(pool.Close)

	// The first connection comes from the pool, the replay dials a fresh one.
	co, err := pool.AcquireConn()
	require.NoError(t, err)
	pool.ReleaseConn(co)

	proxy, err := NewReverseProxy(testhelpers.MustParseURL(backend.URL), nil, false, true, false, pool)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	res := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		defer close(done)
		proxy.ServeHTTP(res, req)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("request replayed in a loop")
	}

	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, int32(2), dials.Load())
}

// TestResponseHeaderTimeout checks that the response header timeout applies when the backend never answers,
// not only once the first response byte has arrived.
func TestResponseHeaderTimeout(t *testing.T) {
	release := make(chan struct{})
	backend := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		<-release
	}))
	t.Cleanup(backend.Close)
	t.Cleanup(func() { close(release) })

	pool := newConnPool(1, 0, 200*time.Millisecond, func() (net.Conn, error) {
		return net.Dial("tcp", backend.Listener.Addr().String())
	})
	t.Cleanup(pool.Close)

	proxy, err := NewReverseProxy(testhelpers.MustParseURL(backend.URL), nil, false, true, false, pool)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	res := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		defer close(done)
		proxy.ServeHTTP(res, req)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("request blocked waiting for the response headers")
	}

	assert.Equal(t, http.StatusGatewayTimeout, res.Code)
}

func TestNoContentLength(t *testing.T) {
	backendListener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = backendListener.Close()
	})

	go func() {
		t.Helper()

		conn, err := backendListener.Accept()
		require.NoError(t, err)

		_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nfoo"))
		require.NoError(t, err)

		// CloseWrite the connection to signal the end of the response.
		if v, ok := conn.(interface{ CloseWrite() error }); ok {
			err = v.CloseWrite()
			require.NoError(t, err)
		}
	}()

	builder := NewProxyBuilder(&transportManagerMock{}, static.FastProxyConfig{})

	serverURL := "http://" + backendListener.Addr().String()

	proxyHandler, err := builder.Build("", testhelpers.MustParseURL(serverURL), true, true)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	res := httptest.NewRecorder()

	proxyHandler.ServeHTTP(res, req)

	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, "foo", res.Body.String())
}

func TestTransferEncodingChunked(t *testing.T) {
	backendServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		flusher, ok := rw.(http.Flusher)
		require.True(t, ok)

		for i := range 3 {
			_, err := fmt.Fprintf(rw, "chunk %d\n", i)
			require.NoError(t, err)

			flusher.Flush()
		}
	}))
	t.Cleanup(backendServer.Close)

	builder := NewProxyBuilder(&transportManagerMock{}, static.FastProxyConfig{})

	proxyHandler, err := builder.Build("", testhelpers.MustParseURL(backendServer.URL), true, true)
	require.NoError(t, err)

	proxyServer := httptest.NewServer(proxyHandler)
	t.Cleanup(proxyServer.Close)

	req, err := http.NewRequest(http.MethodGet, proxyServer.URL, http.NoBody)
	require.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	t.Cleanup(func() { _ = res.Body.Close() })

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, []string{"chunked"}, res.TransferEncoding)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	assert.Equal(t, "chunk 0\nchunk 1\nchunk 2\n", string(body))
}

func TestConnectRequest(t *testing.T) {
	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		callCount++
	}))
	t.Cleanup(server.Close)

	builder := NewProxyBuilder(&transportManagerMock{}, static.FastProxyConfig{})

	serverURL, err := url.JoinPath(server.URL)
	require.NoError(t, err)

	proxyHandler, err := builder.Build("", testhelpers.MustParseURL(serverURL), true, true)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodConnect, "/", http.NoBody)
	res := httptest.NewRecorder()

	proxyHandler.ServeHTTP(res, req)

	assert.Equal(t, 0, callCount)
	assert.Equal(t, http.StatusNotImplemented, res.Code)
}

func TestXForwardedFor(t *testing.T) {
	testCases := []struct {
		desc                  string
		notAppendXFF          bool
		incomingXFF           string
		expectedXFF           string
		expectedXFFNotPresent bool
	}{
		{
			desc:         "appends RemoteAddr when notAppendXFF is false",
			notAppendXFF: false,
			incomingXFF:  "",
			expectedXFF:  "192.0.2.1",
		},
		{
			desc:         "appends RemoteAddr to existing XFF when notAppendXFF is false",
			notAppendXFF: false,
			incomingXFF:  "203.0.113.1",
			expectedXFF:  "203.0.113.1, 192.0.2.1",
		},
		{
			desc:                  "does not append RemoteAddr when notAppendXFF is true and no incoming XFF",
			notAppendXFF:          true,
			incomingXFF:           "",
			expectedXFFNotPresent: true,
		},
		{
			desc:         "preserves existing XFF when notAppendXFF is true",
			notAppendXFF: true,
			incomingXFF:  "203.0.113.1",
			expectedXFF:  "203.0.113.1",
		},
		{
			desc:         "preserves multiple XFF values when notAppendXFF is true",
			notAppendXFF: true,
			incomingXFF:  "203.0.113.1, 198.51.100.1",
			expectedXFF:  "203.0.113.1, 198.51.100.1",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			var receivedXFF string
			var xffPresent bool

			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				receivedXFF = req.Header.Get("X-Forwarded-For")
				xffPresent = req.Header.Get("X-Forwarded-For") != "" || len(req.Header["X-Forwarded-For"]) > 0
				rw.WriteHeader(http.StatusOK)
			}))
			t.Cleanup(server.Close)

			builder := NewProxyBuilder(&transportManagerMock{}, static.FastProxyConfig{})

			proxyHandler, err := builder.Build("", testhelpers.MustParseURL(server.URL), true, false)
			require.NoError(t, err)

			ctx := t.Context()
			if test.notAppendXFF {
				ctx = proxyhttputil.SetNotAppendXFF(ctx)
			}

			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			req = req.WithContext(ctx)
			req.RemoteAddr = "192.0.2.1:12345"

			if test.incomingXFF != "" {
				req.Header.Set("X-Forwarded-For", test.incomingXFF)
			}

			res := httptest.NewRecorder()
			proxyHandler.ServeHTTP(res, req)

			assert.Equal(t, http.StatusOK, res.Code)

			if test.expectedXFFNotPresent {
				assert.False(t, xffPresent, "X-Forwarded-For header should not be present")
			} else {
				assert.Equal(t, test.expectedXFF, receivedXFF)
			}
		})
	}
}

// TestStickyConnPool ensures that a connection-bound NTLM/Negotiate authentication is isolated to the frontend
// connection that carried the credential, and never leaks the authenticated backend connection to another one.
func TestStickyConnPool(t *testing.T) {
	const backendConnHeader = "X-Backend-Conn"

	// RemoteAddr identifies the backend connection, so the test can observe which backend connection serves a request.
	backend := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set(backendConnHeader, req.RemoteAddr)
	}))
	t.Cleanup(backend.Close)

	// Requests are issued serially, so at most one connection is ever in flight and a pool never holds more than
	// one idle connection: a single idle slot makes the reused backend connection deterministic to compare.
	builder := NewProxyBuilder(&transportManagerMock{serversTransport: &dynamic.ServersTransport{MaxIdleConnsPerHost: 1}}, static.FastProxyConfig{})

	proxyHandler, err := builder.Build("default@internal", testhelpers.MustParseURL(backend.URL), false, false)
	require.NoError(t, err)

	// Each context stands for a distinct frontend connection, as the entrypoint would set it up.
	backendConnFor := func(ctx context.Context, header, credential string) string {
		t.Helper()

		req := httptest.NewRequest(http.MethodGet, "/", http.NoBody).WithContext(ctx)
		if credential != "" {
			req.Header.Set(header, credential)
		}

		res := httptest.NewRecorder()
		proxyHandler.ServeHTTP(res, req)
		require.Equal(t, http.StatusOK, res.Code)

		return res.Header().Get(backendConnHeader)
	}

	firstConn := AddConnPoolsOnContext(context.Background())
	secondConn := AddConnPoolsOnContext(context.Background())
	thirdConn := AddConnPoolsOnContext(context.Background())

	// Credential-less requests are served from the shared pool, so they reuse the same backend connection.
	sharedConn := backendConnFor(firstConn, "", "")
	assert.Equal(t, sharedConn, backendConnFor(secondConn, "", ""))

	// A request carrying an NTLM credential is moved onto a pool dedicated to its frontend connection.
	firstStickyConn := backendConnFor(firstConn, "Authorization", "NTLM TlRMTVNTUAAB")
	assert.NotEqual(t, sharedConn, firstStickyConn)
	// That dedicated connection is then reused even by a later credential-less request, which is what a
	// connection-bound authentication relies on.
	assert.Equal(t, firstStickyConn, backendConnFor(firstConn, "", ""))

	// A client must never be served by the connection another one authenticated on.
	assert.NotEqual(t, firstStickyConn, backendConnFor(secondConn, "", ""))
	// Authenticated clients each get their own dedicated connection, never shared with one another.
	assert.NotEqual(t, firstStickyConn, backendConnFor(secondConn, "Authorization", "NTLM TlRMTVNTUAAB"))

	// A credential carried in Proxy-Authorization binds the connection the same way.
	thirdStickyConn := backendConnFor(thirdConn, "Proxy-Authorization", "NTLM TlRMTVNTUAAB")
	assert.NotEqual(t, sharedConn, thirdStickyConn)
	assert.Equal(t, thirdStickyConn, backendConnFor(thirdConn, "", ""))
}

func TestReleaseConnPoolsOnContext(t *testing.T) {
	var backendConns atomic.Int32

	backend := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	backend.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			backendConns.Add(1)
		case http.StateClosed, http.StateHijacked:
			backendConns.Add(-1)
		}
	}
	backend.Start()
	t.Cleanup(backend.Close)

	builder := NewProxyBuilder(&transportManagerMock{serversTransport: &dynamic.ServersTransport{MaxIdleConnsPerHost: 1}}, static.FastProxyConfig{})

	proxyHandler, err := builder.Build("default@internal", testhelpers.MustParseURL(backend.URL), false, false)
	require.NoError(t, err)

	ctx := AddConnPoolsOnContext(t.Context())

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody).WithContext(ctx)
	req.Header.Set("Authorization", "NTLM TlRMTVNTUAAB")

	res := httptest.NewRecorder()
	proxyHandler.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)

	require.EqualValues(t, 1, backendConns.Load())

	ReleaseConnPoolsOnContext(ctx)

	assert.Eventually(t, func() bool { return backendConns.Load() == 0 }, time.Second, 10*time.Millisecond)
}

func TestReleaseConnPoolsOnContextWithoutConnPools(t *testing.T) {
	assert.NotPanics(t, func() {
		ReleaseConnPoolsOnContext(t.Context())
		ReleaseConnPoolsOnContext(AddConnPoolsOnContext(t.Context()))
	})
}

// closeNotifyingConn signals on closed once the connection is closed.
type closeNotifyingConn struct {
	net.Conn

	closed chan<- struct{}
}

func (c closeNotifyingConn) Close() error {
	err := c.Conn.Close()
	c.closed <- struct{}{}
	return err
}

type transportManagerMock struct {
	tlsConfig        *tls.Config
	serversTransport *dynamic.ServersTransport
}

func (r *transportManagerMock) GetTLSConfig(_ string) (*tls.Config, error) {
	return r.tlsConfig, nil
}

func (r *transportManagerMock) Get(_ string) (*dynamic.ServersTransport, error) {
	if r.serversTransport != nil {
		return r.serversTransport, nil
	}
	return &dynamic.ServersTransport{}, nil
}

// failingWriteConn fails every write.
type failingWriteConn struct {
	net.Conn
}

func (c failingWriteConn) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}
