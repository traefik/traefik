package ingress

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/types"
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
				"ingress.kubernetes.io/foo":                                     "bar",
				"traefik.ingress.kubernetes.io/foo":                             "bar",
				"traefik.ingress.kubernetes.io/router.pathmatcher":              "foobar",
				"traefik.ingress.kubernetes.io/router.entrypoints":              "foobar,foobar",
				"traefik.ingress.kubernetes.io/router.middlewares":              "foobar,foobar",
				"traefik.ingress.kubernetes.io/router.priority":                 "42",
				"traefik.ingress.kubernetes.io/router.rulesyntax":               "foobar",
				"traefik.ingress.kubernetes.io/router.tls":                      "true",
				"traefik.ingress.kubernetes.io/router.tls.certresolver":         "foobar",
				"traefik.ingress.kubernetes.io/router.tls.domains.0.main":       "foobar",
				"traefik.ingress.kubernetes.io/router.tls.domains.0.sans":       "foobar,foobar",
				"traefik.ingress.kubernetes.io/router.tls.domains.1.main":       "foobar",
				"traefik.ingress.kubernetes.io/router.tls.domains.1.sans":       "foobar,foobar",
				"traefik.ingress.kubernetes.io/router.tls.options":              "foobar",
				"traefik.ingress.kubernetes.io/router.observability.accessLogs": "true",
				"traefik.ingress.kubernetes.io/router.observability.metrics":    "true",
				"traefik.ingress.kubernetes.io/router.observability.tracing":    "true",
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
						TraceVerbosity: types.MinimalVerbosity,
					},
				},
			},
		},
		{
			desc: "simple TLS annotation",
			annotations: map[string]string{
				"traefik.ingress.kubernetes.io/router.tls": "true",
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
				"ingress.kubernetes.io/foo":                                    "bar",
				"traefik.ingress.kubernetes.io/foo":                            "bar",
				"traefik.ingress.kubernetes.io/service.serversscheme":          "protocol",
				"traefik.ingress.kubernetes.io/service.serverstransport":       "foobar@file",
				"traefik.ingress.kubernetes.io/service.passhostheader":         "true",
				"traefik.ingress.kubernetes.io/service.nativelb":               "true",
				"traefik.ingress.kubernetes.io/service.sticky.cookie":          "true",
				"traefik.ingress.kubernetes.io/service.sticky.cookie.httponly": "true",
				"traefik.ingress.kubernetes.io/service.sticky.cookie.name":     "foobar",
				"traefik.ingress.kubernetes.io/service.sticky.cookie.secure":   "true",
				"traefik.ingress.kubernetes.io/service.sticky.cookie.samesite": "none",
				"traefik.ingress.kubernetes.io/service.sticky.cookie.domain":   "foo.com",
				"traefik.ingress.kubernetes.io/service.sticky.cookie.path":     "foobar",
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
				"traefik.ingress.kubernetes.io/service.sticky.cookie": "true",
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
				"ingress.kubernetes.io/foo":                                     "bar",
				"traefik.ingress.kubernetes.io/foo":                             "bar",
				"traefik.ingress.kubernetes.io/router.pathmatcher":              "foobar",
				"traefik.ingress.kubernetes.io/router.entrypoints":              "foobar,foobar",
				"traefik.ingress.kubernetes.io/router.middlewares":              "foobar,foobar",
				"traefik.ingress.kubernetes.io/router.priority":                 "42",
				"traefik.ingress.kubernetes.io/router.rulesyntax":               "foobar",
				"traefik.ingress.kubernetes.io/router.tls":                      "true",
				"traefik.ingress.kubernetes.io/router.tls.certresolver":         "foobar",
				"traefik.ingress.kubernetes.io/router.tls.domains.0.main":       "foobar",
				"traefik.ingress.kubernetes.io/router.tls.domains.0.sans":       "foobar,foobar",
				"traefik.ingress.kubernetes.io/router.tls.domains.1.main":       "foobar",
				"traefik.ingress.kubernetes.io/router.tls.domains.1.sans":       "foobar,foobar",
				"traefik.ingress.kubernetes.io/router.tls.options":              "foobar",
				"traefik.ingress.kubernetes.io/router.observability.accessLogs": "true",
				"traefik.ingress.kubernetes.io/router.observability.metrics":    "true",
				"traefik.ingress.kubernetes.io/router.observability.tracing":    "true",
			},
			expected: map[string]string{
				"traefik.foo":                             "bar",
				"traefik.router.pathmatcher":              "foobar",
				"traefik.router.entrypoints":              "foobar,foobar",
				"traefik.router.middlewares":              "foobar,foobar",
				"traefik.router.priority":                 "42",
				"traefik.router.rulesyntax":               "foobar",
				"traefik.router.tls":                      "true",
				"traefik.router.tls.certresolver":         "foobar",
				"traefik.router.tls.domains[0].main":      "foobar",
				"traefik.router.tls.domains[0].sans":      "foobar,foobar",
				"traefik.router.tls.domains[1].main":      "foobar",
				"traefik.router.tls.domains[1].sans":      "foobar,foobar",
				"traefik.router.tls.options":              "foobar",
				"traefik.router.observability.accessLogs": "true",
				"traefik.router.observability.metrics":    "true",
				"traefik.router.observability.tracing":    "true",
			},
		},
		{
			desc: "service annotations",
			annotations: map[string]string{
				"traefik.ingress.kubernetes.io/service.serversscheme":          "protocol",
				"traefik.ingress.kubernetes.io/service.serverstransport":       "foobar@file",
				"traefik.ingress.kubernetes.io/service.passhostheader":         "true",
				"traefik.ingress.kubernetes.io/service.sticky.cookie":          "true",
				"traefik.ingress.kubernetes.io/service.sticky.cookie.httponly": "true",
				"traefik.ingress.kubernetes.io/service.sticky.cookie.name":     "foobar",
				"traefik.ingress.kubernetes.io/service.sticky.cookie.secure":   "true",
			},
			expected: map[string]string{
				"traefik.service.passhostheader":         "true",
				"traefik.service.serversscheme":          "protocol",
				"traefik.service.serverstransport":       "foobar@file",
				"traefik.service.sticky.cookie":          "true",
				"traefik.service.sticky.cookie.httponly": "true",
				"traefik.service.sticky.cookie.name":     "foobar",
				"traefik.service.sticky.cookie.secure":   "true",
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
