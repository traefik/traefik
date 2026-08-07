package crd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	auth "github.com/abbot/go-http-auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/provider"
	traefikcrdfake "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned/fake"
	traefikv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v2/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v2/pkg/tls"
	"github.com/traefik/traefik/v2/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

var _ provider.Provider = (*Provider)(nil)

func TestLoadIngressRouteTCPs(t *testing.T) {
	testCases := []struct {
		desc               string
		ingressClass       string
		paths              []string
		allowEmptyServices bool
		expected           *dynamic.Configuration
	}{
		{
			desc: "Empty",
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Simple Ingress Route, with foo entrypoint",
			paths: []string{"tcp/services.yml", "tcp/simple.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Simple Ingress Route, with foo entrypoint and middleware",
			paths: []string{"tcp/services.yml", "tcp/with_middleware.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Middlewares: []string{"default-ipwhitelist-72ecdadecb80eec37e7b", "foo-ipwhitelist-b4114023c2ec108dc3a6"},
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipwhitelist-72ecdadecb80eec37e7b": {
							IPWhiteList: &dynamic.TCPIPWhiteList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						"foo-ipwhitelist-b4114023c2ec108dc3a6": {
							IPWhiteList: &dynamic.TCPIPWhiteList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Middlewares in ingress route config are normalized",
			paths: []string{"tcp/services.yml", "tcp/with_middleware_multiple_hyphens.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Middlewares: []string{"default-multiple-hyphens-5c9512a3890cb0f6fbe2"},
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-multiple-hyphens-5c9512a3890cb0f6fbe2": {
							IPWhiteList: &dynamic.TCPIPWhiteList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Simple Ingress Route, with foo entrypoint and crossprovider middleware",
			paths: []string{"tcp/services.yml", "tcp/with_middleware_cross_provider.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Middlewares: []string{"default-ipwhitelist-72ecdadecb80eec37e7b", "foo-ipwhitelist-b4114023c2ec108dc3a6", "ipwhitelist@file", "ipwhitelist-foo@file"},
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipwhitelist-72ecdadecb80eec37e7b": {
							IPWhiteList: &dynamic.TCPIPWhiteList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						"foo-ipwhitelist-b4114023c2ec108dc3a6": {
							IPWhiteList: &dynamic.TCPIPWhiteList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "One ingress Route with two different rules",
			paths: []string{"tcp/services.yml", "tcp/with_two_rules.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
						},
						"default-test-route-1-73e4808ba7263def92b2": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-1-lb-1109d6313f2dbb16a8e6",
							Rule:        "HostSNI(`bar.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
						"default-test-route-1-lb-1109d6313f2dbb16a8e6": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "One ingress Route with two identical rules",
			paths: []string{"tcp/services.yml", "tcp/with_two_identical_rules.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					// The two identical rules do not overwrite each other anymore:
					// the route index is part of the generated names.
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
						},
						"default-test-route-1-73e4808ba7263def92b2": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-1-lb-1109d6313f2dbb16a8e6",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
						"default-test-route-1-lb-1109d6313f2dbb16a8e6": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "One ingress Route with two different services",
			paths: []string{"tcp/services.yml", "tcp/with_two_services.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-wrr-a8d5191cb9281af18d2d",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-test-route-0-wrr-whoamitcp-8000-4987436aa9bf251cc2e1",
										Weight: new(2),
									},
									{
										Name:   "default-test-route-0-wrr-whoamitcp2-8080-80d7a6e501859a4d4f62",
										Weight: new(3),
									},
								},
							},
						},
						"default-test-route-0-wrr-whoamitcp-8000-4987436aa9bf251cc2e1": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
						"default-test-route-0-wrr-whoamitcp2-8080-80d7a6e501859a4d4f62": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.3:8080",
									},
									{
										Address: "10.10.0.4:8080",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "One ingress Route with different services namespaces",
			paths: []string{"tcp/services.yml", "tcp/with_different_services_ns.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-wrr-a8d5191cb9281af18d2d",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-test-route-0-wrr-whoamitcp-8000-4987436aa9bf251cc2e1",
										Weight: new(2),
									},
									{
										Name:   "default-test-route-0-wrr-whoamitcp2-8080-80d7a6e501859a4d4f62",
										Weight: new(3),
									},
									{
										Name:   "default-test-route-0-wrr-whoamitcp3-8083-ce7ce8106b0d356747a1",
										Weight: new(4),
									},
								},
							},
						},
						"default-test-route-0-wrr-whoamitcp-8000-4987436aa9bf251cc2e1": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
						"default-test-route-0-wrr-whoamitcp2-8080-80d7a6e501859a4d4f62": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.3:8080",
									},
									{
										Address: "10.10.0.4:8080",
									},
								},
							},
						},
						"default-test-route-0-wrr-whoamitcp3-8083-ce7ce8106b0d356747a1": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.7:8083",
									},
									{
										Address: "10.10.0.8:8083",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:         "Ingress class does not match",
			paths:        []string{"tcp/services.yml", "tcp/simple.yml"},
			ingressClass: "tchouk",
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Route with empty rule value is ignored",
			paths: []string{"tcp/services.yml", "tcp/with_no_rule_value.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "TLS",
			paths: []string{"tcp/services.yml", "tcp/with_tls.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							Certificate: tls.Certificate{
								CertFile: tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  tls.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with passthrough",
			paths: []string{"tcp/services.yml", "tcp/with_tls_passthrough.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "TLS with tls options",
			paths: []string{"tcp/services.yml", "tcp/with_tls_options.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"default-foo-ce203de49508e8c70b06": {
							MinVersion: "VersionTLS12",
							CipherSuites: []string{
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
							},
							ClientAuth: tls.ClientAuth{
								CAFiles: []tls.FileOrContent{
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								ClientAuthType: "VerifyClientCertIfGiven",
							},
							SniStrict: true,
							ALPNProtocols: []string{
								"h2",
								"http/1.1",
								"acme-tls/1",
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "default-foo-ce203de49508e8c70b06",
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with tls options and specific namespace",
			paths: []string{"tcp/services.yml", "tcp/with_tls_options_and_specific_namespace.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"myns-foo-d809e81d5ccf461aab60": {
							MinVersion: "VersionTLS12",
							CipherSuites: []string{
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
							},
							ClientAuth: tls.ClientAuth{
								CAFiles: []tls.FileOrContent{
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								ClientAuthType: "VerifyClientCertIfGiven",
							},
							SniStrict: true,
							ALPNProtocols: []string{
								"h2",
								"http/1.1",
								"acme-tls/1",
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "myns-foo-d809e81d5ccf461aab60",
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with bad tls options",
			paths: []string{"tcp/services.yml", "tcp/with_bad_tls_options.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"default-foo-ce203de49508e8c70b06": {
							MinVersion: "VersionTLS12",
							CipherSuites: []string{
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
							},
							ClientAuth: tls.ClientAuth{
								CAFiles: []tls.FileOrContent{
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								ClientAuthType: "VerifyClientCertIfGiven",
							},
							SniStrict: true,
							ALPNProtocols: []string{
								"h2",
								"http/1.1",
								"acme-tls/1",
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "default-foo-ce203de49508e8c70b06",
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with unknown tls options",
			paths: []string{"tcp/services.yml", "tcp/with_unknown_tls_options.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"default-foo-ce203de49508e8c70b06": {
							MinVersion: "VersionTLS12",
							ALPNProtocols: []string{
								"h2",
								"http/1.1",
								"acme-tls/1",
							},
							CipherSuites: []string{
								"TLS_AES_128_GCM_SHA256",
								"TLS_AES_256_GCM_SHA384",
								"TLS_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "default-unknown-d612e69053993bc351c4",
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with unknown tls options namespace",
			paths: []string{"tcp/services.yml", "tcp/with_unknown_tls_options_namespace.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"default-foo-ce203de49508e8c70b06": {
							MinVersion: "VersionTLS12",
							ALPNProtocols: []string{
								"h2",
								"http/1.1",
								"acme-tls/1",
							},
							CipherSuites: []string{
								"TLS_AES_128_GCM_SHA256",
								"TLS_AES_256_GCM_SHA384",
								"TLS_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "unknown-foo-60df693b64851d2ececa",
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with ACME",
			paths: []string{"tcp/services.yml", "tcp/with_tls_acme.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "TCP with terminationDelay",
			paths: []string{"tcp/services.yml", "tcp/with_termination_delay.yml"},
			expected: &dynamic.Configuration{
				TLS: &dynamic.TLSConfiguration{},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
								TerminationDelay: new(500),
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with tls Store",
			paths: []string{"tcp/services.yml", "tcp/with_tls_store.yml"},
			expected: &dynamic.Configuration{
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{
						"default": {
							DefaultCertificate: &tls.Certificate{
								CertFile: tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  tls.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Simple Ingress Route, with externalName service",
			paths: []string{"tcp/services.yml", "tcp/with_externalname.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "external.domain:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Ingress Route, externalName service with port",
			paths: []string{"tcp/services.yml", "tcp/with_externalname_with_port.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "external.domain:80",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Ingress Route, externalName service without port",
			paths: []string{"tcp/services.yml", "tcp/with_externalname_without_ports.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-bfddf9467dc588639f54",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Ingress Route with IPv6 backends",
			paths: []string{
				"services.yml", "with_ipv6.yml",
				"tcp/services.yml", "tcp/with_ipv6.yml",
				"udp/services.yml", "udp/with_ipv6.yml",
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "[fd00:10:244:0:1::3]:8080",
									},
								},
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-wrr-a8d5191cb9281af18d2d",
							Rule:        "HostSNI(`*`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-test-route-0-wrr-whoamitcp-ipv6-8080-6b332c5cb5a60aff780e",
										Weight: new(1),
									},
									{
										Name:   "default-test-route-0-wrr-external-service-with-ipv6-8080-24bdf291749302824f85",
										Weight: new(1),
									},
								},
							},
						},
						"default-test-route-0-wrr-whoamitcp-ipv6-8080-6b332c5cb5a60aff780e": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "[fd00:10:244:0:1::3]:8080",
									},
									{
										Address: "[2001:db8:85a3:8d3:1319:8a2e:370:7348]:8080",
									},
								},
							},
						},
						"default-test-route-0-wrr-external-service-with-ipv6-8080-24bdf291749302824f85": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "[fe80::200:5aee:feaa:20a2]:8080",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-wrr-a8d5191cb9281af18d2d",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-whoami-ipv6-8080-b9d3955a0d99f192038b": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://[2001:db8:85a3:8d3:1319:8a2e:370:7348]:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-external-svc-with-ipv6-8080-92969401a95b06835aa0": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://[2001:db8:85a3:8d3:1319:8a2e:370:7347]:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-ipv6-8080-b9d3955a0d99f192038b",
										Weight: new(1),
									},
									{
										Name:   "default-external-svc-with-ipv6-8080-92969401a95b06835aa0",
										Weight: new(1),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "TCP with proxyProtocol Version",
			paths: []string{"tcp/services.yml", "tcp/with_proxyprotocol.yml"},
			expected: &dynamic.Configuration{
				TLS: &dynamic.TLSConfiguration{},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
								ProxyProtocol: &dynamic.ProxyProtocol{Version: 2},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Ingress Route, empty service disallowed",
			paths: []string{"tcp/services.yml", "tcp/with_empty_services.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-bfddf9467dc588639f54",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:               "Ingress Route, empty service allowed",
			allowEmptyServices: true,
			paths:              []string{"tcp/services.yml", "tcp/with_empty_services.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.expected == nil {
				return
			}

			p := Provider{
				IngressClass:              test.ingressClass,
				AllowCrossNamespace:       true,
				AllowExternalNameServices: true,
				AllowEmptyServices:        test.allowEmptyServices,
			}

			clientMock := newClientMock(test.paths...)
			conf := p.loadConfigurationFromCRD(t.Context(), clientMock)
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadIngressRoutes(t *testing.T) {
	testCases := []struct {
		desc                    string
		ingressClass            string
		paths                   []string
		expected                *dynamic.Configuration
		allowCrossNamespace     bool
		allowEmptyServices      bool
		crossProviderNamespaces []string
	}{
		{
			desc: "Empty",
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Simple Ingress Route, with foo entrypoint",
			paths: []string{"services.yml", "simple.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                "Simple Ingress Route with middleware",
			allowCrossNamespace: true,
			paths:               []string{"services.yml", "with_middleware.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test2-route-0-f5a7fc3c74a43f56b526": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-0-lb-1b6127ca68ad95acf267",
							Rule:        "Host(`foo.com`) && PathPrefix(`/tobestripped`)",
							Priority:    12,
							Middlewares: []string{"default-stripprefix-a1d92c4f7c7ec6eb4634", "default-ratelimit-b8e74655f71c8bbfd816", "foo-addprefix-727ab48a4ad9f4b63353"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ratelimit-b8e74655f71c8bbfd816": {
							RateLimit: &dynamic.RateLimit{
								Average: 6,
								Burst:   12,
								Period:  ptypes.Duration(60 * time.Second),
								SourceCriterion: &dynamic.SourceCriterion{
									IPStrategy: &dynamic.IPStrategy{
										ExcludedIPs: []string{"127.0.0.1/32", "192.168.1.7"},
									},
								},
							},
						},
						"default-stripprefix-a1d92c4f7c7ec6eb4634": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/tobestripped"},
							},
						},
						"foo-addprefix-727ab48a4ad9f4b63353": {
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/tobeadded",
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test2-route-0-lb-1b6127ca68ad95acf267": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                "Middlewares in ingress route config are normalized",
			allowCrossNamespace: true,
			paths:               []string{"services.yml", "with_middleware_multiple_hyphens.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test2-route-0-f5a7fc3c74a43f56b526": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-0-lb-1b6127ca68ad95acf267",
							Rule:        "Host(`foo.com`) && PathPrefix(`/tobestripped`)",
							Priority:    12,
							Middlewares: []string{"default-multiple-hyphens-5c9512a3890cb0f6fbe2"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-multiple-hyphens-5c9512a3890cb0f6fbe2": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/tobestripped"},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test2-route-0-lb-1b6127ca68ad95acf267": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                    "Simple Ingress Route with middleware crossprovider",
			crossProviderNamespaces: []string{"default"},
			allowCrossNamespace:     true,
			paths:                   []string{"services.yml", "with_middleware_cross_provider.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test2-route-0-f5a7fc3c74a43f56b526": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-0-lb-1b6127ca68ad95acf267",
							Rule:        "Host(`foo.com`) && PathPrefix(`/tobestripped`)",
							Priority:    12,
							Middlewares: []string{"default-stripprefix-a1d92c4f7c7ec6eb4634", "foo-addprefix-727ab48a4ad9f4b63353", "basicauth@file", "redirect@file"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-stripprefix-a1d92c4f7c7ec6eb4634": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/tobestripped"},
							},
						},
						"foo-addprefix-727ab48a4ad9f4b63353": {
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/tobeadded",
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test2-route-0-lb-1b6127ca68ad95acf267": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "One ingress Route with two different rules",
			paths: []string{"services.yml", "with_two_rules.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-1-73e4808ba7263def92b2": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Service:     "default-test-route-1-lb-1109d6313f2dbb16a8e6",
							Priority:    14,
						},
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-1-lb-1109d6313f2dbb16a8e6": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "One ingress Route with two different services",
			paths: []string{"services.yml", "with_two_services.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-wrr-a8d5191cb9281af18d2d",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80-b08b5838443b99387085",
										Weight: new(1),
									},
									{
										Name:   "default-whoami2-8080-0b62ec73429bbf0049ac",
										Weight: new(1),
									},
								},
							},
						},
						"default-whoami-80-b08b5838443b99387085": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-whoami2-8080-0b62ec73429bbf0049ac": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "one kube service (== servers lb) in a services wrr",
			paths: []string{"with_services_lb0.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-wrr1-3b6294b40828fe78faec",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-wrr1-3b6294b40828fe78faec": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080-e336dbc96b59b3100172",
										Weight: new(1),
									},
								},
							},
						},
						"default-whoami5-8080-e336dbc96b59b3100172": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "traefik service without ingress route",
			paths: []string{"with_services_only.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-wrr1-3b6294b40828fe78faec": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080-e336dbc96b59b3100172",
										Weight: new(1),
									},
								},
							},
						},
						"default-whoami5-8080-e336dbc96b59b3100172": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "One ingress Route with two different services, each with two services, balancing servers nested",
			paths: []string{"with_services_lb1.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-wrr-a8d5191cb9281af18d2d",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-wrr1-3b6294b40828fe78faec",
										Weight: new(1),
									},
									{
										Name:   "default-wrr2-504732714158b42ee5b2",
										Weight: new(1),
									},
								},
							},
						},
						"default-wrr1-3b6294b40828fe78faec": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami4-80-57836cd12fe622b48cbc",
										Weight: new(1),
									},
									{
										Name:   "default-whoami5-8080-e336dbc96b59b3100172",
										Weight: new(1),
									},
								},
							},
						},
						"default-whoami4-80-57836cd12fe622b48cbc": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-whoami5-8080-e336dbc96b59b3100172": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-wrr2-504732714158b42ee5b2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami6-80-ec19f7af9040d4774735",
										Weight: new(1),
									},
									{
										Name:   "default-whoami7-8080-e22feebc9a2cb11bcbe6",
										Weight: new(1),
									},
								},
							},
						},
						"default-whoami6-80-ec19f7af9040d4774735": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.5:80",
									},
									{
										URL: "http://10.10.0.6:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-whoami7-8080-e22feebc9a2cb11bcbe6": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:8080",
									},
									{
										URL: "http://10.10.0.8:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "one wrr and one kube service (== servers lb) in a wrr",
			paths: []string{"with_services_lb2.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-wrr1-3b6294b40828fe78faec",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-wrr1-3b6294b40828fe78faec": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-wrr2-504732714158b42ee5b2",
										Weight: new(1),
									},
									{
										Name:   "default-whoami5-8080-e336dbc96b59b3100172",
										Weight: new(1),
									},
								},
							},
						},
						"default-wrr2-504732714158b42ee5b2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080-e336dbc96b59b3100172",
										Weight: new(1),
									},
								},
							},
						},
						"default-whoami5-8080-e336dbc96b59b3100172": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "services lb, servers lb, and mirror service, all in a wrr",
			paths: []string{"with_services_lb3.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-wrr1-3b6294b40828fe78faec",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-wrr1-3b6294b40828fe78faec": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-wrr2-504732714158b42ee5b2",
										Weight: new(1),
									},
									{
										Name:   "default-whoami5-8080-e336dbc96b59b3100172",
										Weight: new(1),
									},
									{
										Name:   "default-mirror1-27e0059f29381e4e4b93",
										Weight: new(1),
									},
								},
							},
						},
						"default-wrr2-504732714158b42ee5b2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080-e336dbc96b59b3100172",
										Weight: new(1),
									},
								},
							},
						},
						"default-mirror1-27e0059f29381e4e4b93": {
							Mirroring: &dynamic.Mirroring{
								Service: "default-whoami5-8080-e336dbc96b59b3100172",
								Mirrors: []dynamic.MirrorService{
									{Name: "default-whoami4-8080-7dc7f6765e3843670364", Percent: 50},
								},
							},
						},
						"default-whoami4-8080-7dc7f6765e3843670364": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-whoami5-8080-e336dbc96b59b3100172": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:                "services lb, servers lb, and mirror service, all in a wrr with different namespaces",
			allowCrossNamespace: true,
			paths:               []string{"with_namespaces.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-wrr-a8d5191cb9281af18d2d",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "baz-whoami6-8080-531e566e1ce14cd3ac91",
										Weight: new(1),
									},
									{
										Name:   "foo-wrr1-3f46e81fadb0e6d5865d",
										Weight: new(1),
									},
									{
										Name:   "foo-mirror2-5ad6655aa920162e663f",
										Weight: new(1),
									},
									{
										Name:   "foo-mirror3-2954533fbd26f06c9da8",
										Weight: new(1),
									},
									{
										Name:   "foo-mirror4-4a1eee1c5f7cb88b0a7d",
										Weight: new(1),
									},
								},
							},
						},
						"baz-whoami6-8080-531e566e1ce14cd3ac91": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.5:8080",
									},
									{
										URL: "http://10.10.0.6:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"foo-wrr1-3f46e81fadb0e6d5865d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "foo-whoami4-8080-bedb6203517cc982698b",
										Weight: new(1),
									},
									{
										Name:   "baz-whoami6-8080-531e566e1ce14cd3ac91",
										Weight: new(1),
									},
									{
										Name:   "foo-mirror1-cbfee77b7fea5a43b4ef",
										Weight: new(1),
									},
									{
										Name:   "bar-wrr2-05f3b853d4318465ac66",
										Weight: new(1),
									},
								},
							},
						},
						"foo-whoami4-8080-bedb6203517cc982698b": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"foo-mirror1-cbfee77b7fea5a43b4ef": {
							Mirroring: &dynamic.Mirroring{
								Service: "foo-whoami5-8080-5a253922f39b73f2d908",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-whoami4-8080-bedb6203517cc982698b"},
									{Name: "baz-whoami6-8080-531e566e1ce14cd3ac91"},
									{Name: "bar-mirrored-653b67a4c37564aca28d"},
								},
							},
						},
						"foo-whoami5-8080-5a253922f39b73f2d908": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"bar-mirrored-653b67a4c37564aca28d": {
							Mirroring: &dynamic.Mirroring{
								Service: "baz-whoami6-8080-531e566e1ce14cd3ac91",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-whoami4-8080-bedb6203517cc982698b", Percent: 50},
								},
							},
						},
						"foo-mirror2-5ad6655aa920162e663f": {
							Mirroring: &dynamic.Mirroring{
								Service: "foo-whoami5-8080-5a253922f39b73f2d908",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-whoami4-8080-bedb6203517cc982698b"},
									{Name: "baz-whoami6-8080-531e566e1ce14cd3ac91"},
									{Name: "bar-mirrored-653b67a4c37564aca28d"},
									{Name: "foo-wrr1-3f46e81fadb0e6d5865d"},
								},
							},
						},
						"foo-mirror3-2954533fbd26f06c9da8": {
							Mirroring: &dynamic.Mirroring{
								Service: "foo-wrr1-3f46e81fadb0e6d5865d",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-whoami4-8080-bedb6203517cc982698b"},
									{Name: "baz-whoami6-8080-531e566e1ce14cd3ac91"},
									{Name: "bar-mirrored-653b67a4c37564aca28d"},
									{Name: "foo-wrr1-3f46e81fadb0e6d5865d"},
								},
							},
						},
						"bar-wrr2-05f3b853d4318465ac66": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "foo-whoami5-8080-5a253922f39b73f2d908",
										Weight: new(1),
									},
								},
							},
						},
						"foo-mirror4-4a1eee1c5f7cb88b0a7d": {
							Mirroring: &dynamic.Mirroring{
								Service: "foo-wrr1-3f46e81fadb0e6d5865d",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-whoami4-8080-bedb6203517cc982698b"},
									{Name: "baz-whoami6-8080-531e566e1ce14cd3ac91"},
									{Name: "bar-mirrored-653b67a4c37564aca28d"},
									{Name: "foo-wrr1-3f46e81fadb0e6d5865d"},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "one kube service (== servers lb) in a mirroring",
			paths: []string{"with_mirroring.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-mirror1-27e0059f29381e4e4b93",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-mirror1-27e0059f29381e4e4b93": {
							Mirroring: &dynamic.Mirroring{
								Service: "default-whoami5-8080-e336dbc96b59b3100172",
								Mirrors: []dynamic.MirrorService{
									{Name: "default-whoami4-8080-7dc7f6765e3843670364", Percent: 50},
								},
							},
						},
						"default-whoami4-8080-7dc7f6765e3843670364": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-whoami5-8080-e336dbc96b59b3100172": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "weighted services in a mirroring",
			paths: []string{"with_mirroring2.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-mirror1-27e0059f29381e4e4b93",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-mirror1-27e0059f29381e4e4b93": {
							Mirroring: &dynamic.Mirroring{
								Service: "default-wrr1-3b6294b40828fe78faec",
								Mirrors: []dynamic.MirrorService{
									{Name: "default-wrr2-504732714158b42ee5b2", Percent: 30},
								},
							},
						},
						"default-wrr1-3b6294b40828fe78faec": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami4-8080-7dc7f6765e3843670364",
										Weight: new(1),
									},
								},
							},
						},
						"default-wrr2-504732714158b42ee5b2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080-e336dbc96b59b3100172",
										Weight: new(1),
									},
								},
							},
						},
						"default-whoami4-8080-7dc7f6765e3843670364": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-whoami5-8080-e336dbc96b59b3100172": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "One ingress Route with two different services, with weights",
			paths: []string{"services.yml", "with_two_services_weight.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-wrr-a8d5191cb9281af18d2d",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80-b08b5838443b99387085",
										Weight: new(10),
									},
									{
										Name:   "default-whoami2-8080-0b62ec73429bbf0049ac",
										Weight: new(0),
									},
								},
							},
						},
						"default-whoami-80-b08b5838443b99387085": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-whoami2-8080-0b62ec73429bbf0049ac": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:         "Ingress class",
			paths:        []string{"services.yml", "simple.yml"},
			ingressClass: "tchouk",
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Route with empty rule value is ignored",
			paths: []string{"services.yml", "with_no_rule_value.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Route with kind not of a rule type (empty kind) is ignored",
			paths: []string{"services.yml", "with_wrong_rule_kind.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS",
			paths: []string{"services.yml", "with_tls.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							Certificate: tls.Certificate{
								CertFile: tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  tls.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with tls options",
			paths: []string{"services.yml", "with_tls_options.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"default-foo-ce203de49508e8c70b06": {
							MinVersion: "VersionTLS12",
							CipherSuites: []string{
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
							},
							ClientAuth: tls.ClientAuth{
								CAFiles: []tls.FileOrContent{
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								ClientAuthType: "VerifyClientCertIfGiven",
							},
							SniStrict: true,
							ALPNProtocols: []string{
								"h2",
								"http/1.1",
								"acme-tls/1",
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-foo-ce203de49508e8c70b06",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with two default tls options",
			paths: []string{"services.yml", "with_default_tls_options.yml", "with_default_tls_options_default_namespace.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-foo-ce203de49508e8c70b06",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with default tls options",
			paths: []string{"services.yml", "with_default_tls_options.yml"},
			expected: &dynamic.Configuration{
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"default": {
							MinVersion: "VersionTLS12",
							CipherSuites: []string{
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
							},
							ClientAuth: tls.ClientAuth{
								CAFiles: []tls.FileOrContent{
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								ClientAuthType: "VerifyClientCertIfGiven",
							},
							SniStrict: true,
							ALPNProtocols: []string{
								"h2",
								"http/1.1",
								"acme-tls/1",
							},
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-foo-ce203de49508e8c70b06",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:                "TLS with tls options and specific namespace",
			paths:               []string{"services.yml", "with_tls_options_and_specific_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"myns-foo-d809e81d5ccf461aab60": {
							MinVersion: "VersionTLS12",
							CipherSuites: []string{
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
							},
							ClientAuth: tls.ClientAuth{
								CAFiles: []tls.FileOrContent{
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								ClientAuthType: "VerifyClientCertIfGiven",
							},
							SniStrict: true,
							ALPNProtocols: []string{
								"h2",
								"http/1.1",
								"acme-tls/1",
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "myns-foo-d809e81d5ccf461aab60",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with bad tls options",
			paths: []string{"services.yml", "with_bad_tls_options.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"default-foo-ce203de49508e8c70b06": {
							MinVersion: "VersionTLS12",
							CipherSuites: []string{
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_RSA_WITH_AES_256_GCM_SHA384",
							},
							ClientAuth: tls.ClientAuth{
								CAFiles: []tls.FileOrContent{
									tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								ClientAuthType: "VerifyClientCertIfGiven",
							},
							SniStrict: true,
							ALPNProtocols: []string{
								"h2",
								"http/1.1",
								"acme-tls/1",
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-foo-ce203de49508e8c70b06",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with unknown tls options",
			paths: []string{"services.yml", "with_unknown_tls_options.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"default-foo-ce203de49508e8c70b06": {
							MinVersion: "VersionTLS12",
							ALPNProtocols: []string{
								"h2",
								"http/1.1",
								"acme-tls/1",
							},
							CipherSuites: []string{
								"TLS_AES_128_GCM_SHA256",
								"TLS_AES_256_GCM_SHA384",
								"TLS_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-unknown-d612e69053993bc351c4",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:                "TLS with unknown tls options namespace",
			paths:               []string{"services.yml", "with_unknown_tls_options_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"default-foo-ce203de49508e8c70b06": {
							MinVersion: "VersionTLS12",
							ALPNProtocols: []string{
								"h2",
								"http/1.1",
								"acme-tls/1",
							},
							CipherSuites: []string{
								"TLS_AES_128_GCM_SHA256",
								"TLS_AES_256_GCM_SHA384",
								"TLS_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "unknown-foo-60df693b64851d2ececa",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with ACME",
			paths: []string{"services.yml", "with_tls_acme.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Simple Ingress Route, defaulting to https for servers",
			paths: []string{"services.yml", "with_https_default.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.5:8443",
									},
									{
										URL: "https://10.10.0.6:8443",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Simple Ingress Route, explicit https scheme",
			paths: []string{"services.yml", "with_https_scheme.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.7:8443",
									},
									{
										URL: "https://10.10.0.8:8443",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Simple Ingress Route, with basic auth middleware",
			paths: []string{"services.yml", "with_auth.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{
						"default-basicauth-4fc831a329fbed27112a": {
							BasicAuth: &dynamic.BasicAuth{
								Users: dynamic.Users{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							},
						},
						"default-digestauth-b0bf024b07b03f2233b6": {
							DigestAuth: &dynamic.DigestAuth{
								Users: dynamic.Users{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							},
						},
						"default-forwardauth-e693d333fa61d6830f14": {
							ForwardAuth: &dynamic.ForwardAuth{
								Address: "test.com",
								TLS: &types.ClientTLS{
									CA:   "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----",
									Cert: "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----",
									Key:  "-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----",
								},
							},
						},
					},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Simple Ingress Route, with test middleware read config from secret",
			paths: []string{"services.yml", "with_plugin_read_secret.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{
						"default-test-secret-23fb605c17af75d7b87e": {
							Plugin: map[string]dynamic.PluginConf{
								"test-secret": map[string]any{
									"user":   "admin",
									"secret": "this_is_the_secret",
								},
							},
						},
					},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Simple Ingress Route, with test middleware read config from deep secret",
			paths: []string{"services.yml", "with_plugin_deep_read_secret.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{
						"default-test-secret-23fb605c17af75d7b87e": {
							Plugin: map[string]dynamic.PluginConf{
								"test-secret": map[string]any{
									"secret_0": map[string]any{
										"secret_1": map[string]any{
											"secret_2": map[string]any{
												"user":   "admin",
												"secret": "this_is_the_very_deep_secret",
											},
										},
									},
								},
							},
						},
					},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Simple Ingress Route, with test middleware read config from an array of secret",
			paths: []string{"services.yml", "with_plugin_read_array_of_secret.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{
						"default-test-secret-23fb605c17af75d7b87e": {
							Plugin: map[string]dynamic.PluginConf{
								"test-secret": map[string]any{
									"secret": []any{"secret_data1", "secret_data2"},
								},
							},
						},
					},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Simple Ingress Route, with test middleware read config from an array of secret",
			paths: []string{"services.yml", "with_plugin_read_array_of_map_contain_secret.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{
						"default-test-secret-23fb605c17af75d7b87e": {
							Plugin: map[string]dynamic.PluginConf{
								"test-secret": map[string]any{
									"users": []any{
										map[string]any{
											"name":   "admin",
											"secret": "admin_password",
										},
										map[string]any{
											"name":   "user",
											"secret": "user_password",
										},
									},
								},
							},
						},
					},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:                "Simple Ingress Route, with test middleware read config from secret that not found",
			paths:               []string{"services.yml", "with_plugin_read_not_exist_secret.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Simple Ingress Route, with error page middleware",
			paths: []string{"services.yml", "with_error_page.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{
						"default-errorpage-28845717dbca24eb7eb9": {
							Errors: &dynamic.ErrorPage{
								Status:              []string{"404", "500"},
								Service:             "default-errorpage-28845717dbca24eb7eb9-errorpage-service",
								Query:               "query",
								ErrorRequestHeaders: []string{"X-Foo-Bar", "X-Bar-Foo"},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-errorpage-28845717dbca24eb7eb9-errorpage-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Error page middleware referencing a TraefikService",
			paths: []string{"services.yml", "with_error_page_traefik_service.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{
						"default-errorpage-28845717dbca24eb7eb9": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "500"},
								Service: "default-errorpage-wrr-9e4f6c31dcb0d29bd174",
								Query:   "query",
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-errorpage-wrr-9e4f6c31dcb0d29bd174": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80-b08b5838443b99387085",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoami-80-b08b5838443b99387085": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Simple Ingress Route, with options",
			paths: []string{"services.yml", "with_options.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader:     new(false),
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: "10s"},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "TLS with tls store",
			paths: []string{"services.yml", "with_tls_store.yml"},
			expected: &dynamic.Configuration{
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{
						"default": {
							DefaultCertificate: &tls.Certificate{
								CertFile: tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  tls.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with tls store containing certificates",
			paths: []string{"services.yml", "with_tls_store_certificates.yml"},
			expected: &dynamic.Configuration{
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							Certificate: tls.Certificate{
								CertFile: tls.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  tls.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
							Stores: []string{"default"},
						},
					},
					Stores: map[string]tls.Store{
						"default": {},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with tls store default two times",
			paths: []string{"services.yml", "with_tls_store.yml", "with_default_tls_store.yml"},
			expected: &dynamic.Configuration{
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "port selected by name (TODO)",
		},
		{
			desc:  "Simple Ingress Route, with externalName service",
			paths: []string{"services.yml", "with_externalname.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://external.domain:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Ingress Route, externalName service with http",
			paths: []string{"services.yml", "with_externalname_with_http.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://external.domain:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Ingress Route, externalName service with https",
			paths: []string{"services.yml", "with_externalname_with_https.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Ingress Route, externalName service without ports",
			paths: []string{"services.yml", "with_externalname_without_ports.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "ServersTransport",
			paths: []string{"services.yml", "with_servers_transport.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{
						"foo-test-0bbbd89c878c4df00916": {
							ServerName:         "test",
							InsecureSkipVerify: true,
							RootCAs:            []tls.FileOrContent{"TESTROOTCAS0", "TESTROOTCAS1", "TESTROOTCAS2", "TESTROOTCAS3", "TESTROOTCAS5", "TESTALLCERTS"},
							Certificates: tls.Certificates{
								{CertFile: "TESTCERT1", KeyFile: "TESTKEY1"},
								{CertFile: "TESTCERT2", KeyFile: "TESTKEY2"},
								{CertFile: "TESTCERT3", KeyFile: "TESTKEY3"},
							},
							MaxIdleConnsPerHost: 42,
							DisableHTTP2:        true,
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:           ptypes.Duration(42 * time.Second),
								ResponseHeaderTimeout: ptypes.Duration(42 * time.Second),
								IdleConnTimeout:       ptypes.Duration(42 * time.Millisecond),
								ReadIdleTimeout:       ptypes.Duration(42 * time.Second),
								PingTimeout:           ptypes.Duration(42 * time.Second),
							},
							PeerCertURI: "foo://bar",
						},
						"default-test-e6d80538b02592c965d8": {
							ServerName: "test",
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(30 * time.Second),
								IdleConnTimeout: ptypes.Duration(90 * time.Second),
								PingTimeout:     ptypes.Duration(15 * time.Second),
							},
						},
					},
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-wrr-a8d5191cb9281af18d2d",
							Rule:        "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-external-svc-with-https-443-de2ffc682f4e9ff377b5": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader:   new(true),
								ServersTransport: "default-test-e6d80538b02592c965d8",
							},
						},
						"default-whoamitls-443-516a8130196c6d0c480c": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.5:8443",
									},
									{
										URL: "https://10.10.0.6:8443",
									},
								},
								PassHostHeader:   new(true),
								ServersTransport: "default-default-test-592f4763608d32419a4e",
							},
						},
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-external-svc-with-https-443-de2ffc682f4e9ff377b5",
										Weight: new(1),
									},
									{
										Name:   "default-whoamitls-443-516a8130196c6d0c480c",
										Weight: new(1),
									},
								},
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Ingress Route, empty service disallowed",
			paths: []string{"services.yml", "with_empty_services.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:               "Ingress Route, empty service allowed",
			allowEmptyServices: true,
			paths:              []string{"services.yml", "with_empty_services.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:               "IngressRoute, service with multiple subsets",
			allowEmptyServices: true,
			paths:              []string{"services.yml", "with_multiple_subsets.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:               "TraefikService, empty service allowed",
			allowEmptyServices: true,
			paths:              []string{"services.yml", "with_empty_services_ts.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Middlewares: []string{"default-test-errorpage-e1e939ab3bb71fe16185"},
							Service:     "default-test-route-0-wrr-a8d5191cb9281af18d2d",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-test-errorpage-e1e939ab3bb71fe16185": {
							Errors: &dynamic.ErrorPage{
								Service: "default-test-errorpage-e1e939ab3bb71fe16185-errorpage-service",
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-test-weighted-4b1c2d77727f0755b988",
										Weight: new(1),
									},
									{
										Name:   "default-test-mirror-52e846c6f16eac7e8401",
										Weight: new(1),
									},
								},
							},
						},
						"default-test-errorpage-e1e939ab3bb71fe16185-errorpage-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								PassHostHeader: new(true),
							},
						},
						"default-test-weighted-4b1c2d77727f0755b988": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-without-endpoints-subsets-80-60a5bd422b9b56cc78ef",
										Weight: new(1),
									},
								},
							},
						},
						"default-test-mirror-52e846c6f16eac7e8401": {
							Mirroring: &dynamic.Mirroring{
								Service: "default-whoami-without-endpoints-subsets-80-60a5bd422b9b56cc78ef",
								Mirrors: []dynamic.MirrorService{
									{
										Name: "default-whoami-without-endpoints-subsets-80-60a5bd422b9b56cc78ef",
									},
									{
										Name: "default-test-weighted-4b1c2d77727f0755b988",
									},
								},
							},
						},
						"default-whoami-without-endpoints-subsets-80-60a5bd422b9b56cc78ef": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.expected == nil {
				return
			}

			p := Provider{
				IngressClass:              test.ingressClass,
				AllowCrossNamespace:       test.allowCrossNamespace,
				AllowExternalNameServices: true,
				AllowEmptyServices:        test.allowEmptyServices,
				CrossProviderNamespaces:   test.crossProviderNamespaces,
			}

			clientMock := newClientMock(test.paths...)
			conf := p.loadConfigurationFromCRD(t.Context(), clientMock)
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadIngressRouteUDPs(t *testing.T) {
	testCases := []struct {
		desc               string
		ingressClass       string
		paths              []string
		allowEmptyServices bool
		expected           *dynamic.Configuration
	}{
		{
			desc: "Empty",
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Simple Ingress Route, with foo entrypoint",
			paths: []string{"udp/services.yml", "udp/simple.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "One ingress Route with two different routes",
			paths: []string{"udp/services.yml", "udp/with_two_routes.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
						},
						"default-test-route-1-73e4808ba7263def92b2": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-1-lb-1109d6313f2dbb16a8e6",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
						"default-test-route-1-lb-1109d6313f2dbb16a8e6": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "10.10.0.3:8080",
									},
									{
										Address: "10.10.0.4:8080",
									},
								},
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "One ingress Route with two different services",
			paths: []string{"udp/services.yml", "udp/with_two_services.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-wrr-a8d5191cb9281af18d2d",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.UDPWeightedRoundRobin{
								Services: []dynamic.UDPWRRService{
									{
										Name:   "default-test-route-0-wrr-whoamiudp-8000-238c381c71b733803182",
										Weight: new(2),
									},
									{
										Name:   "default-test-route-0-wrr-whoamiudp2-8080-9d4476593bfbc7db55b0",
										Weight: new(3),
									},
								},
							},
						},
						"default-test-route-0-wrr-whoamiudp-8000-238c381c71b733803182": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
						"default-test-route-0-wrr-whoamiudp2-8080-9d4476593bfbc7db55b0": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "10.10.0.3:8080",
									},
									{
										Address: "10.10.0.4:8080",
									},
								},
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "One ingress Route with different services namespaces",
			paths: []string{"udp/services.yml", "udp/with_different_services_ns.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-wrr-a8d5191cb9281af18d2d",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.UDPWeightedRoundRobin{
								Services: []dynamic.UDPWRRService{
									{
										Name:   "default-test-route-0-wrr-whoamiudp-8000-238c381c71b733803182",
										Weight: new(2),
									},
									{
										Name:   "default-test-route-0-wrr-whoamiudp2-8080-9d4476593bfbc7db55b0",
										Weight: new(3),
									},
									{
										Name:   "default-test-route-0-wrr-whoamiudp3-8083-cec5196fa18b6005f2e2",
										Weight: new(4),
									},
								},
							},
						},
						"default-test-route-0-wrr-whoamiudp-8000-238c381c71b733803182": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
						"default-test-route-0-wrr-whoamiudp2-8080-9d4476593bfbc7db55b0": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "10.10.0.3:8080",
									},
									{
										Address: "10.10.0.4:8080",
									},
								},
							},
						},
						"default-test-route-0-wrr-whoamiudp3-8083-cec5196fa18b6005f2e2": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "10.10.0.7:8083",
									},
									{
										Address: "10.10.0.8:8083",
									},
								},
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Simple Ingress Route, with externalName service",
			paths: []string{"udp/services.yml", "udp/with_externalname.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "external.domain:8000",
									},
								},
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Ingress Route, externalName service with port",
			paths: []string{"udp/services.yml", "udp/with_externalname_with_port.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "external.domain:80",
									},
								},
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Ingress Route, externalName service without port",
			paths: []string{"udp/services.yml", "udp/with_externalname_without_ports.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-bfddf9467dc588639f54",
						},
					},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:         "Ingress class does not match",
			paths:        []string{"udp/services.yml", "udp/simple.yml"},
			ingressClass: "tchouk",
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Ingress Route, empty service disallowed",
			paths: []string{"udp/services.yml", "udp/with_empty_services.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-bfddf9467dc588639f54",
						},
					},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:               "Ingress Route, empty service allowed",
			allowEmptyServices: true,
			paths:              []string{"udp/services.yml", "udp/with_empty_services.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.expected == nil {
				return
			}

			p := Provider{
				IngressClass:              test.ingressClass,
				AllowCrossNamespace:       true,
				AllowExternalNameServices: true,
				AllowEmptyServices:        test.allowEmptyServices,
			}

			clientMock := newClientMock(test.paths...)
			conf := p.loadConfigurationFromCRD(t.Context(), clientMock)
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestParseServiceProtocol(t *testing.T) {
	testCases := []struct {
		desc          string
		scheme        string
		portName      string
		portNumber    int32
		expected      string
		expectedError bool
	}{
		{
			desc:       "Empty scheme and name",
			scheme:     "",
			portName:   "",
			portNumber: 1000,
			expected:   "http",
		},
		{
			desc:       "h2c scheme and emptyname",
			scheme:     "h2c",
			portName:   "",
			portNumber: 1000,
			expected:   "h2c",
		},
		{
			desc:          "invalid scheme",
			scheme:        "foo",
			portName:      "",
			portNumber:    1000,
			expectedError: true,
		},
		{
			desc:       "Empty scheme and https name",
			scheme:     "",
			portName:   "https-secure",
			portNumber: 1000,
			expected:   "https",
		},
		{
			desc:       "Empty scheme and port number",
			scheme:     "",
			portName:   "",
			portNumber: 443,
			expected:   "https",
		},
		{
			desc:       "https scheme",
			scheme:     "https",
			portName:   "",
			portNumber: 1000,
			expected:   "https",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			protocol, err := parseServiceProtocol(test.scheme, test.portName, test.portNumber)
			if test.expectedError {
				assert.Error(t, err)
			} else {
				assert.Equal(t, test.expected, protocol)
			}
		})
	}
}

func TestGetServicePort(t *testing.T) {
	testCases := []struct {
		desc        string
		svc         *corev1.Service
		port        intstr.IntOrString
		expected    *corev1.ServicePort
		expectError bool
	}{
		{
			desc:        "Basic",
			expectError: true,
		},
		{
			desc: "Matching ports, with no service type",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port: 80,
						},
					},
				},
			},
			port: intstr.FromInt(80),
			expected: &corev1.ServicePort{
				Port: 80,
			},
		},
		{
			desc: "Matching ports 0",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{},
					},
				},
			},
			expectError: true,
		},
		{
			desc: "Matching ports 0 (with external name)",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeExternalName,
					Ports: []corev1.ServicePort{
						{},
					},
				},
			},
			expectError: true,
		},
		{
			desc: "Matching named port",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name: "http",
							Port: 80,
						},
					},
				},
			},
			port: intstr.FromString("http"),
			expected: &corev1.ServicePort{
				Name: "http",
				Port: 80,
			},
		},
		{
			desc: "Matching named port (with external name)",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeExternalName,
					Ports: []corev1.ServicePort{
						{
							Name: "http",
							Port: 80,
						},
					},
				},
			},
			port: intstr.FromString("http"),
			expected: &corev1.ServicePort{
				Name: "http",
				Port: 80,
			},
		},
		{
			desc: "Mismatching, only port(Ingress) defined",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{},
			},
			port:        intstr.FromInt(80),
			expectError: true,
		},
		{
			desc: "Mismatching, only named port(Ingress) defined",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{},
			},
			port:        intstr.FromString("http"),
			expectError: true,
		},
		{
			desc: "Mismatching, only port(Ingress) defined with external name",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeExternalName,
				},
			},
			port: intstr.FromInt(80),
			expected: &corev1.ServicePort{
				Port: 80,
			},
		},
		{
			desc: "Mismatching, only named port(Ingress) defined with external name",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeExternalName,
				},
			},
			port:        intstr.FromString("http"),
			expectError: true,
		},
		{
			desc: "Mismatching, only Service port defined",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port: 80,
						},
					},
				},
			},
			expectError: true,
		},
		{
			desc: "Mismatching, only Service port defined with external name",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeExternalName,
					Ports: []corev1.ServicePort{
						{
							Port: 80,
						},
					},
				},
			},
			expectError: true,
		},
		{
			desc: "Two different ports defined",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port: 80,
						},
					},
				},
			},
			port:        intstr.FromInt(443),
			expectError: true,
		},
		{
			desc: "Two different named ports defined",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name: "foo",
							Port: 80,
						},
					},
				},
			},
			port:        intstr.FromString("bar"),
			expectError: true,
		},
		{
			desc: "Two different ports defined (with external name)",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeExternalName,
					Ports: []corev1.ServicePort{
						{
							Port: 80,
						},
					},
				},
			},
			port: intstr.FromInt(443),
			expected: &corev1.ServicePort{
				Port: 443,
			},
		},
		{
			desc: "Two different named ports defined (with external name)",
			svc: &corev1.Service{
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeExternalName,
					Ports: []corev1.ServicePort{
						{
							Name: "foo",
							Port: 80,
						},
					},
				},
			},
			port:        intstr.FromString("bar"),
			expectError: true,
		},
	}
	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			actual, err := getServicePort(test.svc, test.port)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.Equal(t, test.expected, actual)
			}
		})
	}
}

