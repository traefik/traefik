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
							Middlewares: []string{"default-ingress-with-custom-headers-rule-0-path-0-custom-headers", "default-ingress-with-custom-headers-rule-0-path-0-retry"},
							Service:     "default-ingress-with-custom-headers-whoami-80",
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
						"default-ingress-with-no-annotation-rule-0-path-0": {
							Rule:        "Host(`whoami.localhost`) && PathPrefix(`/`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-ingress-with-no-annotation-rule-0-path-0-retry"},
							Service:     "default-ingress-with-no-annotation-whoami-80",
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
						"default-ingress-with-no-annotation-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whoami.localhost`) && Path(`/basicauth`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-basicauth-rule-0-path-0-basic-auth", "default-ingress-with-basicauth-rule-0-path-0-retry"},
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
						"default-ingress-with-basicauth-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whoami.localhost`) && Path(`/forwardauth`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-forwardauth-rule-0-path-0-forward-auth", "default-ingress-with-forwardauth-rule-0-path-0-retry"},
							Service:     "default-ingress-with-forwardauth-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-forwardauth-rule-0-path-0-forward-auth": {
							ForwardAuth: &dynamic.ForwardAuth{
								Address:             "http://whoami.default.svc/",
								AuthResponseHeaders: []string{"X-Foo"},
								AuthSigninURL:       "https://auth.example.com/oauth2/start?rd=foo",
								Interpolate:         true,
							},
						},
						"default-ingress-with-forwardauth-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
						"default-ingress-with-ssl-redirect-rule-0-path-0": {
							Rule:        "Host(`sslredirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-ingress-with-ssl-redirect-rule-0-path-0-retry"},
							Service:     "default-ingress-with-ssl-redirect-whoami-80",
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
							Rule:        "Host(`withoutsslredirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
							Middlewares: []string{"default-ingress-without-ssl-redirect-rule-0-path-0-retry"},
							Service:     "default-ingress-without-ssl-redirect-whoami-80",
						},
						"default-ingress-with-force-ssl-redirect-rule-0-path-0": {
							Rule:        "Host(`forcesslredirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-force-ssl-redirect-rule-0-path-0-redirect-scheme", "default-ingress-with-force-ssl-redirect-rule-0-path-0-retry"},
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
						"default-ingress-with-ssl-redirect-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-without-ssl-redirect-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-force-ssl-redirect-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-without-ssl-redirect": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-force-ssl-redirect": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`sticky.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-sticky-rule-0-path-0-retry"},
							Service:     "default-ingress-with-sticky-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-sticky-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`proxy-ssl.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-ssl-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-ssl-whoami-tls-443",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-ssl-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`cors.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-cors-rule-0-path-0-cors", "default-ingress-with-cors-rule-0-path-0-retry"},
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
						"default-ingress-with-cors-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`service-upstream.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-service-upstream-rule-0-path-0-retry"},
							Service:     "default-ingress-with-service-upstream-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-service-upstream-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`upstream-vhost.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-upstream-vhost-rule-0-path-0-vhost", "default-ingress-with-upstream-vhost-rule-0-path-0-retry"},
							Service:     "default-ingress-with-upstream-vhost-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-upstream-vhost-rule-0-path-0-vhost": {
							Headers: &dynamic.Headers{
								CustomRequestHeaders: map[string]string{"Host": "upstream-host-header-value"},
							},
						},
						"default-ingress-with-upstream-vhost-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`use-regex.localhost`) && PathRegexp(`^/test(.*)`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-use-regex-rule-0-path-0-retry"},
							Service:     "default-ingress-with-use-regex-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-use-regex-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`rewrite-target.localhost`) && PathRegexp(`^/something(/|$)(.*)`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-rewrite-target-whoami-80",
							Middlewares: []string{"default-ingress-with-rewrite-target-rule-0-path-0-rewrite-target", "default-ingress-with-rewrite-target-rule-0-path-0-retry"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-rewrite-target-rule-0-path-0-rewrite-target": {
							ReplacePathRegex: &dynamic.ReplacePathRegex{
								Regex:       "/something(/|$)(.*)",
								Replacement: "/$2",
							},
						},
						"default-ingress-with-rewrite-target-rule-0-path-0-retry": {
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
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-ingress-with-rewrite-target": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`rewrite-target-no-regex.localhost`) && Path(`/original`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-rewrite-target-no-regex-whoami-80",
							Middlewares: []string{"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-rewrite-target", "default-ingress-with-rewrite-target-no-regex-rule-0-path-0-retry"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-rewrite-target": {
							ReplacePath: &dynamic.ReplacePath{
								Path: "/rewritten",
							},
						},
						"default-ingress-with-rewrite-target-no-regex-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`app-root.localhost`) && (Path(`/bar`) || PathPrefix(`/bar/`))",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-app-root-whoami-80",
							Middlewares: []string{"default-ingress-with-app-root-rule-0-path-0-app-root", "default-ingress-with-app-root-rule-0-path-0-retry"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-app-root-rule-0-path-0-app-root": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`app-root.localhost`) && (Path(`/bar`) || PathPrefix(`/bar/`))",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-app-root-rule-0-path-0-retry"},
							Service:     "default-ingress-with-app-root-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-app-root-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`www.host.localhost`) && PathPrefix(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-www-host-rule-0-path-0-retry"},
							Service:     "default-ingress-with-www-host-whoami-80",
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`host.localhost`) && PathPrefix(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-host-rule-0-path-0-retry"},
							Service:     "default-ingress-with-host-whoami-80",
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`host.localhost`) && PathPrefix(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-host-rule-0-path-0-retry"},
							Service:     "default-ingress-with-host-whoami-80",
						},
						"default-ingress-with-www-host-rule-0-path-0": {
							Rule:        "Host(`www.host.localhost`) && PathPrefix(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-www-host-rule-0-path-0-retry"},
							Service:     "default-ingress-with-www-host-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-host-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 3,
							},
						},
						"default-ingress-with-www-host-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-host": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whitelist-source-range.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-single-ip-rule-0-path-0-allowed-source-range", "default-ingress-with-whitelist-single-ip-rule-0-path-0-retry"},
							Service:     "default-ingress-with-whitelist-single-ip-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-whitelist-single-ip-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.20.1"},
							},
						},
						"default-ingress-with-whitelist-single-ip-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whitelist-source-range.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-single-cidr-rule-0-path-0-allowed-source-range", "default-ingress-with-whitelist-single-cidr-rule-0-path-0-retry"},
							Service:     "default-ingress-with-whitelist-single-cidr-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-whitelist-single-cidr-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24"},
							},
						},
						"default-ingress-with-whitelist-single-cidr-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whitelist-source-range.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-allowed-source-range", "default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-retry"},
							Service:     "default-ingress-with-whitelist-multiple-ip-and-cidr-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24", "10.0.0.0/8", "192.168.20.1"},
							},
						},
						"default-ingress-with-whitelist-multiple-ip-and-cidr-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whitelist-source-range.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-whitelist-empty-rule-0-path-0-retry"},
							Service:     "default-ingress-with-whitelist-empty-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-whitelist-empty-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`allowlist-source-range.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-allowlist-empty-rule-0-path-0-retry"},
							Service:     "default-ingress-with-allowlist-empty-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-allowlist-empty-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`allowlist-source-range.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-allowlist-single-ip-rule-0-path-0-allowed-source-range", "default-ingress-with-allowlist-single-ip-rule-0-path-0-retry"},
							Service:     "default-ingress-with-allowlist-single-ip-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-allowlist-single-ip-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.20.1"},
							},
						},
						"default-ingress-with-allowlist-single-ip-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`allowlist-source-range.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-allowlist-single-cidr-rule-0-path-0-allowed-source-range", "default-ingress-with-allowlist-single-cidr-rule-0-path-0-retry"},
							Service:     "default-ingress-with-allowlist-single-cidr-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-allowlist-single-cidr-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24"},
							},
						},
						"default-ingress-with-allowlist-single-cidr-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`allowlist-source-range.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0-allowed-source-range", "default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0-retry"},
							Service:     "default-ingress-with-allowlist-multiple-ip-and-cidr-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0-allowed-source-range": {
							IPAllowList: &dynamic.IPAllowList{
								SourceRange: []string{"192.168.1.0/24", "10.0.0.0/8", "192.168.20.1"},
							},
						},
						"default-ingress-with-allowlist-multiple-ip-and-cidr-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`permanent-redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-permanent-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-permanent-redirect-rule-0-path-0-redirect", "default-ingress-with-permanent-redirect-rule-0-path-0-retry"},
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
						"default-ingress-with-permanent-redirect-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`permanent-redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-permanent-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-permanent-redirect-rule-0-path-0-redirect", "default-ingress-with-permanent-redirect-rule-0-path-0-retry"},
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
						"default-ingress-with-permanent-redirect-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`permanent-redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-permanent-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-permanent-redirect-rule-0-path-0-redirect", "default-ingress-with-permanent-redirect-rule-0-path-0-retry"},
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
						"default-ingress-with-permanent-redirect-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-redirect-rule-0-path-0-redirect", "default-ingress-with-redirect-rule-0-path-0-retry"},
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
						"default-ingress-with-redirect-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`temporal-redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-temporal-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-temporal-redirect-rule-0-path-0-redirect", "default-ingress-with-temporal-redirect-rule-0-path-0-retry"},
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
						"default-ingress-with-temporal-redirect-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`temporal-redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-temporal-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-temporal-redirect-rule-0-path-0-redirect", "default-ingress-with-temporal-redirect-rule-0-path-0-retry"},
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
						"default-ingress-with-temporal-redirect-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`temporal-redirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-temporal-redirect-whoami-80",
							Middlewares: []string{"default-ingress-with-temporal-redirect-rule-0-path-0-redirect", "default-ingress-with-temporal-redirect-rule-0-path-0-retry"},
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
						"default-ingress-with-temporal-redirect-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whoami.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-timeout-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-timeout-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-timeout-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(30 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whoami.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-timeout-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-timeout-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-timeout-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(30 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whoami.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-timeout-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-timeout-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-timeout-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(30 * time.Second),
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
						"default-ingress-with-auth-tls-secret-rule-0-path-0": {
							Rule:        "Host(`auth-tls-secret.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-auth-tls-secret-rule-0-path-0-retry"},
							Service:     "default-ingress-with-auth-tls-secret-whoami-80",
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
						"default-ingress-with-auth-tls-secret-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
						"default-ingress-with-auth-tls-verify-client-rule-0-path-0": {
							Rule:        "Host(`auth-tls-verify-client.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-auth-tls-verify-client-rule-0-path-0-retry"},
							Service:     "default-ingress-with-auth-tls-verify-client-whoami-80",
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
						"default-ingress-with-auth-tls-verify-client-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whoami.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-custom-http-errors-and-default-backend-whoami-80",
							Middlewares: []string{"default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-custom-http-errors", "default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-retry"},
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
						"default-ingress-with-custom-http-errors-and-default-backend-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
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
										URL: "http://10.10.0.7:8000",
									},
									{
										URL: "http://10.10.0.8:8000",
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
						"default-ingress-with-custom-http-errors-and-default-backend": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:                           "Custom HTTP Errors",
			defaultBackendServiceName:      "whoami_b",
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
							Rule:        "Host(`whoami.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-custom-http-errors-whoami-80",
							Middlewares: []string{"default-ingress-with-custom-http-errors-rule-0-path-0-custom-http-errors", "default-ingress-with-custom-http-errors-rule-0-path-0-retry"},
						},
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
						"default-ingress-with-custom-http-errors-rule-0-path-0-retry": {
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
						"default-backend": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:8000",
									},
									{
										URL: "http://10.10.0.8:8000",
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whoami.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-default-backend-annotation-empty-80",
							Middlewares: []string{"default-ingress-with-default-backend-annotation-rule-0-path-0-retry"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-default-backend-annotation-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{Attempts: 3},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-ingress-with-default-backend-annotation-empty-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.7:8000",
									},
									{
										URL: "http://10.10.0.8:8000",
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`hostname.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-body-size-rule-0-path-0-buffering", "default-ingress-with-proxy-body-size-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-body-size-whoami-80",
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
						"default-ingress-with-proxy-body-size-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`hostname.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-client-body-buffer-size-rule-0-path-0-buffering", "default-ingress-with-client-body-buffer-size-rule-0-path-0-retry"},
							Service:     "default-ingress-with-client-body-buffer-size-whoami-80",
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
						"default-ingress-with-client-body-buffer-size-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`hostname.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-body-size-and-client-body-buffer-size-rule-0-path-0-buffering", "default-ingress-with-proxy-body-size-and-client-body-buffer-size-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-body-size-and-client-body-buffer-size-whoami-80",
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
						"default-ingress-with-proxy-body-size-and-client-body-buffer-size-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`hostname.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-buffer-size-rule-0-path-0-buffering", "default-ingress-with-proxy-buffer-size-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-buffer-size-whoami-80",
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
						"default-ingress-with-proxy-buffer-size-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`hostname.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-buffers-number-rule-0-path-0-buffering", "default-ingress-with-proxy-buffers-number-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-buffers-number-whoami-80",
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
						"default-ingress-with-proxy-buffers-number-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`hostname.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-buffer-size-and-number-rule-0-path-0-buffering", "default-ingress-with-proxy-buffer-size-and-number-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-buffer-size-and-number-whoami-80",
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
						"default-ingress-with-proxy-buffer-size-and-number-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`hostname.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-proxy-max-temp-file-size-rule-0-path-0-buffering", "default-ingress-with-proxy-max-temp-file-size-rule-0-path-0-retry"},
							Service:     "default-ingress-with-proxy-max-temp-file-size-whoami-80",
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
						"default-ingress-with-proxy-max-temp-file-size-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
						"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0": {
							Rule:        "Host(`auth-tls-pass-certificate-to-upstream.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-pass-certificate-to-upstream", "default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-retry"},
							Service:     "default-ingress-with-auth-tls-pass-certificate-to-upstream-whoami-80",
							TLS: &dynamic.RouterTLSConfig{
								Options: "default-ingress-with-auth-tls-pass-certificate-to-upstream-default-ca-secret",
							},
						},
						"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-http": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`auth-tls-pass-certificate-to-upstream.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-redirect-scheme"},
							Service:     "noop@internal",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-pass-certificate-to-upstream": {
							AuthTLSPassCertificateToUpstream: &dynamic.AuthTLSPassCertificateToUpstream{
								ClientAuthType: tls.RequireAndVerifyClientCert,
								CAFiles:        nil,
							},
						},
						"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:                 "https",
								ForcePermanentRedirect: true,
							},
						},
						"default-ingress-with-auth-tls-pass-certificate-to-upstream-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whoami.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-next-upstream-rule-0-path-0-retry"},
						},
						"default-ingress-with-proxy-next-upstream-off-rule-0-path-0": {
							Rule:       "Host(`whoami.localhost`) && Path(`/`)",
							RuleSyntax: "default",
							Service:    "default-ingress-with-proxy-next-upstream-off-whoami-80",
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-proxy-next-upstream-off": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whoami.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-tries-unlimited-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-next-upstream-tries-unlimited-rule-0-path-0-retry"},
						},
						"default-ingress-with-proxy-next-upstream-tries-rule-0-path-0": {
							Rule:        "Host(`whoami.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-tries-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-next-upstream-tries-rule-0-path-0-retry"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-next-upstream-tries-unlimited-rule-0-path-0-retry": {
							Retry: &dynamic.Retry{
								Attempts: 2,
							},
						},
						"default-ingress-with-proxy-next-upstream-tries-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
							},
						},
						"default-ingress-with-proxy-next-upstream-tries": {
							ForwardingTimeouts: &dynamic.ForwardingTimeouts{
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
							Rule:        "Host(`whoami.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Service:     "default-ingress-with-proxy-next-upstream-timeout-whoami-80",
							Middlewares: []string{"default-ingress-with-proxy-next-upstream-timeout-rule-0-path-0-retry"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-proxy-next-upstream-timeout-rule-0-path-0-retry": {
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
								DialTimeout:  ptypes.Duration(60 * time.Second),
								ReadTimeout:  ptypes.Duration(60 * time.Second),
								WriteTimeout: ptypes.Duration(60 * time.Second),
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
