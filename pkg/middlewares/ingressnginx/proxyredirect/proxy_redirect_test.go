package proxyredirect

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestProxyRedirect(t *testing.T) {
	testCases := []struct {
		desc            string
		config          dynamic.ProxyRedirect
		reqHost         string
		reqTLS          bool
		respHeaders     map[string]string
		expectedHeaders map[string]string
	}{
		{
			desc: "replaces location prefix with target URL",
			config: dynamic.ProxyRedirect{
				From: "http://backend:8080/",
				To:   "https://example.com/",
			},
			reqHost: "example.com",
			respHeaders: map[string]string{
				"Location": "http://backend:8080/dashboard",
			},
			expectedHeaders: map[string]string{
				"Location": "https://example.com/dashboard",
			},
		},
		{
			desc: "replaces scheme prefix",
			config: dynamic.ProxyRedirect{
				From: "http://",
				To:   "https://",
			},
			reqHost: "example.com",
			respHeaders: map[string]string{
				"Location": "http://example.com/dashboard",
			},
			expectedHeaders: map[string]string{
				"Location": "https://example.com/dashboard",
			},
		},
		{
			desc: "relative target with absolute source inserts server name and port",
			config: dynamic.ProxyRedirect{
				From: "http://backend:8080/two/",
				To:   "/one/",
			},
			reqHost: "example.com:9000",
			respHeaders: map[string]string{
				"Location": "http://backend:8080/two/some/uri",
			},
			expectedHeaders: map[string]string{
				"Location": "http://example.com:9000/one/some/uri",
			},
		},
		{
			desc: "default target uses scheme and best_http_host",
			config: dynamic.ProxyRedirect{
				From: "http://backend:8080/",
				To:   "default",
			},
			reqHost: "example.com",
			reqTLS:  true,
			respHeaders: map[string]string{
				"Location": "http://backend:8080/login",
			},
			expectedHeaders: map[string]string{
				"Location": "https://example.com/login",
			},
		},
		{
			desc: "empty target defaults to scheme and best_http_host",
			config: dynamic.ProxyRedirect{
				From: "http://backend:8080/",
			},
			reqHost: "example.com",
			respHeaders: map[string]string{
				"Location": "http://backend:8080/login",
			},
			expectedHeaders: map[string]string{
				"Location": "http://example.com/login",
			},
		},
		{
			desc: "variable interpolation in target",
			config: dynamic.ProxyRedirect{
				From: "http://backend:8080/",
				To:   "$scheme://$best_http_host/app/",
			},
			reqHost: "example.com:8443",
			reqTLS:  true,
			respHeaders: map[string]string{
				"Location": "http://backend:8080/home",
			},
			expectedHeaders: map[string]string{
				"Location": "https://example.com:8443/app/home",
			},
		},
		{
			desc: "case-sensitive regular expression replacement",
			config: dynamic.ProxyRedirect{
				From: `~^http://backend:\d+(/.*)$`,
				To:   "https://example.com$1",
			},
			reqHost: "example.com",
			respHeaders: map[string]string{
				"Location": "http://backend:8080/test/path",
			},
			expectedHeaders: map[string]string{
				"Location": "https://example.com/test/path",
			},
		},
		{
			desc: "case-insensitive regular expression replacement",
			config: dynamic.ProxyRedirect{
				From: `~*^http://BACKEND:\d+(/.*)$`,
				To:   "https://example.com$1",
			},
			reqHost: "example.com",
			respHeaders: map[string]string{
				"Location": "http://backend:8080/test/path",
			},
			expectedHeaders: map[string]string{
				"Location": "https://example.com/test/path",
			},
		},
		{
			desc: "case-sensitive regular expression replacement with relative target does not insert host",
			config: dynamic.ProxyRedirect{
				From: `~^http://backend:\d+(/.*)$`,
				To:   "/prefix$1",
			},
			reqHost: "example.com",
			respHeaders: map[string]string{
				"Location": "http://backend:8080/test/path",
			},
			expectedHeaders: map[string]string{
				"Location": "/prefix/test/path",
			},
		},
		{
			desc: "rewrites Refresh header",
			config: dynamic.ProxyRedirect{
				From: "http://backend:8080/",
				To:   "https://example.com/",
			},
			reqHost: "example.com",
			respHeaders: map[string]string{
				"Refresh": "5; url=http://backend:8080/next",
			},
			expectedHeaders: map[string]string{
				"Refresh": "5; url=https://example.com/next",
			},
		},
		{
			desc: "rewrites quoted Refresh header",
			config: dynamic.ProxyRedirect{
				From: "http://backend:8080/",
				To:   "https://example.com/",
			},
			reqHost: "example.com",
			respHeaders: map[string]string{
				"Refresh": `0; url="http://backend:8080/next"`,
			},
			expectedHeaders: map[string]string{
				"Refresh": `0; url="https://example.com/next"`,
			},
		},
		{
			desc: "does not rewrite non-matching location",
			config: dynamic.ProxyRedirect{
				From: "http://backend:8080/",
				To:   "https://example.com/",
			},
			reqHost: "example.com",
			respHeaders: map[string]string{
				"Location": "http://other-service:9090/path",
			},
			expectedHeaders: map[string]string{
				"Location": "http://other-service:9090/path",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				for k, v := range tc.respHeaders {
					rw.Header().Set(k, v)
				}
				rw.WriteHeader(http.StatusFound)
			})

			handler, err := New(context.Background(), next, tc.config, "test-proxy-redirect")
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "http://"+tc.reqHost+"/", nil)
			req.Host = tc.reqHost
			if tc.reqTLS {
				req.TLS = &tls.ConnectionState{}
			}

			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			for k, expectedVal := range tc.expectedHeaders {
				assert.Equal(t, expectedVal, recorder.Header().Get(k))
			}
		})
	}
}

func TestProxyRedirectValidation(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})

	_, err := New(context.Background(), next, dynamic.ProxyRedirect{From: ""}, "test")
	require.Error(t, err)

	_, err = New(context.Background(), next, dynamic.ProxyRedirect{From: "~["}, "test")
	require.Error(t, err)

	_, err = New(context.Background(), next, dynamic.ProxyRedirect{From: "~*["}, "test")
	require.Error(t, err)
}