func TestCrossNamespace(t *testing.T) {
	testCases := []struct {
		desc                string
		allowCrossNamespace bool
		ingressClass        string
		paths               []string
		expected            *dynamic.Configuration
	}{
		{
			desc: "Empty",
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "HTTP middleware cross namespace disallowed",
			paths: []string{"services.yml", "with_middleware_cross_namespace.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-crossnamespace-route-1-262ab643fa6e14e5f5e8": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-1-lb-0ce7ef3cc5e5221ad93f",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bir`)",
							Priority:    12,
							Middlewares: []string{"default-test-errorpage-e1e939ab3bb71fe16185"},
						},
						"default-test-crossnamespace-route-3-777508f24b1bbecc80f3": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-3-lb-37e219f6ea3be495ffa7",
							Rule:        "Host(`foo.com`) && PathPrefix(`/chain`)",
							Priority:    12,
							Middlewares: []string{
								"default-test-chain-88cb3a598b2461561eca",
								"default-test-chain-cross-provider-1d4a50816120e573b00c",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"cross-ns-stripprefix-6977af38b2936444fec6": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes:   []string{"/stripit"},
								ForceSlash: false,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test-crossnamespace-route-1-lb-0ce7ef3cc5e5221ad93f": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-test-crossnamespace-route-3-lb-37e219f6ea3be495ffa7": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                "HTTP middleware cross namespace allowed",
			paths:               []string{"services.yml", "with_middleware_cross_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-crossnamespace-route-0-30e803916cecd24023a7": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-0-lb-7c97e893920902f4c4fb",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							Middlewares: []string{
								"cross-ns-stripprefix-6977af38b2936444fec6",
							},
						},
						"default-test-crossnamespace-route-1-262ab643fa6e14e5f5e8": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-1-lb-0ce7ef3cc5e5221ad93f",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bir`)",
							Priority:    12,
							Middlewares: []string{"default-test-errorpage-e1e939ab3bb71fe16185"},
						},
						"default-test-crossnamespace-route-2-34212371bf9022fddc6d": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-2-lb-ebbed47881305fcd1030",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bur`)",
							Priority:    12,
							Middlewares: []string{"cross-ns-stripprefix@kubernetescrd"},
						},
						"default-test-crossnamespace-route-3-777508f24b1bbecc80f3": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-3-lb-37e219f6ea3be495ffa7",
							Rule:        "Host(`foo.com`) && PathPrefix(`/chain`)",
							Priority:    12,
							Middlewares: []string{
								"default-test-chain-88cb3a598b2461561eca",
								"default-test-chain-cross-provider-1d4a50816120e573b00c",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"cross-ns-stripprefix-6977af38b2936444fec6": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes:   []string{"/stripit"},
								ForceSlash: false,
							},
						},
						"default-test-errorpage-e1e939ab3bb71fe16185": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"500-599"},
								Service: "default-test-errorpage-e1e939ab3bb71fe16185-errorpage-service",
								Query:   "/{status}.html",
							},
						},
						"default-test-chain-88cb3a598b2461561eca": {
							Chain: &dynamic.Chain{
								Middlewares: []string{"cross-ns-stripprefix-6977af38b2936444fec6"},
							},
						},
						"default-test-chain-cross-provider-1d4a50816120e573b00c": {
							Chain: &dynamic.Chain{
								Middlewares: []string{"other-middleware@kubernetescrd"},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test-crossnamespace-route-0-lb-7c97e893920902f4c4fb": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-test-crossnamespace-route-1-lb-0ce7ef3cc5e5221ad93f": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-test-errorpage-e1e939ab3bb71fe16185-errorpage-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-test-crossnamespace-route-2-lb-ebbed47881305fcd1030": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-test-crossnamespace-route-3-lb-37e219f6ea3be495ffa7": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                "HTTP cross namespace allowed",
			paths:               []string{"services.yml", "with_cross_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-cross-ns-route-0-43e82ad3b9d755894bb8": {
							EntryPoints: []string{"foo"},
							Service:     "default-cross-ns-route-0-wrr-98f7f5a30f2f29b81fa7",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
						"default-cross-ns-route-1-0ed85e47bfdaf9c429a2": {
							EntryPoints: []string{"foo"},
							Service:     "default-cross-ns-route-1-lb-8bdc5a04b2a40bf59b5e",
							Rule:        "Host(`bar.com`) && PathPrefix(`/foo`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-cross-ns-route-0-wrr-98f7f5a30f2f29b81fa7": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80-6a7a2b78fd9b52601354",
										Weight: new(1),
									},
									{
										Name:   "default-tr-svc-wrr1-65f91db84ed92e091532",
										Weight: new(1),
									},
									{
										Name:   "cross-ns-tr-svc-wrr2-35846c0f6a99ca063e5c",
										Weight: new(1),
									},
									{
										Name:   "default-tr-svc-mirror1-8130c8b7f004b659eeb3",
										Weight: new(1),
									},
									{
										Name:   "cross-ns-tr-svc-mirror2-3a1eebdec12b22b77a0f",
										Weight: new(1),
									},
								},
							},
						},
						"default-cross-ns-route-1-lb-8bdc5a04b2a40bf59b5e": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader:   new(true),
								ServersTransport: "foo-test@kubernetescrd",
							},
						},
						"cross-ns-whoami-svc-80-6a7a2b78fd9b52601354": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"default-tr-svc-wrr1-65f91db84ed92e091532": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80-6a7a2b78fd9b52601354",
										Weight: new(1),
									},
								},
							},
						},
						"cross-ns-tr-svc-wrr2-35846c0f6a99ca063e5c": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80-6a7a2b78fd9b52601354",
										Weight: new(1),
									},
								},
							},
						},
						"default-tr-svc-mirror1-8130c8b7f004b659eeb3": {
							Mirroring: &dynamic.Mirroring{
								Service: "default-whoami-80-b08b5838443b99387085",
								Mirrors: []dynamic.MirrorService{
									{
										Name:    "cross-ns-whoami-svc-80-6a7a2b78fd9b52601354",
										Percent: 20,
									},
								},
							},
						},
						"cross-ns-tr-svc-mirror2-3a1eebdec12b22b77a0f": {
							Mirroring: &dynamic.Mirroring{
								Service: "cross-ns-whoami-svc-80-6a7a2b78fd9b52601354",
								Mirrors: []dynamic.MirrorService{
									{
										Name:    "cross-ns-whoami-svc-80-6a7a2b78fd9b52601354",
										Percent: 20,
									},
								},
							},
						},
						"default-whoami-80-b08b5838443b99387085": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "HTTP cross namespace disallowed",
			paths: []string{"services.yml", "with_cross_namespace.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"cross-ns-tr-svc-wrr2-35846c0f6a99ca063e5c": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80-6a7a2b78fd9b52601354",
										Weight: new(1),
									},
								},
							},
						},
						"cross-ns-whoami-svc-80-6a7a2b78fd9b52601354": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
						"cross-ns-tr-svc-mirror2-3a1eebdec12b22b77a0f": {
							Mirroring: &dynamic.Mirroring{
								Service: "cross-ns-whoami-svc-80-6a7a2b78fd9b52601354",
								Mirrors: []dynamic.MirrorService{
									{
										Name:    "cross-ns-whoami-svc-80-6a7a2b78fd9b52601354",
										Percent: 20,
									},
								},
							},
						},
						"default-whoami-80-b08b5838443b99387085": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                "HTTP ServersTransport cross namespace allowed",
			paths:               []string{"services.yml", "with_servers_transport_cross_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader:   new(true),
								ServersTransport: "cross-ns-st-cross-ns@kubernetescrd",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"cross-ns-st-cross-ns-77710ecac094a47836bf": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:           30000000000,
								ResponseHeaderTimeout: 0,
								IdleConnTimeout:       90000000000,
								ReadIdleTimeout:       0,
								PingTimeout:           15000000000,
							},
							DisableHTTP2: true,
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "HTTP ServersTransport cross namespace disallowed",
			paths: []string{"services.yml", "with_servers_transport_cross_namespace.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services:    map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"cross-ns-st-cross-ns-77710ecac094a47836bf": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:           30000000000,
								ResponseHeaderTimeout: 0,
								IdleConnTimeout:       90000000000,
								ReadIdleTimeout:       0,
								PingTimeout:           15000000000,
							},
							DisableHTTP2: true,
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                "HTTP Service cross namespace via @kubernetescrd suffix allowed",
			paths:               []string{"services.yml", "with_service_cross_namespace_kubernetescrd.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "cross-ns-tr-svc-ea661fba252322e8fd8e",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"cross-ns-tr-svc-b2116afd1555971d704d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80-6a7a2b78fd9b52601354",
										Weight: new(1),
									},
								},
							},
						},
						"cross-ns-whoami-svc-80-6a7a2b78fd9b52601354": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "HTTP Service cross namespace via @kubernetescrd suffix disallowed",
			paths: []string{"services.yml", "with_service_cross_namespace_kubernetescrd.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"cross-ns-tr-svc-b2116afd1555971d704d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80-6a7a2b78fd9b52601354",
										Weight: new(1),
									},
								},
							},
						},
						"cross-ns-whoami-svc-80-6a7a2b78fd9b52601354": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                "HTTP TLSOption cross namespace allowed",
			paths:               []string{"services.yml", "with_tls_options_cross_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "cross-ns-tls-options-cn-843f4bfed2e94b4d9396",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"cross-ns-tls-options-cn-843f4bfed2e94b4d9396": {
							MinVersion:    "VersionTLS12",
							ALPNProtocols: []string{"h2", "http/1.1", "acme-tls/1"},
							CipherSuites: []string{
								"TLS_AES_128_GCM_SHA256",
								"TLS_AES_256_GCM_SHA384",
								"TLS_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
							},
						},
					},
				},
			},
		},
		{
			desc:                "HTTP TLSOption cross namespace disallowed",
			paths:               []string{"services.yml", "with_tls_options_cross_namespace.yml"},
			allowCrossNamespace: false,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"cross-ns-tls-options-cn-843f4bfed2e94b4d9396": {
							MinVersion:    "VersionTLS12",
							ALPNProtocols: []string{"h2", "http/1.1", "acme-tls/1"},
							CipherSuites: []string{
								"TLS_AES_128_GCM_SHA256",
								"TLS_AES_256_GCM_SHA384",
								"TLS_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
							},
						},
					},
				},
			},
		},
		{
			desc:  "TCP middleware cross namespace disallowed",
			paths: []string{"tcp/services.yml", "tcp/with_middleware_with_cross_namespace.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Middlewares: []string{"default-ipwhitelist-72ecdadecb80eec37e7b"},
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipwhitelist-72ecdadecb80eec37e7b": {
							IPWhiteList: &dynamic.TCPIPWhiteList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						"cross-ns-ipwhitelist-161d1949c9328b78aa30": {
							IPWhiteList: &dynamic.TCPIPWhiteList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                "TCP middleware cross namespace allowed",
			paths:               []string{"tcp/services.yml", "tcp/with_middleware_with_cross_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Middlewares: []string{"default-ipwhitelist-72ecdadecb80eec37e7b"},
							Rule:        "HostSNI(`foo.com`)",
						},
						"default-test-route-1-73e4808ba7263def92b2": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-1-lb-1109d6313f2dbb16a8e6",
							Middlewares: []string{"cross-ns-ipwhitelist-161d1949c9328b78aa30"},
							Rule:        "HostSNI(`bar.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipwhitelist-72ecdadecb80eec37e7b": {
							IPWhiteList: &dynamic.TCPIPWhiteList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						"cross-ns-ipwhitelist-161d1949c9328b78aa30": {
							IPWhiteList: &dynamic.TCPIPWhiteList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-1-lb-1109d6313f2dbb16a8e6": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                "TCP cross namespace allowed",
			paths:               []string{"tcp/services.yml", "tcp/with_cross_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "TCP cross namespace disallowed",
			paths: []string{"tcp/services.yml", "tcp/with_cross_namespace.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					// The router that references the invalid service will be discarded.
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-bfddf9467dc588639f54",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                "TCP TLSOption cross namespace allowed",
			paths:               []string{"tcp/services.yml", "tcp/with_tls_options_cross_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "cross-ns-tls-options-cn-843f4bfed2e94b4d9396",
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"cross-ns-tls-options-cn-843f4bfed2e94b4d9396": {
							MinVersion:    "VersionTLS12",
							ALPNProtocols: []string{"h2", "http/1.1", "acme-tls/1"},
							CipherSuites: []string{
								"TLS_AES_128_GCM_SHA256",
								"TLS_AES_256_GCM_SHA384",
								"TLS_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
							},
						},
					},
				},
			},
		},
		{
			desc:                "TCP TLSOption cross namespace disallowed",
			paths:               []string{"tcp/services.yml", "tcp/with_tls_options_cross_namespace.yml"},
			allowCrossNamespace: false,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"cross-ns-tls-options-cn-843f4bfed2e94b4d9396": {
							MinVersion:    "VersionTLS12",
							ALPNProtocols: []string{"h2", "http/1.1", "acme-tls/1"},
							CipherSuites: []string{
								"TLS_AES_128_GCM_SHA256",
								"TLS_AES_256_GCM_SHA384",
								"TLS_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
								"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
								"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
								"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
								"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256",
								"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256",
							},
						},
					},
				},
			},
		},
		{
			desc:                "UDP cross namespace allowed",
			paths:               []string{"udp/services.yml", "udp/with_cross_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "UDP cross namespace disallowed",
			paths: []string{"udp/services.yml", "udp/with_cross_namespace.yml"},
			expected: &dynamic.Configuration{
				// The router that references the invalid service will be discarded.
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-bfddf9467dc588639f54",
						},
					},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var k8sObjects []runtime.Object
			var crdObjects []runtime.Object
			for _, path := range test.paths {
				yamlContent, err := os.ReadFile(filepath.FromSlash("./fixtures/" + path))
				if err != nil {
					panic(err)
				}

				objects := k8s.MustParseYaml(yamlContent)
				for _, obj := range objects {
					switch o := obj.(type) {
					case *corev1.Service, *corev1.Endpoints, *corev1.Secret:
						k8sObjects = append(k8sObjects, o)
					case *traefikv1alpha1.IngressRoute:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.IngressRouteTCP:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.IngressRouteUDP:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.Middleware:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.MiddlewareTCP:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.TraefikService:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.TLSOption:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.TLSStore:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.ServersTransport:
						crdObjects = append(crdObjects, o)
					default:
					}
				}
			}

			kubeClient := kubefake.NewClientset(k8sObjects...)
			crdClient := traefikcrdfake.NewClientset(crdObjects...)

			client := newClientImpl(kubeClient, crdClient)

			stopCh := make(chan struct{})

			eventCh, err := client.WatchAll([]string{"default", "cross-ns"}, stopCh)
			require.NoError(t, err)

			if k8sObjects != nil || crdObjects != nil {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{AllowCrossNamespace: test.allowCrossNamespace}

			conf := p.loadConfigurationFromCRD(t.Context(), client)
			assert.Equal(t, test.expected, conf)
		})
	}
}

