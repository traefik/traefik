package fast

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"

	"github.com/armon/go-socks5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/static"
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

type transportManagerMock struct {
	tlsConfig *tls.Config
}

func (r *transportManagerMock) GetTLSConfig(_ string) (*tls.Config, error) {
	return r.tlsConfig, nil
}

func (r *transportManagerMock) Get(_ string) (*dynamic.ServersTransport, error) {
	return &dynamic.ServersTransport{}, nil
}
