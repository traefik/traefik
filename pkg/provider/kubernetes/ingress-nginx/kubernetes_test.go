package ingressnginx

import (
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
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
			},
		},
		{
			desc: "Custom Headers",
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
							Rule:        "Host(`whoami.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-custom-headers-rule-0-path-0-custom-headers"},
							Service:     "default-ingress-with-custom-headers-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-custom-headers-rule-0-path-0-custom-headers": {
							Headers: &dynamic.Headers{
								CustomResponseHeaders: map[string]string{"X-Custom-Header": "some-random-string"},
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
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-custom-headers": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
						"default-ingress-with-no-annotation-rule-0-path-0": {
							Rule:       "Host(`whoami.localhost`) && PathPrefix(`/`)",
							RuleSyntax: "default",
							TLS:        &dynamic.RouterTLSConfig{},
							Service:    "default-ingress-with-no-annotation-whoami-80",
						},
						"default-ingress-with-no-annotation-rule-0-path-0-http": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`whoami.localhost`) && PathPrefix(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-no-annotation-rule-0-path-0-redirect-scheme"},
							Service:     "noop@internal",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-no-annotation-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
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
								DialTimeout: ptypes.Duration(60 * time.Second),
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
					Options: map[string]tls.Options{},
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
							Rule:        "Host(`whoami.localhost`) && Path(`/basicauth`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-basicauth-rule-0-path-0-basic-auth"},
							Service:     "default-ingress-with-basicauth-whoami-80",
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`whoami.localhost`) && Path(`/forwardauth`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-forwardauth-rule-0-path-0-forward-auth"},
							Service:     "default-ingress-with-forwardauth-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-forwardauth-rule-0-path-0-forward-auth": {
							ForwardAuth: &dynamic.ForwardAuth{
								Address:             "http://whoami.default.svc/",
								AuthResponseHeaders: []string{"X-Foo"},
								AuthSigninURL:       "https://auth.example.com/oauth2/start?rd=foo",
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
						"default-ingress-with-ssl-redirect-rule-0-path-0": {
							Rule:       "Host(`sslredirect.localhost`) && Path(`/`)",
							RuleSyntax: "default",
							TLS:        &dynamic.RouterTLSConfig{},
							Service:    "default-ingress-with-ssl-redirect-whoami-80",
						},
						"default-ingress-with-ssl-redirect-rule-0-path-0-http": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`sslredirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-ssl-redirect-rule-0-path-0-redirect-scheme"},
							Service:     "noop@internal",
						},
						"default-ingress-without-ssl-redirect-rule-0-path-0-http": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`withoutsslredirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-without-ssl-redirect-whoami-80",
						},
						"default-ingress-without-ssl-redirect-rule-0-path-0": {
							Rule:       "Host(`withoutsslredirect.localhost`) && Path(`/`)",
							RuleSyntax: "default",
							TLS:        &dynamic.RouterTLSConfig{},
							Service:    "default-ingress-without-ssl-redirect-whoami-80",
						},
						"default-ingress-with-force-ssl-redirect-rule-0-path-0": {
							Rule:        "Host(`forcesslredirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-force-ssl-redirect-rule-0-path-0-redirect-scheme"},
							Service:     "default-ingress-with-force-ssl-redirect-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-ssl-redirect-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-ingress-with-force-ssl-redirect-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-without-ssl-redirect": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-force-ssl-redirect": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout: ptypes.Duration(60 * time.Second),
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
					Options: map[string]tls.Options{},
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
							Rule:       "HostSNI(`passthrough.whoami.localhost`)",
							RuleSyntax: "default",
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
										Address: "10.10.0.5:8443",
									},
									{
										Address: "10.10.0.6:8443",
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
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:       "Host(`sticky.localhost`) && Path(`/`)",
							RuleSyntax: "default",
							Service:    "default-ingress-with-sticky-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:       "Host(`proxy-ssl.localhost`) && Path(`/`)",
							RuleSyntax: "default",
							Service:    "default-ingress-with-proxy-ssl-whoami-tls-443",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-proxy-ssl-whoami-tls-443": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.5:8443",
									},
									{
										URL: "https://10.10.0.6:8443",
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`cors.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-cors-rule-0-path-0-cors"},
							Service:     "default-ingress-with-cors-whoami-80",
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:       "Host(`service-upstream.localhost`) && Path(`/`)",
							RuleSyntax: "default",
							Service:    "default-ingress-with-service-upstream-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`upstream-vhost.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-upstream-vhost-rule-0-path-0-vhost"},
							Service:     "default-ingress-with-upstream-vhost-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-upstream-vhost-rule-0-path-0-vhost": {
							Headers: &dynamic.Headers{
								CustomRequestHeaders: map[string]string{"Host": "upstream-host-header-value"},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:       "Host(`use-regex.localhost`) && PathRegexp(`^/test(.*)`)",
							RuleSyntax: "default",
							Service:    "default-ingress-with-use-regex-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`rewrite-target.localhost`) && PathRegexp(`^/something(/|$)(.*)`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-rewrite-target-whoami-80",
							Middlewares: []string{"default-ingress-with-rewrite-target-rule-0-path-0-rewrite-target"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-rewrite-target-rule-0-path-0-rewrite-target": {
							ReplacePathRegex: &dynamic.ReplacePathRegex{
								Regex:       "/something(/|$)(.*)",
								Replacement: "/$2",
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
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-rewrite-target": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`app-root.localhost`) && (Path(`/bar`) || PathPrefix(`/bar/`))",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-app-root-whoami-80",
							Middlewares: []string{"default-ingress-with-app-root-rule-0-path-0-app-root"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-app-root-rule-0-path-0-app-root": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `^(https?://[^/]+)/$`,
								Replacement: "$1/foo",
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:       "Host(`app-root.localhost`) && (Path(`/bar`) || PathPrefix(`/bar/`))",
							RuleSyntax: "default",
							Service:    "default-ingress-with-app-root-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
			},
		},
		{
			desc: "From To WWW Redirect - www host",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/22-ingress-with-www-host.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-ingress-with-www-host-rule-0-path-0": {
							Rule:       "Host(`www.host.localhost`) && PathPrefix(`/`)",
							RuleSyntax: "default",
							Service:    "default-ingress-with-www-host-whoami-80",
						},
						"default-ingress-with-www-host-rule-0-path-0-from-to-www-redirect": {
							Rule:        "Host(`host.localhost`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-www-host-whoami-80",
							Middlewares: []string{"default-ingress-with-www-host-rule-0-path-0-from-to-www-redirect"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-www-host-rule-0-path-0-from-to-www-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `(https?)://[^/]+:([0-9]+)/(.*)`,
								Replacement: "$1://www.host.localhost:$2/$3",
								Permanent:   true,
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
			},
		},
		{
			desc: "From To WWW Redirect - host",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/22-ingress-with-host.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-ingress-with-host-rule-0-path-0": {
							Rule:       "Host(`host.localhost`) && PathPrefix(`/`)",
							RuleSyntax: "default",
							Service:    "default-ingress-with-host-whoami-80",
						},
						"default-ingress-with-host-rule-0-path-0-from-to-www-redirect": {
							Rule:        "Host(`www.host.localhost`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-host-whoami-80",
							Middlewares: []string{"default-ingress-with-host-rule-0-path-0-from-to-www-redirect"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-host-rule-0-path-0-from-to-www-redirect": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       `(https?)://[^/]+:([0-9]+)/(.*)`,
								Replacement: "$1://host.localhost:$2/$3",
								Permanent:   true,
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
			},
		},
		{
			desc: "From To WWW Redirect - multiple ingresses",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/22-ingresses-with-www-redirect.yml",
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:  map[string]*dynamic.TCPRouter{},
					Services: map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-ingress-with-host-rule-0-path-0": {
							Rule:       "Host(`host.localhost`) && PathPrefix(`/`)",
							RuleSyntax: "default",
							Service:    "default-ingress-with-host-whoami-80",
						},
						"default-ingress-with-www-host-rule-0-path-0": {
							Rule:       "Host(`www.host.localhost`) && PathPrefix(`/`)",
							RuleSyntax: "default",
							Service:    "default-ingress-with-www-host-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-host": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:       "PathPrefix(`/`)",
							RuleSyntax: "default",
							Priority:   math.MinInt32,
							Service:    "default-backend",
						},
						"default-backend-tls": {
							Rule:       "PathPrefix(`/`)",
							RuleSyntax: "default",
							Priority:   math.MinInt32,
							TLS:        &dynamic.RouterTLSConfig{},
							Service:    "default-backend",
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
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`whitelist-source-range.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-single-ip-rule-0-path-0-whitelist-source-range"},
							Service:     "default-ingress-with-whitelist-single-ip-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-whitelist-single-ip-rule-0-path-0-whitelist-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.20.1"},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`whitelist-source-range.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-single-cidr-rule-0-path-0-whitelist-source-range"},
							Service:     "default-ingress-with-whitelist-single-cidr-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-whitelist-single-cidr-rule-0-path-0-whitelist-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24"},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`whitelist-source-range.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-whitelist-source-range"},
							Service:     "default-ingress-with-whitelist-multiple-ip-and-cidr-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-whitelist-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24", "10.0.0.0/8", "192.168.20.1"},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`whitelist-source-range.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: nil,
							Service:     "default-ingress-with-whitelist-empty-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`permanent-redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-permanent-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-permanent-redirect-rule-0-path-0-redirect"},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`permanent-redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-permanent-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-permanent-redirect-rule-0-path-0-redirect"},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`permanent-redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-permanent-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-permanent-redirect-rule-0-path-0-redirect"},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-redirect-rule-0-path-0-redirect"},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`temporal-redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-temporal-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-temporal-redirect-rule-0-path-0-redirect"},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`temporal-redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-temporal-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-temporal-redirect-rule-0-path-0-redirect"},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:        "Host(`temporal-redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-temporal-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-temporal-redirect-rule-0-path-0-redirect"},
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
								DialTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
							Rule:       "Host(`whoami.localhost`) && Path(`/`)",
							RuleSyntax: "default",
							Service:    "default-ingress-with-proxy-timeout-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
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
								DialTimeout: ptypes.Duration(30 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{
					Options: map[string]tls.Options{},
				},
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
						"default-ingress-with-auth-tls-secret-rule-0-path-0": {
							Rule:       "Host(`auth-tls-secret.localhost`) && Path(`/`)",
							RuleSyntax: "default",
							Service:    "default-ingress-with-auth-tls-secret-whoami-80",
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-ingress-with-auth-tls-secret-default-ca-secret",
							},
						},
						"default-ingress-with-auth-tls-secret-rule-0-path-0-http": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`auth-tls-secret.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-auth-tls-secret-rule-0-path-0-redirect-scheme"},
							Service:     "noop@internal",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-auth-tls-secret-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
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
								DialTimeout: ptypes.Duration(60 * time.Second),
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
								ClientAuthType: "RequireAndVerifyClientCert",
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
						"default-ingress-with-auth-tls-verify-client-rule-0-path-0": {
							Rule:       "Host(`auth-tls-verify-client.localhost`) && Path(`/`)",
							RuleSyntax: "default",
							Service:    "default-ingress-with-auth-tls-verify-client-whoami-80",
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-ingress-with-auth-tls-verify-client-default-ca-secret",
							},
						},
						"default-ingress-with-auth-tls-verify-client-rule-0-path-0-http": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`auth-tls-verify-client.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-auth-tls-verify-client-rule-0-path-0-redirect-scheme"},
							Service:     "noop@internal",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-auth-tls-verify-client-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
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
								DialTimeout: ptypes.Duration(60 * time.Second),
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
								ClientAuthType: "VerifyClientCertIfGiven",
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
				NonTLSEntryPoints:              []string{"web"},
			}
			p.SetDefaults()

			conf := p.loadConfiguration(t.Context())
			assert.Equal(t, test.expected, conf)
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
