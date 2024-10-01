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
	"github.com/traefik/traefik/v3/pkg/tls/generate"
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
			backendURL, backendCert := newBackendServer(t, test.tls, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				_, _ = rw.Write([]byte("backend"))
			}))

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

				go func() { _, _ = io.Copy(connHj, conn) }()
				_, _ = io.Copy(conn, connHj)
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
				proxyServer := httptest.NewServer(proxyHandler)
				t.Cleanup(proxyServer.Close)

				proxyURL = proxyServer.URL
				proxyCert = proxyServer.Certificate()
			}

			certPool := x509.NewCertPool()
			if proxyCert != nil {
				certPool.AddCert(proxyCert)
			}
			if backendCert != nil {
				cert, err := x509.ParseCertificate(backendCert.Certificate[0])
				require.NoError(t, err)

				certPool.AddCert(cert)
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

			reverseProxy, err := builder.Build("foo", testhelpers.MustParseURL(backendURL), false)
			require.NoError(t, err)

			reverseProxyServer := httptest.NewServer(reverseProxy)
			t.Cleanup(reverseProxyServer.Close)

			client := http.Client{Timeout: 5 * time.Second}

			resp, err := client.Get(reverseProxyServer.URL)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, "backend", string(body))
			assert.True(t, proxyCalled)
		})
	}
}

func newCertificate(t *testing.T, domain string) *tls.Certificate {
	t.Helper()

	certPEM, keyPEM, err := generate.KeyPair(domain, time.Time{})
	require.NoError(t, err)

	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)

	return &certificate
}

func newBackendServer(t *testing.T, isTLS bool, handler http.Handler) (string, *tls.Certificate) {
	t.Helper()

	var ln net.Listener
	var err error
	var cert *tls.Certificate

	scheme := "http"
	domain := "backend.localhost"
	if isTLS {
		scheme = "https"

		cert = newCertificate(t, domain)

		ln, err = tls.Listen("tcp", ":0", &tls.Config{Certificates: []tls.Certificate{*cert}})
		require.NoError(t, err)
	} else {
		ln, err = net.Listen("tcp", ":0")
		require.NoError(t, err)
	}

	srv := &http.Server{Handler: handler}
	go func() { _ = srv.Serve(ln) }()

	t.Cleanup(func() { _ = srv.Close() })

	_, port, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)

	backendURL := fmt.Sprintf("%s://%s:%s", scheme, domain, port)

	return backendURL, cert
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
