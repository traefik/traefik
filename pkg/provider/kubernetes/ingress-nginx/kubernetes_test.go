package ingressnginx

import (
	"errors"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-acme/lego/v4/challenge/tlsalpn01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/utils/ptr"
)

func TestLoadIngresses(t *testing.T) {
	testCases := []struct {
		desc                           string
		ingressClass                   string
		defaultBackendServiceName      string
		defaultBackendServiceNamespace string
		allowCrossNamespaceResources   bool
		globalAllowedResponseHeaders   []string
		allowSnippetAnnotations        bool
		globalAuthURL                  string
		strictValidatePathType         *bool
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
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
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
						"default-ingress-with-custom-headers-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-custom-headers-rule-0-path-0-custom-headers", "default-ingress-with-custom-headers-rule-0-path-0-retry"},
							Service:     "default-ingress-with-custom-headers-whoami-80",
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("cross-namespace.localhost") && Path("/")`,
							Middlewares: []string{"default-ingress-with-cross-namespace-headers-rule-0-path-0-custom-headers", "default-ingress-with-cross-namespace-headers-rule-0-path-0-retry"},
							RuleSyntax:  "default",
							Service:     "default-ingress-with-cross-namespace-headers-whoami-80",
						},
						"default-ingress-with-custom-headers-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-custom-headers-rule-0-path-0-tls-custom-headers", "default-ingress-with-custom-headers-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-custom-headers-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("cross-namespace.localhost") && Path("/")`,
							Middlewares: []string{"default-ingress-with-cross-namespace-headers-rule-0-path-0-tls-custom-headers", "default-ingress-with-cross-namespace-headers-rule-0-path-0-tls-retry"},
							RuleSyntax:  "default",
							Service:     "default-ingress-with-cross-namespace-headers-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-custom-headers-rule-0-path-0-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-ingress-with-custom-headers-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Cross-Header": "cross-value"},
							},
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-custom-headers-rule-0-path-0-tls-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-ingress-with-custom-headers-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0-tls-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Cross-Header": "cross-value"},
							},
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-custom-headers-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-custom-headers",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-cross-namespace-headers-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-cross-namespace-headers",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-custom-headers": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-cross-namespace-headers": {
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
						"default-ingress-with-custom-headers-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-custom-headers-rule-0-path-0-custom-headers", "default-ingress-with-custom-headers-rule-0-path-0-retry"},
							Service:     "default-ingress-with-custom-headers-whoami-80",
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("cross-namespace.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-cross-namespace-headers-rule-0-path-0-custom-headers", "default-ingress-with-cross-namespace-headers-rule-0-path-0-retry"},
							Service:     "default-ingress-with-cross-namespace-headers-whoami-80",
						},
						"default-ingress-with-custom-headers-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-custom-headers-rule-0-path-0-tls-custom-headers", "default-ingress-with-custom-headers-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-custom-headers-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("cross-namespace.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-cross-namespace-headers-rule-0-path-0-tls-custom-headers", "default-ingress-with-cross-namespace-headers-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-cross-namespace-headers-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-custom-headers-rule-0-path-0-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-ingress-with-custom-headers-rule-0-path-0-tls-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-ingress-with-custom-headers-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-custom-headers-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Cross-Header": "cross-value"},
							},
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0-tls-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Cross-Header": "cross-value"},
							},
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-custom-headers-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-custom-headers",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-cross-namespace-headers-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-cross-namespace-headers",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-custom-headers": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-cross-namespace-headers": {
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
			globalAllowedResponseHeaders: []string{"X-Custom-Header"},
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
						"default-ingress-with-custom-headers-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-custom-headers-rule-0-path-0-custom-headers", "default-ingress-with-custom-headers-rule-0-path-0-retry"},
							Service:     "default-ingress-with-custom-headers-whoami-80",
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("cross-namespace.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-cross-namespace-headers-whoami-80",
						},
						"default-ingress-with-custom-headers-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-custom-headers-rule-0-path-0-tls-custom-headers", "default-ingress-with-custom-headers-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-custom-headers-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-cross-namespace-headers-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("cross-namespace.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-cross-namespace-headers-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-custom-headers-rule-0-path-0-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-ingress-with-custom-headers-rule-0-path-0-tls-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
							},
						},
						"default-ingress-with-custom-headers-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-custom-headers-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-custom-headers-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-custom-headers",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-cross-namespace-headers-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-cross-namespace-headers",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-custom-headers": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-cross-namespace-headers": {
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
						"default-ingress-with-no-annotation-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-ingress-with-no-annotation-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-no-annotation-whoami-80",
						},
						"default-ingress-with-no-annotation-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-no-annotation-rule-0-path-0-retry"},
							Service:     "default-ingress-with-no-annotation-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-no-annotation-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-no-annotation-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-no-annotation-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-no-annotation",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-no-annotation": {
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
							Certificate: tls.Certificate{
								CertFile: "-----BEGIN CERTIFICATE-----",
								KeyFile:  "-----BEGIN CERTIFICATE-----",
							},
						},
					},
				},
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
						"default-ingress-with-basicauth-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/basicauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-basicauth-rule-0-path-0-basic-auth", "default-ingress-with-basicauth-rule-0-path-0-retry"},
							Service:     "default-ingress-with-basicauth-whoami-80",
						},
						"default-ingress-with-basicauth-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/basicauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-basicauth-rule-0-path-0-tls-basic-auth", "default-ingress-with-basicauth-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-basicauth-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-basicauth-rule-0-path-0-basic-auth": {
							BasicAuth: &dynamic.BasicAuth{
								Users: dynamic.Users{
									"user:{SHA}W6ph5Mm5Pz8GgiULbPgzG37mj9g=",
								},
								Realm: "Authentication Required",
							},
						},
						"default-ingress-with-basicauth-rule-0-path-0-tls-basic-auth": {
							BasicAuth: &dynamic.BasicAuth{
								Users: dynamic.Users{
									"user:{SHA}W6ph5Mm5Pz8GgiULbPgzG37mj9g=",
								},
								Realm: "Authentication Required",
							},
						},
						"default-ingress-with-basicauth-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-basicauth-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-basicauth-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-basicauth",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-basicauth": {
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
						"default-ingress-with-forwardauth-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/forwardauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-forwardauth-rule-0-path-0-snippet", "default-ingress-with-forwardauth-rule-0-path-0-retry"},
							Service:     "default-ingress-with-forwardauth-whoami-80",
						},
						"default-ingress-with-forwardauth-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/forwardauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-forwardauth-rule-0-path-0-tls-snippet", "default-ingress-with-forwardauth-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-forwardauth-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-forwardauth-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address:             "http://whoami.default.svc/",
									Method:              http.MethodGet,
									AuthResponseHeaders: []string{"X-Foo", "X-Bar"},
									AuthSigninURL:       "https://auth.example.com/oauth2/start?rd=foo",
								},
							},
						},
						"default-ingress-with-forwardauth-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address:             "http://whoami.default.svc/",
									Method:              http.MethodGet,
									AuthResponseHeaders: []string{"X-Foo", "X-Bar"},
									AuthSigninURL:       "https://auth.example.com/oauth2/start?rd=foo",
								},
							},
						},
						"default-ingress-with-forwardauth-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-forwardauth-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-forwardauth-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-forwardauth",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-forwardauth": {
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
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
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
						"default-ingress-with-forwardauth-snippet-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/forwardauth-snippet")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-forwardauth-snippet-rule-0-path-0-snippet", "default-ingress-with-forwardauth-snippet-rule-0-path-0-retry"},
							Service:     "default-ingress-with-forwardauth-snippet-whoami-80",
						},
						"default-ingress-with-forwardauth-snippet-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/forwardauth-snippet")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-forwardauth-snippet-rule-0-path-0-tls-snippet", "default-ingress-with-forwardauth-snippet-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-forwardauth-snippet-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-forwardauth-snippet-rule-0-path-0-snippet": {
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
						"default-ingress-with-forwardauth-snippet-rule-0-path-0-tls-snippet": {
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
						"default-ingress-with-forwardauth-snippet-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-forwardauth-snippet-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-forwardauth-snippet-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-forwardauth-snippet",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-forwardauth-snippet": {
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
						"default-ingress-without-auth-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-without-auth-rule-0-path-0-snippet", "default-ingress-without-auth-rule-0-path-0-retry"},
							Service:     "default-ingress-without-auth-whoami-80",
						},
						"default-ingress-without-auth-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-without-auth-rule-0-path-0-tls-snippet", "default-ingress-without-auth-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-without-auth-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-without-auth-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address: "http://auth.example.com/verify",
								},
							},
						},
						"default-ingress-without-auth-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address: "http://auth.example.com/verify",
								},
							},
						},
						"default-ingress-without-auth-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
						"default-ingress-without-auth-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-without-auth-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-without-auth",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-without-auth": {
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
						"default-ingress-with-global-auth-disabled-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-global-auth-disabled-rule-0-path-0-retry"},
							Service:     "default-ingress-with-global-auth-disabled-whoami-80",
						},
						"default-ingress-with-global-auth-disabled-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-global-auth-disabled-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-global-auth-disabled-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-global-auth-disabled-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
						"default-ingress-with-global-auth-disabled-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-global-auth-disabled-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-global-auth-disabled",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-global-auth-disabled": {
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
						"default-ingress-with-forwardauth-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/forwardauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-forwardauth-rule-0-path-0-snippet", "default-ingress-with-forwardauth-rule-0-path-0-retry"},
							Service:     "default-ingress-with-forwardauth-whoami-80",
						},
						"default-ingress-with-forwardauth-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/forwardauth")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-forwardauth-rule-0-path-0-tls-snippet", "default-ingress-with-forwardauth-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-forwardauth-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-forwardauth-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address:             "http://whoami.default.svc/",
									Method:              http.MethodGet,
									AuthResponseHeaders: []string{"X-Foo", "X-Bar"},
									AuthSigninURL:       "https://auth.example.com/oauth2/start?rd=foo",
								},
							},
						},
						"default-ingress-with-forwardauth-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								Auth: &dynamic.Auth{
									Address:             "http://whoami.default.svc/",
									Method:              http.MethodGet,
									AuthResponseHeaders: []string{"X-Foo", "X-Bar"},
									AuthSigninURL:       "https://auth.example.com/oauth2/start?rd=foo",
								},
							},
						},
						"default-ingress-with-forwardauth-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
						"default-ingress-with-forwardauth-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-forwardauth-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-forwardauth",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-forwardauth": {
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
						"default-ingress-without-auth-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-without-auth-rule-0-path-0-retry"},
							Service:     "default-ingress-without-auth-whoami-80",
						},
						"default-ingress-without-auth-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-without-auth-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-without-auth-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-without-auth-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
						"default-ingress-without-auth-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-without-auth-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-without-auth",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-without-auth": {
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
						"default-ingress-with-ssl-redirect-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("sslredirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-ingress-with-ssl-redirect-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-ssl-redirect-whoami-80",
						},
						"default-ingress-with-ssl-redirect-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("sslredirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-ssl-redirect-rule-0-path-0-retry"},
							Service:     "default-ingress-with-ssl-redirect-whoami-80",
						},
						"default-ingress-without-ssl-redirect-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("withoutsslredirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-without-ssl-redirect-rule-0-path-0-retry"},
							Service:     "default-ingress-without-ssl-redirect-whoami-80",
						},
						"default-ingress-without-ssl-redirect-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("withoutsslredirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-ingress-without-ssl-redirect-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-without-ssl-redirect-whoami-80",
						},
						"default-ingress-with-force-ssl-redirect-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("forcesslredirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-force-ssl-redirect-rule-0-path-0-retry"},
							Service:     "default-ingress-with-force-ssl-redirect-whoami-80",
						},
						"default-ingress-with-force-ssl-redirect-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("forcesslredirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-force-ssl-redirect-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-force-ssl-redirect-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-ssl-redirect-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-ssl-redirect-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-without-ssl-redirect-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-without-ssl-redirect-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-force-ssl-redirect-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-force-ssl-redirect-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-ssl-redirect-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-ssl-redirect",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-without-ssl-redirect-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-without-ssl-redirect",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-force-ssl-redirect-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-force-ssl-redirect",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-ssl-redirect": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-without-ssl-redirect": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-force-ssl-redirect": {
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
							Certificate: tls.Certificate{
								CertFile: "-----BEGIN CERTIFICATE-----",
								KeyFile:  "-----BEGIN CERTIFICATE-----",
							},
						},
					},
				},
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
						"default-ingress-with-ssl-passthrough-passthrough-whoami-localhost": {
							EntryPoints: []string{"https"},
							Rule:        `HostSNI("passthrough.whoami.localhost")`,
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
							Service: "default-whoami-tls-443",
						},
					},
					Services: map[string]*dynamic.TCPService{
						"default-whoami-tls-443": {
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
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
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
						"default-ingress-with-sticky-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("sticky.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-sticky-rule-0-path-0-retry"},
							Service:     "default-ingress-with-sticky-whoami-80",
						},
						"default-ingress-with-sticky-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("sticky.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-sticky-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-sticky-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-sticky-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-sticky-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-sticky-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-sticky",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "foobar",
										Domain:   "foo.localhost",
										HTTPOnly: true,
										MaxAge:   42,
										Expires:  42,
										Path:     ptr.To("/foobar"),
										SameSite: "none",
										Secure:   true,
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-sticky": {
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
						"default-ingress-with-proxy-ssl-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("proxy-ssl.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-ssl-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-ssl-whoami-tls-443",
						},
						"default-ingress-with-proxy-ssl-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("proxy-ssl.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-ssl-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-proxy-ssl-whoami-tls-443",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-ssl-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-proxy-ssl-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-ssl-whoami-tls-443": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-ssl",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-ssl": {
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
						"default-ingress-with-cors-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("cors.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-cors-rule-0-path-0-cors", "default-ingress-with-cors-rule-0-path-0-retry"},
							Service:     "default-ingress-with-cors-whoami-80",
						},
						"default-ingress-with-cors-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("cors.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-cors-rule-0-path-0-tls-cors", "default-ingress-with-cors-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-cors-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-cors-rule-0-path-0-cors": {
							Headers: &dynamic.Headers{
								AccessControlAllowCredentials: true,
								AccessControlAllowHeaders:     []string{"X-Foo"},
								AccessControlAllowMethods:     []string{"PUT", "GET", "POST", "OPTIONS"},
								AccessControlAllowOriginList:  []string{"*"},
								AccessControlExposeHeaders:    []string{"X-Forwarded-For", "X-Forwarded-Host"},
								AccessControlMaxAge:           42,
							},
						},
						"default-ingress-with-cors-rule-0-path-0-tls-cors": {
							Headers: &dynamic.Headers{
								AccessControlAllowCredentials: true,
								AccessControlAllowHeaders:     []string{"X-Foo"},
								AccessControlAllowMethods:     []string{"PUT", "GET", "POST", "OPTIONS"},
								AccessControlAllowOriginList:  []string{"*"},
								AccessControlExposeHeaders:    []string{"X-Forwarded-For", "X-Forwarded-Host"},
								AccessControlMaxAge:           42,
							},
						},
						"default-ingress-with-cors-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-cors-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-cors-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-cors",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-cors": {
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
						"default-ingress-with-service-upstream-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("service-upstream.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-service-upstream-rule-0-path-0-retry"},
							Service:     "default-ingress-with-service-upstream-whoami-80",
						},
						"default-ingress-with-service-upstream-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("service-upstream.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-service-upstream-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-service-upstream-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-service-upstream-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-service-upstream-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-service-upstream-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.10.1:80",
									},
								},
								Strategy:         "wrr",
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-service-upstream",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-service-upstream": {
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
						"default-ingress-with-upstream-vhost-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("upstream-vhost.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-upstream-vhost-rule-0-path-0-vhost", "default-ingress-with-upstream-vhost-rule-0-path-0-retry"},
							Service:     "default-ingress-with-upstream-vhost-whoami-80",
						},
						"default-ingress-with-upstream-vhost-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("upstream-vhost.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-upstream-vhost-rule-0-path-0-tls-vhost", "default-ingress-with-upstream-vhost-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-upstream-vhost-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-upstream-vhost-rule-0-path-0-vhost": {
							Headers: &dynamic.Headers{
								CustomRequestHeaders: map[string]string{"Host": "upstream-host-header-value"},
							},
						},
						"default-ingress-with-upstream-vhost-rule-0-path-0-tls-vhost": {
							Headers: &dynamic.Headers{
								CustomRequestHeaders: map[string]string{"Host": "upstream-host-header-value"},
							},
						},
						"default-ingress-with-upstream-vhost-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-upstream-vhost-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-upstream-vhost-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-upstream-vhost",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-upstream-vhost": {
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
						"default-ingress-with-x-forwarded-prefix-no-rewrite-target-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("x-forwarded-prefix.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-x-forwarded-prefix-no-rewrite-target-rule-0-path-0-retry"},
							Service:     "default-ingress-with-x-forwarded-prefix-no-rewrite-target-whoami-80",
						},
						"default-ingress-with-x-forwarded-prefix-no-rewrite-target-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("x-forwarded-prefix.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-x-forwarded-prefix-no-rewrite-target-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-x-forwarded-prefix-no-rewrite-target-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-x-forwarded-prefix-no-rewrite-target-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-x-forwarded-prefix-no-rewrite-target-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-x-forwarded-prefix-no-rewrite-target-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-x-forwarded-prefix-no-rewrite-target",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-x-forwarded-prefix-no-rewrite-target": {
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
						"default-ingress-with-x-forwarded-prefix-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("x-forwarded-prefix.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-x-forwarded-prefix-rule-0-path-0-rewrite-target", "default-ingress-with-x-forwarded-prefix-rule-0-path-0-retry"},
							Service:     "default-ingress-with-x-forwarded-prefix-whoami-80",
						},
						"default-ingress-with-x-forwarded-prefix-regex-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("x-forwarded-prefix-regex.localhost") && PathRegexp("(?i)^/(something)(/.+)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-x-forwarded-prefix-regex-rule-0-path-0-rewrite-target", "default-ingress-with-x-forwarded-prefix-regex-rule-0-path-0-retry"},
							Service:     "default-ingress-with-x-forwarded-prefix-regex-whoami-80",
						},
						"default-ingress-with-x-forwarded-prefix-three-groups-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("x-forwarded-prefix-three-groups.localhost") && PathRegexp("(?i)^/(prefix)/(sub)/(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-x-forwarded-prefix-three-groups-rule-0-path-0-rewrite-target", "default-ingress-with-x-forwarded-prefix-three-groups-rule-0-path-0-retry"},
							Service:     "default-ingress-with-x-forwarded-prefix-three-groups-whoami-80",
						},
						"default-ingress-with-x-forwarded-prefix-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("x-forwarded-prefix.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-x-forwarded-prefix-rule-0-path-0-tls-rewrite-target", "default-ingress-with-x-forwarded-prefix-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-x-forwarded-prefix-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-x-forwarded-prefix-regex-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("x-forwarded-prefix-regex.localhost") && PathRegexp("(?i)^/(something)(/.+)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-x-forwarded-prefix-regex-rule-0-path-0-tls-rewrite-target", "default-ingress-with-x-forwarded-prefix-regex-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-x-forwarded-prefix-regex-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-x-forwarded-prefix-three-groups-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("x-forwarded-prefix-three-groups.localhost") && PathRegexp("(?i)^/(prefix)/(sub)/(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-x-forwarded-prefix-three-groups-rule-0-path-0-tls-rewrite-target", "default-ingress-with-x-forwarded-prefix-three-groups-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-x-forwarded-prefix-three-groups-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-x-forwarded-prefix-three-groups-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:            "(?i)/(prefix)/(sub)/(.*)",
								Replacement:      "/$3",
								XForwardedPrefix: "/$1/$2",
							},
						},
						"default-ingress-with-x-forwarded-prefix-three-groups-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:            "(?i)/(prefix)/(sub)/(.*)",
								Replacement:      "/$3",
								XForwardedPrefix: "/$1/$2",
							},
						},
						"default-ingress-with-x-forwarded-prefix-three-groups-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-x-forwarded-prefix-three-groups-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-x-forwarded-prefix-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:            "(?i)/",
								Replacement:      "/path",
								XForwardedPrefix: "x-forwarded-prefix-header-value",
							},
						},
						"default-ingress-with-x-forwarded-prefix-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:            "(?i)/",
								Replacement:      "/path",
								XForwardedPrefix: "x-forwarded-prefix-header-value",
							},
						},
						"default-ingress-with-x-forwarded-prefix-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-x-forwarded-prefix-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-x-forwarded-prefix-regex-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:            "(?i)/(something)(/.+)",
								Replacement:      "$2",
								XForwardedPrefix: "$1",
							},
						},
						"default-ingress-with-x-forwarded-prefix-regex-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:            "(?i)/(something)(/.+)",
								Replacement:      "$2",
								XForwardedPrefix: "$1",
							},
						},
						"default-ingress-with-x-forwarded-prefix-regex-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-x-forwarded-prefix-regex-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-x-forwarded-prefix-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-x-forwarded-prefix",
							},
						},
						"default-ingress-with-x-forwarded-prefix-three-groups-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-x-forwarded-prefix-three-groups",
							},
						},
						"default-ingress-with-x-forwarded-prefix-regex-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-x-forwarded-prefix-regex",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-x-forwarded-prefix": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-x-forwarded-prefix-three-groups": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-x-forwarded-prefix-regex": {
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
						"default-ingress-with-use-regex-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("use-regex.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-use-regex-rule-0-path-0-retry"},
							Service:     "default-ingress-with-use-regex-whoami-80",
						},
						"default-ingress-with-use-regex-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("use-regex.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-use-regex-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-use-regex-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-use-regex-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-use-regex-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-use-regex-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-use-regex",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-use-regex": {
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
						"default-ingress-a-with-use-regex-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-a-with-use-regex-rule-0-path-0-retry"},
							Service:     "default-ingress-a-with-use-regex-whoami-80",
						},
						"default-ingress-b-without-use-regex-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/static")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-b-without-use-regex-rule-0-path-0-retry"},
							Service:     "default-ingress-b-without-use-regex-whoami-80",
						},
						"default-ingress-a-with-use-regex-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-a-with-use-regex-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-a-with-use-regex-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-b-without-use-regex-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/static")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-b-without-use-regex-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-b-without-use-regex-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-a-with-use-regex-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-a-with-use-regex-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-b-without-use-regex-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-b-without-use-regex-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-a-with-use-regex-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-a-with-use-regex",
							},
						},
						"default-ingress-b-without-use-regex-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-b-without-use-regex",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-a-with-use-regex": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-b-without-use-regex": {
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
						"default-ingress-a-with-use-regex-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("regex.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-a-with-use-regex-rule-0-path-0-retry"},
							Service:     "default-ingress-a-with-use-regex-whoami-80",
						},
						"default-ingress-b-without-use-regex-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("no-regex.localhost") && (Path("/static") || PathPrefix("/static/"))`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-b-without-use-regex-rule-0-path-0-retry"},
							Service:     "default-ingress-b-without-use-regex-whoami-80",
						},
						"default-ingress-a-with-use-regex-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("regex.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-a-with-use-regex-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-a-with-use-regex-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-b-without-use-regex-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("no-regex.localhost") && (Path("/static") || PathPrefix("/static/"))`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-b-without-use-regex-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-b-without-use-regex-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-a-with-use-regex-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-a-with-use-regex-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-b-without-use-regex-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-b-without-use-regex-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-a-with-use-regex-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-a-with-use-regex",
							},
						},
						"default-ingress-b-without-use-regex-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-b-without-use-regex",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-a-with-use-regex": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-b-without-use-regex": {
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
						"default-ingress-with-rewrite-target-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("rewrite-target.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-rewrite-target-whoami-80",
							Middlewares: []string{"default-ingress-with-rewrite-target-rule-0-path-0-rewrite-target", "default-ingress-with-rewrite-target-rule-0-path-0-retry"},
						},
						"default-ingress-with-rewrite-target-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("rewrite-target.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-rewrite-target-whoami-80",
							Middlewares: []string{"default-ingress-with-rewrite-target-rule-0-path-0-tls-rewrite-target", "default-ingress-with-rewrite-target-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-rewrite-target-no-regex-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("rewrite-target-no-regex.localhost") && (Path("/something") || PathPrefix("/something/"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-rewrite-target-no-regex-whoami-80",
							Middlewares: []string{"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-retry"},
						},
						"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("rewrite-target-no-regex.localhost") && (Path("/something") || PathPrefix("/something/"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-rewrite-target-no-regex-whoami-80",
							Middlewares: []string{"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-rewrite-target-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "(?i)/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-ingress-with-rewrite-target-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "(?i)/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-ingress-with-rewrite-target-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-rewrite-target-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-rewrite-target-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-rewrite-target",
							},
						},
						"default-ingress-with-rewrite-target-no-regex-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-rewrite-target-no-regex",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-rewrite-target": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-rewrite-target-no-regex": {
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
						"default-ingress-with-rewrite-target-no-regex-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("rewrite-target-no-regex.localhost") && Path("/original")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-rewrite-target-no-regex-whoami-80",
							Middlewares: []string{
								"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-rewrite-target",
								"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-retry",
							},
						},
						"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("rewrite-target-no-regex.localhost") && Path("/original")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-rewrite-target-no-regex-whoami-80",
							Middlewares: []string{
								"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-tls-rewrite-target",
								"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "(?i)/original",
								Replacement: "/rewritten",
							},
						},
						"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "(?i)/original",
								Replacement: "/rewritten",
							},
						},
						"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-rewrite-target-no-regex-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-rewrite-target-no-regex",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-rewrite-target-no-regex": {
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
						"default-ingress-a-with-rewrite-target-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-a-with-rewrite-target-whoami-80",
							Middlewares: []string{"default-ingress-a-with-rewrite-target-rule-0-path-0-rewrite-target", "default-ingress-a-with-rewrite-target-rule-0-path-0-retry"},
						},
						"default-ingress-b-without-rewrite-target-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/static")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-b-without-rewrite-target-whoami-80",
							Middlewares: []string{"default-ingress-b-without-rewrite-target-rule-0-path-0-retry"},
						},
						"default-ingress-a-with-rewrite-target-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-a-with-rewrite-target-whoami-80",
							Middlewares: []string{"default-ingress-a-with-rewrite-target-rule-0-path-0-tls-rewrite-target", "default-ingress-a-with-rewrite-target-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-b-without-rewrite-target-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("shared.localhost") && PathRegexp("(?i)^/static")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-b-without-rewrite-target-whoami-80",
							Middlewares: []string{"default-ingress-b-without-rewrite-target-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-a-with-rewrite-target-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "(?i)/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-ingress-a-with-rewrite-target-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "(?i)/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-ingress-a-with-rewrite-target-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
						"default-ingress-a-with-rewrite-target-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
						"default-ingress-b-without-rewrite-target-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
						"default-ingress-b-without-rewrite-target-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-a-with-rewrite-target-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-a-with-rewrite-target",
							},
						},
						"default-ingress-b-without-rewrite-target-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-b-without-rewrite-target",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-a-with-rewrite-target": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-b-without-rewrite-target": {
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
						"default-ingress-a-with-rewrite-target-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("rewrite.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-a-with-rewrite-target-whoami-80",
							Middlewares: []string{"default-ingress-a-with-rewrite-target-rule-0-path-0-rewrite-target", "default-ingress-a-with-rewrite-target-rule-0-path-0-retry"},
						},
						"default-ingress-b-without-rewrite-target-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("no-rewrite.localhost") && (Path("/static") || PathPrefix("/static/"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-b-without-rewrite-target-whoami-80",
							Middlewares: []string{"default-ingress-b-without-rewrite-target-rule-0-path-0-retry"},
						},
						"default-ingress-a-with-rewrite-target-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("rewrite.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-a-with-rewrite-target-whoami-80",
							Middlewares: []string{"default-ingress-a-with-rewrite-target-rule-0-path-0-tls-rewrite-target", "default-ingress-a-with-rewrite-target-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-b-without-rewrite-target-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("no-rewrite.localhost") && (Path("/static") || PathPrefix("/static/"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-b-without-rewrite-target-whoami-80",
							Middlewares: []string{"default-ingress-b-without-rewrite-target-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-a-with-rewrite-target-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "(?i)/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-ingress-a-with-rewrite-target-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "(?i)/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-ingress-a-with-rewrite-target-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
						"default-ingress-a-with-rewrite-target-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
						"default-ingress-b-without-rewrite-target-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
						"default-ingress-b-without-rewrite-target-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-a-with-rewrite-target-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-a-with-rewrite-target",
							},
						},
						"default-ingress-b-without-rewrite-target-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-b-without-rewrite-target",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-a-with-rewrite-target": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-b-without-rewrite-target": {
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
						"default-ingress-with-rewrite-target-use-regex-false-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("rewrite-target.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-rewrite-target-use-regex-false-whoami-80",
							Middlewares: []string{"default-ingress-with-rewrite-target-use-regex-false-rule-0-path-0-rewrite-target", "default-ingress-with-rewrite-target-use-regex-false-rule-0-path-0-retry"},
						},
						"default-ingress-with-rewrite-target-use-regex-false-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("rewrite-target.localhost") && PathRegexp("(?i)^/something(/|$)(.*)")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-rewrite-target-use-regex-false-whoami-80",
							Middlewares: []string{"default-ingress-with-rewrite-target-use-regex-false-rule-0-path-0-tls-rewrite-target", "default-ingress-with-rewrite-target-use-regex-false-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-rewrite-target-use-regex-false-rule-0-path-0-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "(?i)/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-ingress-with-rewrite-target-use-regex-false-rule-0-path-0-tls-rewrite-target": {
							RewriteTarget: &dynamic.RewriteTarget{
								Regex:       "(?i)/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-ingress-with-rewrite-target-use-regex-false-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
						"default-ingress-with-rewrite-target-use-regex-false-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-rewrite-target-use-regex-false-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-rewrite-target-use-regex-false",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-rewrite-target-use-regex-false": {
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
						"default-ingress-with-app-root-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-app-root-whoami-80",
							Middlewares: []string{"default-ingress-with-app-root-rule-0-path-0-app-root", "default-ingress-with-app-root-rule-0-path-0-retry"},
						},
						"default-ingress-with-app-root-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-app-root-whoami-80",
							Middlewares: []string{"default-ingress-with-app-root-rule-0-path-0-tls-app-root", "default-ingress-with-app-root-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-app-root-rule-0-path-0-app-root": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?://[^/]+)/$`,
								Replacement: "$1/foo",
							},
						},
						"default-ingress-with-app-root-rule-0-path-0-tls-app-root": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?://[^/]+)/$`,
								Replacement: "$1/foo",
							},
						},
						"default-ingress-with-app-root-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-app-root-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-app-root-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-app-root",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-app-root": {
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
						"default-ingress-with-app-root-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-app-root-rule-0-path-0-retry"},
							Service:     "default-ingress-with-app-root-whoami-80",
						},
						"default-ingress-with-app-root-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("app-root.localhost") && (Path("/bar") || PathPrefix("/bar/"))`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-app-root-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-app-root-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-app-root-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-app-root-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-app-root-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-app-root",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-app-root": {
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
						"default-ingress-with-www-host-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("www.host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-www-host-rule-0-path-0-retry"},
							Service:     "default-ingress-with-www-host-whoami-80",
						},
						"default-ingress-with-www-host-rule-0-path-0-from-to-www-redirect": {
							EntryPoints: []string{"http"},
							Rule:        `Host("host.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-www-host-whoami-80",
							Middlewares: []string{"default-ingress-with-www-host-rule-0-path-0-from-to-www-redirect"},
						},
						"default-ingress-with-www-host-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("www.host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-www-host-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-www-host-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-www-host-rule-0-path-0-tls-from-to-www-redirect": {
							EntryPoints: []string{"https"},
							Rule:        `Host("host.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-www-host-whoami-80",
							Middlewares: []string{"default-ingress-with-www-host-rule-0-path-0-tls-from-to-www-redirect"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-www-host-rule-0-path-0-from-to-www-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `(https?)://[^/:]+(:[0-9]+)?/(.*)`,
								Replacement: "$1://www.host.localhost$2/$3",
								StatusCode:  ptr.To(http.StatusPermanentRedirect),
							},
						},
						"default-ingress-with-www-host-rule-0-path-0-tls-from-to-www-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `(https?)://[^/:]+(:[0-9]+)?/(.*)`,
								Replacement: "$1://www.host.localhost$2/$3",
								StatusCode:  ptr.To(http.StatusPermanentRedirect),
							},
						},
						"default-ingress-with-www-host-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-www-host-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-www-host-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-www-host",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-www-host": {
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
						"default-ingress-with-host-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-host-rule-0-path-0-retry"},
							Service:     "default-ingress-with-host-whoami-80",
						},
						"default-ingress-with-host-rule-0-path-0-from-to-www-redirect": {
							EntryPoints: []string{"http"},
							Rule:        `Host("www.host.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-host-whoami-80",
							Middlewares: []string{"default-ingress-with-host-rule-0-path-0-from-to-www-redirect"},
						},
						"default-ingress-with-host-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-host-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-host-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-host-rule-0-path-0-tls-from-to-www-redirect": {
							EntryPoints: []string{"https"},
							Rule:        `Host("www.host.localhost")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-host-whoami-80",
							Middlewares: []string{"default-ingress-with-host-rule-0-path-0-tls-from-to-www-redirect"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-host-rule-0-path-0-from-to-www-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `(https?)://[^/:]+(:[0-9]+)?/(.*)`,
								Replacement: "$1://host.localhost$2/$3",
								StatusCode:  ptr.To(http.StatusPermanentRedirect),
							},
						},
						"default-ingress-with-host-rule-0-path-0-tls-from-to-www-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `(https?)://[^/:]+(:[0-9]+)?/(.*)`,
								Replacement: "$1://host.localhost$2/$3",
								StatusCode:  ptr.To(http.StatusPermanentRedirect),
							},
						},
						"default-ingress-with-host-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-host-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-host-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-host",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-host": {
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
						"default-ingress-with-host-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-host-rule-0-path-0-retry"},
							Service:     "default-ingress-with-host-whoami-80",
						},
						"default-ingress-with-www-host-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("www.host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-www-host-rule-0-path-0-retry"},
							Service:     "default-ingress-with-www-host-whoami-80",
						},
						"default-ingress-with-host-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-host-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-host-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-www-host-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("www.host.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-www-host-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-www-host-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-host-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-host-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-www-host-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-www-host-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-host-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-host",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-www-host-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-www-host",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-www-host": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-host": {
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
						},
						"default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							TLS:         &dynamic.RouterTLSConfig{},
							Service:     "default-backend",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
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
								PassHostHeader: ptr.To(true),
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
						"default-ingress-with-whitelist-single-ip-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-single-ip-rule-0-path-0-allowed-source-range", "default-ingress-with-whitelist-single-ip-rule-0-path-0-retry"},
							Service:     "default-ingress-with-whitelist-single-ip-whoami-80",
						},
						"default-ingress-with-whitelist-single-ip-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-single-ip-rule-0-path-0-tls-allowed-source-range", "default-ingress-with-whitelist-single-ip-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-whitelist-single-ip-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-whitelist-single-ip-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.20.1"},
							},
						},
						"default-ingress-with-whitelist-single-ip-rule-0-path-0-tls-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.20.1"},
							},
						},
						"default-ingress-with-whitelist-single-ip-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-whitelist-single-ip-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-whitelist-single-ip-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-whitelist-single-ip",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-whitelist-single-ip": {
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
						"default-ingress-with-whitelist-single-cidr-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-single-cidr-rule-0-path-0-allowed-source-range", "default-ingress-with-whitelist-single-cidr-rule-0-path-0-retry"},
							Service:     "default-ingress-with-whitelist-single-cidr-whoami-80",
						},
						"default-ingress-with-whitelist-single-cidr-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-single-cidr-rule-0-path-0-tls-allowed-source-range", "default-ingress-with-whitelist-single-cidr-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-whitelist-single-cidr-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-whitelist-single-cidr-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24"},
							},
						},
						"default-ingress-with-whitelist-single-cidr-rule-0-path-0-tls-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24"},
							},
						},
						"default-ingress-with-whitelist-single-cidr-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-whitelist-single-cidr-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-whitelist-single-cidr-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-whitelist-single-cidr",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-whitelist-single-cidr": {
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
						"default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-allowed-source-range", "default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-retry"},
							Service:     "default-ingress-with-whitelist-multiple-ip-and-cidr-whoami-80",
						},
						"default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-tls-allowed-source-range", "default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-whitelist-multiple-ip-and-cidr-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24", "10.0.0.0/8", "192.168.20.1"},
							},
						},
						"default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-tls-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24", "10.0.0.0/8", "192.168.20.1"},
							},
						},
						"default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-whitelist-multiple-ip-and-cidr-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-whitelist-multiple-ip-and-cidr",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-whitelist-multiple-ip-and-cidr": {
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
						"default-ingress-with-whitelist-empty-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-empty-rule-0-path-0-retry"},
							Service:     "default-ingress-with-whitelist-empty-whoami-80",
						},
						"default-ingress-with-whitelist-empty-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whitelist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-empty-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-whitelist-empty-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-whitelist-empty-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-whitelist-empty-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-whitelist-empty-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-whitelist-empty",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-whitelist-empty": {
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
						"default-ingress-with-allowlist-empty-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-allowlist-empty-rule-0-path-0-retry"},
							Service:     "default-ingress-with-allowlist-empty-whoami-80",
						},
						"default-ingress-with-allowlist-empty-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-allowlist-empty-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-allowlist-empty-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-allowlist-empty-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-allowlist-empty-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-allowlist-empty-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-allowlist-empty",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-allowlist-empty": {
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
						"default-ingress-with-allowlist-single-ip-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-allowlist-single-ip-rule-0-path-0-allowed-source-range", "default-ingress-with-allowlist-single-ip-rule-0-path-0-retry"},
							Service:     "default-ingress-with-allowlist-single-ip-whoami-80",
						},
						"default-ingress-with-allowlist-single-ip-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-allowlist-single-ip-rule-0-path-0-tls-allowed-source-range", "default-ingress-with-allowlist-single-ip-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-allowlist-single-ip-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-allowlist-single-ip-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.20.1"},
							},
						},
						"default-ingress-with-allowlist-single-ip-rule-0-path-0-tls-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.20.1"},
							},
						},
						"default-ingress-with-allowlist-single-ip-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-allowlist-single-ip-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-allowlist-single-ip-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-allowlist-single-ip",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-allowlist-single-ip": {
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
						"default-ingress-with-allowlist-single-cidr-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-allowlist-single-cidr-rule-0-path-0-allowed-source-range", "default-ingress-with-allowlist-single-cidr-rule-0-path-0-retry"},
							Service:     "default-ingress-with-allowlist-single-cidr-whoami-80",
						},
						"default-ingress-with-allowlist-single-cidr-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-allowlist-single-cidr-rule-0-path-0-tls-allowed-source-range", "default-ingress-with-allowlist-single-cidr-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-allowlist-single-cidr-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-allowlist-single-cidr-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24"},
							},
						},
						"default-ingress-with-allowlist-single-cidr-rule-0-path-0-tls-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24"},
							},
						},
						"default-ingress-with-allowlist-single-cidr-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-allowlist-single-cidr-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-allowlist-single-cidr-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-allowlist-single-cidr",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-allowlist-single-cidr": {
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
						"default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0-allowed-source-range", "default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0-retry"},
							Service:     "default-ingress-with-allowlist-multiple-ip-and-cidr-whoami-80",
						},
						"default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("allowlist-source-range.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0-tls-allowed-source-range", "default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-allowlist-multiple-ip-and-cidr-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24", "10.0.0.0/8", "192.168.20.1"},
							},
						},
						"default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0-tls-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24", "10.0.0.0/8", "192.168.20.1"},
							},
						},
						"default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-allowlist-multiple-ip-and-cidr-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-allowlist-multiple-ip-and-cidr",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-allowlist-multiple-ip-and-cidr": {
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
						"default-ingress-with-access-log-enabled-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("accesslog-enabled.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-access-log-enabled-rule-0-path-0-retry"},
							Service:     "default-ingress-with-access-log-enabled-whoami-80",
							Observability: &dynamic.RouterObservabilityConfig{
								AccessLogs: ptr.To(true),
							},
						},
						"default-ingress-with-access-log-enabled-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("accesslog-enabled.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-access-log-enabled-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-access-log-enabled-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								AccessLogs: ptr.To(true),
							},
						},
						"default-ingress-with-access-log-disabled-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("accesslog-disabled.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-access-log-disabled-rule-0-path-0-retry"},
							Service:     "default-ingress-with-access-log-disabled-whoami-80",
							Observability: &dynamic.RouterObservabilityConfig{
								AccessLogs: ptr.To(false),
							},
						},
						"default-ingress-with-access-log-disabled-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("accesslog-disabled.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-access-log-disabled-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-access-log-disabled-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
							Observability: &dynamic.RouterObservabilityConfig{
								AccessLogs: ptr.To(false),
							},
						},
						"default-ingress-with-access-log-default-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("accesslog-default.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-access-log-default-rule-0-path-0-retry"},
							Service:     "default-ingress-with-access-log-default-whoami-80",
						},
						"default-ingress-with-access-log-default-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("accesslog-default.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-access-log-default-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-access-log-default-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-access-log-enabled-rule-0-path-0-retry":      {Retry: &dynamic.Retry{Attempts: 3}},
						"default-ingress-with-access-log-enabled-rule-0-path-0-tls-retry":  {Retry: &dynamic.Retry{Attempts: 3}},
						"default-ingress-with-access-log-disabled-rule-0-path-0-retry":     {Retry: &dynamic.Retry{Attempts: 3}},
						"default-ingress-with-access-log-disabled-rule-0-path-0-tls-retry": {Retry: &dynamic.Retry{Attempts: 3}},
						"default-ingress-with-access-log-default-rule-0-path-0-retry":      {Retry: &dynamic.Retry{Attempts: 3}},
						"default-ingress-with-access-log-default-rule-0-path-0-tls-retry":  {Retry: &dynamic.Retry{Attempts: 3}},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-access-log-enabled-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-access-log-enabled",
							},
						},
						"default-ingress-with-access-log-disabled-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-access-log-disabled",
							},
						},
						"default-ingress-with-access-log-default-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-access-log-default",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-access-log-enabled": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-access-log-disabled": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-access-log-default": {
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
						"default-ingress-with-permanent-redirect-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("permanent-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-permanent-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-permanent-redirect-rule-0-path-0-redirect", "default-ingress-with-permanent-redirect-rule-0-path-0-retry"},
						},
						"default-ingress-with-permanent-redirect-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("permanent-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-permanent-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-permanent-redirect-rule-0-path-0-tls-redirect", "default-ingress-with-permanent-redirect-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-permanent-redirect-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusMovedPermanently),
							},
						},
						"default-ingress-with-permanent-redirect-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusMovedPermanently),
							},
						},
						"default-ingress-with-permanent-redirect-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-permanent-redirect-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-permanent-redirect-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-permanent-redirect",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-permanent-redirect": {
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
						"default-ingress-with-permanent-redirect-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("permanent-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-permanent-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-permanent-redirect-rule-0-path-0-redirect", "default-ingress-with-permanent-redirect-rule-0-path-0-retry"},
						},
						"default-ingress-with-permanent-redirect-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("permanent-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-permanent-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-permanent-redirect-rule-0-path-0-tls-redirect", "default-ingress-with-permanent-redirect-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-permanent-redirect-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusMovedPermanently),
							},
						},
						"default-ingress-with-permanent-redirect-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusMovedPermanently),
							},
						},
						"default-ingress-with-permanent-redirect-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-permanent-redirect-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-permanent-redirect-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-permanent-redirect",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-permanent-redirect": {
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
						"default-ingress-with-permanent-redirect-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("permanent-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-permanent-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-permanent-redirect-rule-0-path-0-redirect", "default-ingress-with-permanent-redirect-rule-0-path-0-retry"},
						},
						"default-ingress-with-permanent-redirect-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("permanent-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-permanent-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-permanent-redirect-rule-0-path-0-tls-redirect", "default-ingress-with-permanent-redirect-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-permanent-redirect-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusMultipleChoices),
							},
						},
						"default-ingress-with-permanent-redirect-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusMultipleChoices),
							},
						},
						"default-ingress-with-permanent-redirect-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-permanent-redirect-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-permanent-redirect-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-permanent-redirect",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-permanent-redirect": {
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
						"default-ingress-with-redirect-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-redirect-rule-0-path-0-redirect", "default-ingress-with-redirect-rule-0-path-0-retry"},
						},
						"default-ingress-with-redirect-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-redirect-rule-0-path-0-tls-redirect", "default-ingress-with-redirect-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-redirect-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusFound),
							},
						},
						"default-ingress-with-redirect-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusFound),
							},
						},
						"default-ingress-with-redirect-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-redirect-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-redirect-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-redirect",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-redirect": {
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
						"default-ingress-with-temporal-redirect-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("temporal-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-temporal-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-temporal-redirect-rule-0-path-0-redirect", "default-ingress-with-temporal-redirect-rule-0-path-0-retry"},
						},
						"default-ingress-with-temporal-redirect-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("temporal-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-temporal-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-temporal-redirect-rule-0-path-0-tls-redirect", "default-ingress-with-temporal-redirect-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-temporal-redirect-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusFound),
							},
						},
						"default-ingress-with-temporal-redirect-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusFound),
							},
						},
						"default-ingress-with-temporal-redirect-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-temporal-redirect-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-temporal-redirect-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-temporal-redirect",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-temporal-redirect": {
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
						"default-ingress-with-temporal-redirect-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("temporal-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-temporal-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-temporal-redirect-rule-0-path-0-redirect", "default-ingress-with-temporal-redirect-rule-0-path-0-retry"},
						},
						"default-ingress-with-temporal-redirect-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("temporal-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-temporal-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-temporal-redirect-rule-0-path-0-tls-redirect", "default-ingress-with-temporal-redirect-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-temporal-redirect-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusFound),
							},
						},
						"default-ingress-with-temporal-redirect-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusFound),
							},
						},
						"default-ingress-with-temporal-redirect-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-temporal-redirect-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-temporal-redirect-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-temporal-redirect",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-temporal-redirect": {
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
						"default-ingress-with-temporal-redirect-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("temporal-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-temporal-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-temporal-redirect-rule-0-path-0-redirect", "default-ingress-with-temporal-redirect-rule-0-path-0-retry"},
						},
						"default-ingress-with-temporal-redirect-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("temporal-redirect.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-temporal-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-temporal-redirect-rule-0-path-0-tls-redirect", "default-ingress-with-temporal-redirect-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-temporal-redirect-rule-0-path-0-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusPermanentRedirect),
							},
						},
						"default-ingress-with-temporal-redirect-rule-0-path-0-tls-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       ".*",
								Replacement: "https://www.google.com",
								StatusCode:  ptr.To(http.StatusPermanentRedirect),
							},
						},
						"default-ingress-with-temporal-redirect-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-temporal-redirect-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-temporal-redirect-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-temporal-redirect",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-temporal-redirect": {
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
						"default-ingress-with-proxy-timeout-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-timeout-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-timeout-whoami-80",
						},
						"default-ingress-with-proxy-timeout-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-timeout-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-proxy-timeout-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-timeout-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-proxy-timeout-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-timeout-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-timeout",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-timeout": {
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
						"default-ingress-with-proxy-timeout-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-timeout-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-timeout-whoami-80",
						},
						"default-ingress-with-proxy-timeout-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-timeout-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-proxy-timeout-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-timeout-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-proxy-timeout-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-timeout-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-timeout",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-timeout": {
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
						"default-ingress-with-proxy-timeout-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-timeout-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-timeout-whoami-80",
						},
						"default-ingress-with-proxy-timeout-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-timeout-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-proxy-timeout-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-timeout-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-proxy-timeout-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-timeout-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-timeout",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-timeout": {
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
						"default-ingress-with-auth-tls-secret-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("auth-tls-secret.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-auth-tls-secret-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-auth-tls-secret-whoami-80",
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-ingress-with-auth-tls-secret-default-ca-secret",
							},
						},
						"default-ingress-with-auth-tls-secret-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("auth-tls-secret.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-auth-tls-secret-rule-0-path-0-retry"},
							Service:     "default-ingress-with-auth-tls-secret-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-auth-tls-secret-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-auth-tls-secret-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-auth-tls-secret-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-auth-tls-secret",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-auth-tls-secret": {
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
							Certificate: tls.Certificate{
								CertFile: "-----BEGIN CERTIFICATE-----",
								KeyFile:  "-----BEGIN CERTIFICATE-----",
							},
						},
					},
					Options: map[string]tls.Options{
						"default-ingress-with-auth-tls-secret-default-ca-secret": {
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
						"default-ingress-with-auth-tls-verify-client-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("auth-tls-verify-client.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-auth-tls-verify-client-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-auth-tls-verify-client-whoami-80",
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-ingress-with-auth-tls-verify-client-default-ca-secret",
							},
						},
						"default-ingress-with-auth-tls-verify-client-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("auth-tls-verify-client.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-auth-tls-verify-client-rule-0-path-0-retry"},
							Service:     "default-ingress-with-auth-tls-verify-client-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-auth-tls-verify-client-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-auth-tls-verify-client-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-auth-tls-verify-client-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-auth-tls-verify-client",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-auth-tls-verify-client": {
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
							Certificate: tls.Certificate{
								CertFile: "-----BEGIN CERTIFICATE-----",
								KeyFile:  "-----BEGIN CERTIFICATE-----",
							},
						},
					},
					Options: map[string]tls.Options{
						"default-ingress-with-auth-tls-verify-client-default-ca-secret": {
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
						"default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-custom-http-errors-and-default-backend-whoami-80",
							Middlewares: []string{"default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-custom-http-errors", "default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-retry"},
						},
						"default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-custom-http-errors-and-default-backend-whoami-80",
							Middlewares: []string{"default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-tls-custom-http-errors", "default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "415"},
								Service: "default-backend-default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0",
								NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-custom-http-errors-and-default-backend"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-tls-custom-http-errors": {
							Errors: &dynamic.ErrorPage{
								Status:  []string{"404", "415"},
								Service: "default-backend-default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-tls", NginxHeaders: &http.Header{
									"X-Namespaces":   {"default"},
									"X-Ingress-Name": {"ingress-with-custom-http-errors-and-default-backend"},
									"X-Service-Name": {"whoami"},
									"X-Service-Port": {"80"},
								},
							},
						},
						"default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-custom-http-errors-and-default-backend-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-custom-http-errors-and-default-backend",
							},
						},
						"default-backend-default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-backend-default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-tls": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.5:8000"},
									{URL: "http://10.10.0.6:8000"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-custom-http-errors-and-default-backend": {
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
						"default-ingress-with-custom-http-errors-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-custom-http-errors-whoami-80",
							Middlewares: []string{
								"default-ingress-with-custom-http-errors-rule-0-path-0-custom-http-errors",
								"default-ingress-with-custom-http-errors-rule-0-path-0-retry",
							},
						},
						"default-ingress-with-custom-http-errors-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-custom-http-errors-whoami-80",
							Middlewares: []string{
								"default-ingress-with-custom-http-errors-rule-0-path-0-tls-custom-http-errors",
								"default-ingress-with-custom-http-errors-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-backend": {
							EntryPoints: []string{"http"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							Service:     "default-backend",
						},
						"default-backend-tls": {
							EntryPoints: []string{"https"},
							Rule:        `PathPrefix("/")`,
							RuleSyntax:  "default",
							Priority:    math.MinInt32,
							TLS:         &dynamic.RouterTLSConfig{},
							Service:     "default-backend",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-custom-http-errors-rule-0-path-0-custom-http-errors": {
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
						"default-ingress-with-custom-http-errors-rule-0-path-0-tls-custom-http-errors": {
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
						"default-ingress-with-custom-http-errors-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-custom-http-errors-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-custom-http-errors-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-custom-http-errors",
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-custom-http-errors": {
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
						"default-ingress-with-custom-http-errors-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-custom-http-errors-whoami-80",
							Middlewares: []string{"default-ingress-with-custom-http-errors-rule-0-path-0-retry"},
						},
						"default-ingress-with-custom-http-errors-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-custom-http-errors-whoami-80",
							Middlewares: []string{"default-ingress-with-custom-http-errors-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-custom-http-errors-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
						"default-ingress-with-custom-http-errors-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-custom-http-errors-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-custom-http-errors",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-custom-http-errors": {
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
						"default-ingress-with-default-backend-annotation-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-default-backend-annotation-empty-80",
							Middlewares: []string{"default-ingress-with-default-backend-annotation-rule-0-path-0-retry"},
						},
						"default-ingress-with-default-backend-annotation-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-default-backend-annotation-empty-80",
							Middlewares: []string{"default-ingress-with-default-backend-annotation-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-default-backend-annotation-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-default-backend-annotation-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-default-backend-annotation-empty-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-default-backend-annotation",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-default-backend-annotation": {
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
						"default-ingress-with-proxy-body-size-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-body-size-rule-0-path-0-buffering", "default-ingress-with-proxy-body-size-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-body-size-whoami-80",
						},
						"default-ingress-with-proxy-body-size-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-body-size-rule-0-path-0-tls-buffering", "default-ingress-with-proxy-body-size-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-proxy-body-size-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-body-size-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								MaxRequestBodyBytes:   10 * 1024 * 1024,
								MemRequestBodyBytes:   defaultClientBodyBufferSize,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-ingress-with-proxy-body-size-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								MaxRequestBodyBytes:   10 * 1024 * 1024,
								MemRequestBodyBytes:   defaultClientBodyBufferSize,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-ingress-with-proxy-body-size-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-proxy-body-size-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-body-size-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-body-size",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-body-size": {
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
						"default-ingress-with-client-body-buffer-size-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-client-body-buffer-size-rule-0-path-0-buffering", "default-ingress-with-client-body-buffer-size-rule-0-path-0-retry"},
							Service:     "default-ingress-with-client-body-buffer-size-whoami-80",
						},
						"default-ingress-with-client-body-buffer-size-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-client-body-buffer-size-rule-0-path-0-tls-buffering", "default-ingress-with-client-body-buffer-size-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-client-body-buffer-size-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-client-body-buffer-size-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								MemRequestBodyBytes:   10 * 1024 * 1024,
								MaxRequestBodyBytes:   defaultProxyBodySize,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-ingress-with-client-body-buffer-size-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								MemRequestBodyBytes:   10 * 1024 * 1024,
								MaxRequestBodyBytes:   defaultProxyBodySize,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-ingress-with-client-body-buffer-size-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-client-body-buffer-size-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-client-body-buffer-size-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-client-body-buffer-size",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-client-body-buffer-size": {
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
						"default-ingress-with-proxy-body-size-and-client-body-buffer-size-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-body-size-and-client-body-buffer-size-rule-0-path-0-buffering", "default-ingress-with-proxy-body-size-and-client-body-buffer-size-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-body-size-and-client-body-buffer-size-whoami-80",
						},
						"default-ingress-with-proxy-body-size-and-client-body-buffer-size-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-body-size-and-client-body-buffer-size-rule-0-path-0-tls-buffering", "default-ingress-with-proxy-body-size-and-client-body-buffer-size-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-proxy-body-size-and-client-body-buffer-size-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-body-size-and-client-body-buffer-size-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								MaxRequestBodyBytes:   10 * 1024 * 1024,
								MemRequestBodyBytes:   10 * 1024,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-ingress-with-proxy-body-size-and-client-body-buffer-size-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								MaxRequestBodyBytes:   10 * 1024 * 1024,
								MemRequestBodyBytes:   10 * 1024,
								MemResponseBodyBytes:  defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								DisableResponseBuffer: true,
							},
						},
						"default-ingress-with-proxy-body-size-and-client-body-buffer-size-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-proxy-body-size-and-client-body-buffer-size-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-body-size-and-client-body-buffer-size-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-body-size-and-client-body-buffer-size",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-body-size-and-client-body-buffer-size": {
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
						"default-ingress-with-proxy-buffer-size-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-buffer-size-rule-0-path-0-buffering", "default-ingress-with-proxy-buffer-size-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-buffer-size-whoami-80",
						},
						"default-ingress-with-proxy-buffer-size-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-buffer-size-rule-0-path-0-tls-buffering", "default-ingress-with-proxy-buffer-size-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-proxy-buffer-size-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-buffer-size-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: 16 * 1024 * int64(defaultProxyBuffersNumber),
								MaxResponseBodyBytes: defaultProxyMaxTempFileSize + (defaultProxyBufferSize * 8),
							},
						},
						"default-ingress-with-proxy-buffer-size-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: 16 * 1024 * int64(defaultProxyBuffersNumber),
								MaxResponseBodyBytes: defaultProxyMaxTempFileSize + (defaultProxyBufferSize * 8),
							},
						},
						"default-ingress-with-proxy-buffer-size-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-proxy-buffer-size-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-buffer-size-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-buffer-size",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-buffer-size": {
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
						"default-ingress-with-proxy-buffers-number-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-buffers-number-rule-0-path-0-buffering", "default-ingress-with-proxy-buffers-number-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-buffers-number-whoami-80",
						},
						"default-ingress-with-proxy-buffers-number-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-buffers-number-rule-0-path-0-tls-buffering", "default-ingress-with-proxy-buffers-number-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-proxy-buffers-number-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-buffers-number-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: defaultProxyBufferSize * 8,
								MaxResponseBodyBytes: defaultProxyMaxTempFileSize + (defaultProxyBufferSize * 8),
							},
						},
						"default-ingress-with-proxy-buffers-number-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: defaultProxyBufferSize * 8,
								MaxResponseBodyBytes: defaultProxyMaxTempFileSize + (defaultProxyBufferSize * 8),
							},
						},
						"default-ingress-with-proxy-buffers-number-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-proxy-buffers-number-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-buffers-number-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-buffers-number",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-buffers-number": {
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
						"default-ingress-with-proxy-buffer-size-and-number-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-buffer-size-and-number-rule-0-path-0-buffering", "default-ingress-with-proxy-buffer-size-and-number-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-buffer-size-and-number-whoami-80",
						},
						"default-ingress-with-proxy-buffer-size-and-number-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-buffer-size-and-number-rule-0-path-0-tls-buffering", "default-ingress-with-proxy-buffer-size-and-number-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-proxy-buffer-size-and-number-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-buffer-size-and-number-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: 16 * 1024 * 8,
								MaxResponseBodyBytes: defaultProxyMaxTempFileSize + (16 * 1024 * 8),
							},
						},
						"default-ingress-with-proxy-buffer-size-and-number-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: 16 * 1024 * 8,
								MaxResponseBodyBytes: defaultProxyMaxTempFileSize + (16 * 1024 * 8),
							},
						},
						"default-ingress-with-proxy-buffer-size-and-number-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-proxy-buffer-size-and-number-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-buffer-size-and-number-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-buffer-size-and-number",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-buffer-size-and-number": {
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
						"default-ingress-with-proxy-max-temp-file-size-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-max-temp-file-size-rule-0-path-0-buffering", "default-ingress-with-proxy-max-temp-file-size-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-max-temp-file-size-whoami-80",
						},
						"default-ingress-with-proxy-max-temp-file-size-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("hostname.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-max-temp-file-size-rule-0-path-0-tls-buffering", "default-ingress-with-proxy-max-temp-file-size-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-proxy-max-temp-file-size-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-max-temp-file-size-rule-0-path-0-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								MaxResponseBodyBytes: (defaultProxyBufferSize * int64(defaultProxyBuffersNumber)) + (100 * 1024 * 1024),
							},
						},
						"default-ingress-with-proxy-max-temp-file-size-rule-0-path-0-tls-buffering": {
							Buffering: &dynamic.Buffering{
								DisableRequestBuffer: true,
								MaxRequestBodyBytes:  defaultProxyBodySize,
								MemRequestBodyBytes:  defaultClientBodyBufferSize,
								MemResponseBodyBytes: defaultProxyBufferSize * int64(defaultProxyBuffersNumber),
								MaxResponseBodyBytes: (defaultProxyBufferSize * int64(defaultProxyBuffersNumber)) + (100 * 1024 * 1024),
							},
						},
						"default-ingress-with-proxy-max-temp-file-size-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-proxy-max-temp-file-size-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-max-temp-file-size-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-max-temp-file-size",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-max-temp-file-size": {
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
						"default-ingress-with-server-snippet-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-server-snippet-whoami-80",
							Middlewares: []string{
								"default-ingress-with-server-snippet-rule-0-path-0-snippet",
								"default-ingress-with-server-snippet-rule-0-path-0-retry",
							},
						},
						"default-ingress-with-server-snippet-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-server-snippet-whoami-80",
							Middlewares: []string{
								"default-ingress-with-server-snippet-rule-0-path-0-tls-snippet",
								"default-ingress-with-server-snippet-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-server-snippet-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet: "add_header X-Server-Snippet \"server-value\";\n",
							},
						},
						"default-ingress-with-server-snippet-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet: "add_header X-Server-Snippet \"server-value\";\n",
							},
						},
						"default-ingress-with-server-snippet-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-server-snippet-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-server-snippet-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-server-snippet",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-server-snippet": {
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
						"default-ingress-with-configuration-snippet-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-configuration-snippet-whoami-80",
							Middlewares: []string{
								"default-ingress-with-configuration-snippet-rule-0-path-0-snippet",
								"default-ingress-with-configuration-snippet-rule-0-path-0-retry",
							},
						},
						"default-ingress-with-configuration-snippet-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-configuration-snippet-whoami-80",
							Middlewares: []string{
								"default-ingress-with-configuration-snippet-rule-0-path-0-tls-snippet",
								"default-ingress-with-configuration-snippet-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-configuration-snippet-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-ingress-with-configuration-snippet-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-ingress-with-configuration-snippet-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-configuration-snippet-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-configuration-snippet-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-configuration-snippet",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-configuration-snippet": {
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
						"default-ingress-with-both-snippets-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-both-snippets-whoami-80",
							Middlewares: []string{
								"default-ingress-with-both-snippets-rule-0-path-0-snippet",
								"default-ingress-with-both-snippets-rule-0-path-0-retry",
							},
						},
						"default-ingress-with-both-snippets-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-both-snippets-whoami-80",
							Middlewares: []string{
								"default-ingress-with-both-snippets-rule-0-path-0-tls-snippet",
								"default-ingress-with-both-snippets-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-both-snippets-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet:        "add_header X-Server-Snippet \"server-value\";\n",
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-ingress-with-both-snippets-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet:        "add_header X-Server-Snippet \"server-value\";\n",
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-ingress-with-both-snippets-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-both-snippets-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-both-snippets-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-both-snippets",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-both-snippets": {
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
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
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
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
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
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
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
						"default-ingress-with-server-snippet-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-server-snippet-whoami-80",
							Middlewares: []string{"default-ingress-with-server-snippet-rule-0-path-0-snippet", "default-ingress-with-server-snippet-rule-0-path-0-retry"},
						},
						"default-ingress-with-server-snippet-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-server-snippet-whoami-80",
							Middlewares: []string{"default-ingress-with-server-snippet-rule-0-path-0-tls-snippet", "default-ingress-with-server-snippet-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-server-snippet-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet: "add_header X-Server-Snippet \"server-value\";\n",
							},
						},
						"default-ingress-with-server-snippet-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet: "add_header X-Server-Snippet \"server-value\";\n",
							},
						},
						"default-ingress-with-server-snippet-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-server-snippet-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-server-snippet-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-server-snippet",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-server-snippet": {
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
						"default-ingress-with-configuration-snippet-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-configuration-snippet-whoami-80",
							Middlewares: []string{"default-ingress-with-configuration-snippet-rule-0-path-0-snippet", "default-ingress-with-configuration-snippet-rule-0-path-0-retry"},
						},
						"default-ingress-with-configuration-snippet-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-configuration-snippet-whoami-80",
							Middlewares: []string{"default-ingress-with-configuration-snippet-rule-0-path-0-tls-snippet", "default-ingress-with-configuration-snippet-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-configuration-snippet-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-ingress-with-configuration-snippet-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-ingress-with-configuration-snippet-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-configuration-snippet-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-configuration-snippet-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-configuration-snippet",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-configuration-snippet": {
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
						"default-ingress-with-both-snippets-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-both-snippets-whoami-80",
							Middlewares: []string{"default-ingress-with-both-snippets-rule-0-path-0-snippet", "default-ingress-with-both-snippets-rule-0-path-0-retry"},
						},
						"default-ingress-with-both-snippets-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("snippet.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-both-snippets-whoami-80",
							Middlewares: []string{"default-ingress-with-both-snippets-rule-0-path-0-tls-snippet", "default-ingress-with-both-snippets-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-both-snippets-rule-0-path-0-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet:        "add_header X-Server-Snippet \"server-value\";\n",
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-ingress-with-both-snippets-rule-0-path-0-tls-snippet": {
							Snippet: &dynamic.Snippet{
								ServerSnippet:        "add_header X-Server-Snippet \"server-value\";\n",
								ConfigurationSnippet: "add_header X-Configuration-Snippet \"configuration-value\";\n",
							},
						},
						"default-ingress-with-both-snippets-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-both-snippets-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-both-snippets-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-both-snippets",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-both-snippets": {
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
						"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("auth-tls-pass-certificate-to-upstream.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-tls-pass-certificate-to-upstream", "default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-auth-tls-pass-certificate-to-upstream-whoami-80",
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-ingress-with-auth-tls-pass-certificate-to-upstream-default-ca-secret",
							},
						},
						"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("auth-tls-pass-certificate-to-upstream.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-retry"},
							Service:     "default-ingress-with-auth-tls-pass-certificate-to-upstream-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-tls-pass-certificate-to-upstream": {
							AuthTLSPassCertificateToUpstream: &dynamic.AuthTLSPassCertificateToUpstream{
								ClientAuthType: tls.RequireAndVerifyClientCert,
								CAFiles:        nil,
							},
						},

						"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-auth-tls-pass-certificate-to-upstream-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-auth-tls-pass-certificate-to-upstream",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-auth-tls-pass-certificate-to-upstream": {
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
							Certificate: tls.Certificate{
								CertFile: "-----BEGIN CERTIFICATE-----",
								KeyFile:  "-----BEGIN CERTIFICATE-----",
							},
						},
					},
					Options: map[string]tls.Options{
						"default-ingress-with-auth-tls-pass-certificate-to-upstream-default-ca-secret": {
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
						"default-ingress-with-proxy-next-upstream-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-next-upstream-rule-0-path-0-retry"},
						},
						"default-ingress-with-proxy-next-upstream-off-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-off-whoami-80",
						},
						"default-ingress-with-proxy-next-upstream-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-next-upstream-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-proxy-next-upstream-off-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-off-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-next-upstream-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts:                 3,
								Status:                   []string{"400"},
								RetryNonIdempotentMethod: true,
							},
						},
						"default-ingress-with-proxy-next-upstream-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts:                 3,
								Status:                   []string{"400"},
								RetryNonIdempotentMethod: true,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-next-upstream-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-next-upstream",
							},
						},
						"default-ingress-with-proxy-next-upstream-off-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-next-upstream-off",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-next-upstream": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-proxy-next-upstream-off": {
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
						"default-ingress-with-proxy-next-upstream-tries-unlimited-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-tries-unlimited-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-next-upstream-tries-unlimited-rule-0-path-0-retry"},
						},
						"default-ingress-with-proxy-next-upstream-tries-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-tries-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-next-upstream-tries-rule-0-path-0-retry"},
						},
						"default-ingress-with-proxy-next-upstream-tries-unlimited-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-tries-unlimited-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-next-upstream-tries-unlimited-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-proxy-next-upstream-tries-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-tries-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-next-upstream-tries-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-next-upstream-tries-unlimited-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 2,
							},
						},
						"default-ingress-with-proxy-next-upstream-tries-unlimited-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 2,
							},
						},
						"default-ingress-with-proxy-next-upstream-tries-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 5,
							},
						},
						"default-ingress-with-proxy-next-upstream-tries-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 5,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-next-upstream-tries-unlimited-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-next-upstream-tries-unlimited",
							},
						},
						"default-ingress-with-proxy-next-upstream-tries-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-next-upstream-tries",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-next-upstream-tries-unlimited": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-proxy-next-upstream-tries": {
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
						"default-ingress-with-proxy-next-upstream-timeout-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-timeout-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-next-upstream-timeout-rule-0-path-0-retry"},
						},
						"default-ingress-with-proxy-next-upstream-timeout-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-timeout-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-next-upstream-timeout-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-next-upstream-timeout-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
								Timeout:  ptypes.Duration(30 * time.Second),
							},
						},
						"default-ingress-with-proxy-next-upstream-timeout-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
								Timeout:  ptypes.Duration(30 * time.Second),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-next-upstream-timeout-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-next-upstream-timeout",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-next-upstream-timeout": {
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
						"default-ingress-with-server-alias-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("original.localhost") || Host("alias1.localhost") || Host("alias2.localhost")) && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-server-alias-rule-0-path-0-retry"},
							Service:     "default-ingress-with-server-alias-whoami-80",
						},
						"default-ingress-with-server-alias-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("original.localhost") || Host("alias1.localhost") || Host("alias2.localhost")) && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-server-alias-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-server-alias-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-server-alias-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-server-alias-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},

					Services: map[string]*dynamic.Service{
						"default-ingress-with-server-alias-whoami-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-server-alias",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-server-alias": {
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
						"default-primary-ingress-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("conflict.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-primary-ingress-rule-0-path-0-retry"},
							Service:     "default-primary-ingress-whoami-80",
						},
						"default-alias-ingress-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("original.localhost") || Host("alias1.localhost")) && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-alias-ingress-rule-0-path-0-retry"},
							Service:     "default-alias-ingress-whoami-80",
						},
						"default-primary-ingress-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("conflict.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-primary-ingress-rule-0-path-0-tls-retry"},
							Service:     "default-primary-ingress-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-alias-ingress-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("original.localhost") || Host("alias1.localhost")) && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-alias-ingress-rule-0-path-0-tls-retry"},
							Service:     "default-alias-ingress-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-primary-ingress-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-primary-ingress-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-alias-ingress-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-alias-ingress-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-primary-ingress-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-primary-ingress",
							},
						},
						"default-alias-ingress-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-alias-ingress",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-primary-ingress": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-alias-ingress": {
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
						"default-ingress-with-proxy-http-version-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("proxy-http-version.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-http-version-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-http-version-rule-0-path-0-retry"},
						},
						"default-ingress-with-proxy-http-version-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("proxy-http-version.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-http-version-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-http-version-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-http-version-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-proxy-http-version-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-http-version-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-http-version",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-http-version": {
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
						"default-ingress-with-proxy-http-version-unsupported-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("proxy-http-version-unsupported.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-http-version-unsupported-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-http-version-unsupported-rule-0-path-0-retry"},
						},
						"default-ingress-with-proxy-http-version-unsupported-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("proxy-http-version-unsupported.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-http-version-unsupported-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-http-version-unsupported-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-http-version-unsupported-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-proxy-http-version-unsupported-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-http-version-unsupported-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-proxy-http-version-unsupported",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-proxy-http-version-unsupported": {
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
						"default-ingress-with-upstream-hash-by-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-upstream-hash-by-whoami-80",
							Middlewares: []string{"default-ingress-with-upstream-hash-by-rule-0-path-0-retry"},
						},
						"default-ingress-with-upstream-hash-by-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-upstream-hash-by-whoami-80",
							Middlewares: []string{"default-ingress-with-upstream-hash-by-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-upstream-hash-by-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-upstream-hash-by-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-upstream-hash-by-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:            "hrw",
								NginxUpstreamHashBy: "$request_uri",
								PassHostHeader:      ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-upstream-hash-by",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-upstream-hash-by": {
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
						"default-ingress-with-canary-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-rule-0-path-0-retry"},
						},
						"default-ingress-with-canary-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-canary-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-canary-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-whoami-80-canary": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-whoami-80-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-ingress-with-canary-whoami-80",
										Weight: ptr.To(100),
									},
									{
										Name:   "default-ingress-with-canary-whoami-80-canary",
										Weight: ptr.To(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-canary": {
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
						"default-ingress-with-canary-and-sticky-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-and-sticky-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-and-sticky-rule-0-path-0-retry"},
						},
						"default-ingress-with-canary-and-sticky-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-and-sticky-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-and-sticky-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-canary-and-sticky-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-and-sticky-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-canary-and-sticky-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "foobar",
										Domain:   "foo.localhost",
										HTTPOnly: true,
										MaxAge:   42,
										Expires:  42,
										Path:     ptr.To("/foobar"),
										SameSite: "none",
										Secure:   true,
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-and-sticky",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-and-sticky-whoami-80-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "foobar",
										Domain:   "foo.localhost",
										HTTPOnly: true,
										MaxAge:   42,
										Expires:  42,
										Path:     ptr.To("/foobar"),
										SameSite: "none",
										Secure:   true,
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-and-sticky",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-and-sticky-whoami-80-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "foobar-wrr",
										Domain:   "foo.localhost",
										HTTPOnly: true,
										MaxAge:   42,
										Expires:  42,
										Path:     ptr.To("/foobar"),
										SameSite: "none",
										Secure:   true,
									},
								},
								Services: []dynamic.WRRService{
									{
										Name:   "default-ingress-with-canary-and-sticky-whoami-80",
										Weight: ptr.To(100),
									},
									{
										Name:   "default-ingress-with-canary-and-sticky-whoami-80-canary",
										Weight: ptr.To(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-canary-and-sticky": {
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
						"default-ingress-with-canary-weight-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-weight-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-weight-rule-0-path-0-retry"},
						},
						"default-ingress-with-canary-weight-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-weight-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-weight-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-canary-weight-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-weight-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-canary-weight-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-weight",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-weight-whoami-80-canary": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-weight",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-weight-whoami-80-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-ingress-with-canary-weight-whoami-80",
										Weight: ptr.To(110),
									},
									{
										Name:   "default-ingress-with-canary-weight-whoami-80-canary",
										Weight: ptr.To(10),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-canary-weight": {
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
						"default-ingress-with-canary-by-header-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-by-header-rule-0-path-0-retry"},
						},
						"default-ingress-with-canary-by-header-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-whoami-80-canary",
							Middlewares: []string{"default-ingress-with-canary-by-header-rule-0-path-0-canary-retry"},
						},
						"default-ingress-with-canary-by-header-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-by-header-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-canary-by-header-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-whoami-80-canary",
							Middlewares: []string{"default-ingress-with-canary-by-header-rule-0-path-0-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-canary-by-header-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-canary-by-header-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-by-header",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-by-header-whoami-80-canary": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-by-header",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-by-header-whoami-80-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-ingress-with-canary-by-header-whoami-80",
										Weight: ptr.To(100),
									},
									{
										Name:   "default-ingress-with-canary-by-header-whoami-80-canary",
										Weight: ptr.To(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-canary-by-header": {
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
						"default-ingress-with-canary-by-header-value-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-value-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-by-header-value-rule-0-path-0-retry"},
						},
						"default-ingress-with-canary-by-header-value-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "bar"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-value-whoami-80-canary",
							Middlewares: []string{"default-ingress-with-canary-by-header-value-rule-0-path-0-canary-retry"},
						},
						"default-ingress-with-canary-by-header-value-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-value-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-by-header-value-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-canary-by-header-value-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "bar"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-value-whoami-80-canary",
							Middlewares: []string{"default-ingress-with-canary-by-header-value-rule-0-path-0-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-canary-by-header-value-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-value-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-value-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-value-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-canary-by-header-value-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-by-header-value",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-by-header-value-whoami-80-canary": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-by-header-value",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-by-header-value-whoami-80-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-ingress-with-canary-by-header-value-whoami-80",
										Weight: ptr.To(100),
									},
									{
										Name:   "default-ingress-with-canary-by-header-value-whoami-80-canary",
										Weight: ptr.To(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-canary-by-header-value": {
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
						"default-ingress-with-canary-by-header-pattern-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-pattern-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-by-header-pattern-rule-0-path-0-retry"},
						},
						"default-ingress-with-canary-by-header-pattern-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (HeaderRegexp("Foo", "bar(.*)"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-pattern-whoami-80-canary",
							Middlewares: []string{"default-ingress-with-canary-by-header-pattern-rule-0-path-0-canary-retry"},
						},
						"default-ingress-with-canary-by-header-pattern-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-pattern-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-by-header-pattern-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-canary-by-header-pattern-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (HeaderRegexp("Foo", "bar(.*)"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-pattern-whoami-80-canary",
							Middlewares: []string{"default-ingress-with-canary-by-header-pattern-rule-0-path-0-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-canary-by-header-pattern-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-pattern-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-pattern-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-pattern-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-canary-by-header-pattern-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-by-header-pattern",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-by-header-pattern-whoami-80-canary": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-by-header-pattern",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-by-header-pattern-whoami-80-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-ingress-with-canary-by-header-pattern-whoami-80",
										Weight: ptr.To(100),
									},
									{
										Name:   "default-ingress-with-canary-by-header-pattern-whoami-80-canary",
										Weight: ptr.To(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-canary-by-header-pattern": {
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
						"default-ingress-with-canary-by-header-misconfigured-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-misconfigured-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-by-header-misconfigured-rule-0-path-0-retry"},
						},
						"default-ingress-with-canary-by-header-misconfigured-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-misconfigured-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-by-header-misconfigured-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-canary-by-header-misconfigured-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-misconfigured-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-canary-by-header-misconfigured-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-by-header-misconfigured",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-by-header-misconfigured-whoami-80-canary": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-by-header-misconfigured",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-by-header-misconfigured-whoami-80-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-ingress-with-canary-by-header-misconfigured-whoami-80",
										Weight: ptr.To(100),
									},
									{
										Name:   "default-ingress-with-canary-by-header-misconfigured-whoami-80-canary",
										Weight: ptr.To(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-canary-by-header-misconfigured": {
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
						"default-ingress-with-canary-by-cookie-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-cookie-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-by-cookie-rule-0-path-0-retry"},
						},
						"default-ingress-with-canary-by-cookie-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (HeaderRegexp("Cookie", "(^|;\\s*)foo=always(;|$)"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-cookie-whoami-80-canary",
							Middlewares: []string{"default-ingress-with-canary-by-cookie-rule-0-path-0-canary-retry"},
						},
						"default-ingress-with-canary-by-cookie-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-cookie-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-by-cookie-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-canary-by-cookie-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (HeaderRegexp("Cookie", "(^|;\\s*)foo=always(;|$)"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-cookie-whoami-80-canary",
							Middlewares: []string{"default-ingress-with-canary-by-cookie-rule-0-path-0-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-canary-by-cookie-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-cookie-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-cookie-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-cookie-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-canary-by-cookie-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-by-cookie",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-by-cookie-whoami-80-canary": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-by-cookie",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-by-cookie-whoami-80-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-ingress-with-canary-by-cookie-whoami-80",
										Weight: ptr.To(100),
									},
									{
										Name:   "default-ingress-with-canary-by-cookie-whoami-80-canary",
										Weight: ptr.To(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-canary-by-cookie": {
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
						"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-and-cookie-and-weight-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-retry"},
						},
						"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always") || (HeaderRegexp("Cookie", "(^|;\\s*)foo=always(;|$)") && !Header("Foo", "never")))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-and-cookie-and-weight-whoami-80-canary",
							Middlewares: []string{"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-canary-retry"},
						},
						"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-non-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "never") || HeaderRegexp("Cookie", "(^|;\\s*)foo=never(;|$)"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-and-cookie-and-weight-whoami-80",
							Middlewares: []string{"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-non-canary-retry"},
						},
						"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-and-cookie-and-weight-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always") || (HeaderRegexp("Cookie", "(^|;\\s*)foo=always(;|$)") && !Header("Foo", "never")))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-and-cookie-and-weight-whoami-80-canary",
							Middlewares: []string{"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-non-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "never") || HeaderRegexp("Cookie", "(^|;\\s*)foo=never(;|$)"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-by-header-and-cookie-and-weight-whoami-80",
							Middlewares: []string{"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-non-canary-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-non-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-by-header-and-cookie-and-weight-rule-0-path-0-non-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-canary-by-header-and-cookie-and-weight-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-by-header-and-cookie-and-weight",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-by-header-and-cookie-and-weight-whoami-80-canary": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-by-header-and-cookie-and-weight",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-by-header-and-cookie-and-weight-whoami-80-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-ingress-with-canary-by-header-and-cookie-and-weight-whoami-80",
										Weight: ptr.To(90),
									},
									{
										Name:   "default-ingress-with-canary-by-header-and-cookie-and-weight-whoami-80-canary",
										Weight: ptr.To(10),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-canary-by-header-and-cookie-and-weight": {
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
						"default-ingress-with-canary-middlewares-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-middlewares-whoami-80-wrr",
							Middlewares: []string{
								"default-ingress-with-canary-middlewares-rule-0-path-0-app-root",
								"default-ingress-with-canary-middlewares-rule-0-path-0-retry",
							},
						},
						"default-ingress-with-canary-middlewares-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-middlewares-whoami-80-canary",
							Middlewares: []string{
								"default-ingress-with-canary-middlewares-rule-0-path-0-canary-app-root",
								"default-ingress-with-canary-middlewares-rule-0-path-0-canary-retry",
							},
						},
						"default-ingress-with-canary-middlewares-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-middlewares-whoami-80-wrr",
							Middlewares: []string{
								"default-ingress-with-canary-middlewares-rule-0-path-0-tls-app-root",
								"default-ingress-with-canary-middlewares-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-canary-middlewares-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-middlewares-whoami-80-canary",
							Middlewares: []string{
								"default-ingress-with-canary-middlewares-rule-0-path-0-canary-tls-app-root",
								"default-ingress-with-canary-middlewares-rule-0-path-0-canary-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-canary-middlewares-rule-0-path-0-app-root": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?://[^/]+)/$`,
								Replacement: "$1/foo",
							},
						},
						"default-ingress-with-canary-middlewares-rule-0-path-0-tls-app-root": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?://[^/]+)/$`,
								Replacement: "$1/foo",
							},
						},
						"default-ingress-with-canary-middlewares-rule-0-path-0-canary-app-root": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?://[^/]+)/$`,
								Replacement: "$1/foo",
							},
						},
						"default-ingress-with-canary-middlewares-rule-0-path-0-canary-tls-app-root": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?://[^/]+)/$`,
								Replacement: "$1/foo",
							},
						},
						"default-ingress-with-canary-middlewares-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-middlewares-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-middlewares-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-middlewares-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-canary-middlewares-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-middlewares",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-middlewares-whoami-80-canary": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-middlewares",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-middlewares-whoami-80-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-ingress-with-canary-middlewares-whoami-80",
										Weight: ptr.To(100),
									},
									{
										Name:   "default-ingress-with-canary-middlewares-whoami-80-canary",
										Weight: ptr.To(0),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-canary-middlewares": {
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
						"default-ingress-with-non-matching-canary-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-non-matching-canary-whoami-80",
							Middlewares: []string{"default-ingress-with-non-matching-canary-rule-0-path-0-retry"},
						},
						"default-ingress-with-non-matching-canary-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-non-matching-canary-whoami-80",
							Middlewares: []string{"default-ingress-with-non-matching-canary-rule-0-path-0-tls-retry"},
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-non-matching-canary-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-non-matching-canary-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-non-matching-canary-whoami-80": {
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
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-non-matching-canary",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-non-matching-canary": {
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
						"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-middlewares-and-tls-whoami-80-wrr",
							Middlewares: []string{"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-app-root", "default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-retry"},
						},
						"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-canary": {
							EntryPoints: []string{"http"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-middlewares-and-tls-whoami-80-canary",
							Middlewares: []string{"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-canary-app-root", "default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-canary-retry"},
						},
						"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("production.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-middlewares-and-tls-whoami-80-wrr",
							Middlewares: []string{
								"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-tls-app-root",
								"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-canary-tls": {
							EntryPoints: []string{"https"},
							Rule:        `(Host("production.localhost") && PathPrefix("/")) && (Header("Foo", "always"))`,
							RuleSyntax:  "default",
							Service:     "default-ingress-with-canary-middlewares-and-tls-whoami-80-canary",
							Middlewares: []string{
								"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-canary-tls-app-root",
								"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-canary-tls-retry",
							},
							TLS: &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-app-root": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?://[^/]+)/$`,
								Replacement: "$1/foo",
							},
						},
						"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-tls-app-root": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?://[^/]+)/$`,
								Replacement: "$1/foo",
							},
						},
						"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-canary-app-root": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?://[^/]+)/$`,
								Replacement: "$1/foo",
							},
						},
						"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-canary-tls-app-root": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?://[^/]+)/$`,
								Replacement: "$1/foo",
							},
						},
						"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-canary-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-canary-middlewares-and-tls-rule-0-path-0-canary-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-canary-middlewares-and-tls-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-middlewares-and-tls",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-middlewares-and-tls-whoami-80-canary": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.7:80"},
									{URL: "http://10.10.0.8:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-canary-middlewares-and-tls",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
						"default-ingress-with-canary-middlewares-and-tls-whoami-80-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{Name: "default-ingress-with-canary-middlewares-and-tls-whoami-80", Weight: ptr.To(100)},
									{Name: "default-ingress-with-canary-middlewares-and-tls-whoami-80-canary", Weight: ptr.To(0)},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-canary-middlewares-and-tls": {
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
							Certificate: tls.Certificate{
								CertFile: "-----BEGIN CERTIFICATE-----",
								KeyFile:  "-----BEGIN CERTIFICATE-----",
							},
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
						"default-ingress-with-limit-rps-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-limit-rps-rule-0-path-0-limit-rps", "default-ingress-with-limit-rps-rule-0-path-0-retry"},
							Service:     "default-ingress-with-limit-rps-whoami-80",
						},
						"default-ingress-with-limit-rps-zero-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami-zero.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-limit-rps-zero-rule-0-path-0-retry"},
							Service:     "default-ingress-with-limit-rps-zero-whoami-80",
						},
						"default-ingress-with-limit-rps-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-limit-rps-rule-0-path-0-tls-limit-rps", "default-ingress-with-limit-rps-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-limit-rps-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-limit-rps-zero-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami-zero.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-limit-rps-zero-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-limit-rps-zero-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-limit-rps-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-limit-rps-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-limit-rps-rule-0-path-0-limit-rps": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Second),
							},
						},
						"default-ingress-with-limit-rps-rule-0-path-0-tls-limit-rps": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Second),
							},
						},
						"default-ingress-with-limit-rps-zero-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-limit-rps-zero-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-limit-rps-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-limit-rps",
							},
						},
						"default-ingress-with-limit-rps-zero-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-limit-rps-zero",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-limit-rps": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-limit-rps-zero": {
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
						"default-ingress-with-limit-rpm-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-limit-rpm-rule-0-path-0-limit-rpm", "default-ingress-with-limit-rpm-rule-0-path-0-retry"},
							Service:     "default-ingress-with-limit-rpm-whoami-80",
						},
						"default-ingress-with-limit-rpm-zero-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami-zero.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-limit-rpm-zero-rule-0-path-0-retry"},
							Service:     "default-ingress-with-limit-rpm-zero-whoami-80",
						},
						"default-ingress-with-limit-rpm-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-limit-rpm-rule-0-path-0-tls-limit-rpm", "default-ingress-with-limit-rpm-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-limit-rpm-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-limit-rpm-zero-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami-zero.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-limit-rpm-zero-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-limit-rpm-zero-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-limit-rpm-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-limit-rpm-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-limit-rpm-rule-0-path-0-limit-rpm": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Minute),
							},
						},
						"default-ingress-with-limit-rpm-rule-0-path-0-tls-limit-rpm": {
							RateLimit: &dynamic.RateLimit{
								Average: 10,
								Burst:   50,
								Period:  ptypes.Duration(time.Minute),
							},
						},
						"default-ingress-with-limit-rpm-zero-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-limit-rpm-zero-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-limit-rpm-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-limit-rpm",
							},
						},
						"default-ingress-with-limit-rpm-zero-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-limit-rpm-zero",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-limit-rpm": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:     ptypes.Duration(60 * time.Second),
								ReadTimeout:     ptypes.Duration(60 * time.Second),
								WriteTimeout:    ptypes.Duration(60 * time.Second),
								IdleConnTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-limit-rpm-zero": {
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
						"default-ingress-with-limit-burst-multiplier-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami-burst.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-limit-burst-multiplier-rule-0-path-0-limit-rps", "default-ingress-with-limit-burst-multiplier-rule-0-path-0-retry"},
							Service:     "default-ingress-with-limit-burst-multiplier-whoami-80",
						},
						"default-ingress-with-limit-burst-multiplier-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami-burst.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-limit-burst-multiplier-rule-0-path-0-tls-limit-rps", "default-ingress-with-limit-burst-multiplier-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-limit-burst-multiplier-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"default-ingress-with-limit-burst-multiplier-zero-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("whoami-burst-zero.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-limit-burst-multiplier-zero-rule-0-path-0-limit-rps", "default-ingress-with-limit-burst-multiplier-zero-rule-0-path-0-retry"},
							Service:     "default-ingress-with-limit-burst-multiplier-zero-whoami-80",
						},
						"default-ingress-with-limit-burst-multiplier-zero-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("whoami-burst-zero.localhost") && Path("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-limit-burst-multiplier-zero-rule-0-path-0-tls-limit-rps", "default-ingress-with-limit-burst-multiplier-zero-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-limit-burst-multiplier-zero-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-limit-burst-multiplier-rule-0-path-0-retry":          {Retry: &dynamic.Retry{Attempts: 3}},
						"default-ingress-with-limit-burst-multiplier-rule-0-path-0-tls-retry":      {Retry: &dynamic.Retry{Attempts: 3}},
						"default-ingress-with-limit-burst-multiplier-zero-rule-0-path-0-retry":     {Retry: &dynamic.Retry{Attempts: 3}},
						"default-ingress-with-limit-burst-multiplier-zero-rule-0-path-0-tls-retry": {Retry: &dynamic.Retry{Attempts: 3}},
						"default-ingress-with-limit-burst-multiplier-rule-0-path-0-limit-rps": {
							RateLimit: &dynamic.RateLimit{Average: 10, Burst: 100, Period: ptypes.Duration(time.Second)},
						},
						"default-ingress-with-limit-burst-multiplier-rule-0-path-0-tls-limit-rps": {
							RateLimit: &dynamic.RateLimit{Average: 10, Burst: 100, Period: ptypes.Duration(time.Second)},
						},
						"default-ingress-with-limit-burst-multiplier-zero-rule-0-path-0-limit-rps": {
							RateLimit: &dynamic.RateLimit{Average: 10, Burst: 50, Period: ptypes.Duration(time.Second)},
						},
						"default-ingress-with-limit-burst-multiplier-zero-rule-0-path-0-tls-limit-rps": {
							RateLimit: &dynamic.RateLimit{Average: 10, Burst: 50, Period: ptypes.Duration(time.Second)},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-limit-burst-multiplier-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers:            []dynamic.Server{{URL: "http://10.10.0.1:80"}, {URL: "http://10.10.0.2:80"}},
								Strategy:           "wrr",
								PassHostHeader:     ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								ServersTransport:   "default-ingress-with-limit-burst-multiplier",
							},
						},
						"default-ingress-with-limit-burst-multiplier-zero-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers:            []dynamic.Server{{URL: "http://10.10.0.1:80"}, {URL: "http://10.10.0.2:80"}},
								Strategy:           "wrr",
								PassHostHeader:     ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{FlushInterval: dynamic.DefaultFlushInterval},
								ServersTransport:   "default-ingress-with-limit-burst-multiplier-zero",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-limit-burst-multiplier": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{DialTimeout: ptypes.Duration(60 * time.Second), ReadTimeout: ptypes.Duration(60 * time.Second), WriteTimeout: ptypes.Duration(60 * time.Second), IdleConnTimeout: ptypes.Duration(60 * time.Second)},
						},
						"default-ingress-with-limit-burst-multiplier-zero": {
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
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
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
			strictValidatePathType: ptr.To(false),
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-ingress-with-use-regex-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("use-regex.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-use-regex-rule-0-path-0-retry"},
							Service:     "default-ingress-with-use-regex-whoami-80",
						},
						"default-ingress-with-use-regex-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("use-regex.localhost") && PathRegexp("(?i)^/test(.*)")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-use-regex-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-use-regex-whoami-80",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-use-regex-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
						"default-ingress-with-use-regex-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-use-regex-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:       "wrr",
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
								ServersTransport: "default-ingress-with-use-regex",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-use-regex": {
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
						"default-ingress-with-wildcard-host-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("*.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-wildcard-host-rule-0-path-0-retry"},
							Service:     "default-ingress-with-wildcard-host-whoami-80",
						},
						"default-ingress-with-wildcard-host-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("*.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-ingress-with-wildcard-host-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-wildcard-host-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-wildcard-host-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-wildcard-host-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-wildcard-host-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-wildcard-host",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-wildcard-host": {
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
						"default-ingress-with-wildcard-host-tls-rule-0-path-0": {
							EntryPoints: []string{"http"},
							Rule:        `Host("*.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-wildcard-host-tls-rule-0-path-0-retry"},
							Service:     "default-ingress-with-wildcard-host-tls-whoami-80",
						},
						"default-ingress-with-wildcard-host-tls-rule-0-path-0-tls": {
							EntryPoints: []string{"https"},
							Rule:        `Host("*.localhost") && PathPrefix("/")`,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-ingress-with-wildcard-host-tls-rule-0-path-0-tls-retry"},
							Service:     "default-ingress-with-wildcard-host-tls-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-wildcard-host-tls-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-wildcard-host-tls-rule-0-path-0-tls-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-wildcard-host-tls-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "http://10.10.0.1:80"},
									{URL: "http://10.10.0.2:80"},
								},
								Strategy:         "wrr",
								PassHostHeader:   ptr.To(true),
								ServersTransport: "default-ingress-with-wildcard-host-tls",
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: dynamic.DefaultFlushInterval,
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-wildcard-host-tls": {
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
							Certificate: tls.Certificate{
								CertFile: "-----BEGIN CERTIFICATE-----",
								KeyFile:  "-----BEGIN CERTIFICATE-----",
							},
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

			k8sObjects := readResources(t, test.paths)
			kubeClient := kubefake.NewClientset(k8sObjects...)
			client := newClient(kubeClient)

			eventCh, err := client.WatchAll(t.Context(), "", "")
			require.NoError(t, err)

			if len(k8sObjects) > 0 {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				k8sClient:                      client,
				defaultBackendServiceName:      test.defaultBackendServiceName,
				defaultBackendServiceNamespace: test.defaultBackendServiceNamespace,
				AllowSnippetAnnotations:        test.allowSnippetAnnotations,
				NonTLSEntryPoints:              []string{"http"},
				TLSEntryPoints:                 []string{"https"},
				allowedHeaders:                 test.globalAllowedResponseHeaders,
				AllowCrossNamespaceResources:   test.allowCrossNamespaceResources,
				GlobalAuthURL:                  test.globalAuthURL,
			}
			p.SetDefaults()
			if test.strictValidatePathType != nil {
				p.StrictValidatePathType = *test.strictValidatePathType
			}

			conf := p.loadConfiguration(t.Context())
			assert.Equal(t, test.expected, conf)
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

func TestProvider_validateConfiguration(t *testing.T) {
	testCases := []struct {
		desc                            string
		globalAllowedResponseHeaders    []string
		expectedAllowedHeaders          []string
		defaultBackendService           string
		expectedBackendServiceName      string
		expectedBackendServiceNamespace string
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			provider := &Provider{
				GlobalAllowedResponseHeaders: test.globalAllowedResponseHeaders,
				DefaultBackendService:        test.defaultBackendService,
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