func Test_isCrossProviderNamespaceAllowed(t *testing.T) {
	testCases := []struct {
		desc      string
		allowList []string
		namespace string
		want      bool
	}{
		{desc: "nil allowList allows any namespace", allowList: nil, namespace: "ns-a", want: true},
		{desc: "empty allowList denies every namespace", allowList: []string{}, namespace: "ns-a", want: false},
		{desc: "namespace in allowList is accepted", allowList: []string{"ns-a"}, namespace: "ns-a", want: true},
		{desc: "namespace not in allowList is rejected", allowList: []string{"ns-b"}, namespace: "ns-a", want: false},
		{desc: "namespace among multiple allowed entries is accepted", allowList: []string{"ns-a", "ns-b"}, namespace: "ns-b", want: true},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			got := isCrossProviderNamespaceAllowed(test.allowList, test.namespace)
			assert.Equal(t, test.want, got)
		})
	}
}

// TestCrossProviderNamespaces_HTTPMiddleware verifies that the
// CrossProviderNamespaces option gates middleware references.
// Plain in-namespace middleware references are not affected.
func TestCrossProviderNamespaces_HTTPMiddleware(t *testing.T) {
	testCases := []struct {
		desc                    string
		crossProviderNamespaces []string
		wantMiddlewares         []string
		wantRouterDropped       bool
	}{
		{
			desc:                    "nil: cross-provider middleware refs are accepted (backward compatible)",
			crossProviderNamespaces: nil,
			wantMiddlewares:         []string{"default-stripprefix-a1d92c4f7c7ec6eb4634", "foo-addprefix-727ab48a4ad9f4b63353", "basicauth@file", "redirect@file"},
		},
		{
			desc:                    "empty list: cross-provider middleware refs are rejected, IngressRoute is dropped",
			crossProviderNamespaces: []string{},
			wantRouterDropped:       true,
		},
		{
			desc:                    "namespace allowed: cross-provider middleware refs are accepted",
			crossProviderNamespaces: []string{"default"},
			wantMiddlewares:         []string{"default-stripprefix-a1d92c4f7c7ec6eb4634", "foo-addprefix-727ab48a4ad9f4b63353", "basicauth@file", "redirect@file"},
		},
		{
			desc:                    "namespace not allowed: cross-provider middleware refs are rejected, IngressRoute is dropped",
			crossProviderNamespaces: []string{"other"},
			wantRouterDropped:       true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{
				AllowCrossNamespace:     true,
				CrossProviderNamespaces: test.crossProviderNamespaces,
			}

			clientMock := newClientMock("services.yml", "with_middleware_cross_provider.yml")
			conf := p.loadConfigurationFromCRD(t.Context(), clientMock)

			router, ok := conf.HTTP.Routers["default-test2-route-0-f5a7fc3c74a43f56b526"]
			if test.wantRouterDropped {
				assert.False(t, ok)
				return
			}

			assert.True(t, ok)
			assert.Equal(t, test.wantMiddlewares, router.Middlewares)
		})
	}
}

