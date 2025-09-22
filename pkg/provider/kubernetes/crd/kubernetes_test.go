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

func pointer[T any](v T) *T { return &v }

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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Middlewares: []string{"default-ipallowlist", "foo-ipallowlist"},
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						"foo-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Middlewares: []string{"default-multiple-hyphens"},
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-multiple-hyphens": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
			paths: []string{"tcp/services.yml", "tcp/with_middleware_crossprovider.yml"},
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Middlewares: []string{"default-ipallowlist", "foo-ipallowlist", "ipallowlist@file", "ipallowlist-foo@file"},
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						"foo-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
						"default-test.route-f44ce589164e656d231c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-f44ce589164e656d231c",
							Rule:        "HostSNI(`bar.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-f44ce589164e656d231c": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp-8000",
										Weight: pointer(2),
									},
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp2-8080",
										Weight: pointer(3),
									},
								},
							},
						},
						"default-test.route-fdd3e9338e47a45efefc-whoamitcp-8000": {
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
						"default-test.route-fdd3e9338e47a45efefc-whoamitcp2-8080": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp-8000",
										Weight: pointer(2),
									},
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp2-8080",
										Weight: pointer(3),
									},
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp3-8083",
										Weight: pointer(4),
									},
								},
							},
						},
						"default-test.route-fdd3e9338e47a45efefc-whoamitcp-8000": {
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
						"default-test.route-fdd3e9338e47a45efefc-whoamitcp2-8080": {
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
						"default-test.route-fdd3e9338e47a45efefc-whoamitcp3-8083": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-foo": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "default-foo",
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"myns-foo": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "myns-foo",
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-foo": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "default-foo",
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-foo": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "default-unknown",
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-foo": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "unknown-foo",
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								TerminationDelay: pointer(500),
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
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
						"default-test.route-673acf455cb2dab0b43a": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-673acf455cb2dab0b43a",
							Rule:        "HostSNI(`*`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-673acf455cb2dab0b43a": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-test.route-673acf455cb2dab0b43a-whoamitcp-ipv6-8080",
										Weight: pointer(1),
									},
									{
										Name:   "default-test.route-673acf455cb2dab0b43a-external.service.with.ipv6-8080",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-test.route-673acf455cb2dab0b43a-whoamitcp-ipv6-8080": {
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
						"default-test.route-673acf455cb2dab0b43a-external.service.with.ipv6-8080": {
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-whoami-ipv6-8080": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://[2001:db8:85a3:8d3:1319:8a2e:370:7348]:8080",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-external-svc-with-ipv6-8080": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://[2001:db8:85a3:8d3:1319:8a2e:370:7347]:8080",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-6b204d94623b3df4370c": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-ipv6-8080",
										Weight: pointer(1),
									},
									{
										Name:   "default-external-svc-with-ipv6-8080",
										Weight: pointer(1),
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"foo-test": {
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
						"default-test": {
							TLS: &dynamic.TLSClientConfig{
								ServerName: "test",
							},
							DialTimeout:      ptypes.Duration(30 * time.Second),
							DialKeepAlive:    ptypes.Duration(15 * time.Second),
							TerminationDelay: ptypes.Duration(100 * time.Millisecond),
						},
					},
					Routers: map[string]*dynamic.TCPRouter{
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc-whoamitcp-8000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.1:8000",
									},
									{
										Address: "10.10.0.2:8000",
									},
								},
								ServersTransport: "default-test",
							},
						},
						"default-test.route-fdd3e9338e47a45efefc-whoamitcp2-8080": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.3:8080",
									},
									{
										Address: "10.10.0.4:8080",
									},
								},
								ServersTransport: "default-default-test",
							},
						},
						"default-test.route-fdd3e9338e47a45efefc": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp-8000",
										Weight: pointer(1),
									},
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp2-8080",
										Weight: pointer(1),
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
			crdClient := traefikcrdfake.NewSimpleClientset(crdObjects...)

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
		desc                string
		ingressClass        string
		paths               []string
		expected            *dynamic.Configuration
		allowCrossNamespace bool
		allowEmptyServices  bool
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							Observability: &dynamic.RouterObservabilityConfig{
								AccessLogs: pointer(true),
								Tracing:    pointer(true),
								Metrics:    pointer(true),
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-test2-route-23c7f4c450289ee29016": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-23c7f4c450289ee29016",
							Rule:        "Host(`foo.com`) && PathPrefix(`/tobestripped`)",
							Priority:    12,
							Middlewares: []string{"default-stripprefix", "default-ratelimit", "foo-addprefix"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ratelimit": {
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
						"default-stripprefix": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/tobestripped"},
							},
						},
						"foo-addprefix": {
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/tobeadded",
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test2-route-23c7f4c450289ee29016": {
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
								PassHostHeader: pointer(true),
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
						"default-test2-route-3c9bf014491ebdba74f7": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-3c9bf014491ebdba74f7",
							Rule:        "Host(`foo.com`) && PathPrefix(`/will-be-limited`)",
							Priority:    12,
							Middlewares: []string{"default-ratelimit"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ratelimit": {
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
									ReadTimeout:    pointer(ptypes.Duration(42 * time.Second)),
									WriteTimeout:   pointer(ptypes.Duration(42 * time.Second)),
									DialTimeout:    pointer(ptypes.Duration(42 * time.Second)),
								},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test2-route-3c9bf014491ebdba74f7": {
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
								PassHostHeader: pointer(true),
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
						"default-test2-route-23c7f4c450289ee29016": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-23c7f4c450289ee29016",
							Rule:        "Host(`foo.com`) && PathPrefix(`/tobestripped`)",
							Priority:    12,
							Middlewares: []string{"default-multiple-hyphens"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-multiple-hyphens": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/tobestripped"},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test2-route-23c7f4c450289ee29016": {
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
								PassHostHeader: pointer(true),
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
			desc:                "Simple Ingress Route with middleware crossprovider",
			allowCrossNamespace: true,
			paths:               []string{"services.yml", "with_middleware_crossprovider.yml"},
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
						"default-test2-route-23c7f4c450289ee29016": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-23c7f4c450289ee29016",
							Rule:        "Host(`foo.com`) && PathPrefix(`/tobestripped`)",
							Priority:    12,
							Middlewares: []string{"default-stripprefix", "foo-addprefix", "basicauth@file", "redirect@file"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-stripprefix": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/tobestripped"},
							},
						},
						"foo-addprefix": {
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/tobeadded",
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test2-route-23c7f4c450289ee29016": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Service:     "default-test-route-6b204d94623b3df4370c",
							Priority:    14,
						},
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-77c62dfe9517144aeeaa",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-77c62dfe9517144aeeaa": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-77c62dfe9517144aeeaa",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-77c62dfe9517144aeeaa": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: pointer(1),
									},
									{
										Name:   "default-whoami2-8080",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-whoami-80": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami2-8080": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default-wrr1",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-whoami5-8080": {
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
								PassHostHeader: pointer(true),
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
						"default-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-whoami5-8080": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-77c62dfe9517144aeeaa",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-77c62dfe9517144aeeaa": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-wrr1",
										Weight: pointer(1),
									},
									{
										Name:   "default-wrr2",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami4-80",
										Weight: pointer(1),
									},
									{
										Name:   "default-whoami5-8080",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-whoami4-80": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami5-8080": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami6-80",
										Weight: pointer(1),
									},
									{
										Name:   "default-whoami7-8080",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-whoami6-80": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami7-8080": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default-wrr1",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-wrr2",
										Weight: pointer(1),
									},
									{
										Name:   "default-whoami5-8080",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-whoami5-8080": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default-wrr1",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-wrr2",
										Weight: pointer(1),
									},
									{
										Name:   "default-whoami5-8080",
										Weight: pointer(1),
									},
									{
										Name:   "default-mirror1",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-mirror1": {
							Mirroring: &dynamic.Mirroring{
								Service: "default-whoami5-8080",
								Mirrors: []dynamic.MirrorService{
									{Name: "default-whoami4-8080", Percent: 50},
								},
							},
						},
						"default-whoami4-8080": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami5-8080": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-77c62dfe9517144aeeaa",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-77c62dfe9517144aeeaa": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								HealthCheck: &dynamic.ServerHealthCheck{
									Path:              "/health",
									Timeout:           5000000000,
									Interval:          15000000000,
									UnhealthyInterval: pointer(ptypes.Duration(15000000000)),
									FollowRedirects:   pointer(true),
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
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-77c62dfe9517144aeeaa",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-77c62dfe9517144aeeaa": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-external-svc-443",
										Weight: func(i int) *int { return &i }(1),
									},
									{
										Name:   "default-external-svc-with-https-443",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-external-svc-443": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								HealthCheck: &dynamic.ServerHealthCheck{
									Path:              "/health1",
									Timeout:           5000000000,
									Interval:          15000000000,
									UnhealthyInterval: pointer(ptypes.Duration(15000000000)),
									FollowRedirects:   pointer(true),
								},
							},
						},
						"default-external-svc-with-https-443": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								HealthCheck: &dynamic.ServerHealthCheck{
									Path:              "/health2",
									Timeout:           5000000000,
									Interval:          20000000000,
									UnhealthyInterval: pointer(ptypes.Duration(20000000000)),
									FollowRedirects:   pointer(true),
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
						"default-external-svc-443": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								HealthCheck: &dynamic.ServerHealthCheck{
									Path:              "/health1",
									Timeout:           5000000000,
									Interval:          15000000000,
									UnhealthyInterval: pointer(ptypes.Duration(15000000000)),
									FollowRedirects:   pointer(true),
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
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-77c62dfe9517144aeeaa",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-77c62dfe9517144aeeaa": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "baz-whoami6-8080",
										Weight: pointer(1),
									},
									{
										Name:   "foo-wrr1",
										Weight: pointer(1),
									},
									{
										Name:   "foo-mirror2",
										Weight: pointer(1),
									},
									{
										Name:   "foo-mirror3",
										Weight: pointer(1),
									},
									{
										Name:   "foo-mirror4",
										Weight: pointer(1),
									},
								},
							},
						},
						"baz-whoami6-8080": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "foo-whoami4-8080",
										Weight: pointer(1),
									},
									{
										Name:   "baz-whoami6-8080",
										Weight: pointer(1),
									},
									{
										Name:   "foo-mirror1",
										Weight: pointer(1),
									},
									{
										Name:   "bar-wrr2",
										Weight: pointer(1),
									},
								},
							},
						},
						"foo-whoami4-8080": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"foo-mirror1": {
							Mirroring: &dynamic.Mirroring{
								Service: "foo-whoami5-8080",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-whoami4-8080"},
									{Name: "baz-whoami6-8080"},
									{Name: "bar-mirrored"},
								},
							},
						},
						"foo-whoami5-8080": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"bar-mirrored": {
							Mirroring: &dynamic.Mirroring{
								Service: "baz-whoami6-8080",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-whoami4-8080", Percent: 50},
								},
							},
						},
						"foo-mirror2": {
							Mirroring: &dynamic.Mirroring{
								Service: "foo-whoami5-8080",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-whoami4-8080"},
									{Name: "baz-whoami6-8080"},
									{Name: "bar-mirrored"},
									{Name: "foo-wrr1"},
								},
							},
						},
						"foo-mirror3": {
							Mirroring: &dynamic.Mirroring{
								Service: "foo-wrr1",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-whoami4-8080"},
									{Name: "baz-whoami6-8080"},
									{Name: "bar-mirrored"},
									{Name: "foo-wrr1"},
								},
							},
						},
						"bar-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "foo-whoami5-8080",
										Weight: pointer(1),
									},
								},
							},
						},
						"foo-mirror4": {
							Mirroring: &dynamic.Mirroring{
								Service: "foo-wrr1",
								Mirrors: []dynamic.MirrorService{
									{Name: "foo-whoami4-8080"},
									{Name: "baz-whoami6-8080"},
									{Name: "bar-mirrored"},
									{Name: "foo-wrr1"},
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
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default-mirror1",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-mirror1": {
							Mirroring: &dynamic.Mirroring{
								Service: "default-whoami5-8080",
								Mirrors: []dynamic.MirrorService{
									{Name: "default-whoami4-8080", Percent: 50},
								},
							},
						},
						"default-whoami4-8080": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami5-8080": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default-mirror1",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-mirror1": {
							Mirroring: &dynamic.Mirroring{
								Service: "default-wrr1",
								Mirrors: []dynamic.MirrorService{
									{Name: "default-wrr2", Percent: 30},
								},
							},
						},
						"default-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami4-8080",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-whoami4-8080": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami5-8080": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-77c62dfe9517144aeeaa",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-77c62dfe9517144aeeaa": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: pointer(10),
									},
									{
										Name:   "default-whoami2-8080",
										Weight: pointer(0),
									},
								},
							},
						},
						"default-whoami-80": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami2-8080": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-02719a68b11e915a4b23": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-02719a68b11e915a4b23",
							Rule:        "/prefix",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-02719a68b11e915a4b23": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-foo": {
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-foo",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-foo",
							},
						},
						"default-test-route-default-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-default-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-foo",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-default-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-foo",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"myns-foo": {
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "myns-foo",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-foo": {
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-foo",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-foo": {
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-unknown",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-foo": {
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "unknown-foo",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-basicauth": {
							BasicAuth: &dynamic.BasicAuth{
								Users: dynamic.Users{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							},
						},
						"default-digestauth": {
							DigestAuth: &dynamic.DigestAuth{
								Users: dynamic.Users{"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/", "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0"},
							},
						},
						"default-forwardauth": {
							ForwardAuth: &dynamic.ForwardAuth{
								Address:     "test.com",
								MaxBodySize: pointer(int64(-1)),
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
						"default-test-secret": {
							Plugin: map[string]dynamic.PluginConf{
								"test-secret": map[string]interface{}{
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
						"default-test-secret": {
							Plugin: map[string]dynamic.PluginConf{
								"test-secret": map[string]interface{}{
									"secret_0": map[string]interface{}{
										"secret_1": map[string]interface{}{
											"secret_2": map[string]interface{}{
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
						"default-test-secret": {
							Plugin: map[string]dynamic.PluginConf{
								"test-secret": map[string]interface{}{
									"secret": []interface{}{"secret_data1", "secret_data2"},
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
						"default-test-secret": {
							Plugin: map[string]dynamic.PluginConf{
								"test-secret": map[string]interface{}{
									"users": []interface{}{
										map[string]interface{}{
											"name":   "admin",
											"secret": "admin_password",
										},
										map[string]interface{}{
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
						"default-errorpage": {
							Errors: &dynamic.ErrorPage{
								Status: []string{"404", "500"},
								StatusRewrites: map[string]int{
									"404": 200,
								},
								Service: "default-errorpage-errorpage-service",
								Query:   "query",
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-errorpage-errorpage-service": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader:     pointer(false),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-test-route-default-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-default-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-route-default-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6f97418635c7e18853da": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6f97418635c7e18853da": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://external.domain:80",
									},
								},
								PassHostHeader: pointer(true),
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
						"default-test-route-6f97418635c7e18853da": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6f97418635c7e18853da": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://external.domain:80",
									},
								},
								PassHostHeader: pointer(true),
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
						"default-test-route-6f97418635c7e18853da": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6f97418635c7e18853da": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader: pointer(true),
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
						"foo-test": {
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
						"default-test": {
							ServerName: "test",
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(30 * time.Second),
								IdleConnTimeout: ptypes.Duration(90 * time.Second),
								PingTimeout:     ptypes.Duration(15 * time.Second),
							},
						},
					},
					Routers: map[string]*dynamic.Router{
						"default-test-route-6f97418635c7e18853da": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-external-svc-with-https-443": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://external.domain:443",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "default-test",
							},
						},
						"default-whoamitls-443": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "default-default-test",
							},
						},
						"default-test-route-6f97418635c7e18853da": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-external-svc-with-https-443",
										Weight: pointer(1),
									},
									{
										Name:   "default-whoamitls-443",
										Weight: pointer(1),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Middlewares: []string{"default-test-errorpage"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-test-errorpage": {
							Errors: &dynamic.ErrorPage{
								Service: "default-test-errorpage-errorpage-service",
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-test-weighted",
										Weight: pointer(1),
									},
									{
										Name:   "default-test-mirror",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-test-errorpage-errorpage-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-weighted": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-without-endpointslice-endpoints-80",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-test-mirror": {
							Mirroring: &dynamic.Mirroring{
								Service: "default-whoami-without-endpointslice-endpoints-80",
								Mirrors: []dynamic.MirrorService{
									{
										Name: "default-whoami-without-endpointslice-endpoints-80",
									},
									{
										Name: "default-test-weighted",
									},
								},
							},
						},
						"default-whoami-without-endpointslice-endpoints-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
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
						"default-test2-route-840425136fbd5d85a4ad": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-840425136fbd5d85a4ad",
							Rule:        "Host(`k8s-service`)",
						},
						"default-test2-route-4f06607bbc69f34a4db5": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-4f06607bbc69f34a4db5",
							Rule:        "Host(`traefik-service`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test2-route-840425136fbd5d85a4ad": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: pointer(1),
									},
									{
										Name:   "default-whoami2-8080",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-test2-route-4f06607bbc69f34a4db5": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-sticky",
										Weight: pointer(1),
									},
									{
										Name:   "default-sticky-default",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-sticky": {
							Weighted: &dynamic.WeightedRoundRobin{
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "cookie",
										Secure:   true,
										HTTPOnly: true,
										SameSite: "none",
										MaxAge:   42,
										Path:     pointer("/foo"),
									},
								},
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami3-8443",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-sticky-default": {
							Weighted: &dynamic.WeightedRoundRobin{
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "cookie",
										Secure:   true,
										HTTPOnly: true,
										SameSite: "none",
										MaxAge:   42,
										Path:     pointer("/"),
									},
								},
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami3-8443",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-whoami2-8080": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "cookie",
										Secure:   true,
										HTTPOnly: true,
										SameSite: "none",
										MaxAge:   42,
										Path:     pointer("/"),
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "cookie",
										Secure:   true,
										HTTPOnly: true,
										SameSite: "none",
										MaxAge:   42,
										Path:     pointer("/foo"),
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami3-8443": {
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
								PassHostHeader: pointer(true),
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.expected == nil {
				return
			}

			k8sObjects, crdObjects := readResources(t, test.paths)

			kubeClient := kubefake.NewClientset(k8sObjects...)
			crdClient := traefikcrdfake.NewSimpleClientset(crdObjects...)

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
				"default-test-route-6b204d94623b3df4370c": {
					EntryPoints: []string{"foo"},
					Service:     "default-test-route-6b204d94623b3df4370c",
					Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
					Priority:    12,
				},
			},
			Middlewares: map[string]*dynamic.Middleware{},
			Services: map[string]*dynamic.Service{
				"default-test-route-6b204d94623b3df4370c": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy:       dynamic.BalancerStrategyWRR,
						PassHostHeader: pointer(true),
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
	crdClient := traefikcrdfake.NewSimpleClientset(crdObjects...)

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

	service, ok := conf.HTTP.Services["default-test-route-6b204d94623b3df4370c"]
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
						},
						"default-test.route-1": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-1",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
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
						"default-test.route-1": {
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
							Weighted: &dynamic.UDPWeightedRoundRobin{
								Services: []dynamic.UDPWRRService{
									{
										Name:   "default-test.route-0-whoamiudp-8000",
										Weight: pointer(2),
									},
									{
										Name:   "default-test.route-0-whoamiudp2-8080",
										Weight: pointer(3),
									},
								},
							},
						},
						"default-test.route-0-whoamiudp-8000": {
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
						"default-test.route-0-whoamiudp2-8080": {
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
							Weighted: &dynamic.UDPWeightedRoundRobin{
								Services: []dynamic.UDPWRRService{
									{
										Name:   "default-test.route-0-whoamiudp-8000",
										Weight: pointer(2),
									},
									{
										Name:   "default-test.route-0-whoamiudp2-8080",
										Weight: pointer(3),
									},
									{
										Name:   "default-test.route-0-whoamiudp3-8083",
										Weight: pointer(4),
									},
								},
							},
						},
						"default-test.route-0-whoamiudp-8000": {
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
						"default-test.route-0-whoamiudp2-8080": {
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
						"default-test.route-0-whoamiudp3-8083": {
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
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
			crdClient := traefikcrdfake.NewSimpleClientset(crdObjects...)

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
						"default-test-crossnamespace-route-9313b71dbe6a649d5049": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-9313b71dbe6a649d5049",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bir`)",
							Priority:    12,
							Middlewares: []string{"default-test-errorpage"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"cross-ns-stripprefix": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/stripit"},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test-crossnamespace-route-9313b71dbe6a649d5049": {
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
								PassHostHeader: pointer(true),
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
						"default-test-crossnamespace-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							Middlewares: []string{
								"cross-ns-stripprefix",
							},
						},
						"default-test-crossnamespace-route-9313b71dbe6a649d5049": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-9313b71dbe6a649d5049",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bir`)",
							Priority:    12,
							Middlewares: []string{"default-test-errorpage"},
						},
						"default-test-crossnamespace-route-a1963878aac7331b7950": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-a1963878aac7331b7950",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bur`)",
							Priority:    12,
							Middlewares: []string{"cross-ns-stripprefix@kubernetescrd"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"cross-ns-stripprefix": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/stripit"},
							},
						},
						"default-test-errorpage": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"500-599"},
								Service: "default-test-errorpage-errorpage-service",
								Query:   "/{status}.html",
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test-crossnamespace-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-crossnamespace-route-9313b71dbe6a649d5049": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-errorpage-errorpage-service": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-test-crossnamespace-route-a1963878aac7331b7950": {
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
								PassHostHeader: pointer(true),
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
						"default-cross-ns-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-cross-ns-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
						"default-cross-ns-route-1bc3efa892379bb93c6e": {
							EntryPoints: []string{"foo"},
							Service:     "default-cross-ns-route-1bc3efa892379bb93c6e",
							Rule:        "Host(`bar.com`) && PathPrefix(`/foo`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-cross-ns-route-6b204d94623b3df4370c": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80",
										Weight: pointer(1),
									},
									{
										Name:   "default-tr-svc-wrr1",
										Weight: pointer(1),
									},
									{
										Name:   "cross-ns-tr-svc-wrr2",
										Weight: pointer(1),
									},
									{
										Name:   "default-tr-svc-mirror1",
										Weight: pointer(1),
									},
									{
										Name:   "cross-ns-tr-svc-mirror2",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-cross-ns-route-1bc3efa892379bb93c6e": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "foo-test@kubernetescrd",
							},
						},
						"cross-ns-whoami-svc-80": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-tr-svc-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80",
										Weight: pointer(1),
									},
								},
							},
						},
						"cross-ns-tr-svc-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80",
										Weight: pointer(1),
									},
								},
							},
						},
						"default-tr-svc-mirror1": {
							Mirroring: &dynamic.Mirroring{
								Service: "default-whoami-80",
								Mirrors: []dynamic.MirrorService{
									{
										Name:    "cross-ns-whoami-svc-80",
										Percent: 20,
									},
								},
							},
						},
						"cross-ns-tr-svc-mirror2": {
							Mirroring: &dynamic.Mirroring{
								Service: "cross-ns-whoami-svc-80",
								Mirrors: []dynamic.MirrorService{
									{
										Name:    "cross-ns-whoami-svc-80",
										Percent: 20,
									},
								},
							},
						},
						"default-whoami-80": {
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
								PassHostHeader: pointer(true),
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
						"cross-ns-tr-svc-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80",
										Weight: pointer(1),
									},
								},
							},
						},
						"cross-ns-whoami-svc-80": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"cross-ns-tr-svc-mirror2": {
							Mirroring: &dynamic.Mirroring{
								Service: "cross-ns-whoami-svc-80",
								Mirrors: []dynamic.MirrorService{
									{
										Name:    "cross-ns-whoami-svc-80",
										Percent: 20,
									},
								},
							},
						},
						"default-whoami-80": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "cross-ns-st-cross-ns@kubernetescrd",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"cross-ns-st-cross-ns": {
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
						"cross-ns-st-cross-ns": {
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "cross-ns-tls-options-cn",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"cross-ns-tls-options-cn": {
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
						"default-test-route-6b204d94623b3df4370c": {
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
								PassHostHeader: pointer(true),
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
						"cross-ns-tls-options-cn": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Middlewares: []string{"default-ipallowlist"},
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						"cross-ns-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Middlewares: []string{"default-ipallowlist"},
							Rule:        "HostSNI(`foo.com`)",
						},
						"default-test.route-f44ce589164e656d231c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-f44ce589164e656d231c",
							Middlewares: []string{"cross-ns-ipallowlist"},
							Rule:        "HostSNI(`bar.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						"cross-ns-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-f44ce589164e656d231c": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"cross-ns-st-cross-ns": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							Priority:    12,
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{
						"cross-ns-st-cross-ns": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "cross-ns-tls-options-cn",
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"cross-ns-tls-options-cn": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"cross-ns-tls-options-cn": {
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
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
			crdClient := traefikcrdfake.NewSimpleClientset(crdObjects...)

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
						"default-test-route-6f97418635c7e18853da": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
							Priority:    0,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6f97418635c7e18853da": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://external.domain:80",
									},
								},
								PassHostHeader: pointer(true),
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
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
			crdClient := traefikcrdfake.NewSimpleClientset(crdObjects...)

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
						"default-test-route-6f97418635c7e18853da": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
							Priority:    0,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6f97418635c7e18853da": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
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
			crdClient := traefikcrdfake.NewSimpleClientset(crdObjects...)

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
						"default-test-route-6f97418635c7e18853da": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
							Priority:    0,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6f97418635c7e18853da": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								Servers: []dynamic.Server{
									{
										URL: "http://172.16.4.4:32456",
									},
								},
								PassHostHeader: pointer(true),
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
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
			crdClient := traefikcrdfake.NewSimpleClientset(crdObjects...)

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
	crdClient := traefikcrdfake.NewSimpleClientset()

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
				assert.Equal(t, "default-my-middleware@kubernetescrd", name)
			}

			backendFunc, ok := r.groupKindBackendFuncs[traefikv1alpha1.SchemeGroupVersion.Group]["TraefikService"]
			require.True(t, ok)

			name, svc, err := backendFunc("my-service", "default")
			test.wantErr(t, err)

			if err == nil {
				assert.Nil(t, svc)
				assert.Equal(t, "default-my-service@kubernetescrd", name)
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
						"default-global-native-lb-6f97418635c7e18853da": {
							EntryPoints: []string{"foo"},
							Service:     "default-global-native-lb-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
							Priority:    0,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-global-native-lb-6f97418635c7e18853da": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
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
						"default-test-route-native-disabled-6f97418635c7e18853da": {
							Service:  "default-test-route-native-disabled-6f97418635c7e18853da",
							Rule:     "Host(`foo.com`)",
							Priority: 0,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-native-disabled-6f97418635c7e18853da": {
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
								PassHostHeader: pointer(true),
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
						"default-test-route-6f97418635c7e18853da": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
							Priority:    0,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6f97418635c7e18853da": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
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
						"default-global-native-lb-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-global-native-lb-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-global-native-lb-fdd3e9338e47a45efefc": {
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
						"default-tcp.route.native-disabled-fdd3e9338e47a45efefc": {
							Service: "default-tcp.route.native-disabled-fdd3e9338e47a45efefc",
							Rule:    "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp.route.native-disabled-fdd3e9338e47a45efefc": {
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
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
						"default-global-native-lb-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-global-native-lb-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-global-native-lb-0": {
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
						"default-udp.route.native-disabled-0": {
							Service: "default-udp.route.native-disabled-0",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-udp.route.native-disabled-0": {
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
			crdClient := traefikcrdfake.NewSimpleClientset(crdObjects...)

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
