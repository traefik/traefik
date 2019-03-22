package crd

import (
	"context"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/provider"
	"github.com/containous/traefik/pkg/tls"
	"github.com/stretchr/testify/assert"
)

var _ provider.Provider = (*Provider)(nil)

func TestLoadIngressRoutes(t *testing.T) {
	testCases := []struct {
		desc         string
		ingressClass string
		paths        []string
		expected     *config.Configuration
	}{
		{
			desc: "Empty",
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc:  "Simple Ingress Route, with foo entrypoint",
			paths: []string{"services.yml", "simple.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"default/test.crd-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"default/test.crd-6b204d94623b3df4370c": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL:    "http://10.10.0.1:80",
										Weight: 1,
									},
									{
										URL:    "http://10.10.0.2:80",
										Weight: 1,
									},
								},
								Method:         "wrr",
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc:  "Simple Ingress Route with middleware",
			paths: []string{"services.yml", "with_middleware.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"default/test2.crd-23c7f4c450289ee29016": {
							EntryPoints: []string{"web"},
							Service:     "default/test2.crd-23c7f4c450289ee29016",
							Rule:        "Host(`foo.com`) && PathPrefix(`/tobestripped`)",
							Priority:    12,
							Middlewares: []string{"default/stripprefix", "foo/addprefix"},
						},
					},
					Middlewares: map[string]*config.Middleware{
						"default/stripprefix": {
							StripPrefix: &config.StripPrefix{
								Prefixes: []string{"/tobestripped"},
							},
						},
						"foo/addprefix": {
							AddPrefix: &config.AddPrefix{
								Prefix: "/tobeadded",
							},
						},
					},
					Services: map[string]*config.Service{
						"default/test2.crd-23c7f4c450289ee29016": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL:    "http://10.10.0.1:80",
										Weight: 1,
									},
									{
										URL:    "http://10.10.0.2:80",
										Weight: 1,
									},
								},
								Method:         "wrr",
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc:  "One ingress Route with two different rules",
			paths: []string{"services.yml", "with_two_rules.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"default/test.crd-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Service:     "default/test.crd-6b204d94623b3df4370c",
							Priority:    14,
						},
						"default/test.crd-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default/test.crd-77c62dfe9517144aeeaa",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"default/test.crd-6b204d94623b3df4370c": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL:    "http://10.10.0.1:80",
										Weight: 1,
									},
									{
										URL:    "http://10.10.0.2:80",
										Weight: 1,
									},
								},
								Method:         "wrr",
								PassHostHeader: true,
							},
						},
						"default/test.crd-77c62dfe9517144aeeaa": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL:    "http://10.10.0.1:80",
										Weight: 1,
									},
									{
										URL:    "http://10.10.0.2:80",
										Weight: 1,
									},
								},
								Method:         "wrr",
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc:  "One ingress Route with two different services, their servers will merge",
			paths: []string{"services.yml", "with_two_services.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"default/test.crd-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default/test.crd-77c62dfe9517144aeeaa",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"default/test.crd-77c62dfe9517144aeeaa": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL:    "http://10.10.0.1:80",
										Weight: 1,
									},
									{
										URL:    "http://10.10.0.2:80",
										Weight: 1,
									},
									{
										URL:    "http://10.10.0.3:8080",
										Weight: 1,
									},
									{
										URL:    "http://10.10.0.4:8080",
										Weight: 1,
									},
								},
								Method:         "wrr",
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc:         "Ingress class",
			paths:        []string{"services.yml", "simple.yml"},
			ingressClass: "tchouk",
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc:  "Route with empty rule value is ignored",
			paths: []string{"services.yml", "with_no_rule_value.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc:  "Route with kind not of a rule type (empty kind) is ignored",
			paths: []string{"services.yml", "with_wrong_rule_kind.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc:  "check rule quoting validity",
			paths: []string{"services.yml", "with_bad_host_rule.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc:  "TLS",
			paths: []string{"services.yml", "with_tls.yml"},
			expected: &config.Configuration{
				TLS: []*tls.Configuration{
					{
						Certificate: &tls.Certificate{
							CertFile: tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
							KeyFile:  tls.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
						},
					},
				},
				TCP: &config.TCPConfiguration{},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"default/test.crd-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default/test.crd-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &config.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"default/test.crd-6b204d94623b3df4370c": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL:    "http://10.10.0.1:80",
										Weight: 1,
									},
									{
										URL:    "http://10.10.0.2:80",
										Weight: 1,
									},
								},
								Method:         "wrr",
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc:  "TLS with ACME",
			paths: []string{"services.yml", "with_tls_acme.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"default/test.crd-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default/test.crd-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &config.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"default/test.crd-6b204d94623b3df4370c": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL:    "http://10.10.0.1:80",
										Weight: 1,
									},
									{
										URL:    "http://10.10.0.2:80",
										Weight: 1,
									},
								},
								Method:         "wrr",
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc:  "Simple Ingress Route, defaulting to https for servers",
			paths: []string{"services.yml", "with_https_default.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"default/test.crd-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"default/test.crd-6b204d94623b3df4370c": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL:    "https://10.10.0.5:443",
										Weight: 1,
									},
									{
										URL:    "https://10.10.0.6:443",
										Weight: 1,
									},
								},
								Method:         "wrr",
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "port selected by name (TODO)",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.expected == nil {
				return
			}

			p := Provider{IngressClass: test.ingressClass}
			conf := p.loadConfigurationFromIngresses(context.Background(), newClientMock(test.paths...))
			assert.Equal(t, test.expected, conf)
		})
	}
}