// TestCrossProviderNamespaces_HTTPServiceTransitivity verifies that the option for a TraefikService chain
// (here: IngressRoute -> Mirror / Weighted TraefikService -> @file service).
func TestCrossProviderNamespaces_HTTPServiceTransitivity(t *testing.T) {
	testCases := []struct {
		desc                    string
		crossProviderNamespaces []string
		wantMirrorService       bool
		wantWeightedService     bool
	}{
		{
			desc:                    "nil: cross-provider service refs accepted (backward compatible)",
			crossProviderNamespaces: nil,
			wantMirrorService:       true,
			wantWeightedService:     true,
		},
		{
			desc:                    "empty list: both Mirror and Weighted TraefikServices are rejected",
			crossProviderNamespaces: []string{},
			wantMirrorService:       false,
			wantWeightedService:     false,
		},
		{
			desc:                    "only the Mirror's namespace is allowed: Weighted is still rejected",
			crossProviderNamespaces: []string{"foo"},
			wantMirrorService:       true,
			wantWeightedService:     false,
		},
		{
			desc:                    "only the Weighted's namespace is allowed: Mirror is still rejected",
			crossProviderNamespaces: []string{"bar"},
			wantMirrorService:       false,
			wantWeightedService:     true,
		},
		{
			desc:                    "both namespaces allowed: both TraefikServices are accepted",
			crossProviderNamespaces: []string{"foo", "bar"},
			wantMirrorService:       true,
			wantWeightedService:     true,
		},
		{
			desc:                    "originating IngressRoute namespace alone is not enough: TraefikService namespace must also be allowed",
			crossProviderNamespaces: []string{"default"},
			wantMirrorService:       false,
			wantWeightedService:     false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{
				AllowCrossNamespace:     true,
				CrossProviderNamespaces: test.crossProviderNamespaces,
			}

			clientMock := newClientMock("services.yml", "with_service_cross_provider.yml")
			conf := p.loadConfigurationFromCRD(t.Context(), clientMock)

			_, mirrorOK := conf.HTTP.Services["foo-mirror-cp-b9e97b251f2755ad29b2"]
			_, weightedOK := conf.HTTP.Services["bar-weighted-cp-e6f8ce202c6b7378f628"]

			assert.Equal(t, test.wantMirrorService, mirrorOK)
			assert.Equal(t, test.wantWeightedService, weightedOK)
		})
	}
}

