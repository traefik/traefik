package forwardedheaders

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServeHTTP(t *testing.T) {
	testCases := []struct {
		desc            string
		insecure        bool
		trustedIps      []string
		incomingHeaders map[string]string
		remoteAddr      string
		expectedHeaders map[string]string
		tls             bool
		websocket       bool
		host            string
	}{
		{
			desc:            "all Empty",
			insecure:        true,
			trustedIps:      nil,
			remoteAddr:      "",
			incomingHeaders: map[string]string{},
			expectedHeaders: map[string]string{
				"X-Forwarded-for":                  "",
				"X-Forwarded-Uri":                  "",
				"X-Forwarded-Method":               "",
				"X-Forwarded-Tls-Client-Cert":      "",
				"X-Forwarded-Tls-Client-Cert-Info": "",
			},
		},
		{
			desc:       "insecure true with incoming X-Forwarded headers",
			insecure:   true,
			trustedIps: nil,
			remoteAddr: "",
			incomingHeaders: map[string]string{
				"X-Forwarded-for":                  "10.0.1.0, 10.0.1.12",
				"X-Forwarded-Uri":                  "/bar",
				"X-Forwarded-Method":               "GET",
				"X-Forwarded-Tls-Client-Cert":      "Cert",
				"X-Forwarded-Tls-Client-Cert-Info": "CertInfo",
			},
			expectedHeaders: map[string]string{
				"X-Forwarded-for":                  "10.0.1.0, 10.0.1.12",
				"X-Forwarded-Uri":                  "/bar",
				"X-Forwarded-Method":               "GET",
				"X-Forwarded-Tls-Client-Cert":      "Cert",
				"X-Forwarded-Tls-Client-Cert-Info": "CertInfo",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers",
			insecure:   false,
			trustedIps: nil,
			remoteAddr: "",
			incomingHeaders: map[string]string{
				"X-Forwarded-for":                  "10.0.1.0, 10.0.1.12",
				"X-Forwarded-Uri":                  "/bar",
				"X-Forwarded-Method":               "GET",
				"X-Forwarded-Tls-Client-Cert":      "Cert",
				"X-Forwarded-Tls-Client-Cert-Info": "CertInfo",
			},
			expectedHeaders: map[string]string{
				"X-Forwarded-for":                  "",
				"X-Forwarded-Uri":                  "",
				"X-Forwarded-Method":               "",
				"X-Forwarded-Tls-Client-Cert":      "",
				"X-Forwarded-Tls-Client-Cert-Info": "",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers and valid Trusted Ips",
			insecure:   false,
			trustedIps: []string{"10.0.1.100"},
			remoteAddr: "10.0.1.100:80",
			incomingHeaders: map[string]string{
				"X-Forwarded-for":                  "10.0.1.0, 10.0.1.12",
				"X-Forwarded-Uri":                  "/bar",
				"X-Forwarded-Method":               "GET",
				"X-Forwarded-Tls-Client-Cert":      "Cert",
				"X-Forwarded-Tls-Client-Cert-Info": "CertInfo",
			},
			expectedHeaders: map[string]string{
				"X-Forwarded-for":                  "10.0.1.0, 10.0.1.12",
				"X-Forwarded-Uri":                  "/bar",
				"X-Forwarded-Method":               "GET",
				"X-Forwarded-Tls-Client-Cert":      "Cert",
				"X-Forwarded-Tls-Client-Cert-Info": "CertInfo",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers and invalid Trusted Ips",
			insecure:   false,
			trustedIps: []string{"10.0.1.100"},
			remoteAddr: "10.0.1.101:80",
			incomingHeaders: map[string]string{
				"X-Forwarded-for":                  "10.0.1.0, 10.0.1.12",
				"X-Forwarded-Uri":                  "/bar",
				"X-Forwarded-Method":               "GET",
				"X-Forwarded-Tls-Client-Cert":      "Cert",
				"X-Forwarded-Tls-Client-Cert-Info": "CertInfo",
			},
			expectedHeaders: map[string]string{
				"X-Forwarded-for":                  "",
				"X-Forwarded-Uri":                  "",
				"X-Forwarded-Method":               "",
				"X-Forwarded-Tls-Client-Cert":      "",
				"X-Forwarded-Tls-Client-Cert-Info": "",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers and valid Trusted Ips CIDR",
			insecure:   false,
			trustedIps: []string{"1.2.3.4/24"},
			remoteAddr: "1.2.3.156:80",
			incomingHeaders: map[string]string{
				"X-Forwarded-for":                  "10.0.1.0, 10.0.1.12",
				"X-Forwarded-Uri":                  "/bar",
				"X-Forwarded-Method":               "GET",
				"X-Forwarded-Tls-Client-Cert":      "Cert",
				"X-Forwarded-Tls-Client-Cert-Info": "CertInfo",
			},
			expectedHeaders: map[string]string{
				"X-Forwarded-for":                  "10.0.1.0, 10.0.1.12",
				"X-Forwarded-Uri":                  "/bar",
				"X-Forwarded-Method":               "GET",
				"X-Forwarded-Tls-Client-Cert":      "Cert",
				"X-Forwarded-Tls-Client-Cert-Info": "CertInfo",
			},
		},
		{
			desc:       "insecure false with incoming X-Forwarded headers and invalid Trusted Ips CIDR",
			insecure:   false,
			trustedIps: []string{"1.2.3.4/24"},
			remoteAddr: "10.0.1.101:80",
			incomingHeaders: map[string]string{
				"X-Forwarded-for":                  "10.0.1.0, 10.0.1.12",
				"X-Forwarded-Uri":                  "/bar",
				"X-Forwarded-Method":               "GET",
				"X-Forwarded-Tls-Client-Cert":      "Cert",
				"X-Forwarded-Tls-Client-Cert-Info": "CertInfo",
			},
			expectedHeaders: map[string]string{
				"X-Forwarded-for":                  "",
				"X-Forwarded-Uri":                  "",
				"X-Forwarded-Method":               "",
				"X-Forwarded-Tls-Client-Cert":      "",
				"X-Forwarded-Tls-Client-Cert-Info": "",
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
			incomingHeaders: map[string]string{
				xRealIP: "10.0.1.12",
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
			incomingHeaders: map[string]string{
				xForwardedProto: "wss",
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
			incomingHeaders: map[string]string{
				xForwardedProto: "https",
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
	}

	for _, test := range testCases {
		test := test
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

			for k, v := range test.incomingHeaders {
				req.Header.Set(k, v)
			}

			m, err := NewXForwarded(test.insecure, test.trustedIps,
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
