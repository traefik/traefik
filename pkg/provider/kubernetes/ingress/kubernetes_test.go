package ingress

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefikhttp "github.com/traefik/traefik/v3/pkg/muxer/http"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

var _ provider.Provider = (*Provider)(nil)

func pointer[T any](v T) *T { return &v }

func TestLoadConfigurationFromIngresses(t *testing.T) {
	testCases := []struct {
		desc                         string
		ingressClass                 string
		expected                     *dynamic.Configuration
		allowEmptyServices           bool
		disableIngressClassLookup    bool
		disableClusterScopeResources bool
		defaultRuleSyntax            string
		strictPrefixMatching         bool
	}{
		{
			desc: "Empty ingresses",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc: "Ingress one rule host only",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc: "Ingress with a basic rule on one path",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.21.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with annotations",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:        "Path(`/bar`)",
							EntryPoints: []string{"ep1", "ep2"},
							Service:     "testing-service1-80",
							Middlewares: []string{"md1", "md2"},
							Priority:    42,
							RuleSyntax:  "v2",
							TLS: &dynamic.RouterTLSConfig{
								CertResolver: "foobar",
								Domains: []types.Domain{
									{
										Main: "domain.com",
										SANs: []string{"one.domain.com", "two.domain.com"},
									},
									{
										Main: "example.com",
										SANs: []string{"one.example.com", "two.example.com"},
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
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "foobar",
										Secure:   true,
										HTTPOnly: true,
										Path:     pointer("/"),
									},
								},
								Servers: []dynamic.Server{
									{
										URL: "protocol://10.10.0.1:8080",
									},
									{
										URL: "protocol://10.21.0.1:8080",
									},
								},
								ServersTransport: "foobar@file",
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with two different rules with one path",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
						"testing-foo": {
							Rule:    "PathPrefix(`/foo`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.21.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with conflicting routers on host",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar-bar-aba9a7d00e9b06a78e16": {
							Rule:    "HostRegexp(`^[a-zA-Z0-9-]+\\.bar$`) && PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
						"testing-bar-bar-636bf36c00fedaab3d44": {
							Rule:    "Host(`bar`) && PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.21.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with conflicting routers on path",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-foo-bar-d0b30949e54d6a7515ca": {
							Rule:    "PathPrefix(`/foo/bar`)",
							Service: "testing-service1-80",
						},
						"testing-foo-bar-dcd54bae39a6d7557f48": {
							Rule:    "PathPrefix(`/foo-bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.21.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress one rule with two paths",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
						"testing-foo": {
							Rule:    "PathPrefix(`/foo`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.21.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress one rule with one path and one host",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.21.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with one host without path",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-example-com": {
							Rule:    "Host(`example.com`)",
							Service: "testing-example-com-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-example-com-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.11.0.1:80",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress one rule with one host and two paths",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
						"testing-traefik-tchouk-foo": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/foo`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.21.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress Two rules with one host and one path",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
						"testing-traefik-courgette-carotte": {
							Rule:    "Host(`traefik.courgette`) && PathPrefix(`/carotte`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.21.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with two services",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
						"testing-traefik-courgette-carotte": {
							Rule:    "Host(`traefik.courgette`) && PathPrefix(`/carotte`)",
							Service: "testing-service2-8082",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.21.0.1:8080",
									},
								},
							},
						},
						"testing-service2-8082": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.2:8080",
									},
									{
										URL: "http://10.21.0.2:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc:               "Ingress with one service without endpoints subset",
			allowEmptyServices: true,
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
					},
				},
			},
		},
		{
			desc:               "Ingress with backend resource",
			allowEmptyServices: true,
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc:               "Ingress without backend",
			allowEmptyServices: true,
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc: "Ingress with one service without endpoint",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc: "Single Service Ingress (without any rules)",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"default-router": {
							Rule:       "PathPrefix(`/`)",
							RuleSyntax: "default",
							Service:    "default-backend",
							Priority:   math.MinInt32,
						},
					},
					Services: map[string]*dynamic.Service{
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.21.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with port value in backend and no pod replica",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8089",
									},
									{
										URL: "http://10.21.0.1:8089",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with port name in backend and no pod replica",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-tchouk",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-tchouk": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8089",
									},
									{
										URL: "http://10.21.0.1:8089",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with port name in backend and 2 pod replica",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-tchouk",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-tchouk": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8089",
									},
									{
										URL: "http://10.10.0.2:8089",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with two paths using same service and different port name",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-tchouk",
						},
						"testing-traefik-tchouk-foo": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/foo`)",
							Service: "testing-service1-carotte",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-tchouk": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8089",
									},
									{
										URL: "http://10.10.0.2:8089",
									},
								},
							},
						},
						"testing-service1-carotte": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8090",
									},
									{
										URL: "http://10.10.0.2:8090",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with a named port matching subset of service pods",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-tchouk",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-tchouk": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8089",
									},
									{
										URL: "http://10.10.0.2:8089",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "2 ingresses in different namespace with same service name",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-tchouk",
						},
						"toto-toto-traefik-tchouk-bar": {
							Rule:    "Host(`toto.traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "toto-service1-tchouk",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-tchouk": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8089",
									},
									{
										URL: "http://10.10.0.2:8089",
									},
								},
							},
						},
						"toto-service1-tchouk": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.11.0.1:8089",
									},
									{
										URL: "http://10.11.0.2:8089",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with unknown service port name",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc: "Ingress with unknown service port",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc: "Ingress with port invalid for one service",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-port-port": {
							Rule:    "Host(`traefik.port`) && PathPrefix(`/port`)",
							Service: "testing-service1-8080",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-8080": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.0.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "TLS support",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-example-com": {
							Rule:    "Host(`example.com`)",
							Service: "testing-example-com-80",
							TLS:     &dynamic.RouterTLSConfig{},
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-example-com-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.11.0.1:80",
									},
								},
							},
						},
					},
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
			},
		},
		{
			desc: "Ingress with a basic rule on one path with https (port == 443)",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-443",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-443": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.1:8443",
									},
									{
										URL: "https://10.21.0.1:8443",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with a basic rule on one path with https (portname == https)",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-8443",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-8443": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.1:8443",
									},
									{
										URL: "https://10.21.0.1:8443",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with a basic rule on one path with https (portname starts with https)",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},

					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-8443",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-8443": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.1:8443",
									},
									{
										URL: "https://10.21.0.1:8443",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Double Single Service Ingress",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"default-router": {
							Rule:       "PathPrefix(`/`)",
							RuleSyntax: "default",
							Service:    "default-backend",
							Priority:   math.MinInt32,
						},
					},
					Services: map[string]*dynamic.Service{
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.30.0.1:8080",
									},
									{
										URL: "http://10.41.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with default traefik ingressClass",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress without provider traefik ingressClass and unknown annotation",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc:         "Ingress with non matching provider traefik ingressClass and annotation",
			ingressClass: "tchouk",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc:         "Ingress with ingressClass without annotation",
			ingressClass: "tchouk",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc:         "Ingress with ingressClass without annotation",
			ingressClass: "toto",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc: "Ingress with wildcard host",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-foobar-com-bar": {
							Rule:    "HostRegexp(`^[a-zA-Z0-9-]+\\.foobar\\.com$`) && PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL:    "http://10.10.0.1:8080",
										Scheme: "",
										Port:   "",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc:              "Ingress with wildcard host syntax v2",
			defaultRuleSyntax: "v2",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-foobar-com-bar": {
							Rule:    "HostRegexp(`{subdomain:[a-zA-Z0-9-]+}.foobar.com`) && PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL:    "http://10.10.0.1:8080",
										Scheme: "",
										Port:   "",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with multiple ingressClasses",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-foo": {
							Rule:    "PathPrefix(`/foo`)",
							Service: "testing-service1-80",
						},
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc:         "Ingress with ingressClasses filter",
			ingressClass: "traefik-lb2",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-foo": {
							Rule:    "PathPrefix(`/foo`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with prefix pathType",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with empty pathType",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "Path(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with exact pathType",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "Path(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with implementationSpecific pathType",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "Path(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with ingress annotation",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// Duplicate test case with the same fixture as the one above, but with the disableIngressClassLookup option to true.
			// Showing that disabling the ingressClass discovery still allow the discovery of ingresses with ingress annotation.
			desc:                      "Ingress with ingress annotation",
			disableIngressClassLookup: true,
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// Duplicate test case with the same fixture as the one above, but with the disableClusterScopeResources option to true.
			// Showing that disabling the ingressClass discovery still allow the discovery of ingresses with ingress annotation.
			desc:                         "Ingress with ingress annotation",
			disableClusterScopeResources: true,
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with ingressClass",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			// Duplicate test case with the same fixture as the one above, but with the disableIngressClassLookup option to true.
			// Showing that disabling the ingressClass discovery avoid discovering Ingresses with an IngressClass.
			desc:                      "Ingress with ingressClass",
			disableIngressClassLookup: true,
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			// Duplicate test case with the same fixture as the one above, but with the disableClusterScopeResources option to true.
			// Showing that disabling the ingressClass discovery avoid discovering Ingresses with an IngressClass.
			desc:                         "Ingress with ingressClass",
			disableClusterScopeResources: true,
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc: "Ingress with named port",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-foobar",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-foobar": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:4711",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with missing ingressClass",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc: "Ingress with defaultbackend",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"default-router": {
							Rule:       "PathPrefix(`/`)",
							RuleSyntax: "default",
							Priority:   math.MinInt32,
							Service:    "default-backend",
						},
					},
					Services: map[string]*dynamic.Service{
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with endpoint conditions",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.10.0.2:8080",
									},
									{
										URL:    "http://10.10.0.3:8080",
										Fenced: true,
									},
									{
										URL:    "http://10.10.0.4:8080",
										Fenced: true,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with strict prefix matching",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-bar": {
							Rule:    "(Path(`/bar`) || PathPrefix(`/bar/`))",
							Service: "testing-service1-80",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8080",
									},
									{
										URL: "http://10.21.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
			strictPrefixMatching: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			clientMock := newClientMock(generateTestFilename(test.desc))
			p := Provider{
				IngressClass:                 test.ingressClass,
				AllowEmptyServices:           test.allowEmptyServices,
				DisableIngressClassLookup:    test.disableIngressClassLookup,
				DisableClusterScopeResources: test.disableClusterScopeResources,
				DefaultRuleSyntax:            test.defaultRuleSyntax,
				StrictPrefixMatching:         test.strictPrefixMatching,
			}
			conf := p.loadConfigurationFromIngresses(t.Context(), clientMock)

			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadConfigurationFromIngressesWithExternalNameServices(t *testing.T) {
	testCases := []struct {
		desc                      string
		ingressClass              string
		allowExternalNameServices bool
		expected                  *dynamic.Configuration
	}{
		{
			desc: "Ingress with service with externalName",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
		{
			desc:                      "Ingress with service with externalName enabled",
			allowExternalNameServices: true,
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-8080",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-8080": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								Servers: []dynamic.Server{
									{
										URL: "http://traefik.wtf:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with IPv6 endpoints",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-example-com-bar": {
							Rule:    "PathPrefix(`/bar`)",
							Service: "testing-service-bar-8080",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service-bar-8080": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://[2001:0db8:3c4d:0015:0000:0000:1a2f:1a2b]:8080",
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
			},
		},
		{
			desc:                      "Ingress with IPv6 endpoints externalname enabled",
			allowExternalNameServices: true,
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-example-com-foo": {
							Rule:    "PathPrefix(`/foo`)",
							Service: "testing-service-foo-8080",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service-foo-8080": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://[2001:0db8:3c4d:0015:0000:0000:1a2f:2a3b]:8080",
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
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			clientMock := newClientMock(generateTestFilename(test.desc))

			p := Provider{IngressClass: test.ingressClass}
			p.AllowExternalNameServices = test.allowExternalNameServices
			conf := p.loadConfigurationFromIngresses(t.Context(), clientMock)

			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadConfigurationFromIngressesWithNativeLB(t *testing.T) {
	testCases := []struct {
		desc         string
		ingressClass string
		expected     *dynamic.Configuration
	}{
		{
			desc: "Ingress with native service lb",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-8080",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-8080": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								PassHostHeader:     pointer(true),
								Servers: []dynamic.Server{
									{
										URL: "http://10.0.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			clientMock := newClientMock(generateTestFilename(test.desc))

			p := Provider{IngressClass: test.ingressClass}
			conf := p.loadConfigurationFromIngresses(t.Context(), clientMock)

			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadConfigurationFromIngressesWithNodePortLB(t *testing.T) {
	testCases := []struct {
		desc                 string
		clusterScopeDisabled bool
		expected             *dynamic.Configuration
	}{
		{
			desc: "Ingress with node port lb",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-8080",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-8080": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								PassHostHeader:     pointer(true),
								Servers: []dynamic.Server{
									{
										URL: "http://172.16.4.4:32456",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc:                 "Ingress with node port lb cluster scope disabled",
			clusterScopeDisabled: true,
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers:     map[string]*dynamic.Router{},
					Services:    map[string]*dynamic.Service{},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			clientMock := newClientMock(generateTestFilename(test.desc))

			p := Provider{DisableClusterScopeResources: test.clusterScopeDisabled}
			conf := p.loadConfigurationFromIngresses(t.Context(), clientMock)

			assert.Equal(t, test.expected, conf)
		})
	}
}

func generateTestFilename(desc string) string {
	return filepath.Join("fixtures", strings.ReplaceAll(desc, " ", "-")+".yml")
}

func TestGetCertificates(t *testing.T) {
	testIngressWithoutHostname := buildIngress(
		iNamespace("testing"),
		iRules(
			iRule(iHost("ep1.example.com")),
			iRule(iHost("ep2.example.com")),
		),
		iTLSes(
			iTLS("test-secret"),
		),
	)

	testIngressWithoutSecret := buildIngress(
		iNamespace("testing"),
		iRules(
			iRule(iHost("ep1.example.com")),
		),
		iTLSes(
			iTLS("", "foo.com"),
		),
	)

	testCases := []struct {
		desc      string
		ingress   *netv1.Ingress
		client    Client
		result    map[string]*tls.CertAndStores
		errResult string
	}{
		{
			desc:    "api client returns error",
			ingress: testIngressWithoutHostname,
			client: clientMock{
				apiSecretError: errors.New("api secret error"),
			},
			errResult: "failed to fetch secret testing/test-secret: api secret error",
		},
		{
			desc:      "api client doesn't find secret",
			ingress:   testIngressWithoutHostname,
			client:    clientMock{},
			errResult: "secret testing/test-secret does not exist",
		},
		{
			desc:    "entry 'tls.crt' in secret missing",
			ingress: testIngressWithoutHostname,
			client: clientMock{
				secrets: []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-secret",
							Namespace: "testing",
						},
						Data: map[string][]byte{
							"tls.key": []byte("tls-key"),
						},
					},
				},
			},
			errResult: "secret testing/test-secret is missing the following TLS data entries: tls.crt",
		},
		{
			desc:    "entry 'tls.key' in secret missing",
			ingress: testIngressWithoutHostname,
			client: clientMock{
				secrets: []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-secret",
							Namespace: "testing",
						},
						Data: map[string][]byte{
							"tls.crt": []byte("tls-crt"),
						},
					},
				},
			},
			errResult: "secret testing/test-secret is missing the following TLS data entries: tls.key",
		},
		{
			desc:    "secret doesn't provide any of the required fields",
			ingress: testIngressWithoutHostname,
			client: clientMock{
				secrets: []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-secret",
							Namespace: "testing",
						},
						Data: map[string][]byte{},
					},
				},
			},
			errResult: "secret testing/test-secret is missing the following TLS data entries: tls.crt, tls.key",
		},
		{
			desc: "add certificates to the configuration",
			ingress: buildIngress(
				iNamespace("testing"),
				iRules(
					iRule(iHost("ep1.example.com")),
					iRule(iHost("ep2.example.com")),
					iRule(iHost("ep3.example.com")),
				),
				iTLSes(
					iTLS("test-secret"),
					iTLS("test-secret2"),
				),
			),
			client: clientMock{
				secrets: []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-secret2",
							Namespace: "testing",
						},
						Data: map[string][]byte{
							"tls.crt": []byte("tls-crt"),
							"tls.key": []byte("tls-key"),
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "test-secret",
							Namespace: "testing",
						},
						Data: map[string][]byte{
							"tls.crt": []byte("tls-crt"),
							"tls.key": []byte("tls-key"),
						},
					},
				},
			},
			result: map[string]*tls.CertAndStores{
				"testing-test-secret": {
					Certificate: tls.Certificate{
						CertFile: types.FileOrContent("tls-crt"),
						KeyFile:  types.FileOrContent("tls-key"),
					},
				},
				"testing-test-secret2": {
					Certificate: tls.Certificate{
						CertFile: types.FileOrContent("tls-crt"),
						KeyFile:  types.FileOrContent("tls-key"),
					},
				},
			},
		},
		{
			desc:    "return nil when no secret is defined",
			ingress: testIngressWithoutSecret,
			client:  clientMock{},
			result:  map[string]*tls.CertAndStores{},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			tlsConfigs := map[string]*tls.CertAndStores{}
			err := getCertificates(t.Context(), test.ingress, test.client, tlsConfigs)

			if test.errResult != "" {
				assert.EqualError(t, err, test.errResult)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.result, tlsConfigs)
			}
		})
	}
}

func TestLoadConfigurationFromIngressesWithNativeLBByDefault(t *testing.T) {
	testCases := []struct {
		desc         string
		ingressClass string
		expected     *dynamic.Configuration
	}{
		{
			desc: "Ingress with native service lb",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"testing-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "testing-service1-8080",
						},
					},
					Services: map[string]*dynamic.Service{
						"testing-service1-8080": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								PassHostHeader:     pointer(true),
								Servers: []dynamic.Server{
									{
										URL: "http://10.0.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with native lb by default",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"default-global-native-lb-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "default-service1-8080",
						},
					},
					Services: map[string]*dynamic.Service{
						"default-service1-8080": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								PassHostHeader:     pointer(true),
								Servers: []dynamic.Server{
									{
										URL: "http://10.0.0.1:8080",
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "Ingress with native lb by default but service has disabled nativelb",
			expected: &dynamic.Configuration{
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Routers: map[string]*dynamic.Router{
						"default-global-native-lb-traefik-tchouk-bar": {
							Rule:    "Host(`traefik.tchouk`) && PathPrefix(`/bar`)",
							Service: "default-native-disabled-svc-web",
						},
					},
					Services: map[string]*dynamic.Service{
						"default-native-disabled-svc-web": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:           dynamic.BalancerStrategyWRR,
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								PassHostHeader:     pointer(true),
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.20:8080",
									},
									{
										URL: "http://10.10.0.21:8080",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			clientMock := newClientMock(generateTestFilename(test.desc))

			p := Provider{
				IngressClass:      test.ingressClass,
				NativeLBByDefault: true,
			}
			conf := p.loadConfigurationFromIngresses(t.Context(), clientMock)

			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestIngressEndpointPublishedService(t *testing.T) {
	testCases := []struct {
		desc                         string
		disableClusterScopeResources bool
		expected                     []netv1.IngressLoadBalancerIngress
	}{
		{
			desc: "Published Service ClusterIP",
			expected: []netv1.IngressLoadBalancerIngress{
				{
					IP: "1.2.3.4",
					Ports: []netv1.IngressPortStatus{
						{Port: 9090, Protocol: "TCP"},
						{Port: 9091, Protocol: "TCP"},
					},
				},
				{
					IP: "5.6.7.8",
					Ports: []netv1.IngressPortStatus{
						{Port: 9090, Protocol: "TCP"},
						{Port: 9091, Protocol: "TCP"},
					},
				},
			},
		},
		{
			desc: "Published Service LoadBalancer",
			expected: []netv1.IngressLoadBalancerIngress{
				{
					IP: "1.2.3.4",
					Ports: []netv1.IngressPortStatus{
						{Port: 9090, Protocol: "TCP"},
						{Port: 9091, Protocol: "TCP"},
					},
				},
				{
					IP: "5.6.7.8",
					Ports: []netv1.IngressPortStatus{
						{Port: 9090, Protocol: "TCP"},
						{Port: 9091, Protocol: "TCP"},
					},
				},
			},
		},
		{
			desc:                         "Published Service NodePort",
			disableClusterScopeResources: true,
		},
		{
			desc: "Published Service NodePort",
			expected: []netv1.IngressLoadBalancerIngress{
				{
					IP: "1.2.3.4",
					Ports: []netv1.IngressPortStatus{
						{Port: 9090, Protocol: "TCP"},
						{Port: 9091, Protocol: "TCP"},
					},
				},
				{
					IP: "5.6.7.8",
					Ports: []netv1.IngressPortStatus{
						{Port: 9090, Protocol: "TCP"},
						{Port: 9091, Protocol: "TCP"},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			k8sObjects := readResources(t, []string{generateTestFilename(test.desc)})
			kubeClient := kubefake.NewClientset(k8sObjects...)

			client := newClientImpl(kubeClient)

			stopCh := make(chan struct{})
			eventCh, err := client.WatchAll(nil, stopCh)
			require.NoError(t, err)

			if k8sObjects != nil {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				DisableClusterScopeResources: test.disableClusterScopeResources,
				IngressEndpoint: &EndpointIngress{
					PublishedService: "default/published-service",
				},
			}
			p.loadConfigurationFromIngresses(t.Context(), client)

			ingress, err := kubeClient.NetworkingV1().Ingresses(metav1.NamespaceDefault).Get(t.Context(), "foo", metav1.GetOptions{})
			require.NoError(t, err)

			assert.Equal(t, test.expected, ingress.Status.LoadBalancer.Ingress)
		})
	}
}

func readResources(t *testing.T, paths []string) []runtime.Object {
	t.Helper()

	var k8sObjects []runtime.Object
	for _, path := range paths {
		yamlContent, err := os.ReadFile(filepath.FromSlash(path))
		if err != nil {
			panic(err)
		}

		objects := k8s.MustParseYaml(yamlContent)
		k8sObjects = append(k8sObjects, objects...)
	}

	return k8sObjects
}

func TestStrictPrefixMatchingRule(t *testing.T) {
	tests := []struct {
		path        string
		requestPath string
		match       bool
	}{ // The tests are taken from https://kubernetes.io/docs/concepts/services-networking/ingress/#examples
		{
			path:        "/foo",
			requestPath: "/foo",
			match:       true,
		},
		{
			path:        "/foo",
			requestPath: "/foo/",
			match:       true,
		},
		{
			path:        "/foo/",
			requestPath: "/foo",
			match:       true,
		},
		{
			path:        "/foo/",
			requestPath: "/foo/",
			match:       true,
		},
		{
			path:        "/aaa/bb",
			requestPath: "/aaa/bbb",
			match:       false,
		},
		{
			path:        "/aaa/bbb",
			requestPath: "/aaa/bbb",
			match:       true,
		},
		{
			path:        "/aaa/bbb/",
			requestPath: "/aaa/bbb",
			match:       true,
		},
		{
			path:        "/aaa/bbb",
			requestPath: "/aaa/bbb/",
			match:       true,
		},
		{
			path:        "/aaa/bbb",
			requestPath: "/aaa/bbb/ccc",
			match:       true,
		},
		{
			path:        "/aaa/bbb",
			requestPath: "/aaa/bbbxyz",
			match:       false,
		},
		{
			path:        "/",
			requestPath: "/aaa/ccc",
			match:       true,
		},
		{
			path:        "/aaa",
			requestPath: "/aaa/ccc",
			match:       true,
		},
		{
			path:        "/...",
			requestPath: "/aaa",
			match:       false,
		},
		{
			path:        "/...",
			requestPath: "/.../",
			match:       true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Prefix match case %s", tt.path), func(t *testing.T) {
			t.Parallel()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			parser, err := traefikhttp.NewSyntaxParser()
			require.NoError(t, err)

			muxer := traefikhttp.NewMuxer(parser)

			rule := buildStrictPrefixMatchingRule(tt.path)
			err = muxer.AddRoute(rule, "", 0, handler)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, http.NoBody)
			muxer.ServeHTTP(w, req)

			if tt.match {
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				assert.Equal(t, http.StatusNotFound, w.Code)
			}
		})
	}
}