// TestCrossProviderNamespaces_HTTPTLSOption verifies that the
// CrossProviderNamespaces option gates @file references in IngressRoute tls.options.
func TestCrossProviderNamespaces_HTTPTLSOption(t *testing.T) {
	testCases := []struct {
		desc                    string
		crossProviderNamespaces []string
		wantRouterDropped       bool
	}{
		{
			desc:                    "nil: cross-provider TLSOption ref is accepted (backward compatible)",
			crossProviderNamespaces: nil,
		},
		{
			desc:                    "empty list: cross-provider TLSOption ref is rejected, IngressRoute is dropped",
			crossProviderNamespaces: []string{},
			wantRouterDropped:       true,
		},
		{
			desc:                    "namespace allowed: cross-provider TLSOption ref is accepted",
			crossProviderNamespaces: []string{"default"},
		},
		{
			desc:                    "namespace not allowed: cross-provider TLSOption ref is rejected, IngressRoute is dropped",
			crossProviderNamespaces: []string{"other"},
			wantRouterDropped:       true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{
				AllowCrossNamespace:     true,
				CrossProviderNamespaces: test.crossProviderNamespaces,
			}

			clientMock := newClientMock("services.yml", "with_tls_option_cross_provider.yml")
			conf := p.loadConfigurationFromCRD(t.Context(), clientMock)

			router, ok := conf.HTTP.Routers["default-test-route-0-bfddf9467dc588639f54"]
			if test.wantRouterDropped {
				assert.False(t, ok)
				return
			}

			require.True(t, ok)
			require.NotNil(t, router.TLS)
			assert.Equal(t, "foo@file", router.TLS.Options)
		})
	}
}

