package proxy

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/proxy/fasthttp"
	"github.com/traefik/traefik/v2/pkg/proxy/httputil"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
)

func Test_PassHostHeader(t *testing.T) {
	testCases := []struct {
		desc         string
		cfg          dynamic.HTTPClientConfig
		proxyBuilder func(*testing.T, *url.URL, *dynamic.HTTPClientConfig) http.Handler
	}{
		{
			desc:         "FastHTTP proxy with passHostHeader",
			proxyBuilder: buildFastHTTPProxy,
			cfg: dynamic.HTTPClientConfig{
				PassHostHeader: true,
			},
		},
		{
			desc:         "FastHTTP proxy without passHostHeader",
			proxyBuilder: buildFastHTTPProxy,
			cfg: dynamic.HTTPClientConfig{
				PassHostHeader: false,
			},
		},
		{
			desc:         "HTTPUtil proxy with passHostHeader",
			proxyBuilder: buildHTTPProxy,
			cfg: dynamic.HTTPClientConfig{
				PassHostHeader: true,
			},
		},
		{
			desc:         "HTTPUtil proxy without passHostHeader",
			proxyBuilder: buildHTTPProxy,
			cfg: dynamic.HTTPClientConfig{
				PassHostHeader: false,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var gotHostHeader string
			backendServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				gotHostHeader = req.Host
			}))

			u := testhelpers.MustParseURL(backendServer.URL)
			handler := test.proxyBuilder(t, u, &test.cfg)

			proxyServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				handler.ServeHTTP(rw, req)
			}))
			t.Cleanup(proxyServer.Close)

			_, err := http.Get(proxyServer.URL)
			require.NoError(t, err)

			target := testhelpers.MustParseURL(proxyServer.URL)
			if !test.cfg.PassHostHeader {
				target = testhelpers.MustParseURL(backendServer.URL)
			}
			assert.Equal(t, target.Host, gotHostHeader)
		})
	}
}

func Test_EscapedPath(t *testing.T) {
	testCases := []struct {
		desc         string
		proxyBuilder func(*testing.T, *url.URL, *dynamic.HTTPClientConfig) http.Handler
		cfg          dynamic.HTTPClientConfig
	}{
		{
			desc:         "FastHTTP proxy",
			proxyBuilder: buildFastHTTPProxy,
		},
		{
			desc:         "HTTPUtil proxy",
			proxyBuilder: buildHTTPProxy,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var gotEscapedPath string
			backendServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				gotEscapedPath = req.URL.EscapedPath()
			}))

			u := testhelpers.MustParseURL(backendServer.URL)
			h := test.proxyBuilder(t, u, &test.cfg)

			proxyServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				h.ServeHTTP(rw, req)
			}))
			t.Cleanup(proxyServer.Close)

			escapedPath := "/%3A%2F%2F"

			_, err := http.Get(proxyServer.URL + escapedPath)
			require.NoError(t, err)

			assert.Equal(t, escapedPath, gotEscapedPath)
		})
	}
}

func buildFastHTTPProxy(t *testing.T, u *url.URL, cfg *dynamic.HTTPClientConfig) http.Handler {
	t.Helper()

	f, err := fasthttp.NewReverseProxy(u, nil, cfg.PassHostHeader, 0, fasthttp.NewConnPool(200, 0, func() (net.Conn, error) {
		return net.Dial("tcp", u.Host)
	}))
	require.NoError(t, err)

	return f
}

func buildHTTPProxy(t *testing.T, u *url.URL, cfg *dynamic.HTTPClientConfig) http.Handler {
	t.Helper()

	f, err := httputil.NewProxyBuilder().Build("default", cfg, nil, u)
	require.NoError(t, err)

	return f
}
