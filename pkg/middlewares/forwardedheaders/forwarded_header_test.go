package forwardedheaders

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeHTTP(t *testing.T) {
	testCases := []struct {
		desc              string
		insecure          bool
		trustedIps        []string
		connectionHeaders []string
		incomingHeaders   map[string][]string
		remoteAddr        string
		expectedHeaders   map[string]string
		tls               bool
		websocket         bool
		host              string
	}{
		{
			desc:            "all Empty",
			insecure:        true,
			trustedIps:      nil,
			remoteAddr:      "",
			incomingHeaders: map[string][]string{},
			expectedHeaders: map[string]string{
				xForwardedFor:               "",
				xForwardedURI:               "",
				xForwardedMethod:            "",
				xForwardedTLSClientCert:     "",
				xForwardedTLSClientCertInfo: "",
			},
		},
		{
			desc:       "insecure true with incoming X-Forwarded headers",
			insecure:   true,
			trustedIps: nil,
			remoteAddr: "",
			incomingHeaders: map[string][]string{
				xForwardedFor:               {"10.0.1.0, 10.0.1.12"},
				xForwardedURI:               {"/bar"},
				xForwardedMethod:            {"GET"},
				xForwardedTLSClientCert:     {"Cert"},
				xForwardedTLSClientCertInfo: {"CertInfo"},
				xForwardedPrefix:            {"/prefix"},
			},
			expectedHeaders: map[string]string{
				xForwardedFor:               "10.0.1.0, 10.0.1.12",
				xForwardedURI:               "/bar",
				xForwardedMethod:            "GET",
				xForwardedTLSClientCert:     "Cert",
				xForwardedTLSClientCertInfo: "CertInfo",
				xForwardedPrefix:            "/prefix",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers",
			insecure:   false,
			trustedIps: nil,
			remoteAddr: "",
			incomingHeaders: map[string][]string{
				xForwardedFor:               {"10.0.1.0, 10.0.1.12"},
				xForwardedURI:               {"/bar"},
				xForwardedMethod:            {"GET"},
				xForwardedTLSClientCert:     {"Cert"},
				xForwardedTLSClientCertInfo: {"CertInfo"},
				xForwardedPrefix:            {"/prefix"},
			},
			expectedHeaders: map[string]string{
				xForwardedFor:               "",
				xForwardedURI:               "",
				xForwardedMethod:            "",
				xForwardedTLSClientCert:     "",
				xForwardedTLSClientCertInfo: "",
				xForwardedPrefix:            "",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers and valid Trusted Ips",
			insecure:   false,
			trustedIps: []string{"10.0.1.100"},
			remoteAddr: "10.0.1.100:80",
			incomingHeaders: map[string][]string{
				xForwardedFor:               {"10.0.1.0, 10.0.1.12"},
				xForwardedURI:               {"/bar"},
				xForwardedMethod:            {"GET"},
				xForwardedTLSClientCert:     {"Cert"},
				xForwardedTLSClientCertInfo: {"CertInfo"},
				xForwardedPrefix:            {"/prefix"},
			},
			expectedHeaders: map[string]string{
				xForwardedFor:               "10.0.1.0, 10.0.1.12",
				xForwardedURI:               "/bar",
				xForwardedMethod:            "GET",
				xForwardedTLSClientCert:     "Cert",
				xForwardedTLSClientCertInfo: "CertInfo",
				xForwardedPrefix:            "/prefix",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers and invalid Trusted Ips",
			insecure:   false,
			trustedIps: []string{"10.0.1.100"},
			remoteAddr: "10.0.1.101:80",
			incomingHeaders: map[string][]string{
				xForwardedFor:               {"10.0.1.0, 10.0.1.12"},
				xForwardedURI:               {"/bar"},
				xForwardedMethod:            {"GET"},
				xForwardedTLSClientCert:     {"Cert"},
				xForwardedTLSClientCertInfo: {"CertInfo"},
				xForwardedPrefix:            {"/prefix"},
			},
			expectedHeaders: map[string]string{
				xForwardedFor:               "",
				xForwardedURI:               "",
				xForwardedMethod:            "",
				xForwardedTLSClientCert:     "",
				xForwardedTLSClientCertInfo: "",
				xForwardedPrefix:            "",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers and valid Trusted Ips CIDR",
			insecure:   false,
			trustedIps: []string{"1.2.3.4/24"},
			remoteAddr: "1.2.3.156:80",
			incomingHeaders: map[string][]string{
				xForwardedFor:               {"10.0.1.0, 10.0.1.12"},
				xForwardedURI:               {"/bar"},
				xForwardedMethod:            {"GET"},
				xForwardedTLSClientCert:     {"Cert"},
				xForwardedTLSClientCertInfo: {"CertInfo"},
				xForwardedPrefix:            {"/prefix"},
			},
			expectedHeaders: map[string]string{
				xForwardedFor:               "10.0.1.0, 10.0.1.12",
				xForwardedURI:               "/bar",
				xForwardedMethod:            "GET",
				xForwardedTLSClientCert:     "Cert",
				xForwardedTLSClientCertInfo: "CertInfo",
				xForwardedPrefix:            "/prefix",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers and invalid Trusted Ips CIDR",
			insecure:   false,
			trustedIps: []string{"1.2.3.4/24"},
			remoteAddr: "10.0.1.101:80",
			incomingHeaders: map[string][]string{
				xForwardedFor:               {"10.0.1.0, 10.0.1.12"},
				xForwardedURI:               {"/bar"},
				xForwardedMethod:            {"GET"},
				xForwardedTLSClientCert:     {"Cert"},
				xForwardedTLSClientCertInfo: {"CertInfo"},
				xForwardedPrefix:            {"/prefix"},
			},
			expectedHeaders: map[string]string{
				xForwardedFor:               "",
				xForwardedURI:               "",
				xForwardedMethod:            "",
				xForwardedTLSClientCert:     "",
				xForwardedTLSClientCertInfo: "",
				xForwardedPrefix:            "",
			},
		},
		{
			desc:     "xForwardedFor with multiple header(s) values",
			insecure: true,
			incomingHeaders: map[string][]string{
				xForwardedFor: {
					"10.0.0.4, 10.0.0.3",
					"10.0.0.2, 10.0.0.1",
					"10.0.0.0",
				},
			},
			expectedHeaders: map[string]string{
				xForwardedFor: "10.0.0.4, 10.0.0.3, 10.0.0.2, 10.0.0.1, 10.0.0.0",
			},
		},
		{
			desc:       "xRealIP populated from remote address",
			remoteAddr: "10.0.1.101:80",
			expectedHeaders: map[string]string{
				xRealIP: "10.0.1.101",
			},
		},
		{
			desc:       "xRealIP was already populated from previous headers",
			insecure:   true,
			remoteAddr: "10.0.1.101:80",
			incomingHeaders: map[string][]string{
				xRealIP: {"10.0.1.12"},
			},
			expectedHeaders: map[string]string{
				xRealIP: "10.0.1.12",
			},
		},
		{
			desc: "xForwardedProto with no tls",
			tls:  false,
			expectedHeaders: map[string]string{
				xForwardedProto: "http",
			},
		},
		{
			desc: "xForwardedProto with tls",
			tls:  true,
			expectedHeaders: map[string]string{
				xForwardedProto: "https",
			},
		},
		{
			desc:      "xForwardedProto with websocket",
			tls:       false,
			websocket: true,
			expectedHeaders: map[string]string{
				xForwardedProto: "ws",
			},
		},
		{
			desc:      "xForwardedProto with websocket and tls",
			tls:       true,
			websocket: true,
			expectedHeaders: map[string]string{
				xForwardedProto: "wss",
			},
		},
		{
			desc:      "xForwardedProto with websocket and tls and already x-forwarded-proto with wss",
			tls:       true,
			websocket: true,
			incomingHeaders: map[string][]string{
				xForwardedProto: {"wss"},
			},
			expectedHeaders: map[string]string{
				xForwardedProto: "wss",
			},
		},
		{
			desc: "xForwardedPort with explicit port",
			host: "foo.com:8080",
			expectedHeaders: map[string]string{
				xForwardedPort: "8080",
			},
		},
		{
			desc: "xForwardedPort with implicit tls port from proto header",
			// setting insecure just so our initial xForwardedProto does not get cleaned
			insecure: true,
			incomingHeaders: map[string][]string{
				xForwardedProto: {"https"},
			},
			expectedHeaders: map[string]string{
				xForwardedProto: "https",
				xForwardedPort:  "443",
			},
		},
		{
			desc: "xForwardedPort with implicit tls port from TLS in req",
			tls:  true,
			expectedHeaders: map[string]string{
				xForwardedPort: "443",
			},
		},
		{
			desc: "xForwardedHost from req host",
			host: "foo.com:8080",
			expectedHeaders: map[string]string{
				xForwardedHost: "foo.com:8080",
			},
		},
		{
			desc: "xForwardedServer from req XForwarded",
			host: "foo.com:8080",
			expectedHeaders: map[string]string{
				xForwardedServer: "foo.com:8080",
			},
		},
		{
			desc:     "Untrusted: Connection header has no effect on X- forwarded headers",
			insecure: false,
			incomingHeaders: map[string][]string{
				connection: {
					xForwardedProto,
					xForwardedFor,
					xForwardedURI,
					xForwardedMethod,
					xForwardedHost,
					xForwardedPort,
					xForwardedTLSClientCert,
					xForwardedTLSClientCertInfo,
					xForwardedPrefix,
					xRealIP,
				},
				xForwardedProto:             {"foo"},
				xForwardedFor:               {"foo"},
				xForwardedURI:               {"foo"},
				xForwardedMethod:            {"foo"},
				xForwardedHost:              {"foo"},
				xForwardedPort:              {"foo"},
				xForwardedTLSClientCert:     {"foo"},
				xForwardedTLSClientCertInfo: {"foo"},
				xForwardedPrefix:            {"foo"},
				xRealIP:                     {"foo"},
			},
			expectedHeaders: map[string]string{
				xForwardedProto:             "http",
				xForwardedFor:               "",
				xForwardedURI:               "",
				xForwardedMethod:            "",
				xForwardedHost:              "",
				xForwardedPort:              "80",
				xForwardedTLSClientCert:     "",
				xForwardedTLSClientCertInfo: "",
				xForwardedPrefix:            "",
				xRealIP:                     "",
				connection:                  "",
			},
		},
		{
			desc:     "Trusted (insecure): Connection header has no effect on X- forwarded headers",
			insecure: true,
			incomingHeaders: map[string][]string{
				connection: {
					xForwardedProto,
					xForwardedFor,
					xForwardedURI,
					xForwardedMethod,
					xForwardedHost,
					xForwardedPort,
					xForwardedTLSClientCert,
					xForwardedTLSClientCertInfo,
					xForwardedPrefix,
					xRealIP,
				},
				xForwardedProto:             {"foo"},
				xForwardedFor:               {"foo"},
				xForwardedURI:               {"foo"},
				xForwardedMethod:            {"foo"},
				xForwardedHost:              {"foo"},
				xForwardedPort:              {"foo"},
				xForwardedTLSClientCert:     {"foo"},
				xForwardedTLSClientCertInfo: {"foo"},
				xForwardedPrefix:            {"foo"},
				xRealIP:                     {"foo"},
			},
			expectedHeaders: map[string]string{
				xForwardedProto:             "foo",
				xForwardedFor:               "foo",
				xForwardedURI:               "foo",
				xForwardedMethod:            "foo",
				xForwardedHost:              "foo",
				xForwardedPort:              "foo",
				xForwardedTLSClientCert:     "foo",
				xForwardedTLSClientCertInfo: "foo",
				xForwardedPrefix:            "foo",
				xRealIP:                     "foo",
				connection:                  "",
			},
		},
		{
			desc:     "Untrusted and Connection: Connection header has no effect on X- forwarded headers",
			insecure: false,
			connectionHeaders: []string{
				xForwardedProto,
				xForwardedFor,
				xForwardedURI,
				xForwardedMethod,
				xForwardedHost,
				xForwardedPort,
				xForwardedTLSClientCert,
				xForwardedTLSClientCertInfo,
				xForwardedPrefix,
				xRealIP,
			},
			incomingHeaders: map[string][]string{
				connection: {
					xForwardedProto,
					xForwardedFor,
					xForwardedURI,
					xForwardedMethod,
					xForwardedHost,
					xForwardedPort,
					xForwardedTLSClientCert,
					xForwardedTLSClientCertInfo,
					xForwardedPrefix,
					xRealIP,
				},
				xForwardedProto:             {"foo"},
				xForwardedFor:               {"foo"},
				xForwardedURI:               {"foo"},
				xForwardedMethod:            {"foo"},
				xForwardedHost:              {"foo"},
				xForwardedPort:              {"foo"},
				xForwardedTLSClientCert:     {"foo"},
				xForwardedTLSClientCertInfo: {"foo"},
				xForwardedPrefix:            {"foo"},
				xRealIP:                     {"foo"},
			},
			expectedHeaders: map[string]string{
				xForwardedProto:             "http",
				xForwardedFor:               "",
				xForwardedURI:               "",
				xForwardedMethod:            "",
				xForwardedHost:              "",
				xForwardedPort:              "80",
				xForwardedTLSClientCert:     "",
				xForwardedTLSClientCertInfo: "",
				xForwardedPrefix:            "",
				xRealIP:                     "",
				connection:                  "",
			},
		},
		{
			desc:     "Trusted (insecure) and Connection: Connection header has no effect on X- forwarded headers",
			insecure: true,
			connectionHeaders: []string{
				xForwardedProto,
				xForwardedFor,
				xForwardedURI,
				xForwardedMethod,
				xForwardedHost,
				xForwardedPort,
				xForwardedTLSClientCert,
				xForwardedTLSClientCertInfo,
				xForwardedPrefix,
				xRealIP,
			},
			incomingHeaders: map[string][]string{
				connection: {
					xForwardedProto,
					xForwardedFor,
					xForwardedURI,
					xForwardedMethod,
					xForwardedHost,
					xForwardedPort,
					xForwardedTLSClientCert,
					xForwardedTLSClientCertInfo,
					xForwardedPrefix,
					xRealIP,
				},
				xForwardedProto:             {"foo"},
				xForwardedFor:               {"foo"},
				xForwardedURI:               {"foo"},
				xForwardedMethod:            {"foo"},
				xForwardedHost:              {"foo"},
				xForwardedPort:              {"foo"},
				xForwardedTLSClientCert:     {"foo"},
				xForwardedTLSClientCertInfo: {"foo"},
				xForwardedPrefix:            {"foo"},
				xRealIP:                     {"foo"},
			},
			expectedHeaders: map[string]string{
				xForwardedProto:             "foo",
				xForwardedFor:               "foo",
				xForwardedURI:               "foo",
				xForwardedMethod:            "foo",
				xForwardedHost:              "foo",
				xForwardedPort:              "foo",
				xForwardedTLSClientCert:     "foo",
				xForwardedTLSClientCertInfo: "foo",
				xForwardedPrefix:            "foo",
				xRealIP:                     "foo",
				connection:                  "",
			},
		},
		{
			desc: "Connection: one remove, and one passthrough header",
			connectionHeaders: []string{
				"foo",
			},
			incomingHeaders: map[string][]string{
				connection: {
					"foo",
				},
				"Foo": {"bar"},
				"Bar": {"foo"},
			},
			expectedHeaders: map[string]string{
				"Bar": "foo",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, "", nil)
			require.NoError(t, err)

			req.RemoteAddr = test.remoteAddr

			if test.tls {
				req.TLS = &tls.ConnectionState{}
			}

			if test.websocket {
				req.Header.Set(connection, "upgrade")
				req.Header.Set(upgrade, "websocket")
			}

			if test.host != "" {
				req.Host = test.host
			}

			for k, values := range test.incomingHeaders {
				for _, value := range values {
					req.Header.Add(k, value)
				}
			}

			m, err := NewXForwarded(test.insecure, test.trustedIps, test.connectionHeaders,
				http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
			require.NoError(t, err)

			if test.host != "" {
				m.hostname = test.host
			}

			m.ServeHTTP(nil, req)

			for k, v := range test.expectedHeaders {
				assert.Equal(t, v, req.Header.Get(k))
			}
		})
	}
}

func Test_isWebsocketRequest(t *testing.T) {
	testCases := []struct {
		desc             string
		connectionHeader string
		upgradeHeader    string
		assert           assert.BoolAssertionFunc
	}{
		{
			desc:             "connection Header multiple values middle",
			connectionHeader: "foo,upgrade,bar",
			upgradeHeader:    "websocket",
			assert:           assert.True,
		},
		{
			desc:             "connection Header multiple values end",
			connectionHeader: "foo,bar,upgrade",
			upgradeHeader:    "websocket",
			assert:           assert.True,
		},
		{
			desc:             "connection Header multiple values begin",
			connectionHeader: "upgrade,foo,bar",
			upgradeHeader:    "websocket",
			assert:           assert.True,
		},
		{
			desc:             "connection Header no upgrade",
			connectionHeader: "foo,bar",
			upgradeHeader:    "websocket",
			assert:           assert.False,
		},
		{
			desc:             "connection Header empty",
			connectionHeader: "",
			upgradeHeader:    "websocket",
			assert:           assert.False,
		},
		{
			desc:             "no header values",
			connectionHeader: "foo,bar",
			upgradeHeader:    "foo,bar",
			assert:           assert.False,
		},
		{
			desc:             "upgrade header multiple values",
			connectionHeader: "upgrade",
			upgradeHeader:    "foo,bar,websocket",
			assert:           assert.True,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)

			req.Header.Set(connection, test.connectionHeader)
			req.Header.Set(upgrade, test.upgradeHeader)

			ok := isWebsocketRequest(req)

			test.assert(t, ok)
		})
	}
}

func TestConnection(t *testing.T) {
	testCases := []struct {
		desc              string
		reqHeaders        map[string]string
		connectionHeaders []string
		expected          http.Header
	}{
		{
			desc: "simple remove",
			reqHeaders: map[string]string{
				"Foo":      "bar",
				connection: "foo",
			},
			expected: http.Header{},
		},
		{
			desc: "remove and upgrade",
			reqHeaders: map[string]string{
				upgrade:    "test",
				"Foo":      "bar",
				connection: "upgrade,foo",
			},
			expected: http.Header{
				upgrade:    []string{"test"},
				connection: []string{"Upgrade"},
			},
		},
		{
			desc: "no remove",
			reqHeaders: map[string]string{
				"Foo":      "bar",
				connection: "fii",
			},
			expected: http.Header{
				"Foo": []string{"bar"},
			},
		},
		{
			desc: "no remove because connection header pass through",
			reqHeaders: map[string]string{
				"Foo":      "bar",
				connection: "Foo",
			},
			connectionHeaders: []string{"Foo"},
			expected: http.Header{
				"Foo":      []string{"bar"},
				connection: []string{"Foo"},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			forwarded, err := NewXForwarded(true, nil, test.connectionHeaders, nil)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "https://localhost", nil)

			for k, v := range test.reqHeaders {
				req.Header.Set(k, v)
			}

			forwarded.removeConnectionHeaders(req)

			assert.Equal(t, test.expected, req.Header)
		})
	}
}
