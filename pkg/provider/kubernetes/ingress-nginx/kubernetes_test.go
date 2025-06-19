package ingressnginx

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				"ingresses/01-ingress-with-basicauth.yml",
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
			desc: "Basic Auth",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/01-ingress-with-basicauth.yml",
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
							Service:     "default-whoami-80",
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
						"default-whoami-80": {
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
				"ingresses/02-ingress-with-forwardauth.yml",
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
							Service:     "default-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-forwardauth-rule-0-path-0-forward-auth": {
							ForwardAuth: &dynamic.ForwardAuth{
								Address:             "http://whoami.default.svc/",
								AuthResponseHeaders: []string{"X-Foo"},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-whoami-80": {
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
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
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
				"ingresses/03-ingress-with-ssl-redirect.yml",
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
							Service:    "default-whoami-80",
						},
						"default-ingress-with-ssl-redirect-rule-0-path-0-redirect": {
							Rule:        "Host(`sslredirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-ssl-redirect-rule-0-path-0-redirect-scheme"},
							Service:     "noop@internal",
						},
						"default-ingress-without-ssl-redirect-rule-0-path-0-http": {
							Rule:       "Host(`withoutsslredirect.localhost`) && Path(`/`)",
							RuleSyntax: "default",
							Service:    "default-whoami-80",
						},
						"default-ingress-without-ssl-redirect-rule-0-path-0": {
							Rule:       "Host(`withoutsslredirect.localhost`) && Path(`/`)",
							RuleSyntax: "default",
							TLS:        &dynamic.RouterTLSConfig{},
							Service:    "default-whoami-80",
						},
						"default-ingress-with-force-ssl-redirect-rule-0-path-0": {
							Rule:       "Host(`forcesslredirect.localhost`) && Path(`/`)",
							RuleSyntax: "default",
							Service:    "default-whoami-80",
						},
						"default-ingress-with-force-ssl-redirect-rule-0-path-0-redirect": {
							Rule:        "Host(`forcesslredirect.localhost`) && Path(`/`)",
							RuleSyntax:  "default",
							Middlewares: []string{"default-ingress-with-force-ssl-redirect-rule-0-path-0-redirect-scheme"},
							Service:     "noop@internal",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-ingress-with-ssl-redirect-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:    "https",
								Permanent: true,
							},
						},
						"default-ingress-with-force-ssl-redirect-rule-0-path-0-redirect-scheme": {
							RedirectScheme: &dynamic.RedirectScheme{
								Scheme:    "https",
								Permanent: true,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-whoami-80": {
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
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
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
				"ingresses/04-ingress-with-ssl-passthrough.yml",
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
				"ingresses/06-ingress-with-sticky.yml",
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
							Service:    "default-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-whoami-80": {
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
								Sticky: &dynamic.Sticky{
									Cookie: &dynamic.Cookie{
										Name:     "foobar",
										Domain:   "foo.localhost",
										HTTPOnly: true,
										MaxAge:   42,
										Path:     ptr.To("/foobar"),
										SameSite: "none",
										Secure:   true,
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
			desc: "Proxy SSL",
			paths: []string{
				"services.yml",
				"secrets.yml",
				"ingressclasses.yml",
				"ingresses/07-ingress-with-proxy-ssl.yml",
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
							Service:    "default-whoami-tls-443",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-whoami-tls-443": {
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
							InsecureSkipVerify: true,
							RootCAs:            []types.FileOrContent{"-----BEGIN CERTIFICATE-----"},
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
				"ingresses/08-ingress-with-cors.yml",
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
							Service:     "default-whoami-80",
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
						"default-whoami-80": {
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
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc: "Service Upstream",
			paths: []string{
				"services.yml",
				"ingressclasses.yml",
				"ingresses/09-ingress-with-service-upstream.yml",
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
							Service:    "default-whoami-80",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-whoami-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.10.1:80",
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
