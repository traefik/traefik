package ingress

import (
	"testing"

	"github.com/baqupio/baqup/v3/pkg/config/dynamic"
	otypes "github.com/baqupio/baqup/v3/pkg/observability/types"
	"github.com/baqupio/baqup/v3/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseRouterConfig(t *testing.T) {
	testCases := []struct {
		desc        string
		annotations map[string]string
		expected    *RouterConfig
	}{
		{
			desc: "router annotations",
			annotations: map[string]string{
				"ingress.kubernetes.io/foo":                                   "bar",
				"baqup.ingress.kubernetes.io/foo":                             "bar",
				"baqup.ingress.kubernetes.io/router.pathmatcher":              "foobar",
				"baqup.ingress.kubernetes.io/router.entrypoints":              "foobar,foobar",
				"baqup.ingress.kubernetes.io/router.middlewares":              "foobar,foobar",
				"baqup.ingress.kubernetes.io/router.priority":                 "42",
				"baqup.ingress.kubernetes.io/router.rulesyntax":               "foobar",
				"baqup.ingress.kubernetes.io/router.tls":                      "true",
				"baqup.ingress.kubernetes.io/router.tls.certresolver":         "foobar",
				"baqup.ingress.kubernetes.io/router.tls.domains.0.main":       "foobar",
				"baqup.ingress.kubernetes.io/router.tls.domains.0.sans":       "foobar,foobar",
				"baqup.ingress.kubernetes.io/router.tls.domains.1.main":       "foobar",
				"baqup.ingress.kubernetes.io/router.tls.domains.1.sans":       "foobar,foobar",
				"baqup.ingress.kubernetes.io/router.tls.options":              "foobar",
				"baqup.ingress.kubernetes.io/router.observability.accessLogs": "true",
				"baqup.ingress.kubernetes.io/router.observability.metrics":    "true",
				"baqup.ingress.kubernetes.io/router.observability.tracing":    "true",
			},
			expected: &RouterConfig{
				Router: &RouterIng{
					PathMatcher: "foobar",
					EntryPoints: []string{"foobar", "foobar"},
					Middlewares: []string{"foobar", "foobar"},
					Priority:    42,
					RuleSyntax:  "foobar",
					TLS: &dynamic.RouterTLSConfig{
						CertResolver: "foobar",
						Domains: []types.Domain{
							{
								Main: "foobar",
								SANs: []string{"foobar", "foobar"},
							},
							{
								Main: "foobar",
								SANs: []string{"foobar", "foobar"},
							},
						},
						Options: "foobar",
					},
					Observability: &dynamic.RouterObservabilityConfig{
						AccessLogs:     pointer(true),
						Tracing:        pointer(true),
						Metrics:        pointer(true),
						TraceVerbosity: otypes.MinimalVerbosity,
					},
				},
			},
		},
		{
			desc: "simple TLS annotation",
			annotations: map[string]string{
				"baqup.ingress.kubernetes.io/router.tls": "true",
			},
			expected: &RouterConfig{
				Router: &RouterIng{
					PathMatcher: "PathPrefix",
					TLS:         &dynamic.RouterTLSConfig{},
				},
			},
		},
		{
			desc:        "empty map",
			annotations: nil,
			expected:    nil,
		},
		{
			desc:        "nil map",
			annotations: nil,
			expected:    nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg, err := parseRouterConfig(test.annotations)
			require.NoError(t, err)

			assert.Equal(t, test.expected, cfg)
		})
	}
}

func Test_parseServiceConfig(t *testing.T) {
	testCases := []struct {
		desc        string
		annotations map[string]string
		expected    *ServiceConfig
	}{
		{
			desc: "service annotations",
			annotations: map[string]string{
				"ingress.kubernetes.io/foo":                                  "bar",
				"baqup.ingress.kubernetes.io/foo":                            "bar",
				"baqup.ingress.kubernetes.io/service.serversscheme":          "protocol",
				"baqup.ingress.kubernetes.io/service.serverstransport":       "foobar@file",
				"baqup.ingress.kubernetes.io/service.passhostheader":         "true",
				"baqup.ingress.kubernetes.io/service.nativelb":               "true",
				"baqup.ingress.kubernetes.io/service.sticky.cookie":          "true",
				"baqup.ingress.kubernetes.io/service.sticky.cookie.httponly": "true",
				"baqup.ingress.kubernetes.io/service.sticky.cookie.name":     "foobar",
				"baqup.ingress.kubernetes.io/service.sticky.cookie.secure":   "true",
				"baqup.ingress.kubernetes.io/service.sticky.cookie.samesite": "none",
				"baqup.ingress.kubernetes.io/service.sticky.cookie.domain":   "foo.com",
				"baqup.ingress.kubernetes.io/service.sticky.cookie.path":     "foobar",
			},
			expected: &ServiceConfig{
				Service: &ServiceIng{
					Sticky: &dynamic.Sticky{
						Cookie: &dynamic.Cookie{
							Name:     "foobar",
							Secure:   true,
							HTTPOnly: true,
							SameSite: "none",
							Domain:   "foo.com",
							Path:     pointer("foobar"),
						},
					},
					ServersScheme:    "protocol",
					ServersTransport: "foobar@file",
					PassHostHeader:   pointer(true),
					NativeLB:         pointer(true),
				},
			},
		},
		{
			desc: "simple sticky annotation",
			annotations: map[string]string{
				"baqup.ingress.kubernetes.io/service.sticky.cookie": "true",
			},
			expected: &ServiceConfig{
				Service: &ServiceIng{
					Sticky: &dynamic.Sticky{
						Cookie: &dynamic.Cookie{
							Path: pointer("/"),
						},
					},
					PassHostHeader: pointer(true),
				},
			},
		},
		{
			desc:        "empty map",
			annotations: map[string]string{},
			expected:    nil,
		},
		{
			desc:        "nil map",
			annotations: nil,
			expected:    nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg, err := parseServiceConfig(test.annotations)
			require.NoError(t, err)

			assert.Equal(t, test.expected, cfg)
		})
	}
}