// TestCrossProviderNamespaces_TCPTLSOption verifies that the
// CrossProviderNamespaces option gates @file references in IngressRouteTCP tls.options.
func TestCrossProviderNamespaces_TCPTLSOption(t *testing.T) {
	testCases := []struct {
		desc                    string
		crossProviderNamespaces []string
		wantRouterDropped       bool
	}{
		{
			desc:                    "nil: cross-provider TLSOption ref is accepted (backward compatible)",
			crossProviderNamespaces: nil,
		},
		{
			desc:                    "empty list: cross-provider TLSOption ref is rejected, IngressRouteTCP is dropped",
			crossProviderNamespaces: []string{},
			wantRouterDropped:       true,
		},
		{
			desc:                    "namespace allowed: cross-provider TLSOption ref is accepted",
			crossProviderNamespaces: []string{"default"},
		},
		{
			desc:                    "namespace not allowed: cross-provider TLSOption ref is rejected, IngressRouteTCP is dropped",
			crossProviderNamespaces: []string{"other"},
			wantRouterDropped:       true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{
				AllowCrossNamespace:     true,
				CrossProviderNamespaces: test.crossProviderNamespaces,
			}

			clientMock := newClientMock("tcp/services.yml", "tcp/with_tls_options_cross_provider.yml")
			conf := p.loadConfigurationFromCRD(t.Context(), clientMock)

			router, ok := conf.TCP.Routers["default-test-route-0-bfddf9467dc588639f54"]
			if test.wantRouterDropped {
				assert.False(t, ok)
				return
			}

			require.True(t, ok)
			require.NotNil(t, router.TLS)
			assert.Equal(t, "foo@file", router.TLS.Options)
		})
	}
}

