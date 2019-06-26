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

func TestLoadIngressRouteTCPs(t *testing.T) {
	testCases := []struct {
		desc         string
		ingressClass string
		paths        []string
		expected     *config.Configuration
	}{
		{
			desc: "Empty",
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
				TLS: &config.TLSConfiguration{},
			},
		},
		{
			desc:  "Simple Ingress Route, with foo entrypoint",
			paths: []string{"tcp/services.yml", "tcp/simple.yml"},
			expected: &config.Configuration{
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"default/test.crd-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Services: map[string]*config.TCPService{
						"default/test.crd-fdd3e9338e47a45efefc": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.2:8000",
										Port:    "",
									},
								},
							},
						},
					},
				},
				TLS: &config.TLSConfiguration{},
			},
		},
		{
			desc:  "One ingress Route with two different rules",
			paths: []string{"tcp/services.yml", "tcp/with_two_rules.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"default/test.crd-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
						"default/test.crd-f44ce589164e656d231c": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-f44ce589164e656d231c",
							Rule:        "HostSNI(`bar.com`)",
						},
					},
					Services: map[string]*config.TCPService{
						"default/test.crd-fdd3e9338e47a45efefc": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.2:8000",
										Port:    "",
									},
								},
							},
						},
						"default/test.crd-f44ce589164e656d231c": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.2:8000",
										Port:    "",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
				TLS: &config.TLSConfiguration{},
			},
		},
		{
			desc:  "One ingress Route with two different services, their servers will merge",
			paths: []string{"tcp/services.yml", "tcp/with_two_services.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"default/test.crd-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Services: map[string]*config.TCPService{
						"default/test.crd-fdd3e9338e47a45efefc": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.2:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.3:8080",
										Port:    "",
									},
									{
										Address: "10.10.0.4:8080",
										Port:    "",
									},
								},
							},
						}},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
				TLS: &config.TLSConfiguration{},
			},
		},
		{
			desc:         "Ingress class does not match",
			paths:        []string{"tcp/services.yml", "tcp/simple.yml"},
			ingressClass: "tchouk",
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
				TLS: &config.TLSConfiguration{},
			},
		},
		{
			desc:  "Route with empty rule value is ignored",
			paths: []string{"tcp/services.yml", "tcp/with_no_rule_value.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
				TLS: &config.TLSConfiguration{},
			},
		},
		{
			desc:  "check rule quoting validity",
			paths: []string{"tcp/services.yml", "tcp/with_bad_host_rule.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
				TLS: &config.TLSConfiguration{},
			},
		},
		{
			desc:  "TLS",
			paths: []string{"tcp/services.yml", "tcp/with_tls.yml"},
			expected: &config.Configuration{
				TLS: &config.TLSConfiguration{
					Certificates: []*tls.Configuration{
						{
							Certificate: tls.Certificate{
								CertFile: tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  tls.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"default/test.crd-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS:         &config.RouterTCPTLSConfig{},
						},
					},
					Services: map[string]*config.TCPService{
						"default/test.crd-fdd3e9338e47a45efefc": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.2:8000",
										Port:    "",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc:  "TLS with passthrough",
			paths: []string{"tcp/services.yml", "tcp/with_tls_passthrough.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"default/test.crd-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &config.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Services: map[string]*config.TCPService{
						"default/test.crd-fdd3e9338e47a45efefc": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.2:8000",
										Port:    "",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
				TLS: &config.TLSConfiguration{},
			},
		},
		{
			desc:  "TLS with tls options",
			paths: []string{"tcp/services.yml", "tcp/with_tls_options.yml"},
			expected: &config.Configuration{
				TLS: &config.TLSConfiguration{
					Options: map[string]tls.Options{
						"default/foo": {
							MinVersion: "VersionTLS12",
							CipherSuites: []string{
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
							},
							ClientCA: tls.ClientCA{
								Files: []tls.FileOrContent{
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								Optional: true,
							},
							SniStrict: true,
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"default/test.crd-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &config.RouterTCPTLSConfig{
								Options: "default/foo",
							},
						},
					},
					Services: map[string]*config.TCPService{
						"default/test.crd-fdd3e9338e47a45efefc": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.2:8000",
										Port:    "",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc:  "TLS with tls options and specific namespace",
			paths: []string{"tcp/services.yml", "tcp/with_tls_options_and_specific_namespace.yml"},
			expected: &config.Configuration{
				TLS: &config.TLSConfiguration{
					Options: map[string]tls.Options{
						"myns/foo": {
							MinVersion: "VersionTLS12",
							CipherSuites: []string{
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
							},
							ClientCA: tls.ClientCA{
								Files: []tls.FileOrContent{
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								Optional: true,
							},
							SniStrict: true,
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"default/test.crd-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &config.RouterTCPTLSConfig{
								Options: "myns/foo",
							},
						},
					},
					Services: map[string]*config.TCPService{
						"default/test.crd-fdd3e9338e47a45efefc": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.2:8000",
										Port:    "",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc:  "TLS with bad tls options",
			paths: []string{"tcp/services.yml", "tcp/with_bad_tls_options.yml"},
			expected: &config.Configuration{
				TLS: &config.TLSConfiguration{
					Options: map[string]tls.Options{
						"default/foo": {
							MinVersion: "VersionTLS12",
							CipherSuites: []string{
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
							},
							ClientCA: tls.ClientCA{
								Files: []tls.FileOrContent{
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								Optional: true,
							},
							SniStrict: true,
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"default/test.crd-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &config.RouterTCPTLSConfig{
								Options: "default/foo",
							},
						},
					},
					Services: map[string]*config.TCPService{
						"default/test.crd-fdd3e9338e47a45efefc": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.2:8000",
										Port:    "",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc:  "TLS with unknown tls options",
			paths: []string{"tcp/services.yml", "tcp/with_unknown_tls_options.yml"},
			expected: &config.Configuration{
				TLS: &config.TLSConfiguration{
					Options: map[string]tls.Options{
						"default/foo": {
							MinVersion: "VersionTLS12",
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"default/test.crd-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &config.RouterTCPTLSConfig{
								Options: "default/unknown",
							},
						},
					},
					Services: map[string]*config.TCPService{
						"default/test.crd-fdd3e9338e47a45efefc": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.2:8000",
										Port:    "",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc:  "TLS with unknown tls options namespace",
			paths: []string{"tcp/services.yml", "tcp/with_unknown_tls_options_namespace.yml"},
			expected: &config.Configuration{
				TLS: &config.TLSConfiguration{
					Options: map[string]tls.Options{
						"default/foo": {
							MinVersion: "VersionTLS12",
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"default/test.crd-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &config.RouterTCPTLSConfig{
								Options: "unknown/foo",
							},
						},
					},
					Services: map[string]*config.TCPService{
						"default/test.crd-fdd3e9338e47a45efefc": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.2:8000",
										Port:    "",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc:  "TLS with ACME",
			paths: []string{"tcp/services.yml", "tcp/with_tls_acme.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"default/test.crd-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default/test.crd-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS:         &config.RouterTCPTLSConfig{},
						},
					},
					Services: map[string]*config.TCPService{
						"default/test.crd-fdd3e9338e47a45efefc": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.2:8000",
										Port:    "",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
				TLS: &config.TLSConfiguration{},
			},
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
			conf := p.loadConfigurationFromCRD(context.Background(), newClientMock(test.paths...))
			assert.Equal(t, test.expected, conf)
		})
	}
}

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
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
				TLS: &config.TLSConfiguration{},
			},
		},
		{
			desc:  "Simple Ingress Route, with foo entrypoint",
			paths: []string{"services.yml", "simple.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
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
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
				TLS: &config.TLSConfiguration{},
			},
		},
		{
			desc:  "Simple Ingress Route with middleware",
			paths: []string{"services.yml", "with_middleware.yml"},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
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
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
				TLS: &config.TLSConfiguration{},
			},
		},
		{
			desc:  "Simple Ingress Route with middleware crossprovider",
			paths: []string{"services.yml", "with_middleware_crossprovider.yml"},
			expected: &config.Configuration{
				TLS: &config.TLSConfiguration{},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"default/test2.crd-23c7f4c450289ee29016": {
							EntryPoints: []string{"web"},
							Service:     "default/test2.crd-23c7f4c450289ee29016",
							Rule:        "Host(`foo.com`) && PathPrefix(`/tobestripped`)",
							Priority:    12,
							Middlewares: []string{"default/stripprefix", "foo/addprefix", "basicauth@file", "redirect@file"},
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
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
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
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
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
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: true,
							},
						},
						"default/test.crd-77c62dfe9517144aeeaa": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
				TLS: &config.TLSConfiguration{},
			},
		},
		{
			desc:  "One ingress Route with two different services, their servers will merge",
			paths: []string{"services.yml", "with_two_services.yml"},
			expected: &config.Configuration{
				TLS: &config.TLSConfiguration{},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
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
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
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
				TLS: &config.TLSConfiguration{},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
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
				TLS: &config.TLSConfiguration{},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
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
				TLS: &config.TLSConfiguration{},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
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
				TLS: &config.TLSConfiguration{},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
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
				TLS: &config.TLSConfiguration{
					Certificates: []*tls.Configuration{
						{
							Certificate: tls.Certificate{
								CertFile: tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  tls.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
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
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc:  "TLS with tls options",
			paths: []string{"services.yml", "with_tls_options.yml"},
			expected: &config.Configuration{
				TLS: &config.TLSConfiguration{
					Options: map[string]tls.Options{
						"default/foo": {
							MinVersion: "VersionTLS12",
							CipherSuites: []string{
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
							},
							ClientCA: tls.ClientCA{
								Files: []tls.FileOrContent{
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								Optional: true,
							},
							SniStrict: true,
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"default/test.crd-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default/test.crd-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &config.RouterTLSConfig{
								Options: "default/foo",
							},
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"default/test.crd-6b204d94623b3df4370c": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc:  "TLS with tls options and specific namespace",
			paths: []string{"services.yml", "with_tls_options_and_specific_namespace.yml"},
			expected: &config.Configuration{
				TLS: &config.TLSConfiguration{
					Options: map[string]tls.Options{
						"myns/foo": {
							MinVersion: "VersionTLS12",
							CipherSuites: []string{
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
							},
							ClientCA: tls.ClientCA{
								Files: []tls.FileOrContent{
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								Optional: true,
							},
							SniStrict: true,
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"default/test.crd-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default/test.crd-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &config.RouterTLSConfig{
								Options: "myns/foo",
							},
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"default/test.crd-6b204d94623b3df4370c": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc:  "TLS with bad tls options",
			paths: []string{"services.yml", "with_bad_tls_options.yml"},
			expected: &config.Configuration{
				TLS: &config.TLSConfiguration{
					Options: map[string]tls.Options{
						"default/foo": {
							MinVersion: "VersionTLS12",
							CipherSuites: []string{
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
							},
							ClientCA: tls.ClientCA{
								Files: []tls.FileOrContent{
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								Optional: true,
							},
							SniStrict: true,
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"default/test.crd-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default/test.crd-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &config.RouterTLSConfig{
								Options: "default/foo",
							},
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"default/test.crd-6b204d94623b3df4370c": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc:  "TLS with unknown tls options",
			paths: []string{"services.yml", "with_unknown_tls_options.yml"},
			expected: &config.Configuration{
				TLS: &config.TLSConfiguration{
					Options: map[string]tls.Options{
						"default/foo": {
							MinVersion: "VersionTLS12",
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"default/test.crd-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default/test.crd-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &config.RouterTLSConfig{
								Options: "default/unknown",
							},
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"default/test.crd-6b204d94623b3df4370c": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc:  "TLS with unknown tls options namespace",
			paths: []string{"services.yml", "with_unknown_tls_options_namespace.yml"},
			expected: &config.Configuration{
				TLS: &config.TLSConfiguration{
					Options: map[string]tls.Options{
						"default/foo": {
							MinVersion: "VersionTLS12",
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"default/test.crd-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default/test.crd-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &config.RouterTLSConfig{
								Options: "unknown/foo",
							},
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"default/test.crd-6b204d94623b3df4370c": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
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
				TLS: &config.TLSConfiguration{},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
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
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
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
				TLS: &config.TLSConfiguration{},
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
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
										URL: "https://10.10.0.5:443",
									},
									{
										URL: "https://10.10.0.6:443",
									},
								},
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
			conf := p.loadConfigurationFromCRD(context.Background(), newClientMock(test.paths...))
			assert.Equal(t, test.expected, conf)
		})
	}
}
