package ingressnginx

import (
	"errors"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-acme/lego/v5/challenge/tlsalpn01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
)

func TestLoadIngresses(t *testing.T) {
	testCases := []struct {
		desc                           string
		ingressClass                   string
		defaultBackendServiceName      string
		defaultBackendServiceNamespace string
		allowCrossNamespaceResources   bool
		globalAllowedResponseHeaders   []string
		ipAllowListStrategy            *dynamic.IPStrategy
		allowSnippetAnnotations        bool
		globalAuthURL                  string
		strictValidatePathType         *bool
		proxyRequestBuffering          bool
		disableNonTLSRouters           bool
		paths                          []string
		expected                       *dynamic.Configuration
	}{
		{
			desc: "Empty, no IngressClass",
			paths: []string{
				"services.yml",
				"ingresses/ingress-with-basicauth.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
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
			desc: "Service unavailable in HTTP",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-service-unavailable-http.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-service-unavailable-http-37-unavailable-11-80-2",
							Middlewares: []string{"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-service-unavailable-http",
									},
								},
							},
						},
						"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-service-unavailable-http-37-unavailable-11-80-2",
							Middlewares: []string{"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-tls-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-service-unavailable-http",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-service-unavailable-http-37-unavailable-11-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "unavailable",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-service-unavailable-http-37": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Service unavailable in HTTP with redirect annotation",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-service-unavailable-http-annotation.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-service-unavailable-http-37-unavailable-11-80-2",
							Middlewares: []string{"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-redirect", "default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-service-unavailable-http",
									},
								},
							},
						},
						"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-service-unavailable-http-37-unavailable-11-80-2",
							Middlewares: []string{"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-tls-redirect", "default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-tls-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-service-unavailable-http",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://foo.bar.com",
								StatusCode:  new(301),
							},
						},
						"default-7-ingress-with-service-unavailable-http-37-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://foo.bar.com",
								StatusCode:  new(301),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-service-unavailable-http-37-unavailable-11-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "unavailable",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-service-unavailable-http-37": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Service unavailable in default backend",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-service-unavailable-default-backend.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-service-unavailable-default-backend-48-rule-0-default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-service-unavailable-default-backend-48-default-backend",
							Middlewares: []string{"default-7-ingress-with-service-unavailable-default-backend-48-rule-0-default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-service-unavailable-default-backend",
									},
								},
							},
						},
						"default-7-ingress-with-service-unavailable-default-backend-48-rule-0-default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-service-unavailable-default-backend-48-default-backend",
							Middlewares: []string{"default-7-ingress-with-service-unavailable-default-backend-48-rule-0-default-backend-tls-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-service-unavailable-default-backend",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-service-unavailable-default-backend-48-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && (Path("/foo") || PathPrefix("/foo/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-service-unavailable-default-backend-48-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-service-unavailable-default-backend-48-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-service-unavailable-default-backend",
									},
								},
							},
						},
						"default-7-ingress-with-service-unavailable-default-backend-48-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && (Path("/foo") || PathPrefix("/foo/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-service-unavailable-default-backend-48-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-service-unavailable-default-backend-48-rule-0-path-0-tls-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-service-unavailable-default-backend",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-service-unavailable-default-backend-48-rule-0-default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-service-unavailable-default-backend-48-rule-0-default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-service-unavailable-default-backend-48-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-service-unavailable-default-backend-48-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-service-unavailable-default-backend-48-default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "unavailable",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-service-unavailable-default-backend-48-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-service-unavailable-default-backend-48",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-service-unavailable-default-backend-48": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                         "Custom Headers",
			allowCrossNamespaceResources: true,
			globalAllowedResponseHeaders: []string{"X-Custom-Header", "X-Cross-Header"},
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-custom-headers.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-custom-headers-27-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-custom-headers-27-rule-0-path-0-custom-headers", "default-7-ingress-with-custom-headers-27-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-custom-headers-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-headers",
									},
								},
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("cross-namespace.localhost") && Path("/")`,
							Middlewares: []string{"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-custom-headers", "default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-retry"},
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-cross-namespace-headers-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-cross-namespace-headers",
									},
								},
							},
						},
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-custom-headers-27-rule-0-path-0-tls-custom-headers", "default-7-ingress-with-custom-headers-27-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-custom-headers-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-headers",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("cross-namespace.localhost") && Path("/")`,
							Middlewares: []string{"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-tls-custom-headers", "default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-tls-retry"},
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-cross-namespace-headers-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-cross-namespace-headers",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Cross-Header": "cross-value"},
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-tls-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-tls-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Cross-Header": "cross-value"},
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-custom-headers-27-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-custom-headers-27",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-cross-namespace-headers-36",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-custom-headers-27": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                         "Custom Headers with invalid header value returns 503",
			globalAllowedResponseHeaders: []string{"X-Custom-Header"},
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-invalid-header-value.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-invalid-header-value-33-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     unavailableServiceName,
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-invalid-header-value",
									},
								},
							},
						},
						"default-7-ingress-with-invalid-header-value-33-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     unavailableServiceName,
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-invalid-header-value",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-invalid-header-value-33-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-invalid-header-value-33",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-invalid-header-value-33": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                         "Custom Headers with cross namespace not allowed",
			globalAllowedResponseHeaders: []string{"X-Custom-Header", "X-Cross-Header"},
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-custom-headers.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-custom-headers-27-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-custom-headers-27-rule-0-path-0-custom-headers", "default-7-ingress-with-custom-headers-27-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-custom-headers-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-headers",
									},
								},
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("cross-namespace.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "unavailable-service",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-cross-namespace-headers",
									},
								},
							},
						},
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-custom-headers-27-rule-0-path-0-tls-custom-headers", "default-7-ingress-with-custom-headers-27-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-custom-headers-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-headers",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("cross-namespace.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "unavailable-service",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-cross-namespace-headers",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-tls-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-custom-headers-27-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-custom-headers-27",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-cross-namespace-headers-36",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-custom-headers-27": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                         "Custom Headers cross namespace with cross namespace allowed",
			globalAllowedResponseHeaders: []string{"X-Custom-Header", "X-Cross-Header"},
			allowCrossNamespaceResources: true,
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-custom-headers.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-custom-headers-27-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-custom-headers-27-rule-0-path-0-custom-headers", "default-7-ingress-with-custom-headers-27-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-custom-headers-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-headers",
									},
								},
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("cross-namespace.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-custom-headers", "default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-cross-namespace-headers-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-cross-namespace-headers",
									},
								},
							},
						},
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-custom-headers-27-rule-0-path-0-tls-custom-headers", "default-7-ingress-with-custom-headers-27-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-custom-headers-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-headers",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("cross-namespace.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-tls-custom-headers", "default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-cross-namespace-headers-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-cross-namespace-headers",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-tls-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-custom-headers-27-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Cross-Header": "cross-value"},
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-tls-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Cross-Header": "cross-value"},
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-custom-headers-27-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-custom-headers-27",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-cross-namespace-headers-36",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-custom-headers-27": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-cross-namespace-headers-36": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "No annotation",
			paths: []string{
				"ingresses/ingress-with-no-annotation.yml",
				"ingressclasses.yml",
				"services.yml",
				"secrets.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-no-annotation-26-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-7-ingress-with-no-annotation-26-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-no-annotation-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-no-annotation",
									},
								},
							},
						},
						"default-7-ingress-with-no-annotation-26-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-ingress-with-no-annotation-26-rule-0-path-0-redirect-scheme",
								"default-7-ingress-with-no-annotation-26-rule-0-path-0-retry",
							},
							Service: "default-7-ingress-with-no-annotation-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-no-annotation",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-no-annotation-26-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-no-annotation-26-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-no-annotation-26-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-no-annotation-26-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-no-annotation-26",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-no-annotation-26": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
				},
			},
		},
		{
			// Regression test for a generated name collision: an ingress name may contain dots, and
			// normalizing it turns my.ingress into my-ingress, so the routers, services and
			// serversTransports of the two Ingresses would silently overwrite one another.
			desc: "Ingress name collision",
			paths: []string{
				"ingresses/ingress-with-name-collision.yml",
				"ingressclasses.yml",
				"services.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-my-ingress-10-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/dash")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-7-my-ingress-10-rule-0-path-0-tls-retry"},
							Service:     "default-7-my-ingress-10-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "my-ingress",
									},
								},
							},
						},
						"default-7-my-ingress-10-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/dash")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-my-ingress-10-rule-0-path-0-retry"},
							Service:     "default-7-my-ingress-10-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "my-ingress",
									},
								},
							},
						},
						"default-7-my.ingress-10-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/dot")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-7-my.ingress-10-rule-0-path-0-tls-retry"},
							Service:     "default-7-my.ingress-10-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "my.ingress",
									},
								},
							},
						},
						"default-7-my.ingress-10-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/dot")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-my.ingress-10-rule-0-path-0-retry"},
							Service:     "default-7-my.ingress-10-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "my.ingress",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-my-ingress-10-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-my-ingress-10-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-my.ingress-10-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-my.ingress-10-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-my-ingress-10-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-my-ingress-10",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-my.ingress-10-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-my.ingress-10",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-my-ingress-10": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-my.ingress-10": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			// Regression test for a generated name collision across namespaces: namespaces may contain
			// dashes, so team/dev-web and team-dev/web would both produce team-dev-web, and the routers,
			// services and serversTransports of one Ingress would silently overwrite the other's.
			desc: "Ingress name collision across namespaces",
			paths: []string{
				"ingresses/ingress-with-namespace-collision.yml",
				"ingressclasses.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"team-4-dev-web-7-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/dash")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"team-4-dev-web-7-rule-0-path-0-tls-retry"},
							Service:     "team-4-dev-web-7-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "team",
										Name:      "dev-web",
									},
								},
							},
						},
						"team-4-dev-web-7-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/dash")`,
							RuleSyntax:  "default",
							Middlewares: []string{"team-4-dev-web-7-rule-0-path-0-retry"},
							Service:     "team-4-dev-web-7-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "team",
										Name:      "dev-web",
									},
								},
							},
						},
						"team-dev-8-web-3-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/dot")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"team-dev-8-web-3-rule-0-path-0-tls-retry"},
							Service:     "team-dev-8-web-3-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "team-dev",
										Name:      "web",
									},
								},
							},
						},
						"team-dev-8-web-3-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/dot")`,
							RuleSyntax:  "default",
							Middlewares: []string{"team-dev-8-web-3-rule-0-path-0-retry"},
							Service:     "team-dev-8-web-3-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "team-dev",
										Name:      "web",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"team-4-dev-web-7-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"team-4-dev-web-7-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"team-dev-8-web-3-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"team-dev-8-web-3-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"team-4-dev-web-7-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "team",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"team-dev-8-web-3-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "team-dev",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"team-4-dev-web-7": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"team-dev-8-web-3": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Basic Auth",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-basicauth.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-basicauth-22-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/basicauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-basicauth-22-rule-0-path-0-basic-auth", "default-7-ingress-with-basicauth-22-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-basicauth-22-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-basicauth",
									},
								},
							},
						},
						"default-7-ingress-with-basicauth-22-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/basicauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-basicauth-22-rule-0-path-0-tls-basic-auth", "default-7-ingress-with-basicauth-22-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-basicauth-22-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-basicauth",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-basicauth-22-rule-0-path-0-basic-auth": {
							BasicAuth: &dynamic.BasicAuth{
								Users: dynamic.Users{
									"urn:k8s:secret:default/basic-auth:auth",
								},
								Realm: "Authentication Required",
							},
							URNRefs: map[string]string{
								"/basicAuth/users/0": "user:{SHA}W6ph5Mm5Pz8GgiULbPgzG37mj9g=",
							},
						},
						"default-7-ingress-with-basicauth-22-rule-0-path-0-tls-basic-auth": {
							BasicAuth: &dynamic.BasicAuth{
								Users: dynamic.Users{
									"urn:k8s:secret:default/basic-auth:auth",
								},
								Realm: "Authentication Required",
							},
							URNRefs: map[string]string{
								"/basicAuth/users/0": "user:{SHA}W6ph5Mm5Pz8GgiULbPgzG37mj9g=",
							},
						},
						"default-7-ingress-with-basicauth-22-whoami-6-80-2-remove-authorization-header": {
							Headers: &dynamic.Headers{
								CustomRequestHeaders: map[string]string{"Authorization": ""},
							},
						},
						"default-7-ingress-with-basicauth-22-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-basicauth-22-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-basicauth-22-whoami-6-80-2": {
							Middlewares: []string{"default-7-ingress-with-basicauth-22-whoami-6-80-2-remove-authorization-header"},
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-basicauth-22",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-basicauth-22": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Digest Auth",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-digestauth.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-digestauth-23-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/digestauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-digestauth-23-rule-0-path-0-digest-auth", "default-7-ingress-with-digestauth-23-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-digestauth-23-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-digestauth",
									},
								},
							},
						},
						"default-7-ingress-with-digestauth-23-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/digestauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-digestauth-23-rule-0-path-0-tls-digest-auth", "default-7-ingress-with-digestauth-23-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-digestauth-23-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-digestauth",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-digestauth-23-rule-0-path-0-digest-auth": {
							DigestAuth: &dynamic.DigestAuth{
								Users: dynamic.Users{
									"urn:k8s:secret:default/digest-auth:auth",
								},
								Realm: "Authentication Required",
							},
							URNRefs: map[string]string{
								"/digestAuth/users/0": "user:Authentication Required:0a132d8fa4138c56c1c2b50e9c9d1a1f",
							},
						},
						"default-7-ingress-with-digestauth-23-rule-0-path-0-tls-digest-auth": {
							DigestAuth: &dynamic.DigestAuth{
								Users: dynamic.Users{
									"urn:k8s:secret:default/digest-auth:auth",
								},
								Realm: "Authentication Required",
							},
							URNRefs: map[string]string{
								"/digestAuth/users/0": "user:Authentication Required:0a132d8fa4138c56c1c2b50e9c9d1a1f",
							},
						},
						"default-7-ingress-with-digestauth-23-whoami-6-80-2-remove-authorization-header": {
							Headers: &dynamic.Headers{
								CustomRequestHeaders: map[string]string{"Authorization": ""},
							},
						},
						"default-7-ingress-with-digestauth-23-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-digestauth-23-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-digestauth-23-whoami-6-80-2": {
							Middlewares: []string{"default-7-ingress-with-digestauth-23-whoami-6-80-2-remove-authorization-header"},
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-digestauth-23",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-digestauth-23": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Basic Auth with missing secret — ingress is skipped entirely",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-basicauth-secret-missing.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
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
			desc: "Forward Auth",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-forwardauth.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-forwardauth-24-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/forwardauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-forwardauth-24-rule-0-path-0-snippet", "default-7-ingress-with-forwardauth-24-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-forwardauth-24-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-forwardauth",
									},
								},
							},
						},
						"default-7-ingress-with-forwardauth-24-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/forwardauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-forwardauth-24-rule-0-path-0-tls-snippet", "default-7-ingress-with-forwardauth-24-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-forwardauth-24-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-forwardauth",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-forwardauth-24-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address:             "http://whoami.default.svc/",
									Method:              http.MethodGet,
									AuthResponseHeaders: []string{"X-Foo", "X-Bar"},
									AuthSigninURL:       "https://auth.example.com/oauth2/start?rd=foo",
								},
							},
						},
						"default-7-ingress-with-forwardauth-24-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address:             "http://whoami.default.svc/",
									Method:              http.MethodGet,
									AuthResponseHeaders: []string{"X-Foo", "X-Bar"},
									AuthSigninURL:       "https://auth.example.com/oauth2/start?rd=foo",
								},
							},
						},
						"default-7-ingress-with-forwardauth-24-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-forwardauth-24-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-forwardauth-24-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-forwardauth-24",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-forwardauth-24": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Forward Auth with auth-snippet - allowSnippetAnnotations disabled",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-forwardauth-snippet.yml",
			},
			allowSnippetAnnotations: false,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
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
			desc: "Forward Auth with auth-snippet - allowSnippetAnnotations enabled",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-forwardauth-snippet.yml",
			},
			allowSnippetAnnotations: true,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-forwardauth-snippet-32-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/forwardauth-snippet")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-forwardauth-snippet-32-rule-0-path-0-snippet", "default-7-ingress-with-forwardauth-snippet-32-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-forwardauth-snippet-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-forwardauth-snippet",
									},
								},
							},
						},
						"default-7-ingress-with-forwardauth-snippet-32-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/forwardauth-snippet")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-forwardauth-snippet-32-rule-0-path-0-tls-snippet", "default-7-ingress-with-forwardauth-snippet-32-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-forwardauth-snippet-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-forwardauth-snippet",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-forwardauth-snippet-32-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address:             "http://whoami.default.svc/",
									AuthResponseHeaders: []string{"X-Auth-Snippet"},
									AuthSigninURL:       "https://auth.example.com/oauth2/start?rd=foo",
									Method:              http.MethodPost,
									Snippet:             "add_header X-Auth-Snippet \"auth-value\";\n",
								},
							},
						},
						"default-7-ingress-with-forwardauth-snippet-32-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address:             "http://whoami.default.svc/",
									AuthResponseHeaders: []string{"X-Auth-Snippet"},
									AuthSigninURL:       "https://auth.example.com/oauth2/start?rd=foo",
									Method:              http.MethodPost,
									Snippet:             "add_header X-Auth-Snippet \"auth-value\";\n",
								},
							},
						},
						"default-7-ingress-with-forwardauth-snippet-32-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-forwardauth-snippet-32-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-forwardauth-snippet-32-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-forwardauth-snippet-32",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-forwardauth-snippet-32": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Global Auth applied when no local auth-url",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-without-auth.yml",
			},
			globalAuthURL: "http://auth.example.com/verify",
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-without-auth-20-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-without-auth-20-rule-0-path-0-snippet", "default-7-ingress-without-auth-20-rule-0-path-0-retry"},
							Service:     "default-7-ingress-without-auth-20-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-without-auth",
									},
								},
							},
						},
						"default-7-ingress-without-auth-20-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-without-auth-20-rule-0-path-0-tls-snippet", "default-7-ingress-without-auth-20-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-without-auth-20-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-without-auth",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-without-auth-20-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address: "http://auth.example.com/verify",
								},
							},
						},
						"default-7-ingress-without-auth-20-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address: "http://auth.example.com/verify",
								},
							},
						},
						"default-7-ingress-without-auth-20-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-without-auth-20-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-without-auth-20-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-without-auth-20",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-without-auth-20": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Global Auth disabled by annotation",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-global-auth-disabled.yml",
			},
			globalAuthURL: "http://auth.example.com/verify",
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-global-auth-disabled-33-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-global-auth-disabled-33-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-global-auth-disabled-33-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-global-auth-disabled",
									},
								},
							},
						},
						"default-7-ingress-with-global-auth-disabled-33-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-global-auth-disabled-33-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-global-auth-disabled-33-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-global-auth-disabled",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-global-auth-disabled-33-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-global-auth-disabled-33-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-global-auth-disabled-33-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-global-auth-disabled-33",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-global-auth-disabled-33": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Local auth-url takes precedence over global auth",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-forwardauth.yml",
			},
			globalAuthURL: "http://global-auth.example.com/verify",
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-forwardauth-24-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/forwardauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-forwardauth-24-rule-0-path-0-snippet", "default-7-ingress-with-forwardauth-24-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-forwardauth-24-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-forwardauth",
									},
								},
							},
						},
						"default-7-ingress-with-forwardauth-24-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/forwardauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-forwardauth-24-rule-0-path-0-tls-snippet", "default-7-ingress-with-forwardauth-24-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-forwardauth-24-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-forwardauth",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-forwardauth-24-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address:             "http://whoami.default.svc/",
									Method:              http.MethodGet,
									AuthResponseHeaders: []string{"X-Foo", "X-Bar"},
									AuthSigninURL:       "https://auth.example.com/oauth2/start?rd=foo",
								},
							},
						},
						"default-7-ingress-with-forwardauth-24-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address:             "http://whoami.default.svc/",
									Method:              http.MethodGet,
									AuthResponseHeaders: []string{"X-Foo", "X-Bar"},
									AuthSigninURL:       "https://auth.example.com/oauth2/start?rd=foo",
								},
							},
						},
						"default-7-ingress-with-forwardauth-24-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-forwardauth-24-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-forwardauth-24-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-forwardauth-24",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-forwardauth-24": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "No global auth when GlobalAuthURL not configured",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-without-auth.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-without-auth-20-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-without-auth-20-rule-0-path-0-retry"},
							Service:     "default-7-ingress-without-auth-20-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-without-auth",
									},
								},
							},
						},
						"default-7-ingress-without-auth-20-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-without-auth-20-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-without-auth-20-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-without-auth",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-without-auth-20-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-without-auth-20-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-without-auth-20-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-without-auth-20",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-without-auth-20": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "SSL Redirect",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-ssl-redirect.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-ssl-redirect-25-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("sslredirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-7-ingress-with-ssl-redirect-25-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-ssl-redirect-25-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-ssl-redirect",
									},
								},
							},
						},
						"default-7-ingress-with-ssl-redirect-25-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("sslredirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-ingress-with-ssl-redirect-25-rule-0-path-0-redirect-scheme",
								"default-7-ingress-with-ssl-redirect-25-rule-0-path-0-retry",
							},
							Service: "default-7-ingress-with-ssl-redirect-25-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-ssl-redirect",
									},
								},
							},
						},
						"default-7-ingress-without-ssl-redirect-28-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("withoutsslredirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-without-ssl-redirect-28-rule-0-path-0-retry"},
							Service:     "default-7-ingress-without-ssl-redirect-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-without-ssl-redirect",
									},
								},
							},
						},
						"default-7-ingress-without-ssl-redirect-28-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("withoutsslredirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-7-ingress-without-ssl-redirect-28-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-without-ssl-redirect-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-without-ssl-redirect",
									},
								},
							},
						},
						"default-7-ingress-with-force-ssl-redirect-31-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("forcesslredirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-ingress-with-force-ssl-redirect-31-rule-0-path-0-redirect-scheme",
								"default-7-ingress-with-force-ssl-redirect-31-rule-0-path-0-retry",
							},
							Service: "default-7-ingress-with-force-ssl-redirect-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-force-ssl-redirect",
									},
								},
							},
						},
						"default-7-ingress-with-force-ssl-redirect-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("forcesslredirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-force-ssl-redirect-31-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-force-ssl-redirect-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-force-ssl-redirect",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-ssl-redirect-25-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-ssl-redirect-25-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-ssl-redirect-25-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-without-ssl-redirect-28-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-without-ssl-redirect-28-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-force-ssl-redirect-31-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-force-ssl-redirect-31-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-force-ssl-redirect-31-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-ssl-redirect-25-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-ssl-redirect-25",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-without-ssl-redirect-28-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-without-ssl-redirect-28",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-force-ssl-redirect-31-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-force-ssl-redirect-31",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-ssl-redirect-25": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-without-ssl-redirect-28": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-force-ssl-redirect-31": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
				},
			},
		},
		{
			// Regression test: force-ssl-redirect with no TLS section on the Ingress (LB-terminated TLS).
			// The HTTP router must use the real backend service so that requests arriving with
			// X-Forwarded-Proto: https (redirect bypassed) are forwarded to the backend instead
			// of hitting noop@internal and receiving HTTP 418.
			desc: "Force SSL Redirect without Ingress TLS section",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-force-ssl-redirect-no-tls.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-force-ssl-redirect-no-tls-38-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("forcesslredirectnotls.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-force-ssl-redirect-no-tls-38-whoami-6-80-2",
							Middlewares: []string{
								"default-7-ingress-with-force-ssl-redirect-no-tls-38-rule-0-path-0-redirect-scheme",
								"default-7-ingress-with-force-ssl-redirect-no-tls-38-rule-0-path-0-retry",
							},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-force-ssl-redirect-no-tls",
									},
								},
							},
						},
						"default-7-ingress-with-force-ssl-redirect-no-tls-38-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("forcesslredirectnotls.localhost") && Path("/")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Service:     "default-7-ingress-with-force-ssl-redirect-no-tls-38-whoami-6-80-2",
							Middlewares: []string{
								"default-7-ingress-with-force-ssl-redirect-no-tls-38-rule-0-path-0-tls-retry",
							},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-force-ssl-redirect-no-tls",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-force-ssl-redirect-no-tls-38-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-force-ssl-redirect-no-tls-38-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-force-ssl-redirect-no-tls-38-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-force-ssl-redirect-no-tls-38-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-force-ssl-redirect-no-tls-38",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-force-ssl-redirect-no-tls-38": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "SSL Passthrough",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-ssl-passthrough.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-7-ingress-with-ssl-passthrough-28-passthrough.whoami.localhost-28": {
							EntryPoints: []string{"https"},
							Rule:        `HostSNI("passthrough.whoami.localhost")`,
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
							Service: "default-7-whoami-tls-10-443-3",
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-7-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.3:8443",
									},
									{
										Address: "10.10.0.4:8443",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-ssl-passthrough-28-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("passthrough.whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-ssl-passthrough-28-whoami-tls-10-443-3",
							Middlewares: []string{"default-7-ingress-with-ssl-passthrough-28-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-ssl-passthrough",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-ssl-passthrough-28-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-ssl-passthrough-28-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: "wrr",
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.3:8443"},
									{URL: "http://10.10.0.4:8443"},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-ssl-passthrough-28",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-tls",
										Port:      "443",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-ssl-passthrough-28": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "SSL Passthrough with hosts differing only by dots and dashes",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-ssl-passthrough-similar-hosts.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-7-ingress-with-ssl-passthrough-similar-hosts-42-passthrough.whoami.localhost-28": {
							EntryPoints: []string{"https"},
							Rule:        `HostSNI("passthrough.whoami.localhost")`,
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
							Service: "default-7-whoami-tls-10-443-3",
						},
						"default-7-ingress-with-ssl-passthrough-similar-hosts-42-passthrough-whoami.localhost-28": {
							EntryPoints: []string{"https"},
							Rule:        `HostSNI("passthrough-whoami.localhost")`,
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
							Service: "default-7-whoami-tls-10-443-3",
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-7-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.3:8443",
									},
									{
										Address: "10.10.0.4:8443",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-ssl-passthrough-similar-hosts-42-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("passthrough.whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-ssl-passthrough-similar-hosts-42-whoami-tls-10-443-3",
							Middlewares: []string{"default-7-ingress-with-ssl-passthrough-similar-hosts-42-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-ssl-passthrough-similar-hosts",
									},
								},
							},
						},
						"default-7-ingress-with-ssl-passthrough-similar-hosts-42-rule-1-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("passthrough-whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-ssl-passthrough-similar-hosts-42-whoami-tls-10-443-3",
							Middlewares: []string{"default-7-ingress-with-ssl-passthrough-similar-hosts-42-rule-1-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-ssl-passthrough-similar-hosts",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-ssl-passthrough-similar-hosts-42-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-ssl-passthrough-similar-hosts-42-rule-1-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-ssl-passthrough-similar-hosts-42-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.3:8443"},
									{URL: "http://10.10.0.4:8443"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-ssl-passthrough-similar-hosts-42",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-tls",
										Port:      "443",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-ssl-passthrough-similar-hosts-42": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "SSL Passthrough with a default backend creates no TLS router for it",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-ssl-passthrough-and-default-backend.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-7-ingress-with-ssl-passthrough-and-default-backend-48-passthrough.whoami.localhost-28": {
							EntryPoints: []string{"https"},
							Rule:        `HostSNI("passthrough.whoami.localhost")`,
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
							Service: "default-7-whoami-tls-10-443-3",
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-7-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.3:8443",
									},
									{
										Address: "10.10.0.4:8443",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-ssl-passthrough-and-default-backend-48-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("passthrough.whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-ssl-passthrough-and-default-backend-48-whoami-tls-10-443-3",
							Middlewares: []string{"default-7-ingress-with-ssl-passthrough-and-default-backend-48-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-ssl-passthrough-and-default-backend",
									},
								},
							},
						},
						"default-7-ingress-with-ssl-passthrough-and-default-backend-48-rule-0-default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `Host("passthrough.whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-ssl-passthrough-and-default-backend-48-default-backend",
							Middlewares: []string{"default-7-ingress-with-ssl-passthrough-and-default-backend-48-rule-0-default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-ssl-passthrough-and-default-backend",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-ssl-passthrough-and-default-backend-48-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-ssl-passthrough-and-default-backend-48-rule-0-default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-ssl-passthrough-and-default-backend-48-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: "wrr",
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.3:8443"},
									{URL: "http://10.10.0.4:8443"},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-ssl-passthrough-and-default-backend-48",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-tls",
										Port:      "443",
									},
								},
							},
						},
						"default-7-ingress-with-ssl-passthrough-and-default-backend-48-default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: "wrr",
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-ssl-passthrough-and-default-backend-48",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-ssl-passthrough-and-default-backend-48": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "SSL Passthrough applies the access control annotations to the HTTP route",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-ssl-passthrough-and-auth.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-7-ingress-with-ssl-passthrough-and-auth-37-passthrough-auth.whoami.localhost-33": {
							EntryPoints: []string{"https"},
							Rule:        `HostSNI("passthrough-auth.whoami.localhost")`,
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
							Service: "default-7-whoami-tls-10-443-3",
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-7-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{Address: "10.10.0.3:8443"},
									{Address: "10.10.0.4:8443"},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-ssl-passthrough-and-auth-37-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("passthrough-auth.whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-ssl-passthrough-and-auth-37-whoami-tls-10-443-3",
							Middlewares: []string{
								"default-7-ingress-with-ssl-passthrough-and-auth-37-rule-0-path-0-basic-auth",
								"default-7-ingress-with-ssl-passthrough-and-auth-37-rule-0-path-0-allowed-source-range",
								"default-7-ingress-with-ssl-passthrough-and-auth-37-rule-0-path-0-retry",
							},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-ssl-passthrough-and-auth",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-ssl-passthrough-and-auth-37-rule-0-path-0-basic-auth": {
							BasicAuth: &dynamic.BasicAuth{
								Users: dynamic.Users{
									"urn:k8s:secret:default/basic-auth:auth",
								},
								Realm: "Authentication Required",
							},
							URNRefs: map[string]string{
								"/basicAuth/users/0": "user:{SHA}W6ph5Mm5Pz8GgiULbPgzG37mj9g=",
							},
						},
						"default-7-ingress-with-ssl-passthrough-and-auth-37-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"10.0.0.0/24"},
							},
						},
						"default-7-ingress-with-ssl-passthrough-and-auth-37-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-ssl-passthrough-and-auth-37-whoami-tls-10-443-3-remove-authorization-header": {
							Headers: &dynamic.Headers{
								CustomRequestHeaders: map[string]string{"Authorization": ""},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-ssl-passthrough-and-auth-37-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: "wrr",
								Servers: []dynamic.Server{
									{URL: "https://10.10.0.3:8443"},
									{URL: "https://10.10.0.4:8443"},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-ssl-passthrough-and-auth-37",
							},
							Middlewares: []string{"default-7-ingress-with-ssl-passthrough-and-auth-37-whoami-tls-10-443-3-remove-authorization-header"},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-tls",
										Port:      "443",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-ssl-passthrough-and-auth-37": {
							InsecureSkipVerify: true,
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			// The TLS certificates of an ingress are loaded before anything can skip it,
			// so the default certificate pool stays consistent even when the serversTransport cannot be built.
			desc: "TLS section certificates are loaded when the serversTransport cannot be built",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-tls-section-and-invalid-proxy-ssl-secret.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
				},
			},
		},
		{
			desc: "SSL Passthrough keeps the TCP router when the auth secret cannot be resolved",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-ssl-passthrough-and-missing-auth-secret.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-7-ingress-with-ssl-passthrough-and-missing-auth-secret-52-passthrough-missing-auth.whoami.localhost-41": {
							EntryPoints: []string{"https"},
							Rule:        `HostSNI("passthrough-missing-auth.whoami.localhost")`,
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
							Service: "default-7-whoami-tls-10-443-3",
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-7-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{Address: "10.10.0.3:8443"},
									{Address: "10.10.0.4:8443"},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
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
			desc: "SSL Passthrough with force-ssl-redirect creates HTTP redirect router",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-ssl-passthrough-and-force-ssl-redirect.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-7-ingress-with-ssl-passthrough-and-force-ssl-redirect-51-passthrough-redirect.whoami.localhost-37": {
							EntryPoints: []string{"https"},
							Rule:        `HostSNI("passthrough-redirect.whoami.localhost")`,
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
							Service: "default-7-whoami-tls-10-443-3",
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-7-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.3:8443",
									},
									{
										Address: "10.10.0.4:8443",
									},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-ssl-passthrough-and-force-ssl-redirect-51-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("passthrough-redirect.whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-ssl-passthrough-and-force-ssl-redirect-51-rule-0-path-0-redirect-scheme", "default-7-ingress-with-ssl-passthrough-and-force-ssl-redirect-51-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-ssl-passthrough-and-force-ssl-redirect",
									},
								},
							},
							Service: "default-7-ingress-with-ssl-passthrough-and-force-ssl-redirect-51-whoami-tls-10-443-3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-ssl-passthrough-and-force-ssl-redirect-51-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-ssl-passthrough-and-force-ssl-redirect-51-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-ssl-passthrough-and-force-ssl-redirect-51-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: "wrr",
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.3:8443"},
									{URL: "http://10.10.0.4:8443"},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-ssl-passthrough-and-force-ssl-redirect-51",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-tls",
										Port:      "443",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-ssl-passthrough-and-force-ssl-redirect-51": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			// ssl-redirect set to false opts out of the redirect the TLS section would otherwise enable.
			desc: "SSL Passthrough with a TLS section and ssl-redirect disabled",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-ssl-passthrough-and-tls-section-and-ssl-redirect-disabled.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-7-ingress-with-ssl-passthrough-and-tls-section-and-ssl-redirect-disabled-70-passthrough-no-redirect.whoami.localhost-40": {
							EntryPoints: []string{"https"},
							Rule:        `HostSNI("passthrough-no-redirect.whoami.localhost")`,
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
							Service: "default-7-whoami-tls-10-443-3",
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-7-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{Address: "10.10.0.3:8443"},
									{Address: "10.10.0.4:8443"},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-ssl-passthrough-and-tls-section-and-ssl-redirect-disabled-70-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("passthrough-no-redirect.whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-ssl-passthrough-and-tls-section-and-ssl-redirect-disabled-70-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-ssl-passthrough-and-tls-section-and-ssl-redirect-disabled",
									},
								},
							},
							Service: "default-7-ingress-with-ssl-passthrough-and-tls-section-and-ssl-redirect-disabled-70-whoami-tls-10-443-3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-ssl-passthrough-and-tls-section-and-ssl-redirect-disabled-70-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-ssl-passthrough-and-tls-section-and-ssl-redirect-disabled-70-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: "wrr",
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.3:8443"},
									{URL: "http://10.10.0.4:8443"},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-ssl-passthrough-and-tls-section-and-ssl-redirect-disabled-70",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-tls",
										Port:      "443",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-ssl-passthrough-and-tls-section-and-ssl-redirect-disabled-70": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
				},
			},
		},
		{
			// The serversTransport only shapes the HTTP router: when it cannot be built,
			// the TCP passthrough router must still be created.
			desc: "SSL Passthrough keeps the TCP router when the serversTransport cannot be built",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-ssl-passthrough-and-invalid-proxy-ssl-secret.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-7-ingress-with-ssl-passthrough-and-invalid-proxy-ssl-secret-57-passthrough-invalid-transport.whoami.localhost-46": {
							EntryPoints: []string{"https"},
							Rule:        `HostSNI("passthrough-invalid-transport.whoami.localhost")`,
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
							Service: "default-7-whoami-tls-10-443-3",
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-7-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{Address: "10.10.0.3:8443"},
									{Address: "10.10.0.4:8443"},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
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
			desc: "Sticky Sessions",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-sticky.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-sticky-19-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("sticky.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-sticky-19-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-sticky-19-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-sticky",
									},
								},
							},
						},
						"default-7-ingress-with-sticky-19-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("sticky.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-sticky-19-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-sticky-19-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-sticky",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-sticky-19-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-sticky-19-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-sticky-19-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-sticky-19",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:               "foobar",
										Domain:             "foo.localhost",
										HTTPOnly:           true,
										MaxAge:             42,
										Expires:            42,
										Path:               new("/foobar"),
										SameSite:           "none",
										Secure:             true,
										PreserveLeadingDot: true,
									},
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-sticky-19": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Proxy SSL",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-ssl.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-ssl-22-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("proxy-ssl.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-ssl-22-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-proxy-ssl-22-whoami-tls-10-443-3",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-ssl",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-ssl-22-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("proxy-ssl.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-ssl-22-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-proxy-ssl-22-whoami-tls-10-443-3",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-ssl",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-ssl-22-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-ssl-22-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-ssl-22-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.3:8443",
									},
									{
										URL: "https://10.10.0.4:8443",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-ssl-22",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-tls",
										Port:      "443",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-ssl-22": {
							ServerName:         "whoami.localhost",
							InsecureSkipVerify: false,
							RootCAs:            []types.FileOrContent{"-----BEGIN CERTIFICATE-----"},
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "CORS",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-cors.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-cors-17-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("cors.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-cors-17-rule-0-path-0-cors", "default-7-ingress-with-cors-17-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-cors-17-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-cors",
									},
								},
							},
						},
						"default-7-ingress-with-cors-17-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("cors.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-cors-17-rule-0-path-0-tls-cors", "default-7-ingress-with-cors-17-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-cors-17-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-cors",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-cors-17-rule-0-path-0-cors": {
							Headers: &dynamic.Headers{
								AccessControlAllowCredentials: true,
								AccessControlAllowHeaders:     []string{"X-Foo"},
								AccessControlAllowMethods:     []string{"PUT", "GET", "POST", "OPTIONS"},
								AccessControlAllowOriginList:  []string{"*"},
								AccessControlExposeHeaders:    []string{"X-Forwarded-For", "X-Forwarded-Host"},
								AccessControlMaxAge:           new(int64(42)),
							},
						},
						"default-7-ingress-with-cors-17-rule-0-path-0-tls-cors": {
							Headers: &dynamic.Headers{
								AccessControlAllowCredentials: true,
								AccessControlAllowHeaders:     []string{"X-Foo"},
								AccessControlAllowMethods:     []string{"PUT", "GET", "POST", "OPTIONS"},
								AccessControlAllowOriginList:  []string{"*"},
								AccessControlExposeHeaders:    []string{"X-Forwarded-For", "X-Forwarded-Host"},
								AccessControlMaxAge:           new(int64(42)),
							},
						},
						"default-7-ingress-with-cors-17-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-cors-17-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-cors-17-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-cors-17",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-cors-17": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "CORS with wildcard origin",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-cors-wildcard-origin.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-cors-wildcard-origin-33-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("cors-wildcard.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-cors-wildcard-origin-33-rule-0-path-0-cors", "default-7-ingress-with-cors-wildcard-origin-33-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-cors-wildcard-origin-33-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-cors-wildcard-origin",
									},
								},
							},
						},
						"default-7-ingress-with-cors-wildcard-origin-33-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("cors-wildcard.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-cors-wildcard-origin-33-rule-0-path-0-tls-cors", "default-7-ingress-with-cors-wildcard-origin-33-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-cors-wildcard-origin-33-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-cors-wildcard-origin",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-cors-wildcard-origin-33-rule-0-path-0-cors": {
							Headers: &dynamic.Headers{
								AccessControlAllowCredentials:     true,
								AccessControlAllowHeaders:         []string{"X-Foo"},
								AccessControlAllowMethods:         []string{"PUT", "GET", "POST", "OPTIONS"},
								AccessControlAllowOriginList:      []string{"https://exact.example.org"},
								AccessControlAllowOriginListRegex: []string{`^(?i)https://[A-Za-z0-9-]+\.example\.com$`},
								AccessControlExposeHeaders:        []string{"X-Forwarded-For", "X-Forwarded-Host"},
								AccessControlMaxAge:               new(int64(42)),
							},
						},
						"default-7-ingress-with-cors-wildcard-origin-33-rule-0-path-0-tls-cors": {
							Headers: &dynamic.Headers{
								AccessControlAllowCredentials:     true,
								AccessControlAllowHeaders:         []string{"X-Foo"},
								AccessControlAllowMethods:         []string{"PUT", "GET", "POST", "OPTIONS"},
								AccessControlAllowOriginList:      []string{"https://exact.example.org"},
								AccessControlAllowOriginListRegex: []string{`^(?i)https://[A-Za-z0-9-]+\.example\.com$`},
								AccessControlExposeHeaders:        []string{"X-Forwarded-For", "X-Forwarded-Host"},
								AccessControlMaxAge:               new(int64(42)),
							},
						},
						"default-7-ingress-with-cors-wildcard-origin-33-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-cors-wildcard-origin-33-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-cors-wildcard-origin-33-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-cors-wildcard-origin-33",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-cors-wildcard-origin-33": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Service Upstream",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-service-upstream.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-service-upstream-29-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("service-upstream.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-service-upstream-29-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-service-upstream-29-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-service-upstream",
									},
								},
							},
						},
						"default-7-ingress-with-service-upstream-29-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("service-upstream.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-service-upstream-29-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-service-upstream-29-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-service-upstream",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-service-upstream-29-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-service-upstream-29-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-service-upstream-29-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.10.1:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-service-upstream-29",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-service-upstream-29": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Upstream vhost",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-upstream-vhost.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-upstream-vhost-27-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("upstream-vhost.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-upstream-vhost-27-rule-0-path-0-vhost", "default-7-ingress-with-upstream-vhost-27-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-upstream-vhost-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-upstream-vhost",
									},
								},
							},
						},
						"default-7-ingress-with-upstream-vhost-27-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("upstream-vhost.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-upstream-vhost-27-rule-0-path-0-tls-vhost", "default-7-ingress-with-upstream-vhost-27-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-upstream-vhost-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-upstream-vhost",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-upstream-vhost-27-rule-0-path-0-vhost": {
							UpstreamVHost: &dynamic.UpstreamVHost{
								VHost: "upstream-host-header-value",
								Vars: map[string]string{
									"$namespace":     "default",
									"$ingress_name":  "ingress-with-upstream-vhost",
									"$location_path": "/",
									"$service_name":  "whoami",
									"$service_port":  "80",
								},
							},
						},
						"default-7-ingress-with-upstream-vhost-27-rule-0-path-0-tls-vhost": {
							UpstreamVHost: &dynamic.UpstreamVHost{
								VHost: "upstream-host-header-value",
								Vars: map[string]string{
									"$namespace":     "default",
									"$ingress_name":  "ingress-with-upstream-vhost",
									"$location_path": "/",
									"$service_name":  "whoami",
									"$service_port":  "80",
								},
							},
						},
						"default-7-ingress-with-upstream-vhost-27-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-upstream-vhost-27-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-upstream-vhost-27-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-upstream-vhost-27",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-upstream-vhost-27": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "X-forwarded-prefix with missing rewrite-target",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-x-forwarded-prefix-no-rewrite-target.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-x-forwarded-prefix-no-rewrite-target-49-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("x-forwarded-prefix.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-x-forwarded-prefix-no-rewrite-target-49-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-x-forwarded-prefix-no-rewrite-target-49-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-x-forwarded-prefix-no-rewrite-target",
									},
								},
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-no-rewrite-target-49-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("x-forwarded-prefix.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-x-forwarded-prefix-no-rewrite-target-49-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-x-forwarded-prefix-no-rewrite-target-49-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-x-forwarded-prefix-no-rewrite-target",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-x-forwarded-prefix-no-rewrite-target-49-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-no-rewrite-target-49-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-no-rewrite-target-49-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-x-forwarded-prefix-no-rewrite-target-49",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-x-forwarded-prefix-no-rewrite-target-49": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "X-forwarded-prefix",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-x-forwarded-prefix.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-x-forwarded-prefix-31-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("x-forwarded-prefix.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-x-forwarded-prefix-31-rule-0-path-0-rewrite-target", "default-7-ingress-with-x-forwarded-prefix-31-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-x-forwarded-prefix-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-x-forwarded-prefix",
									},
								},
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-regex-37-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("x-forwarded-prefix-regex.localhost") && PathRegexp("(?i)^/(something)(/.+)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-x-forwarded-prefix-regex-37-rule-0-path-0-rewrite-target", "default-7-ingress-with-x-forwarded-prefix-regex-37-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-x-forwarded-prefix-regex-37-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-x-forwarded-prefix-regex",
									},
								},
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-three-groups-44-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("x-forwarded-prefix-three-groups.localhost") && PathRegexp("(?i)^/(prefix)/(sub)/(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-x-forwarded-prefix-three-groups-44-rule-0-path-0-rewrite-target", "default-7-ingress-with-x-forwarded-prefix-three-groups-44-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-x-forwarded-prefix-three-groups-44-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-x-forwarded-prefix-three-groups",
									},
								},
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("x-forwarded-prefix.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-x-forwarded-prefix-31-rule-0-path-0-tls-rewrite-target", "default-7-ingress-with-x-forwarded-prefix-31-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-x-forwarded-prefix-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-x-forwarded-prefix",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-x-forwarded-prefix-regex-37-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("x-forwarded-prefix-regex.localhost") && PathRegexp("(?i)^/(something)(/.+)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-x-forwarded-prefix-regex-37-rule-0-path-0-tls-rewrite-target", "default-7-ingress-with-x-forwarded-prefix-regex-37-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-x-forwarded-prefix-regex-37-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-x-forwarded-prefix-regex",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-x-forwarded-prefix-three-groups-44-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("x-forwarded-prefix-three-groups.localhost") && PathRegexp("(?i)^/(prefix)/(sub)/(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-x-forwarded-prefix-three-groups-44-rule-0-path-0-tls-rewrite-target", "default-7-ingress-with-x-forwarded-prefix-three-groups-44-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-x-forwarded-prefix-three-groups-44-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-x-forwarded-prefix-three-groups",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-x-forwarded-prefix-three-groups-44-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:            "/(prefix)/(sub)/(.*)",
								Replacement:      "/$3",
								XForwardedPrefix: "/$1/$2",
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-three-groups-44-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:            "/(prefix)/(sub)/(.*)",
								Replacement:      "/$3",
								XForwardedPrefix: "/$1/$2",
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-three-groups-44-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-three-groups-44-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-31-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:            "/",
								Replacement:      "/path",
								XForwardedPrefix: "x-forwarded-prefix-header-value",
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-31-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:            "/",
								Replacement:      "/path",
								XForwardedPrefix: "x-forwarded-prefix-header-value",
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-31-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-31-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-regex-37-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:            "/(something)(/.+)",
								Replacement:      "$2",
								XForwardedPrefix: "$1",
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-regex-37-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:            "/(something)(/.+)",
								Replacement:      "$2",
								XForwardedPrefix: "$1",
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-regex-37-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-regex-37-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-31-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-x-forwarded-prefix-31",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-three-groups-44-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-x-forwarded-prefix-three-groups-44",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-regex-37-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-x-forwarded-prefix-regex-37",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-x-forwarded-prefix-31": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-three-groups-44": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-x-forwarded-prefix-regex-37": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Use Regex",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-use-regex.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-use-regex-22-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("use-regex.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-use-regex-22-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-use-regex-22-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-use-regex",
									},
								},
							},
						},
						"default-7-ingress-with-use-regex-22-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("use-regex.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-use-regex-22-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-use-regex-22-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-use-regex",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-use-regex-22-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-use-regex-22-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-use-regex-22-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-use-regex-22",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-use-regex-22": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Use Regex cross ingress",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-use-regex-cross-ingress.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-a-with-use-regex-24-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-a-with-use-regex-24-rule-0-path-0-retry"},
							Service:     "default-7-ingress-a-with-use-regex-24-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-a-with-use-regex",
									},
								},
							},
						},
						"default-7-ingress-b-without-use-regex-27-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/static")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-b-without-use-regex-27-rule-0-path-0-retry"},
							Service:     "default-7-ingress-b-without-use-regex-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-b-without-use-regex",
									},
								},
							},
						},
						"default-7-ingress-a-with-use-regex-24-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-a-with-use-regex-24-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-a-with-use-regex-24-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-a-with-use-regex",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-b-without-use-regex-27-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/static")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-b-without-use-regex-27-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-b-without-use-regex-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-b-without-use-regex",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-a-with-use-regex-24-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-a-with-use-regex-24-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-b-without-use-regex-27-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-b-without-use-regex-27-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-a-with-use-regex-24-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-a-with-use-regex-24",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-b-without-use-regex-27-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-b-without-use-regex-27",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-a-with-use-regex-24": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-b-without-use-regex-27": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Use Regex isolated",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-use-regex-isolated.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-a-with-use-regex-24-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("regex.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-a-with-use-regex-24-rule-0-path-0-retry"},
							Service:     "default-7-ingress-a-with-use-regex-24-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-a-with-use-regex",
									},
								},
							},
						},
						"default-7-ingress-b-without-use-regex-27-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("no-regex.localhost") && (Path("/static") || PathPrefix("/static/"))`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-b-without-use-regex-27-rule-0-path-0-retry"},
							Service:     "default-7-ingress-b-without-use-regex-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-b-without-use-regex",
									},
								},
							},
						},
						"default-7-ingress-a-with-use-regex-24-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("regex.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-a-with-use-regex-24-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-a-with-use-regex-24-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-a-with-use-regex",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-b-without-use-regex-27-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("no-regex.localhost") && (Path("/static") || PathPrefix("/static/"))`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-b-without-use-regex-27-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-b-without-use-regex-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-b-without-use-regex",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-a-with-use-regex-24-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-a-with-use-regex-24-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-b-without-use-regex-27-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-b-without-use-regex-27-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-a-with-use-regex-24-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-a-with-use-regex-24",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-b-without-use-regex-27-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-b-without-use-regex-27",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-a-with-use-regex-24": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-b-without-use-regex-27": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Rewrite Target",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-rewrite-target.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-rewrite-target-27-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("rewrite-target.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-rewrite-target-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-rewrite-target",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-rewrite-target-27-rule-0-path-0-rewrite-target", "default-7-ingress-with-rewrite-target-27-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-rewrite-target-27-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("rewrite-target.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-rewrite-target-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-rewrite-target",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-rewrite-target-27-rule-0-path-0-tls-rewrite-target", "default-7-ingress-with-rewrite-target-27-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("rewrite-target-no-regex.localhost") && (Path("/something") || PathPrefix("/something/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-rewrite-target-no-regex-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-rewrite-target-no-regex",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("rewrite-target-no-regex.localhost") && (Path("/something") || PathPrefix("/something/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-rewrite-target-no-regex-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-rewrite-target-no-regex",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-rewrite-target-27-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-7-ingress-with-rewrite-target-27-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-7-ingress-with-rewrite-target-27-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-rewrite-target-27-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-rewrite-target-27-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-rewrite-target-27",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-rewrite-target-no-regex-36-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-rewrite-target-no-regex-36",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-rewrite-target-27": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-rewrite-target-no-regex-36": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Rewrite Target with ImplementationSpecific path and absolute URL",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-rewrite-target-absolute-url.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-rewrite-target-absolute-url-40-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("rewrite-target-absolute-url.localhost") && PathRegexp("(?i)^/original(?:/(.*))?")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-rewrite-target-absolute-url-40-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-rewrite-target-absolute-url",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-rewrite-target-absolute-url-40-rule-0-path-0-rewrite-target", "default-7-ingress-with-rewrite-target-absolute-url-40-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-rewrite-target-absolute-url-40-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("rewrite-target-absolute-url.localhost") && PathRegexp("(?i)^/original(?:/(.*))?")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-rewrite-target-absolute-url-40-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-rewrite-target-absolute-url",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-rewrite-target-absolute-url-40-rule-0-path-0-tls-rewrite-target", "default-7-ingress-with-rewrite-target-absolute-url-40-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-rewrite-target-absolute-url-40-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/original(?:/(.*))?",
								Replacement: "https://bar.example.org/$1",
							},
						},
						"default-7-ingress-with-rewrite-target-absolute-url-40-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/original(?:/(.*))?",
								Replacement: "https://bar.example.org/$1",
							},
						},
						"default-7-ingress-with-rewrite-target-absolute-url-40-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-rewrite-target-absolute-url-40-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-rewrite-target-absolute-url-40-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-rewrite-target-absolute-url-40",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-rewrite-target-absolute-url-40": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Rewrite Target with ImplementationSpecific trailing capture group path",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-rewrite-target-trailing-group.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-rewrite-target-trailing-group-42-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("rewrite-target-trailing-group.localhost") && PathRegexp("(?i)^/original/(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-rewrite-target-trailing-group-42-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-rewrite-target-trailing-group",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-rewrite-target-trailing-group-42-rule-0-path-0-rewrite-target", "default-7-ingress-with-rewrite-target-trailing-group-42-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-rewrite-target-trailing-group-42-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("rewrite-target-trailing-group.localhost") && PathRegexp("(?i)^/original/(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-rewrite-target-trailing-group-42-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-rewrite-target-trailing-group",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-rewrite-target-trailing-group-42-rule-0-path-0-tls-rewrite-target", "default-7-ingress-with-rewrite-target-trailing-group-42-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-rewrite-target-trailing-group-42-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/original/(.*)",
								Replacement: "/newpath/$1",
							},
						},
						"default-7-ingress-with-rewrite-target-trailing-group-42-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/original/(.*)",
								Replacement: "/newpath/$1",
							},
						},
						"default-7-ingress-with-rewrite-target-trailing-group-42-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-rewrite-target-trailing-group-42-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-rewrite-target-trailing-group-42-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-rewrite-target-trailing-group-42",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-rewrite-target-trailing-group-42": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Rewrite Target without use-regex",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-rewrite-target-no-regex.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("rewrite-target-no-regex.localhost") && Path("/original")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-rewrite-target-no-regex-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-rewrite-target-no-regex",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-rewrite-target",
								"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-retry",
							},
						},
						"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("rewrite-target-no-regex.localhost") && Path("/original")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-rewrite-target-no-regex-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-rewrite-target-no-regex",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-tls-rewrite-target",
								"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/original",
								Replacement: "/rewritten",
							},
						},
						"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/original",
								Replacement: "/rewritten",
							},
						},
						"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-rewrite-target-no-regex-36-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-rewrite-target-no-regex-36-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-rewrite-target-no-regex-36",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-rewrite-target-no-regex-36": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Rewrite target cross ingress",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-rewrite-target-cross-ingress.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-a-with-rewrite-target-29-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-a-with-rewrite-target",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-rewrite-target", "default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-retry"},
						},
						"default-7-ingress-b-without-rewrite-target-32-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/static")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-b-without-rewrite-target-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-b-without-rewrite-target",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-b-without-rewrite-target-32-rule-0-path-0-retry"},
						},
						"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-a-with-rewrite-target-29-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-a-with-rewrite-target",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-tls-rewrite-target", "default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-b-without-rewrite-target-32-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/static")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-b-without-rewrite-target-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-b-without-rewrite-target",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-b-without-rewrite-target-32-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-b-without-rewrite-target-32-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-b-without-rewrite-target-32-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-a-with-rewrite-target-29-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-a-with-rewrite-target-29",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-b-without-rewrite-target-32-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-b-without-rewrite-target-32",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-a-with-rewrite-target-29": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-b-without-rewrite-target-32": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Rewrite target isolated",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-rewrite-target-isolated.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("rewrite.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-a-with-rewrite-target-29-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-a-with-rewrite-target",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-rewrite-target", "default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-retry"},
						},
						"default-7-ingress-b-without-rewrite-target-32-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("no-rewrite.localhost") && (Path("/static") || PathPrefix("/static/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-b-without-rewrite-target-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-b-without-rewrite-target",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-b-without-rewrite-target-32-rule-0-path-0-retry"},
						},
						"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("rewrite.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-a-with-rewrite-target-29-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-a-with-rewrite-target",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-tls-rewrite-target", "default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-b-without-rewrite-target-32-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("no-rewrite.localhost") && (Path("/static") || PathPrefix("/static/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-b-without-rewrite-target-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-b-without-rewrite-target",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-b-without-rewrite-target-32-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-a-with-rewrite-target-29-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-b-without-rewrite-target-32-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-b-without-rewrite-target-32-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-a-with-rewrite-target-29-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-a-with-rewrite-target-29",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-b-without-rewrite-target-32-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-b-without-rewrite-target-32",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-a-with-rewrite-target-29": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-b-without-rewrite-target-32": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Rewrite target with use-regex false still applies regex",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-rewrite-target-use-regex-false.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-rewrite-target-use-regex-false-43-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("rewrite-target.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-rewrite-target-use-regex-false-43-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-rewrite-target-use-regex-false",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-rewrite-target-use-regex-false-43-rule-0-path-0-rewrite-target", "default-7-ingress-with-rewrite-target-use-regex-false-43-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-rewrite-target-use-regex-false-43-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("rewrite-target.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-rewrite-target-use-regex-false-43-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-rewrite-target-use-regex-false",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-rewrite-target-use-regex-false-43-rule-0-path-0-tls-rewrite-target", "default-7-ingress-with-rewrite-target-use-regex-false-43-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-rewrite-target-use-regex-false-43-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-7-ingress-with-rewrite-target-use-regex-false-43-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-7-ingress-with-rewrite-target-use-regex-false-43-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-rewrite-target-use-regex-false-43-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-rewrite-target-use-regex-false-43-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-rewrite-target-use-regex-false-43",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-rewrite-target-use-regex-false-43": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "App Root",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-app-root.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-app-root-21-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-21-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-21-rule-0-path-0-app-root", "default-7-ingress-with-app-root-21-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-app-root-21-rule-0-path-0-app-root-redirect": {
							EntryPoints: []string{"http"},
							Rule:        `Host("app-root.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "unavailable-service",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-21-rule-0-path-0-app-root"},
						},
						"default-7-ingress-with-app-root-21-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-21-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-21-rule-0-path-0-tls-app-root", "default-7-ingress-with-app-root-21-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-app-root-21-rule-0-path-0-tls-app-root-redirect": {
							EntryPoints: []string{"https"},
							Rule:        `Host("app-root.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "unavailable-service",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-21-rule-0-path-0-tls-app-root"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-app-root-21-rule-0-path-0-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-21-rule-0-path-0-tls-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-21-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-app-root-21-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-app-root-21-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-app-root-21",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-app-root-21": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "App Root - server alias",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-app-root-and-server-alias.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-app-root-alias-27-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("app-root.localhost") || Host("alias1.localhost")) && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-alias-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-alias",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-alias-27-rule-0-path-0-app-root", "default-7-ingress-with-app-root-alias-27-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-app-root-alias-27-rule-0-path-0-app-root-redirect": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("app-root.localhost") || Host("alias1.localhost")) && Path("/")`,
							RuleSyntax:  "default",
							Service:     "unavailable-service",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-alias",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-alias-27-rule-0-path-0-app-root"},
						},
						"default-7-ingress-with-app-root-alias-27-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("app-root.localhost") || Host("alias1.localhost")) && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-alias-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-alias",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-alias-27-rule-0-path-0-tls-app-root", "default-7-ingress-with-app-root-alias-27-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-app-root-alias-27-rule-0-path-0-tls-app-root-redirect": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("app-root.localhost") || Host("alias1.localhost")) && Path("/")`,
							RuleSyntax:  "default",
							Service:     "unavailable-service",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-alias",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-alias-27-rule-0-path-0-tls-app-root"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-app-root-alias-27-rule-0-path-0-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-alias-27-rule-0-path-0-tls-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-alias-27-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-app-root-alias-27-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-app-root-alias-27-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-app-root-alias-27",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-app-root-alias-27": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "App Root - root path declared",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-app-root-and-root-path.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-app-root-root-path-31-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("app-root.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-root-path-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-root-path",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-root-path-31-rule-0-path-0-app-root", "default-7-ingress-with-app-root-root-path-31-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-app-root-root-path-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("app-root.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-root-path-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-root-path",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-root-path-31-rule-0-path-0-tls-app-root", "default-7-ingress-with-app-root-root-path-31-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-app-root-root-path-31-rule-0-path-1": {
							EntryPoints: []string{"http"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-root-path-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-root-path",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-root-path-31-rule-0-path-1-app-root", "default-7-ingress-with-app-root-root-path-31-rule-0-path-1-retry"},
						},
						"default-7-ingress-with-app-root-root-path-31-rule-0-path-1-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-root-path-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-root-path",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-root-path-31-rule-0-path-1-tls-app-root", "default-7-ingress-with-app-root-root-path-31-rule-0-path-1-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-app-root-root-path-31-rule-0-path-0-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-root-path-31-rule-0-path-0-tls-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-root-path-31-rule-0-path-1-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-root-path-31-rule-0-path-1-tls-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-root-path-31-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-app-root-root-path-31-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-app-root-root-path-31-rule-0-path-1-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-app-root-root-path-31-rule-0-path-1-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-app-root-root-path-31-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-app-root-root-path-31",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-app-root-root-path-31": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "App Root - root path declared by another Ingress on the same host",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-app-root-shared-host.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-app-root-shared-host-root-38-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("app-root.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-shared-host-root-38-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-shared-host-root",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-shared-host-root-38-rule-0-path-0-app-root", "default-7-ingress-with-app-root-shared-host-root-38-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-app-root-shared-host-root-38-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("app-root.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-shared-host-root-38-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-shared-host-root",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-shared-host-root-38-rule-0-path-0-tls-app-root", "default-7-ingress-with-app-root-shared-host-root-38-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-app-root-shared-host-33-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-shared-host-33-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-shared-host",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-shared-host-33-rule-0-path-0-app-root", "default-7-ingress-with-app-root-shared-host-33-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-app-root-shared-host-33-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-shared-host-33-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-shared-host",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-shared-host-33-rule-0-path-0-tls-app-root", "default-7-ingress-with-app-root-shared-host-33-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-app-root-shared-host-root-38-rule-0-path-0-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-shared-host-root-38-rule-0-path-0-tls-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-shared-host-33-rule-0-path-0-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-shared-host-33-rule-0-path-0-tls-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-shared-host-root-38-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-app-root-shared-host-root-38-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-app-root-shared-host-33-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-app-root-shared-host-33-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-app-root-shared-host-root-38-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-app-root-shared-host-root-38",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-app-root-shared-host-33-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-app-root-shared-host-33",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-app-root-shared-host-root-38": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-app-root-shared-host-33": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "App Root - default backend declared",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-app-root-and-default-backend.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-app-root-and-default-backend-41-rule-0-default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `Host("app-root.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-and-default-backend-41-default-backend",
							Middlewares: []string{"default-7-ingress-with-app-root-and-default-backend-41-rule-0-default-backend-app-root", "default-7-ingress-with-app-root-and-default-backend-41-rule-0-default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-and-default-backend",
									},
								},
							},
						},
						"default-7-ingress-with-app-root-and-default-backend-41-rule-0-default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("app-root.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-and-default-backend-41-default-backend",
							Middlewares: []string{"default-7-ingress-with-app-root-and-default-backend-41-rule-0-default-backend-tls-app-root", "default-7-ingress-with-app-root-and-default-backend-41-rule-0-default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-and-default-backend",
									},
								},
							},
						},
						"default-7-ingress-with-app-root-and-default-backend-41-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-and-default-backend-41-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-app-root-and-default-backend-41-rule-0-path-0-app-root", "default-7-ingress-with-app-root-and-default-backend-41-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-and-default-backend",
									},
								},
							},
						},
						"default-7-ingress-with-app-root-and-default-backend-41-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-and-default-backend-41-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-app-root-and-default-backend-41-rule-0-path-0-tls-app-root", "default-7-ingress-with-app-root-and-default-backend-41-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-and-default-backend",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-app-root-and-default-backend-41-rule-0-default-backend-app-root": {
							AppRoot: &dynamic.AppRoot{Path: "/foo"},
						},
						"default-7-ingress-with-app-root-and-default-backend-41-rule-0-default-backend-tls-app-root": {
							AppRoot: &dynamic.AppRoot{Path: "/foo"},
						},
						"default-7-ingress-with-app-root-and-default-backend-41-rule-0-path-0-app-root": {
							AppRoot: &dynamic.AppRoot{Path: "/foo"},
						},
						"default-7-ingress-with-app-root-and-default-backend-41-rule-0-path-0-tls-app-root": {
							AppRoot: &dynamic.AppRoot{Path: "/foo"},
						},
						"default-7-ingress-with-app-root-and-default-backend-41-rule-0-default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-app-root-and-default-backend-41-rule-0-default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-app-root-and-default-backend-41-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-app-root-and-default-backend-41-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-app-root-and-default-backend-41-default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-app-root-and-default-backend-41",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-app-root-and-default-backend-41-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-app-root-and-default-backend-41",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-app-root-and-default-backend-41": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "App Root - no host declared",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-app-root-no-host.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-app-root-no-host-29-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `(Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-no-host-29-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-no-host",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-no-host-29-rule-0-path-0-app-root", "default-7-ingress-with-app-root-no-host-29-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-app-root-no-host-29-rule-0-path-0-app-root-redirect": {
							EntryPoints: []string{"http"},
							Rule:        `Path("/")`,
							RuleSyntax:  "default",
							Service:     "unavailable-service",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-no-host",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-no-host-29-rule-0-path-0-app-root"},
						},
						"default-7-ingress-with-app-root-no-host-29-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-app-root-no-host-29-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-no-host",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-no-host-29-rule-0-path-0-tls-app-root", "default-7-ingress-with-app-root-no-host-29-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-app-root-no-host-29-rule-0-path-0-tls-app-root-redirect": {
							EntryPoints: []string{"https"},
							Rule:        `Path("/")`,
							RuleSyntax:  "default",
							Service:     "unavailable-service",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root-no-host",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-app-root-no-host-29-rule-0-path-0-tls-app-root"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-app-root-no-host-29-rule-0-path-0-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-no-host-29-rule-0-path-0-tls-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-app-root-no-host-29-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-app-root-no-host-29-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-app-root-no-host-29-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-app-root-no-host-29",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-app-root-no-host-29": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "App Root - no prefix slash",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-app-root-wrong.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-app-root-21-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-app-root-21-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-app-root-21-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root",
									},
								},
							},
						},
						"default-7-ingress-with-app-root-21-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-app-root-21-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-app-root-21-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-app-root",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-app-root-21-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-app-root-21-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-app-root-21-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-app-root-21",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-app-root-21": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "From To WWW Redirect - www host",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-www-host.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-www-host-21-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("www.host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-www-host-21-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-www-host-21-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-www-host",
									},
								},
							},
						},
						"default-7-ingress-with-www-host-21-rule-0-path-0-from-to-www-redirect": {
							EntryPoints: []string{"http"},
							Rule:        `Host("host.localhost")`,
							RuleSyntax:  "default",
							Service:     unavailableServiceName,
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-www-host",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-www-host-21-rule-0-path-0-from-to-www-redirect"},
						},
						"default-7-ingress-with-www-host-21-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("www.host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-www-host-21-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-www-host-21-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-www-host",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-www-host-21-rule-0-path-0-tls-from-to-www-redirect": {
							EntryPoints: []string{"https"},
							Rule:        `Host("host.localhost")`,
							RuleSyntax:  "default",
							Service:     unavailableServiceName,
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-www-host",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-www-host-21-rule-0-path-0-tls-from-to-www-redirect"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-www-host-21-rule-0-path-0-from-to-www-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?)://(?:\[[^/\]]*\]|[^/:]+)(:[0-9]+)?[^/]*/(.*?)/?$`,
								Replacement: "$1://www.host.localhost$2/$3",
								StatusCode:  new(http.StatusPermanentRedirect),
							},
						},
						"default-7-ingress-with-www-host-21-rule-0-path-0-tls-from-to-www-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?)://(?:\[[^/\]]*\]|[^/:]+)(:[0-9]+)?[^/]*/(.*?)/?$`,
								Replacement: "$1://www.host.localhost$2/$3",
								StatusCode:  new(http.StatusPermanentRedirect),
							},
						},
						"default-7-ingress-with-www-host-21-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-www-host-21-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-www-host-21-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-www-host-21",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-www-host-21": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "From To WWW Redirect - host",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-host.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-host-17-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-host-17-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-host-17-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-host",
									},
								},
							},
						},
						"default-7-ingress-with-host-17-rule-0-path-0-from-to-www-redirect": {
							EntryPoints: []string{"http"},
							Rule:        `Host("www.host.localhost")`,
							RuleSyntax:  "default",
							Service:     unavailableServiceName,
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-host",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-host-17-rule-0-path-0-from-to-www-redirect"},
						},
						"default-7-ingress-with-host-17-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-host-17-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-host-17-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-host",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-host-17-rule-0-path-0-tls-from-to-www-redirect": {
							EntryPoints: []string{"https"},
							Rule:        `Host("www.host.localhost")`,
							RuleSyntax:  "default",
							Service:     unavailableServiceName,
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-host",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-host-17-rule-0-path-0-tls-from-to-www-redirect"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-host-17-rule-0-path-0-from-to-www-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?)://(?:\[[^/\]]*\]|[^/:]+)(:[0-9]+)?[^/]*/(.*?)/?$`,
								Replacement: "$1://host.localhost$2/$3",
								StatusCode:  new(http.StatusPermanentRedirect),
							},
						},
						"default-7-ingress-with-host-17-rule-0-path-0-tls-from-to-www-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?)://(?:\[[^/\]]*\]|[^/:]+)(:[0-9]+)?[^/]*/(.*?)/?$`,
								Replacement: "$1://host.localhost$2/$3",
								StatusCode:  new(http.StatusPermanentRedirect),
							},
						},
						"default-7-ingress-with-host-17-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-host-17-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-host-17-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-host-17",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-host-17": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "From To WWW Redirect - multiple ingresses",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-www-redirect.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-host-17-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-host-17-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-host-17-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-host",
									},
								},
							},
						},
						"default-7-ingress-with-www-host-21-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("www.host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-www-host-21-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-www-host-21-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-www-host",
									},
								},
							},
						},
						"default-7-ingress-with-host-17-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-host-17-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-host-17-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-host",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-www-host-21-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("www.host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-www-host-21-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-www-host-21-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-www-host",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-host-17-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-host-17-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-www-host-21-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-www-host-21-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-host-17-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-host-17",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-www-host-21-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-www-host-21",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-www-host-21": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-host-17": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                           "Default Backend",
			defaultBackendServiceName:      "whoami",
			defaultBackendServiceNamespace: "default",
			paths: []string{
				"services.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
									},
								},
							},
						},
						"default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							TLS:         &dynamic.RouterTLSConfig{},
							Service:     "default-backend",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:8000",
									},
									{
										URL: "http://10.10.0.2:8000",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "0",
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
			desc: "Ingress default backend without rules emits catch-all routers with middleware annotations applied",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-no-rules.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-cors", "default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules",
									},
								},
							},
						},
						"default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-tls-cors", "default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-backend-cors": {
							Headers: &dynamic.Headers{
								AccessControlAllowCredentials: true,
								AccessControlExposeHeaders:    []string{},
								AccessControlAllowHeaders:     []string{"DNT", "Keep-Alive", "User-Agent", "X-Requested-With", "If-Modified-Since", "Cache-Control", "Content-Type", "Range", "Authorization"},
								AccessControlAllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH", "OPTIONS"},
								AccessControlAllowOriginList:  []string{"*"},
								AccessControlMaxAge:           new(int64(1728000)),
							},
						},
						"default-backend-tls-cors": {
							Headers: &dynamic.Headers{
								AccessControlAllowCredentials: true,
								AccessControlExposeHeaders:    []string{},
								AccessControlAllowHeaders:     []string{"DNT", "Keep-Alive", "User-Agent", "X-Requested-With", "If-Modified-Since", "Cache-Control", "Content-Type", "Range", "Authorization"},
								AccessControlAllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH", "OPTIONS"},
								AccessControlAllowOriginList:  []string{"*"},
								AccessControlMaxAge:           new(int64(1728000)),
							},
						},
						"default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-no-rules-37",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-no-rules-37": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Ingress default backend without rules inherits basic auth on the catch-all routers",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-no-rules-basic-auth.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-basic-auth", "default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-basic-auth",
									},
								},
							},
						},
						"default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-tls-basic-auth", "default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-basic-auth",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-backend-basic-auth": {
							BasicAuth: &dynamic.BasicAuth{
								Users: dynamic.Users{
									"urn:k8s:secret:default/basic-auth:auth",
								},
								Realm: "Authentication Required",
							},
							URNRefs: map[string]string{
								"/basicAuth/users/0": "user:{SHA}W6ph5Mm5Pz8GgiULbPgzG37mj9g=",
							},
						},
						"default-backend-tls-basic-auth": {
							BasicAuth: &dynamic.BasicAuth{
								Users: dynamic.Users{
									"urn:k8s:secret:default/basic-auth:auth",
								},
								Realm: "Authentication Required",
							},
							URNRefs: map[string]string{
								"/basicAuth/users/0": "user:{SHA}W6ph5Mm5Pz8GgiULbPgzG37mj9g=",
							},
						},
						"default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-backend-remove-authorization-header": {
							Headers: &dynamic.Headers{
								CustomRequestHeaders: map[string]string{"Authorization": ""},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-no-rules-basic-auth-48",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Middlewares: []string{"default-backend-remove-authorization-header"},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-no-rules-basic-auth-48": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                         "Ingress default backend without rules inherits custom headers on the catch-all routers",
			globalAllowedResponseHeaders: []string{"X-Custom-Header"},
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-no-rules-custom-headers.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-custom-headers", "default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-custom-headers",
									},
								},
							},
						},
						"default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-tls-custom-headers", "default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-custom-headers",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-backend-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-backend-tls-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-no-rules-custom-headers-52",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-no-rules-custom-headers-52": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Ingress default backend without rules inherits the client-auth TLS option on the catch-all TLS router",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-no-rules-auth-tls.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-auth-tls",
									},
								},
							},
						},
						"default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-tls-retry"},
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-7-ca-secret-9-require-and-verify-client-cert",
							},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-auth-tls",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-no-rules-auth-tls-46",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-no-rules-auth-tls-46": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"default-7-ca-secret-9-require-and-verify-client-cert": {
							ClientAuth: tls.ClientAuth{
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----"},
								ClientAuthType: tls.RequireAndVerifyClientCert,
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
							ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
						},
					},
				},
			},
		},
		{
			desc: "Ingress per-host default backend inherits basic auth on the fallback router",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-per-host-basic-auth.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-per-host-basic-auth-48-default-backend",
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-default-backend-basic-auth", "default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-per-host-basic-auth",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-per-host-basic-auth-48-default-backend",
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-default-backend-tls-basic-auth", "default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-per-host-basic-auth",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/foo")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-per-host-basic-auth-48-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-path-0-basic-auth", "default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-per-host-basic-auth",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/foo")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-per-host-basic-auth-48-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-path-0-tls-basic-auth", "default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-per-host-basic-auth",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-default-backend-basic-auth": {
							BasicAuth: &dynamic.BasicAuth{
								Users: dynamic.Users{"urn:k8s:secret:default/basic-auth:auth"},
								Realm: "Authentication Required",
							},
							URNRefs: map[string]string{
								"/basicAuth/users/0": "user:{SHA}W6ph5Mm5Pz8GgiULbPgzG37mj9g=",
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-default-backend-tls-basic-auth": {
							BasicAuth: &dynamic.BasicAuth{
								Users: dynamic.Users{"urn:k8s:secret:default/basic-auth:auth"},
								Realm: "Authentication Required",
							},
							URNRefs: map[string]string{
								"/basicAuth/users/0": "user:{SHA}W6ph5Mm5Pz8GgiULbPgzG37mj9g=",
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-path-0-basic-auth": {
							BasicAuth: &dynamic.BasicAuth{
								Users: dynamic.Users{"urn:k8s:secret:default/basic-auth:auth"},
								Realm: "Authentication Required",
							},
							URNRefs: map[string]string{
								"/basicAuth/users/0": "user:{SHA}W6ph5Mm5Pz8GgiULbPgzG37mj9g=",
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-path-0-tls-basic-auth": {
							BasicAuth: &dynamic.BasicAuth{
								Users: dynamic.Users{"urn:k8s:secret:default/basic-auth:auth"},
								Realm: "Authentication Required",
							},
							URNRefs: map[string]string{
								"/basicAuth/users/0": "user:{SHA}W6ph5Mm5Pz8GgiULbPgzG37mj9g=",
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-default-backend-remove-authorization-header": {
							Headers: &dynamic.Headers{
								CustomRequestHeaders: map[string]string{"Authorization": ""},
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-whoami-6-80-2-remove-authorization-header": {
							Headers: &dynamic.Headers{
								CustomRequestHeaders: map[string]string{"Authorization": ""},
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-default-backend": {
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-basic-auth-48-default-backend-remove-authorization-header"},
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-per-host-basic-auth-48",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-basic-auth-48-whoami-6-80-2": {
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-basic-auth-48-whoami-6-80-2-remove-authorization-header"},
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-per-host-basic-auth-48",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-per-host-basic-auth-48": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                         "Ingress per-host default backend inherits custom headers on the fallback router",
			globalAllowedResponseHeaders: []string{"X-Custom-Header"},
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-per-host-custom-headers.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-per-host-custom-headers-52-default-backend",
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-default-backend-custom-headers", "default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-per-host-custom-headers",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-per-host-custom-headers-52-default-backend",
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-default-backend-tls-custom-headers", "default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-per-host-custom-headers",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/foo")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-per-host-custom-headers-52-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-path-0-custom-headers", "default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-per-host-custom-headers",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/foo")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-per-host-custom-headers-52-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-path-0-tls-custom-headers", "default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-per-host-custom-headers",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-default-backend-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-default-backend-tls-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-path-0-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-path-0-tls-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-per-host-custom-headers-52",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-headers-52-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-per-host-custom-headers-52",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-per-host-custom-headers-52": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Ingress default backend without rules inherits custom http errors on the catch-all routers",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-no-rules-custom-http-errors.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-custom-http-errors", "default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-custom-http-errors",
									},
								},
							},
						},
						"default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-tls-custom-http-errors", "default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-custom-http-errors",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-backend-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "415"},
								Service: "default-backend-default-backend",
								NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-default-backend-no-rules-custom-http-errors"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-backend-tls-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "415"},
								Service: "default-backend-default-backend-tls",
								NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-default-backend-no-rules-custom-http-errors"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-no-rules-custom-http-errors-56",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-backend-default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.5:8000"},
									{URL: "http://10.10.0.6:8000"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "",
										Port:      "",
									},
								},
							},
						},
						"default-backend-default-backend-tls": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.5:8000"},
									{URL: "http://10.10.0.6:8000"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "",
										Port:      "",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-no-rules-custom-http-errors-56": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Ingress per-host default backend inherits custom http errors on the fallback router",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-per-host-custom-http-errors.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-per-host-custom-http-errors-56-default-backend",
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend-custom-http-errors", "default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-per-host-custom-http-errors",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-per-host-custom-http-errors-56-default-backend",
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend-tls-custom-http-errors", "default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-per-host-custom-http-errors",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/foo")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-per-host-custom-http-errors-56-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0-custom-http-errors", "default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-per-host-custom-http-errors",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/foo")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-per-host-custom-http-errors-56-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0-tls-custom-http-errors", "default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-per-host-custom-http-errors",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "415"},
								Service: "default-backend-default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend",
								NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-default-backend-per-host-custom-http-errors"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend-tls-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "415"},
								Service: "default-backend-default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend-tls",
								NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-default-backend-per-host-custom-http-errors"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "415"},
								Service: "default-backend-default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0",
								NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-default-backend-per-host-custom-http-errors"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0-tls-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "415"},
								Service: "default-backend-default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0-tls",
								NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-default-backend-per-host-custom-http-errors"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-per-host-custom-http-errors-56",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-per-host-custom-http-errors-56",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-backend-default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.5:8000"},
									{URL: "http://10.10.0.6:8000"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-b",
										Port:      "0",
									},
								},
							},
						},
						"default-backend-default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-default-backend-tls": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.5:8000"},
									{URL: "http://10.10.0.6:8000"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-b",
										Port:      "0",
									},
								},
							},
						},
						"default-backend-default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.5:8000"},
									{URL: "http://10.10.0.6:8000"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-b",
										Port:      "0",
									},
								},
							},
						},
						"default-backend-default-7-ingress-with-default-backend-per-host-custom-http-errors-56-rule-0-path-0-tls": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.5:8000"},
									{URL: "http://10.10.0.6:8000"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-b",
										Port:      "0",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-per-host-custom-http-errors-56": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Ingress default backend without rules inherits ssl-redirect on the catch-all http router",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-no-rules-tls.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-redirect-scheme", "default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-tls",
									},
								},
							},
						},
						"default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-tls",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-backend-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-no-rules-tls-41",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-no-rules-tls-41": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
				},
			},
		},
		{
			desc: "Ingress default backend respects backend-protocol and affinity annotations",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-annotations.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-default-backend-annotations-40-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-annotations-40-whoami-tls-10-443-3",
							Middlewares: []string{"default-7-ingress-with-default-backend-annotations-40-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-annotations",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-annotations-40-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-annotations-40-whoami-tls-10-443-3",
							Middlewares: []string{"default-7-ingress-with-default-backend-annotations-40-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-annotations",
									},
								},
							},
						},
						// The ingress-level default backend should also use the ingress annotations
						// (backend-protocol=HTTPS → https:// scheme, affinity=cookie → sticky session).
						"default-7-ingress-with-default-backend-annotations-40-rule-0-default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-annotations-40-default-backend",
							Middlewares: []string{"default-7-ingress-with-default-backend-annotations-40-rule-0-default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-annotations",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-annotations-40-rule-0-default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-annotations-40-default-backend",
							Middlewares: []string{"default-7-ingress-with-default-backend-annotations-40-rule-0-default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-annotations",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-default-backend-annotations-40-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-annotations-40-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-annotations-40-rule-0-default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-annotations-40-rule-0-default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						// Regular path backend: https:// scheme + sticky cookie from annotations.
						"default-7-ingress-with-default-backend-annotations-40-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "https://10.10.0.3:8443"},
									{URL: "https://10.10.0.4:8443"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-annotations-40",
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:               "MYSTICKYNESS",
										HTTPOnly:           true,
										Path:               new("/"),
										PreserveLeadingDot: true,
									},
								},
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-tls",
										Port:      "443",
									},
								},
							},
						},
						// Ingress-level default backend: must also use https:// scheme + sticky cookie.
						"default-7-ingress-with-default-backend-annotations-40-default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "https://10.10.0.3:8443"},
									{URL: "https://10.10.0.4:8443"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-annotations-40",
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:               "MYSTICKYNESS",
										HTTPOnly:           true,
										Path:               new("/"),
										PreserveLeadingDot: true,
									},
								},
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-tls",
										Port:      "443",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-annotations-40": {
							ServerName:         "",
							InsecureSkipVerify: true,
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Ingress per-host default backend with multiple hosts",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-multiple-hosts.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-default-backend-multiple-hosts-43-rule-0-default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `Host("one.whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-multiple-hosts-43-default-backend",
							Middlewares: []string{"default-7-ingress-with-default-backend-multiple-hosts-43-rule-0-default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-multiple-hosts",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-multiple-hosts-43-rule-0-default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("one.whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-multiple-hosts-43-default-backend",
							Middlewares: []string{"default-7-ingress-with-default-backend-multiple-hosts-43-rule-0-default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-multiple-hosts",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-multiple-hosts-43-rule-1-default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `Host("two.whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-multiple-hosts-43-default-backend",
							Middlewares: []string{"default-7-ingress-with-default-backend-multiple-hosts-43-rule-1-default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-multiple-hosts",
									},
								},
							},
						},
						"default-7-ingress-with-default-backend-multiple-hosts-43-rule-1-default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("two.whoami.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-multiple-hosts-43-default-backend",
							Middlewares: []string{"default-7-ingress-with-default-backend-multiple-hosts-43-rule-1-default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-multiple-hosts",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-default-backend-multiple-hosts-43-rule-0-default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-multiple-hosts-43-rule-0-default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-multiple-hosts-43-rule-1-default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-multiple-hosts-43-rule-1-default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-default-backend-multiple-hosts-43-default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-default-backend-multiple-hosts-43",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-multiple-hosts-43": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "WhitelistSourceRange with single IP",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-whitelist-single-ip.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-whitelist-single-ip-32-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-whitelist-single-ip-32-rule-0-path-0-allowed-source-range", "default-7-ingress-with-whitelist-single-ip-32-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-whitelist-single-ip-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-whitelist-single-ip",
									},
								},
							},
						},
						"default-7-ingress-with-whitelist-single-ip-32-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-whitelist-single-ip-32-rule-0-path-0-tls-allowed-source-range", "default-7-ingress-with-whitelist-single-ip-32-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-whitelist-single-ip-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-whitelist-single-ip",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-whitelist-single-ip-32-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.20.1"},
							},
						},
						"default-7-ingress-with-whitelist-single-ip-32-rule-0-path-0-tls-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.20.1"},
							},
						},
						"default-7-ingress-with-whitelist-single-ip-32-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-whitelist-single-ip-32-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-whitelist-single-ip-32-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-whitelist-single-ip-32",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-whitelist-single-ip-32": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "WhitelistSourceRange with single CIDR",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-whitelist-single-cidr.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-whitelist-single-cidr-34-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-whitelist-single-cidr-34-rule-0-path-0-allowed-source-range", "default-7-ingress-with-whitelist-single-cidr-34-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-whitelist-single-cidr-34-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-whitelist-single-cidr",
									},
								},
							},
						},
						"default-7-ingress-with-whitelist-single-cidr-34-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-whitelist-single-cidr-34-rule-0-path-0-tls-allowed-source-range", "default-7-ingress-with-whitelist-single-cidr-34-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-whitelist-single-cidr-34-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-whitelist-single-cidr",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-whitelist-single-cidr-34-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24"},
							},
						},
						"default-7-ingress-with-whitelist-single-cidr-34-rule-0-path-0-tls-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24"},
							},
						},
						"default-7-ingress-with-whitelist-single-cidr-34-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-whitelist-single-cidr-34-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-whitelist-single-cidr-34-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-whitelist-single-cidr-34",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-whitelist-single-cidr-34": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "WhitelistSourceRange when specified multiple IP/CIDR",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-whitelist-multiple-ip-and-cidr.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-whitelist-multiple-ip-and-cidr-43-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-whitelist-multiple-ip-and-cidr-43-rule-0-path-0-allowed-source-range", "default-7-ingress-with-whitelist-multiple-ip-and-cidr-43-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-whitelist-multiple-ip-and-cidr-43-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-whitelist-multiple-ip-and-cidr",
									},
								},
							},
						},
						"default-7-ingress-with-whitelist-multiple-ip-and-cidr-43-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-whitelist-multiple-ip-and-cidr-43-rule-0-path-0-tls-allowed-source-range", "default-7-ingress-with-whitelist-multiple-ip-and-cidr-43-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-whitelist-multiple-ip-and-cidr-43-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-whitelist-multiple-ip-and-cidr",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-whitelist-multiple-ip-and-cidr-43-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24", "10.0.0.0/8", "192.168.20.1"},
							},
						},
						"default-7-ingress-with-whitelist-multiple-ip-and-cidr-43-rule-0-path-0-tls-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24", "10.0.0.0/8", "192.168.20.1"},
							},
						},
						"default-7-ingress-with-whitelist-multiple-ip-and-cidr-43-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-whitelist-multiple-ip-and-cidr-43-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-whitelist-multiple-ip-and-cidr-43-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-whitelist-multiple-ip-and-cidr-43",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-whitelist-multiple-ip-and-cidr-43": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "WhitelistSourceRange when empty ignored",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-whitelist-empty.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-whitelist-empty-28-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-whitelist-empty-28-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-whitelist-empty-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-whitelist-empty",
									},
								},
							},
						},
						"default-7-ingress-with-whitelist-empty-28-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-whitelist-empty-28-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-whitelist-empty-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-whitelist-empty",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-whitelist-empty-28-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-whitelist-empty-28-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-whitelist-empty-28-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-whitelist-empty-28",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-whitelist-empty-28": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "AllowlistSourceRange when empty ignored",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-allowlist-empty.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-allowlist-empty-28-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-allowlist-empty-28-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-allowlist-empty-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-allowlist-empty",
									},
								},
							},
						},
						"default-7-ingress-with-allowlist-empty-28-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-allowlist-empty-28-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-allowlist-empty-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-allowlist-empty",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-allowlist-empty-28-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-allowlist-empty-28-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-allowlist-empty-28-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-allowlist-empty-28",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-allowlist-empty-28": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "AllowlistSourceRange with single IP",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-allowlist-single-ip.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-allowed-source-range", "default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-allowlist-single-ip-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-allowlist-single-ip",
									},
								},
							},
						},
						"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-tls-allowed-source-range", "default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-allowlist-single-ip-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-allowlist-single-ip",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.20.1"},
							},
						},
						"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-tls-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.20.1"},
							},
						},
						"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-allowlist-single-ip-32-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-allowlist-single-ip-32",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-allowlist-single-ip-32": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                "AllowlistSourceRange with global IPAllowListDepth",
			ipAllowListStrategy: &dynamic.IPStrategy{Depth: 2},
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-allowlist-single-ip.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-allowed-source-range", "default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-allowlist-single-ip-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-allowlist-single-ip",
									},
								},
							},
						},
						"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-tls-allowed-source-range", "default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-allowlist-single-ip-32-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-allowlist-single-ip",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.20.1"},
								IPStrategy: &dynamic.IPStrategy{
									Depth: 2,
								},
							},
						},
						"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-tls-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.20.1"},
								IPStrategy: &dynamic.IPStrategy{
									Depth: 2,
								},
							},
						},
						"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-allowlist-single-ip-32-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-allowlist-single-ip-32-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-allowlist-single-ip-32",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-allowlist-single-ip-32": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "AllowlistSourceRange with single CIDR",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-allowlist-single-cidr.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-allowlist-single-cidr-34-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-allowlist-single-cidr-34-rule-0-path-0-allowed-source-range", "default-7-ingress-with-allowlist-single-cidr-34-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-allowlist-single-cidr-34-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-allowlist-single-cidr",
									},
								},
							},
						},
						"default-7-ingress-with-allowlist-single-cidr-34-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-allowlist-single-cidr-34-rule-0-path-0-tls-allowed-source-range", "default-7-ingress-with-allowlist-single-cidr-34-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-allowlist-single-cidr-34-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-allowlist-single-cidr",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-allowlist-single-cidr-34-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24"},
							},
						},
						"default-7-ingress-with-allowlist-single-cidr-34-rule-0-path-0-tls-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24"},
							},
						},
						"default-7-ingress-with-allowlist-single-cidr-34-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-allowlist-single-cidr-34-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-allowlist-single-cidr-34-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-allowlist-single-cidr-34",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-allowlist-single-cidr-34": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "AllowlistSourceRange when specified multiple IP/CIDR",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-allowlist-multiple-ip-and-cidr.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-allowlist-multiple-ip-and-cidr-43-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-allowlist-multiple-ip-and-cidr-43-rule-0-path-0-allowed-source-range", "default-7-ingress-with-allowlist-multiple-ip-and-cidr-43-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-allowlist-multiple-ip-and-cidr-43-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-allowlist-multiple-ip-and-cidr",
									},
								},
							},
						},
						"default-7-ingress-with-allowlist-multiple-ip-and-cidr-43-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-allowlist-multiple-ip-and-cidr-43-rule-0-path-0-tls-allowed-source-range", "default-7-ingress-with-allowlist-multiple-ip-and-cidr-43-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-allowlist-multiple-ip-and-cidr-43-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-allowlist-multiple-ip-and-cidr",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-allowlist-multiple-ip-and-cidr-43-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24", "10.0.0.0/8", "192.168.20.1"},
							},
						},
						"default-7-ingress-with-allowlist-multiple-ip-and-cidr-43-rule-0-path-0-tls-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24", "10.0.0.0/8", "192.168.20.1"},
							},
						},
						"default-7-ingress-with-allowlist-multiple-ip-and-cidr-43-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-allowlist-multiple-ip-and-cidr-43-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-allowlist-multiple-ip-and-cidr-43-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-allowlist-multiple-ip-and-cidr-43",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-allowlist-multiple-ip-and-cidr-43": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Enable Access Log",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-access-log.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-access-log-enabled-31-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("accesslog-enabled.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-access-log-enabled-31-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-access-log-enabled-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								AccessLogs: new(true),
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-access-log-enabled",
									},
								},
							},
						},
						"default-7-ingress-with-access-log-enabled-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("accesslog-enabled.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-access-log-enabled-31-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-access-log-enabled-31-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								AccessLogs: new(true),
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-access-log-enabled",
									},
								},
							},
						},
						"default-7-ingress-with-access-log-disabled-32-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("accesslog-disabled.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-access-log-disabled-32-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-access-log-disabled-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								AccessLogs: new(false),
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-access-log-disabled",
									},
								},
							},
						},
						"default-7-ingress-with-access-log-disabled-32-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("accesslog-disabled.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-access-log-disabled-32-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-access-log-disabled-32-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								AccessLogs: new(false),
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-access-log-disabled",
									},
								},
							},
						},
						"default-7-ingress-with-access-log-default-31-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("accesslog-default.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-access-log-default-31-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-access-log-default-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-access-log-default",
									},
								},
							},
						},
						"default-7-ingress-with-access-log-default-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("accesslog-default.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-access-log-default-31-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-access-log-default-31-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-access-log-default",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-access-log-enabled-31-rule-0-path-0-retry":      {Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)}},
						"default-7-ingress-with-access-log-enabled-31-rule-0-path-0-tls-retry":  {Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)}},
						"default-7-ingress-with-access-log-disabled-32-rule-0-path-0-retry":     {Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)}},
						"default-7-ingress-with-access-log-disabled-32-rule-0-path-0-tls-retry": {Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)}},
						"default-7-ingress-with-access-log-default-31-rule-0-path-0-retry":      {Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)}},
						"default-7-ingress-with-access-log-default-31-rule-0-path-0-tls-retry":  {Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)}},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-access-log-enabled-31-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-access-log-enabled-31",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-access-log-disabled-32-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-access-log-disabled-32",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-access-log-default-31-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-access-log-default-31",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-access-log-enabled-31": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-access-log-disabled-32": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-access-log-default-31": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Permanent Redirect",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-permanent-redirect.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("permanent-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-permanent-redirect-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-permanent-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-redirect", "default-7-ingress-with-permanent-redirect-31-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("permanent-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-permanent-redirect-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-permanent-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls-redirect", "default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusMovedPermanently),
							},
						},
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusMovedPermanently),
							},
						},
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-permanent-redirect-31-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-permanent-redirect-31",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-permanent-redirect-31": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Permanent Redirect Code - wrong code",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-permanent-redirect-code-wrong-code.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("permanent-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-permanent-redirect-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-permanent-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-redirect", "default-7-ingress-with-permanent-redirect-31-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("permanent-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-permanent-redirect-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-permanent-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls-redirect", "default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusMovedPermanently),
							},
						},
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusMovedPermanently),
							},
						},
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-permanent-redirect-31-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-permanent-redirect-31",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-permanent-redirect-31": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Permanent Redirect Code - correct code",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-permanent-redirect-code-correct-code.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("permanent-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-permanent-redirect-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-permanent-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-redirect", "default-7-ingress-with-permanent-redirect-31-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("permanent-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-permanent-redirect-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-permanent-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls-redirect", "default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusMultipleChoices),
							},
						},
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusMultipleChoices),
							},
						},
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-permanent-redirect-31-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-permanent-redirect-31-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-permanent-redirect-31",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-permanent-redirect-31": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Temporal Redirect takes precedence over Permanent Redirect",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-temporal-and-permanent-redirect.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-redirect-21-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-redirect-21-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-redirect-21-rule-0-path-0-redirect", "default-7-ingress-with-redirect-21-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-redirect-21-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-redirect-21-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-redirect-21-rule-0-path-0-tls-redirect", "default-7-ingress-with-redirect-21-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-redirect-21-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusFound),
							},
						},
						"default-7-ingress-with-redirect-21-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusFound),
							},
						},
						"default-7-ingress-with-redirect-21-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-redirect-21-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-redirect-21-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-redirect-21",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-redirect-21": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Temporal Redirect",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-temporal-redirect.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("temporal-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-temporal-redirect-30-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-temporal-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-redirect", "default-7-ingress-with-temporal-redirect-30-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("temporal-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-temporal-redirect-30-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-temporal-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls-redirect", "default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusFound),
							},
						},
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusFound),
							},
						},
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-temporal-redirect-30-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-temporal-redirect-30",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-temporal-redirect-30": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Temporal Redirect Code - wrong code",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-temporal-redirect-code-wrong-code.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("temporal-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-temporal-redirect-30-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-temporal-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-redirect", "default-7-ingress-with-temporal-redirect-30-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("temporal-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-temporal-redirect-30-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-temporal-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls-redirect", "default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusFound),
							},
						},
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusFound),
							},
						},
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-temporal-redirect-30-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-temporal-redirect-30",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-temporal-redirect-30": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Temporal Redirect Code - correct code",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-temporal-redirect-code-correct-code.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("temporal-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-temporal-redirect-30-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-temporal-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-redirect", "default-7-ingress-with-temporal-redirect-30-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("temporal-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-temporal-redirect-30-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-temporal-redirect",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls-redirect", "default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusPermanentRedirect),
							},
						},
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  new(http.StatusPermanentRedirect),
							},
						},
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-temporal-redirect-30-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-temporal-redirect-30-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-temporal-redirect-30",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-temporal-redirect-30": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Permanent Redirect with Use Regex",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-redirect-use-regex.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-redirect-use-regex-31-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("redirect-regex.localhost") && PathRegexp("(?i)^/(example)(/.*)")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-redirect-use-regex-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-redirect-use-regex",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-redirect-use-regex-31-rule-0-path-0-redirect", "default-7-ingress-with-redirect-use-regex-31-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-redirect-use-regex-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("redirect-regex.localhost") && PathRegexp("(?i)^/(example)(/.*)")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-redirect-use-regex-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-redirect-use-regex",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-redirect-use-regex-31-rule-0-path-0-tls-redirect", "default-7-ingress-with-redirect-use-regex-31-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-redirect-use-regex-31-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^https?://[^/]+/(example)(/.*)`,
								Replacement: "https://redirect.example.com/$1$2",
								StatusCode:  new(http.StatusMovedPermanently),
							},
						},
						"default-7-ingress-with-redirect-use-regex-31-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^https?://[^/]+/(example)(/.*)`,
								Replacement: "https://redirect.example.com/$1$2",
								StatusCode:  new(http.StatusMovedPermanently),
							},
						},
						"default-7-ingress-with-redirect-use-regex-31-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-redirect-use-regex-31-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-redirect-use-regex-31-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-redirect-use-regex-31",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-redirect-use-regex-31": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Proxy connect timeout",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-timeout.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-timeout-26-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-proxy-timeout-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-timeout",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-proxy-timeout-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-timeout",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-timeout-26-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-timeout-26",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-timeout-26": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(30 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Proxy read timeout",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-read-timeout.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-timeout-26-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-proxy-timeout-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-timeout",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-proxy-timeout-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-timeout",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-timeout-26-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-timeout-26",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-timeout-26": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(30 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Proxy send timeout",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-send-timeout.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-timeout-26-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-proxy-timeout-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-timeout",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-proxy-timeout-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-timeout",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-timeout-26-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-timeout-26-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-timeout-26",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-timeout-26": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(30 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Auth TLS secret",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-auth-tls-secret.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-auth-tls-secret-28-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("auth-tls-secret.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-auth-tls-secret-28-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-auth-tls-secret-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-secret",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-7-ca-secret-9-require-and-verify-client-cert",
							},
						},
						"default-7-ingress-with-auth-tls-secret-28-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("auth-tls-secret.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-ingress-with-auth-tls-secret-28-rule-0-path-0-redirect-scheme",
								"default-7-ingress-with-auth-tls-secret-28-rule-0-path-0-retry",
							},
							Service: "default-7-ingress-with-auth-tls-secret-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-secret",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-auth-tls-secret-28-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-auth-tls-secret-28-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-auth-tls-secret-28-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-auth-tls-secret-28-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-auth-tls-secret-28",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-auth-tls-secret-28": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
					Options: map[string]tls.Options{
						"default-7-ca-secret-9-require-and-verify-client-cert": {
							ClientAuth: tls.ClientAuth{
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----"},
								ClientAuthType: tls.RequireAndVerifyClientCert,
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
							ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
						},
					},
				},
			},
		},
		{
			// Regression test for a TLS option name collision: with only the secret name
			// length-prefixed, ns=a/secret=1-b and ns=a-3/secret=b concatenate to the identical
			// pre-Normalize string ("a-3-1-b-require-and-verify-client-cert" both ways), so two
			// unrelated Ingresses in different namespaces would resolve to the same cached TLS
			// option and silently share one another's CA. Both the namespace and the secret name
			// must be length-prefixed to keep the two distinct.
			desc: "Auth TLS secret namespace collision",
			paths: []string{
				"secrets-tls-option-namespace-collision.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-auth-tls-secret-namespace-collision.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"a-1-ingress-with-auth-tls-secret-namespace-collision-victim-55-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("tls-option-collision-victim.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"a-1-ingress-with-auth-tls-secret-namespace-collision-victim-55-rule-0-path-0-tls-retry"},
							Service:     "a-1-ingress-with-auth-tls-secret-namespace-collision-victim-55-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "a",
										Name:      "ingress-with-auth-tls-secret-namespace-collision-victim",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "a-1-1-b-3-require-and-verify-client-cert",
							},
						},
						"a-1-ingress-with-auth-tls-secret-namespace-collision-victim-55-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("tls-option-collision-victim.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"a-1-ingress-with-auth-tls-secret-namespace-collision-victim-55-rule-0-path-0-retry"},
							Service:     "a-1-ingress-with-auth-tls-secret-namespace-collision-victim-55-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "a",
										Name:      "ingress-with-auth-tls-secret-namespace-collision-victim",
									},
								},
							},
						},
						"a-3-3-ingress-with-auth-tls-secret-namespace-collision-attacker-57-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("tls-option-collision-attacker.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"a-3-3-ingress-with-auth-tls-secret-namespace-collision-attacker-57-rule-0-path-0-tls-retry"},
							Service:     "a-3-3-ingress-with-auth-tls-secret-namespace-collision-attacker-57-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "a-3",
										Name:      "ingress-with-auth-tls-secret-namespace-collision-attacker",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "a-3-3-b-1-require-and-verify-client-cert",
							},
						},
						"a-3-3-ingress-with-auth-tls-secret-namespace-collision-attacker-57-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("tls-option-collision-attacker.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"a-3-3-ingress-with-auth-tls-secret-namespace-collision-attacker-57-rule-0-path-0-retry"},
							Service:     "a-3-3-ingress-with-auth-tls-secret-namespace-collision-attacker-57-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "a-3",
										Name:      "ingress-with-auth-tls-secret-namespace-collision-attacker",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"a-1-ingress-with-auth-tls-secret-namespace-collision-victim-55-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"a-1-ingress-with-auth-tls-secret-namespace-collision-victim-55-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"a-3-3-ingress-with-auth-tls-secret-namespace-collision-attacker-57-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"a-3-3-ingress-with-auth-tls-secret-namespace-collision-attacker-57-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"a-1-ingress-with-auth-tls-secret-namespace-collision-victim-55-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "a",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"a-3-3-ingress-with-auth-tls-secret-namespace-collision-attacker-57-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "a-3",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"a-1-ingress-with-auth-tls-secret-namespace-collision-victim-55": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"a-3-3-ingress-with-auth-tls-secret-namespace-collision-attacker-57": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"a-1-1-b-3-require-and-verify-client-cert": {
							ClientAuth: tls.ClientAuth{
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----VICTIM-----END CERTIFICATE-----"},
								ClientAuthType: tls.RequireAndVerifyClientCert,
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
							ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
						},
						"a-3-3-b-1-require-and-verify-client-cert": {
							ClientAuth: tls.ClientAuth{
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----ATTACKER-----END CERTIFICATE-----"},
								ClientAuthType: tls.RequireAndVerifyClientCert,
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
							ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
						},
					},
				},
			},
		},
		{
			desc: "Auth TLS secret cross-namespace reference is rejected even when the TLS option is cached",
			paths: []string{
				"secrets-tls-option-cross-namespace.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-auth-tls-secret-cross-namespace.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"a-1-ingress-with-auth-tls-secret-cross-namespace-owner-50-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("tls-option-cross-namespace-owner.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"a-1-ingress-with-auth-tls-secret-cross-namespace-owner-50-rule-0-path-0-tls-retry"},
							Service:     "a-1-ingress-with-auth-tls-secret-cross-namespace-owner-50-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "a",
										Name:      "ingress-with-auth-tls-secret-cross-namespace-owner",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "a-1-ca-2-require-and-verify-client-cert",
							},
						},
						"a-1-ingress-with-auth-tls-secret-cross-namespace-owner-50-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("tls-option-cross-namespace-owner.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"a-1-ingress-with-auth-tls-secret-cross-namespace-owner-50-rule-0-path-0-retry"},
							Service:     "a-1-ingress-with-auth-tls-secret-cross-namespace-owner-50-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "a",
										Name:      "ingress-with-auth-tls-secret-cross-namespace-owner",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"a-1-ingress-with-auth-tls-secret-cross-namespace-owner-50-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"a-1-ingress-with-auth-tls-secret-cross-namespace-owner-50-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"a-1-ingress-with-auth-tls-secret-cross-namespace-owner-50-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "a",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"a-1-ingress-with-auth-tls-secret-cross-namespace-owner-50": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"a-1-ca-2-require-and-verify-client-cert": {
							ClientAuth: tls.ClientAuth{
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----OWNER-----END CERTIFICATE-----"},
								ClientAuthType: tls.RequireAndVerifyClientCert,
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
							ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
						},
					},
				},
			},
		},
		{
			// Regression test for a TLS option name collision: a secret name may contain dots, and
			// normalizing the name turns my.secret into my-secret without changing its length, so
			// the length suffixes alone cannot keep the two secrets apart. Two Ingresses referencing
			// them would resolve to the same cached TLS option and share one another's CA.
			desc: "Auth TLS secret name collision",
			paths: []string{
				"services.yml",
				"secrets-tls-option-name-collision.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-auth-tls-secret-name-collision.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-auth-tls-secret-name-collision-dash-48-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("tls-option-name-collision-dash.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-auth-tls-secret-name-collision-dash-48-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-auth-tls-secret-name-collision-dash-48-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-secret-name-collision-dash",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-7-my-secret-9-require-and-verify-client-cert",
							},
						},
						"default-7-ingress-with-auth-tls-secret-name-collision-dash-48-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("tls-option-name-collision-dash.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-auth-tls-secret-name-collision-dash-48-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-auth-tls-secret-name-collision-dash-48-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-secret-name-collision-dash",
									},
								},
							},
						},
						"default-7-ingress-with-auth-tls-secret-name-collision-dot-47-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("tls-option-name-collision-dot.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-auth-tls-secret-name-collision-dot-47-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-auth-tls-secret-name-collision-dot-47-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-secret-name-collision-dot",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-7-my.secret-9-require-and-verify-client-cert",
							},
						},
						"default-7-ingress-with-auth-tls-secret-name-collision-dot-47-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("tls-option-name-collision-dot.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-auth-tls-secret-name-collision-dot-47-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-auth-tls-secret-name-collision-dot-47-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-secret-name-collision-dot",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-auth-tls-secret-name-collision-dash-48-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-auth-tls-secret-name-collision-dash-48-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-auth-tls-secret-name-collision-dot-47-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-auth-tls-secret-name-collision-dot-47-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-auth-tls-secret-name-collision-dash-48-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-auth-tls-secret-name-collision-dash-48",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-auth-tls-secret-name-collision-dot-47-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-auth-tls-secret-name-collision-dot-47",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-auth-tls-secret-name-collision-dash-48": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-auth-tls-secret-name-collision-dot-47": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{
						"default-7-my-secret-9-require-and-verify-client-cert": {
							ClientAuth: tls.ClientAuth{
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----DASH-----END CERTIFICATE-----"},
								ClientAuthType: tls.RequireAndVerifyClientCert,
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
							ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
						},
						"default-7-my.secret-9-require-and-verify-client-cert": {
							ClientAuth: tls.ClientAuth{
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----DOT-----END CERTIFICATE-----"},
								ClientAuthType: tls.RequireAndVerifyClientCert,
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
							ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
						},
					},
				},
			},
		},
		{
			desc: "Auth TLS verify client",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-auth-tls-verify-client.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-auth-tls-verify-client-35-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("auth-tls-verify-client.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-auth-tls-verify-client-35-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-auth-tls-verify-client-35-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-verify-client",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-7-ca-secret-9-verify-client-cert-if-given",
							},
						},
						"default-7-ingress-with-auth-tls-verify-client-35-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("auth-tls-verify-client.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-ingress-with-auth-tls-verify-client-35-rule-0-path-0-redirect-scheme",
								"default-7-ingress-with-auth-tls-verify-client-35-rule-0-path-0-retry",
							},
							Service: "default-7-ingress-with-auth-tls-verify-client-35-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-verify-client",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-auth-tls-verify-client-35-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-auth-tls-verify-client-35-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-auth-tls-verify-client-35-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-auth-tls-verify-client-35-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-auth-tls-verify-client-35",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-auth-tls-verify-client-35": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
					Options: map[string]tls.Options{
						"default-7-ca-secret-9-verify-client-cert-if-given": {
							ClientAuth: tls.ClientAuth{
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----"},
								ClientAuthType: tls.VerifyClientCertIfGiven,
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
							ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
						},
					},
				},
			},
		},
		{
			desc: "Auth TLS secret shared across Ingresses on the same host",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-auth-tls-secret-shared-host.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-protected-ingress-17-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("mtls-conflict.localhost") && Path("/protected")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-protected-ingress-17-rule-0-path-0-tls-retry"},
							Service:     "default-7-protected-ingress-17-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "protected-ingress",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-7-ca-secret-9-require-and-verify-client-cert",
							},
						},
						"default-7-protected-ingress-17-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("mtls-conflict.localhost") && Path("/protected")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-protected-ingress-17-rule-0-path-0-redirect-scheme",
								"default-7-protected-ingress-17-rule-0-path-0-retry",
							},
							Service: "default-7-protected-ingress-17-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "protected-ingress",
									},
								},
							},
						},
						"default-7-same-host-second-ingress-24-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("mtls-conflict.localhost") && Path("/second")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-same-host-second-ingress-24-rule-0-path-0-tls-retry"},
							Service:     "default-7-same-host-second-ingress-24-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "same-host-second-ingress",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-7-ca-secret-9-require-and-verify-client-cert",
							},
						},
						"default-7-same-host-second-ingress-24-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("mtls-conflict.localhost") && Path("/second")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-same-host-second-ingress-24-rule-0-path-0-redirect-scheme",
								"default-7-same-host-second-ingress-24-rule-0-path-0-retry",
							},
							Service: "default-7-same-host-second-ingress-24-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "same-host-second-ingress",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-protected-ingress-17-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-protected-ingress-17-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-protected-ingress-17-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-same-host-second-ingress-24-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-same-host-second-ingress-24-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-same-host-second-ingress-24-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-protected-ingress-17-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-protected-ingress-17",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-same-host-second-ingress-24-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-same-host-second-ingress-24",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-protected-ingress-17": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-same-host-second-ingress-24": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
					Options: map[string]tls.Options{
						"default-7-ca-secret-9-require-and-verify-client-cert": {
							ClientAuth: tls.ClientAuth{
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----"},
								ClientAuthType: tls.RequireAndVerifyClientCert,
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
							ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
						},
					},
				},
			},
		},
		{
			desc: "Auth TLS secret applied to every Ingress on the same host",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-auth-tls-secret-mixed-host.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-public-ingress-14-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("mtls-mixed.localhost") && Path("/public")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-public-ingress-14-rule-0-path-0-tls-retry"},
							Service:     "default-7-public-ingress-14-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "public-ingress",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-7-ca-secret-9-require-and-verify-client-cert",
							},
						},
						"default-7-public-ingress-14-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("mtls-mixed.localhost") && Path("/public")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-public-ingress-14-rule-0-path-0-retry"},
							Service:     "default-7-public-ingress-14-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "public-ingress",
									},
								},
							},
						},
						"default-7-protected-ingress-17-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("mtls-mixed.localhost") && Path("/protected")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-protected-ingress-17-rule-0-path-0-tls-retry"},
							Service:     "default-7-protected-ingress-17-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "protected-ingress",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-7-ca-secret-9-require-and-verify-client-cert",
							},
						},
						"default-7-protected-ingress-17-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("mtls-mixed.localhost") && Path("/protected")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-protected-ingress-17-rule-0-path-0-redirect-scheme",
								"default-7-protected-ingress-17-rule-0-path-0-retry",
							},
							Service: "default-7-protected-ingress-17-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "protected-ingress",
									},
								},
							},
						},
						"default-7-optional-ingress-16-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("mtls-mixed.localhost") && Path("/optional")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-optional-ingress-16-rule-0-path-0-tls-retry"},
							Service:     "default-7-optional-ingress-16-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "optional-ingress",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-7-ca-secret-9-require-and-verify-client-cert",
							},
						},
						"default-7-optional-ingress-16-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("mtls-mixed.localhost") && Path("/optional")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-optional-ingress-16-rule-0-path-0-retry"},
							Service:     "default-7-optional-ingress-16-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "optional-ingress",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-public-ingress-14-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-public-ingress-14-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-protected-ingress-17-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-protected-ingress-17-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-protected-ingress-17-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-optional-ingress-16-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-optional-ingress-16-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-public-ingress-14-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-public-ingress-14",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-protected-ingress-17-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-protected-ingress-17",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-optional-ingress-16-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-optional-ingress-16",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-public-ingress-14": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-protected-ingress-17": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-optional-ingress-16": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
					Options: map[string]tls.Options{
						"default-7-ca-secret-9-require-and-verify-client-cert": {
							ClientAuth: tls.ClientAuth{
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----"},
								ClientAuthType: tls.RequireAndVerifyClientCert,
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
							ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
						},
					},
				},
			},
		},
		{
			desc: "Custom HTTP Errors and Default backend annotation",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-custom-http-errors-and-default-backend.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-custom-http-errors-and-default-backend-51-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-http-errors-and-default-backend",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0-custom-http-errors", "default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-custom-http-errors-and-default-backend-51-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-http-errors-and-default-backend",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0-tls-custom-http-errors", "default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "415"},
								Service: "default-backend-default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0",
								NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-custom-http-errors-and-default-backend"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0-tls-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "415"},
								Service: "default-backend-default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0-tls", NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-custom-http-errors-and-default-backend"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-custom-http-errors-and-default-backend-51-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-custom-http-errors-and-default-backend-51",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-backend-default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.5:8000",
									},
									{
										URL: "http://10.10.0.6:8000",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-b",
										Port:      "0",
									},
								},
							},
						},
						"default-backend-default-7-ingress-with-custom-http-errors-and-default-backend-51-rule-0-path-0-tls": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.5:8000"},
									{URL: "http://10.10.0.6:8000"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-b",
										Port:      "0",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-custom-http-errors-and-default-backend-51": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                           "Custom HTTP Errors",
			defaultBackendServiceName:      "whoami-b",
			defaultBackendServiceNamespace: "default",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-custom-http-errors.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-custom-http-errors-31-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-custom-http-errors-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-http-errors",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-custom-http-errors",
								"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-retry",
							},
						},
						"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-custom-http-errors-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-http-errors",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-tls-custom-http-errors",
								"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
									},
								},
							},
						},
						"default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							TLS:         &dynamic.RouterTLSConfig{},
							Service:     "default-backend",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "415"},
								Service: "default-backend",
								NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-custom-http-errors"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-tls-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "415"},
								Service: "default-backend",
								NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-custom-http-errors"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-custom-http-errors-31-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-custom-http-errors-31",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.5:8000",
									},
									{
										URL: "http://10.10.0.6:8000",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-b",
										Port:      "0",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-custom-http-errors-31": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Custom HTTP Errors without default backend",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-custom-http-errors.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-custom-http-errors-31-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-custom-http-errors-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-http-errors",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-custom-http-errors-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-http-errors",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-custom-http-errors-31-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-custom-http-errors-31-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-custom-http-errors-31",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-custom-http-errors-31": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Default backend annotation",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-annotation.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-default-backend-annotation-39-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-annotation-39-empty-5-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-annotation",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-default-backend-annotation-39-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-default-backend-annotation-39-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-default-backend-annotation-39-empty-5-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-annotation",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-default-backend-annotation-39-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-default-backend-annotation-39-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-default-backend-annotation-39-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-default-backend-annotation-39-empty-5-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.5:8000",
									},
									{
										URL: "http://10.10.0.6:8000",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-default-backend-annotation-39",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "empty",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-annotation-39": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Buffering with proxy body size of 10MB",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-body-size.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-body-size-28-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-body-size-28-rule-0-path-0-buffering", "default-7-ingress-with-proxy-body-size-28-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-proxy-body-size-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-body-size",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-body-size-28-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-body-size-28-rule-0-path-0-tls-buffering", "default-7-ingress-with-proxy-body-size-28-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-proxy-body-size-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-body-size",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-body-size-28-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								MaxRequestBodyBytes:   10 * 1024 * 1024,
								MemRequestBodyBytes:   defaultClientBodyBufferSize,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-7-ingress-with-proxy-body-size-28-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								MaxRequestBodyBytes:   10 * 1024 * 1024,
								MemRequestBodyBytes:   defaultClientBodyBufferSize,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-7-ingress-with-proxy-body-size-28-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(int64(10 * 1024 * 1024)),
							},
						},
						"default-7-ingress-with-proxy-body-size-28-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(int64(10 * 1024 * 1024)),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-body-size-28-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-body-size-28",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-body-size-28": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                  "Proxy request buffering off keeps network retry when no status codes are set",
			proxyRequestBuffering: true,
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-request-buffering-off.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-request-buffering-off-40-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-request-buffering-off-40-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-request-buffering-off",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-proxy-request-buffering-off-40-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-proxy-request-buffering-off-40-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-request-buffering-off-40-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-request-buffering-off",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-proxy-request-buffering-off-40-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-request-buffering-off-40-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-request-buffering-off-40-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-request-buffering-off-40-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-request-buffering-off-40",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-request-buffering-off-40": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Proxy body size influences retry max request body bytes",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-body-size-retry.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-body-size-retry-34-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-body-size-retry-34-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-proxy-body-size-retry-34-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-body-size-retry",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-body-size-retry-34-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-body-size-retry-34-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-proxy-body-size-retry-34-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-body-size-retry",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-body-size-retry-34-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(int64(5 * 1024 * 1024)),
							},
						},
						"default-7-ingress-with-proxy-body-size-retry-34-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(int64(5 * 1024 * 1024)),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-body-size-retry-34-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-body-size-retry-34",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-body-size-retry-34": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Proxy body size zero influences retry max request body bytes",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-body-size-zero-retry.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-body-size-zero-retry-39-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-body-size-zero-retry-39-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-proxy-body-size-zero-retry-39-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-body-size-zero-retry",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-body-size-zero-retry-39-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-body-size-zero-retry-39-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-proxy-body-size-zero-retry-39-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-body-size-zero-retry",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-body-size-zero-retry-39-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(int64(-1)),
							},
						},
						"default-7-ingress-with-proxy-body-size-zero-retry-39-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(int64(-1)),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-body-size-zero-retry-39-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-body-size-zero-retry-39",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-body-size-zero-retry-39": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "HTTP status retry disabled when request buffering off",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-http-status-retry-and-buffering-off.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-http-status-retry-and-buffering-off-48-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-http-status-retry-and-buffering-off-48-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-http-status-retry-and-buffering-off-48-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-http-status-retry-and-buffering-off",
									},
								},
							},
						},
						"default-7-ingress-with-http-status-retry-and-buffering-off-48-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-http-status-retry-and-buffering-off-48-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-http-status-retry-and-buffering-off-48-rule-0-path-0-tls-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-http-status-retry-and-buffering-off",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-http-status-retry-and-buffering-off-48-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								Status:              []string{"502"},
								MaxRequestBodyBytes: new(int64(0)),
							},
						},
						"default-7-ingress-with-http-status-retry-and-buffering-off-48-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								Status:              []string{"502"},
								MaxRequestBodyBytes: new(int64(0)),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-http-status-retry-and-buffering-off-48-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-http-status-retry-and-buffering-off-48",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-http-status-retry-and-buffering-off-48": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Buffering with client body buffer size of 10MB",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-client-body-buffer-size.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-client-body-buffer-size-36-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-client-body-buffer-size-36-rule-0-path-0-buffering", "default-7-ingress-with-client-body-buffer-size-36-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-client-body-buffer-size-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-client-body-buffer-size",
									},
								},
							},
						},
						"default-7-ingress-with-client-body-buffer-size-36-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-client-body-buffer-size-36-rule-0-path-0-tls-buffering", "default-7-ingress-with-client-body-buffer-size-36-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-client-body-buffer-size-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-client-body-buffer-size",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-client-body-buffer-size-36-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								MemRequestBodyBytes:   10 * 1024 * 1024,
								MaxRequestBodyBytes:   defaultProxyBodySize,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-7-ingress-with-client-body-buffer-size-36-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								MemRequestBodyBytes:   10 * 1024 * 1024,
								MaxRequestBodyBytes:   defaultProxyBodySize,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-7-ingress-with-client-body-buffer-size-36-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-client-body-buffer-size-36-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-client-body-buffer-size-36-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-client-body-buffer-size-36",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-client-body-buffer-size-36": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Buffering with proxy body size and client body buffer",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-body-size-and-client-body-buffer-size.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56-rule-0-path-0-buffering", "default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-body-size-and-client-body-buffer-size",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56-rule-0-path-0-tls-buffering", "default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-body-size-and-client-body-buffer-size",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								MaxRequestBodyBytes:   10 * 1024 * 1024,
								MemRequestBodyBytes:   10 * 1024,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								MaxRequestBodyBytes:   10 * 1024 * 1024,
								MemRequestBodyBytes:   10 * 1024,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(int64(10 * 1024 * 1024)),
							},
						},
						"default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(int64(10 * 1024 * 1024)),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-body-size-and-client-body-buffer-size-56": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Buffering with proxy buffer size",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-buffer-size.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-buffer-size-30-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-buffer-size-30-rule-0-path-0-buffering", "default-7-ingress-with-proxy-buffer-size-30-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-proxy-buffer-size-30-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-buffer-size",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-buffer-size-30-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-buffer-size-30-rule-0-path-0-tls-buffering", "default-7-ingress-with-proxy-buffer-size-30-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-proxy-buffer-size-30-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-buffer-size",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-buffer-size-30-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: 16 * 1024 * int64(defaultProxyBuffersNumber),
								MaxResponseBodyBytes: defaultProxyMaxTempFileSize + (defaultProxyBufferSize * 8),
							},
						},
						"default-7-ingress-with-proxy-buffer-size-30-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: 16 * 1024 * int64(defaultProxyBuffersNumber),
								MaxResponseBodyBytes: defaultProxyMaxTempFileSize + (defaultProxyBufferSize * 8),
							},
						},
						"default-7-ingress-with-proxy-buffer-size-30-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-buffer-size-30-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-buffer-size-30-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-buffer-size-30",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-buffer-size-30": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Buffering with proxy buffers number",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-buffers-number.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-buffers-number-33-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-buffers-number-33-rule-0-path-0-buffering", "default-7-ingress-with-proxy-buffers-number-33-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-proxy-buffers-number-33-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-buffers-number",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-buffers-number-33-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-buffers-number-33-rule-0-path-0-tls-buffering", "default-7-ingress-with-proxy-buffers-number-33-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-proxy-buffers-number-33-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-buffers-number",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-buffers-number-33-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: defaultProxyBufferSize * 8,
								MaxResponseBodyBytes: defaultProxyMaxTempFileSize + (defaultProxyBufferSize * 8),
							},
						},
						"default-7-ingress-with-proxy-buffers-number-33-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: defaultProxyBufferSize * 8,
								MaxResponseBodyBytes: defaultProxyMaxTempFileSize + (defaultProxyBufferSize * 8),
							},
						},
						"default-7-ingress-with-proxy-buffers-number-33-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-buffers-number-33-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-buffers-number-33-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-buffers-number-33",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-buffers-number-33": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Buffering with proxy buffer size and proxy buffers number",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-buffer-size-and-number.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-buffer-size-and-number-41-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-buffer-size-and-number-41-rule-0-path-0-buffering", "default-7-ingress-with-proxy-buffer-size-and-number-41-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-proxy-buffer-size-and-number-41-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-buffer-size-and-number",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-buffer-size-and-number-41-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-buffer-size-and-number-41-rule-0-path-0-tls-buffering", "default-7-ingress-with-proxy-buffer-size-and-number-41-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-proxy-buffer-size-and-number-41-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-buffer-size-and-number",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-buffer-size-and-number-41-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: 16 * 1024 * 8,
								MaxResponseBodyBytes: defaultProxyMaxTempFileSize + (16 * 1024 * 8),
							},
						},
						"default-7-ingress-with-proxy-buffer-size-and-number-41-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: 16 * 1024 * 8,
								MaxResponseBodyBytes: defaultProxyMaxTempFileSize + (16 * 1024 * 8),
							},
						},
						"default-7-ingress-with-proxy-buffer-size-and-number-41-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-buffer-size-and-number-41-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-buffer-size-and-number-41-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-buffer-size-and-number-41",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-buffer-size-and-number-41": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Buffering with proxy max temp file size",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-max-temp-file-size.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-max-temp-file-size-37-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-max-temp-file-size-37-rule-0-path-0-buffering", "default-7-ingress-with-proxy-max-temp-file-size-37-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-proxy-max-temp-file-size-37-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-max-temp-file-size",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-max-temp-file-size-37-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-proxy-max-temp-file-size-37-rule-0-path-0-tls-buffering", "default-7-ingress-with-proxy-max-temp-file-size-37-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-proxy-max-temp-file-size-37-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-max-temp-file-size",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-max-temp-file-size-37-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								MaxResponseBodyBytes: (defaultProxyBufferSize * int64(defaultProxyBuffersNumber)) + (100 * 1024 * 1024),
							},
						},
						"default-7-ingress-with-proxy-max-temp-file-size-37-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								MaxResponseBodyBytes: (defaultProxyBufferSize * int64(defaultProxyBuffersNumber)) + (100 * 1024 * 1024),
							},
						},
						"default-7-ingress-with-proxy-max-temp-file-size-37-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-max-temp-file-size-37-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-max-temp-file-size-37-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-max-temp-file-size-37",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-max-temp-file-size-37": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                    "Server snippet with allowSnippetAnnotations enabled",
			allowSnippetAnnotations: true,
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-server-snippet.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-server-snippet-27-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-server-snippet-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-server-snippet",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-server-snippet-27-rule-0-path-0-snippet",
								"default-7-ingress-with-server-snippet-27-rule-0-path-0-retry",
							},
						},
						"default-7-ingress-with-server-snippet-27-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-server-snippet-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-server-snippet",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-server-snippet-27-rule-0-path-0-tls-snippet",
								"default-7-ingress-with-server-snippet-27-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-server-snippet-27-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet: "add_header X-Server-Snippet \"server-value\";\n",
							},
						},
						"default-7-ingress-with-server-snippet-27-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet: "add_header X-Server-Snippet \"server-value\";\n",
							},
						},
						"default-7-ingress-with-server-snippet-27-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-server-snippet-27-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-server-snippet-27-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-server-snippet-27",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-server-snippet-27": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                    "Configuration snippet with allowSnippetAnnotations enabled",
			allowSnippetAnnotations: true,
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-configuration-snippet.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-configuration-snippet-34-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-configuration-snippet-34-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-configuration-snippet",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-snippet",
								"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-retry",
							},
						},
						"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-configuration-snippet-34-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-configuration-snippet",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-tls-snippet",
								"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-configuration-snippet-34-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-configuration-snippet-34",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-configuration-snippet-34": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                    "Both snippets with allowSnippetAnnotations enabled",
			allowSnippetAnnotations: true,
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-both-snippets.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-both-snippets-26-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-both-snippets-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-both-snippets",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-both-snippets-26-rule-0-path-0-snippet",
								"default-7-ingress-with-both-snippets-26-rule-0-path-0-retry",
							},
						},
						"default-7-ingress-with-both-snippets-26-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-both-snippets-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-both-snippets",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-both-snippets-26-rule-0-path-0-tls-snippet",
								"default-7-ingress-with-both-snippets-26-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-both-snippets-26-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet:        "add_header X-Server-Snippet \"server-value\";\n",
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-7-ingress-with-both-snippets-26-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet:        "add_header X-Server-Snippet \"server-value\";\n",
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-7-ingress-with-both-snippets-26-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-both-snippets-26-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-both-snippets-26-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-both-snippets-26",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-both-snippets-26": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                    "Server snippet with allowSnippetAnnotations disabled",
			allowSnippetAnnotations: false,
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-server-snippet.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
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
			desc:                    "Configuration snippet with allowSnippetAnnotations disabled",
			allowSnippetAnnotations: false,
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-configuration-snippet.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
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
			desc:                    "Both snippets with allowSnippetAnnotations disabled",
			allowSnippetAnnotations: false,
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-both-snippets.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
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
			desc:                    "Server snippet with allowSnippetAnnotations enabled",
			allowSnippetAnnotations: true,
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-server-snippet.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-server-snippet-27-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-server-snippet-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-server-snippet",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-server-snippet-27-rule-0-path-0-snippet", "default-7-ingress-with-server-snippet-27-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-server-snippet-27-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-server-snippet-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-server-snippet",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-server-snippet-27-rule-0-path-0-tls-snippet", "default-7-ingress-with-server-snippet-27-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-server-snippet-27-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet: "add_header X-Server-Snippet \"server-value\";\n",
							},
						},
						"default-7-ingress-with-server-snippet-27-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet: "add_header X-Server-Snippet \"server-value\";\n",
							},
						},
						"default-7-ingress-with-server-snippet-27-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-server-snippet-27-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-server-snippet-27-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-server-snippet-27",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-server-snippet-27": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                    "Configuration snippet with allowSnippetAnnotations enabled",
			allowSnippetAnnotations: true,
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-configuration-snippet.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-configuration-snippet-34-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-configuration-snippet-34-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-configuration-snippet",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-snippet", "default-7-ingress-with-configuration-snippet-34-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-configuration-snippet-34-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-configuration-snippet",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-tls-snippet", "default-7-ingress-with-configuration-snippet-34-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-configuration-snippet-34-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-configuration-snippet-34-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-configuration-snippet-34",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-configuration-snippet-34": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                    "Both snippets with allowSnippetAnnotations enabled",
			allowSnippetAnnotations: true,
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-both-snippets.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-both-snippets-26-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-both-snippets-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-both-snippets",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-both-snippets-26-rule-0-path-0-snippet", "default-7-ingress-with-both-snippets-26-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-both-snippets-26-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-both-snippets-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-both-snippets",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-both-snippets-26-rule-0-path-0-tls-snippet", "default-7-ingress-with-both-snippets-26-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-both-snippets-26-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet:        "add_header X-Server-Snippet \"server-value\";\n",
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-7-ingress-with-both-snippets-26-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet:        "add_header X-Server-Snippet \"server-value\";\n",
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-7-ingress-with-both-snippets-26-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-both-snippets-26-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-both-snippets-26-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-both-snippets-26",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-both-snippets-26": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Auth TLS pass certificate to upstream",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-auth-tls-pass-certificate-to-upstream.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("auth-tls-pass-certificate-to-upstream.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-rule-0-path-0-tls-pass-certificate-to-upstream", "default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-pass-certificate-to-upstream",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-7-ca-secret-9-require-and-verify-client-cert",
							},
						},
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("auth-tls-pass-certificate-to-upstream.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-rule-0-path-0-redirect-scheme",
								"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-rule-0-path-0-pass-certificate-to-upstream",
								"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-rule-0-path-0-retry",
							},
							Service: "default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-pass-certificate-to-upstream",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-rule-0-path-0-tls-pass-certificate-to-upstream": {
							AuthTLSPassCertificateToUpstream: &dynamic.AuthTLSPassCertificateToUpstream{
								ClientAuthType: tls.RequireAndVerifyClientCert,
								CAFiles:        nil,
							},
						},

						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-rule-0-path-0-pass-certificate-to-upstream": {
							AuthTLSPassCertificateToUpstream: &dynamic.AuthTLSPassCertificateToUpstream{
								ClientAuthType: tls.RequireAndVerifyClientCert,
								CAFiles:        nil,
							},
						},
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-50": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
					Options: map[string]tls.Options{
						"default-7-ca-secret-9-require-and-verify-client-cert": {
							ClientAuth: tls.ClientAuth{
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----"},
								ClientAuthType: tls.RequireAndVerifyClientCert,
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
							ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
						},
					},
				},
			},
		},
		{
			desc: "Auth TLS pass certificate to upstream with ssl-redirect disabled",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-auth-tls-pass-cert-ssl-redirect-disabled.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("auth-tls-pass-cert-ssl-redirect-disabled.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53-rule-0-path-0-tls-pass-certificate-to-upstream", "default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-pass-cert-ssl-redirect-disabled",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-7-ca-secret-9-require-and-verify-client-cert",
							},
						},
						"default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("auth-tls-pass-cert-ssl-redirect-disabled.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53-rule-0-path-0-pass-certificate-to-upstream",
								"default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53-rule-0-path-0-retry",
							},
							Service: "default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-pass-cert-ssl-redirect-disabled",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53-rule-0-path-0-tls-pass-certificate-to-upstream": {
							AuthTLSPassCertificateToUpstream: &dynamic.AuthTLSPassCertificateToUpstream{
								ClientAuthType: tls.RequireAndVerifyClientCert,
								CAFiles:        nil,
							},
						},

						"default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53-rule-0-path-0-pass-certificate-to-upstream": {
							AuthTLSPassCertificateToUpstream: &dynamic.AuthTLSPassCertificateToUpstream{
								ClientAuthType: tls.RequireAndVerifyClientCert,
								CAFiles:        nil,
							},
						},
						"default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-auth-tls-pass-cert-ssl-redirect-disabled-53": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
					Options: map[string]tls.Options{
						"default-7-ca-secret-9-require-and-verify-client-cert": {
							ClientAuth: tls.ClientAuth{
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----"},
								ClientAuthType: tls.RequireAndVerifyClientCert,
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
							ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
						},
					},
				},
			},
		},
		{
			desc: "Proxy next upstream",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-next-upstream.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-32-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-next-upstream-off-36-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-off-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream-off",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-32-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-tls-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-proxy-next-upstream-off-36-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-off-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream-off",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:                 3,
								Status:                   []string{"400"},
								RetryNonIdempotentMethod: true,
								MaxRequestBodyBytes:      new(int64(0)),
							},
						},
						"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:                 3,
								Status:                   []string{"400"},
								RetryNonIdempotentMethod: true,
								MaxRequestBodyBytes:      new(int64(0)),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-next-upstream-32-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-next-upstream-32",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-next-upstream-off-36-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-next-upstream-off-36",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-next-upstream-32": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-proxy-next-upstream-off-36": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Proxy next upstream with status only conditions",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-next-upstream-status-only.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-next-upstream-status-only-44-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-status-only-44-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-proxy-next-upstream-status-only-44-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream-status-only",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-next-upstream-status-only-44-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-status-only-44-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-proxy-next-upstream-status-only-44-rule-0-path-0-tls-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream-status-only",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-next-upstream-status-only-44-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:                   3,
								Status:                     []string{"502", "503"},
								DisableRetryOnNetworkError: true,
								MaxRequestBodyBytes:        new(int64(0)),
							},
						},
						"default-7-ingress-with-proxy-next-upstream-status-only-44-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:                   3,
								Status:                     []string{"502", "503"},
								DisableRetryOnNetworkError: true,
								MaxRequestBodyBytes:        new(int64(0)),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-next-upstream-status-only-44-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-next-upstream-status-only-44",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-next-upstream-status-only-44": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Proxy next upstream with proxyRequestBuffering true",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-next-upstream.yml",
			},
			proxyRequestBuffering: true,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-buffering",
								"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-retry",
							},
						},
						"default-7-ingress-with-proxy-next-upstream-off-36-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-off-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream-off",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-proxy-next-upstream-off-36-rule-0-path-0-buffering",
							},
						},
						"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-tls-buffering",
								"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-proxy-next-upstream-off-36-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-off-36-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream-off",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-proxy-next-upstream-off-36-rule-0-path-0-tls-buffering",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-next-upstream-off-36-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								MaxRequestBodyBytes:   defaultProxyBodySize,
								MemRequestBodyBytes:   defaultClientBodyBufferSize,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-7-ingress-with-proxy-next-upstream-off-36-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								MaxRequestBodyBytes:   defaultProxyBodySize,
								MemRequestBodyBytes:   defaultClientBodyBufferSize,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								MaxRequestBodyBytes:   defaultProxyBodySize,
								MemRequestBodyBytes:   defaultClientBodyBufferSize,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								MaxRequestBodyBytes:   defaultProxyBodySize,
								MemRequestBodyBytes:   defaultClientBodyBufferSize,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:                 3,
								Status:                   []string{"400"},
								RetryNonIdempotentMethod: true,
								MaxRequestBodyBytes:      new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-next-upstream-32-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:                 3,
								Status:                   []string{"400"},
								RetryNonIdempotentMethod: true,
								MaxRequestBodyBytes:      new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-next-upstream-32-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-next-upstream-32",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-next-upstream-off-36-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-next-upstream-off-36",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-next-upstream-32": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-proxy-next-upstream-off-36": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Proxy next upstream tries",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-next-upstream-tries.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-next-upstream-tries-unlimited-48-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-tries-unlimited-48-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream-tries-unlimited",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-proxy-next-upstream-tries-unlimited-48-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-proxy-next-upstream-tries-38-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-tries-38-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream-tries",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-proxy-next-upstream-tries-38-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-proxy-next-upstream-tries-unlimited-48-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-tries-unlimited-48-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream-tries-unlimited",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-proxy-next-upstream-tries-unlimited-48-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-proxy-next-upstream-tries-38-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-tries-38-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream-tries",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-proxy-next-upstream-tries-38-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-next-upstream-tries-unlimited-48-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            2,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-next-upstream-tries-unlimited-48-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            2,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-next-upstream-tries-38-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            5,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-next-upstream-tries-38-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            5,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-next-upstream-tries-unlimited-48-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-next-upstream-tries-unlimited-48",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-proxy-next-upstream-tries-38-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-next-upstream-tries-38",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-next-upstream-tries-unlimited-48": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-proxy-next-upstream-tries-38": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Proxy next upstream timeout",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-next-upstream-timeout.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-next-upstream-timeout-40-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-timeout-40-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream-timeout",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-proxy-next-upstream-timeout-40-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-proxy-next-upstream-timeout-40-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-next-upstream-timeout-40-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-next-upstream-timeout",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-proxy-next-upstream-timeout-40-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-next-upstream-timeout-40-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								Timeout:             ptypes.Duration(30 * time.Second),
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-next-upstream-timeout-40-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								Timeout:             ptypes.Duration(30 * time.Second),
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-next-upstream-timeout-40-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-next-upstream-timeout-40",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-next-upstream-timeout-40": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Server Alias",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-server-alias.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-server-alias-25-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("original.localhost") || Host("alias1.localhost") || Host("alias2.localhost")) && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-server-alias-25-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-server-alias-25-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-server-alias",
									},
								},
							},
						},
						"default-7-ingress-with-server-alias-25-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("original.localhost") || Host("alias1.localhost") || Host("alias2.localhost")) && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-server-alias-25-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-server-alias-25-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-server-alias",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-server-alias-25-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-server-alias-25-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},

					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-server-alias-25-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-server-alias-25",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-server-alias-25": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Server Alias with Conflict",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-server-alias-conflict.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-primary-ingress-15-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("conflict.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-primary-ingress-15-rule-0-path-0-retry"},
							Service:     "default-7-primary-ingress-15-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "primary-ingress",
									},
								},
							},
						},
						"default-7-alias-ingress-13-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("original.localhost") || Host("alias1.localhost")) && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-alias-ingress-13-rule-0-path-0-retry"},
							Service:     "default-7-alias-ingress-13-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "alias-ingress",
									},
								},
							},
						},
						"default-7-primary-ingress-15-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("conflict.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-primary-ingress-15-rule-0-path-0-tls-retry"},
							Service:     "default-7-primary-ingress-15-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "primary-ingress",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-alias-ingress-13-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("original.localhost") || Host("alias1.localhost")) && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-alias-ingress-13-rule-0-path-0-tls-retry"},
							Service:     "default-7-alias-ingress-13-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "alias-ingress",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-primary-ingress-15-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-primary-ingress-15-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-alias-ingress-13-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-alias-ingress-13-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-primary-ingress-15-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-primary-ingress-15",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-alias-ingress-13-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-alias-ingress-13",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-primary-ingress-15": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-alias-ingress-13": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Server Alias with Alias-Alias Conflict",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-server-alias-alias-conflict.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-first-ingress-13-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("first.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-first-ingress-13-rule-0-path-0-retry"},
							Service:     "default-7-first-ingress-13-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Namespace: "default",
										Name:      "first-ingress",
										Kind:      "Ingress",
									},
								},
							},
						},
						"default-7-second-ingress-14-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("second.localhost") || Host("shared.localhost")) && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-second-ingress-14-rule-0-path-0-retry"},
							Service:     "default-7-second-ingress-14-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Namespace: "default",
										Name:      "second-ingress",
										Kind:      "Ingress",
									},
								},
							},
						},
						"default-7-first-ingress-13-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("first.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-first-ingress-13-rule-0-path-0-tls-retry"},
							Service:     "default-7-first-ingress-13-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Namespace: "default",
										Name:      "first-ingress",
										Kind:      "Ingress",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-second-ingress-14-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("second.localhost") || Host("shared.localhost")) && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-second-ingress-14-rule-0-path-0-tls-retry"},
							Service:     "default-7-second-ingress-14-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Namespace: "default",
										Name:      "second-ingress",
										Kind:      "Ingress",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-first-ingress-13-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-first-ingress-13-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-second-ingress-14-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-second-ingress-14-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-first-ingress-13-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-first-ingress-13",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-second-ingress-14-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-second-ingress-14",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-first-ingress-13": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-second-ingress-14": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Proxy HTTP version 1.1",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-http-version.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-http-version-31-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("proxy-http-version.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-http-version-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-http-version",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-proxy-http-version-31-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-proxy-http-version-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("proxy-http-version.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-http-version-31-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-http-version",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-proxy-http-version-31-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-http-version-31-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-http-version-31-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-http-version-31-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-http-version-31",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-http-version-31": {
							DisableHTTP2: true,
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Proxy HTTP version 1.0 (unsupported)",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-proxy-http-version-unsupported.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-proxy-http-version-unsupported-43-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("proxy-http-version-unsupported.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-http-version-unsupported-43-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-http-version-unsupported",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-proxy-http-version-unsupported-43-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-proxy-http-version-unsupported-43-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("proxy-http-version-unsupported.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-proxy-http-version-unsupported-43-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-proxy-http-version-unsupported",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-proxy-http-version-unsupported-43-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-proxy-http-version-unsupported-43-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-proxy-http-version-unsupported-43-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-proxy-http-version-unsupported-43-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-proxy-http-version-unsupported-43",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-proxy-http-version-unsupported-43": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Upstream Hash By",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-upstream-hash-by.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-upstream-hash-by-29-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-upstream-hash-by-29-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-upstream-hash-by",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-upstream-hash-by-29-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-upstream-hash-by-29-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-upstream-hash-by-29-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-upstream-hash-by",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-upstream-hash-by-29-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-upstream-hash-by-29-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-upstream-hash-by-29-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-upstream-hash-by-29-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:            "hrw",
								NginxUpstreamHashBy: "$request_uri",
								PassHostHeader:      new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-upstream-hash-by-29",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-upstream-hash-by-29": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Canary",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-canary.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-canary-19-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-19-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-19-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-canary-19-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-19-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-19-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-canary-19-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-19-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-canary-19-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-19",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-19-whoami-6-80-2-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:80",
									},
									{
										URL: "http://10.10.0.8:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-19",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "canary",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-19-whoami-6-80-2-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-7-ingress-with-canary-19-whoami-6-80-2",
										Weight: new(100),
									},
									{
										Name:   "default-7-ingress-with-canary-19-whoami-6-80-2-canary",
										Weight: new(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-canary-19": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Canary with sticky",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-canary-and-sticky.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-canary-and-sticky-30-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-and-sticky-30-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-and-sticky",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-and-sticky-30-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-canary-and-sticky-30-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-and-sticky-30-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-and-sticky",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-and-sticky-30-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-canary-and-sticky-30-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-and-sticky-30-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-canary-and-sticky-30-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:               "foobar",
										Domain:             "foo.localhost",
										HTTPOnly:           true,
										MaxAge:             42,
										Expires:            42,
										Path:               new("/foobar"),
										SameSite:           "none",
										Secure:             true,
										PreserveLeadingDot: true,
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
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-and-sticky-30",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-and-sticky-30-whoami-6-80-2-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:               "foobar",
										Domain:             "foo.localhost",
										HTTPOnly:           true,
										MaxAge:             42,
										Expires:            42,
										Path:               new("/foobar"),
										SameSite:           "none",
										Secure:             true,
										PreserveLeadingDot: true,
									},
								},
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:80",
									},
									{
										URL: "http://10.10.0.8:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-and-sticky-30",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "canary",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-and-sticky-30-whoami-6-80-2-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:               "foobar-wrr",
										Domain:             "foo.localhost",
										HTTPOnly:           true,
										MaxAge:             42,
										Expires:            42,
										Path:               new("/foobar"),
										SameSite:           "none",
										Secure:             true,
										PreserveLeadingDot: true,
									},
								},
								Services: []dynamic.WRRService{
									{
										Name:   "default-7-ingress-with-canary-and-sticky-30-whoami-6-80-2",
										Weight: new(100),
									},
									{
										Name:   "default-7-ingress-with-canary-and-sticky-30-whoami-6-80-2-canary",
										Weight: new(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-canary-and-sticky-30": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Canary with weight",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-canary-weight.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-canary-weight-26-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-weight-26-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-weight",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-weight-26-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-canary-weight-26-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-weight-26-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-weight",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-weight-26-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-canary-weight-26-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-weight-26-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-canary-weight-26-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-weight-26",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-weight-26-whoami-6-80-2-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:80",
									},
									{
										URL: "http://10.10.0.8:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-weight-26",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "canary",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-weight-26-whoami-6-80-2-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-7-ingress-with-canary-weight-26-whoami-6-80-2",
										Weight: new(110),
									},
									{
										Name:   "default-7-ingress-with-canary-weight-26-whoami-6-80-2-canary",
										Weight: new(10),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-canary-weight-26": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Canary with header",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-canary-by-header.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-canary-by-header-29-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-29-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-29-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-canary-by-header-29-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-29-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-29-rule-0-path-0-canary-retry"},
						},
						"default-7-ingress-with-canary-by-header-29-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-29-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-29-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-canary-by-header-29-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-29-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-29-rule-0-path-0-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-canary-by-header-29-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-29-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-29-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-29-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-canary-by-header-29-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-header-29",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-header-29-whoami-6-80-2-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:80",
									},
									{
										URL: "http://10.10.0.8:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-header-29",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "canary",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-header-29-whoami-6-80-2-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-7-ingress-with-canary-by-header-29-whoami-6-80-2",
										Weight: new(100),
									},
									{
										Name:   "default-7-ingress-with-canary-by-header-29-whoami-6-80-2-canary",
										Weight: new(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-canary-by-header-29": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Canary with header value",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-canary-by-header-value.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-canary-by-header-value-35-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-value-35-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-value",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-value-35-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-canary-by-header-value-35-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "bar"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-value-35-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-value",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-value-35-rule-0-path-0-canary-retry"},
						},
						"default-7-ingress-with-canary-by-header-value-35-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-value-35-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-value",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-value-35-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-canary-by-header-value-35-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "bar"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-value-35-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-value",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-value-35-rule-0-path-0-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-canary-by-header-value-35-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-value-35-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-value-35-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-value-35-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-canary-by-header-value-35-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-header-value-35",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-header-value-35-whoami-6-80-2-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:80",
									},
									{
										URL: "http://10.10.0.8:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-header-value-35",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "canary",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-header-value-35-whoami-6-80-2-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-7-ingress-with-canary-by-header-value-35-whoami-6-80-2",
										Weight: new(100),
									},
									{
										Name:   "default-7-ingress-with-canary-by-header-value-35-whoami-6-80-2-canary",
										Weight: new(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-canary-by-header-value-35": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Canary with header pattern",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-canary-by-header-pattern.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-canary-by-header-pattern-37-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-pattern-37-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-pattern",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-pattern-37-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-canary-by-header-pattern-37-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (HeaderRegexp("Foo", "bar(.*)"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-pattern-37-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-pattern",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-pattern-37-rule-0-path-0-canary-retry"},
						},
						"default-7-ingress-with-canary-by-header-pattern-37-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-pattern-37-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-pattern",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-pattern-37-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-canary-by-header-pattern-37-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (HeaderRegexp("Foo", "bar(.*)"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-pattern-37-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-pattern",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-pattern-37-rule-0-path-0-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-canary-by-header-pattern-37-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-pattern-37-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-pattern-37-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-pattern-37-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-canary-by-header-pattern-37-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-header-pattern-37",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-header-pattern-37-whoami-6-80-2-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:80",
									},
									{
										URL: "http://10.10.0.8:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-header-pattern-37",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "canary",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-header-pattern-37-whoami-6-80-2-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-7-ingress-with-canary-by-header-pattern-37-whoami-6-80-2",
										Weight: new(100),
									},
									{
										Name:   "default-7-ingress-with-canary-by-header-pattern-37-whoami-6-80-2-canary",
										Weight: new(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-canary-by-header-pattern-37": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Canary with header misconfigured",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-canary-by-header-misconfigured.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-canary-by-header-misconfigured-43-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-misconfigured-43-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-misconfigured",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-misconfigured-43-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-canary-by-header-misconfigured-43-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-misconfigured-43-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-misconfigured",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-misconfigured-43-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-canary-by-header-misconfigured-43-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-misconfigured-43-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-canary-by-header-misconfigured-43-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-header-misconfigured-43",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-header-misconfigured-43-whoami-6-80-2-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:80",
									},
									{
										URL: "http://10.10.0.8:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-header-misconfigured-43",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "canary",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-header-misconfigured-43-whoami-6-80-2-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-7-ingress-with-canary-by-header-misconfigured-43-whoami-6-80-2",
										Weight: new(100),
									},
									{
										Name:   "default-7-ingress-with-canary-by-header-misconfigured-43-whoami-6-80-2-canary",
										Weight: new(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-canary-by-header-misconfigured-43": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Canary with cookie",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-canary-by-cookie.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-canary-by-cookie-29-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-cookie-29-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-cookie",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-cookie-29-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-canary-by-cookie-29-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (HeaderRegexp("Cookie", "(^|;\\s*)foo\\.bar=always(;|$)"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-cookie-29-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-cookie",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-cookie-29-rule-0-path-0-canary-retry"},
						},
						"default-7-ingress-with-canary-by-cookie-29-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-cookie-29-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-cookie",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-cookie-29-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-canary-by-cookie-29-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (HeaderRegexp("Cookie", "(^|;\\s*)foo\\.bar=always(;|$)"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-cookie-29-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-cookie",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-cookie-29-rule-0-path-0-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-canary-by-cookie-29-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-cookie-29-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-cookie-29-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-cookie-29-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-canary-by-cookie-29-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-cookie-29",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-cookie-29-whoami-6-80-2-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:80",
									},
									{
										URL: "http://10.10.0.8:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-cookie-29",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "canary",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-cookie-29-whoami-6-80-2-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-7-ingress-with-canary-by-cookie-29-whoami-6-80-2",
										Weight: new(100),
									},
									{
										Name:   "default-7-ingress-with-canary-by-cookie-29-whoami-6-80-2-canary",
										Weight: new(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-canary-by-cookie-29": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			// Every router of the location gets its own allowlist router, so that the rule
			// length keeps ordering them: the canary allowlist rule stays the longest, and
			// an allowlisted canary request is not captured by the primary allowlist router.
			desc: "Canary with cookie and limit allowlist",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-canary-and-limit-allowlist.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-and-limit-allowlist-39-whoami-6-80-2-wrr",
							Middlewares: []string{"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-limit-rps", "default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-and-limit-allowlist",
									},
								},
							},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-and-limit-allowlist-39-whoami-6-80-2-wrr",
							Middlewares: []string{"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-tls-limit-rps", "default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-and-limit-allowlist",
									},
								},
							},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-limit-allowlist": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && ClientIP("10.0.0.0/24")`,
							RuleSyntax:  "default",
							Priority:    48,
							Service:     "default-7-ingress-with-canary-and-limit-allowlist-39-whoami-6-80-2-wrr",
							Middlewares: []string{"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-limit-allowlist-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-and-limit-allowlist",
									},
								},
							},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-tls-limit-allowlist": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && ClientIP("10.0.0.0/24")`,
							RuleSyntax:  "default",
							Priority:    48,
							Service:     "default-7-ingress-with-canary-and-limit-allowlist-39-whoami-6-80-2-wrr",
							Middlewares: []string{"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-tls-limit-allowlist-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-and-limit-allowlist",
									},
								},
							},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (HeaderRegexp("Cookie", "(^|;\\s*)foo\\.bar=always(;|$)"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-and-limit-allowlist-39-whoami-6-80-2-canary",
							Middlewares: []string{"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-limit-rps", "default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-and-limit-allowlist",
									},
								},
							},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (HeaderRegexp("Cookie", "(^|;\\s*)foo\\.bar=always(;|$)"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-and-limit-allowlist-39-whoami-6-80-2-canary",
							Middlewares: []string{"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-tls-limit-rps", "default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-and-limit-allowlist",
									},
								},
							},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-limit-allowlist": {
							EntryPoints: []string{"http"},
							Rule:        `((Host("production.localhost") && PathPrefix("/")) && (HeaderRegexp("Cookie", "(^|;\\s*)foo\\.bar=always(;|$)"))) && ClientIP("10.0.0.0/24")`,
							RuleSyntax:  "default",
							Priority:    112,
							Service:     "default-7-ingress-with-canary-and-limit-allowlist-39-whoami-6-80-2-canary",
							Middlewares: []string{"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-limit-allowlist-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-and-limit-allowlist",
									},
								},
							},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-tls-limit-allowlist": {
							EntryPoints: []string{"https"},
							Rule:        `((Host("production.localhost") && PathPrefix("/")) && (HeaderRegexp("Cookie", "(^|;\\s*)foo\\.bar=always(;|$)"))) && ClientIP("10.0.0.0/24")`,
							RuleSyntax:  "default",
							Priority:    112,
							Service:     "default-7-ingress-with-canary-and-limit-allowlist-39-whoami-6-80-2-canary",
							Middlewares: []string{"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-tls-limit-allowlist-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-and-limit-allowlist",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-limit-allowlist-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-tls-limit-allowlist-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-limit-allowlist-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-tls-limit-allowlist-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-limit-rps": {
							RateLimit: &dynamic.RateLimit{Average: 10, Burst: 50, Period: ptypes.Duration(time.Second)},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-tls-limit-rps": {
							RateLimit: &dynamic.RateLimit{Average: 10, Burst: 50, Period: ptypes.Duration(time.Second)},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-limit-rps": {
							RateLimit: &dynamic.RateLimit{Average: 10, Burst: 50, Period: ptypes.Duration(time.Second)},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-rule-0-path-0-canary-tls-limit-rps": {
							RateLimit: &dynamic.RateLimit{Average: 10, Burst: 50, Period: ptypes.Duration(time.Second)},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-and-limit-allowlist-39",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-whoami-6-80-2-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.7:80"},
									{URL: "http://10.10.0.8:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-and-limit-allowlist-39",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "canary",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-and-limit-allowlist-39-whoami-6-80-2-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-7-ingress-with-canary-and-limit-allowlist-39-whoami-6-80-2",
										Weight: new(100),
									},
									{
										Name:   "default-7-ingress-with-canary-and-limit-allowlist-39-whoami-6-80-2-canary",
										Weight: new(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-canary-and-limit-allowlist-39": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Canary with header, cookie, and weight",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-canary-by-header-and-cookie-and-weight.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-and-cookie-and-weight",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always") || (HeaderRegexp("Cookie", "(^|;\\s*)foo=always(;|$)") && !Header("Foo", "never")))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-and-cookie-and-weight",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-canary-retry"},
						},
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-non-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "never") || HeaderRegexp("Cookie", "(^|;\\s*)foo=never(;|$)"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-and-cookie-and-weight",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-non-canary-retry"},
						},
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-and-cookie-and-weight",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always") || (HeaderRegexp("Cookie", "(^|;\\s*)foo=always(;|$)") && !Header("Foo", "never")))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-and-cookie-and-weight",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-non-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "never") || HeaderRegexp("Cookie", "(^|;\\s*)foo=never(;|$)"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header-and-cookie-and-weight",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-non-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-non-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-rule-0-path-0-non-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-header-and-cookie-and-weight-51",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-whoami-6-80-2-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:80",
									},
									{
										URL: "http://10.10.0.8:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-header-and-cookie-and-weight-51",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "canary",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-whoami-6-80-2-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-whoami-6-80-2",
										Weight: new(90),
									},
									{
										Name:   "default-7-ingress-with-canary-by-header-and-cookie-and-weight-51-whoami-6-80-2-canary",
										Weight: new(10),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-canary-by-header-and-cookie-and-weight-51": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Canary with middlewares on production Ingress",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-canary-middlewares.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-canary-middlewares-31-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-middlewares-31-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-middlewares",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-app-root",
								"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-retry",
							},
						},
						"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-middlewares-31-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-middlewares",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-canary-app-root",
								"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-canary-retry",
							},
						},
						"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-middlewares-31-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-middlewares",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-tls-app-root",
								"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-middlewares-31-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-middlewares",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-canary-tls-app-root",
								"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-canary-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-tls-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-canary-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-canary-tls-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-middlewares-31-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-canary-middlewares-31-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-middlewares-31",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-middlewares-31-whoami-6-80-2-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:80",
									},
									{
										URL: "http://10.10.0.8:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-middlewares-31",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "canary",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-middlewares-31-whoami-6-80-2-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-7-ingress-with-canary-middlewares-31-whoami-6-80-2",
										Weight: new(100),
									},
									{
										Name:   "default-7-ingress-with-canary-middlewares-31-whoami-6-80-2-canary",
										Weight: new(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-canary-middlewares-31": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Canary with non matching canaries",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-non-matching-canary.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-non-matching-canary-32-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-non-matching-canary-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-non-matching-canary",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-non-matching-canary-32-rule-0-path-0-retry"},
						},
						"default-7-ingress-with-non-matching-canary-32-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-non-matching-canary-32-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-non-matching-canary",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-non-matching-canary-32-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-non-matching-canary-32-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-non-matching-canary-32-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-non-matching-canary-32-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-non-matching-canary-32",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-non-matching-canary-32": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Canary with TLS on production Ingress",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"secrets.yml",
				"ingresses/ingresses-with-canary-middlewares-and-tls.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-middlewares-and-tls-39-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-middlewares-and-tls",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-redirect-scheme",
								"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-app-root",
								"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-retry",
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-middlewares-and-tls-39-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-middlewares-and-tls",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-canary-redirect-scheme",
								"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-canary-app-root",
								"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-canary-retry",
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-middlewares-and-tls-39-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-middlewares-and-tls",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-tls-app-root",
								"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-middlewares-and-tls-39-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-middlewares-and-tls",
									},
								},
							},
							Middlewares: []string{
								"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-canary-tls-app-root",
								"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-canary-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-tls-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-canary-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-canary-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-canary-tls-app-root": {
							AppRoot: &dynamic.AppRoot{
								Path: "/foo",
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-middlewares-and-tls-39",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-whoami-6-80-2-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.7:80"},
									{URL: "http://10.10.0.8:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-middlewares-and-tls-39",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "canary",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-middlewares-and-tls-39-whoami-6-80-2-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{Name: "default-7-ingress-with-canary-middlewares-and-tls-39-whoami-6-80-2", Weight: new(100)},
									{Name: "default-7-ingress-with-canary-middlewares-and-tls-39-whoami-6-80-2-canary", Weight: new(0)},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-canary-middlewares-and-tls-39": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
				},
			},
		},
		{
			desc: "Limit RPS",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-limit-rps.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-limit-rps-22-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-rps-22-rule-0-path-0-limit-rps", "default-7-ingress-with-limit-rps-22-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-limit-rps-22-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-rps",
									},
								},
							},
						},
						"default-7-ingress-with-limit-rps-zero-27-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami-zero.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-rps-zero-27-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-limit-rps-zero-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-rps-zero",
									},
								},
							},
						},
						"default-7-ingress-with-limit-rps-22-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-rps-22-rule-0-path-0-tls-limit-rps", "default-7-ingress-with-limit-rps-22-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-limit-rps-22-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-rps",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-limit-rps-zero-27-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami-zero.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-rps-zero-27-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-limit-rps-zero-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-rps-zero",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-limit-rps-22-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-limit-rps-22-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-limit-rps-22-rule-0-path-0-limit-rps": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Second),
							},
						},
						"default-7-ingress-with-limit-rps-22-rule-0-path-0-tls-limit-rps": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Second),
							},
						},
						"default-7-ingress-with-limit-rps-zero-27-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-limit-rps-zero-27-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-limit-rps-22-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-limit-rps-22",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-limit-rps-zero-27-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-limit-rps-zero-27",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-limit-rps-22": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-limit-rps-zero-27": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Limit RPS with the legacy limit-whitelist alias",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-limit-whitelist.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-limit-whitelist-28-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-whitelist-28-rule-0-path-0-limit-rps", "default-7-ingress-with-limit-whitelist-28-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-limit-whitelist-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-whitelist",
									},
								},
							},
						},
						"default-7-ingress-with-limit-whitelist-28-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-whitelist-28-rule-0-path-0-tls-limit-rps", "default-7-ingress-with-limit-whitelist-28-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-limit-whitelist-28-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-whitelist",
									},
								},
							},
						},
						"default-7-ingress-with-limit-whitelist-28-rule-0-path-0-limit-allowlist": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("whoami.localhost") && Path("/")) && ClientIP("10.0.0.0/24", "192.168.1.1")`,
							RuleSyntax:  "default",
							Priority:    38,
							Middlewares: []string{"default-7-ingress-with-limit-whitelist-28-rule-0-path-0-limit-allowlist-retry"},
							Service:     "default-7-ingress-with-limit-whitelist-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-whitelist",
									},
								},
							},
						},
						"default-7-ingress-with-limit-whitelist-28-rule-0-path-0-tls-limit-allowlist": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("whoami.localhost") && Path("/")) && ClientIP("10.0.0.0/24", "192.168.1.1")`,
							RuleSyntax:  "default",
							Priority:    38,
							Middlewares: []string{"default-7-ingress-with-limit-whitelist-28-rule-0-path-0-tls-limit-allowlist-retry"},
							Service:     "default-7-ingress-with-limit-whitelist-28-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-whitelist",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-limit-whitelist-28-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-limit-whitelist-28-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-limit-whitelist-28-rule-0-path-0-limit-allowlist-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-limit-whitelist-28-rule-0-path-0-tls-limit-allowlist-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-limit-whitelist-28-rule-0-path-0-limit-rps": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Second),
							},
						},
						"default-7-ingress-with-limit-whitelist-28-rule-0-path-0-tls-limit-rps": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Second),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-limit-whitelist-28-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-limit-whitelist-28",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-limit-whitelist-28": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Limit RPS and connections with allowlist",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-limit-allowlist.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-limit-allowlist-28-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-limit-rps", "default-7-ingress-with-limit-allowlist-28-rule-0-path-0-limit-connections", "default-7-ingress-with-limit-allowlist-28-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-limit-allowlist-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-allowlist",
									},
								},
							},
						},
						"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-tls-limit-rps", "default-7-ingress-with-limit-allowlist-28-rule-0-path-0-tls-limit-connections", "default-7-ingress-with-limit-allowlist-28-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-limit-allowlist-28-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-allowlist",
									},
								},
							},
						},
						"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-limit-allowlist": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("whoami.localhost") && Path("/")) && ClientIP("10.0.0.0/24", "192.168.1.1")`,
							RuleSyntax:  "default",
							Priority:    38,
							Middlewares: []string{"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-limit-allowlist-retry"},
							Service:     "default-7-ingress-with-limit-allowlist-28-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-allowlist",
									},
								},
							},
						},
						"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-tls-limit-allowlist": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("whoami.localhost") && Path("/")) && ClientIP("10.0.0.0/24", "192.168.1.1")`,
							RuleSyntax:  "default",
							Priority:    38,
							Middlewares: []string{"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-tls-limit-allowlist-retry"},
							Service:     "default-7-ingress-with-limit-allowlist-28-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-allowlist",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-limit-allowlist-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-tls-limit-allowlist-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-limit-rps": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Second),
							},
						},
						"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-tls-limit-rps": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Second),
							},
						},
						"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-limit-connections": {
							InFlightReq: &dynamic.InFlightReq{
								Amount: 10,
								SourceCriterion: &dynamic.SourceCriterion{
									IPStrategy: &dynamic.IPStrategy{},
								},
							},
						},
						"default-7-ingress-with-limit-allowlist-28-rule-0-path-0-tls-limit-connections": {
							InFlightReq: &dynamic.InFlightReq{
								Amount: 10,
								SourceCriterion: &dynamic.SourceCriterion{
									IPStrategy: &dynamic.IPStrategy{},
								},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-limit-allowlist-28-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-limit-allowlist-28",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-limit-allowlist-28": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Limit allowlist without any limit",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-limit-allowlist-no-limit.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-limit-allowlist-no-limit-37-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-allowlist-no-limit-37-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-limit-allowlist-no-limit-37-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-allowlist-no-limit",
									},
								},
							},
						},
						"default-7-ingress-with-limit-allowlist-no-limit-37-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-allowlist-no-limit-37-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-limit-allowlist-no-limit-37-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-allowlist-no-limit",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-limit-allowlist-no-limit-37-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-limit-allowlist-no-limit-37-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-limit-allowlist-no-limit-37-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-limit-allowlist-no-limit-37",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-limit-allowlist-no-limit-37": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Limit allowlist on an ingress default backend without rules",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-no-rules-limit-allowlist.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-limit-rps", "default-backend-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-limit-allowlist",
									},
								},
							},
						},
						"default-backend-limit-allowlist": {
							EntryPoints: []string{"http"},
							Rule:        `(PathPrefix("/")) && ClientIP("10.0.0.0/24")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32 + 1,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-limit-allowlist-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-limit-allowlist",
									},
								},
							},
						},
						"default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-tls-limit-rps", "default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-limit-allowlist",
									},
								},
							},
						},
						"default-backend-tls-limit-allowlist": {
							EntryPoints: []string{"https"},
							Rule:        `(PathPrefix("/")) && ClientIP("10.0.0.0/24")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32 + 1,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-tls-limit-allowlist-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules-limit-allowlist",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-backend-limit-rps": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Second),
							},
						},
						"default-backend-tls-limit-rps": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Second),
							},
						},
						"default-backend-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-backend-limit-allowlist-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-backend-tls-limit-allowlist-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-no-rules-limit-allowlist-53",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-no-rules-limit-allowlist-53": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Limit RPM",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-limit-rpm.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-limit-rpm-22-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-rpm-22-rule-0-path-0-limit-rpm", "default-7-ingress-with-limit-rpm-22-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-limit-rpm-22-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-rpm",
									},
								},
							},
						},
						"default-7-ingress-with-limit-rpm-zero-27-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami-zero.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-rpm-zero-27-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-limit-rpm-zero-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-rpm-zero",
									},
								},
							},
						},
						"default-7-ingress-with-limit-rpm-22-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-rpm-22-rule-0-path-0-tls-limit-rpm", "default-7-ingress-with-limit-rpm-22-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-limit-rpm-22-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-rpm",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-limit-rpm-zero-27-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami-zero.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-rpm-zero-27-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-limit-rpm-zero-27-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-rpm-zero",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-limit-rpm-22-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-limit-rpm-22-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-limit-rpm-22-rule-0-path-0-limit-rpm": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Minute),
							},
						},
						"default-7-ingress-with-limit-rpm-22-rule-0-path-0-tls-limit-rpm": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Minute),
							},
						},
						"default-7-ingress-with-limit-rpm-zero-27-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-limit-rpm-zero-27-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-limit-rpm-22-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-limit-rpm-22",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-limit-rpm-zero-27-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-limit-rpm-zero-27",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-limit-rpm-22": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-7-ingress-with-limit-rpm-zero-27": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Limit Burst Multiplier",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-limit-burst-multiplier.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{Routers: map[string]*dynamic.TCPRouter{}, Services: map[string]*dynamic.TCPService{}},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-limit-burst-multiplier-35-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami-burst.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-burst-multiplier-35-rule-0-path-0-limit-rps", "default-7-ingress-with-limit-burst-multiplier-35-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-limit-burst-multiplier-35-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-burst-multiplier",
									},
								},
							},
						},
						"default-7-ingress-with-limit-burst-multiplier-35-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami-burst.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-burst-multiplier-35-rule-0-path-0-tls-limit-rps", "default-7-ingress-with-limit-burst-multiplier-35-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-limit-burst-multiplier-35-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-burst-multiplier",
									},
								},
							},
						},
						"default-7-ingress-with-limit-burst-multiplier-zero-40-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami-burst-zero.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-burst-multiplier-zero-40-rule-0-path-0-limit-rps", "default-7-ingress-with-limit-burst-multiplier-zero-40-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-limit-burst-multiplier-zero-40-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-burst-multiplier-zero",
									},
								},
							},
						},
						"default-7-ingress-with-limit-burst-multiplier-zero-40-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami-burst-zero.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-burst-multiplier-zero-40-rule-0-path-0-tls-limit-rps", "default-7-ingress-with-limit-burst-multiplier-zero-40-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-limit-burst-multiplier-zero-40-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-burst-multiplier-zero",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-limit-burst-multiplier-35-rule-0-path-0-retry":          {Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)}},
						"default-7-ingress-with-limit-burst-multiplier-35-rule-0-path-0-tls-retry":      {Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)}},
						"default-7-ingress-with-limit-burst-multiplier-zero-40-rule-0-path-0-retry":     {Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)}},
						"default-7-ingress-with-limit-burst-multiplier-zero-40-rule-0-path-0-tls-retry": {Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)}},
						"default-7-ingress-with-limit-burst-multiplier-35-rule-0-path-0-limit-rps": {
							RateLimit: &dynamic.RateLimit{Average: 10, Burst: 100, Period: ptypes.Duration(time.Second)},
						},
						"default-7-ingress-with-limit-burst-multiplier-35-rule-0-path-0-tls-limit-rps": {
							RateLimit: &dynamic.RateLimit{Average: 10, Burst: 100, Period: ptypes.Duration(time.Second)},
						},
						"default-7-ingress-with-limit-burst-multiplier-zero-40-rule-0-path-0-limit-rps": {
							RateLimit: &dynamic.RateLimit{Average: 10, Burst: 50, Period: ptypes.Duration(time.Second)},
						},
						"default-7-ingress-with-limit-burst-multiplier-zero-40-rule-0-path-0-tls-limit-rps": {
							RateLimit: &dynamic.RateLimit{Average: 10, Burst: 50, Period: ptypes.Duration(time.Second)},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-limit-burst-multiplier-35-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers:            []dynamic.Server{{URL: "http://10.10.0.1:80"}, {URL: "http://10.10.0.2:80"}},
								Strategy:           "wrr",
								PassHostHeader:     new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								ServersTransport:   "default-7-ingress-with-limit-burst-multiplier-35",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-limit-burst-multiplier-zero-40-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers:            []dynamic.Server{{URL: "http://10.10.0.1:80"}, {URL: "http://10.10.0.2:80"}},
								Strategy:           "wrr",
								PassHostHeader:     new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								ServersTransport:   "default-7-ingress-with-limit-burst-multiplier-zero-40",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-limit-burst-multiplier-35": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{DialTimeout: ptypes.Duration(60 * time.Second), ReadTimeout: ptypes.Duration(60 * time.Second), WriteTimeout: ptypes.Duration(60 * time.Second), IdleConnTimeout: ptypes.Duration(60 * time.Second)},
						},
						"default-7-ingress-with-limit-burst-multiplier-zero-40": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{DialTimeout: ptypes.Duration(60 * time.Second), ReadTimeout: ptypes.Duration(60 * time.Second), WriteTimeout: ptypes.Duration(60 * time.Second), IdleConnTimeout: ptypes.Duration(60 * time.Second)},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Limit connections",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-limit-connections.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-limit-connections-30-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-connections-30-rule-0-path-0-limit-connections", "default-7-ingress-with-limit-connections-30-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-limit-connections-30-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-connections",
									},
								},
							},
						},
						"default-7-ingress-with-limit-connections-30-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-limit-connections-30-rule-0-path-0-tls-limit-connections", "default-7-ingress-with-limit-connections-30-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-limit-connections-30-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-limit-connections",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-limit-connections-30-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-limit-connections-30-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-limit-connections-30-rule-0-path-0-limit-connections": {
							InFlightReq: &dynamic.InFlightReq{
								Amount: 10,
								SourceCriterion: &dynamic.SourceCriterion{
									IPStrategy: &dynamic.IPStrategy{},
								},
							},
						},
						"default-7-ingress-with-limit-connections-30-rule-0-path-0-tls-limit-connections": {
							InFlightReq: &dynamic.InFlightReq{
								Amount: 10,
								SourceCriterion: &dynamic.SourceCriterion{
									IPStrategy: &dynamic.IPStrategy{},
								},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-limit-connections-30-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers:            []dynamic.Server{{URL: "http://10.10.0.1:80"}, {URL: "http://10.10.0.2:80"}},
								Strategy:           "wrr",
								PassHostHeader:     new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								ServersTransport:   "default-7-ingress-with-limit-connections-30",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-limit-connections-30": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{DialTimeout: ptypes.Duration(60 * time.Second), ReadTimeout: ptypes.Duration(60 * time.Second), WriteTimeout: ptypes.Duration(60 * time.Second), IdleConnTimeout: ptypes.Duration(60 * time.Second)},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Use Regex with Prefix pathType and StrictValidatePathType enabled",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-use-regex-prefix-pathtype.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
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
			desc: "Use Regex with Prefix pathType and StrictValidatePathType disabled",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-use-regex-prefix-pathtype.yml",
			},
			strictValidatePathType: new(false),
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-use-regex-22-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("use-regex.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-use-regex-22-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-use-regex-22-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-use-regex",
									},
								},
							},
						},
						"default-7-ingress-with-use-regex-22-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("use-regex.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-use-regex-22-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-use-regex-22-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-use-regex",
									},
								},
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-use-regex-22-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
						"default-7-ingress-with-use-regex-22-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3, MaxRequestBodyBytes: new(defaultProxyBodySize)},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-use-regex-22-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-use-regex-22",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-use-regex-22": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Wildcard host",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-wildcard-host.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-wildcard-host-26-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("*.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-wildcard-host-26-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-wildcard-host-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-wildcard-host",
									},
								},
							},
						},
						"default-7-ingress-with-wildcard-host-26-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("*.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-7-ingress-with-wildcard-host-26-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-wildcard-host-26-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-wildcard-host",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-wildcard-host-26-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-wildcard-host-26-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-wildcard-host-26-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-wildcard-host-26",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-wildcard-host-26": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Wildcard host with TLS",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"secrets.yml",
				"ingresses/ingress-with-wildcard-host-tls.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-wildcard-host-tls-30-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("*.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-ingress-with-wildcard-host-tls-30-rule-0-path-0-redirect-scheme",
								"default-7-ingress-with-wildcard-host-tls-30-rule-0-path-0-retry",
							},
							Service: "default-7-ingress-with-wildcard-host-tls-30-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-wildcard-host-tls",
									},
								},
							},
						},
						"default-7-ingress-with-wildcard-host-tls-30-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("*.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-7-ingress-with-wildcard-host-tls-30-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-wildcard-host-tls-30-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-wildcard-host-tls",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-wildcard-host-tls-30-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-wildcard-host-tls-30-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-wildcard-host-tls-30-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-wildcard-host-tls-30-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-wildcard-host-tls-30",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-wildcard-host-tls-30": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
				},
			},
		},
		{
			desc: "Ingress with multiple paths and one invalid path with StrictValidatePathType drops the whole ingress",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-invalid-path-strict-validate.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
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
			desc: "External name service",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-external-name.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-external-name-26-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("external.localhost") && Path("/external-with-matching-port")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-external-name-26-rule-0-path-0-retry"},
							Service:     "default-7-ingress-with-external-name-26-external-8-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-external-name",
									},
								},
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("external.localhost") && Path("/external-with-matching-port")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-external-name-26-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-external-name-26-external-8-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-external-name",
									},
								},
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-1": {
							EntryPoints: []string{"http"},
							Rule:        `Host("external.localhost") && Path("/external-with-matching-named-port")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-external-name-26-rule-0-path-1-retry"},
							Service:     "default-7-ingress-with-external-name-26-external-8-http-4",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-external-name",
									},
								},
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-1-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("external.localhost") && Path("/external-with-matching-named-port")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-external-name-26-rule-0-path-1-tls-retry"},
							Service:     "default-7-ingress-with-external-name-26-external-8-http-4",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-external-name",
									},
								},
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-2": {
							EntryPoints: []string{"http"},
							Rule:        `Host("external.localhost") && Path("/external-with-non-matching-port")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-external-name-26-rule-0-path-2-retry"},
							Service:     "default-7-ingress-with-external-name-26-external-8-3000-4",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-external-name",
									},
								},
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-2-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("external.localhost") && Path("/external-with-non-matching-port")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-external-name-26-rule-0-path-2-tls-retry"},
							Service:     "default-7-ingress-with-external-name-26-external-8-3000-4",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-external-name",
									},
								},
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-3": {
							EntryPoints: []string{"http"},
							Rule:        `Host("external.localhost") && Path("/external-with-non-matching-named-port")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-external-name-26-rule-0-path-3-retry"},
							Service:     "default-7-ingress-with-external-name-26-external-8-foo-3",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-external-name",
									},
								},
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-3-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("external.localhost") && Path("/external-with-non-matching-named-port")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-external-name-26-rule-0-path-3-tls-retry"},
							Service:     "default-7-ingress-with-external-name-26-external-8-foo-3",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-external-name",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-external-name-26-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-1-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-1-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-2-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-2-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-3-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-external-name-26-rule-0-path-3-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-external-name-26-external-8-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://external.com:8080",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-external-name-26",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "external",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-external-name-26-external-8-3000-4": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://external.com:3000",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-external-name-26",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "external",
										Port:      "3000",
									},
								},
							},
						},
						"default-7-ingress-with-external-name-26-external-8-http-4": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://external.com:8080",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-external-name-26",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "external",
										Port:      "http",
									},
								},
							},
						},
						"default-7-ingress-with-external-name-26-external-8-foo-3": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://external.com:0",
									},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-external-name-26",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "external",
										Port:      "foo",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-external-name-26": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			// Two TLS sections that reference distinct Secrets must produce two TLS certificate entries,
			// even when the Secrets carry identical PEM content.
			desc: "TLS multiple secrets are not deduplicated by PEM content",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-tls-multi-secrets.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-tls-multi-secrets-30-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("first.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-ingress-with-tls-multi-secrets-30-rule-0-path-0-redirect-scheme",
								"default-7-ingress-with-tls-multi-secrets-30-rule-0-path-0-retry",
							},
							Service: "default-7-ingress-with-tls-multi-secrets-30-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-tls-multi-secrets",
									},
								},
							},
						},
						"default-7-ingress-with-tls-multi-secrets-30-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("first.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-tls-multi-secrets-30-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-tls-multi-secrets-30-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-tls-multi-secrets",
									},
								},
							},
						},
						"default-7-ingress-with-tls-multi-secrets-30-rule-1-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("second.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-ingress-with-tls-multi-secrets-30-rule-1-path-0-redirect-scheme",
								"default-7-ingress-with-tls-multi-secrets-30-rule-1-path-0-retry",
							},
							Service: "default-7-ingress-with-tls-multi-secrets-30-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-tls-multi-secrets",
									},
								},
							},
						},
						"default-7-ingress-with-tls-multi-secrets-30-rule-1-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("second.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-tls-multi-secrets-30-rule-1-path-0-tls-retry"},
							Service:     "default-7-ingress-with-tls-multi-secrets-30-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-tls-multi-secrets",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-tls-multi-secrets-30-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-tls-multi-secrets-30-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-tls-multi-secrets-30-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-tls-multi-secrets-30-rule-1-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-tls-multi-secrets-30-rule-1-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-tls-multi-secrets-30-rule-1-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-tls-multi-secrets-30-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-tls-multi-secrets-30",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-tls-multi-secrets-30": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
				},
			},
		},
		{
			// An ingress with ssl-passthrough must still load TLS certificates from its Spec.TLS section,
			// so the certificate is registered as a default cert even though TCP passthrough takes over the actual routing.
			// The TLS section also enables the SSL redirect on the HTTP router, as it does for regular locations.
			desc: "SSL Passthrough still loads TLS section certificates",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-ssl-passthrough-and-tls-section.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-7-ingress-with-ssl-passthrough-and-tls-section-44-passthrough-tls.whoami.localhost-32": {
							EntryPoints: []string{"https"},
							Rule:        `HostSNI("passthrough-tls.whoami.localhost")`,
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
							Service: "default-7-whoami-tls-10-443-3",
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-7-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{Address: "10.10.0.3:8443"},
									{Address: "10.10.0.4:8443"},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-ssl-passthrough-and-tls-section-44-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("passthrough-tls.whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-ssl-passthrough-and-tls-section-44-rule-0-path-0-redirect-scheme", "default-7-ingress-with-ssl-passthrough-and-tls-section-44-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-ssl-passthrough-and-tls-section",
									},
								},
							},
							Service: "default-7-ingress-with-ssl-passthrough-and-tls-section-44-whoami-tls-10-443-3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-ssl-passthrough-and-tls-section-44-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-ssl-passthrough-and-tls-section-44-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-ssl-passthrough-and-tls-section-44-whoami-tls-10-443-3": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: "wrr",
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.3:8443"},
									{URL: "http://10.10.0.4:8443"},
								},
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-ssl-passthrough-and-tls-section-44",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-tls",
										Port:      "443",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-ssl-passthrough-and-tls-section-44": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
				},
			},
		},
		{
			// When auth-tls-pass-certificate-to-upstream is enabled together with auth-tls-verify-client=optional_no_ca (RequestClientCert),
			// the auth-tls-secret's CA bytes must be forwarded to the upstream via the AuthTLSPassCertificateToUpstream middleware's CAFiles field.
			desc: "Auth TLS pass certificate to upstream with optional_no_ca populates CAFiles",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("auth-tls-pass-cert-optional-no-ca.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-rule-0-path-0-redirect-scheme",
								"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-rule-0-path-0-pass-certificate-to-upstream",
								"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-rule-0-path-0-retry",
							},
							Service: "default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-whoami-6-80-2",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca",
									},
								},
							},
						},
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("auth-tls-pass-cert-optional-no-ca.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-rule-0-path-0-tls-pass-certificate-to-upstream", "default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-rule-0-path-0-tls-retry"},
							Service:     "default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-whoami-6-80-2",
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-7-ca-secret-9-request-client-cert",
							},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-rule-0-path-0-pass-certificate-to-upstream": {
							AuthTLSPassCertificateToUpstream: &dynamic.AuthTLSPassCertificateToUpstream{
								ClientAuthType: tls.RequestClientCert,
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----"},
							},
						},
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-rule-0-path-0-tls-pass-certificate-to-upstream": {
							AuthTLSPassCertificateToUpstream: &dynamic.AuthTLSPassCertificateToUpstream{
								ClientAuthType: tls.RequestClientCert,
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----"},
							},
						},
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-auth-tls-pass-certificate-to-upstream-optional-no-ca-65": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
					Options: map[string]tls.Options{
						"default-7-ca-secret-9-request-client-cert": {
							ClientAuth: tls.ClientAuth{
								CAFiles:        []types.FileOrContent{"-----BEGIN CERTIFICATE-----"},
								ClientAuthType: tls.RequestClientCert,
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
							ALPNProtocols: []string{"h2", "http/1.1", tlsalpn01.ACMETLS1Protocol},
						},
					},
				},
			},
		},
		{
			// When an ingress combines custom-http-errors + a per-ingress default-backend annotation + upstream-hash-by,
			// the per-router error-backend service must inherit the upstream-hash-by configuration (HRW strategy + NginxUpstreamHashBy)
			// from the ingress location config — not fall back to the default WRR strategy.
			desc: "Custom HTTP errors error-backend service inherits upstream-hash-by",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-custom-http-errors-and-upstream-hash-by.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-whoami-6-80-2",
							Middlewares: []string{
								"default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0-custom-http-errors",
								"default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0-retry",
							},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-http-errors-and-upstream-hash-by",
									},
								},
							},
						},
						"default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-whoami-6-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{
								"default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0-tls-custom-http-errors",
								"default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0-tls-retry",
							},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-custom-http-errors-and-upstream-hash-by",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "500"},
								Service: "default-backend-default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0",
								NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-custom-http-errors-and-upstream-hash-by"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0-tls-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "500"},
								Service: "default-backend-default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0-tls",
								NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-custom-http-errors-and-upstream-hash-by"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:            dynamic.BalancerStrategyHRW,
								NginxUpstreamHashBy: "$request_uri",
								PassHostHeader:      new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52",
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-backend-default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.5:8000"},
									{URL: "http://10.10.0.6:8000"},
								},
								Strategy:            dynamic.BalancerStrategyHRW,
								NginxUpstreamHashBy: "$request_uri",
								PassHostHeader:      new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-b",
										Port:      "0",
									},
								},
							},
						},
						"default-backend-default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52-rule-0-path-0-tls": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.5:8000"},
									{URL: "http://10.10.0.6:8000"},
								},
								Strategy:            dynamic.BalancerStrategyHRW,
								NginxUpstreamHashBy: "$request_uri",
								PassHostHeader:      new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-b",
										Port:      "0",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-custom-http-errors-and-upstream-hash-by-52": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Ingress with endpoint conditions",
			paths: []string{
				"ingressclasses.yml",
				"ingresses/ingress-with-endpoint-conditions.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-endpoint-conditions-32-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-endpoint-conditions-32-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-endpoint-conditions-32-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-endpoint-conditions",
									},
								},
							},
						},
						"default-7-ingress-with-endpoint-conditions-32-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-endpoint-conditions-32-whoami-6-80-2",
							Middlewares: []string{"default-7-ingress-with-endpoint-conditions-32-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-endpoint-conditions",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-endpoint-conditions-32-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-endpoint-conditions-32-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-endpoint-conditions-32-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80", Fenced: true},
								},
								Strategy:         dynamic.BalancerStrategyWRR,
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-endpoint-conditions-32",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-endpoint-conditions-32": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Disable non-TLS routers",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-disable-http-entrypoint.yml",
				"secrets.yml",
			},
			disableNonTLSRouters: true,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-disable-http-entrypoint-31-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-disable-http-entrypoint-31-service1-8-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-7-ingress-disable-http-entrypoint-31-rule-0-path-0-tls-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-disable-http-entrypoint",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-disable-http-entrypoint-31-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-disable-http-entrypoint-31-service1-8-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "service1",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-disable-http-entrypoint-31": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Disable non-TLS routers with limit allowlist",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-disable-http-entrypoint-limit-allowlist.yml",
				"secrets.yml",
			},
			disableNonTLSRouters: true,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-disable-http-entrypoint-limit-allowlist-47-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-disable-http-entrypoint-limit-allowlist-47-service1-8-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-7-ingress-disable-http-entrypoint-limit-allowlist-47-rule-0-path-0-tls-limit-rps", "default-7-ingress-disable-http-entrypoint-limit-allowlist-47-rule-0-path-0-tls-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-disable-http-entrypoint-limit-allowlist",
									},
								},
							},
						},
						"default-7-ingress-disable-http-entrypoint-limit-allowlist-47-rule-0-path-0-tls-limit-allowlist": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("whoami.localhost") && PathPrefix("/")) && ClientIP("10.0.0.0/24")`,
							RuleSyntax:  "default",
							Priority:    44,
							Service:     "default-7-ingress-disable-http-entrypoint-limit-allowlist-47-service1-8-80-2",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-7-ingress-disable-http-entrypoint-limit-allowlist-47-rule-0-path-0-tls-limit-allowlist-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-disable-http-entrypoint-limit-allowlist",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-disable-http-entrypoint-limit-allowlist-47-rule-0-path-0-tls-limit-rps": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Second),
							},
						},
						"default-7-ingress-disable-http-entrypoint-limit-allowlist-47-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-disable-http-entrypoint-limit-allowlist-47-rule-0-path-0-tls-limit-allowlist-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-disable-http-entrypoint-limit-allowlist-47-service1-8-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "service1",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-disable-http-entrypoint-limit-allowlist-47": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Disable non-TLS routers with canary by header",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingresses-with-canary-by-header.yml",
			},
			disableNonTLSRouters: true,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-canary-by-header-29-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-29-whoami-6-80-2-wrr",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-29-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-7-ingress-with-canary-by-header-29-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always"))`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-canary-by-header-29-whoami-6-80-2-canary",
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-canary-by-header",
									},
								},
							},
							Middlewares: []string{"default-7-ingress-with-canary-by-header-29-rule-0-path-0-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-canary-by-header-29-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-canary-by-header-29-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-7-ingress-with-canary-by-header-29-whoami-6-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.1:80",
									},
									{
										URL: "http://10.10.0.2:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-header-29",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-header-29-whoami-6-80-2-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:80",
									},
									{
										URL: "http://10.10.0.8:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-canary-by-header-29",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "canary",
										Port:      "80",
									},
								},
							},
						},
						"default-7-ingress-with-canary-by-header-29-whoami-6-80-2-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-7-ingress-with-canary-by-header-29-whoami-6-80-2",
										Weight: new(100),
									},
									{
										Name:   "default-7-ingress-with-canary-by-header-29-whoami-6-80-2-canary",
										Weight: new(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-canary-by-header-29": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                 "Disable non-TLS routers with default backend",
			disableNonTLSRouters: true,
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-default-backend-no-rules.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
							Middlewares: []string{"default-backend-tls-cors", "default-backend-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-default-backend-no-rules",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-backend-tls-cors": {
							Headers: &dynamic.Headers{
								AccessControlAllowCredentials: true,
								AccessControlExposeHeaders:    []string{},
								AccessControlAllowHeaders:     []string{"DNT", "Keep-Alive", "User-Agent", "X-Requested-With", "If-Modified-Since", "Cache-Control", "Content-Type", "Range", "Authorization"},
								AccessControlAllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH", "OPTIONS"},
								AccessControlAllowOriginList:  []string{"*"},
								AccessControlMaxAge:           new(int64(1728000)),
							},
						},
						"default-backend-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   new(true),
								ServersTransport: "default-7-ingress-with-default-backend-no-rules-37",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami",
										Port:      "80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-default-backend-no-rules-37": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Auth TLS secret missing — ingress is skipped entirely",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/ingress-with-auth-tls-secret-missing.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       "wrr",
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Certificates: []*tls.CertAndStores{
						{
							CertFile: "-----BEGIN CERTIFICATE-----",
							KeyFile:  "-----BEGIN CERTIFICATE-----",
						},
					},
				},
			},
		},
		{
			desc: "EndpointSlice with nil port name",
			paths: []string{
				"ingressclasses.yml",
				"services.yml",
				"ingresses/ingress-with-nil-endpointslice-port-name.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-nil-endpointslice-port-name-40-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-nil-endpointslice-port-name-40-whoami-nil-port-name-20-80-2",
							Middlewares: []string{"default-7-ingress-with-nil-endpointslice-port-name-40-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-nil-endpointslice-port-name",
									},
								},
							},
						},
						"default-7-ingress-with-nil-endpointslice-port-name-40-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-nil-endpointslice-port-name-40-whoami-nil-port-name-20-80-2",
							Middlewares: []string{"default-7-ingress-with-nil-endpointslice-port-name-40-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-nil-endpointslice-port-name",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-nil-endpointslice-port-name-40-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-nil-endpointslice-port-name-40-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-nil-endpointslice-port-name-40-whoami-nil-port-name-20-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-nil-port-name",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-nil-endpointslice-port-name-40": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "EndpointSlice with nil port value",
			paths: []string{
				"ingressclasses.yml",
				"services.yml",
				"ingresses/ingress-with-nil-endpointslice-port-value.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-7-ingress-with-nil-endpointslice-port-value-41-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-nil-endpointslice-port-value-41-whoami-nil-port-value-21-80-2",
							Middlewares: []string{"default-7-ingress-with-nil-endpointslice-port-value-41-rule-0-path-0-retry"},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-nil-endpointslice-port-value",
									},
								},
							},
						},
						"default-7-ingress-with-nil-endpointslice-port-value-41-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-7-ingress-with-nil-endpointslice-port-value-41-whoami-nil-port-value-21-80-2",
							Middlewares: []string{"default-7-ingress-with-nil-endpointslice-port-value-41-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								Metadata: &dynamic.ObservabilityMetadata{
									Ingress: &dynamic.KubernetesMetadata{
										Kind:      "Ingress",
										Namespace: "default",
										Name:      "ingress-with-nil-endpointslice-port-value",
									},
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-7-ingress-with-nil-endpointslice-port-value-41-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
						"default-7-ingress-with-nil-endpointslice-port-value-41-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:            3,
								MaxRequestBodyBytes: new(defaultProxyBodySize),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-7-ingress-with-nil-endpointslice-port-value-41-whoami-nil-port-value-21-80-2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
							Observability: &dynamic.ServiceObservabilityConfig{
								Metadata: &dynamic.ServiceObservabilityMetadata{
									Kubernetes: &dynamic.KubernetesServiceMetadata{
										Namespace: "default",
										Name:      "whoami-nil-port-value",
										Port:      "80",
									},
								},
							},
						},
						"unavailable-service": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								PassHostHeader: new(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-7-ingress-with-nil-endpointslice-port-value-41": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			k8sObjects := readResources(t, test.paths)
			kubeClient := kubefake.NewClientset(k8sObjects...)
			client := newClient(kubeClient)

			eventCh, err := client.WatchAll(t.Context(), "", "")
			require.NoError(t, err)

			if len(k8sObjects) > 0 {
				// just wait for the first event
				<-eventCh
			}

			nonTLSEntryPoints := []string{"http"}
			if test.disableNonTLSRouters {
				nonTLSEntryPoints = []string{}
			}

			p := Provider{
				k8sClient:                      client,
				defaultBackendServiceName:      test.defaultBackendServiceName,
				defaultBackendServiceNamespace: test.defaultBackendServiceNamespace,
				AllowSnippetAnnotations:        test.allowSnippetAnnotations,
				NonTLSEntryPoints:              nonTLSEntryPoints,
				DisableNonTLSRouters:           test.disableNonTLSRouters,
				TLSEntryPoints:                 []string{"https"},
				allowedHeaders:                 test.globalAllowedResponseHeaders,
				IPAllowListStrategy:            test.ipAllowListStrategy,
				AllowCrossNamespaceResources:   test.allowCrossNamespaceResources,
				GlobalAuthURL:                  test.globalAuthURL,
				ProxyRequestBuffering:          test.proxyRequestBuffering,
			}
			p.SetDefaults()
			if test.strictValidatePathType != nil {
				p.StrictValidatePathType = *test.strictValidatePathType
			}

			conf := p.loadConfiguration(t.Context())
			assert.Equal(t, test.expected, conf)

			for name, middleware := range conf.HTTP.Middlewares {
				_, err := middleware.ResolveURNRefs()
				assert.NoError(t, err, name)
			}
		})
	}
}

func TestNginxSizeToBytes(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		err      error
		expected int64
	}{
		{
			desc:     "Testing no unit",
			expected: 100,
			value:    "100",
		},
		{
			desc:     "Testing unit b",
			expected: 100,
			value:    "100b",
		},
		{
			desc:     "Testing unit B",
			expected: 100,
			value:    "100B",
		},
		{
			desc:     "Testing unit KB",
			expected: 100 * 1024,
			value:    "100k",
		},
		{
			desc:     "Testing unit MB",
			expected: 100 * 1024 * 1024,
			value:    "100m",
		},
		{
			desc:     "Testing unit GB",
			expected: 100 * 1024 * 1024 * 1024,
			value:    "100g",
		},
		{
			desc:     "Testing unit GB with whitespaces",
			expected: 100 * 1024 * 1024 * 1024,
			value:    " 100 g ",
		},
		{
			desc:     "Testing unit KB uppercase",
			expected: 100 * 1024,
			value:    "100K",
		},
		{
			desc:     "Testing unit MB uppercase",
			expected: 100 * 1024 * 1024,
			value:    "100M",
		},
		{
			desc:     "Testing unit GB uppercase",
			expected: 100 * 1024 * 1024 * 1024,
			value:    "100G",
		},
		{
			desc:     "Testing invalid input",
			expected: 0,
			value:    "100A",
			err:      errors.New("unable to parse number 100A"),
		},
		{
			desc:     "Testing pipe character is invalid",
			expected: 0,
			value:    "100|",
			err:      errors.New("unable to parse number 100|"),
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			size, err := nginxSizeToBytes(test.value)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.expected, size)
		})
	}
}

func TestLoadConfigurationIngressStatus(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		desc         string
		paths        []string
		ingressNames []string
		wantStatus   bool
	}{
		{
			desc:         "SSL passthrough ingress",
			paths:        []string{"ingresses/ingress-with-ssl-passthrough.yml"},
			ingressNames: []string{"ingress-with-ssl-passthrough"},
			wantStatus:   true,
		},
		{
			desc:         "normal ingress",
			paths:        []string{"ingresses/ingress-with-host.yml"},
			ingressNames: []string{"ingress-with-host"},
			wantStatus:   true,
		},
		{
			desc:         "canary ingress",
			paths:        []string{"ingresses/ingresses-with-canary.yml"},
			ingressNames: []string{"ingress-with-canary", "canary"},
			wantStatus:   true,
		},
		{
			desc:         "skipped ingress",
			paths:        []string{"secrets.yml", "ingresses/ingress-with-auth-tls-secret-missing.yml"},
			ingressNames: []string{"ingress-with-auth-tls-secret-missing"},
		},
		{
			desc:         "SSL passthrough ingress without a root path",
			paths:        []string{"ingresses/ingress-with-ssl-passthrough-no-root.yml"},
			ingressNames: []string{"ingress-with-ssl-passthrough-no-root"},
			wantStatus:   true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			paths := append([]string{"services.yml", "ingressclasses.yml"}, test.paths...)
			k8sObjects := readResources(t, paths)
			kubeClient := kubefake.NewClientset(k8sObjects...)
			client := newClient(kubeClient)

			eventCh, err := client.WatchAll(t.Context(), "", "")
			require.NoError(t, err)
			<-eventCh

			p := Provider{
				PublishStatusAddress: []string{"203.0.113.10"},
				k8sClient:            client,
				NonTLSEntryPoints:    []string{"http"},
				TLSEntryPoints:       []string{"https"},
			}
			p.SetDefaults()

			require.NotNil(t, p.loadConfiguration(t.Context()))

			for _, name := range test.ingressNames {
				ing, err := kubeClient.NetworkingV1().Ingresses("default").Get(t.Context(), name, metav1.GetOptions{})
				require.NoError(t, err)

				if test.wantStatus {
					assert.Equal(t, []netv1.IngressLoadBalancerIngress{{IP: "203.0.113.10"}}, ing.Status.LoadBalancer.Ingress)
				} else {
					assert.Empty(t, ing.Status.LoadBalancer.Ingress)
				}
			}
		})
	}
}

func controllerPod(name, nodeName string, ready bool) *corev1.Pod {
	status := corev1.ConditionFalse
	if ready {
		status = corev1.ConditionTrue
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default", Labels: map[string]string{"app": "traefik"}},
		Spec:       corev1.PodSpec{NodeName: nodeName},
		Status: corev1.PodStatus{
			Phase:      corev1.PodRunning,
			Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: status}},
		},
	}
}

func node(name, internalIP, externalIP string) *corev1.Node {
	addresses := []corev1.NodeAddress{{Type: corev1.NodeInternalIP, Address: internalIP}}
	if externalIP != "" {
		addresses = append(addresses, corev1.NodeAddress{Type: corev1.NodeExternalIP, Address: externalIP})
	}

	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status:     corev1.NodeStatus{Addresses: addresses},
	}
}

func TestLoadConfigurationIngressStatusFromNodes(t *testing.T) {
	testCases := []struct {
		desc           string
		internalIP     bool
		podName        string
		expectedStatus []netv1.IngressLoadBalancerIngress
	}{
		{
			desc:    "external node addresses, with the internal one as a fallback",
			podName: "traefik-controller-1",
			expectedStatus: []netv1.IngressLoadBalancerIngress{
				{IP: "203.0.113.1"},
				{IP: "10.0.0.2"},
			},
		},
		{
			desc:       "internal node addresses",
			internalIP: true,
			podName:    "traefik-controller-1",
			expectedStatus: []netv1.IngressLoadBalancerIngress{
				{IP: "10.0.0.1"},
				{IP: "10.0.0.2"},
			},
		},
		{
			desc:    "unknown controller pod",
			podName: "missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Setenv("POD_NAME", test.podName)
			t.Setenv("POD_NAMESPACE", "default")

			k8sObjects := readResources(t, []string{"services.yml", "ingressclasses.yml", "ingresses/ingress-with-host.yml"})
			k8sObjects = append(k8sObjects,
				controllerPod("traefik-controller-1", "node-1", true),
				controllerPod("traefik-controller-2", "node-2", true),
				controllerPod("traefik-controller-3", "node-3", false),
				&corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{Name: "other-controller", Namespace: "default", Labels: map[string]string{"app": "other"}},
					Spec:       corev1.PodSpec{NodeName: "node-4"},
					Status: corev1.PodStatus{
						Phase:      corev1.PodRunning,
						Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}},
					},
				},
				node("node-1", "10.0.0.1", "203.0.113.1"),
				node("node-2", "10.0.0.2", ""),
				node("node-3", "10.0.0.3", "203.0.113.3"),
				node("node-4", "10.0.0.4", "203.0.113.4"),
			)
			kubeClient := kubefake.NewClientset(k8sObjects...)
			client := newClient(kubeClient)

			eventCh, err := client.WatchAll(t.Context(), "", "")
			require.NoError(t, err)
			<-eventCh

			p := Provider{
				ReportNodeInternalIPAddress: test.internalIP,
				k8sClient:                   client,
				NonTLSEntryPoints:           []string{"http"},
				TLSEntryPoints:              []string{"https"},
			}
			p.SetDefaults()

			require.NotNil(t, p.loadConfiguration(t.Context()))

			ing, err := kubeClient.NetworkingV1().Ingresses("default").Get(t.Context(), "ingress-with-host", metav1.GetOptions{})
			require.NoError(t, err)

			assert.ElementsMatch(t, test.expectedStatus, ing.Status.LoadBalancer.Ingress)
		})
	}
}

func TestProvider_validateConfiguration(t *testing.T) {
	testCases := []struct {
		desc                            string
		globalAllowedResponseHeaders    []string
		expectedAllowedHeaders          []string
		defaultBackendService           string
		expectedBackendServiceName      string
		expectedBackendServiceNamespace string
		disableNonTLSRouters            bool
		httpEntryPoint                  string
		expectError                     bool
	}{
		{
			desc:                         "Valid headers only",
			globalAllowedResponseHeaders: []string{"X-Custom-Header", "X-Another-Header", "Content-Type"},
			expectedAllowedHeaders:       []string{"X-Custom-Header", "X-Another-Header", "Content-Type"},
		},
		{
			desc:                         "Mixed valid and invalid headers",
			globalAllowedResponseHeaders: []string{"X-Custom-Header", "Invalid Header With Spaces", "X-Valid_Header-123"},
			expectedAllowedHeaders:       []string{"X-Custom-Header", "X-Valid_Header-123"},
		},
		{
			desc:                         "All invalid headers",
			globalAllowedResponseHeaders: []string{"Invalid Header", "Another Bad Header!", "@#$%"},
			expectedAllowedHeaders:       nil,
		},
		{
			desc:                         "Empty list",
			globalAllowedResponseHeaders: []string{},
			expectedAllowedHeaders:       nil,
		},
		{
			desc:                            "Valid default backend service",
			defaultBackendService:           "namespace/name",
			expectedBackendServiceName:      "name",
			expectedBackendServiceNamespace: "namespace",
		},
		{
			desc:                  "Invalid default backend service",
			defaultBackendService: "namespace",
			expectError:           true,
		},
		{
			desc:                 "Disable non-TLS routers with httpEntryPoint set",
			disableNonTLSRouters: true,
			httpEntryPoint:       "web",
			expectError:          true,
		},
		{
			desc:                 "Disable non-TLS routers without httpEntryPoint",
			disableNonTLSRouters: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			provider := &Provider{
				GlobalAllowedResponseHeaders: test.globalAllowedResponseHeaders,
				DefaultBackendService:        test.defaultBackendService,
				DisableNonTLSRouters:         test.disableNonTLSRouters,
				HTTPEntryPoint:               test.httpEntryPoint,
			}

			err := provider.validateConfiguration()
			if test.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectedAllowedHeaders, provider.allowedHeaders)
			assert.Equal(t, test.expectedBackendServiceName, provider.defaultBackendServiceName)
			assert.Equal(t, test.expectedBackendServiceNamespace, provider.defaultBackendServiceNamespace)
		})
	}
}

func TestHeaderRegexp(t *testing.T) {
	testCases := []struct {
		desc     string
		header   string
		expected bool
	}{
		{
			desc:     "Valid header with alphanumeric characters",
			header:   "X-Custom-Header",
			expected: true,
		},
		{
			desc:     "Valid header with underscores",
			header:   "X_Custom_Header",
			expected: true,
		},
		{
			desc:     "Valid header with numbers",
			header:   "Header123",
			expected: true,
		},
		{
			desc:     "Valid header with mixed case",
			header:   "Content-Type",
			expected: true,
		},
		{
			desc:     "Invalid header with spaces",
			header:   "Invalid Header",
			expected: false,
		},
		{
			desc:     "Invalid header with special characters",
			header:   "Header@#$",
			expected: false,
		},
		{
			desc:     "Invalid header with dots",
			header:   "Header.Name",
			expected: false,
		},
		{
			desc:     "Empty header",
			header:   "",
			expected: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := headerRegexp.MatchString(test.header)
			assert.Equal(t, test.expected, result, "Header: %q", test.header)
		})
	}
}

func TestHeaderValueRegexp(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected bool
	}{
		{
			desc:     "Simple string value",
			value:    "simple-value",
			expected: true,
		},
		{
			desc:     "Value with spaces",
			value:    "value with spaces",
			expected: true,
		},
		{
			desc:     "Value with various allowed characters",
			value:    "value:with;various,allowed.characters/and\\backslash\"quotes'single?!(){}[]@<>=+-*#$&`|~^%",
			expected: true,
		},
		{
			desc:     "Value with newline",
			value:    "value\nwith\nnewline",
			expected: false,
		},
		{
			desc:     "Value with tab",
			value:    "value\twith\ttab",
			expected: false,
		},
		{
			desc:     "Empty value",
			value:    "",
			expected: false,
		},
		{
			desc:     "Value with unicode",
			value:    "value-with-unicode-€",
			expected: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			result := headerValueRegexp.MatchString(test.value)
			assert.Equal(t, test.expected, result, "Value: %q", test.value)
		})
	}
}

func readResources(t *testing.T, paths []string) []runtime.Object {
	t.Helper()

	var k8sObjects []runtime.Object
	for _, path := range paths {
		yamlContent, err := os.ReadFile(filepath.FromSlash("./fixtures/" + path))
		if err != nil {
			panic(err)
		}

		k8sObjects = append(k8sObjects, k8s.MustParseYaml(yamlContent)...)
	}

	return k8sObjects
}
