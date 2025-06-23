package ingressnginx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/utils/ptr"
)

func Test_parseIngressConfig(t *testing.T) {
	tests := []struct {
		desc        string
		annotations map[string]string
		expected    ingressConfig
	}{
		{
			desc: "all fields set",
			annotations: map[string]string{
				"nginx.ingress.kubernetes.io/ssl-passthrough":         "true",
				"nginx.ingress.kubernetes.io/affinity":                "cookie",
				"nginx.ingress.kubernetes.io/session-cookie-name":     "mycookie",
				"nginx.ingress.kubernetes.io/session-cookie-secure":   "true",
				"nginx.ingress.kubernetes.io/session-cookie-path":     "/foo",
				"nginx.ingress.kubernetes.io/session-cookie-domain":   "example.com",
				"nginx.ingress.kubernetes.io/session-cookie-samesite": "Strict",
				"nginx.ingress.kubernetes.io/session-cookie-max-age":  "3600",
				"nginx.ingress.kubernetes.io/backend-protocol":        "HTTPS",
				"nginx.ingress.kubernetes.io/cors-expose-headers":     "foo, bar",
			},
			expected: ingressConfig{
				SSLPassthrough:        ptr.To(true),
				Affinity:              ptr.To("cookie"),
				SessionCookieName:     ptr.To("mycookie"),
				SessionCookieSecure:   ptr.To(true),
				SessionCookiePath:     ptr.To("/foo"),
				SessionCookieDomain:   ptr.To("example.com"),
				SessionCookieSameSite: ptr.To("Strict"),
				SessionCookieMaxAge:   ptr.To(3600),
				BackendProtocol:       ptr.To("HTTPS"),
				CORSExposeHeaders:     ptr.To([]string{"foo", "bar"}),
			},
		},
		{
			desc: "missing fields",
			annotations: map[string]string{
				"nginx.ingress.kubernetes.io/ssl-passthrough": "false",
			},
			expected: ingressConfig{
				SSLPassthrough: ptr.To(false),
			},
		},
		{
			desc: "invalid bool and int",
			annotations: map[string]string{
				"nginx.ingress.kubernetes.io/ssl-passthrough":                     "notabool",
				"nginx.ingress.kubernetes.io/session-cookie-max-age (in seconds)": "notanint",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var ing netv1.Ingress
			ing.SetAnnotations(test.annotations)

			cfg, err := parseIngressConfig(&ing)
			require.NoError(t, err)

			assert.Equal(t, test.expected, cfg)
		})
	}
}