// TestCrossProviderNamespaces_HTTPServersTransport verifies that the
// CrossProviderNamespaces option gates @file references in service.serversTransport.
func TestCrossProviderNamespaces_HTTPServersTransport(t *testing.T) {
	testCases := []struct {
		desc                    string
		crossProviderNamespaces []string
		wantServiceDropped      bool
	}{
		{
			desc:                    "nil: cross-provider ServersTransport ref is accepted (backward compatible)",
			crossProviderNamespaces: nil,
		},
		{
			desc:                    "empty list: cross-provider ServersTransport ref is rejected, service is dropped",
			crossProviderNamespaces: []string{},
			wantServiceDropped:      true,
		},
		{
			desc:                    "namespace allowed: cross-provider ServersTransport ref is accepted",
			crossProviderNamespaces: []string{"default"},
		},
		{
			desc:                    "namespace not allowed: cross-provider ServersTransport ref is rejected, service is dropped",
			crossProviderNamespaces: []string{"other"},
			wantServiceDropped:      true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{
				AllowCrossNamespace:     true,
				CrossProviderNamespaces: test.crossProviderNamespaces,
			}

			clientMock := newClientMock("services.yml", "with_servers_transport_cross_provider.yml")
			conf := p.loadConfigurationFromCRD(t.Context(), clientMock)

			service, ok := conf.HTTP.Services["default-test-route-0-lb-611f3105c4cbbc8cca31"]
			if test.wantServiceDropped {
				assert.False(t, ok)
				return
			}

			require.True(t, ok)
			require.NotNil(t, service.LoadBalancer)
			assert.Equal(t, "foo@file", service.LoadBalancer.ServersTransport)
		})
	}
}

