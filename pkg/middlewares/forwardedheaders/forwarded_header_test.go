package forwardedheaders

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/middlewares/forwardedheaders/xheaders"
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
				xheaders.ForwardedFor:               "",
				xheaders.ForwardedURI:               "",
				xheaders.ForwardedMethod:            "",
				xheaders.ForwardedTLSClientCert:     "",
				xheaders.ForwardedTLSClientCertInfo: "",
			},
		},
		{
			desc:       "insecure true with incoming X-Forwarded headers",
			insecure:   true,
			trustedIps: nil,
			remoteAddr: "",
			incomingHeaders: map[string][]string{
				xheaders.ForwardedFor:               {"10.0.1.0, 10.0.1.12"},
				xheaders.ForwardedURI:               {"/bar"},
				xheaders.ForwardedMethod:            {"GET"},
				xheaders.ForwardedTLSClientCert:     {"Cert"},
				xheaders.ForwardedTLSClientCertInfo: {"CertInfo"},
				xheaders.ForwardedPrefix:            {"/prefix"},
				"X_Forwarded_Proto":                 {"https"},
				"X_Forwarded_For":                   {"10.0.0.1"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedFor:               "10.0.1.0, 10.0.1.12",
				xheaders.ForwardedURI:               "/bar",
				xheaders.ForwardedMethod:            "GET",
				xheaders.ForwardedTLSClientCert:     "Cert",
				xheaders.ForwardedTLSClientCertInfo: "CertInfo",
				xheaders.ForwardedPrefix:            "/prefix",
				"X_Forwarded_Proto":                 "https",
				"X_Forwarded_For":                   "10.0.0.1",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers",
			insecure:   false,
			trustedIps: nil,
			remoteAddr: "",
			incomingHeaders: map[string][]string{
				xheaders.ForwardedFor:               {"10.0.1.0, 10.0.1.12"},
				xheaders.ForwardedURI:               {"/bar"},
				xheaders.ForwardedMethod:            {"GET"},
				xheaders.ForwardedTLSClientCert:     {"Cert"},
				xheaders.ForwardedTLSClientCertInfo: {"CertInfo"},
				xheaders.ForwardedPrefix:            {"/prefix"},
				"X_Forwarded_Proto":                 {"https"},
				"X_Forwarded_For":                   {"10.0.0.1"},
				"X_Forwarded_Host":                  {"evil.example"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedFor:               "",
				xheaders.ForwardedURI:               "",
				xheaders.ForwardedMethod:            "",
				xheaders.ForwardedTLSClientCert:     "",
				xheaders.ForwardedTLSClientCertInfo: "",
				xheaders.ForwardedPrefix:            "",
				"X_Forwarded_Proto":                 "",
				"X_Forwarded_For":                   "",
				"X_Forwarded_Host":                  "",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers and valid Trusted Ips",
			insecure:   false,
			trustedIps: []string{"10.0.1.100"},
			remoteAddr: "10.0.1.100:80",
			incomingHeaders: map[string][]string{
				xheaders.ForwardedFor:               {"10.0.1.0, 10.0.1.12"},
				xheaders.ForwardedURI:               {"/bar"},
				xheaders.ForwardedMethod:            {"GET"},
				xheaders.ForwardedTLSClientCert:     {"Cert"},
				xheaders.ForwardedTLSClientCertInfo: {"CertInfo"},
				xheaders.ForwardedPrefix:            {"/prefix"},
				"X_Forwarded_Proto":                 {"https"},
				"X_Forwarded_For":                   {"10.0.0.1"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedFor:               "10.0.1.0, 10.0.1.12",
				xheaders.ForwardedURI:               "/bar",
				xheaders.ForwardedMethod:            "GET",
				xheaders.ForwardedTLSClientCert:     "Cert",
				xheaders.ForwardedTLSClientCertInfo: "CertInfo",
				xheaders.ForwardedPrefix:            "/prefix",
				"X_Forwarded_Proto":                 "https",
				"X_Forwarded_For":                   "10.0.0.1",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers and invalid Trusted Ips",
			insecure:   false,
			trustedIps: []string{"10.0.1.100"},
			remoteAddr: "10.0.1.101:80",
			incomingHeaders: map[string][]string{
				xheaders.ForwardedFor:               {"10.0.1.0, 10.0.1.12"},
				xheaders.ForwardedURI:               {"/bar"},
				xheaders.ForwardedMethod:            {"GET"},
				xheaders.ForwardedTLSClientCert:     {"Cert"},
				xheaders.ForwardedTLSClientCertInfo: {"CertInfo"},
				xheaders.ForwardedPrefix:            {"/prefix"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedFor:               "",
				xheaders.ForwardedURI:               "",
				xheaders.ForwardedMethod:            "",
				xheaders.ForwardedTLSClientCert:     "",
				xheaders.ForwardedTLSClientCertInfo: "",
				xheaders.ForwardedPrefix:            "",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers and valid Trusted Ips CIDR",
			insecure:   false,
			trustedIps: []string{"1.2.3.4/24"},
			remoteAddr: "1.2.3.156:80",
			incomingHeaders: map[string][]string{
				xheaders.ForwardedFor:               {"10.0.1.0, 10.0.1.12"},
				xheaders.ForwardedURI:               {"/bar"},
				xheaders.ForwardedMethod:            {"GET"},
				xheaders.ForwardedTLSClientCert:     {"Cert"},
				xheaders.ForwardedTLSClientCertInfo: {"CertInfo"},
				xheaders.ForwardedPrefix:            {"/prefix"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedFor:               "10.0.1.0, 10.0.1.12",
				xheaders.ForwardedURI:               "/bar",
				xheaders.ForwardedMethod:            "GET",
				xheaders.ForwardedTLSClientCert:     "Cert",
				xheaders.ForwardedTLSClientCertInfo: "CertInfo",
				xheaders.ForwardedPrefix:            "/prefix",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers and invalid Trusted Ips CIDR",
			insecure:   false,
			trustedIps: []string{"1.2.3.4/24"},
			remoteAddr: "10.0.1.101:80",
			incomingHeaders: map[string][]string{
				xheaders.ForwardedFor:               {"10.0.1.0, 10.0.1.12"},
				xheaders.ForwardedURI:               {"/bar"},
				xheaders.ForwardedMethod:            {"GET"},
				xheaders.ForwardedTLSClientCert:     {"Cert"},
				xheaders.ForwardedTLSClientCertInfo: {"CertInfo"},
				xheaders.ForwardedPrefix:            {"/prefix"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedFor:               "",
				xheaders.ForwardedURI:               "",
				xheaders.ForwardedMethod:            "",
				xheaders.ForwardedTLSClientCert:     "",
				xheaders.ForwardedTLSClientCertInfo: "",
				xheaders.ForwardedPrefix:            "",
			},
		},
		{
			desc:     "xheaders.ForwardedFor with multiple header(s) values",
			insecure: true,
			incomingHeaders: map[string][]string{
				xheaders.ForwardedFor: {
					"10.0.0.4, 10.0.0.3",
					"10.0.0.2, 10.0.0.1",
					"10.0.0.0",
				},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedFor: "10.0.0.4, 10.0.0.3, 10.0.0.2, 10.0.0.1, 10.0.0.0",
			},
		},
		{
			desc:       "xheaders.RealIP populated from remote address",
			remoteAddr: "10.0.1.101:80",
			expectedHeaders: map[string]string{
				xheaders.RealIP: "10.0.1.101",
			},
		},
		{
			desc:       "xheaders.RealIP was already populated from previous headers",
			insecure:   true,
			remoteAddr: "10.0.1.101:80",
			incomingHeaders: map[string][]string{
				xheaders.RealIP: {"10.0.1.12"},
			},
			expectedHeaders: map[string]string{
				xheaders.RealIP: "10.0.1.12",
			},
		},
		{
			desc: "xheaders.ForwardedProto with no tls",
			tls:  false,
			expectedHeaders: map[string]string{
				xheaders.ForwardedProto: "http",
			},
		},
		{
			desc: "xheaders.ForwardedProto with tls",
			tls:  true,
			expectedHeaders: map[string]string{
				xheaders.ForwardedProto: "https",
			},
		},
		{
			desc:      "xheaders.ForwardedProto with websocket",
			tls:       false,
			websocket: true,
			expectedHeaders: map[string]string{
				xheaders.ForwardedProto: "ws",
			},
		},
		{
			desc:      "xheaders.ForwardedProto with websocket and tls",
			tls:       true,
			websocket: true,
			expectedHeaders: map[string]string{
				xheaders.ForwardedProto: "wss",
			},
		},
		{
			desc:      "xheaders.ForwardedProto with websocket and tls and already x-forwarded-proto with wss",
			tls:       true,
			websocket: true,
			incomingHeaders: map[string][]string{
				xheaders.ForwardedProto: {"wss"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedProto: "wss",
			},
		},
		{
			desc: "xheaders.ForwardedPort with explicit port",
			host: "foo.com:8080",
			expectedHeaders: map[string]string{
				xheaders.ForwardedPort: "8080",
			},
		},
		{
			desc: "xheaders.ForwardedPort with implicit tls port from proto header",
			// setting insecure just so our initial xheaders.ForwardedProto does not get cleaned
			insecure: true,
			incomingHeaders: map[string][]string{
				xheaders.ForwardedProto: {"https"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedProto: "https",
				xheaders.ForwardedPort:  "443",
			},
		},
		{
			desc: "xheaders.ForwardedPort with implicit tls port from TLS in req",
			tls:  true,
			expectedHeaders: map[string]string{
				xheaders.ForwardedPort: "443",
			},
		},
		{
			desc: "xheaders.ForwardedHost from req host",
			host: "foo.com:8080",
			expectedHeaders: map[string]string{
				xheaders.ForwardedHost: "foo.com:8080",
			},
		},
		{
			desc: "xheaders.ForwardedServer from req XForwarded",
			host: "foo.com:8080",
			expectedHeaders: map[string]string{
				xheaders.ForwardedServer: "foo.com:8080",
			},
		},
		{
			desc:     "Untrusted: Connection header has no effect on X- forwarded headers",
			insecure: false,
			incomingHeaders: map[string][]string{
				connection: {
					xheaders.ForwardedProto,
					xheaders.ForwardedFor,
					xheaders.ForwardedURI,
					xheaders.ForwardedMethod,
					xheaders.ForwardedHost,
					xheaders.ForwardedPort,
					xheaders.ForwardedTLSClientCert,
					xheaders.ForwardedTLSClientCertInfo,
					xheaders.ForwardedPrefix,
					xheaders.RealIP,
					"X_Forwarded_Proto",
				},
				xheaders.ForwardedProto:             {"foo"},
				xheaders.ForwardedFor:               {"foo"},
				xheaders.ForwardedURI:               {"foo"},
				xheaders.ForwardedMethod:            {"foo"},
				xheaders.ForwardedHost:              {"foo"},
				xheaders.ForwardedPort:              {"foo"},
				xheaders.ForwardedTLSClientCert:     {"foo"},
				xheaders.ForwardedTLSClientCertInfo: {"foo"},
				xheaders.ForwardedPrefix:            {"foo"},
				xheaders.RealIP:                     {"foo"},
				"X_Forwarded_Proto":                 {"spoofed"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedProto:             "http",
				xheaders.ForwardedFor:               "",
				xheaders.ForwardedURI:               "",
				xheaders.ForwardedMethod:            "",
				xheaders.ForwardedHost:              "",
				xheaders.ForwardedPort:              "80",
				xheaders.ForwardedTLSClientCert:     "",
				xheaders.ForwardedTLSClientCertInfo: "",
				xheaders.ForwardedPrefix:            "",
				xheaders.RealIP:                     "",
				"X_Forwarded_Proto":                 "",
				connection:                          "",
			},
		},
		{
			desc:     "Trusted (insecure): Connection header has no effect on X- forwarded headers",
			insecure: true,
			incomingHeaders: map[string][]string{
				connection: {
					xheaders.ForwardedProto,
					xheaders.ForwardedFor,
					xheaders.ForwardedURI,
					xheaders.ForwardedMethod,
					xheaders.ForwardedHost,
					xheaders.ForwardedPort,
					xheaders.ForwardedTLSClientCert,
					xheaders.ForwardedTLSClientCertInfo,
					xheaders.ForwardedPrefix,
					xheaders.RealIP,
				},
				xheaders.ForwardedProto:             {"foo"},
				xheaders.ForwardedFor:               {"foo"},
				xheaders.ForwardedURI:               {"foo"},
				xheaders.ForwardedMethod:            {"foo"},
				xheaders.ForwardedHost:              {"foo"},
				xheaders.ForwardedPort:              {"foo"},
				xheaders.ForwardedTLSClientCert:     {"foo"},
				xheaders.ForwardedTLSClientCertInfo: {"foo"},
				xheaders.ForwardedPrefix:            {"foo"},
				xheaders.RealIP:                     {"foo"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedProto:             "foo",
				xheaders.ForwardedFor:               "foo",
				xheaders.ForwardedURI:               "foo",
				xheaders.ForwardedMethod:            "foo",
				xheaders.ForwardedHost:              "foo",
				xheaders.ForwardedPort:              "foo",
				xheaders.ForwardedTLSClientCert:     "foo",
				xheaders.ForwardedTLSClientCertInfo: "foo",
				xheaders.ForwardedPrefix:            "foo",
				xheaders.RealIP:                     "foo",
				connection:                          "",
			},
		},
		{
			desc:     "Untrusted and Connection: Connection header has no effect on X- forwarded headers",
			insecure: false,
			connectionHeaders: []string{
				xheaders.ForwardedProto,
				xheaders.ForwardedFor,
				xheaders.ForwardedURI,
				xheaders.ForwardedMethod,
				xheaders.ForwardedHost,
				xheaders.ForwardedPort,
				xheaders.ForwardedTLSClientCert,
				xheaders.ForwardedTLSClientCertInfo,
				xheaders.ForwardedPrefix,
				xheaders.RealIP,
			},
			incomingHeaders: map[string][]string{
				connection: {
					xheaders.ForwardedProto,
					xheaders.ForwardedFor,
					xheaders.ForwardedURI,
					xheaders.ForwardedMethod,
					xheaders.ForwardedHost,
					xheaders.ForwardedPort,
					xheaders.ForwardedTLSClientCert,
					xheaders.ForwardedTLSClientCertInfo,
					xheaders.ForwardedPrefix,
					xheaders.RealIP,
				},
				xheaders.ForwardedProto:             {"foo"},
				xheaders.ForwardedFor:               {"foo"},
				xheaders.ForwardedURI:               {"foo"},
				xheaders.ForwardedMethod:            {"foo"},
				xheaders.ForwardedHost:              {"foo"},
				xheaders.ForwardedPort:              {"foo"},
				xheaders.ForwardedTLSClientCert:     {"foo"},
				xheaders.ForwardedTLSClientCertInfo: {"foo"},
				xheaders.ForwardedPrefix:            {"foo"},
				xheaders.RealIP:                     {"foo"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedProto:             "http",
				xheaders.ForwardedFor:               "",
				xheaders.ForwardedURI:               "",
				xheaders.ForwardedMethod:            "",
				xheaders.ForwardedHost:              "",
				xheaders.ForwardedPort:              "80",
				xheaders.ForwardedTLSClientCert:     "",
				xheaders.ForwardedTLSClientCertInfo: "",
				xheaders.ForwardedPrefix:            "",
				xheaders.RealIP:                     "",
				connection:                          "",
			},
		},
		{
			desc:     "Trusted (insecure) and Connection: Connection header has no effect on X- forwarded headers",
			insecure: true,
			connectionHeaders: []string{
				xheaders.ForwardedProto,
				xheaders.ForwardedFor,
				xheaders.ForwardedURI,
				xheaders.ForwardedMethod,
				xheaders.ForwardedHost,
				xheaders.ForwardedPort,
				xheaders.ForwardedTLSClientCert,
				xheaders.ForwardedTLSClientCertInfo,
				xheaders.ForwardedPrefix,
				xheaders.RealIP,
			},
			incomingHeaders: map[string][]string{
				connection: {
					xheaders.ForwardedProto,
					xheaders.ForwardedFor,
					xheaders.ForwardedURI,
					xheaders.ForwardedMethod,
					xheaders.ForwardedHost,
					xheaders.ForwardedPort,
					xheaders.ForwardedTLSClientCert,
					xheaders.ForwardedTLSClientCertInfo,
					xheaders.ForwardedPrefix,
					xheaders.RealIP,
				},
				xheaders.ForwardedProto:             {"foo"},
				xheaders.ForwardedFor:               {"foo"},
				xheaders.ForwardedURI:               {"foo"},
				xheaders.ForwardedMethod:            {"foo"},
				xheaders.ForwardedHost:              {"foo"},
				xheaders.ForwardedPort:              {"foo"},
				xheaders.ForwardedTLSClientCert:     {"foo"},
				xheaders.ForwardedTLSClientCertInfo: {"foo"},
				xheaders.ForwardedPrefix:            {"foo"},
				xheaders.RealIP:                     {"foo"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedProto:             "foo",
				xheaders.ForwardedFor:               "foo",
				xheaders.ForwardedURI:               "foo",
				xheaders.ForwardedMethod:            "foo",
				xheaders.ForwardedHost:              "foo",
				xheaders.ForwardedPort:              "foo",
				xheaders.ForwardedTLSClientCert:     "foo",
				xheaders.ForwardedTLSClientCertInfo: "foo",
				xheaders.ForwardedPrefix:            "foo",
				xheaders.RealIP:                     "foo",
				connection:                          "",
			},
		},
		{
			desc:     "Trusted (insecure) and Connection: Testing case sensitivity on connection Headers param",
			insecure: true,
			connectionHeaders: []string{
				strings.ToLower(xheaders.ForwardedProto),
				strings.ToLower(xheaders.ForwardedFor),
				strings.ToLower(xheaders.ForwardedURI),
				strings.ToLower(xheaders.ForwardedMethod),
				strings.ToLower(xheaders.ForwardedHost),
				strings.ToLower(xheaders.ForwardedPort),
				strings.ToLower(xheaders.ForwardedTLSClientCert),
				strings.ToLower(xheaders.ForwardedTLSClientCertInfo),
				strings.ToLower(xheaders.ForwardedPrefix),
				strings.ToLower(xheaders.RealIP),
			},
			incomingHeaders: map[string][]string{
				connection: {
					xheaders.ForwardedProto,
					xheaders.ForwardedFor,
					xheaders.ForwardedURI,
					xheaders.ForwardedMethod,
					xheaders.ForwardedHost,
					xheaders.ForwardedPort,
					xheaders.ForwardedTLSClientCert,
					xheaders.ForwardedTLSClientCertInfo,
					xheaders.ForwardedPrefix,
					xheaders.RealIP,
				},
				xheaders.ForwardedProto:             {"foo"},
				xheaders.ForwardedFor:               {"foo"},
				xheaders.ForwardedURI:               {"foo"},
				xheaders.ForwardedMethod:            {"foo"},
				xheaders.ForwardedHost:              {"foo"},
				xheaders.ForwardedPort:              {"foo"},
				xheaders.ForwardedTLSClientCert:     {"foo"},
				xheaders.ForwardedTLSClientCertInfo: {"foo"},
				xheaders.ForwardedPrefix:            {"foo"},
				xheaders.RealIP:                     {"foo"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedProto:             "foo",
				xheaders.ForwardedFor:               "foo",
				xheaders.ForwardedURI:               "foo",
				xheaders.ForwardedMethod:            "foo",
				xheaders.ForwardedHost:              "foo",
				xheaders.ForwardedPort:              "foo",
				xheaders.ForwardedTLSClientCert:     "foo",
				xheaders.ForwardedTLSClientCertInfo: "foo",
				xheaders.ForwardedPrefix:            "foo",
				xheaders.RealIP:                     "foo",
				connection:                          "",
			},
		},
		{
			desc:     "Trusted (insecure) and Connection: Testing case sensitivity on X- forwarded headers",
			insecure: true,
			incomingHeaders: map[string][]string{
				connection: {
					strings.ToLower(xheaders.ForwardedProto),
					strings.ToLower(xheaders.ForwardedFor),
					strings.ToLower(xheaders.ForwardedURI),
					strings.ToLower(xheaders.ForwardedMethod),
					strings.ToLower(xheaders.ForwardedHost),
					strings.ToLower(xheaders.ForwardedPort),
					strings.ToLower(xheaders.ForwardedTLSClientCert),
					strings.ToLower(xheaders.ForwardedTLSClientCertInfo),
					strings.ToLower(xheaders.ForwardedPrefix),
					strings.ToLower(xheaders.RealIP),
				},
				xheaders.ForwardedProto:             {"foo"},
				xheaders.ForwardedFor:               {"foo"},
				xheaders.ForwardedURI:               {"foo"},
				xheaders.ForwardedMethod:            {"foo"},
				xheaders.ForwardedHost:              {"foo"},
				xheaders.ForwardedPort:              {"foo"},
				xheaders.ForwardedTLSClientCert:     {"foo"},
				xheaders.ForwardedTLSClientCertInfo: {"foo"},
				xheaders.ForwardedPrefix:            {"foo"},
				xheaders.RealIP:                     {"foo"},
			},
			expectedHeaders: map[string]string{
				xheaders.ForwardedProto:             "foo",
				xheaders.ForwardedFor:               "foo",
				xheaders.ForwardedURI:               "foo",
				xheaders.ForwardedMethod:            "foo",
				xheaders.ForwardedHost:              "foo",
				xheaders.ForwardedPort:              "foo",
				xheaders.ForwardedTLSClientCert:     "foo",
				xheaders.ForwardedTLSClientCertInfo: "foo",
				xheaders.ForwardedPrefix:            "foo",
				xheaders.RealIP:                     "foo",
				connection:                          "",
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
					"bar",
				},
				"Foo": {"bar"},
				"Bar": {"foo"},
			},
			expectedHeaders: map[string]string{
				"Bar": "",
				"Foo": "bar",
			},
		},
		{
			desc:       "insecure false preserves non-matching underscore headers",
			insecure:   false,
			remoteAddr: "10.0.1.101:80",
			incomingHeaders: map[string][]string{
				"X_Custom_Header":   {"value"},
				"X_Forwarded_Proto": {"spoofed"},
			},
			expectedHeaders: map[string]string{
				"X_Custom_Header":   "value",
				"X_Forwarded_Proto": "",
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
