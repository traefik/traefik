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
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider"
	traefikcrdfake "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/clientset/versioned/fake"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/gateway"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	kubefake "k8s.io/client-go/kubernetes/fake"
	kscheme "k8s.io/client-go/kubernetes/scheme"
)

var _ provider.Provider = (*Provider)(nil)

func init() {
	// required by k8s.MustParseYaml
	err := traefikv1alpha1.AddToScheme(kscheme.Scheme)
	if err != nil {
		panic(err)
	}
}

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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Simple Ingress Route, with foo entrypoint, tls encryption to service",
			paths: []string{"tcp/services.yml", "tcp/with_tls_service.yml"},
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
										TLS:     true,
									},
									{
										Address: "10.10.0.2:8000",
										TLS:     true,
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
							Middlewares: []string{"default-ipallowlist-8054e4c269217e4851c6", "foo-ipallowlist-e7f4cc21f66a9220e0bf"},
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipallowlist-8054e4c269217e4851c6": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						"foo-ipallowlist-e7f4cc21f66a9220e0bf": {
							IPAllowList: &dynamic.TCPIPAllowList{
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
							IPAllowList: &dynamic.TCPIPAllowList{
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
							Middlewares: []string{"default-ipallowlist-8054e4c269217e4851c6", "foo-ipallowlist-e7f4cc21f66a9220e0bf", "ipallowlist@file", "ipallowlist-foo@file"},
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipallowlist-8054e4c269217e4851c6": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						"foo-ipallowlist-e7f4cc21f66a9220e0bf": {
							IPAllowList: &dynamic.TCPIPAllowList{
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-test-route-0-wrr-0-default-whoamitcp-8000-2e12f708facee87e75e6",
										Weight: new(2),
									},
									{
										Name:   "default-test-route-0-wrr-1-default-whoamitcp2-8080-49dbd6d4eaf55546e68e",
										Weight: new(3),
									},
								},
							},
						},
						"default-test-route-0-wrr-0-default-whoamitcp-8000-2e12f708facee87e75e6": {
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
						"default-test-route-0-wrr-1-default-whoamitcp2-8080-49dbd6d4eaf55546e68e": {
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-test-route-0-wrr-0-default-whoamitcp-8000-2e12f708facee87e75e6",
										Weight: new(2),
									},
									{
										Name:   "default-test-route-0-wrr-1-default-whoamitcp2-8080-49dbd6d4eaf55546e68e",
										Weight: new(3),
									},
									{
										Name:   "default-test-route-0-wrr-2-ns3-whoamitcp3-8083-9d12f33f04765888e6b5",
										Weight: new(4),
									},
								},
							},
						},
						"default-test-route-0-wrr-0-default-whoamitcp-8000-2e12f708facee87e75e6": {
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
						"default-test-route-0-wrr-1-default-whoamitcp2-8080-49dbd6d4eaf55546e68e": {
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
						"default-test-route-0-wrr-2-ns3-whoamitcp3-8083-9d12f33f04765888e6b5": {
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								CAFiles: []types.FileOrContent{
									types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								ClientAuthType: "VerifyClientCertIfGiven",
							},
							SniStrict:             true,
							DisableSessionTickets: true,
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								CAFiles: []types.FileOrContent{
									types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								CAFiles: []types.FileOrContent{
									types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								TerminationDelay: new(500),
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					// The router that references the invalid service will be discarded.
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-bfddf9467dc588639f54",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-test-route-0-wrr-0-default-whoamitcp-ipv6-8080-b2b0b74255e3be36bc93",
										Weight: new(1),
									},
									{
										Name:   "default-test-route-0-wrr-1-default-external-service-with-ipv6-8080-43fbcebb03cafd31e512",
										Weight: new(1),
									},
								},
							},
						},
						"default-test-route-0-wrr-0-default-whoamitcp-ipv6-8080-b2b0b74255e3be36bc93": {
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
						"default-test-route-0-wrr-1-default-external-service-with-ipv6-8080-43fbcebb03cafd31e512": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "[fe80::200:5aee:feaa:20a2]:8080",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
						"default-test-route-0-wrr-0-default-whoami-ipv6-8080-eed7adeccb0c37670f58": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://[2001:db8:85a3:8d3:1319:8a2e:370:7348]:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-0-wrr-1-default-external-svc-with-ipv6-8080-fff50966e3b8a3ca929e": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://[2001:db8:85a3:8d3:1319:8a2e:370:7347]:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-test-route-0-wrr-0-default-whoami-ipv6-8080-eed7adeccb0c37670f58",
										Weight: new(1),
									},
									{
										Name:   "default-test-route-0-wrr-1-default-external-svc-with-ipv6-8080-fff50966e3b8a3ca929e",
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc:  "TCP with ServersTransport",
			paths: []string{"tcp/services.yml", "tcp/with_servers_transport.yml"},
			expected: &dynamic.Configuration{
				TLS: &dynamic.TLSConfiguration{},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					ServersTransports: map[string]*dynamic.TCPServersTransport{
						"foo-test-0bbbd89c878c4df00916": {
							TLS: &dynamic.TLSClientConfig{
								ServerName:         "test",
								InsecureSkipVerify: true,
								RootCAs:            []types.FileOrContent{"TESTROOTCAS0", "TESTROOTCAS1", "TESTROOTCAS2", "TESTROOTCAS3", "TESTROOTCAS5", "TESTALLCERTS", "TESTROOTCASFROMCONFIGMAP", "TESTROOTCAS6"},
								Certificates: tls.Certificates{
									{CertFile: "TESTCERT1", KeyFile: "TESTKEY1"},
									{CertFile: "TESTCERT2", KeyFile: "TESTKEY2"},
									{CertFile: "TESTCERT3", KeyFile: "TESTKEY3"},
								},
								PeerCertURI: "foo://bar",
								Spiffe: &dynamic.Spiffe{
									IDs: []string{
										"spiffe://foo/buz",
										"spiffe://bar/biz",
									},
									TrustDomain: "spiffe://lol",
								},
							},
							DialTimeout:      ptypes.Duration(42 * time.Second),
							DialKeepAlive:    ptypes.Duration(42 * time.Second),
							TerminationDelay: ptypes.Duration(42 * time.Second),
						},
						"default-test-e6d80538b02592c965d8": {
							TLS: &dynamic.TLSClientConfig{
								ServerName: "test",
							},
							DialTimeout:      ptypes.Duration(30 * time.Second),
							DialKeepAlive:    ptypes.Duration(15 * time.Second),
							TerminationDelay: ptypes.Duration(100 * time.Millisecond),
						},
					},
					Routers: map[string]*dynamic.TCPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-wrr-a8d5191cb9281af18d2d",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test-route-0-wrr-0-default-whoamitcp-8000-2e12f708facee87e75e6": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
								ServersTransport: "default-test-e6d80538b02592c965d8",
							},
						},
						"default-test-route-0-wrr-1-default-whoamitcp2-8080-49dbd6d4eaf55546e68e": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.3:8080",
									},
									{
										Address: "10.10.0.4:8080",
									},
								},
								ServersTransport: "default-default-test-592f4763608d32419a4e",
							},
						},
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-test-route-0-wrr-0-default-whoamitcp-8000-2e12f708facee87e75e6",
										Weight: new(1),
									},
									{
										Name:   "default-test-route-0-wrr-1-default-whoamitcp2-8080-49dbd6d4eaf55546e68e",
										Weight: new(1),
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
			desc:  "Ingress Route, empty service disallowed",
			paths: []string{"tcp/services.yml", "tcp/with_empty_services.yml"},
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
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc:               "Ingress Route, nil endpointslice port name",
			allowEmptyServices: true,
			paths:              []string{"tcp/services.yml", "tcp/with_nil_endpointslice_port_name.yml"},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc:               "Ingress Route, nil endpointslice port value",
			allowEmptyServices: true,
			paths:              []string{"tcp/services.yml", "tcp/with_nil_endpointslice_port_value.yml"},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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

			k8sObjects, crdObjects := readResources(t, test.paths)

			kubeClient := kubefake.NewClientset(k8sObjects...)
			crdClient := traefikcrdfake.NewClientset(crdObjects...)

			client := newClientImpl(kubeClient, crdClient)

			stopCh := make(chan struct{})

			eventCh, err := client.WatchAll(nil, stopCh)
			require.NoError(t, err)

			if k8sObjects != nil || crdObjects != nil {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				IngressClass:              test.ingressClass,
				AllowCrossNamespace:       true,
				AllowExternalNameServices: true,
				AllowEmptyServices:        test.allowEmptyServices,
			}

			conf := p.loadConfigurationFromCRD(t.Context(), client)
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							Observability: &dynamic.RouterObservabilityConfig{
								AccessLogs: new(true),
								Tracing:    new(true),
								Metrics:    new(true),
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:                "Simple Ingress Route with middleware",
			allowCrossNamespace: true,
			paths:               []string{"services.yml", "with_middleware.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:                "Simple Ingress Route with middleware ratelimit",
			allowCrossNamespace: true,
			paths:               []string{"services.yml", "with_ratelimit.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test2-route-0-f5a7fc3c74a43f56b526": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-0-lb-1b6127ca68ad95acf267",
							Rule:        "Host(`foo.com`) && PathPrefix(`/will-be-limited`)",
							Priority:    12,
							Middlewares: []string{"default-ratelimit-b8e74655f71c8bbfd816"},
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
								Redis: &dynamic.Redis{
									Endpoints: []string{"127.0.0.1:6379"},
									Username:  "user",
									Password:  "password",
									TLS: &types.ClientTLS{
										CA:   "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----",
										Cert: "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----",
										Key:  "-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----",
									},
									DB:             0,
									PoolSize:       42,
									MaxActiveConns: 42,
									ReadTimeout:    new(ptypes.Duration(42 * time.Second)),
									WriteTimeout:   new(ptypes.Duration(42 * time.Second)),
									DialTimeout:    new(ptypes.Duration(42 * time.Second)),
								},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test2-route-0-lb-1b6127ca68ad95acf267": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:                "Middlewares in ingress route config are normalized",
			allowCrossNamespace: true,
			paths:               []string{"services.yml", "with_middleware_multiple_hyphens.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:  "One ingress Route with two different services",
			paths: []string{"services.yml", "with_two_services.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-test-route-0-wrr-0-default-whoami-80-841cdbc191cff57e966e",
										Weight: new(1),
									},
									{
										Name:   "default-test-route-0-wrr-1-default-whoami2-8080-cc3602540bbe7b16a4f6",
										Weight: new(1),
									},
								},
							},
						},
						"default-test-route-0-wrr-0-default-whoami-80-841cdbc191cff57e966e": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-0-wrr-1-default-whoami2-8080-cc3602540bbe7b16a4f6": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-wrr1-wrr-0-default-whoami5-8080-c0ae3d501350255537be",
										Weight: new(1),
									},
								},
							},
						},
						"default-wrr1-wrr-0-default-whoami5-8080-c0ae3d501350255537be": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-wrr1-3b6294b40828fe78faec": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-wrr1-wrr-0-default-whoami5-8080-c0ae3d501350255537be",
										Weight: new(1),
									},
								},
							},
						},
						"default-wrr1-wrr-0-default-whoami5-8080-c0ae3d501350255537be": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-wrr1-wrr-0-default-whoami4-80-32dc0859b9e1b4b7b655",
										Weight: new(1),
									},
									{
										Name:   "default-wrr1-wrr-1-default-whoami5-8080-d47feba41bf8c095e5f4",
										Weight: new(1),
									},
								},
							},
						},
						"default-wrr1-wrr-0-default-whoami4-80-32dc0859b9e1b4b7b655": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-wrr1-wrr-1-default-whoami5-8080-d47feba41bf8c095e5f4": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-wrr2-504732714158b42ee5b2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-wrr2-wrr-0-default-whoami6-80-cdd2715cdae85ce74320",
										Weight: new(1),
									},
									{
										Name:   "default-wrr2-wrr-1-default-whoami7-8080-ae9735706c5b0eff5fe5",
										Weight: new(1),
									},
								},
							},
						},
						"default-wrr2-wrr-0-default-whoami6-80-cdd2715cdae85ce74320": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.5:80",
									},
									{
										URL: "http://10.10.0.6:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-wrr2-wrr-1-default-whoami7-8080-ae9735706c5b0eff5fe5": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:8080",
									},
									{
										URL: "http://10.10.0.8:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-wrr1-wrr-1-default-whoami5-8080-d47feba41bf8c095e5f4",
										Weight: new(1),
									},
								},
							},
						},
						"default-wrr2-504732714158b42ee5b2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-wrr2-wrr-0-default-whoami5-8080-0b3643349720777594db",
										Weight: new(1),
									},
								},
							},
						},
						"default-wrr1-wrr-1-default-whoami5-8080-d47feba41bf8c095e5f4": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-wrr2-wrr-0-default-whoami5-8080-0b3643349720777594db": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-wrr1-wrr-1-default-whoami5-8080-d47feba41bf8c095e5f4",
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
										Name:   "default-wrr2-wrr-0-default-whoami5-8080-0b3643349720777594db",
										Weight: new(1),
									},
								},
							},
						},
						"default-mirror1-27e0059f29381e4e4b93": {
							Mirroring: &dynamic.Mirroring{
								Service: "default-mirror1-mirroring-default-whoami5-8080-39a38a4ab26d269f730e",
								Mirrors: []dynamic.MirrorService{
									{Name: "default-mirror1-mirror-0-default-whoami4-8080-b4a44e5c4684ecfe3b32", Percent: 50},
								},
							},
						},
						"default-mirror1-mirror-0-default-whoami4-8080-b4a44e5c4684ecfe3b32": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-mirror1-mirroring-default-whoami5-8080-39a38a4ab26d269f730e": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-wrr1-wrr-1-default-whoami5-8080-d47feba41bf8c095e5f4": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-wrr2-wrr-0-default-whoami5-8080-0b3643349720777594db": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "with one external service and health check",
			paths: []string{"services.yml", "with_one_external_service_and_health_check.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								HealthCheck: &dynamic.ServerHealthCheck{
									Path:              "/health",
									Timeout:           5000000000,
									Interval:          15000000000,
									UnhealthyInterval: new(ptypes.Duration(15000000000)),
									FollowRedirects:   new(true),
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
			desc:  "with two external services and health check",
			paths: []string{"services.yml", "with_two_external_services_and_health_check.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
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
										Name:   "default-test-route-0-wrr-0-default-external-svc-443-0e5045562d61c940929b",
										Weight: func(i int) *int { return &i }(1),
									},
									{
										Name:   "default-test-route-0-wrr-1-default-external-svc-with-https-443-f6e0eba42dfeb3659998",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-test-route-0-wrr-0-default-external-svc-443-0e5045562d61c940929b": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								HealthCheck: &dynamic.ServerHealthCheck{
									Path:              "/health1",
									Timeout:           5000000000,
									Interval:          15000000000,
									UnhealthyInterval: new(ptypes.Duration(15000000000)),
									FollowRedirects:   new(true),
								},
							},
						},
						"default-test-route-0-wrr-1-default-external-svc-with-https-443-f6e0eba42dfeb3659998": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								HealthCheck: &dynamic.ServerHealthCheck{
									Path:              "/health2",
									Timeout:           5000000000,
									Interval:          20000000000,
									UnhealthyInterval: new(ptypes.Duration(20000000000)),
									FollowRedirects:   new(true),
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
			desc:  "with one external service and one regular service and health check",
			paths: []string{"services.yml", "with_one_external_svc_and_regular_svc_health_check.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-wrr-0-default-external-svc-443-0e5045562d61c940929b": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								HealthCheck: &dynamic.ServerHealthCheck{
									Path:              "/health1",
									Timeout:           5000000000,
									Interval:          15000000000,
									UnhealthyInterval: new(ptypes.Duration(15000000000)),
									FollowRedirects:   new(true),
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-test-route-0-wrr-0-baz-whoami6-8080-5f688feace15393f0cf4",
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
						"bar-mirrored-mirroring-baz-whoami6-8080-2f84f10f112b1e086d0c": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.5:8080",
									},
									{
										URL: "http://10.10.0.6:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-0-wrr-0-baz-whoami6-8080-5f688feace15393f0cf4": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.5:8080",
									},
									{
										URL: "http://10.10.0.6:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-mirror1-mirror-1-baz-whoami6-8080-17ccb90611e6695cdf21": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.5:8080",
									},
									{
										URL: "http://10.10.0.6:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-mirror2-mirror-1-baz-whoami6-8080-37b125df0691ccdcf51d": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.5:8080",
									},
									{
										URL: "http://10.10.0.6:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-mirror3-mirror-1-baz-whoami6-8080-565564a0fe7869555a55": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.5:8080",
									},
									{
										URL: "http://10.10.0.6:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-mirror4-mirror-1-baz-whoami6-8080-6c814d1d71f291a6d79f": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.5:8080",
									},
									{
										URL: "http://10.10.0.6:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-wrr1-wrr-1-baz-whoami6-8080-f5fc0957839b0bd25184": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.5:8080",
									},
									{
										URL: "http://10.10.0.6:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-wrr1-3f46e81fadb0e6d5865d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "foo-wrr1-wrr-0-foo-whoami4-8080-bf09457480a5cceee6c1",
										Weight: new(1),
									},
									{
										Name:   "foo-wrr1-wrr-1-baz-whoami6-8080-f5fc0957839b0bd25184",
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
						"bar-mirrored-mirror-0-foo-whoami4-8080-c63b3e22b331440db79c": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-mirror1-mirror-0-foo-whoami4-8080-8db3166770ed85651a2a": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-mirror2-mirror-0-foo-whoami4-8080-b9801c68744a6ff21506": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-mirror3-mirror-0-foo-whoami4-8080-66a698402622fd1730e7": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-mirror4-mirror-0-foo-whoami4-8080-2f15ab2094a278db09dd": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-wrr1-wrr-0-foo-whoami4-8080-bf09457480a5cceee6c1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-mirror1-cbfee77b7fea5a43b4ef": {
							Mirroring: &dynamic.Mirroring{
								Service: "foo-mirror1-mirroring-foo-whoami5-8080-d3c09cf2a7b706675cee",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-mirror1-mirror-0-foo-whoami4-8080-8db3166770ed85651a2a"},
									{Name: "foo-mirror1-mirror-1-baz-whoami6-8080-17ccb90611e6695cdf21"},
									{Name: "bar-mirrored-653b67a4c37564aca28d"},
								},
							},
						},
						"bar-wrr2-wrr-0-foo-whoami5-8080-c694d181753f5106b49a": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-mirror1-mirroring-foo-whoami5-8080-d3c09cf2a7b706675cee": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-mirror2-mirroring-foo-whoami5-8080-ff83c253b23c208af503": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"bar-mirrored-653b67a4c37564aca28d": {
							Mirroring: &dynamic.Mirroring{
								Service: "bar-mirrored-mirroring-baz-whoami6-8080-2f84f10f112b1e086d0c",
								Mirrors: []dynamic.MirrorService{
									{Name: "bar-mirrored-mirror-0-foo-whoami4-8080-c63b3e22b331440db79c", Percent: 50},
								},
							},
						},
						"foo-mirror2-5ad6655aa920162e663f": {
							Mirroring: &dynamic.Mirroring{
								Service: "foo-mirror2-mirroring-foo-whoami5-8080-ff83c253b23c208af503",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-mirror2-mirror-0-foo-whoami4-8080-b9801c68744a6ff21506"},
									{Name: "foo-mirror2-mirror-1-baz-whoami6-8080-37b125df0691ccdcf51d"},
									{Name: "bar-mirrored-653b67a4c37564aca28d"},
									{Name: "foo-wrr1-3f46e81fadb0e6d5865d"},
								},
							},
						},
						"foo-mirror3-2954533fbd26f06c9da8": {
							Mirroring: &dynamic.Mirroring{
								Service: "foo-wrr1-3f46e81fadb0e6d5865d",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-mirror3-mirror-0-foo-whoami4-8080-66a698402622fd1730e7"},
									{Name: "foo-mirror3-mirror-1-baz-whoami6-8080-565564a0fe7869555a55"},
									{Name: "bar-mirrored-653b67a4c37564aca28d"},
									{Name: "foo-wrr1-3f46e81fadb0e6d5865d"},
								},
							},
						},
						"bar-wrr2-05f3b853d4318465ac66": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "bar-wrr2-wrr-0-foo-whoami5-8080-c694d181753f5106b49a",
										Weight: new(1),
									},
								},
							},
						},
						"foo-mirror4-4a1eee1c5f7cb88b0a7d": {
							Mirroring: &dynamic.Mirroring{
								Service: "foo-wrr1-3f46e81fadb0e6d5865d",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-mirror4-mirror-0-foo-whoami4-8080-2f15ab2094a278db09dd"},
									{Name: "foo-mirror4-mirror-1-baz-whoami6-8080-6c814d1d71f291a6d79f"},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Service: "default-mirror1-mirroring-default-whoami5-8080-39a38a4ab26d269f730e",
								Mirrors: []dynamic.MirrorService{
									{Name: "default-mirror1-mirror-0-default-whoami4-8080-b4a44e5c4684ecfe3b32", Percent: 50},
								},
							},
						},
						"default-mirror1-mirror-0-default-whoami4-8080-b4a44e5c4684ecfe3b32": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-mirror1-mirroring-default-whoami5-8080-39a38a4ab26d269f730e": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-wrr1-wrr-0-default-whoami4-8080-45eaa417be0b2799bd43",
										Weight: new(1),
									},
								},
							},
						},
						"default-wrr2-504732714158b42ee5b2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-wrr2-wrr-0-default-whoami5-8080-0b3643349720777594db",
										Weight: new(1),
									},
								},
							},
						},
						"default-wrr1-wrr-0-default-whoami4-8080-45eaa417be0b2799bd43": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-wrr2-wrr-0-default-whoami5-8080-0b3643349720777594db": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "one kube services in a highest random weight",
			paths: []string{"with_highest_random_weight.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-hrw1-4f594398168ee3534bfa",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-hrw1-4f594398168ee3534bfa": {
							HighestRandomWeight: &dynamic.HighestRandomWeight{
								Services: []dynamic.HRWService{
									{Name: "default-hrw1-hrw-0-default-whoami1-8080-36800e91e9cc50817b89", Weight: new(10)},
									{Name: "default-hrw1-hrw-1-default-whoami2-8080-176f7b5406c3ebbfc595", Weight: new(20)},
								},
							},
						},
						"default-hrw1-hrw-0-default-whoami1-8080-36800e91e9cc50817b89": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-hrw1-hrw-1-default-whoami2-8080-176f7b5406c3ebbfc595": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-test-route-0-wrr-0-default-whoami-80-841cdbc191cff57e966e",
										Weight: new(10),
									},
									{
										Name:   "default-test-route-0-wrr-1-default-whoami2-8080-cc3602540bbe7b16a4f6",
										Weight: new(0),
									},
								},
							},
						},
						"default-test-route-0-wrr-0-default-whoami-80-841cdbc191cff57e966e": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-0-wrr-1-default-whoami2-8080-cc3602540bbe7b16a4f6": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc:  "Route with empty kind is allowed",
			paths: []string{"services.yml", "with_empty_rule_kind.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "/prefix",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
					},
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
								CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
								CAFiles: []types.FileOrContent{
									types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								},
								ClientAuthType: "VerifyClientCertIfGiven",
							},
							SniStrict:             true,
							DisableSessionTickets: true,
							ALPNProtocols: []string{
								"h2",
								"http/1.1",
								"acme-tls/1",
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
						"default-test-route-default-0-ec5a13bc121647a6c17c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-default-0-lb-c466dfd42416a88b5425",
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-default-0-lb-c466dfd42416a88b5425": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
								CAFiles: []types.FileOrContent{
									types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
								CAFiles: []types.FileOrContent{
									types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
									types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
								CAFiles: []types.FileOrContent{
									types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.5:8443",
									},
									{
										URL: "https://10.10.0.6:8443",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.7:8443",
									},
									{
										URL: "https://10.10.0.8:8443",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Address:     "test.com",
								MaxBodySize: new(int64(-1)),
								HeaderField: "X-Header-Field",
								TLS: &dynamic.ClientTLS{
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{
						"default-errorpage-28845717dbca24eb7eb9": {
							Errors: &dynamic.ErrorPage{
								Status: []string{"404", "500"},
								StatusRewrites: map[string]int{
									"404": 200,
								},
								Service:             "default-errorpage-28845717dbca24eb7eb9-errorpage-service",
								Query:               "query",
								ErrorRequestHeaders: []string{"X-Foo-Bar", "X-Bar-Foo"},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-errorpage-28845717dbca24eb7eb9-errorpage-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-errorpage-wrr-wrr-0-default-whoami-80-850a48734a03d7fc586d",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-errorpage-wrr-wrr-0-default-whoami-80-850a48734a03d7fc586d": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader:     new(false),
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: ptypes.Duration(10 * time.Second)},
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
								CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
								CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
						"default-test-route-default-0-ec5a13bc121647a6c17c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-default-0-lb-c466dfd42416a88b5425",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-default-0-lb-c466dfd42416a88b5425": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://external.domain:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:  "Ingress Route, externalName service with http",
			paths: []string{"services.yml", "with_externalname_with_http.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://external.domain:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:  "Ingress Route, externalName service with https",
			paths: []string{"services.yml", "with_externalname_with_https.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:  "Ingress Route, externalName service without ports",
			paths: []string{"services.yml", "with_externalname_without_ports.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{
						"foo-test-0bbbd89c878c4df00916": {
							ServerName:         "test",
							InsecureSkipVerify: true,
							RootCAs:            []types.FileOrContent{"TESTROOTCAS0", "TESTROOTCAS1", "TESTROOTCAS2", "TESTROOTCAS3", "TESTROOTCAS5", "TESTALLCERTS", "TESTROOTCASFROMCONFIGMAP", "TESTROOTCAS6"},
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
							Spiffe: &dynamic.Spiffe{
								IDs: []string{
									"spiffe://foo/buz",
									"spiffe://bar/biz",
								},
								TrustDomain: "spiffe://lol",
							},
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
						"default-test-route-0-wrr-0-default-external-svc-with-https-443-ea033d96506ad8eeed18": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "default-test-e6d80538b02592c965d8",
							},
						},
						"default-test-route-0-wrr-1-default-whoamitls-443-883d100359bb28821f1a": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.5:8443",
									},
									{
										URL: "https://10.10.0.6:8443",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "default-default-test-592f4763608d32419a4e",
							},
						},
						"default-test-route-0-wrr-a8d5191cb9281af18d2d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-test-route-0-wrr-0-default-external-svc-with-https-443-ea033d96506ad8eeed18",
										Weight: new(1),
									},
									{
										Name:   "default-test-route-0-wrr-1-default-whoamitls-443-883d100359bb28821f1a",
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:               "Ingress Route, nil endpointslice port name",
			allowEmptyServices: true,
			paths:              []string{"services.yml", "with_nil_endpointslice_port_name.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:               "Ingress Route, nil endpointslice port value",
			allowEmptyServices: true,
			paths:              []string{"services.yml", "with_nil_endpointslice_port_value.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:               "IngressRoute, service with multiple endpoint addresses on endpointslice",
			allowEmptyServices: true,
			paths:              []string{"services.yml", "with_multiple_endpointaddresses.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
									{
										URL:    "http://10.10.0.3:80",
										Fenced: true,
									},
									{
										URL:    "http://10.10.0.4:80",
										Fenced: true,
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:               "IngressRoute, service with duplicated endpointaddresses",
			allowEmptyServices: true,
			paths:              []string{"services.yml", "with_duplicated_endpointaddresses.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:               "TraefikService, empty service allowed",
			allowEmptyServices: true,
			paths:              []string{"services.yml", "with_empty_services_ts.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-weighted-4b1c2d77727f0755b988": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-test-weighted-wrr-0-default-whoami-without-endpointslice-endpoints-80-1ee642fe7f69a60d28ad",
										Weight: new(1),
									},
								},
							},
						},
						"default-test-mirror-52e846c6f16eac7e8401": {
							Mirroring: &dynamic.Mirroring{
								Service: "default-test-mirror-mirroring-default-whoami-without-endpointslice-endpoints-80-11ddb80f93e043d1e2f8",
								Mirrors: []dynamic.MirrorService{
									{
										Name: "default-test-mirror-mirror-0-default-whoami-without-endpointslice-endpoints-80-a3e214354aa7f5ca572b",
									},
									{
										Name: "default-test-weighted-4b1c2d77727f0755b988",
									},
								},
							},
						},
						"default-test-weighted-wrr-0-default-whoami-without-endpointslice-endpoints-80-1ee642fe7f69a60d28ad": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-mirror-mirroring-default-whoami-without-endpointslice-endpoints-80-11ddb80f93e043d1e2f8": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-mirror-mirror-0-default-whoami-without-endpointslice-endpoints-80-a3e214354aa7f5ca572b": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:                "Simple Ingress Route with sticky",
			allowCrossNamespace: true,
			paths:               []string{"services.yml", "with_sticky.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test2-route-1-d33442e90133f17a4b21": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-1-wrr-ac56d989fd0c2ad035c3",
							Rule:        "Host(`k8s-service`)",
						},
						"default-test2-route-0-f5a7fc3c74a43f56b526": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-0-wrr-a5db881f770ecd0b4aaa",
							Rule:        "Host(`traefik-service`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test2-route-1-wrr-ac56d989fd0c2ad035c3": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-test2-route-1-wrr-0-default-whoami-80-30ae93b7cfa6354042b0",
										Weight: new(1),
									},
									{
										Name:   "default-test2-route-1-wrr-1-default-whoami2-8080-1b3a9386a117598185d6",
										Weight: new(1),
									},
								},
							},
						},
						"default-test2-route-0-wrr-a5db881f770ecd0b4aaa": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-sticky-e228713921ed9e895e44",
										Weight: new(1),
									},
									{
										Name:   "default-sticky-default-203685fbd3c56f5b3cc7",
										Weight: new(1),
									},
								},
							},
						},
						"default-sticky-e228713921ed9e895e44": {
							Weighted: &dynamic.WeightedRoundRobin{
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "cookie",
										Secure:   true,
										HTTPOnly: true,
										SameSite: "none",
										MaxAge:   42,
										Path:     new("/foo"),
									},
								},
								Services: []dynamic.WRRService{
									{
										Name:   "default-sticky-wrr-0-default-whoami3-8443-5819c4ea33505f1a587b",
										Weight: new(1),
									},
								},
							},
						},
						"default-sticky-default-203685fbd3c56f5b3cc7": {
							Weighted: &dynamic.WeightedRoundRobin{
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "cookie",
										Secure:   true,
										HTTPOnly: true,
										SameSite: "none",
										MaxAge:   42,
										Path:     new("/"),
									},
								},
								Services: []dynamic.WRRService{
									{
										Name:   "default-sticky-default-wrr-0-default-whoami3-8443-7542d84f8726b39492e9",
										Weight: new(1),
									},
								},
							},
						},
						"default-test2-route-1-wrr-1-default-whoami2-8080-1b3a9386a117598185d6": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "cookie",
										Secure:   true,
										HTTPOnly: true,
										SameSite: "none",
										MaxAge:   42,
										Path:     new("/"),
									},
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test2-route-1-wrr-0-default-whoami-80-30ae93b7cfa6354042b0": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "cookie",
										Secure:   true,
										HTTPOnly: true,
										SameSite: "none",
										MaxAge:   42,
										Path:     new("/foo"),
									},
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-sticky-default-wrr-0-default-whoami3-8443-7542d84f8726b39492e9": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:8443",
									},
									{
										URL: "http://10.10.0.8:8443",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-sticky-wrr-0-default-whoami3-8443-5819c4ea33505f1a587b": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:8443",
									},
									{
										URL: "http://10.10.0.8:8443",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:  "IngressRoute with single parent (single route)",
			paths: []string{"parent_refs_services.yml", "parent_refs_single_parent_single_route.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-parent-single-0-4753a8a36e885f2fe210": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`parent.example.com`)",
						},
						"default-child-single-0-1b1d3113bada1db6053b": {
							Service:    "default-child-single-0-lb-b0bb30b2190e9b07f9af",
							Rule:       "Path(`/api`)",
							ParentRefs: []string{"default-parent-single-0-4753a8a36e885f2fe210"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-child-single-0-lb-b0bb30b2190e9b07f9af": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.2.1:9000",
									},
									{
										URL: "http://10.10.2.2:9000",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:  "IngressRoute with single parent (multiple routes) - all parent routers in ParentRefs",
			paths: []string{"parent_refs_services.yml", "parent_refs_single_parent_multiple_routes.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-parent-multi-0-18540b158e3c5e3841d3": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`api.example.com`) && PathPrefix(`/v1`)",
						},
						"default-parent-multi-1-a73d2d848da56e75d84a": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`api.example.com`) && PathPrefix(`/v2`)",
						},
						"default-child-multi-routes-0-50dd18d0dff916f8b6c7": {
							Service:    "default-child-multi-routes-0-lb-16b798d34b52de0faee5",
							Rule:       "Path(`/users`)",
							ParentRefs: []string{"default-parent-multi-0-18540b158e3c5e3841d3", "default-parent-multi-1-a73d2d848da56e75d84a"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-child-multi-routes-0-lb-16b798d34b52de0faee5": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.5.1:9000",
									},
									{
										URL: "http://10.10.5.2:9000",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:  "IngressRoute with multiple parents",
			paths: []string{"parent_refs_services.yml", "parent_refs_multiple_parents.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-parent-a-0-8a221361b3450f77e26d": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`a.example.com`)",
						},
						"default-parent-b-0-4081ea04cc4c45a81b05": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`b.example.com`)",
						},
						"default-child-multi-parents-0-d7fcfffde5a3647b3e88": {
							Service:    "default-child-multi-parents-0-lb-953363c4a23b0f862045",
							Rule:       "Path(`/shared`)",
							ParentRefs: []string{"default-parent-a-0-8a221361b3450f77e26d", "default-parent-b-0-4081ea04cc4c45a81b05"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-child-multi-parents-0-lb-953363c4a23b0f862045": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.8.1:9000",
									},
									{
										URL: "http://10.10.8.2:9000",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:  "IngressRoute with missing parent - routers skipped",
			paths: []string{"parent_refs_services.yml", "parent_refs_missing_parent.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc:                "IngressRoute with cross-namespace parent allowed",
			allowCrossNamespace: true,
			paths:               []string{"parent_refs_services.yml", "parent_refs_cross_namespace_allowed.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"ns-a-parent-cross-0-5ade07cc379127464652": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`cross.example.com`)",
						},
						"ns-b-child-cross-allowed-0-002a2c170cff672f4ca8": {
							Service:    "ns-b-child-cross-allowed-0-lb-fc7c12bfd4b67f97faba",
							Rule:       "Path(`/cross`)",
							ParentRefs: []string{"ns-a-parent-cross-0-5ade07cc379127464652"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"ns-b-child-cross-allowed-0-lb-fc7c12bfd4b67f97faba": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.11.1:9000",
									},
									{
										URL: "http://10.10.11.2:9000",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:                "IngressRoute with cross-namespace parent denied",
			allowCrossNamespace: false,
			paths:               []string{"parent_refs_services.yml", "parent_refs_cross_namespace_denied.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"ns-a-parent-cross-0-5ade07cc379127464652": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`cross.example.com`)",
						},
					},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "IngressRoute with parent namespace defaulting to child namespace",
			paths: []string{"parent_refs_services.yml", "parent_refs_default_namespace.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-parent-default-0-d954bafeda6af6db07e0": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`default.example.com`)",
						},
						"default-child-same-0-c36b00801bf4fea8f2e7": {
							Service:    "default-child-same-0-lb-63a469cdf710093bb363",
							Rule:       "Path(`/same`)",
							ParentRefs: []string{"default-parent-default-0-d954bafeda6af6db07e0"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-child-same-0-lb-63a469cdf710093bb363": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.14.1:9000",
									},
									{
										URL: "http://10.10.14.2:9000",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:  "Simple Ingress Route with leasttime strategy",
			paths: []string{"services.yml", "with_leasttime_strategy.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-0-lb-611f3105c4cbbc8cca31",
							Rule:        "Host(`foo.com`) && PathPrefix(`/leasttime`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyLeastTime,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:                "routes referencing the same Kubernetes Service with different options, namespaces, or twice",
			paths:               []string{"services.yml", "with_shared_kube_service.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-plain-0-61d05419e42137e08325": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-plain-0-wrr-ef2770a4da49ac554d36",
							Rule:        "Host(`foo.com`)",
						},
						"default-test-route-mtls-0-92af783804e8785c8713": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-mtls-0-wrr-2245a22754736c16aef2",
							Rule:        "Host(`bar.com`)",
						},
						"default-test-route-namespaces-0-ea1d4f64f8c8255ca532": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-namespaces-0-wrr-f3ae47cd9c9e11d10187",
							Rule:        "Host(`baz.com`)",
						},
						"default-test-route-duplicate-0-6898409405858118881e": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-duplicate-0-wrr-322bf05328a17302a3fe",
							Rule:        "Host(`qux.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-plain-0-wrr-ef2770a4da49ac554d36": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-test-route-plain-0-wrr-0-default-whoami-80-80f3f0d96676b5d6aabd",
										Weight: new(1),
									},
									{
										Name:   "default-test-route-plain-0-wrr-1-default-whoami2-8080-21bb23bf82bc2deb07cb",
										Weight: new(1),
									},
								},
							},
						},
						"default-test-route-mtls-0-wrr-2245a22754736c16aef2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-test-route-mtls-0-wrr-0-default-whoami-80-3ba88bba361a35456918",
										Weight: new(1),
									},
									{
										Name:   "default-test-route-mtls-0-wrr-1-default-whoami2-8080-29d4d7ae3b46d21abac0",
										Weight: new(1),
									},
								},
							},
						},
						// The two routes reference the same Kubernetes Service, but only one of them
						// sets a serversTransport: the generated services must not be merged.
						"default-test-route-plain-0-wrr-0-default-whoami-80-80f3f0d96676b5d6aabd": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-mtls-0-wrr-0-default-whoami-80-3ba88bba361a35456918": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "default-mtls-ac0dc9093199ac1d191e",
							},
						},
						"default-test-route-plain-0-wrr-1-default-whoami2-8080-21bb23bf82bc2deb07cb": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-mtls-0-wrr-1-default-whoami2-8080-29d4d7ae3b46d21abac0": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-namespaces-0-wrr-f3ae47cd9c9e11d10187": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-test-route-namespaces-0-wrr-0-default-whoami-80-48cbb7d9deef59894ff1",
										Weight: new(1),
									},
									{
										Name:   "default-test-route-namespaces-0-wrr-1-cross-ns-whoami-80-cad2fc128a1f96830ba6",
										Weight: new(1),
									},
								},
							},
						},
						// The same Kubernetes Service name and port in two namespaces.
						"default-test-route-namespaces-0-wrr-0-default-whoami-80-48cbb7d9deef59894ff1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-namespaces-0-wrr-1-cross-ns-whoami-80-cad2fc128a1f96830ba6": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.10:80",
									},
									{
										URL: "http://10.10.0.11:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-duplicate-0-wrr-322bf05328a17302a3fe": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-test-route-duplicate-0-wrr-0-default-whoami-80-622f5a999cc8069e8289",
										Weight: new(1),
									},
									{
										Name:   "default-test-route-duplicate-0-wrr-1-default-whoami-80-1b4ce17e91b1391856ef",
										Weight: new(1),
									},
								},
							},
						},
						// The same Kubernetes Service referenced twice by the same parent.
						"default-test-route-duplicate-0-wrr-0-default-whoami-80-622f5a999cc8069e8289": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-duplicate-0-wrr-1-default-whoami-80-1b4ce17e91b1391856ef": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(false),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.expected == nil {
				return
			}

			k8sObjects, crdObjects := readResources(t, test.paths)

			kubeClient := kubefake.NewClientset(k8sObjects...)
			crdClient := traefikcrdfake.NewClientset(crdObjects...)

			client := newClientImpl(kubeClient, crdClient)

			stopCh := make(chan struct{})

			eventCh, err := client.WatchAll(nil, stopCh)
			require.NoError(t, err)

			if k8sObjects != nil || crdObjects != nil {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				IngressClass:              test.ingressClass,
				AllowCrossNamespace:       test.allowCrossNamespace,
				AllowExternalNameServices: true,
				AllowEmptyServices:        test.allowEmptyServices,
				CrossProviderNamespaces:   test.crossProviderNamespaces,
			}

			conf := p.loadConfigurationFromCRD(t.Context(), client)
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadIngressRoutes_multipleEndpointAddresses(t *testing.T) {
	wantConf := &dynamic.Configuration{
		UDP: &dynamic.UDPConfiguration{
			Routers:  map[string]*dynamic.UDPRouter{},
			Services: map[string]*dynamic.UDPService{},
		},
		TCP: &dynamic.TCPConfiguration{
			Routers:           map[string]*dynamic.TCPRouter{},
			Middlewares:       map[string]*dynamic.TCPMiddleware{},
			Services:          map[string]*dynamic.TCPService{},
			ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
						Strategy:       dynamic.BalancerStrategyWRR,
						PassHostHeader: new(true),
						ResponseForwarding: &dynamic.ResponseForwarding{
							FlushInterval: ptypes.Duration(100 * time.Millisecond),
						},
					},
				},
			},
			ServersTransports: map[string]*dynamic.ServersTransport{},
		},
		TLS: &dynamic.TLSConfiguration{},
	}
	wantServers := []dynamic.Server{
		{
			URL: "http://10.10.0.3:8080",
		},
		{
			URL: "http://10.10.0.4:8080",
		},
		{
			URL: "http://10.10.0.5:8080",
		},
		{
			URL: "http://10.10.0.6:8080",
		},
	}

	k8sObjects, crdObjects := readResources(t, []string{"services.yml", "with_multiple_endpointslices.yml"})

	kubeClient := kubefake.NewClientset(k8sObjects...)
	crdClient := traefikcrdfake.NewClientset(crdObjects...)

	client := newClientImpl(kubeClient, crdClient)

	stopCh := make(chan struct{})

	eventCh, err := client.WatchAll(nil, stopCh)
	require.NoError(t, err)

	if k8sObjects != nil || crdObjects != nil {
		// just wait for the first event
		<-eventCh
	}

	p := Provider{}
	conf := p.loadConfigurationFromCRD(t.Context(), client)

	service, ok := conf.HTTP.Services["default-test-route-0-lb-611f3105c4cbbc8cca31"]
	require.True(t, ok)
	require.NotNil(t, service)
	require.NotNil(t, service.LoadBalancer)
	assert.ElementsMatch(t, wantServers, service.LoadBalancer.Servers)

	service.LoadBalancer.Servers = nil
	assert.Equal(t, wantConf, conf)
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-test-route-0-wrr-0-default-whoamiudp-8000-9edd68bfc44e240a3568",
										Weight: new(2),
									},
									{
										Name:   "default-test-route-0-wrr-1-default-whoamiudp2-8080-348c1ca838c04e27b11c",
										Weight: new(3),
									},
								},
							},
						},
						"default-test-route-0-wrr-0-default-whoamiudp-8000-9edd68bfc44e240a3568": {
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
						"default-test-route-0-wrr-1-default-whoamiudp2-8080-348c1ca838c04e27b11c": {
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-test-route-0-wrr-0-default-whoamiudp-8000-9edd68bfc44e240a3568",
										Weight: new(2),
									},
									{
										Name:   "default-test-route-0-wrr-1-default-whoamiudp2-8080-348c1ca838c04e27b11c",
										Weight: new(3),
									},
									{
										Name:   "default-test-route-0-wrr-2-ns3-whoamiudp3-8083-723c477c198e302dac8d",
										Weight: new(4),
									},
								},
							},
						},
						"default-test-route-0-wrr-0-default-whoamiudp-8000-9edd68bfc44e240a3568": {
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
						"default-test-route-0-wrr-1-default-whoamiudp2-8080-348c1ca838c04e27b11c": {
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
						"default-test-route-0-wrr-2-ns3-whoamiudp3-8083-723c477c198e302dac8d": {
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					// The router that references the invalid service will be discarded.
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-bfddf9467dc588639f54",
						},
					},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					// The router that references the invalid service will be discarded.
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-bfddf9467dc588639f54",
						},
					},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc:               "Ingress Route, nil endpointslice port name",
			allowEmptyServices: true,
			paths:              []string{"udp/services.yml", "udp/with_nil_endpointslice_port_name.yml"},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc:               "Ingress Route, nil endpointslice port value",
			allowEmptyServices: true,
			paths:              []string{"udp/services.yml", "udp/with_nil_endpointslice_port_value.yml"},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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

			k8sObjects, crdObjects := readResources(t, test.paths)

			kubeClient := kubefake.NewClientset(k8sObjects...)
			crdClient := traefikcrdfake.NewClientset(crdObjects...)

			client := newClientImpl(kubeClient, crdClient)

			stopCh := make(chan struct{})

			eventCh, err := client.WatchAll(nil, stopCh)
			require.NoError(t, err)

			if k8sObjects != nil || crdObjects != nil {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				IngressClass:              test.ingressClass,
				AllowCrossNamespace:       true,
				AllowExternalNameServices: true,
				AllowEmptyServices:        test.allowEmptyServices,
			}

			conf := p.loadConfigurationFromCRD(t.Context(), client)
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Prefixes: []string{"/stripit"},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test-crossnamespace-route-1-lb-0ce7ef3cc5e5221ad93f": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-crossnamespace-route-3-lb-37e219f6ea3be495ffa7": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:                "HTTP middleware cross namespace allowed",
			paths:               []string{"services.yml", "with_middleware_cross_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Prefixes: []string{"/stripit"},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-crossnamespace-route-1-lb-0ce7ef3cc5e5221ad93f": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-errorpage-e1e939ab3bb71fe16185-errorpage-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-crossnamespace-route-2-lb-ebbed47881305fcd1030": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-crossnamespace-route-3-lb-37e219f6ea3be495ffa7": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:                "HTTP cross namespace allowed",
			paths:               []string{"services.yml", "with_cross_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "default-cross-ns-route-0-wrr-0-cross-ns-whoami-svc-80-6af2a4511c485b57a8e2",
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "foo-test@kubernetescrd",
							},
						},
						"cross-ns-tr-svc-wrr2-wrr-0-cross-ns-whoami-svc-80-7dd406c2f895c8366975": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-cross-ns-route-0-wrr-0-cross-ns-whoami-svc-80-6af2a4511c485b57a8e2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-tr-svc-mirror1-mirror-0-cross-ns-whoami-svc-80-19dd5305d098a310be0c": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-tr-svc-wrr1-wrr-0-cross-ns-whoami-svc-80-8cc813cb7c586442dbf1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"cross-ns-tr-svc-mirror2-mirroring-cross-ns-whoami-svc-80-64a4c9b9339d949a7d72": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"cross-ns-tr-svc-mirror2-mirror-0-cross-ns-whoami-svc-80-7efda2788fa67c88ec32": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-tr-svc-wrr1-65f91db84ed92e091532": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-tr-svc-wrr1-wrr-0-cross-ns-whoami-svc-80-8cc813cb7c586442dbf1",
										Weight: new(1),
									},
								},
							},
						},
						"cross-ns-tr-svc-wrr2-35846c0f6a99ca063e5c": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-tr-svc-wrr2-wrr-0-cross-ns-whoami-svc-80-7dd406c2f895c8366975",
										Weight: new(1),
									},
								},
							},
						},
						"default-tr-svc-mirror1-8130c8b7f004b659eeb3": {
							Mirroring: &dynamic.Mirroring{
								Service: "default-tr-svc-mirror1-mirroring-default-whoami-80-1bc91306c38dde55119a",
								Mirrors: []dynamic.MirrorService{
									{
										Name:    "default-tr-svc-mirror1-mirror-0-cross-ns-whoami-svc-80-19dd5305d098a310be0c",
										Percent: 20,
									},
								},
							},
						},
						"cross-ns-tr-svc-mirror2-3a1eebdec12b22b77a0f": {
							Mirroring: &dynamic.Mirroring{
								Service: "cross-ns-tr-svc-mirror2-mirroring-cross-ns-whoami-svc-80-64a4c9b9339d949a7d72",
								Mirrors: []dynamic.MirrorService{
									{
										Name:    "cross-ns-tr-svc-mirror2-mirror-0-cross-ns-whoami-svc-80-7efda2788fa67c88ec32",
										Percent: 20,
									},
								},
							},
						},
						"default-tr-svc-mirror1-mirroring-default-whoami-80-1bc91306c38dde55119a": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:  "HTTP cross namespace disallowed",
			paths: []string{"services.yml", "with_cross_namespace.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"cross-ns-tr-svc-wrr2-35846c0f6a99ca063e5c": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-tr-svc-wrr2-wrr-0-cross-ns-whoami-svc-80-7dd406c2f895c8366975",
										Weight: new(1),
									},
								},
							},
						},
						"cross-ns-tr-svc-wrr2-wrr-0-cross-ns-whoami-svc-80-7dd406c2f895c8366975": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"cross-ns-tr-svc-mirror2-mirroring-cross-ns-whoami-svc-80-64a4c9b9339d949a7d72": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"cross-ns-tr-svc-mirror2-mirror-0-cross-ns-whoami-svc-80-7efda2788fa67c88ec32": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"cross-ns-tr-svc-mirror2-3a1eebdec12b22b77a0f": {
							Mirroring: &dynamic.Mirroring{
								Service: "cross-ns-tr-svc-mirror2-mirroring-cross-ns-whoami-svc-80-64a4c9b9339d949a7d72",
								Mirrors: []dynamic.MirrorService{
									{
										Name:    "cross-ns-tr-svc-mirror2-mirror-0-cross-ns-whoami-svc-80-7efda2788fa67c88ec32",
										Percent: 20,
									},
								},
							},
						},
						"default-tr-svc-mirror1-mirroring-default-whoami-80-1bc91306c38dde55119a": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:                "HTTP ServersTransport cross namespace allowed",
			paths:               []string{"services.yml", "with_servers_transport_cross_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Name:   "cross-ns-tr-svc-wrr-0-cross-ns-whoami-svc-80-8841556deb972a597a82",
										Weight: new(1),
									},
								},
							},
						},
						"cross-ns-tr-svc-wrr-0-cross-ns-whoami-svc-80-8841556deb972a597a82": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:  "HTTP Service cross namespace via @kubernetescrd suffix disallowed",
			paths: []string{"services.yml", "with_service_cross_namespace_kubernetescrd.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"cross-ns-tr-svc-b2116afd1555971d704d": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-tr-svc-wrr-0-cross-ns-whoami-svc-80-8841556deb972a597a82",
										Weight: new(1),
									},
								},
							},
						},
						"cross-ns-tr-svc-wrr-0-cross-ns-whoami-svc-80-8841556deb972a597a82": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
			desc:                "HTTP TLSOption cross namespace allowed",
			paths:               []string{"services.yml", "with_tls_options_cross_namespace.yml"},
			allowCrossNamespace: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-0-lb-611f3105c4cbbc8cca31": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
							Middlewares: []string{"default-ipallowlist-8054e4c269217e4851c6"},
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipallowlist-8054e4c269217e4851c6": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						"cross-ns-ipallowlist-796441bd1f2f6ce6d7ab": {
							IPAllowList: &dynamic.TCPIPAllowList{
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
							Middlewares: []string{"default-ipallowlist-8054e4c269217e4851c6"},
							Rule:        "HostSNI(`foo.com`)",
						},
						"default-test-route-1-73e4808ba7263def92b2": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-1-lb-1109d6313f2dbb16a8e6",
							Middlewares: []string{"cross-ns-ipallowlist-796441bd1f2f6ce6d7ab"},
							Rule:        "HostSNI(`bar.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipallowlist-8054e4c269217e4851c6": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						"cross-ns-ipallowlist-796441bd1f2f6ce6d7ab": {
							IPAllowList: &dynamic.TCPIPAllowList{
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc:                "TCP ServersTransport cross namespace allowed",
			paths:               []string{"tcp/services.yml", "tcp/with_servers_transport_cross_namespace.yml"},
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
							Rule:        "HostSNI(`foo.com`)",
							Priority:    12,
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
								ServersTransport: "cross-ns-st-cross-ns@kubernetescrd",
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{
						"cross-ns-st-cross-ns-77710ecac094a47836bf": {
							DialTimeout:      ptypes.Duration(30 * time.Second),
							DialKeepAlive:    0,
							TerminationDelay: ptypes.Duration(100 * time.Millisecond),
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
			desc:  "TCP ServersTransport cross namespace disallowed",
			paths: []string{"tcp/services.yml", "tcp/with_servers_transport_cross_namespace.yml"},
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
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{
						"cross-ns-st-cross-ns-77710ecac094a47836bf": {
							DialTimeout:      30000000000,
							DialKeepAlive:    0,
							TerminationDelay: ptypes.Duration(100 * time.Millisecond),
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "UDP cross namespace disallowed",
			paths: []string{"udp/services.yml", "udp/with_cross_namespace.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					// The router that references the invalid service will be discarded.
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-bfddf9467dc588639f54",
						},
					},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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

			k8sObjects, crdObjects := readResources(t, test.paths)

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

			k8sObjects, crdObjects := readResources(t, []string{"services.yml", "with_middleware_cross_provider.yml"})

			kubeClient := kubefake.NewClientset(k8sObjects...)
			crdClient := traefikcrdfake.NewClientset(crdObjects...)

			client := newClientImpl(kubeClient, crdClient)

			stopCh := make(chan struct{})

			eventCh, err := client.WatchAll(nil, stopCh)
			require.NoError(t, err)

			if k8sObjects != nil || crdObjects != nil {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				AllowCrossNamespace:     true,
				CrossProviderNamespaces: test.crossProviderNamespaces,
			}

			conf := p.loadConfigurationFromCRD(t.Context(), client)

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

			k8sObjects, crdObjects := readResources(t, []string{"services.yml", "with_service_cross_provider.yml"})

			kubeClient := kubefake.NewClientset(k8sObjects...)
			crdClient := traefikcrdfake.NewClientset(crdObjects...)

			client := newClientImpl(kubeClient, crdClient)

			stopCh := make(chan struct{})

			eventCh, err := client.WatchAll(nil, stopCh)
			require.NoError(t, err)

			if k8sObjects != nil || crdObjects != nil {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				AllowCrossNamespace:     true,
				CrossProviderNamespaces: test.crossProviderNamespaces,
			}

			conf := p.loadConfigurationFromCRD(t.Context(), client)

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

			k8sObjects, crdObjects := readResources(t, []string{"services.yml", "with_tls_option_cross_provider.yml"})

			kubeClient := kubefake.NewClientset(k8sObjects...)
			crdClient := traefikcrdfake.NewClientset(crdObjects...)

			client := newClientImpl(kubeClient, crdClient)

			stopCh := make(chan struct{})

			eventCh, err := client.WatchAll(nil, stopCh)
			require.NoError(t, err)

			if k8sObjects != nil || crdObjects != nil {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				AllowCrossNamespace:     true,
				CrossProviderNamespaces: test.crossProviderNamespaces,
			}

			conf := p.loadConfigurationFromCRD(t.Context(), client)

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

			k8sObjects, crdObjects := readResources(t, []string{"tcp/services.yml", "tcp/with_tls_options_cross_provider.yml"})

			kubeClient := kubefake.NewClientset(k8sObjects...)
			crdClient := traefikcrdfake.NewClientset(crdObjects...)

			client := newClientImpl(kubeClient, crdClient)

			stopCh := make(chan struct{})

			eventCh, err := client.WatchAll(nil, stopCh)
			require.NoError(t, err)

			if k8sObjects != nil || crdObjects != nil {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				AllowCrossNamespace:     true,
				CrossProviderNamespaces: test.crossProviderNamespaces,
			}

			conf := p.loadConfigurationFromCRD(t.Context(), client)

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

			k8sObjects, crdObjects := readResources(t, []string{"services.yml", "with_servers_transport_cross_provider.yml"})

			kubeClient := kubefake.NewClientset(k8sObjects...)
			crdClient := traefikcrdfake.NewClientset(crdObjects...)

			client := newClientImpl(kubeClient, crdClient)

			stopCh := make(chan struct{})

			eventCh, err := client.WatchAll(nil, stopCh)
			require.NoError(t, err)

			if k8sObjects != nil || crdObjects != nil {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				AllowCrossNamespace:     true,
				CrossProviderNamespaces: test.crossProviderNamespaces,
			}

			conf := p.loadConfigurationFromCRD(t.Context(), client)

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

// TestCrossProviderNamespaces_TCPServersTransport verifies that the
// CrossProviderNamespaces option gates @file references in IngressRouteTCP service.serversTransport.
func TestCrossProviderNamespaces_TCPServersTransport(t *testing.T) {
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

			k8sObjects, crdObjects := readResources(t, []string{"tcp/services.yml", "tcp/with_servers_transport_cross_provider.yml"})

			kubeClient := kubefake.NewClientset(k8sObjects...)
			crdClient := traefikcrdfake.NewClientset(crdObjects...)

			client := newClientImpl(kubeClient, crdClient)

			stopCh := make(chan struct{})

			eventCh, err := client.WatchAll(nil, stopCh)
			require.NoError(t, err)

			if k8sObjects != nil || crdObjects != nil {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				AllowCrossNamespace:     true,
				CrossProviderNamespaces: test.crossProviderNamespaces,
			}

			conf := p.loadConfigurationFromCRD(t.Context(), client)

			service, ok := conf.TCP.Services["default-test-route-0-lb-611f3105c4cbbc8cca31"]
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://external.domain:80",
									},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "UDP ExternalName service disallowed",
			paths: []string{"udp/services.yml", "udp/with_externalname_service.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					// The router that references the invalid service will be discarded.
					Routers: map[string]*dynamic.UDPRouter{
						"default-test-route-0-bfddf9467dc588639f54": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-0-bfddf9467dc588639f54",
						},
					},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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

			k8sObjects, crdObjects := readResources(t, test.paths)

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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Address: "10.10.0.1:9000",
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			k8sObjects, crdObjects := readResources(t, test.paths)

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

func TestNodePortLB(t *testing.T) {
	testCases := []struct {
		desc                string
		paths               []string
		disableClusterScope bool
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
			desc:  "HTTP with node port LB",
			paths: []string{"services.yml", "with_node_port_lb.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								Servers: []dynamic.Server{
									{
										URL: "http://172.16.4.4:32456",
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
			paths: []string{"tcp/services.yml", "tcp/with_node_port_service_lb.yml"},
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Address: "172.16.4.4:32456",
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
			paths: []string{"udp/services.yml", "udp/with_node_port_service_lb.yml"},
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
										Address: "172.16.4.4:32456",
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                "HTTP with node port LB, cluster scope resources disabled",
			paths:               []string{"services.yml", "with_node_port_lb.yml"},
			disableClusterScope: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
			desc:                "TCP with native Service LB, cluster scope resources disabled",
			paths:               []string{"tcp/services.yml", "tcp/with_node_port_service_lb.yml"},
			disableClusterScope: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc:                "UDP with native Service LB, cluster scope resources disabled",
			paths:               []string{"udp/services.yml", "udp/with_node_port_service_lb.yml"},
			disableClusterScope: true,
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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

			k8sObjects, crdObjects := readResources(t, test.paths)

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

			p := Provider{
				DisableClusterScopeResources: test.disableClusterScope,
			}

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

	if len(k8sObjects) != 0 {
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

func TestFillExtensionBuilderRegistry(t *testing.T) {
	testCases := []struct {
		desc       string
		namespaces []string
		wantErr    require.ErrorAssertionFunc
	}{
		{
			desc:    "no filter on namespaces",
			wantErr: require.NoError,
		},
		{
			desc:       "filter on default namespace",
			namespaces: []string{"default"},
			wantErr:    require.NoError,
		},
		{
			desc:       "filter on not-default namespace",
			namespaces: []string{"not-default"},
			wantErr:    require.Error,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			r := &extensionBuilderRegistryMock{}

			p := Provider{Namespaces: test.namespaces}
			p.FillExtensionBuilderRegistry(r)

			filterFunc, ok := r.groupKindFilterFuncs[traefikv1alpha1.SchemeGroupVersion.Group]["Middleware"]
			require.True(t, ok)

			name, conf, err := filterFunc("my-middleware", "default")
			test.wantErr(t, err)

			if err == nil {
				assert.Nil(t, conf)
				assert.Equal(t, "default-my-middleware-d81541a12f73cb5dbd86@kubernetescrd", name)
			}

			backendFunc, ok := r.groupKindBackendFuncs[traefikv1alpha1.SchemeGroupVersion.Group]["TraefikService"]
			require.True(t, ok)

			name, svc, err := backendFunc("my-service", "default")
			test.wantErr(t, err)

			if err == nil {
				assert.Nil(t, svc)
				assert.Equal(t, "default-my-service-5734ab168fd0d22edb04@kubernetescrd", name)
			}
		})
	}
}

func readResources(t *testing.T, paths []string) ([]runtime.Object, []runtime.Object) {
	t.Helper()

	var k8sObjects []runtime.Object
	var crdObjects []runtime.Object
	for _, path := range paths {
		yamlContent, err := os.ReadFile(filepath.FromSlash("./fixtures/" + path))
		if err != nil {
			panic(err)
		}

		objects := k8s.MustParseYaml(yamlContent)
		for _, obj := range objects {
			switch obj.GetObjectKind().GroupVersionKind().Group {
			case "traefik.io":
				crdObjects = append(crdObjects, obj)
			default:
				k8sObjects = append(k8sObjects, obj)
			}
		}
	}

	return k8sObjects, crdObjects
}

type extensionBuilderRegistryMock struct {
	groupKindFilterFuncs  map[string]map[string]gateway.BuildFilterFunc
	groupKindBackendFuncs map[string]map[string]gateway.BuildBackendFunc
}

// RegisterFilterFuncs registers an allowed Group, Kind, and builder for the Filter ExtensionRef objects.
func (p *extensionBuilderRegistryMock) RegisterFilterFuncs(group, kind string, builderFunc gateway.BuildFilterFunc) {
	if p.groupKindFilterFuncs == nil {
		p.groupKindFilterFuncs = map[string]map[string]gateway.BuildFilterFunc{}
	}

	if p.groupKindFilterFuncs[group] == nil {
		p.groupKindFilterFuncs[group] = map[string]gateway.BuildFilterFunc{}
	}

	p.groupKindFilterFuncs[group][kind] = builderFunc
}

// RegisterBackendFuncs registers an allowed Group, Kind, and builder for the Backend ExtensionRef objects.
func (p *extensionBuilderRegistryMock) RegisterBackendFuncs(group, kind string, builderFunc gateway.BuildBackendFunc) {
	if p.groupKindBackendFuncs == nil {
		p.groupKindBackendFuncs = map[string]map[string]gateway.BuildBackendFunc{}
	}

	if p.groupKindBackendFuncs[group] == nil {
		p.groupKindBackendFuncs[group] = map[string]gateway.BuildBackendFunc{}
	}

	p.groupKindBackendFuncs[group][kind] = builderFunc
}

func TestGlobalNativeLB(t *testing.T) {
	testCases := []struct {
		desc              string
		ingressClass      string
		paths             []string
		NativeLBByDefault bool
		expected          *dynamic.Configuration
	}{
		{
			desc: "Empty",
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
			desc:              "HTTP with global native Service LB",
			paths:             []string{"services.yml", "with_global_native_service_lb.yml"},
			NativeLBByDefault: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers: map[string]*dynamic.Router{
						"default-global-native-lb-0-c7017dbf869e86fd74df": {
							EntryPoints: []string{"foo"},
							Service:     "default-global-native-lb-0-lb-f96ce723008c72404a3b",
							Rule:        "Host(`foo.com`)",
							Priority:    0,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-global-native-lb-0-lb-f96ce723008c72404a3b": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
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
			desc:              "HTTP with global native Service LB but service reference has nativeLB disabled",
			paths:             []string{"services.yml", "with_native_service_lb_disabled.yml"},
			NativeLBByDefault: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					ServersTransports: map[string]*dynamic.ServersTransport{},
					Routers: map[string]*dynamic.Router{
						"default-test-route-native-disabled-0-d6af44a15afcb6558d38": {
							Service:  "default-test-route-native-disabled-0-lb-7bbd9079b6681868969b",
							Rule:     "Host(`foo.com`)",
							Priority: 0,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-native-disabled-0-lb-7bbd9079b6681868969b": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.20:80",
									},
									{
										URL: "http://10.10.0.21:80",
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
			desc:  "HTTP with native Service LB in ingressroute",
			paths: []string{"services.yml", "with_native_service_lb.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
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
			desc:              "TCP with global native Service LB",
			paths:             []string{"tcp/services.yml", "tcp/with_global_native_service_lb.yml"},
			NativeLBByDefault: true,
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers: map[string]*dynamic.TCPRouter{
						"default-global-native-lb-0-c7017dbf869e86fd74df": {
							EntryPoints: []string{"foo"},
							Service:     "default-global-native-lb-0-lb-f96ce723008c72404a3b",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-global-native-lb-0-lb-f96ce723008c72404a3b": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:9000",
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
			desc:              "TCP with global native Service LB but service reference has nativeLB disabled",
			paths:             []string{"tcp/services.yml", "tcp/with_native_service_lb_disabled.yml"},
			NativeLBByDefault: true,
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-route-native-disabled-0-2b41ebbf5fe24604f9ac": {
							Service: "default-tcp-route-native-disabled-0-lb-5b12f849c5795a51b7e4",
							Rule:    "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-route-native-disabled-0-lb-5b12f849c5795a51b7e4": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.30:9000",
										Port:    "",
									},
									{
										Address: "10.10.0.31:9000",
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
			desc:  "TCP with native Service LB in ingressroute",
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
										Address: "10.10.0.1:9000",
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
			desc:  "UDP with native Service LB in ingressroute",
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:              "UDP with global native Service LB",
			paths:             []string{"udp/services.yml", "udp/with_global_native_service_lb.yml"},
			NativeLBByDefault: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-global-native-lb-0-c7017dbf869e86fd74df": {
							EntryPoints: []string{"foo"},
							Service:     "default-global-native-lb-0-lb-f96ce723008c72404a3b",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-global-native-lb-0-lb-f96ce723008c72404a3b": {
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:              "UDP with global native Service LB but service reference has nativeLB disabled",
			paths:             []string{"udp/services.yml", "udp/with_native_service_lb_disabled.yml"},
			NativeLBByDefault: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-udp-route-native-disabled-0-74ffa6894d902ac9a2d7": {
							Service: "default-udp-route-native-disabled-0-lb-a91f9683f8d7d811914f",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-udp-route-native-disabled-0-lb-a91f9683f8d7d811914f": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "10.10.0.30:8000",
										Port:    "",
									},
									{
										Address: "10.10.0.31:8000",
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
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
						k8sObjects = append(k8sObjects, o)
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

			p := Provider{NativeLBByDefault: test.NativeLBByDefault}

			conf := p.loadConfigurationFromCRD(t.Context(), client)
			assert.Equal(t, test.expected, conf)
		})
	}
}