func TestExternalNameService(t *testing.T) {
	testCases := []struct {
		desc                     string
		allowExternalNameService bool
		ingressClass             string
		paths                    []string
		expected                 *dynamic.Configuration
	}{
		{
			desc: "Empty",
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                     "HTTP ExternalName services allowed",
			paths:                    []string{"services.yml", "with_externalname_with_http.yml"},
			allowExternalNameService: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`)",
							Priority:    0,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://external.domain:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "HTTP Externalname services disallowed",
			paths: []string{"services.yml", "with_externalname_with_http.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                     "TCP ExternalName services allowed",
			paths:                    []string{"tcp/services.yml", "tcp/with_externalname_with_port.yml"},
			allowExternalNameService: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "external.domain:80",
										Port:    "",
									},
								},
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "TCP ExternalName services disallowed",
			paths: []string{"tcp/services.yml", "tcp/with_externalname_with_port.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					// The router that references the invalid service will be discarded.
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-bfddf9467dc588639f54",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                     "UDP ExternalName services allowed",
			paths:                    []string{"udp/services.yml", "udp/with_externalname_service.yml"},
			allowExternalNameService: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "external.domain:80",
										Port:    "",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "UDP ExternalName service disallowed",
			paths: []string{"udp/services.yml", "udp/with_externalname_service.yml"},
			expected: &dynamic.Configuration{
				// The router that references the invalid service will be discarded.
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-bfddf9467dc588639f54",
						},
					},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var k8sObjects []runtime.Object
			var crdObjects []runtime.Object
			for _, path := range test.paths {
				yamlContent, err := os.ReadFile(filepath.FromSlash("./fixtures/" + path))
				if err != nil {
					panic(err)
				}

				objects := k8s.MustParseYaml(yamlContent)
				for _, obj := range objects {
					switch o := obj.(type) {
					case *corev1.Service, *corev1.Endpoints, *corev1.Secret:
						k8sObjects = append(k8sObjects, o)
					case *traefikv1alpha1.IngressRoute:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.IngressRouteTCP:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.IngressRouteUDP:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.Middleware:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.TraefikService:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.TLSOption:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.TLSStore:
						crdObjects = append(crdObjects, o)
					default:
					}
				}
			}

			kubeClient := kubefake.NewClientset(k8sObjects...)
			crdClient := traefikcrdfake.NewClientset(crdObjects...)

			client := newClientImpl(kubeClient, crdClient)

			stopCh := make(chan struct{})

			eventCh, err := client.WatchAll([]string{"default", "cross-ns"}, stopCh)
			require.NoError(t, err)

			if k8sObjects != nil || crdObjects != nil {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{AllowExternalNameServices: test.allowExternalNameService}

			conf := p.loadConfigurationFromCRD(t.Context(), client)
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestNativeLB(t *testing.T) {
	testCases := []struct {
		desc         string
		ingressClass string
		paths        []string
		expected     *dynamic.Configuration
	}{
		{
			desc: "Empty",
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "HTTP with native Service LB",
			paths: []string{"services.yml", "with_native_service_lb.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`)",
							Priority:    0,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
								},
								PassHostHeader: new(true),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "TCP with native Service LB",
			paths: []string{"tcp/services.yml", "tcp/with_native_service_lb.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
								},
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "UDP with native Service LB",
			paths: []string{"udp/services.yml", "udp/with_native_service_lb.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "10.10.0.1:8000",
										Port:    "",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var k8sObjects []runtime.Object
			var crdObjects []runtime.Object
			for _, path := range test.paths {
				yamlContent, err := os.ReadFile(filepath.FromSlash("./fixtures/" + path))
				if err != nil {
					panic(err)
				}

				objects := k8s.MustParseYaml(yamlContent)
				for _, obj := range objects {
					switch o := obj.(type) {
					case *corev1.Service, *corev1.Endpoints, *corev1.Secret:
						k8sObjects = append(k8sObjects, o)
					case *traefikv1alpha1.IngressRoute:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.IngressRouteTCP:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.IngressRouteUDP:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.Middleware:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.TraefikService:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.TLSOption:
						crdObjects = append(crdObjects, o)
					case *traefikv1alpha1.TLSStore:
						crdObjects = append(crdObjects, o)
					default:
					}
				}
			}

			kubeClient := kubefake.NewClientset(k8sObjects...)
			crdClient := traefikcrdfake.NewClientset(crdObjects...)

			client := newClientImpl(kubeClient, crdClient)

			stopCh := make(chan struct{})

			eventCh, err := client.WatchAll([]string{"default", "cross-ns"}, stopCh)
			require.NoError(t, err)

			if k8sObjects != nil || crdObjects != nil {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{}

			conf := p.loadConfigurationFromCRD(t.Context(), client)
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestCreateBasicAuthCredentials(t *testing.T) {
	var k8sObjects []runtime.Object
	yamlContent, err := os.ReadFile(filepath.FromSlash("./fixtures/basic_auth_secrets.yml"))
	if err != nil {
		panic(err)
	}

	objects := k8s.MustParseYaml(yamlContent)
	for _, obj := range objects {
		switch o := obj.(type) {
		case *corev1.Secret:
			k8sObjects = append(k8sObjects, o)
		default:
		}
	}

	kubeClient := kubefake.NewClientset(k8sObjects...)
	crdClient := traefikcrdfake.NewClientset()

	client := newClientImpl(kubeClient, crdClient)

	stopCh := make(chan struct{})

	eventCh, err := client.WatchAll([]string{"default"}, stopCh)
	require.NoError(t, err)

	if k8sObjects != nil {
		// just wait for the first event
		<-eventCh
	}

	// Testing for username/password components in basic-auth secret
	basicAuth, secretErr := createBasicAuthMiddleware(client, "default", &traefikv1alpha1.BasicAuth{Secret: "basic-auth-secret"})
	require.NoError(t, secretErr)
	require.Len(t, basicAuth.Users, 1)

	components := strings.Split(basicAuth.Users[0], ":")
	require.Len(t, components, 2)

	username := components[0]
	hashedPassword := components[1]

	require.Equal(t, "user", username)
	require.Equal(t, "{SHA}W6ph5Mm5Pz8GgiULbPgzG37mj9g=", hashedPassword)
	assert.True(t, auth.CheckSecret("password", hashedPassword))

	// Testing for username/password components in htpasswd secret
	basicAuth, secretErr = createBasicAuthMiddleware(client, "default", &traefikv1alpha1.BasicAuth{Secret: "auth-secret"})
	require.NoError(t, secretErr)
	require.Len(t, basicAuth.Users, 2)

	components = strings.Split(basicAuth.Users[1], ":")
	require.Len(t, components, 2)

	username = components[0]
	hashedPassword = components[1]

	assert.Equal(t, "test2", username)
	assert.Equal(t, "$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0", hashedPassword)
	assert.True(t, auth.CheckSecret("test2", hashedPassword))
}
