package proxy

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
	"github.com/traefik/traefik/v2/pkg/tls/client"
)

func Int32(i int32) *int32 {
	return &i
}

// LocalhostCert is a PEM-encoded TLS cert
// for host example.com, www.example.com
// expiring at Jan 29 16:00:00 2084 GMT.
// go run $GOROOT/src/crypto/tls/generate_cert.go  --rsa-bits 1024 --host example.com,www.example.com --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
var LocalhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIICDDCCAXWgAwIBAgIQH20JmcOlcRWHNuf62SYwszANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMCAXDTcwMDEwMTAwMDAwMFoYDzIwODQwMTI5MTYw
MDAwWjASMRAwDgYDVQQKEwdBY21lIENvMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCB
iQKBgQC0qINy3F4oq6viDnlpDDE5J08iSRGggg6EylJKBKZfphEG2ufgK78Dufl3
+7b0LlEY2AeZHwviHODqC9a6ihj1ZYQk0/djAh+OeOhFEWu+9T/VP8gVFarFqT8D
Opy+hrG7YJivUIzwb4fmJQRI7FajzsnGyM6LiXLU+0qzb7ZO/QIDAQABo2EwXzAO
BgNVHQ8BAf8EBAMCAqQwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUw
AwEB/zAnBgNVHREEIDAeggtleGFtcGxlLmNvbYIPd3d3LmV4YW1wbGUuY29tMA0G
CSqGSIb3DQEBCwUAA4GBAB+eluoQYzyyMfeEEAOtlldevx5MtDENT05NB0WI+91R
we7mX8lv763u0XuCWPxbHszhclI6FFjoQef0Z1NYLRm8ZRq58QqWDFZ3E6wdDK+B
+OWvkW+hRavo6R9LzIZPfbv8yBo4M9PK/DXw8hLqH7VkkI+Gh793iH7Ugd4A7wvT
-----END CERTIFICATE-----`)

// LocalhostKey is the private key for localhostCert.
var LocalhostKey = []byte(`-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBALSog3LcXiirq+IO
eWkMMTknTyJJEaCCDoTKUkoEpl+mEQba5+ArvwO5+Xf7tvQuURjYB5kfC+Ic4OoL
1rqKGPVlhCTT92MCH4546EURa771P9U/yBUVqsWpPwM6nL6GsbtgmK9QjPBvh+Yl
BEjsVqPOycbIzouJctT7SrNvtk79AgMBAAECgYB1wMT1MBgbkFIXpXGTfAP1id61
rUTVBxCpkypx3ngHLjo46qRq5Hi72BN4FlTY8fugIudI8giP2FztkMvkiLDc4m0p
Gn+QMJzjlBjjTuNLvLy4aSmNRLIC3mtbx9PdU71DQswEpJHFj/vmsxbuSrG1I1YE
r1reuSo2ow6fOAjXLQJBANpz+RkOiPSPuvl+gi1sp2pLuynUJVDVqWZi386YRpfg
DiKCLpqwqYDkOozm/fwFALvwXKGmsyyL43HO8eI+2NsCQQDTtY32V+02GPecdsyq
msK06EPVTSaYwj9Mm+q709KsmYFHLXDqXjcKV4UgKYKRPz7my1fXodMmGmfuh1a3
/HMHAkEAmOQKN0tA90mRJwUvvvMIyRBv0fq0kzq28P3KfiF9ZtZdjjFmxMVYHOmf
QPZ6VGR7+w1jB5BQXqEZcpHQIPSzeQJBAIy9tZJ/AYNlNbcegxEnsSjy/6VdlLsY
51vWi0Yym2uC4R6gZuBnoc+OP0ISVmqY0Qg9RjhjrCs4gr9f2ZaWjSECQCxqZMq1
3viJ8BGCC0m/5jv1EHur3YgwphYCkf4Li6DKwIdMLk1WXkTcPIY3V2Jqj8rPEB5V
rqPRSAtd/h6oZbs=
-----END PRIVATE KEY-----`)

func TestKeepConnectionWhenSameConfiguration(t *testing.T) {
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	connCount := Int32(0)
	srv.Config.ConnState = func(conn net.Conn, state http.ConnState) {
		if state == http.StateNew {
			atomic.AddInt32(connCount, 1)
		}
	}

	cert, err := tls.X509KeyPair(LocalhostCert, LocalhostKey)
	require.NoError(t, err)

	srv.TLS = &tls.Config{Certificates: []tls.Certificate{cert}}
	srv.StartTLS()

	tlsClientConfigGetter := client.NewTLSConfigManager(nil)
	proxyBuilder := NewBuilder(tlsClientConfigGetter)

	dynamicConf := map[string]*dynamic.ServersTransport{
		"test": {
			HTTP: &dynamic.HTTPClientConfig{EnableHTTP2: true},
			TLS: &dynamic.TLSClientConfig{
				ServerName: "example.com",
				RootCAs:    []traefiktls.FileOrContent{traefiktls.FileOrContent(LocalhostCert)},
			},
		},
	}

	for i := 0; i < 10; i++ {
		tlsClientConfigGetter.Update(dynamicConf)
		proxyBuilder.Update(dynamicConf)

		proxy, err := proxyBuilder.Build("test", testhelpers.MustParseURL(srv.URL))
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, srv.URL, http.NoBody)
		req.URL.Scheme = "h2c"
		rw := httptest.NewRecorder()
		proxy.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusOK, rw.Result().StatusCode)
	}

	count := atomic.LoadInt32(connCount)
	require.EqualValues(t, 1, count)

	dynamicConf = map[string]*dynamic.ServersTransport{
		"test": {
			HTTP: &dynamic.HTTPClientConfig{},
			TLS: &dynamic.TLSClientConfig{
				ServerName: "www.example.com",
				RootCAs:    []traefiktls.FileOrContent{traefiktls.FileOrContent(LocalhostCert)},
			},
		},
	}

	tlsClientConfigGetter.Update(dynamicConf)
	proxyBuilder.Update(dynamicConf)

	proxy, err := proxyBuilder.Build("test", testhelpers.MustParseURL(srv.URL))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, srv.URL, http.NoBody)
	rw := httptest.NewRecorder()
	proxy.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Result().StatusCode)

	count = atomic.LoadInt32(connCount)
	assert.EqualValues(t, 2, count)
}

func TestDisableHTTP2(t *testing.T) {
	testCases := []struct {
		desc          string
		disableHTTP2  bool
		serverHTTP2   bool
		expectedProto string
	}{
		{
			desc:          "HTTP1 capable client with HTTP1 server",
			disableHTTP2:  true,
			expectedProto: "HTTP/1.1",
		},
		{
			desc:          "HTTP1 capable client with HTTP2 server",
			disableHTTP2:  true,
			serverHTTP2:   true,
			expectedProto: "HTTP/1.1",
		},
		{
			desc:          "HTTP2 capable client with HTTP1 server",
			expectedProto: "HTTP/1.1",
		},
		{
			desc:          "HTTP2 capable client with HTTP2 server",
			serverHTTP2:   true,
			expectedProto: "HTTP/2.0",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var gotProto string
			srv := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				gotProto = req.Proto
				rw.WriteHeader(http.StatusOK)
			}))

			srv.EnableHTTP2 = test.serverHTTP2
			srv.StartTLS()

			tlsClientConfigGetter := client.NewTLSConfigManager(nil)
			proxyBuilder := NewBuilder(tlsClientConfigGetter)

			dynamicConf := map[string]*dynamic.ServersTransport{
				"test": {
					HTTP: &dynamic.HTTPClientConfig{
						EnableHTTP2: !test.disableHTTP2,
					},
					TLS: &dynamic.TLSClientConfig{
						InsecureSkipVerify: true,
					},
				},
			}

			tlsClientConfigGetter.Update(dynamicConf)
			proxyBuilder.Update(dynamicConf)

			proxy, err := proxyBuilder.Build("test", testhelpers.MustParseURL(srv.URL))
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, srv.URL, http.NoBody)
			rw := httptest.NewRecorder()
			proxy.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusOK, rw.Result().StatusCode)
			// Proxy doesn't keep the proto that's why it is verified directly in the backend server.
			assert.Equal(t, test.expectedProto, gotProto)
		})
	}
}
