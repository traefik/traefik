package ingressnginx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	netv1 "k8s.io/api/networking/v1"
)

func Test_parseIngressConfig(t *testing.T) {
	tests := []struct {
		desc        string
		annotations map[string]string
		expected    IngressConfig
	}{
		{
			desc: "all fields set",
			annotations: map[string]string{
				"nginx.ingress.kubernetes.io/ssl-passthrough":          "true",
				"nginx.ingress.kubernetes.io/affinity":                 "cookie",
				"nginx.ingress.kubernetes.io/session-cookie-name":      "mycookie",
				"nginx.ingress.kubernetes.io/session-cookie-secure":    "true",
				"nginx.ingress.kubernetes.io/session-cookie-path":      "/foo",
				"nginx.ingress.kubernetes.io/session-cookie-domain":    "example.com",
				"nginx.ingress.kubernetes.io/session-cookie-samesite":  "Strict",
				"nginx.ingress.kubernetes.io/session-cookie-max-age":   "3600",
				"nginx.ingress.kubernetes.io/backend-protocol":         "HTTPS",
				"nginx.ingress.kubernetes.io/cors-expose-headers":      "foo, bar",
				"nginx.ingress.kubernetes.io/auth-url":                 "http://auth.example.com/verify",
				"nginx.ingress.kubernetes.io/auth-signin":              "https://auth.example.com/oauth2/start?rd=foo",
				"nginx.ingress.kubernetes.io/proxy-connect-timeout":    "30",
				"nginx.ingress.kubernetes.io/proxy-request-buffering":  "on",
				"nginx.ingress.kubernetes.io/client-body-buffer-size":  "16k",
				"nginx.ingress.kubernetes.io/proxy-body-size":          "16k",
				"nginx.ingress.kubernetes.io/proxy-buffering":          "on",
				"nginx.ingress.kubernetes.io/proxy-buffer-size":        "16k",
				"nginx.ingress.kubernetes.io/proxy-buffers-number":     "8",
				"nginx.ingress.kubernetes.io/proxy-max-temp-file-size": "100m",
				"nginx.ingress.kubernetes.io/limit-rpm":                "120",
				"nginx.ingress.kubernetes.io/x-forwarded-prefix":       "/test",
				"nginx.ingress.kubernetes.io/upstream-vhost":           "upstream-vhost",
			},
			expected: IngressConfig{
				SSLPassthrough:        new(true),
				Affinity:              new("cookie"),
				SessionCookieName:     new("mycookie"),
				SessionCookieSecure:   new(true),
				SessionCookiePath:     new("/foo"),
				SessionCookieDomain:   new("example.com"),
				SessionCookieSameSite: new("Strict"),
				SessionCookieMaxAge:   new(3600),
				BackendProtocol:       new("HTTPS"),
				CORSExposeHeaders:     new([]string{"foo", "bar"}),
				AuthURL:               new("http://auth.example.com/verify"),
				AuthSignin:            new("https://auth.example.com/oauth2/start?rd=foo"),
				ProxyConnectTimeout:   new(30),
				ProxyRequestBuffering: new("on"),
				ClientBodyBufferSize:  new("16k"),
				ProxyBodySize:         new("16k"),
				ProxyBuffering:        new("on"),
				ProxyBufferSize:       new("16k"),
				ProxyBuffersNumber:    new(8),
				ProxyMaxTempFileSize:  new("100m"),
				LimitRPM:              new(120),
				XForwardedPrefix:      new("/test"),
				UpstreamVHost:         new("upstream-vhost"),
			},
		},
		{
			desc: "missing fields",
			annotations: map[string]string{
				"nginx.ingress.kubernetes.io/ssl-passthrough": "false",
			},
			expected: IngressConfig{
				SSLPassthrough: new(false),
			},
		},
		{
			desc: "invalid bool and int",
			annotations: map[string]string{
				"nginx.ingress.kubernetes.io/ssl-passthrough":                     "notabool",
				"nginx.ingress.kubernetes.io/session-cookie-max-age (in seconds)": "notanint",
			},
			expected: IngressConfig{
				SSLPassthrough: new(false),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var ing netv1.Ingress
			ing.SetAnnotations(test.annotations)

			assert.Equal(t, test.expected, parseIngressConfig(&ing))
		})
	}
}