func Test_convertAnnotations(t *testing.T) {
	testCases := []struct {
		desc        string
		annotations map[string]string
		expected    map[string]string
	}{
		{
			desc: "router annotations",
			annotations: map[string]string{
				"ingress.kubernetes.io/foo":                                   "bar",
				"baqup.ingress.kubernetes.io/foo":                             "bar",
				"baqup.ingress.kubernetes.io/router.pathmatcher":              "foobar",
				"baqup.ingress.kubernetes.io/router.entrypoints":              "foobar,foobar",
				"baqup.ingress.kubernetes.io/router.middlewares":              "foobar,foobar",
				"baqup.ingress.kubernetes.io/router.priority":                 "42",
				"baqup.ingress.kubernetes.io/router.rulesyntax":               "foobar",
				"baqup.ingress.kubernetes.io/router.tls":                      "true",
				"baqup.ingress.kubernetes.io/router.tls.certresolver":         "foobar",
				"baqup.ingress.kubernetes.io/router.tls.domains.0.main":       "foobar",
				"baqup.ingress.kubernetes.io/router.tls.domains.0.sans":       "foobar,foobar",
				"baqup.ingress.kubernetes.io/router.tls.domains.1.main":       "foobar",
				"baqup.ingress.kubernetes.io/router.tls.domains.1.sans":       "foobar,foobar",
				"baqup.ingress.kubernetes.io/router.tls.options":              "foobar",
				"baqup.ingress.kubernetes.io/router.observability.accessLogs": "true",
				"baqup.ingress.kubernetes.io/router.observability.metrics":    "true",
				"baqup.ingress.kubernetes.io/router.observability.tracing":    "true",
			},
			expected: map[string]string{
				"baqup.foo":                             "bar",
				"baqup.router.pathmatcher":              "foobar",
				"baqup.router.entrypoints":              "foobar,foobar",
				"baqup.router.middlewares":              "foobar,foobar",
				"baqup.router.priority":                 "42",
				"baqup.router.rulesyntax":               "foobar",
				"baqup.router.tls":                      "true",
				"baqup.router.tls.certresolver":         "foobar",
				"baqup.router.tls.domains[0].main":      "foobar",
				"baqup.router.tls.domains[0].sans":      "foobar,foobar",
				"baqup.router.tls.domains[1].main":      "foobar",
				"baqup.router.tls.domains[1].sans":      "foobar,foobar",
				"baqup.router.tls.options":              "foobar",
				"baqup.router.observability.accessLogs": "true",
				"baqup.router.observability.metrics":    "true",
				"baqup.router.observability.tracing":    "true",
			},
		},
		{
			desc: "service annotations",
			annotations: map[string]string{
				"baqup.ingress.kubernetes.io/service.serversscheme":          "protocol",
				"baqup.ingress.kubernetes.io/service.serverstransport":       "foobar@file",
				"baqup.ingress.kubernetes.io/service.passhostheader":         "true",
				"baqup.ingress.kubernetes.io/service.sticky.cookie":          "true",
				"baqup.ingress.kubernetes.io/service.sticky.cookie.httponly": "true",
				"baqup.ingress.kubernetes.io/service.sticky.cookie.name":     "foobar",
				"baqup.ingress.kubernetes.io/service.sticky.cookie.secure":   "true",
			},
			expected: map[string]string{
				"baqup.service.passhostheader":         "true",
				"baqup.service.serversscheme":          "protocol",
				"baqup.service.serverstransport":       "foobar@file",
				"baqup.service.sticky.cookie":          "true",
				"baqup.service.sticky.cookie.httponly": "true",
				"baqup.service.sticky.cookie.name":     "foobar",
				"baqup.service.sticky.cookie.secure":   "true",
			},
		},
		{
			desc:        "empty map",
			annotations: map[string]string{},
			expected:    nil,
		},
		{
			desc:        "nil map",
			annotations: nil,
			expected:    nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			labels := convertAnnotations(test.annotations)

			assert.Equal(t, test.expected, labels)
		})
	}
}
