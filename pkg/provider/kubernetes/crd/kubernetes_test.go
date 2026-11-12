package crd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	auth "github.com/abbot/go-http-auth"
	"github.com/rs/zerolog"
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
		safeNaming         bool
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
			desc:       "Simple Ingress Route, with foo entrypoint, safe naming",
			paths:      []string{"tcp/services.yml", "tcp/simple.yml"},
			safeNaming: true,
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
						"default_test.route_0": {
							EntryPoints: []string{"foo"},
							Service:     "default_test.route_0_lb",
							Rule:        "HostSNI(`foo.com`)",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default_test.route_0_lb": {
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
										TLS:     true,
									},
									{
										Address: "10.10.0.2:8000",
										TLS:     true,
									},
								},
							},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "MiddlewareTCP", Namespace: "default", Name: "ipallowlist", Path: "/spec"},
									},
								},
							},
						},
						"foo-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "MiddlewareTCP", Namespace: "foo", Name: "ipallowlist", Path: "/spec"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
			desc:       "Simple Ingress Route, with foo entrypoint and middleware, safe naming",
			paths:      []string{"tcp/services.yml", "tcp/with_middleware.yml"},
			safeNaming: true,
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
						"default_test.route_0": {
							EntryPoints: []string{"foo"},
							Service:     "default_test.route_0_lb",
							Middlewares: []string{"default_ipallowlist", "foo_ipallowlist"},
							Rule:        "HostSNI(`foo.com`)",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default_ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "MiddlewareTCP", Namespace: "default", Name: "ipallowlist", Path: "/spec"},
									},
								},
							},
						},
						"foo_ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "MiddlewareTCP", Namespace: "foo", Name: "ipallowlist", Path: "/spec"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default_test.route_0_lb": {
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-multiple-hyphens": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "MiddlewareTCP", Namespace: "default", Name: "multiple---hyphens", Path: "/spec"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Middlewares: []string{"default-ipallowlist", "foo-ipallowlist", "ipallowlist@file", "ipallowlist-foo@file"},
							Rule:        "HostSNI(`foo.com`)",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "MiddlewareTCP", Namespace: "default", Name: "ipallowlist", Path: "/spec"},
									},
								},
							},
						},
						"foo-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "MiddlewareTCP", Namespace: "foo", Name: "ipallowlist", Path: "/spec"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-test.route-f44ce589164e656d231c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-f44ce589164e656d231c",
							Rule:        "HostSNI(`bar.com`)",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/1/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/1/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp-8000",
										Weight: new(2),
									},
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp2-8080",
										Weight: new(3),
									},
								},
							},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "8080",
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
			desc:       "One ingress Route with two different services, safe naming",
			paths:      []string{"tcp/services.yml", "tcp/with_two_services.yml"},
			safeNaming: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default_test.route_0": {
							EntryPoints: []string{"foo"},
							Service:     "default_test.route_0_wrr",
							Rule:        "HostSNI(`foo.com`)",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default_test.route_0_wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default_test.route_0_wrr_0_default_whoamitcp_8000",
										Weight: new(2),
									},
									{
										Name:   "default_test.route_0_wrr_1_default_whoamitcp2_8080",
										Weight: new(3),
									},
								},
							},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
									},
								},
							},
						},
						"default_test.route_0_wrr_0_default_whoamitcp_8000": {
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
								},
							},
						},
						"default_test.route_0_wrr_1_default_whoamitcp2_8080": {
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "8080",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp-8000",
										Weight: new(2),
									},
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp2-8080",
										Weight: new(3),
									},
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp3-8083",
										Weight: new(4),
									},
								},
							},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "8080",
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "ns3", Name: "whoamitcp3", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/2"},
									},
									ServicePort: "8083",
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
							CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
							KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
			desc:       "TLS with tls options, safe naming",
			paths:      []string{"tcp/services.yml", "tcp/with_tls_options.yml"},
			safeNaming: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"default_foo": {
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
						"default_test.route_0": {
							EntryPoints: []string{"foo"},
							Service:     "default_test.route_0_lb",
							Rule:        "HostSNI(`foo.com`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "default_foo",
							},

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default_test.route_0_lb": {
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
			desc:       "TLS with tls Store, safe naming",
			paths:      []string{"tcp/services.yml", "tcp/with_tls_store.yml"},
			safeNaming: true,
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
						"default_test.route_0": {
							EntryPoints: []string{"foo"},
							Service:     "default_test.route_0_lb",
							Rule:        "HostSNI(`foo.com`)",
							TLS:         &dynamic.RouterTCPTLSConfig{},

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default_test.route_0_lb": {
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external-svc-tcp", Path: "/spec"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external.service.with.port.tcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
			desc:         "Simple TCP Ingress Route, with ingressClassName",
			paths:        []string{"tcp/services.yml", "tcp/with_ingressclassname.yml"},
			ingressClass: "traefik-lb",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
			desc:         "Simple TCP Ingress Route, with ingressClassName and deprecated annotation",
			paths:        []string{"tcp/services.yml", "tcp/with_ingressclassname_and_deprecated_annotation.yml"},
			ingressClass: "traefik-lb",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
			desc:         "Simple TCP Ingress Route, with deprecated annotation only",
			paths:        []string{"tcp/services.yml", "tcp/with_deprecated_annotation_only.yml"},
			ingressClass: "traefik-lb",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp-ipv6", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8080",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-673acf455cb2dab0b43a": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-test.route-673acf455cb2dab0b43a-whoamitcp-ipv6-8080",
										Weight: new(1),
									},
									{
										Name:   "default-test.route-673acf455cb2dab0b43a-external.service.with.ipv6-8080",
										Weight: new(1),
									},
								},
							},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp-ipv6", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8080",
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external.service.with.ipv6", Path: "/spec"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "8080",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami-ipv6", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8080",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external-svc-with-ipv6", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "8080",
								},
							},
						},
						"default-test-route-6b204d94623b3df4370c": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-ipv6-8080",
										Weight: new(1),
									},
									{
										Name:   "default-external-svc-with-ipv6-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								ProxyProtocol: &dynamic.ProxyProtocol{Version: 2},
							},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "ServersTransportTCP", Namespace: "foo", Name: "test", Path: "/spec"},
									},
								},
							},
						},
						"default-test": {
							TLS: &dynamic.TLSClientConfig{
								ServerName: "test",
							},
							DialTimeout:      ptypes.Duration(30 * time.Second),
							DialKeepAlive:    ptypes.Duration(15 * time.Second),
							TerminationDelay: ptypes.Duration(100 * time.Millisecond),

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "ServersTransportTCP", Namespace: "default", Name: "test", Path: "/spec"},
									},
								},
							},
						},
					},
					Routers: map[string]*dynamic.TCPRouter{
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
								ServersTransport: "default-default-test",
							},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "8080",
								},
							},
						},
						"default-test.route-fdd3e9338e47a45efefc": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp-8000",
										Weight: new(1),
									},
									{
										Name:   "default-test.route-fdd3e9338e47a45efefc-whoamitcp2-8080",
										Weight: new(1),
									},
								},
							},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp-without-endpointslice-endpoints", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp-nil-port-name", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
						"default-test.route-fdd3e9338e47a45efefc": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-fdd3e9338e47a45efefc",
							Rule:        "HostSNI(`foo.com`)",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-test.route-fdd3e9338e47a45efefc": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp-nil-port-value", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
				SafeNaming:                &test.safeNaming,
			}
			require.NoError(t, p.Init())

			conf := p.loadConfigurationFromCRD(t.Context(), client)
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadIngressRoutes(t *testing.T) {
	testCases := []struct {
		desc                         string
		ingressClass                 string
		paths                        []string
		expected                     *dynamic.Configuration
		allowCrossNamespace          bool
		allowEmptyServices           bool
		crossProviderNamespaces      []string
		defaultTLSResourcesNamespace string
		safeNaming                   bool
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
								AccessLogs: new(true),
								Tracing:    new(true),
								Metrics:    new(true),
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
			desc:       "Simple Ingress Route, with foo entrypoint, safe naming",
			paths:      []string{"services.yml", "simple.yml"},
			safeNaming: true,
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
						"default_test.route_0": {
							EntryPoints: []string{"foo"},
							Service:     "default_test.route_0_lb",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							Observability: &dynamic.RouterObservabilityConfig{
								AccessLogs: new(true),
								Tracing:    new(true),
								Metrics:    new(true),
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default_test.route_0_lb": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
			desc:  "Simple Ingress Route, with responding timeouts",
			paths: []string{"services.yml", "with_responding_timeouts.yml"},
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
							RespondingTimeouts: &dynamic.RouterRespondingTimeouts{
								ReadTimeout:  new(ptypes.Duration(5 * time.Second)),
								WriteTimeout: new(ptypes.Duration(10 * time.Second)),
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
			desc:  "Simple Ingress Route, with invalid responding timeouts",
			paths: []string{"services.yml", "with_invalid_responding_timeouts.yml"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/tls"},
									},
								},
							},
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "ratelimit", Path: "/spec"},
									},
								},
							},
						},
						"default-stripprefix": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/tobestripped"},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "stripprefix", Path: "/spec"},
									},
								},
							},
						},
						"foo-addprefix": {
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/tobeadded",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "foo", Name: "addprefix", Path: "/spec"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
			desc:                "Simple Ingress Route with middleware, safe naming",
			allowCrossNamespace: true,
			paths:               []string{"services.yml", "with_middleware.yml"},
			safeNaming:          true,
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
						"default_test2.route_0": {
							EntryPoints: []string{"web"},
							Service:     "default_test2.route_0_lb",
							Rule:        "Host(`foo.com`) && PathPrefix(`/tobestripped`)",
							Priority:    12,
							Middlewares: []string{"default_stripprefix", "default_ratelimit", "foo_addprefix"},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default_ratelimit": {
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "ratelimit", Path: "/spec"},
									},
								},
							},
						},
						"default_stripprefix": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/tobestripped"},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "stripprefix", Path: "/spec"},
									},
								},
							},
						},
						"foo_addprefix": {
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/tobeadded",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "foo", Name: "addprefix", Path: "/spec"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default_test2.route_0_lb": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/tls"},
									},
								},
							},
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
									Username:  "urn:k8s:secret:redissecret:username",
									Password:  "urn:k8s:secret:redissecret:password",
									TLS: &types.ClientTLS{
										CA:   "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----",
										Cert: "urn:k8s:secret:tlssecret:tls.crt",
										Key:  "urn:k8s:secret:tlssecret:tls.key",
									},
									DB:             0,
									PoolSize:       42,
									MaxActiveConns: 42,
									ReadTimeout:    new(ptypes.Duration(42 * time.Second)),
									WriteTimeout:   new(ptypes.Duration(42 * time.Second)),
									DialTimeout:    new(ptypes.Duration(42 * time.Second)),
								},
							},
							URNRefs: map[string]string{
								"/rateLimit/redis/username": "user",
								"/rateLimit/redis/password": "password",
								"/rateLimit/redis/tls/cert": "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----",
								"/rateLimit/redis/tls/key":  "-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "ratelimit", Path: "/spec"},
									},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-multiple-hyphens": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/tobestripped"},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "multiple---hyphens", Path: "/spec"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
						"default-test2-route-23c7f4c450289ee29016": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-23c7f4c450289ee29016",
							Rule:        "Host(`foo.com`) && PathPrefix(`/tobestripped`)",
							Priority:    12,
							Middlewares: []string{"default-stripprefix", "foo-addprefix", "basicauth@file", "redirect@file"},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-stripprefix": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/tobestripped"},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "stripprefix", Path: "/spec"},
									},
								},
							},
						},
						"foo-addprefix": {
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/tobeadded",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "foo", Name: "addprefix", Path: "/spec"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-77c62dfe9517144aeeaa",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/1/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-77c62dfe9517144aeeaa": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: new(1),
									},
									{
										Name:   "default-whoami2-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "8080",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami5", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "8080",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:       "one kube service (== servers lb) in a services wrr, safe naming",
			paths:      []string{"with_services_lb0.yml"},
			safeNaming: true,
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
						"default_test.route_0": {
							EntryPoints: []string{"web"},
							Service:     "default_wrr1",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default_wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default_wrr1_wrr_0_default_whoami5_8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec"},
									},
								},
							},
						},
						"default_wrr1_wrr_0_default_whoami5_8080": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami5", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "8080",
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
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami5", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "8080",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-77c62dfe9517144aeeaa": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-wrr1",
										Weight: new(1),
									},
									{
										Name:   "default-wrr2",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
									},
								},
							},
						},
						"default-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami4-80",
										Weight: new(1),
									},
									{
										Name:   "default-whoami5-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami4", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami5", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec/weighted/services/1"},
									},
									ServicePort: "8080",
								},
							},
						},
						"default-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami6-80",
										Weight: new(1),
									},
									{
										Name:   "default-whoami7-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr2", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami6", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr2", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami7", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr2", Path: "/spec/weighted/services/1"},
									},
									ServicePort: "8080",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-wrr2",
										Weight: new(1),
									},
									{
										Name:   "default-whoami5-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec"},
									},
								},
							},
						},
						"default-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr2", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami5", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec/weighted/services/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr2", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "8080",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-wrr2",
										Weight: new(1),
									},
									{
										Name:   "default-whoami5-8080",
										Weight: new(1),
									},
									{
										Name:   "default-mirror1",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec"},
									},
								},
							},
						},
						"default-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr2", Path: "/spec"},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "mirror1", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami4", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "mirror1", Path: "/spec/mirroring/mirrors/0"},
									},
									ServicePort: "8080",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami5", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "mirror1", Path: "/spec/mirroring"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec/weighted/services/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr2", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "8080",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external-svc", Path: "/spec"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "443",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external-svc", Path: "/spec"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "443",
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external-svc-with-https", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "443",
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external-svc", Path: "/spec"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "443",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-77c62dfe9517144aeeaa": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "baz-whoami6-8080",
										Weight: new(1),
									},
									{
										Name:   "foo-wrr1",
										Weight: new(1),
									},
									{
										Name:   "foo-mirror2",
										Weight: new(1),
									},
									{
										Name:   "foo-mirror3",
										Weight: new(1),
									},
									{
										Name:   "foo-mirror4",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "baz", Name: "whoami6", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "bar", Name: "mirrored", Path: "/spec/mirroring"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror1", Path: "/spec/mirroring/mirrors/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror2", Path: "/spec/mirroring/mirrors/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror3", Path: "/spec/mirroring/mirrors/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror4", Path: "/spec/mirroring/mirrors/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "wrr1", Path: "/spec/weighted/services/1"},
									},
									ServicePort: "8080",
								},
							},
						},
						"foo-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "foo-whoami4-8080",
										Weight: new(1),
									},
									{
										Name:   "baz-whoami6-8080",
										Weight: new(1),
									},
									{
										Name:   "foo-mirror1",
										Weight: new(1),
									},
									{
										Name:   "bar-wrr2",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "wrr1", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "foo", Name: "whoami4", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "bar", Name: "mirrored", Path: "/spec/mirroring/mirrors/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror1", Path: "/spec/mirroring/mirrors/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror2", Path: "/spec/mirroring/mirrors/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror3", Path: "/spec/mirroring/mirrors/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror4", Path: "/spec/mirroring/mirrors/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "wrr1", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "8080",
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror1", Path: "/spec"},
									},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "foo", Name: "whoami5", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "bar", Name: "wrr2", Path: "/spec/weighted/services/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror1", Path: "/spec/mirroring"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror2", Path: "/spec/mirroring"},
									},
									ServicePort: "8080",
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "bar", Name: "mirrored", Path: "/spec"},
									},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror2", Path: "/spec"},
									},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror3", Path: "/spec"},
									},
								},
							},
						},
						"bar-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "foo-whoami5-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "bar", Name: "wrr2", Path: "/spec"},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "foo", Name: "mirror4", Path: "/spec"},
									},
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "mirror1", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami4", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "mirror1", Path: "/spec/mirroring/mirrors/0"},
									},
									ServicePort: "8080",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami5", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "mirror1", Path: "/spec/mirroring"},
									},
									ServicePort: "8080",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:       "one kube service (== servers lb) in a mirroring, safe naming",
			paths:      []string{"with_mirroring.yml"},
			safeNaming: true,
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
						"default_test.route_0": {
							EntryPoints: []string{"web"},
							Service:     "default_mirror1",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default_mirror1": {
							Mirroring: &dynamic.Mirroring{
								Service: "default_mirror1_mirroring_default_whoami5_8080",
								Mirrors: []dynamic.MirrorService{
									{Name: "default_mirror1_mirror_0_default_whoami4_8080", Percent: 50},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "mirror1", Path: "/spec"},
									},
								},
							},
						},
						"default_mirror1_mirror_0_default_whoami4_8080": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami4", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "mirror1", Path: "/spec/mirroring/mirrors/0"},
									},
									ServicePort: "8080",
								},
							},
						},
						"default_mirror1_mirroring_default_whoami5_8080": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami5", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "mirror1", Path: "/spec/mirroring"},
									},
									ServicePort: "8080",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "mirror1", Path: "/spec"},
									},
								},
							},
						},
						"default-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami4-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec"},
									},
								},
							},
						},
						"default-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr2", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami4", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "8080",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami5", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr2", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "8080",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TraefikService with service middleware",
			paths: []string{"services.yml", "with_traefik_service_middleware.yml"},
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
							Service:     "default-test-weighted",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-stripprefix": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/tobestripped"},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "stripprefix", Path: "/spec"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test-weighted": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "test-weighted", Path: "/spec"},
									},
								},
							},
						},
						"default-whoami-80": {
							Middlewares: []string{"default-stripprefix"},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "test-weighted", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "80",
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
						"default-test-route-77c62dfe9517144aeeaa": {
							EntryPoints: []string{"web"},
							Service:     "default-hrw1",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-hrw1": {
							HighestRandomWeight: &dynamic.HighestRandomWeight{
								Services: []dynamic.HRWService{
									{Name: "default-whoami1-8080", Weight: new(10)},
									{Name: "default-whoami2-8080", Weight: new(20)},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "hrw1", Path: "/spec"},
									},
								},
							},
						},
						"default-whoami1-8080": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami1", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "hrw1", Path: "/spec/highestRandomWeight/services/0"},
									},
									ServicePort: "8080",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "hrw1", Path: "/spec/highestRandomWeight/services/1"},
									},
									ServicePort: "8080",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "one kube service in a failover",
			paths: []string{"with_failover.yml"},
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
							Service:     "default-failover1",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-failover1": {
							Failover: &dynamic.Failover{
								Service:  "default-whoami5-8080",
								Fallback: "default-whoami4-8080",
								Errors:   &dynamic.FailoverError{},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "failover1", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami4", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "failover1", Path: "/spec/failover/fallback"},
									},
									ServicePort: "8080",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami5", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "failover1", Path: "/spec/failover/service"},
									},
									ServicePort: "8080",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "weighted services in a failover",
			paths: []string{"with_failover2.yml"},
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
							Service:     "default-failover1",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-failover1": {
							Failover: &dynamic.Failover{
								Service:  "default-wrr1",
								Fallback: "default-wrr2",
								Errors:   &dynamic.FailoverError{},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "failover1", Path: "/spec"},
									},
								},
							},
						},
						"default-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami4-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec"},
									},
								},
							},
						},
						"default-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami5-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr2", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami4", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr1", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "8080",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami5", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "wrr2", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "8080",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "failover with errors configuration",
			paths: []string{"with_failover_errors.yml"},
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
							Service:     "default-failover1",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-failover1": {
							Failover: &dynamic.Failover{
								Service:  "default-whoami5-8080",
								Fallback: "default-whoami4-8080",
								Errors: &dynamic.FailoverError{
									Status:              []string{"500-504", "404"},
									MaxRequestBodyBytes: new(int64(1048576)),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "failover1", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami4", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "failover1", Path: "/spec/failover/fallback"},
									},
									ServicePort: "8080",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami5", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "failover1", Path: "/spec/failover/service"},
									},
									ServicePort: "8080",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-77c62dfe9517144aeeaa": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: new(10),
									},
									{
										Name:   "default-whoami2-8080",
										Weight: new(0),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "8080",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:       "One ingress Route with two different services, with weights, safe naming",
			paths:      []string{"services.yml", "with_two_services_weight.yml"},
			safeNaming: true,
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
						"default_test.route_0": {
							EntryPoints: []string{"web"},
							Service:     "default_test.route_0_wrr",
							Rule:        "Host(`foo.com`) && PathPrefix(`/foo`)",
							Priority:    12,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default_test.route_0_wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default_test.route_0_wrr_0_default_whoami_80",
										Weight: new(10),
									},
									{
										Name:   "default_test.route_0_wrr_1_default_whoami2_8080",
										Weight: new(0),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
									},
								},
							},
						},
						"default_test.route_0_wrr_0_default_whoami_80": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
						"default_test.route_0_wrr_1_default_whoami2_8080": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "8080",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
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
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-test-route-shared-6f97418635c7e18853da": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-shared-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-test-route-shared-1f773b7f0ac1aad6d729": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-shared-1f773b7f0ac1aad6d729",
							Rule:        "Host(`bar.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-test-route-shared-b539448199544049f6d7": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-shared-b539448199544049f6d7",
							Rule:        "Host(`baz.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/2"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-test-route-shared-ce4a3976e8b9cf98a2c5": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-shared-ce4a3976e8b9cf98a2c5",
							Rule:        "Host(`qux.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/3"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-shared-6f97418635c7e18853da": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: new(1),
									},
									{
										Name:   "default-whoami2-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/0/services"},
									},
								},
							},
						},
						"default-test-route-shared-1f773b7f0ac1aad6d729": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: new(1),
									},
									{
										Name:   "default-whoami2-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/1/services"},
									},
								},
							},
						},
						"default-test-route-shared-b539448199544049f6d7": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: new(1),
									},
									{
										Name:   "cross-ns-whoami-80",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/2/services"},
									},
								},
							},
						},
						"default-test-route-shared-ce4a3976e8b9cf98a2c5": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: new(1),
									},
									{
										Name:   "default-whoami-80",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/3/services"},
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
								PassHostHeader: new(false),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/0/services/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/1/services/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/2/services/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/3/services/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/3/services/1"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/0/services/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/1/services/1"},
									},
									ServicePort: "8080",
								},
							},
						},
						"cross-ns-whoami-80": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "cross-ns", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/2/services/1"},
									},
									ServicePort: "80",
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
			desc:                "routes referencing the same Kubernetes Service with different options, namespaces, or twice, safe naming",
			paths:               []string{"services.yml", "with_shared_kube_service.yml"},
			allowCrossNamespace: true,
			safeNaming:          true,
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
						"default_test.route.shared_0": {
							EntryPoints: []string{"web"},
							Service:     "default_test.route.shared_0_wrr",
							Rule:        "Host(`foo.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/tls"},
									},
								},
							},
						},
						"default_test.route.shared_1": {
							EntryPoints: []string{"web"},
							Service:     "default_test.route.shared_1_wrr",
							Rule:        "Host(`bar.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/tls"},
									},
								},
							},
						},
						"default_test.route.shared_2": {
							EntryPoints: []string{"web"},
							Service:     "default_test.route.shared_2_wrr",
							Rule:        "Host(`baz.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/2"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/tls"},
									},
								},
							},
						},
						"default_test.route.shared_3": {
							EntryPoints: []string{"web"},
							Service:     "default_test.route.shared_3_wrr",
							Rule:        "Host(`qux.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/3"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default_test.route.shared_0_wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default_test.route.shared_0_wrr_0_default_whoami_80",
										Weight: new(1),
									},
									{
										Name:   "default_test.route.shared_0_wrr_1_default_whoami2_8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/0/services"},
									},
								},
							},
						},
						"default_test.route.shared_0_wrr_0_default_whoami_80": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
						"default_test.route.shared_0_wrr_1_default_whoami2_8080": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "8080",
								},
							},
						},
						"default_test.route.shared_1_wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default_test.route.shared_1_wrr_0_default_whoami_80",
										Weight: new(1),
									},
									{
										Name:   "default_test.route.shared_1_wrr_1_default_whoami2_8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/1/services"},
									},
								},
							},
						},
						"default_test.route.shared_1_wrr_0_default_whoami_80": {
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
								ServersTransport: "default_mtls",
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/1/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
						"default_test.route.shared_1_wrr_1_default_whoami2_8080": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/1/services/1"},
									},
									ServicePort: "8080",
								},
							},
						},
						"default_test.route.shared_2_wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default_test.route.shared_2_wrr_0_default_whoami_80",
										Weight: new(1),
									},
									{
										Name:   "default_test.route.shared_2_wrr_1_cross-ns_whoami_80",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/2/services"},
									},
								},
							},
						},
						"default_test.route.shared_2_wrr_0_default_whoami_80": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/2/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
						"default_test.route.shared_2_wrr_1_cross-ns_whoami_80": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "cross-ns", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/2/services/1"},
									},
									ServicePort: "80",
								},
							},
						},
						"default_test.route.shared_3_wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default_test.route.shared_3_wrr_0_default_whoami_80",
										Weight: new(1),
									},
									{
										Name:   "default_test.route.shared_3_wrr_1_default_whoami_80",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/3/services"},
									},
								},
							},
						},
						"default_test.route.shared_3_wrr_0_default_whoami_80": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/3/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
						"default_test.route.shared_3_wrr_1_default_whoami_80": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.shared", Path: "/spec/routes/3/services/1"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
							KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:       "TLS with tls options, safe naming",
			paths:      []string{"services.yml", "with_tls_options.yml"},
			safeNaming: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"default_foo": {
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
						"default_test.route_0": {
							EntryPoints: []string{"web"},
							Service:     "default_test.route_0_lb",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "default_foo",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default_test.route_0_lb": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:                         "TLS with default tls options in the allowed namespace",
			paths:                        []string{"services.yml", "with_default_tls_options.yml"},
			defaultTLSResourcesNamespace: "foo",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:                         "TLS with default tls options in a not allowed namespace",
			paths:                        []string{"services.yml", "with_default_tls_options.yml"},
			defaultTLSResourcesNamespace: "default",
			expected: &dynamic.Configuration{
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:                         "TLS with two default tls options and an allowed namespace",
			paths:                        []string{"services.yml", "with_default_tls_options.yml", "with_default_tls_options_default_namespace.yml"},
			defaultTLSResourcesNamespace: "default",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitls", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "443",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami3", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8443",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:         "Simple Ingress Route, with ingressClassName",
			paths:        []string{"services.yml", "with_ingressclassname.yml"},
			ingressClass: "traefik-lb",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:         "Simple Ingress Route, with ingressClassName and deprecated annotation",
			paths:        []string{"services.yml", "with_ingressclassname_and_deprecated_annotation.yml"},
			ingressClass: "traefik-lb",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:         "Simple Ingress Route, with deprecated annotation only",
			paths:        []string{"services.yml", "with_deprecated_annotation_only.yml"},
			ingressClass: "traefik-lb",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
								Users: dynamic.Users{"urn:k8s:secret:authsecret:users", "urn:k8s:secret:authsecret:users"},
							},
							URNRefs: map[string]string{
								"/basicAuth/users/0": "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"/basicAuth/users/1": "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "basicauth", Path: "/spec"},
									},
								},
							},
						},
						"default-digestauth": {
							DigestAuth: &dynamic.DigestAuth{
								Users: dynamic.Users{"urn:k8s:secret:authsecret:users", "urn:k8s:secret:authsecret:users"},
							},
							URNRefs: map[string]string{
								"/digestAuth/users/0": "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
								"/digestAuth/users/1": "test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "digestauth", Path: "/spec"},
									},
								},
							},
						},
						"default-forwardauth": {
							ForwardAuth: &dynamic.ForwardAuth{
								Address:     "test.com",
								MaxBodySize: new(int64(-1)),
								HeaderField: "X-Header-Field",
								TLS: &dynamic.ClientTLS{
									CA:   "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----",
									Cert: "urn:k8s:secret:tlssecret:tls.crt",
									Key:  "urn:k8s:secret:tlssecret:tls.key",
								},
							},
							URNRefs: map[string]string{
								"/forwardAuth/tls/cert": "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----",
								"/forwardAuth/tls/key":  "-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "forwardauth", Path: "/spec"},
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
								"test-secret": map[string]any{
									"user":   "admin",
									"secret": "urn:k8s:secret:name:key",
								},
							},
							URNRefs: map[string]string{
								"/plugin/test-secret/secret": "this_is_the_secret",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-secret", Path: "/spec"},
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
								"test-secret": map[string]any{
									"secret_0": map[string]any{
										"secret_1": map[string]any{
											"secret_2": map[string]any{
												"user":   "admin",
												"secret": "urn:k8s:secret:name:key",
											},
										},
									},
								},
							},
							URNRefs: map[string]string{
								"/plugin/test-secret/secret_0/secret_1/secret_2/secret": "this_is_the_very_deep_secret",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-secret", Path: "/spec"},
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
								"test-secret": map[string]any{
									"secret": []any{"urn:k8s:secret:name:key1", "urn:k8s:secret:name:key2"},
								},
							},
							URNRefs: map[string]string{
								"/plugin/test-secret/secret/0": "secret_data1",
								"/plugin/test-secret/secret/1": "secret_data2",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-secret", Path: "/spec"},
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
								"test-secret": map[string]any{
									"users": []any{
										map[string]any{
											"name":   "admin",
											"secret": "urn:k8s:secret:name:key1",
										},
										map[string]any{
											"name":   "user",
											"secret": "urn:k8s:secret:name:key2",
										},
									},
								},
							},
							URNRefs: map[string]string{
								"/plugin/test-secret/users/0/secret": "admin_password",
								"/plugin/test-secret/users/1/secret": "user_password",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-secret", Path: "/spec"},
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
			desc:  "Simple Ingress Route, with tap middleware",
			paths: []string{"services.yml", "with_tap.yml"},
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
						"default-tap": {
							Tap: &dynamic.Tap{
								Request: &dynamic.TapRequest{
									Service:             "default-tap-tap-request-service",
									Path:                "/records/requests",
									RecordBody:          true,
									MaxRecordBodySize:   new(int64(1024)),
									RejectOnRecordError: true,
									Timeout:             ptypes.Duration(5 * time.Second),
								},
								Response: &dynamic.TapResponse{
									Service:        "default-tap-tap-response-service",
									Path:           "/",
									Timeout:        ptypes.Duration(30 * time.Second),
									RequestHeaders: []string{"X-Request-Id"},
								},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "tap", Path: "/spec"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-tap-tap-request-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "tap", Path: "/spec/tap/request/service"},
									},
									ServicePort: "80",
								},
							},
						},
						"default-tap-tap-response-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "tap", Path: "/spec/tap/response/service"},
									},
									ServicePort: "80",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "Simple Ingress Route, with test middleware read config from configMap",
			paths: []string{"services.yml", "with_plugin_read_configmap.yml"},
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
						"default-test-configmap": {
							Plugin: map[string]dynamic.PluginConf{
								"test-configmap": map[string]any{
									"instructions": "Series are selected by their OpenTelemetry name here.",
									"notes":        "the_bytes_form",
									"literal":      "urn:k8s:not-a-scheme:grounding:instructions",
								},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-configmap", Path: "/spec"},
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
			desc:  "Simple Ingress Route, with test middleware read config from configMap that does not exist",
			paths: []string{"services.yml", "with_plugin_read_not_exist_configmap.yml"},
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
			desc:  "Simple Ingress Route, with test middleware read config from configMap without the key",
			paths: []string{"services.yml", "with_plugin_read_configmap_missing_key.yml"},
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
			desc:  "Simple Ingress Route, with test middleware read config from a malformed configMap URN",
			paths: []string{"services.yml", "with_plugin_read_malformed_configmap_urn.yml"},
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
								Service:             "default-errorpage-errorpage-service",
								Query:               "query",
								ErrorRequestHeaders: []string{"X-Foo-Bar", "X-Bar-Foo"},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "errorpage", Path: "/spec"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "errorpage", Path: "/spec/errors/service"},
									},
									ServicePort: "80",
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
						"default-errorpage": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "500"},
								Service: "default-errorpage-wrr",
								Query:   "query",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "errorpage", Path: "/spec"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-errorpage-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "errorpage-wrr", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "errorpage-wrr", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader:     new(false),
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: ptypes.Duration(10 * time.Second)},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:       "TLS with tls store, safe naming",
			paths:      []string{"services.yml", "with_tls_store.yml"},
			safeNaming: true,
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
						"default_test.route_0": {
							EntryPoints: []string{"web"},
							Service:     "default_test.route_0_lb",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default_test.route_0_lb": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:  "TLS with a non-default named tls store",
			paths: []string{"with_named_tls_store.yml"},
			expected: &dynamic.Configuration{
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{
						"default-mystore": {
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
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:       "TLS with a non-default named tls store, safe naming",
			paths:      []string{"with_named_tls_store.yml"},
			safeNaming: true,
			expected: &dynamic.Configuration{
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{
						"default_mystore": {
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
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
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
							CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
							KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							Stores:   []string{"default"},
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-test-route-default-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-default-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:                         "TLS with tls store in the allowed namespace",
			paths:                        []string{"services.yml", "with_tls_store.yml"},
			defaultTLSResourcesNamespace: "default",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:                         "TLS with tls store in a not allowed namespace",
			paths:                        []string{"services.yml", "with_default_tls_store.yml"},
			defaultTLSResourcesNamespace: "default",
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
						"default-test-route-default-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-default-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:                         "TLS with tls store default two times and an allowed namespace",
			paths:                        []string{"services.yml", "with_tls_store.yml", "with_default_tls_store.yml"},
			defaultTLSResourcesNamespace: "default",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-test-route-default-6b204d94623b3df4370c": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-default-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS:         &dynamic.RouterTLSConfig{},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.default", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external-svc", Path: "/spec"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external-svc-with-http", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external-svc-with-https", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "443",
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
							CipherSuites:        []string{"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256", "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"},
							MinVersion:          "VersionTLS11",
							MaxVersion:          "VersionTLS12",
							MaxIdleConnsPerHost: 42,
							DisableHTTP2:        true,
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:           ptypes.Duration(42 * time.Second),
								ResponseHeaderTimeout: ptypes.Duration(42 * time.Second),
								IdleConnTimeout:       ptypes.Duration(42 * time.Millisecond),
								ReadIdleTimeout:       ptypes.Duration(42 * time.Second),
								PingTimeout:           ptypes.Duration(42 * time.Second),
								ReadTimeout:           ptypes.Duration(42 * time.Second),
								WriteTimeout:          ptypes.Duration(42 * time.Second),
							},
							PeerCertURI: "foo://bar",
							PeerCertSANs: []tls.SAN{
								{
									Type:  tls.SANDNSNameType,
									Value: "foo.com",
								},
								{
									Type:  tls.SANURIType,
									Value: "foo://bar",
								},
							},
							Spiffe: &dynamic.Spiffe{
								IDs: []string{
									"spiffe://foo/buz",
									"spiffe://bar/biz",
								},
								TrustDomain: "spiffe://lol",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "ServersTransport", Namespace: "foo", Name: "test", Path: "/spec"},
									},
								},
							},
						},
						"default-test": {
							ServerName: "test",
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(30 * time.Second),
								IdleConnTimeout: ptypes.Duration(90 * time.Second),
								PingTimeout:     ptypes.Duration(15 * time.Second),
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "ServersTransport", Namespace: "default", Name: "test", Path: "/spec"},
									},
								},
							},
						},
					},
					Routers: map[string]*dynamic.Router{
						"default-test-route-6f97418635c7e18853da": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "default-test",
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external-svc-with-https", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "443",
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "default-default-test",
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitls", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "443",
								},
							},
						},
						"default-test-route-6f97418635c7e18853da": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-external-svc-with-https-443",
										Weight: new(1),
									},
									{
										Name:   "default-whoamitls-443",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
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
			desc:       "ServersTransport, safe naming",
			paths:      []string{"services.yml", "with_servers_transport.yml"},
			safeNaming: true,
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
						"foo_test": {
							ServerName:         "test",
							InsecureSkipVerify: true,
							RootCAs:            []types.FileOrContent{"TESTROOTCAS0", "TESTROOTCAS1", "TESTROOTCAS2", "TESTROOTCAS3", "TESTROOTCAS5", "TESTALLCERTS", "TESTROOTCASFROMCONFIGMAP", "TESTROOTCAS6"},
							Certificates: tls.Certificates{
								{CertFile: "TESTCERT1", KeyFile: "TESTKEY1"},
								{CertFile: "TESTCERT2", KeyFile: "TESTKEY2"},
								{CertFile: "TESTCERT3", KeyFile: "TESTKEY3"},
							},
							CipherSuites:        []string{"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256", "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"},
							MinVersion:          "VersionTLS11",
							MaxVersion:          "VersionTLS12",
							MaxIdleConnsPerHost: 42,
							DisableHTTP2:        true,
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:           ptypes.Duration(42 * time.Second),
								ResponseHeaderTimeout: ptypes.Duration(42 * time.Second),
								IdleConnTimeout:       ptypes.Duration(42 * time.Millisecond),
								ReadIdleTimeout:       ptypes.Duration(42 * time.Second),
								PingTimeout:           ptypes.Duration(42 * time.Second),
								ReadTimeout:           ptypes.Duration(42 * time.Second),
								WriteTimeout:          ptypes.Duration(42 * time.Second),
							},
							PeerCertURI: "foo://bar",
							PeerCertSANs: []tls.SAN{
								{
									Type:  tls.SANDNSNameType,
									Value: "foo.com",
								},
								{
									Type:  tls.SANURIType,
									Value: "foo://bar",
								},
							},
							Spiffe: &dynamic.Spiffe{
								IDs: []string{
									"spiffe://foo/buz",
									"spiffe://bar/biz",
								},
								TrustDomain: "spiffe://lol",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "ServersTransport", Namespace: "foo", Name: "test", Path: "/spec"},
									},
								},
							},
						},
						"default_test": {
							ServerName: "test",
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(30 * time.Second),
								IdleConnTimeout: ptypes.Duration(90 * time.Second),
								PingTimeout:     ptypes.Duration(15 * time.Second),
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "ServersTransport", Namespace: "default", Name: "test", Path: "/spec"},
									},
								},
							},
						},
					},
					Routers: map[string]*dynamic.Router{
						"default_test.route_0": {
							EntryPoints: []string{"foo"},
							Service:     "default_test.route_0_wrr",
							Rule:        "Host(`foo.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default_test.route_0_wrr_0_default_external-svc-with-https_443": {
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
								ServersTransport: "default_test",
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external-svc-with-https", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "443",
								},
							},
						},
						"default_test.route_0_wrr_1_default_whoamitls_443": {
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
								ServersTransport: "default_default-test",
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitls", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "443",
								},
							},
						},
						"default_test.route_0_wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default_test.route_0_wrr_0_default_external-svc-with-https_443",
										Weight: new(1),
									},
									{
										Name:   "default_test.route_0_wrr_1_default_whoamitls_443",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami-without-endpointslice-endpoints", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami-nil-port-name", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami-nil-port-value", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami-svc-multiple-endpointaddresses", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami-svc-duplicated-endpointaddresses", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8080",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-test-errorpage": {
							Errors: &dynamic.ErrorPage{
								Service: "default-test-errorpage-errorpage-service",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-errorpage", Path: "/spec"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-test-route-6b204d94623b3df4370c": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-test-weighted",
										Weight: new(1),
									},
									{
										Name:   "default-test-mirror",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
									},
								},
							},
						},
						"default-test-errorpage-errorpage-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami-without-endpointslice-endpoints", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-errorpage", Path: "/spec/errors/service"},
									},
									ServicePort: "80",
								},
							},
						},
						"default-test-weighted": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-without-endpointslice-endpoints-80",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "test-weighted", Path: "/spec"},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "test-mirror", Path: "/spec"},
									},
								},
							},
						},
						"default-whoami-without-endpointslice-endpoints-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami-without-endpointslice-endpoints", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "test-mirror", Path: "/spec/mirroring"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "test-mirror", Path: "/spec/mirroring/mirrors/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "test-weighted", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-test2-route-4f06607bbc69f34a4db5": {
							EntryPoints: []string{"web"},
							Service:     "default-test2-route-4f06607bbc69f34a4db5",
							Rule:        "Host(`traefik-service`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test2-route-840425136fbd5d85a4ad": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: new(1),
									},
									{
										Name:   "default-whoami2-8080",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/1/services"},
									},
								},
							},
						},
						"default-test2-route-4f06607bbc69f34a4db5": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-sticky",
										Weight: new(1),
									},
									{
										Name:   "default-sticky-default",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/0/services"},
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
										Path:     new("/foo"),
									},
								},
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami3-8443",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "sticky", Path: "/spec"},
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
										Path:     new("/"),
									},
								},
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami3-8443",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "sticky-default", Path: "/spec"},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/1/services/1"},
									},
									ServicePort: "8080",
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test2.route", Path: "/spec/routes/1/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami3", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "sticky", Path: "/spec/weighted/services/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "sticky-default", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "8443",
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
						"default-parent-single-3c07cfffe8e5f876a01e": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`parent.example.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-single", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-single", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-single", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-single", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-child-single-2bba0a3de1b50b70a519": {
							Service:    "default-child-single-2bba0a3de1b50b70a519",
							Rule:       "Path(`/api`)",
							ParentRefs: []string{"default-parent-single-3c07cfffe8e5f876a01e"},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-single", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-single", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-single", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-single", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-child-single-2bba0a3de1b50b70a519": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "child-service", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-single", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "9000",
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
						"default-parent-multi-4aac0d541c2b669a2d5d": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`api.example.com`) && PathPrefix(`/v1`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-multi", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-multi", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-multi", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-multi", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-parent-multi-0af1ca0a94f5b87a125e": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`api.example.com`) && PathPrefix(`/v2`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-multi", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-multi", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-multi", Path: "/spec/routes/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-multi", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-child-multi-routes-b0479051e6a353d66211": {
							Service:    "default-child-multi-routes-b0479051e6a353d66211",
							Rule:       "Path(`/users`)",
							ParentRefs: []string{"default-parent-multi-4aac0d541c2b669a2d5d", "default-parent-multi-0af1ca0a94f5b87a125e"},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-multi-routes", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-multi-routes", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-multi-routes", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-multi-routes", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-child-multi-routes-b0479051e6a353d66211": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "users-service", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-multi-routes", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "9000",
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
						"default-parent-a-629990b524bf9a1a8d27": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`a.example.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-a", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-a", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-a", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-a", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-parent-b-add617f9b95cff009054": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`b.example.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-b", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-b", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-b", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-b", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-child-multi-parents-8013b5025acddd1761d1": {
							Service:    "default-child-multi-parents-8013b5025acddd1761d1",
							Rule:       "Path(`/shared`)",
							ParentRefs: []string{"default-parent-a-629990b524bf9a1a8d27", "default-parent-b-add617f9b95cff009054"},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-multi-parents", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-multi-parents", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-multi-parents", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-multi-parents", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-child-multi-parents-8013b5025acddd1761d1": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "shared-service", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-multi-parents", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "9000",
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
						"ns-a-parent-cross-74575ab54671a3ede28c": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`cross.example.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "ns-a", Name: "parent-cross", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "ns-a", Name: "parent-cross", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "ns-a", Name: "parent-cross", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "ns-a", Name: "parent-cross", Path: "/spec/tls"},
									},
								},
							},
						},
						"ns-b-child-cross-allowed-0bad04de665623bf2362": {
							Service:    "ns-b-child-cross-allowed-0bad04de665623bf2362",
							Rule:       "Path(`/cross`)",
							ParentRefs: []string{"ns-a-parent-cross-74575ab54671a3ede28c"},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "ns-b", Name: "child-cross-allowed", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "ns-b", Name: "child-cross-allowed", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "ns-b", Name: "child-cross-allowed", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "ns-b", Name: "child-cross-allowed", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"ns-b-child-cross-allowed-0bad04de665623bf2362": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "ns-b", Name: "cross-service", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "ns-b", Name: "child-cross-allowed", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "9000",
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
						"ns-a-parent-cross-74575ab54671a3ede28c": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`cross.example.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "ns-a", Name: "parent-cross", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "ns-a", Name: "parent-cross", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "ns-a", Name: "parent-cross", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "ns-a", Name: "parent-cross", Path: "/spec/tls"},
									},
								},
							},
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
						"default-parent-default-9b8ab283eeed3eb66561": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`default.example.com`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-default", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-default", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-default", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "parent-default", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-child-same-9234eba1edcfbd8a7723": {
							Service:    "default-child-same-9234eba1edcfbd8a7723",
							Rule:       "Path(`/same`)",
							ParentRefs: []string{"default-parent-default-9b8ab283eeed3eb66561"},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-same", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-same", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-same", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-same", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-child-same-9234eba1edcfbd8a7723": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "same-service", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "child-same", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "9000",
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
						"default-test-route-55869f6407935ccfa805": {
							EntryPoints: []string{"web"},
							Service:     "default-test-route-55869f6407935ccfa805",
							Rule:        "Host(`foo.com`) && PathPrefix(`/leasttime`)",
							Priority:    12,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-test-route-55869f6407935ccfa805": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8080",
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
				IngressClass:                 test.ingressClass,
				AllowCrossNamespace:          test.allowCrossNamespace,
				AllowExternalNameServices:    true,
				AllowEmptyServices:           test.allowEmptyServices,
				CrossProviderNamespaces:      test.crossProviderNamespaces,
				DefaultTLSResourcesNamespace: test.defaultTLSResourcesNamespace,
				SafeNaming:                   &test.safeNaming,
			}
			require.NoError(t, p.Init())

			conf := p.loadConfigurationFromCRD(t.Context(), client)
			assert.Equal(t, test.expected, conf)

			for name, middleware := range conf.HTTP.Middlewares {
				_, err := middleware.ResolveURNRefs()
				assert.NoError(t, err, name)
			}
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
					Metadata: &dynamic.Metadata{
						Kubernetes: &dynamic.KubernetesMetadata{
							Sources: dynamic.KubernetesSources{
								{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
								{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
								{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
								{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
							},
						},
					},
				},
			},
			Middlewares: map[string]*dynamic.Middleware{},
			Services: map[string]*dynamic.Service{
				"default-test-route-6b204d94623b3df4370c": {
					LoadBalancer: &dynamic.ServersLoadBalancer{
						Strategy:       dynamic.BalancerStrategyWRR,
						PassHostHeader: new(true),
						ResponseForwarding: &dynamic.ResponseForwarding{
							FlushInterval: ptypes.Duration(100 * time.Millisecond),
						},
					},
					Metadata: &dynamic.ServiceMetadata{
						Kubernetes: &dynamic.KubernetesServiceMetadata{
							Sources: dynamic.KubernetesSources{
								{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami-svc-multiple-endpointslices", Path: "/spec/ports/1"},
								{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
							},
							ServicePort: "8080",
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
		safeNaming         bool
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
			desc:       "Simple Ingress Route, with foo entrypoint, safe naming",
			paths:      []string{"udp/services.yml", "udp/simple.yml"},
			safeNaming: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default_test.route_0": {
							EntryPoints: []string{"foo"},
							Service:     "default_test.route_0_lb",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default_test.route_0_lb": {
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
						},
						"default-test.route-1": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-1",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/1"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/1/services/0"},
									},
									ServicePort: "8080",
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
			desc:         "Simple UDP Ingress Route, with ingressClassName",
			paths:        []string{"udp/services.yml", "udp/with_ingressclassname.yml"},
			ingressClass: "traefik-lb",
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
			desc:         "Simple UDP Ingress Route, with ingressClassName and deprecated annotation",
			paths:        []string{"udp/services.yml", "udp/with_ingressclassname_and_deprecated_annotation.yml"},
			ingressClass: "traefik-lb",
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
			desc:         "Simple UDP Ingress Route, with deprecated annotation only",
			paths:        []string{"udp/services.yml", "udp/with_deprecated_annotation_only.yml"},
			ingressClass: "traefik-lb",
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
							Weighted: &dynamic.UDPWeightedRoundRobin{
								Services: []dynamic.UDPWRRService{
									{
										Name:   "default-test.route-0-whoamiudp-8000",
										Weight: new(2),
									},
									{
										Name:   "default-test.route-0-whoamiudp2-8080",
										Weight: new(3),
									},
								},
							},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "8080",
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
			desc:       "One ingress Route with two different services, safe naming",
			paths:      []string{"udp/services.yml", "udp/with_two_services.yml"},
			safeNaming: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default_test.route_0": {
							EntryPoints: []string{"foo"},
							Service:     "default_test.route_0_wrr",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default_test.route_0_wrr": {
							Weighted: &dynamic.UDPWeightedRoundRobin{
								Services: []dynamic.UDPWRRService{
									{
										Name:   "default_test.route_0_wrr_0_default_whoamiudp_8000",
										Weight: new(2),
									},
									{
										Name:   "default_test.route_0_wrr_1_default_whoamiudp2_8080",
										Weight: new(3),
									},
								},
							},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
									},
								},
							},
						},
						"default_test.route_0_wrr_0_default_whoamiudp_8000": {
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
								},
							},
						},
						"default_test.route_0_wrr_1_default_whoamiudp2_8080": {
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "8080",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
							Weighted: &dynamic.UDPWeightedRoundRobin{
								Services: []dynamic.UDPWRRService{
									{
										Name:   "default-test.route-0-whoamiudp-8000",
										Weight: new(2),
									},
									{
										Name:   "default-test.route-0-whoamiudp2-8080",
										Weight: new(3),
									},
									{
										Name:   "default-test.route-0-whoamiudp3-8083",
										Weight: new(4),
									},
								},
							},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services"},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp2", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/1"},
									},
									ServicePort: "8080",
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "ns3", Name: "whoamiudp3", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/2"},
									},
									ServicePort: "8083",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external-svc-udp", Path: "/spec"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external.service.with.port.udp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp-without-endpointslice-endpoints", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
			desc:               "Ingress Route, nil endpointslice port name",
			allowEmptyServices: true,
			paths:              []string{"udp/services.yml", "udp/with_nil_endpointslice_port_name.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp-nil-port-name", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
			desc:               "Ingress Route, nil endpointslice port value",
			allowEmptyServices: true,
			paths:              []string{"udp/services.yml", "udp/with_nil_endpointslice_port_value.yml"},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"default-test.route-0": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-0",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.UDPService{
						"default-test.route-0": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamiudp-nil-port-value", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
				SafeNaming:                &test.safeNaming,
			}
			require.NoError(t, p.Init())

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
		safeNaming          bool
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-test-crossnamespace-route-4932ffbbcd99474df323": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-4932ffbbcd99474df323",
							Rule:        "Host(`foo.com`) && PathPrefix(`/chain`)",
							Priority:    12,
							Middlewares: []string{
								"default-test-chain",
								"default-test-chain-cross-provider",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/3"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"cross-ns-stripprefix": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/stripit"},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "cross-ns", Name: "stripprefix", Path: "/spec"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/1/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
						"default-test-crossnamespace-route-4932ffbbcd99474df323": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/3/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-test-crossnamespace-route-9313b71dbe6a649d5049": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-9313b71dbe6a649d5049",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bir`)",
							Priority:    12,
							Middlewares: []string{"default-test-errorpage"},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-test-crossnamespace-route-a1963878aac7331b7950": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-a1963878aac7331b7950",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bur`)",
							Priority:    12,
							Middlewares: []string{"cross-ns-stripprefix@kubernetescrd"},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/2"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-test-crossnamespace-route-4932ffbbcd99474df323": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-crossnamespace-route-4932ffbbcd99474df323",
							Rule:        "Host(`foo.com`) && PathPrefix(`/chain`)",
							Priority:    12,
							Middlewares: []string{
								"default-test-chain",
								"default-test-chain-cross-provider",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/3"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"cross-ns-stripprefix": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/stripit"},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "cross-ns", Name: "stripprefix", Path: "/spec"},
									},
								},
							},
						},
						"default-test-errorpage": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"500-599"},
								Service: "default-test-errorpage-errorpage-service",
								Query:   "/{status}.html",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-errorpage", Path: "/spec"},
									},
								},
							},
						},
						"default-test-chain": {
							Chain: &dynamic.Chain{
								Middlewares: []string{"cross-ns-stripprefix"},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-chain", Path: "/spec"},
									},
								},
							},
						},
						"default-test-chain-cross-provider": {
							Chain: &dynamic.Chain{
								Middlewares: []string{"other-middleware@kubernetescrd"},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-chain-cross-provider", Path: "/spec"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/1/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "cross-ns", Name: "whoami-svc", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-errorpage", Path: "/spec/errors/service"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/2/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
						"default-test-crossnamespace-route-4932ffbbcd99474df323": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/3/services/0"},
									},
									ServicePort: "80",
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
			desc:                "HTTP middleware cross namespace allowed, safe naming",
			paths:               []string{"services.yml", "with_middleware_cross_namespace.yml"},
			allowCrossNamespace: true,
			safeNaming:          true,
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
						"default_test-crossnamespace.route_0": {
							EntryPoints: []string{"foo"},
							Service:     "default_test-crossnamespace.route_0_lb",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							Middlewares: []string{
								"cross-ns_stripprefix",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default_test-crossnamespace.route_1": {
							EntryPoints: []string{"foo"},
							Service:     "default_test-crossnamespace.route_1_lb",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bir`)",
							Priority:    12,
							Middlewares: []string{"default_test-errorpage"},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default_test-crossnamespace.route_2": {
							EntryPoints: []string{"foo"},
							Service:     "default_test-crossnamespace.route_2_lb",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bur`)",
							Priority:    12,
							Middlewares: []string{"cross-ns-stripprefix@kubernetescrd"},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/2"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default_test-crossnamespace.route_3": {
							EntryPoints: []string{"foo"},
							Service:     "default_test-crossnamespace.route_3_lb",
							Rule:        "Host(`foo.com`) && PathPrefix(`/chain`)",
							Priority:    12,
							Middlewares: []string{
								"default_test-chain",
								"default_test-chain-cross-provider",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/3"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"cross-ns_stripprefix": {
							StripPrefix: &dynamic.StripPrefix{
								Prefixes: []string{"/stripit"},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "cross-ns", Name: "stripprefix", Path: "/spec"},
									},
								},
							},
						},
						"default_test-errorpage": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"500-599"},
								Service: "default_test-errorpage-errorpage-service",
								Query:   "/{status}.html",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-errorpage", Path: "/spec"},
									},
								},
							},
						},
						"default_test-chain": {
							Chain: &dynamic.Chain{
								Middlewares: []string{"cross-ns_stripprefix"},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-chain", Path: "/spec"},
									},
								},
							},
						},
						"default_test-chain-cross-provider": {
							Chain: &dynamic.Chain{
								Middlewares: []string{"other-middleware@kubernetescrd"},
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-chain-cross-provider", Path: "/spec"},
									},
								},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default_test-crossnamespace.route_0_lb": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
						"default_test-crossnamespace.route_1_lb": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/1/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
						"default_test-errorpage-errorpage-service": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "cross-ns", Name: "whoami-svc", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "Middleware", Namespace: "default", Name: "test-errorpage", Path: "/spec/errors/service"},
									},
									ServicePort: "80",
								},
							},
						},
						"default_test-crossnamespace.route_2_lb": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/2/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
						"default_test-crossnamespace.route_3_lb": {
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test-crossnamespace.route", Path: "/spec/routes/3/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "cross-ns-route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "cross-ns-route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "cross-ns-route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "cross-ns-route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-cross-ns-route-1bc3efa892379bb93c6e": {
							EntryPoints: []string{"foo"},
							Service:     "default-cross-ns-route-1bc3efa892379bb93c6e",
							Rule:        "Host(`bar.com`) && PathPrefix(`/foo`)",
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "cross-ns-route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "cross-ns-route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "cross-ns-route", Path: "/spec/routes/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "cross-ns-route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-cross-ns-route-6b204d94623b3df4370c": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80",
										Weight: new(1),
									},
									{
										Name:   "default-tr-svc-wrr1",
										Weight: new(1),
									},
									{
										Name:   "cross-ns-tr-svc-wrr2",
										Weight: new(1),
									},
									{
										Name:   "default-tr-svc-mirror1",
										Weight: new(1),
									},
									{
										Name:   "cross-ns-tr-svc-mirror2",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "cross-ns-route", Path: "/spec/routes/0/services"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "foo-test@kubernetescrd",
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "cross-ns", Name: "whoami-svc", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "cross-ns-route", Path: "/spec/routes/1/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "cross-ns", Name: "whoami-svc", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "cross-ns-route", Path: "/spec/routes/0/services/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc-mirror2", Path: "/spec/mirroring"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc-mirror2", Path: "/spec/mirroring/mirrors/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc-wrr2", Path: "/spec/weighted/services/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "tr-svc-mirror1", Path: "/spec/mirroring/mirrors/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "tr-svc-wrr1", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "80",
								},
							},
						},
						"default-tr-svc-wrr1": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "tr-svc-wrr1", Path: "/spec"},
									},
								},
							},
						},
						"cross-ns-tr-svc-wrr2": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc-wrr2", Path: "/spec"},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "tr-svc-mirror1", Path: "/spec"},
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc-mirror2", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "tr-svc-mirror1", Path: "/spec/mirroring"},
									},
									ServicePort: "80",
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
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc-wrr2", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "cross-ns", Name: "whoami-svc", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc-mirror2", Path: "/spec/mirroring"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc-mirror2", Path: "/spec/mirroring/mirrors/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc-wrr2", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc-mirror2", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "default", Name: "tr-svc-mirror1", Path: "/spec/mirroring"},
									},
									ServicePort: "80",
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "cross-ns-st-cross-ns@kubernetescrd",
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
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
								ReadTimeout:           0,
								WriteTimeout:          0,
							},
							DisableHTTP2: true,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "ServersTransport", Namespace: "cross-ns", Name: "st-cross-ns", Path: "/spec"},
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
								ReadTimeout:           0,
								WriteTimeout:          0,
							},
							DisableHTTP2: true,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "ServersTransport", Namespace: "cross-ns", Name: "st-cross-ns", Path: "/spec"},
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "cross-ns-tr-svc",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"cross-ns-tr-svc": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "cross-ns", Name: "whoami-svc", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "80",
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
						"cross-ns-tr-svc": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "cross-ns-whoami-svc-80",
										Weight: new(1),
									},
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc", Path: "/spec"},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "cross-ns", Name: "whoami-svc", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "TraefikService", Namespace: "cross-ns", Name: "tr-svc", Path: "/spec/weighted/services/0"},
									},
									ServicePort: "80",
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
						"default-test-route-6b204d94623b3df4370c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6b204d94623b3df4370c",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`)",
							Priority:    12,
							TLS: &dynamic.RouterTLSConfig{
								Options: "cross-ns-tls-options-cn",
							},
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoami", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "MiddlewareTCP", Namespace: "default", Name: "ipallowlist", Path: "/spec"},
									},
								},
							},
						},
						"cross-ns-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "MiddlewareTCP", Namespace: "cross-ns", Name: "ipallowlist", Path: "/spec"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
						"default-test.route-f44ce589164e656d231c": {
							EntryPoints: []string{"foo"},
							Service:     "default-test.route-f44ce589164e656d231c",
							Middlewares: []string{"cross-ns-ipallowlist"},
							Rule:        "HostSNI(`bar.com`)",

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/1"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"default-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "MiddlewareTCP", Namespace: "default", Name: "ipallowlist", Path: "/spec"},
									},
								},
							},
						},
						"cross-ns-ipallowlist": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "MiddlewareTCP", Namespace: "cross-ns", Name: "ipallowlist", Path: "/spec"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/1/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "cross-ns", Name: "whoamitcp-cross-ns", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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
								ServersTransport: "cross-ns-st-cross-ns@kubernetescrd",
							},

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{
						"cross-ns-st-cross-ns": {
							DialTimeout:      ptypes.Duration(30 * time.Second),
							DialKeepAlive:    0,
							TerminationDelay: ptypes.Duration(100 * time.Millisecond),

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "ServersTransportTCP", Namespace: "cross-ns", Name: "st-cross-ns", Path: "/spec"},
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{
						"cross-ns-st-cross-ns": {
							DialTimeout:      30000000000,
							DialKeepAlive:    0,
							TerminationDelay: ptypes.Duration(100 * time.Millisecond),

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "ServersTransportTCP", Namespace: "cross-ns", Name: "st-cross-ns", Path: "/spec"},
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "whoamitcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "cross-ns", Name: "whoamiudp-cross-ns", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

			p := Provider{
				AllowCrossNamespace: test.allowCrossNamespace,
				SafeNaming:          &test.safeNaming,
			}
			require.NoError(t, p.Init())

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
			wantMiddlewares:         []string{"default-stripprefix", "foo-addprefix", "basicauth@file", "redirect@file"},
		},
		{
			desc:                    "empty list: cross-provider middleware refs are rejected, IngressRoute is dropped",
			crossProviderNamespaces: []string{},
			wantRouterDropped:       true,
		},
		{
			desc:                    "namespace allowed: cross-provider middleware refs are accepted",
			crossProviderNamespaces: []string{"default"},
			wantMiddlewares:         []string{"default-stripprefix", "foo-addprefix", "basicauth@file", "redirect@file"},
		},
		{
			desc:                    "namespace not allowed: cross-provider middleware refs are rejected, IngressRoute is dropped",
			crossProviderNamespaces: []string{"other"},
			wantRouterDropped:       true,
		},
	}

	for _, safeNaming := range []bool{false, true} {
		for _, test := range testCases {
			desc := test.desc
			if safeNaming {
				desc += ", safe naming"
			}

			t.Run(desc, func(t *testing.T) {
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
					SafeNaming:              &safeNaming,
				}
				require.NoError(t, p.Init())

				conf := p.loadConfigurationFromCRD(t.Context(), client)

				routerKey := "default-test2-route-23c7f4c450289ee29016"
				wantMiddlewares := test.wantMiddlewares
				if safeNaming {
					routerKey = "default_test2.route_0"
					if wantMiddlewares != nil {
						wantMiddlewares = []string{"default_stripprefix", "foo_addprefix", "basicauth@file", "redirect@file"}
					}
				}

				router, ok := conf.HTTP.Routers[routerKey]
				if test.wantRouterDropped {
					assert.False(t, ok)
					return
				}

				assert.True(t, ok)
				assert.Equal(t, wantMiddlewares, router.Middlewares)
			})
		}
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

	for _, safeNaming := range []bool{false, true} {
		for _, test := range testCases {
			desc := test.desc
			if safeNaming {
				desc += ", safe naming"
			}

			t.Run(desc, func(t *testing.T) {
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
					SafeNaming:              &safeNaming,
				}
				require.NoError(t, p.Init())

				conf := p.loadConfigurationFromCRD(t.Context(), client)

				mirrorKey, weightedKey := "foo-mirror-cp", "bar-weighted-cp"
				if safeNaming {
					mirrorKey, weightedKey = "foo_mirror-cp", "bar_weighted-cp"
				}

				_, mirrorOK := conf.HTTP.Services[mirrorKey]
				_, weightedOK := conf.HTTP.Services[weightedKey]

				assert.Equal(t, test.wantMirrorService, mirrorOK)
				assert.Equal(t, test.wantWeightedService, weightedOK)
			})
		}
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
			require.NoError(t, p.Init())

			conf := p.loadConfigurationFromCRD(t.Context(), client)

			router, ok := conf.HTTP.Routers["default-test-route-6b204d94623b3df4370c"]
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
			require.NoError(t, p.Init())

			conf := p.loadConfigurationFromCRD(t.Context(), client)

			router, ok := conf.TCP.Routers["default-test.route-fdd3e9338e47a45efefc"]
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
			require.NoError(t, p.Init())

			conf := p.loadConfigurationFromCRD(t.Context(), client)

			service, ok := conf.HTTP.Services["default-test-route-6b204d94623b3df4370c"]
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
			require.NoError(t, p.Init())

			conf := p.loadConfigurationFromCRD(t.Context(), client)

			service, ok := conf.TCP.Services["default-test.route-fdd3e9338e47a45efefc"]
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
						"default-test-route-6f97418635c7e18853da": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
							Priority:    0,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external-svc-with-http", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external.service.with.port.tcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "external.service.with.port.udp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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
						"default-test-route-6f97418635c7e18853da": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
							Priority:    0,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
								PassHostHeader: new(true),
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "native-svc", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "native-svc-tcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "9000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "native-svc-udp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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
						"default-test-route-6f97418635c7e18853da": {
							EntryPoints: []string{"foo"},
							Service:     "default-test-route-6f97418635c7e18853da",
							Rule:        "Host(`foo.com`)",
							Priority:    0,
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
								PassHostHeader: new(true),
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "nodeport-svc", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "nodeport-svc-tcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "nodeport-svc-udp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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
	refs := k8s.URNRefs{}
	basicAuth, secretErr := createBasicAuthMiddleware(client, "default", &traefikv1alpha1.BasicAuth{Secret: "basic-auth-secret"}, refs)
	require.NoError(t, secretErr)
	assert.Equal(t, dynamic.Users{"urn:k8s:secret:basic-auth-secret:password"}, basicAuth.Users)
	require.Len(t, refs, 1)

	components := strings.Split(refs["/basicAuth/users/0"], ":")
	require.Len(t, components, 2)

	username := components[0]
	hashedPassword := components[1]

	require.Equal(t, "user", username)
	require.Equal(t, "{SHA}W6ph5Mm5Pz8GgiULbPgzG37mj9g=", hashedPassword)
	assert.True(t, auth.CheckSecret("password", hashedPassword))

	// Testing for username/password components in htpasswd secret
	refs = k8s.URNRefs{}
	basicAuth, secretErr = createBasicAuthMiddleware(client, "default", &traefikv1alpha1.BasicAuth{Secret: "auth-secret"}, refs)
	require.NoError(t, secretErr)
	assert.Equal(t, dynamic.Users{"urn:k8s:secret:auth-secret:users", "urn:k8s:secret:auth-secret:users"}, basicAuth.Users)
	require.Len(t, refs, 2)

	components = strings.Split(refs["/basicAuth/users/1"], ":")
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "global-native-lb", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "global-native-lb", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "global-native-lb", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "global-native-lb", Path: "/spec/tls"},
									},
								},
							},
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
								PassHostHeader: new(true),
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "native-svc", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "global-native-lb", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.native-disabled", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.native-disabled", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.native-disabled", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.native-disabled", Path: "/spec/tls"},
									},
								},
							},
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
								PassHostHeader: new(true),
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "native-disabled-svc", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route.native-disabled", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
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
							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/parentRefs"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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
								PassHostHeader: new(true),
							},
							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "native-svc", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRoute", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "80",
								},
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "global-native-lb", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "global-native-lb", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "global-native-lb", Path: "/spec/tls"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "native-svc-tcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "global-native-lb", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "9000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "tcp.route.native-disabled", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "tcp.route.native-disabled", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "tcp.route.native-disabled", Path: "/spec/tls"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "native-disabled-svc-tcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "tcp.route.native-disabled", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "9000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/tls"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "native-svc-tcp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteTCP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "9000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "native-svc-udp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "test.route", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "global-native-lb", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "global-native-lb", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "native-svc-udp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "global-native-lb", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

							Metadata: &dynamic.Metadata{
								Kubernetes: &dynamic.KubernetesMetadata{
									Sources: dynamic.KubernetesSources{
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "udp.route.native-disabled", Path: "/spec/entryPoints"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "udp.route.native-disabled", Path: "/spec/routes/0"},
									},
								},
							},
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

							Metadata: &dynamic.ServiceMetadata{
								Kubernetes: &dynamic.KubernetesServiceMetadata{
									Sources: dynamic.KubernetesSources{
										{Version: "v1", Kind: "Service", Namespace: "default", Name: "native-disabled-svc-udp", Path: "/spec/ports/0"},
										{Group: "traefik.io", Version: "v1alpha1", Kind: "IngressRouteUDP", Namespace: "default", Name: "udp.route.native-disabled", Path: "/spec/routes/0/services/0"},
									},
									ServicePort: "8000",
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

func TestAddToConfig(t *testing.T) {
	service := func(url string, sources ...string) *dynamic.Service {
		svc := &dynamic.Service{
			LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: url}}},
			Metadata:     &dynamic.ServiceMetadata{Kubernetes: &dynamic.KubernetesServiceMetadata{}},
		}
		for _, source := range sources {
			svc.Metadata.Kubernetes.Sources = append(svc.Metadata.Kubernetes.Sources, dynamic.KubernetesRef{Kind: "TraefikService", Name: source})
		}

		return svc
	}

	testCases := []struct {
		desc         string
		existing     *dynamic.Service
		obj          *dynamic.Service
		expected     *dynamic.Service
		expectedWarn bool
	}{
		{
			desc:     "new name",
			obj:      service("http://10.0.0.1:80", "wrr1"),
			expected: service("http://10.0.0.1:80", "wrr1"),
		},
		{
			desc:     "same object",
			existing: service("http://10.0.0.1:80", "wrr1"),
			obj:      service("http://10.0.0.1:80", "wrr1"),
			expected: service("http://10.0.0.1:80", "wrr1"),
		},
		{
			desc:     "shared object",
			existing: service("http://10.0.0.1:80", "wrr1"),
			obj:      service("http://10.0.0.1:80", "wrr2"),
			expected: service("http://10.0.0.1:80", "wrr1", "wrr2"),
		},
		{
			desc:         "collision",
			existing:     service("http://10.0.0.1:80", "wrr1"),
			obj:          service("http://10.0.0.2:80", "wrr2"),
			expected:     service("http://10.0.0.2:80", "wrr1", "wrr2"),
			expectedWarn: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := zerolog.New(&buf)

			objects := map[string]*dynamic.Service{}
			if test.existing != nil {
				objects["svc"] = test.existing
			}

			addToConfig(&logger, "service", "svc", objects, test.obj)

			assert.Equal(t, test.expected, objects["svc"])
			assert.Equal(t, test.expectedWarn, strings.Contains(buf.String(), "Name collision"))
		})
	}
}
