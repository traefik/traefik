package gateway

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
	gatev1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

var _ provider.Provider = (*Provider)(nil)

func TestLoadHTTPRoutes(t *testing.T) {
	testCases := []struct {
		desc                string
		ingressClass        string
		paths               []string
		expected            *dynamic.Configuration
		entryPoints         map[string]Entrypoint
		experimentalChannel bool
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
			desc:  "Empty because missing entry point",
			paths: []string{"services.yml", "httproute/simple.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":443",
			}},
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
			desc:  "Empty because no http route defined",
			paths: []string{"services.yml", "httproute/without_httproute.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
			desc: "Empty caused by missing GatewayClass",
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
			paths: []string{"services.yml", "httproute/without_gatewayclass.yml"},
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
			desc: "Empty caused by unknown GatewayClass controller desc",
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
			paths: []string{"services.yml", "httproute/gatewayclass_with_unknown_controller.yml"},
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
			desc: "Empty caused by multi ports service with wrong TargetPort",
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
			paths: []string{"services.yml", "httproute/with_wrong_service_port.yml"},
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
			desc: "Empty caused by HTTPS without TLS",
			entryPoints: map[string]Entrypoint{"websecure": {
				Address: ":443",
			}},
			paths: []string{"services.yml", "httproute/with_protocol_https_without_tls.yml"},
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
			desc: "Empty caused by HTTPS with TLS passthrough",
			entryPoints: map[string]Entrypoint{"websecure": {
				Address: ":443",
			}},
			paths: []string{"services.yml", "httproute/with_protocol_https_with_tls_passthrough.yml"},
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
			desc: "Empty caused by HTTPRoute with protocol TLS",
			entryPoints: map[string]Entrypoint{"websecure": {
				Address: ":443",
			}},
			paths: []string{"services.yml", "httproute/with_protocol_tls.yml"},
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
			desc: "Empty caused by HTTPRoute with protocol TCP",
			entryPoints: map[string]Entrypoint{"websecure": {
				Address: ":443",
			}},
			paths: []string{"services.yml", "httproute/with_protocol_tcp.yml"},
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
			desc: "Empty caused by TCPRoute with protocol HTTP",
			entryPoints: map[string]Entrypoint{"websecure": {
				Address: ":9000",
			}},
			paths: []string{"services.yml", "tcproute/with_protocol_http.yml"},
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
			desc: "Empty caused by TCPRoute with protocol HTTPS",
			entryPoints: map[string]Entrypoint{"websecure": {
				Address: ":9000",
			}},
			paths: []string{"services.yml", "tcproute/with_protocol_https.yml"},
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
			desc: "Empty caused by TLSRoute with protocol TCP",
			entryPoints: map[string]Entrypoint{"websecure": {
				Address: ":9001",
			}},
			paths: []string{"services.yml", "tlsroute/with_protocol_tcp.yml"},
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
			desc: "Empty caused by TLSRoute with protocol HTTP",
			entryPoints: map[string]Entrypoint{"websecure": {
				Address: ":9001",
			}},
			paths: []string{"services.yml", "tlsroute/with_protocol_http.yml"},
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
			desc: "Empty caused by TLSRoute with protocol HTTPS",
			entryPoints: map[string]Entrypoint{"websecure": {
				Address: ":9001",
			}},
			paths: []string{"services.yml", "tlsroute/with_protocol_https.yml"},
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
			desc: "Empty caused use http entrypoint with tls activated with HTTPRoute",
			entryPoints: map[string]Entrypoint{"websecure": {
				Address:        ":443",
				HasHTTPTLSConf: true,
			}},
			paths: []string{"services.yml", "httproute/simple_with_tls_entrypoint.yml"},
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
			desc: "Empty caused unsupported HTTPRoute rule",
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
			paths: []string{"services.yml", "httproute/simple_with_bad_rule.yml"},
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
			desc:  "Empty because no tcp route defined tls protocol",
			paths: []string{"services.yml", "tcproute/without_tcproute_tls_protocol.yml"},
			entryPoints: map[string]Entrypoint{"TCP": {
				Address: ":8080",
			}},
			experimentalChannel: true,
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
			desc: "Empty caused by HTTPRoute with TLS configuration",
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
			paths: []string{"services.yml", "httproute/with_tls_configuration.yml"},
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
			desc:  "Simple HTTPRoute",
			paths: []string{"services.yml", "httproute/simple.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
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
			desc:  "Simple HTTPRoute, with api@internal service",
			paths: []string{"services.yml", "httproute/simple_to_api_internal.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "api@internal",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
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
			desc:  "Simple HTTPRoute with protocol HTTPS",
			paths: []string{"services.yml", "httproute/with_protocol_https.yml"},
			entryPoints: map[string]Entrypoint{"websecure": {
				Address: ":443",
			}},
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
						"default-http-app-1-my-gateway-websecure-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"websecure"},
							Service:     "default-http-app-1-my-gateway-websecure-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-websecure-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
								CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
			},
		},
		{
			desc:  "Simple HTTPRoute, with multiple hosts",
			paths: []string{"services.yml", "httproute/with_multiple_host.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-66e726cd8903b49727ae": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-66e726cd8903b49727ae-wrr",
							Rule:        "(Host(`foo.com`) || Host(`bar.com`)) && PathPrefix(`/`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-66e726cd8903b49727ae-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
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
			desc:  "Simple HTTPRoute, with two hosts one wildcard",
			paths: []string{"services.yml", "with_two_hosts_one_wildcard.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-3b78e2feb3295ddd87f0": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-3b78e2feb3295ddd87f0-wrr",
							Rule:        "(Host(`foo.com`) || HostRegexp(`^[a-zA-Z0-9-]+\\.bar\\.com$`)) && PathPrefix(`/`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-3b78e2feb3295ddd87f0-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
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
			desc:  "Simple HTTPRoute, with one host and a wildcard",
			paths: []string{"services.yml", "with_two_hosts_wildcard.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-b0521a61fb43068694b4": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-b0521a61fb43068694b4-wrr",
							Rule:        "(Host(`foo.com`) || HostRegexp(`^[a-zA-Z0-9-]+\\.foo\\.com$`)) && PathPrefix(`/`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-b0521a61fb43068694b4-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
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
			desc:  "One HTTPRoute with two different rules",
			paths: []string{"services.yml", "httproute/two_rules.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
							Service:     "default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr",
						},
						"default-http-app-1-my-gateway-web-d737b4933fa88e68ab8a": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && Path(`/bir`)",
							RuleSyntax:  "v3",
							Service:     "default-http-app-1-my-gateway-web-d737b4933fa88e68ab8a-wrr",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-http-app-1-my-gateway-web-d737b4933fa88e68ab8a-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami2-8080",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami2-8080": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: ptr.To(true),
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
			desc:  "One HTTPRoute with one rule two targets",
			paths: []string{"services.yml", "httproute/one_rule_two_targets.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
							Service:     "default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
									{
										Name:   "default-whoami2-8080",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami2-8080": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.3:8080",
									},
									{
										URL: "http://10.10.0.4:8080",
									},
								},
								PassHostHeader: ptr.To(true),
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
			desc:  "Two Gateways and one HTTPRoute",
			paths: []string{"services.yml", "httproute/with_two_gateways_one_httproute.yml"},
			entryPoints: map[string]Entrypoint{
				"web": {
					Address: ":80",
				},
				"websecure": {
					Address: ":443",
				},
			},
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
						"default-http-app-1-my-gateway-http-web-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-http-web-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
						},
						"default-http-app-1-my-gateway-https-websecure-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"websecure"},
							Service:     "default-http-app-1-my-gateway-https-websecure-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-http-web-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-http-app-1-my-gateway-https-websecure-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
								CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
			},
		},
		{
			desc:  "Gateway with two listeners and one HTTPRoute",
			paths: []string{"services.yml", "httproute/with_two_listeners_one_httproute.yml"},
			entryPoints: map[string]Entrypoint{
				"web": {
					Address: ":80",
				},
				"websecure": {
					Address: ":443",
				},
			},
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
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
						},
						"default-http-app-1-my-gateway-websecure-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"websecure"},
							Service:     "default-http-app-1-my-gateway-websecure-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-http-app-1-my-gateway-websecure-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
								CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
			},
		},
		{
			desc:  "Simple HTTPRoute, with several rules",
			paths: []string{"services.yml", "httproute/with_several_rules.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-4a1b73e6f83804949a37": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-4a1b73e6f83804949a37-wrr",
							Rule:        "Host(`foo.com`) && PathPrefix(`/bar`) && Header(`my-header`,`foo`) && Header(`my-header2`,`bar`)",
							RuleSyntax:  "v3",
						},
						"default-http-app-1-my-gateway-web-aaba0f24fd26e1ca2276": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-aaba0f24fd26e1ca2276-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`) && Header(`my-header`,`bar`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-4a1b73e6f83804949a37-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-http-app-1-my-gateway-web-aaba0f24fd26e1ca2276-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
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
			desc:  "HTTPRoute with Same namespace selector",
			paths: []string{"services.yml", "httproute/with_namespace_same.yml"},
			entryPoints: map[string]Entrypoint{
				"web": {Address: ":80"},
			},
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
						"default-http-app-default-my-gateway-web-efde1997778109a1f6eb": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-default-my-gateway-web-efde1997778109a1f6eb-wrr",
							Rule:        "Host(`foo.com`) && Path(`/foo`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-default-my-gateway-web-efde1997778109a1f6eb-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
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
			desc:  "HTTPRoute with All namespace selector",
			paths: []string{"services.yml", "httproute/with_namespace_all.yml"},
			entryPoints: map[string]Entrypoint{
				"web": {Address: ":80"},
			},
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
						"default-http-app-default-my-gateway-web-efde1997778109a1f6eb": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-default-my-gateway-web-efde1997778109a1f6eb-wrr",
							Rule:        "Host(`foo.com`) && Path(`/foo`)",
							RuleSyntax:  "v3",
						},
						"bar-http-app-bar-my-gateway-web-66f5c78d03d948e36597": {
							EntryPoints: []string{"web"},
							Service:     "bar-http-app-bar-my-gateway-web-66f5c78d03d948e36597-wrr",
							Rule:        "Host(`bar.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-default-my-gateway-web-efde1997778109a1f6eb-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"bar-http-app-bar-my-gateway-web-66f5c78d03d948e36597-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "bar-whoami-bar-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"bar-whoami-bar-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.11:80",
									},
									{
										URL: "http://10.10.0.12:80",
									},
								},
								PassHostHeader: ptr.To(true),
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
			desc:  "HTTPRoute with namespace selector",
			paths: []string{"services.yml", "httproute/with_namespace_selector.yml"},
			entryPoints: map[string]Entrypoint{
				"web": {Address: ":80"},
			},
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
						"bar-http-app-bar-my-gateway-web-66f5c78d03d948e36597": {
							EntryPoints: []string{"web"},
							Service:     "bar-http-app-bar-my-gateway-web-66f5c78d03d948e36597-wrr",
							Rule:        "Host(`bar.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"bar-http-app-bar-my-gateway-web-66f5c78d03d948e36597-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "bar-whoami-bar-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"bar-whoami-bar-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.11:80",
									},
									{
										URL: "http://10.10.0.12:80",
									},
								},
								PassHostHeader: ptr.To(true),
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
			desc:  "Simple HTTPRoute, request header modifier",
			paths: []string{"services.yml", "httproute/filter_request_header_modifier.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4-wrr",
							Rule:        "Host(`example.org`) && PathPrefix(`/`)",
							RuleSyntax:  "v3",
							Middlewares: []string{"default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4-requestheadermodifier-0"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4-requestheadermodifier-0": {
							RequestHeaderModifier: &dynamic.RequestHeaderModifier{
								Set:    map[string]string{"X-Foo": "Bar"},
								Add:    map[string]string{"X-Bar": "Foo"},
								Remove: []string{"X-Baz"},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
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
			desc:  "Simple HTTPRoute, redirect HTTP to HTTPS",
			paths: []string{"services.yml", "httproute/filter_http_to_https.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4-wrr",
							Rule:        "Host(`example.org`) && PathPrefix(`/`)",
							RuleSyntax:  "v3",
							Middlewares: []string{"default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4-requestredirect-0"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4-requestredirect-0": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       "^[a-z]+:\\/\\/(?P<userInfo>.+@)?(?P<hostname>\\[[\\w:\\.]+\\]|[\\w\\._-]+)(?P<port>:\\d+)?\\/(?P<path>.*)",
								Replacement: "https://${userinfo}${hostname}${port}/${path}",
								Permanent:   true,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
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
			desc:  "Simple HTTPRoute, redirect HTTP to HTTPS with hostname",
			paths: []string{"services.yml", "httproute/filter_http_to_https_with_hostname_and_port.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4-wrr",
							Rule:        "Host(`example.org`) && PathPrefix(`/`)",
							RuleSyntax:  "v3",
							Middlewares: []string{"default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4-requestredirect-0"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4-requestredirect-0": {
							RedirectRegex: &dynamic.RedirectRegex{
								Regex:       "^[a-z]+:\\/\\/(?P<userInfo>.+@)?(?P<hostname>\\[[\\w:\\.]+\\]|[\\w\\._-]+)(?P<port>:\\d+)?\\/(?P<path>.*)",
								Replacement: "http://${userinfo}example.com:443/${path}",
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-364ce6ec04c3d49b19c4-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
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

			p := Provider{EntryPoints: test.entryPoints, ExperimentalChannel: test.experimentalChannel}

			conf := p.loadConfigurationFromGateway(context.Background(), newClientMock(test.paths...))
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadHTTPRoutes_backendExtensionRef(t *testing.T) {
	testCases := []struct {
		desc                  string
		paths                 []string
		groupKindBackendFuncs map[string]map[string]BuildBackendFunc
		expected              *dynamic.Configuration
		entryPoints           map[string]Entrypoint
	}{
		{
			desc:  "Simple HTTPRoute with TraefikService",
			paths: []string{"services.yml", "httproute/simple_with_TraefikService.yml"},
			groupKindBackendFuncs: map[string]map[string]BuildBackendFunc{
				traefikv1alpha1.GroupName: {"TraefikService": func(name, namespace string) (string, *dynamic.Service, error) {
					return name, nil, nil
				}},
			},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "whoami",
										Weight: func(i int) *int { return &i }(1),
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
			desc:  "Simple HTTPRoute with TraefikService with service configuration",
			paths: []string{"services.yml", "httproute/simple_with_TraefikService.yml"},
			groupKindBackendFuncs: map[string]map[string]BuildBackendFunc{
				traefikv1alpha1.GroupName: {"TraefikService": func(name, namespace string) (string, *dynamic.Service, error) {
					return name, &dynamic.Service{LoadBalancer: &dynamic.ServersLoadBalancer{Servers: []dynamic.Server{{URL: "foobar"}}}}, nil
				}},
			},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "whoami",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"whoami": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{URL: "foobar"},
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
			desc:  "Simple HTTPRoute with invalid TraefikService kind",
			paths: []string{"services.yml", "httproute/simple_with_TraefikService.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
			desc:  "Simple HTTPRoute with backendFunc error",
			paths: []string{"services.yml", "httproute/simple_with_TraefikService.yml"},
			groupKindBackendFuncs: map[string]map[string]BuildBackendFunc{
				traefikv1alpha1.GroupName: {"TraefikService": func(name, namespace string) (string, *dynamic.Service, error) {
					return "", nil, errors.New("BOOM")
				}},
			},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
			desc:  "Simple HTTPRoute, with myservice@file service",
			paths: []string{"services.yml", "httproute/simple_cross_provider.yml"},
			groupKindBackendFuncs: map[string]map[string]BuildBackendFunc{
				traefikv1alpha1.GroupName: {"TraefikService": func(name, namespace string) (string, *dynamic.Service, error) {
					// func should never be executed in case of cross-provider reference.
					return "", nil, errors.New("BOOM")
				}},
			},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "service@file",
										Weight: func(i int) *int { return &i }(1),
									},
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
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

			p := Provider{EntryPoints: test.entryPoints}
			for group, kindFuncs := range test.groupKindBackendFuncs {
				for kind, backendFunc := range kindFuncs {
					p.RegisterBackendFuncs(group, kind, backendFunc)
				}
			}
			conf := p.loadConfigurationFromGateway(context.Background(), newClientMock(test.paths...))
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadHTTPRoutes_filterExtensionRef(t *testing.T) {
	testCases := []struct {
		desc                 string
		groupKindFilterFuncs map[string]map[string]BuildFilterFunc
		expected             *dynamic.Configuration
		entryPoints          map[string]Entrypoint
	}{
		{
			desc: "HTTPRoute with ExtensionRef filter",
			groupKindFilterFuncs: map[string]map[string]BuildFilterFunc{
				traefikv1alpha1.GroupName: {"Middleware": func(name, namespace string) (string, *dynamic.Middleware, error) {
					return namespace + "-" + name, nil, nil
				}},
			},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
							Middlewares: []string{"default-my-middleware"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
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
			desc: "HTTPRoute with ExtensionRef filter and create middleware",
			groupKindFilterFuncs: map[string]map[string]BuildFilterFunc{
				traefikv1alpha1.GroupName: {"Middleware": func(name, namespace string) (string, *dynamic.Middleware, error) {
					return namespace + "-" + name, &dynamic.Middleware{Headers: &dynamic.Headers{CustomRequestHeaders: map[string]string{"Test-Header": "Test"}}}, nil
				}},
			},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							RuleSyntax:  "v3",
							Middlewares: []string{"default-my-middleware"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-my-middleware": {Headers: &dynamic.Headers{CustomRequestHeaders: map[string]string{"Test-Header": "Test"}}},
					},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
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
			desc: "ExtensionRef filter: Unknown",
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
			desc: "ExtensionRef filter with filterFunc error",
			groupKindFilterFuncs: map[string]map[string]BuildFilterFunc{
				traefikv1alpha1.GroupName: {"Middleware": func(name, namespace string) (string, *dynamic.Middleware, error) {
					return "", nil, errors.New("BOOM")
				}},
			},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
			}},
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.expected == nil {
				return
			}

			p := Provider{EntryPoints: test.entryPoints}
			for group, kindFuncs := range test.groupKindFilterFuncs {
				for kind, filterFunc := range kindFuncs {
					p.RegisterFilterFuncs(group, kind, filterFunc)
				}
			}
			conf := p.loadConfigurationFromGateway(context.Background(), newClientMock([]string{"services.yml", "httproute/filter_extension_ref.yml"}...))
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadTCPRoutes(t *testing.T) {
	testCases := []struct {
		desc         string
		ingressClass string
		paths        []string
		expected     *dynamic.Configuration
		entryPoints  map[string]Entrypoint
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
			desc:  "Empty because missing entry point",
			paths: []string{"services.yml", "tcproute/simple.yml"},
			entryPoints: map[string]Entrypoint{"TCP": {
				Address: ":8000",
			}},
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
			desc:  "Empty because no tcp route defined",
			paths: []string{"services.yml", "tcproute/without_tcproute.yml"},
			entryPoints: map[string]Entrypoint{"TCP": {
				Address: ":8080",
			}},
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
			desc: "Empty caused by missing GatewayClass",
			entryPoints: map[string]Entrypoint{"TCP": {
				Address: ":8080",
			}},
			paths: []string{"services.yml", "tcproute/without_gatewayclass.yml"},
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
			desc: "Empty caused by unknown GatewayClass controller desc",
			entryPoints: map[string]Entrypoint{"TCP": {
				Address: ":8080",
			}},
			paths: []string{"services.yml", "tcproute/gatewayclass_with_unknown_controller.yml"},
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
			desc: "Empty caused by HTTPRoute with TLS configuration",
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":8080",
			}},
			paths: []string{"services.yml", "tcproute/with_tls_configuration.yml"},
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
			desc: "Empty caused by multi ports service with wrong TargetPort",
			entryPoints: map[string]Entrypoint{"TCP": {
				Address: ":8080",
			}},
			paths: []string{"services.yml", "tcproute/with_wrong_service_port.yml"},
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
			desc:  "Simple TCPRoute",
			paths: []string{"services.yml", "tcproute/simple.yml"},
			entryPoints: map[string]Entrypoint{
				"tcp": {Address: ":9000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-1-my-tcp-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "default-tcp-app-1-my-tcp-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-1-my-tcp-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
			desc:  "Multiple TCPRoute",
			paths: []string{"services.yml", "tcproute/with_multiple_routes.yml"},
			entryPoints: map[string]Entrypoint{
				"tcp-1":   {Address: ":9000"},
				"tcp-2":   {Address: ":10000"},
				"not-tcp": {Address: ":11000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-1-my-tcp-gateway-tcp-1-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp-1"},
							Service:     "default-tcp-app-1-my-tcp-gateway-tcp-1-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
						"default-tcp-app-2-my-tcp-gateway-tcp-2-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp-2"},
							Service:     "default-tcp-app-2-my-tcp-gateway-tcp-2-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-1-my-tcp-gateway-tcp-1-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-tcp-app-2-my-tcp-gateway-tcp-2-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-10000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
									},
								},
							},
						},
						"default-whoamitcp-10000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:10000",
									},
									{
										Address: "10.10.0.10:10000",
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
			desc:  "TCPRoute with multiple rules",
			paths: []string{"services.yml", "tcproute/with_multiple_rules.yml"},
			entryPoints: map[string]Entrypoint{
				"tcp-1":   {Address: ":9000"},
				"tcp-2":   {Address: ":10000"},
				"not-tcp": {Address: ":11000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-my-tcp-gateway-tcp-1-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp-1"},
							Service:     "default-tcp-app-my-tcp-gateway-tcp-1-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-my-tcp-gateway-tcp-1-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-tcp-app-my-tcp-gateway-tcp-1-e3b0c44298fc1c149afb-wrr-0",
										Weight: func(i int) *int { return &i }(1),
									},
									{
										Name:   "default-tcp-app-my-tcp-gateway-tcp-1-e3b0c44298fc1c149afb-wrr-1",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-tcp-app-my-tcp-gateway-tcp-1-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-tcp-app-my-tcp-gateway-tcp-1-e3b0c44298fc1c149afb-wrr-1": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-10000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
									},
								},
							},
						},
						"default-whoamitcp-10000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:10000",
									},
									{
										Address: "10.10.0.10:10000",
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
			desc:  "Simple TCPRoute, with backendRef",
			paths: []string{"services.yml", "tcproute/simple_cross_provider.yml"},
			entryPoints: map[string]Entrypoint{
				"tcp": {Address: ":9000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-1-my-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "default-tcp-app-1-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-1-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "service@file",
										Weight: func(i int) *int { return &i }(1),
									},
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
			desc:  "Simple TCPRoute, with TLS",
			paths: []string{"services.yml", "tcproute/with_protocol_tls.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-1-my-gateway-tls-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "default-tcp-app-1-my-gateway-tls-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-1-my-gateway-tls-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{{
									Name:   "default-whoamitcp-9000",
									Weight: func(i int) *int { return &i }(1),
								}},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
			desc:  "TCPRoute with Same namespace selector",
			paths: []string{"services.yml", "tcproute/with_namespace_same.yml"},
			entryPoints: map[string]Entrypoint{
				"tcp": {Address: ":9000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-default-my-tcp-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "default-tcp-app-default-my-tcp-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-default-my-tcp-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
			desc:  "TCPRoute with All namespace selector",
			paths: []string{"services.yml", "tcproute/with_namespace_all.yml"},
			entryPoints: map[string]Entrypoint{
				"tcp": {Address: ":9000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-default-my-tcp-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "default-tcp-app-default-my-tcp-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
						"bar-tcp-app-bar-my-tcp-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "bar-tcp-app-bar-my-tcp-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-default-my-tcp-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"bar-tcp-app-bar-my-tcp-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
									},
								},
							},
						},
						"bar-whoamitcp-bar-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.13:9000",
									},
									{
										Address: "10.10.0.14:9000",
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
			desc:  "TCPRoute with namespace selector",
			paths: []string{"services.yml", "tcproute/with_namespace_selector.yml"},
			entryPoints: map[string]Entrypoint{
				"tcp": {Address: ":9000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"bar-tcp-app-bar-my-tcp-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "bar-tcp-app-bar-my-tcp-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"bar-tcp-app-bar-my-tcp-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"bar-whoamitcp-bar-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.13:9000",
									},
									{
										Address: "10.10.0.14:9000",
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.expected == nil {
				return
			}

			p := Provider{EntryPoints: test.entryPoints, ExperimentalChannel: true}
			conf := p.loadConfigurationFromGateway(context.Background(), newClientMock(test.paths...))
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadTLSRoutes(t *testing.T) {
	testCases := []struct {
		desc         string
		ingressClass string
		paths        []string
		expected     *dynamic.Configuration
		entryPoints  map[string]Entrypoint
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
			desc:  "Empty because no matching entry point",
			paths: []string{"services.yml", "tlsroute/simple_TLS_to_TCPRoute.yml"},
			entryPoints: map[string]Entrypoint{"TCP": {
				Address: ":8000",
			}},
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
			desc:  "Empty because no tls route defined",
			paths: []string{"services.yml", "tlsroute/without_tlsroute.yml"},
			entryPoints: map[string]Entrypoint{"TCP": {
				Address: ":8080",
			}},
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
			desc: "Empty caused by missing GatewayClass",
			entryPoints: map[string]Entrypoint{"TCP": {
				Address: ":8080",
			}},
			paths: []string{"services.yml", "tlsroute/without_gatewayclass.yml"},
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
			desc: "Empty caused by unknown GatewayClass controller desc",
			entryPoints: map[string]Entrypoint{"TCP": {
				Address: ":8080",
			}},
			paths: []string{"services.yml", "tlsroute/gatewayclass_with_unknown_controller.yml"},
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
			desc: "Empty caused by multi ports service with wrong TargetPort",
			entryPoints: map[string]Entrypoint{"TCP": {
				Address: ":8080",
			}},
			paths: []string{"services.yml", "tlsroute/with_wrong_service_port.yml"},
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
			desc: "Empty caused by mixed routes with wrong parent ref",
			entryPoints: map[string]Entrypoint{
				"tcp": {
					Address: ":9000",
				},
				"tcp-tls": {
					Address: ":9443",
				},
				"http": {
					Address: ":80",
				},
			},
			paths: []string{"services.yml", "mixed/with_wrong_routes_selector.yml"},
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
			desc:  "Empty caused by simple TLSRoute with invalid SNI matching",
			paths: []string{"services.yml", "tlsroute/with_invalid_SNI_matching.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9001"},
			},
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
			desc:  "Simple TLS listener to TCPRoute in Terminate mode",
			paths: []string{"services.yml", "tlsroute/simple_TLS_to_TCPRoute.yml"},
			entryPoints: map[string]Entrypoint{
				"tcp": {Address: ":9000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-1-my-tls-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "default-tcp-app-1-my-tls-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-1-my-tls-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
			desc:  "Simple TLS listener to TCPRoute in Passthrough mode",
			paths: []string{"services.yml", "tlsroute/simple_TLS_to_TCPRoute_passthrough.yml"},
			entryPoints: map[string]Entrypoint{
				"tcp": {Address: ":9000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-1-my-tls-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "default-tcp-app-1-my-tls-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-1-my-tls-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
			desc:  "Simple TLS listener to TLSRoute",
			paths: []string{"services.yml", "tlsroute/simple_TLS_to_TLSRoute.yml"},
			entryPoints: map[string]Entrypoint{
				"tcp": {Address: ":9000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tls-app-1-my-tls-gateway-tcp-f0dd0dd89f82eae1c270": {
							EntryPoints: []string{"tcp"},
							Service:     "default-tls-app-1-my-tls-gateway-tcp-f0dd0dd89f82eae1c270-wrr-0",
							Rule:        "HostSNI(`foo.example.com`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tls-app-1-my-tls-gateway-tcp-f0dd0dd89f82eae1c270-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
			desc:  "Multiple TLSRoute",
			paths: []string{"services.yml", "tlsroute/with_multiple_routes_kind.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9000"},
				"tcp": {Address: ":10000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-1-my-tls-gateway-tls-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "default-tcp-app-1-my-tls-gateway-tls-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
						"default-tls-app-1-my-tls-gateway-tcp-673acf455cb2dab0b43a": {
							EntryPoints: []string{"tcp"},
							Service:     "default-tls-app-1-my-tls-gateway-tcp-673acf455cb2dab0b43a-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-1-my-tls-gateway-tls-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-tls-app-1-my-tls-gateway-tcp-673acf455cb2dab0b43a-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-10000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
									},
								},
							},
						},
						"default-whoamitcp-10000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:10000",
									},
									{
										Address: "10.10.0.10:10000",
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
			desc:  "Simple TLSRoute, with backendRef",
			paths: []string{"services.yml", "tlsroute/simple_cross_provider.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-1-my-gateway-tls-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "default-tcp-app-1-my-gateway-tls-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-1-my-gateway-tls-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "service@file",
										Weight: func(i int) *int { return &i }(1),
									},
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
			desc:  "Simple TLSRoute, with Passthrough and TLS configuration should raise a warn",
			paths: []string{"services.yml", "tlsroute/with_passthrough_tls.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9001"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tls-app-1-my-gateway-tls-f0dd0dd89f82eae1c270": {
							EntryPoints: []string{"tls"},
							Service:     "default-tls-app-1-my-gateway-tls-f0dd0dd89f82eae1c270-wrr-0",
							Rule:        "HostSNI(`foo.example.com`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tls-app-1-my-gateway-tls-f0dd0dd89f82eae1c270-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
			desc:  "Simple TLSRoute, with Passthrough",
			paths: []string{"services.yml", "tlsroute/with_passthrough.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9001"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tls-app-1-my-gateway-tls-f0dd0dd89f82eae1c270": {
							EntryPoints: []string{"tls"},
							Service:     "default-tls-app-1-my-gateway-tls-f0dd0dd89f82eae1c270-wrr-0",
							Rule:        "HostSNI(`foo.example.com`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tls-app-1-my-gateway-tls-f0dd0dd89f82eae1c270-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
			desc:  "Simple TLSRoute, with single SNI matching",
			paths: []string{"services.yml", "tlsroute/with_SNI_matching.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9001"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tls-app-1-my-gateway-tls-f0dd0dd89f82eae1c270": {
							EntryPoints: []string{"tls"},
							Service:     "default-tls-app-1-my-gateway-tls-f0dd0dd89f82eae1c270-wrr-0",
							Rule:        "HostSNI(`foo.example.com`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tls-app-1-my-gateway-tls-f0dd0dd89f82eae1c270-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
			desc:  "Simple TLSRoute, with multiple SNI matching",
			paths: []string{"services.yml", "tlsroute/with_multiple_SNI_matching.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9001"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tls-app-1-my-gateway-tls-d5342d75658583f03593": {
							EntryPoints: []string{"tls"},
							Service:     "default-tls-app-1-my-gateway-tls-d5342d75658583f03593-wrr-0",
							Rule:        "HostSNI(`foo.example.com`) || HostSNI(`bar.example.com`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tls-app-1-my-gateway-tls-d5342d75658583f03593-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
			desc:  "TLSRoute with Same namespace selector",
			paths: []string{"services.yml", "tlsroute/with_namespace_same.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9001"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tls-app-default-my-gateway-tls-06ae57dcf13ab4c60ee5": {
							EntryPoints: []string{"tls"},
							Service:     "default-tls-app-default-my-gateway-tls-06ae57dcf13ab4c60ee5-wrr-0",
							Rule:        "HostSNI(`foo.default`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tls-app-default-my-gateway-tls-06ae57dcf13ab4c60ee5-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
			desc:  "TLSRoute with All namespace selector",
			paths: []string{"services.yml", "tlsroute/with_namespace_all.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9001"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tls-app-default-my-gateway-tls-06ae57dcf13ab4c60ee5": {
							EntryPoints: []string{"tls"},
							Service:     "default-tls-app-default-my-gateway-tls-06ae57dcf13ab4c60ee5-wrr-0",
							Rule:        "HostSNI(`foo.default`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
						"bar-tls-app-bar-my-gateway-tls-2279fe75c5156dc5eb26": {
							EntryPoints: []string{"tls"},
							Service:     "bar-tls-app-bar-my-gateway-tls-2279fe75c5156dc5eb26-wrr-0",
							Rule:        "HostSNI(`foo.bar`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tls-app-default-my-gateway-tls-06ae57dcf13ab4c60ee5-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"bar-tls-app-bar-my-gateway-tls-2279fe75c5156dc5eb26-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
									},
								},
							},
						},
						"bar-whoamitcp-bar-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.13:9000",
									},
									{
										Address: "10.10.0.14:9000",
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
			desc:  "TLSRoute with namespace selector",
			paths: []string{"services.yml", "tlsroute/with_namespace_selector.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9001"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"bar-tls-app-bar-my-gateway-tls-2279fe75c5156dc5eb26": {
							EntryPoints: []string{"tls"},
							Service:     "bar-tls-app-bar-my-gateway-tls-2279fe75c5156dc5eb26-wrr-0",
							Rule:        "HostSNI(`foo.bar`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"bar-tls-app-bar-my-gateway-tls-2279fe75c5156dc5eb26-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"bar-whoamitcp-bar-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.13:9000",
									},
									{
										Address: "10.10.0.14:9000",
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
			desc:  "TLSRoute with multiple rules",
			paths: []string{"services.yml", "tlsroute/with_multiple_rules.yml"},
			entryPoints: map[string]Entrypoint{
				"tcp-1": {Address: ":9000"},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tls-app-my-gateway-tcp-1-673acf455cb2dab0b43a": {
							EntryPoints: []string{"tcp-1"},
							Service:     "default-tls-app-my-gateway-tcp-1-673acf455cb2dab0b43a-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tls-app-my-gateway-tcp-1-673acf455cb2dab0b43a-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-tls-app-my-gateway-tcp-1-673acf455cb2dab0b43a-wrr-0",
										Weight: func(i int) *int { return &i }(1),
									},
									{
										Name:   "default-tls-app-my-gateway-tcp-1-673acf455cb2dab0b43a-wrr-1",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-tls-app-my-gateway-tcp-1-673acf455cb2dab0b43a-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-tls-app-my-gateway-tcp-1-673acf455cb2dab0b43a-wrr-1": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-10000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
									},
								},
							},
						},
						"default-whoamitcp-10000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:10000",
									},
									{
										Address: "10.10.0.10:10000",
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.expected == nil {
				return
			}

			p := Provider{EntryPoints: test.entryPoints, ExperimentalChannel: true}
			conf := p.loadConfigurationFromGateway(context.Background(), newClientMock(test.paths...))
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadMixedRoutes(t *testing.T) {
	testCases := []struct {
		desc                string
		ingressClass        string
		paths               []string
		expected            *dynamic.Configuration
		entryPoints         map[string]Entrypoint
		experimentalChannel bool
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
			desc:  "Empty caused by unsupported listener.Protocol",
			paths: []string{"services.yml", "mixed/with_bad_listener_protocol.yml"},
			entryPoints: map[string]Entrypoint{
				"web": {Address: ":9080"},
			},
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
			desc: "Empty caused by unsupported listener.Route.Kind",
			entryPoints: map[string]Entrypoint{
				"web": {Address: ":9080"},
			},
			paths: []string{"services.yml", "mixed/with_bad_listener_route_kind.yml"},
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
			desc:  "Empty caused by listener.Protocol does not support listener.Route.Kind",
			paths: []string{"services.yml", "mixed/with_incompatible_protocol_and_route_kind.yml"},
			entryPoints: map[string]Entrypoint{
				"web": {Address: ":9080"},
			},
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
			desc:  "Mixed routes",
			paths: []string{"services.yml", "mixed/simple.yml"},
			entryPoints: map[string]Entrypoint{
				"web":       {Address: ":9080"},
				"websecure": {Address: ":9443"},
				"tcp":       {Address: ":9000"},
				"tls-1":     {Address: ":10000"},
				"tls-2":     {Address: ":11000"},
			},
			experimentalChannel: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-1-my-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "default-tcp-app-1-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
						"default-tcp-app-1-my-gateway-tls-1-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-1"},
							Service:     "default-tcp-app-1-my-gateway-tls-1-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
						"default-tls-app-1-my-gateway-tls-2-59130f7db6718b7700c1": {
							EntryPoints: []string{"tls-2"},
							Service:     "default-tls-app-1-my-gateway-tls-2-59130f7db6718b7700c1-wrr-0",
							Rule:        "HostSNI(`pass.tls.foo.example.com`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-1-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-tcp-app-1-my-gateway-tls-1-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-tls-app-1-my-gateway-tls-2-59130f7db6718b7700c1-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-http-app-1-my-gateway-web-a431b128267aabc954fd": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-1-my-gateway-web-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							RuleSyntax:  "v3",
						},
						"default-http-app-1-my-gateway-websecure-a431b128267aabc954fd": {
							EntryPoints: []string{"websecure"},
							Service:     "default-http-app-1-my-gateway-websecure-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-1-my-gateway-web-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-http-app-1-my-gateway-websecure-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
								CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
			},
		},
		{
			desc:  "Empty caused by mixed routes with multiple listeners using same hostname, port and protocol",
			paths: []string{"services.yml", "mixed/with_multiple_listeners_using_same_hostname_port_protocol.yml"},
			entryPoints: map[string]Entrypoint{
				"web": {Address: ":9080"},
				"tcp": {Address: ":9000"},
				"tls": {Address: ":9001"},
			},
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
			desc:  "Mixed routes with Same namespace selector",
			paths: []string{"services.yml", "mixed/with_namespace_same.yml"},
			entryPoints: map[string]Entrypoint{
				"web":       {Address: ":9080"},
				"websecure": {Address: ":9443"},
				"tcp":       {Address: ":9000"},
				"tls-1":     {Address: ":10000"},
				"tls-2":     {Address: ":11000"},
			},
			experimentalChannel: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-default-my-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "default-tcp-app-default-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
						"default-tcp-app-default-my-gateway-tls-1-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-1"},
							Service:     "default-tcp-app-default-my-gateway-tls-1-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
						"default-tls-app-default-my-gateway-tls-2-59130f7db6718b7700c1": {
							EntryPoints: []string{"tls-2"},
							Service:     "default-tls-app-default-my-gateway-tls-2-59130f7db6718b7700c1-wrr-0",
							Rule:        "HostSNI(`pass.tls.foo.example.com`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-default-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-tcp-app-default-my-gateway-tls-1-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-tls-app-default-my-gateway-tls-2-59130f7db6718b7700c1-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-http-app-default-my-gateway-web-a431b128267aabc954fd": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-default-my-gateway-web-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							RuleSyntax:  "v3",
						},
						"default-http-app-default-my-gateway-websecure-a431b128267aabc954fd": {
							EntryPoints: []string{"websecure"},
							Service:     "default-http-app-default-my-gateway-websecure-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-default-my-gateway-web-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-http-app-default-my-gateway-websecure-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
								CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
			},
		},
		{
			desc:  "Mixed routes with All namespace selector",
			paths: []string{"services.yml", "mixed/with_namespace_all.yml"},
			entryPoints: map[string]Entrypoint{
				"web":       {Address: ":9080"},
				"websecure": {Address: ":9443"},
				"tcp":       {Address: ":9000"},
				"tls-1":     {Address: ":10000"},
				"tls-2":     {Address: ":11000"},
			},
			experimentalChannel: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-default-my-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "default-tcp-app-default-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
						"default-tcp-app-default-my-gateway-tls-1-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-1"},
							Service:     "default-tcp-app-default-my-gateway-tls-1-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
						"default-tls-app-default-my-gateway-tls-2-59130f7db6718b7700c1": {
							EntryPoints: []string{"tls-2"},
							Service:     "default-tls-app-default-my-gateway-tls-2-59130f7db6718b7700c1-wrr-0",
							Rule:        "HostSNI(`pass.tls.foo.example.com`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
						"bar-tcp-app-bar-my-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "bar-tcp-app-bar-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
						"bar-tcp-app-bar-my-gateway-tls-1-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-1"},
							Service:     "bar-tcp-app-bar-my-gateway-tls-1-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-default-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-tcp-app-default-my-gateway-tls-1-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-tls-app-default-my-gateway-tls-2-59130f7db6718b7700c1-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
									},
								},
							},
						},
						"bar-whoamitcp-bar-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.13:9000",
									},
									{
										Address: "10.10.0.14:9000",
									},
								},
							},
						},
						"bar-tcp-app-bar-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"bar-tcp-app-bar-my-gateway-tls-1-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-http-app-default-my-gateway-web-a431b128267aabc954fd": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-default-my-gateway-web-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							RuleSyntax:  "v3",
						},
						"default-http-app-default-my-gateway-websecure-a431b128267aabc954fd": {
							EntryPoints: []string{"websecure"},
							Service:     "default-http-app-default-my-gateway-websecure-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"bar-http-app-bar-my-gateway-web-a431b128267aabc954fd": {
							EntryPoints: []string{"web"},
							Service:     "bar-http-app-bar-my-gateway-web-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							RuleSyntax:  "v3",
						},
						"bar-http-app-bar-my-gateway-websecure-a431b128267aabc954fd": {
							EntryPoints: []string{"websecure"},
							Service:     "bar-http-app-bar-my-gateway-websecure-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-default-my-gateway-web-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-http-app-default-my-gateway-websecure-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"bar-whoami-bar-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.11:80",
									},
									{
										URL: "http://10.10.0.12:80",
									},
								},
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"bar-http-app-bar-my-gateway-web-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "bar-whoami-bar-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"bar-http-app-bar-my-gateway-websecure-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "bar-whoami-bar-80",
										Weight: func(i int) *int { return &i }(1),
									},
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
								CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
			},
		},
		{
			desc:  "Mixed routes with Selector Route Binding",
			paths: []string{"services.yml", "mixed/with_namespace_selector.yml"},
			entryPoints: map[string]Entrypoint{
				"web":       {Address: ":9080"},
				"websecure": {Address: ":9443"},
				"tcp":       {Address: ":9000"},
				"tls-1":     {Address: ":10000"},
				"tls-2":     {Address: ":11000"},
			},
			experimentalChannel: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"bar-tcp-app-bar-my-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "bar-tcp-app-bar-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
						"bar-tcp-app-bar-my-gateway-tls-1-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-1"},
							Service:     "bar-tcp-app-bar-my-gateway-tls-1-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
						"bar-tls-app-bar-my-gateway-tls-2-59130f7db6718b7700c1": {
							EntryPoints: []string{"tls-2"},
							Service:     "bar-tls-app-bar-my-gateway-tls-2-59130f7db6718b7700c1-wrr-0",
							Rule:        "HostSNI(`pass.tls.foo.example.com`)",
							RuleSyntax:  "v3",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"bar-whoamitcp-bar-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.13:9000",
									},
									{
										Address: "10.10.0.14:9000",
									},
								},
							},
						},
						"bar-tcp-app-bar-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"bar-tcp-app-bar-my-gateway-tls-1-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"bar-tls-app-bar-my-gateway-tls-2-59130f7db6718b7700c1-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"bar-http-app-bar-my-gateway-web-a431b128267aabc954fd": {
							EntryPoints: []string{"web"},
							Service:     "bar-http-app-bar-my-gateway-web-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							RuleSyntax:  "v3",
						},
						"bar-http-app-bar-my-gateway-websecure-a431b128267aabc954fd": {
							EntryPoints: []string{"websecure"},
							Service:     "bar-http-app-bar-my-gateway-websecure-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"bar-whoami-bar-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.11:80",
									},
									{
										URL: "http://10.10.0.12:80",
									},
								},
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"bar-http-app-bar-my-gateway-web-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "bar-whoami-bar-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"bar-http-app-bar-my-gateway-websecure-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "bar-whoami-bar-80",
										Weight: func(i int) *int { return &i }(1),
									},
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
								CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
							},
						},
					},
				},
			},
		},
		{
			desc:  "Mixed routes with core group",
			paths: []string{"services.yml", "mixed/with_core_group.yml"},
			entryPoints: map[string]Entrypoint{
				"web":       {Address: ":9080"},
				"websecure": {Address: ":9443"},
				"tcp":       {Address: ":9000"},
				"tls":       {Address: ":10000"},
			},
			experimentalChannel: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-default-my-gateway-tcp-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "default-tcp-app-default-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
						},
						"default-tcp-app-default-my-gateway-tls-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "default-tcp-app-default-my-gateway-tls-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-default-my-gateway-tcp-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-tcp-app-default-my-gateway-tls-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"default-http-app-default-my-gateway-web-a431b128267aabc954fd": {
							EntryPoints: []string{"web"},
							Service:     "default-http-app-default-my-gateway-web-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							RuleSyntax:  "v3",
						},
						"default-http-app-default-my-gateway-websecure-a431b128267aabc954fd": {
							EntryPoints: []string{"websecure"},
							Service:     "default-http-app-default-my-gateway-websecure-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"default-http-app-default-my-gateway-web-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-http-app-default-my-gateway-websecure-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
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
								CertFile: types.FileOrContent("-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----"),
								KeyFile:  types.FileOrContent("-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"),
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

			if test.expected == nil {
				return
			}

			p := Provider{EntryPoints: test.entryPoints, ExperimentalChannel: test.experimentalChannel}
			conf := p.loadConfigurationFromGateway(context.Background(), newClientMock(test.paths...))
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadRoutesWithReferenceGrants(t *testing.T) {
	testCases := []struct {
		desc                string
		ingressClass        string
		paths               []string
		expected            *dynamic.Configuration
		entryPoints         map[string]Entrypoint
		experimentalChannel bool
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
			desc:  "Empty because ReferenceGrant for Secret is missing",
			paths: []string{"services.yml", "referencegrant/for_secret_missing.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9000"},
			},
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
			desc:  "Empty because ReferenceGrant spec.from does not match",
			paths: []string{"services.yml", "referencegrant/for_secret_not_matching_from.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9000"},
			},
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
			desc:  "Empty because ReferenceGrant spec.to does not match",
			paths: []string{"services.yml", "referencegrant/for_secret_not_matching_to.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9000"},
			},
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
			desc:  "For Secret",
			paths: []string{"services.yml", "referencegrant/for_secret.yml"},
			entryPoints: map[string]Entrypoint{
				"tls": {Address: ":9000"},
			},
			experimentalChannel: true,
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"default-tcp-app-1-my-gateway-tls-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "default-tcp-app-1-my-gateway-tls-e3b0c44298fc1c149afb-wrr-0",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "v3",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"default-tcp-app-1-my-gateway-tls-e3b0c44298fc1c149afb-wrr-0": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{{
									Name:   "default-whoamitcp-9000",
									Weight: func(i int) *int { return &i }(1),
								}},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.9:9000",
									},
									{
										Address: "10.10.0.10:9000",
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.expected == nil {
				return
			}

			p := Provider{EntryPoints: test.entryPoints, ExperimentalChannel: test.experimentalChannel}
			conf := p.loadConfigurationFromGateway(context.Background(), newClientMock(test.paths...))
			assert.Equal(t, test.expected, conf)
		})
	}
}

func Test_hostRule(t *testing.T) {
	testCases := []struct {
		desc         string
		hostnames    []gatev1.Hostname
		expectedRule string
		expectErr    bool
	}{
		{
			desc:         "Empty rule and matches",
			expectedRule: "",
		},
		{
			desc: "One Host",
			hostnames: []gatev1.Hostname{
				"Foo",
			},
			expectedRule: "Host(`Foo`)",
		},
		{
			desc: "Multiple Hosts",
			hostnames: []gatev1.Hostname{
				"Foo",
				"Bar",
				"Bir",
			},
			expectedRule: "(Host(`Foo`) || Host(`Bar`) || Host(`Bir`))",
		},
		{
			desc: "Multiple Hosts with empty one",
			hostnames: []gatev1.Hostname{
				"Foo",
				"",
				"Bir",
			},
			expectedRule: "",
		},
		{
			desc: "Multiple empty hosts",
			hostnames: []gatev1.Hostname{
				"",
				"",
				"",
			},
			expectedRule: "",
		},
		{
			desc: "Several Host and wildcard",
			hostnames: []gatev1.Hostname{
				"*.bar.foo",
				"bar.foo",
				"foo.foo",
			},
			expectedRule: "(HostRegexp(`^[a-zA-Z0-9-]+\\.bar\\.foo$`) || Host(`bar.foo`) || Host(`foo.foo`))",
		},
		{
			desc: "Host with wildcard",
			hostnames: []gatev1.Hostname{
				"*.bar.foo",
			},
			expectedRule: "HostRegexp(`^[a-zA-Z0-9-]+\\.bar\\.foo$`)",
		},
		{
			desc: "Alone wildcard",
			hostnames: []gatev1.Hostname{
				"*",
				"*.foo.foo",
			},
		},
		{
			desc: "Multiple alone Wildcard",
			hostnames: []gatev1.Hostname{
				"foo.foo",
				"*.*",
			},
			expectErr: true,
		},
		{
			desc: "Multiple Wildcard",
			hostnames: []gatev1.Hostname{
				"foo.foo",
				"*.toto.*.bar.foo",
			},
			expectErr: true,
		},
		{
			desc: "Multiple subdomain with misplaced wildcard",
			hostnames: []gatev1.Hostname{
				"foo.foo",
				"toto.*.bar.foo",
			},
			expectErr: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			rule, err := hostRule(test.hostnames)

			assert.Equal(t, test.expectedRule, rule)
			if test.expectErr {
				assert.Error(t, err)
			}
		})
	}
}

func Test_extractRule(t *testing.T) {
	testCases := []struct {
		desc          string
		routeRule     gatev1.HTTPRouteRule
		hostRule      string
		expectedRule  string
		expectedError bool
	}{
		{
			desc:         "Empty rule and matches",
			expectedRule: "PathPrefix(`/`)",
		},
		{
			desc:         "One Host rule without matches",
			hostRule:     "Host(`foo.com`)",
			expectedRule: "Host(`foo.com`) && PathPrefix(`/`)",
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPHeaderMatch",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{Headers: nil},
				},
			},
			expectedRule: "",
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPHeaderMatch Type",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Headers: []gatev1.HTTPHeaderMatch{
							{Type: nil, Name: "foo", Value: "bar"},
						},
					},
				},
			},
			expectedRule: "",
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPPathMatch",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{Path: nil},
				},
			},
			expectedRule: "",
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPPathMatch Type",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  nil,
							Value: ptr.To("/foo/"),
						},
					},
				},
			},
			expectedRule: "",
		},
		{
			desc: "One HTTPRouteMatch with nil HTTPPathMatch Values",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  pathMatchTypePtr(gatev1.PathMatchExact),
							Value: nil,
						},
					},
				},
			},
			expectedRule: "",
		},
		{
			desc: "One Path in matches",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  pathMatchTypePtr(gatev1.PathMatchExact),
							Value: ptr.To("/foo/"),
						},
					},
				},
			},
			expectedRule: "Path(`/foo/`)",
		},
		{
			desc: "One Path in matches and another unknown",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  pathMatchTypePtr(gatev1.PathMatchExact),
							Value: ptr.To("/foo/"),
						},
					},
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  pathMatchTypePtr("unknown"),
							Value: ptr.To("/foo/"),
						},
					},
				},
			},
			expectedError: true,
		},
		{
			desc: "One Path in matches and another empty",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  pathMatchTypePtr(gatev1.PathMatchExact),
							Value: ptr.To("/foo/"),
						},
					},
					{},
				},
			},
			expectedRule: "Path(`/foo/`)",
		},
		{
			desc: "Path OR Header rules",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  pathMatchTypePtr(gatev1.PathMatchExact),
							Value: ptr.To("/foo/"),
						},
					},
					{
						Headers: []gatev1.HTTPHeaderMatch{
							{
								Type:  headerMatchTypePtr(gatev1.HeaderMatchExact),
								Name:  "my-header",
								Value: "foo",
							},
						},
					},
				},
			},
			expectedRule: "Path(`/foo/`) || Header(`my-header`,`foo`)",
		},
		{
			desc: "Path && Header rules",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  pathMatchTypePtr(gatev1.PathMatchExact),
							Value: ptr.To("/foo/"),
						},
						Headers: []gatev1.HTTPHeaderMatch{
							{
								Type:  headerMatchTypePtr(gatev1.HeaderMatchExact),
								Name:  "my-header",
								Value: "foo",
							},
						},
					},
				},
			},
			expectedRule: "Path(`/foo/`) && Header(`my-header`,`foo`)",
		},
		{
			desc:     "Host && Path && Header rules",
			hostRule: "Host(`foo.com`)",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  pathMatchTypePtr(gatev1.PathMatchExact),
							Value: ptr.To("/foo/"),
						},
						Headers: []gatev1.HTTPHeaderMatch{
							{
								Type:  headerMatchTypePtr(gatev1.HeaderMatchExact),
								Name:  "my-header",
								Value: "foo",
							},
						},
					},
				},
			},
			expectedRule: "Host(`foo.com`) && Path(`/foo/`) && Header(`my-header`,`foo`)",
		},
		{
			desc:     "Host && (Path || Header) rules",
			hostRule: "Host(`foo.com`)",
			routeRule: gatev1.HTTPRouteRule{
				Matches: []gatev1.HTTPRouteMatch{
					{
						Path: &gatev1.HTTPPathMatch{
							Type:  pathMatchTypePtr(gatev1.PathMatchExact),
							Value: ptr.To("/foo/"),
						},
					},
					{
						Headers: []gatev1.HTTPHeaderMatch{
							{
								Type:  headerMatchTypePtr(gatev1.HeaderMatchExact),
								Name:  "my-header",
								Value: "foo",
							},
						},
					},
				},
			},
			expectedRule: "Host(`foo.com`) && (Path(`/foo/`) || Header(`my-header`,`foo`))",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rule, err := extractRule(test.routeRule, test.hostRule)
			if test.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectedRule, rule)
		})
	}
}

func Test_hostSNIRule(t *testing.T) {
	testCases := []struct {
		desc         string
		hostnames    []gatev1.Hostname
		expectedRule string
		expectError  bool
	}{
		{
			desc:         "Empty",
			expectedRule: "HostSNI(`*`)",
		},
		{
			desc:         "Empty hostname",
			hostnames:    []gatev1.Hostname{""},
			expectedRule: "HostSNI(`*`)",
		},
		{
			desc:        "Unsupported wildcard",
			hostnames:   []gatev1.Hostname{"*"},
			expectError: true,
		},
		{
			desc:         "Supported wildcard",
			hostnames:    []gatev1.Hostname{"*.foo"},
			expectedRule: "HostSNIRegexp(`^[a-zA-Z0-9-]+\\.foo$`)",
		},
		{
			desc:        "Multiple malformed wildcard",
			hostnames:   []gatev1.Hostname{"*.foo.*"},
			expectError: true,
		},
		{
			desc:         "Some empty hostnames",
			hostnames:    []gatev1.Hostname{"foo", "", "bar"},
			expectedRule: "HostSNI(`foo`) || HostSNI(`bar`)",
		},
		{
			desc:         "Valid hostname",
			hostnames:    []gatev1.Hostname{"foo"},
			expectedRule: "HostSNI(`foo`)",
		},
		{
			desc:         "Multiple valid hostnames",
			hostnames:    []gatev1.Hostname{"foo", "bar"},
			expectedRule: "HostSNI(`foo`) || HostSNI(`bar`)",
		},
		{
			desc:         "Multiple valid hostnames with wildcard",
			hostnames:    []gatev1.Hostname{"bar.foo", "foo.foo", "*.foo"},
			expectedRule: "HostSNI(`bar.foo`) || HostSNI(`foo.foo`) || HostSNIRegexp(`^[a-zA-Z0-9-]+\\.foo$`)",
		},
		{
			desc:         "Multiple overlapping hostnames",
			hostnames:    []gatev1.Hostname{"foo", "bar", "foo", "baz"},
			expectedRule: "HostSNI(`foo`) || HostSNI(`bar`) || HostSNI(`baz`)",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rule, err := hostSNIRule(test.hostnames)
			if test.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expectedRule, rule)
		})
	}
}

func Test_shouldAttach(t *testing.T) {
	testCases := []struct {
		desc           string
		gateway        *gatev1.Gateway
		listener       gatev1.Listener
		routeNamespace string
		routeSpec      gatev1.CommonRouteSpec
		expectedAttach bool
	}{
		{
			desc: "No ParentRefs",
			gateway: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gateway",
					Namespace: "default",
				},
			},
			listener: gatev1.Listener{
				Name: "foo",
			},
			routeNamespace: "default",
			routeSpec: gatev1.CommonRouteSpec{
				ParentRefs: nil,
			},
			expectedAttach: false,
		},
		{
			desc: "Unsupported Kind",
			gateway: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gateway",
					Namespace: "default",
				},
			},
			listener: gatev1.Listener{
				Name: "foo",
			},
			routeNamespace: "default",
			routeSpec: gatev1.CommonRouteSpec{
				ParentRefs: []gatev1.ParentReference{
					{
						SectionName: sectionNamePtr("bar"),
						Name:        "gateway",
						Namespace:   namespacePtr("default"),
						Kind:        kindPtr("Foo"),
						Group:       groupPtr(gatev1.GroupName),
					},
				},
			},
			expectedAttach: false,
		},
		{
			desc: "Unsupported Group",
			gateway: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gateway",
					Namespace: "default",
				},
			},
			listener: gatev1.Listener{
				Name: "foo",
			},
			routeNamespace: "default",
			routeSpec: gatev1.CommonRouteSpec{
				ParentRefs: []gatev1.ParentReference{
					{
						SectionName: sectionNamePtr("bar"),
						Name:        "gateway",
						Namespace:   namespacePtr("default"),
						Kind:        kindPtr("Gateway"),
						Group:       groupPtr("foo.com"),
					},
				},
			},
			expectedAttach: false,
		},
		{
			desc: "Kind is nil",
			gateway: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gateway",
					Namespace: "default",
				},
			},
			listener: gatev1.Listener{
				Name: "foo",
			},
			routeNamespace: "default",
			routeSpec: gatev1.CommonRouteSpec{
				ParentRefs: []gatev1.ParentReference{
					{
						SectionName: sectionNamePtr("bar"),
						Name:        "gateway",
						Namespace:   namespacePtr("default"),
						Group:       groupPtr(gatev1.GroupName),
					},
				},
			},
			expectedAttach: false,
		},
		{
			desc: "Group is nil",
			gateway: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gateway",
					Namespace: "default",
				},
			},
			listener: gatev1.Listener{
				Name: "foo",
			},
			routeNamespace: "default",
			routeSpec: gatev1.CommonRouteSpec{
				ParentRefs: []gatev1.ParentReference{
					{
						SectionName: sectionNamePtr("bar"),
						Name:        "gateway",
						Namespace:   namespacePtr("default"),
						Kind:        kindPtr("Gateway"),
					},
				},
			},
			expectedAttach: false,
		},
		{
			desc: "SectionName does not match a listener desc",
			gateway: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gateway",
					Namespace: "default",
				},
			},
			listener: gatev1.Listener{
				Name: "foo",
			},
			routeNamespace: "default",
			routeSpec: gatev1.CommonRouteSpec{
				ParentRefs: []gatev1.ParentReference{
					{
						SectionName: sectionNamePtr("bar"),
						Name:        "gateway",
						Namespace:   namespacePtr("default"),
						Group:       groupPtr(gatev1.GroupName),
						Kind:        kindPtr("Gateway"),
					},
				},
			},
			expectedAttach: false,
		},
		{
			desc: "Namespace does not match the Gateway namespace",
			gateway: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gateway",
					Namespace: "default",
				},
			},
			listener: gatev1.Listener{
				Name: "foo",
			},
			routeNamespace: "default",
			routeSpec: gatev1.CommonRouteSpec{
				ParentRefs: []gatev1.ParentReference{
					{
						SectionName: sectionNamePtr("bar"),
						Name:        "gateway",
						Namespace:   namespacePtr("bar"),
						Group:       groupPtr(gatev1.GroupName),
						Kind:        kindPtr("Gateway"),
					},
				},
			},
			expectedAttach: false,
		},
		{
			desc: "Route namespace does not match the Gateway namespace",
			gateway: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gateway",
					Namespace: "default",
				},
			},
			listener: gatev1.Listener{
				Name: "foo",
			},
			routeNamespace: "bar",
			routeSpec: gatev1.CommonRouteSpec{
				ParentRefs: []gatev1.ParentReference{
					{
						SectionName: sectionNamePtr("bar"),
						Name:        "gateway",
						Group:       groupPtr(gatev1.GroupName),
						Kind:        kindPtr("Gateway"),
					},
				},
			},
			expectedAttach: false,
		},
		{
			desc: "Unsupported Kind",
			gateway: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gateway",
					Namespace: "default",
				},
			},
			listener: gatev1.Listener{
				Name: "foo",
			},
			routeNamespace: "default",
			routeSpec: gatev1.CommonRouteSpec{
				ParentRefs: []gatev1.ParentReference{
					{
						SectionName: sectionNamePtr("bar"),
						Name:        "gateway",
						Namespace:   namespacePtr("default"),
						Kind:        kindPtr("Gateway"),
						Group:       groupPtr(gatev1.GroupName),
					},
				},
			},
			expectedAttach: false,
		},
		{
			desc: "Route namespace matches the Gateway namespace",
			gateway: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gateway",
					Namespace: "default",
				},
			},
			listener: gatev1.Listener{
				Name: "foo",
			},
			routeNamespace: "default",
			routeSpec: gatev1.CommonRouteSpec{
				ParentRefs: []gatev1.ParentReference{
					{
						SectionName: sectionNamePtr("foo"),
						Name:        "gateway",
						Kind:        kindPtr("Gateway"),
						Group:       groupPtr(gatev1.GroupName),
					},
				},
			},
			expectedAttach: true,
		},
		{
			desc: "Namespace matches the Gateway namespace",
			gateway: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gateway",
					Namespace: "default",
				},
			},
			listener: gatev1.Listener{
				Name: "foo",
			},
			routeNamespace: "bar",
			routeSpec: gatev1.CommonRouteSpec{
				ParentRefs: []gatev1.ParentReference{
					{
						SectionName: sectionNamePtr("foo"),
						Name:        "gateway",
						Namespace:   namespacePtr("default"),
						Kind:        kindPtr("Gateway"),
						Group:       groupPtr(gatev1.GroupName),
					},
				},
			},
			expectedAttach: true,
		},
		{
			desc: "Only one ParentRef matches the Gateway",
			gateway: &gatev1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gateway",
					Namespace: "default",
				},
			},
			listener: gatev1.Listener{
				Name: "foo",
			},
			routeNamespace: "bar",
			routeSpec: gatev1.CommonRouteSpec{
				ParentRefs: []gatev1.ParentReference{
					{
						Name:      "gateway2",
						Namespace: namespacePtr("default"),
						Kind:      kindPtr("Gateway"),
						Group:     groupPtr(gatev1.GroupName),
					},
					{
						Name:      "gateway",
						Namespace: namespacePtr("default"),
						Kind:      kindPtr("Gateway"),
						Group:     groupPtr(gatev1.GroupName),
					},
				},
			},
			expectedAttach: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := shouldAttach(test.gateway, test.listener, test.routeNamespace, test.routeSpec)
			assert.Equal(t, test.expectedAttach, got)
		})
	}
}

func Test_matchingHostnames(t *testing.T) {
	testCases := []struct {
		desc      string
		listener  gatev1.Listener
		hostnames []gatev1.Hostname
		want      []gatev1.Hostname
	}{
		{
			desc: "Empty",
		},
		{
			desc: "Only listener hostname",
			listener: gatev1.Listener{
				Hostname: hostnamePtr("foo.com"),
			},
			want: []gatev1.Hostname{"foo.com"},
		},
		{
			desc:      "Only Route hostname",
			hostnames: []gatev1.Hostname{"foo.com"},
			want:      []gatev1.Hostname{"foo.com"},
		},
		{
			desc: "Matching hostname",
			listener: gatev1.Listener{
				Hostname: hostnamePtr("foo.com"),
			},
			hostnames: []gatev1.Hostname{"foo.com"},
			want:      []gatev1.Hostname{"foo.com"},
		},
		{
			desc: "Matching hostname with wildcard",
			listener: gatev1.Listener{
				Hostname: hostnamePtr("*.foo.com"),
			},
			hostnames: []gatev1.Hostname{"*.foo.com"},
			want:      []gatev1.Hostname{"*.foo.com"},
		},
		{
			desc: "Matching subdomain with listener wildcard",
			listener: gatev1.Listener{
				Hostname: hostnamePtr("*.foo.com"),
			},
			hostnames: []gatev1.Hostname{"bar.foo.com"},
			want:      []gatev1.Hostname{"bar.foo.com"},
		},
		{
			desc: "Matching subdomain with route hostname wildcard",
			listener: gatev1.Listener{
				Hostname: hostnamePtr("bar.foo.com"),
			},
			hostnames: []gatev1.Hostname{"*.foo.com"},
			want:      []gatev1.Hostname{"bar.foo.com"},
		},
		{
			desc: "Non matching root domain with listener wildcard",
			listener: gatev1.Listener{
				Hostname: hostnamePtr("*.foo.com"),
			},
			hostnames: []gatev1.Hostname{"foo.com"},
		},
		{
			desc: "Non matching root domain with route hostname wildcard",
			listener: gatev1.Listener{
				Hostname: hostnamePtr("foo.com"),
			},
			hostnames: []gatev1.Hostname{"*.foo.com"},
		},
		{
			desc: "Multiple route hostnames with one matching route hostname",
			listener: gatev1.Listener{
				Hostname: hostnamePtr("*.foo.com"),
			},
			hostnames: []gatev1.Hostname{"bar.com", "test.foo.com", "test.buz.com"},
			want:      []gatev1.Hostname{"test.foo.com"},
		},
		{
			desc: "Multiple route hostnames with non matching route hostname",
			listener: gatev1.Listener{
				Hostname: hostnamePtr("*.fuz.com"),
			},
			hostnames: []gatev1.Hostname{"bar.com", "test.foo.com", "test.buz.com"},
		},
		{
			desc: "Multiple route hostnames with multiple matching route hostnames",
			listener: gatev1.Listener{
				Hostname: hostnamePtr("*.foo.com"),
			},
			hostnames: []gatev1.Hostname{"toto.foo.com", "test.foo.com", "test.buz.com"},
			want:      []gatev1.Hostname{"toto.foo.com", "test.foo.com"},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := matchingHostnames(test.listener, test.hostnames)
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_getAllowedRoutes(t *testing.T) {
	testCases := []struct {
		desc                string
		listener            gatev1.Listener
		supportedRouteKinds []gatev1.RouteGroupKind
		wantKinds           []gatev1.RouteGroupKind
		wantErr             bool
	}{
		{
			desc: "Empty",
		},
		{
			desc: "Empty AllowedRoutes",
			supportedRouteKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: groupPtr(gatev1.GroupName)},
			},
			wantKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: groupPtr(gatev1.GroupName)},
			},
		},
		{
			desc: "AllowedRoutes with unsupported Group",
			listener: gatev1.Listener{
				AllowedRoutes: &gatev1.AllowedRoutes{
					Kinds: []gatev1.RouteGroupKind{{
						Kind: kindTLSRoute, Group: groupPtr("foo"),
					}},
				},
			},
			supportedRouteKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: groupPtr(gatev1.GroupName)},
			},
			wantErr: true,
		},
		{
			desc: "AllowedRoutes with nil Group",
			listener: gatev1.Listener{
				AllowedRoutes: &gatev1.AllowedRoutes{
					Kinds: []gatev1.RouteGroupKind{{
						Kind: kindTLSRoute, Group: nil,
					}},
				},
			},
			supportedRouteKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: groupPtr(gatev1.GroupName)},
			},
			wantErr: true,
		},
		{
			desc: "AllowedRoutes with unsupported Kind",
			listener: gatev1.Listener{
				AllowedRoutes: &gatev1.AllowedRoutes{
					Kinds: []gatev1.RouteGroupKind{{
						Kind: "foo", Group: groupPtr(gatev1.GroupName),
					}},
				},
			},
			supportedRouteKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: groupPtr(gatev1.GroupName)},
			},
			wantErr: true,
		},
		{
			desc: "Supported AllowedRoutes",
			listener: gatev1.Listener{
				AllowedRoutes: &gatev1.AllowedRoutes{
					Kinds: []gatev1.RouteGroupKind{{
						Kind: kindTLSRoute, Group: groupPtr(gatev1.GroupName),
					}},
				},
			},
			supportedRouteKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: groupPtr(gatev1.GroupName)},
			},
			wantKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: groupPtr(gatev1.GroupName)},
			},
		},
		{
			desc: "Supported AllowedRoutes with duplicates",
			listener: gatev1.Listener{
				AllowedRoutes: &gatev1.AllowedRoutes{
					Kinds: []gatev1.RouteGroupKind{
						{Kind: kindTLSRoute, Group: groupPtr(gatev1.GroupName)},
						{Kind: kindTCPRoute, Group: groupPtr(gatev1.GroupName)},
						{Kind: kindTLSRoute, Group: groupPtr(gatev1.GroupName)},
						{Kind: kindTCPRoute, Group: groupPtr(gatev1.GroupName)},
					},
				},
			},
			supportedRouteKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: groupPtr(gatev1.GroupName)},
				{Kind: kindTCPRoute, Group: groupPtr(gatev1.GroupName)},
			},
			wantKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: groupPtr(gatev1.GroupName)},
				{Kind: kindTCPRoute, Group: groupPtr(gatev1.GroupName)},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got, conditions := getAllowedRouteKinds(&gatev1.Gateway{}, test.listener, test.supportedRouteKinds)
			if test.wantErr {
				require.NotEmpty(t, conditions, "no conditions")
				return
			}

			require.Empty(t, conditions)
			assert.Equal(t, test.wantKinds, got)
		})
	}
}

func Test_makeListenerKey(t *testing.T) {
	testCases := []struct {
		desc        string
		listener    gatev1.Listener
		expectedKey string
	}{
		{
			desc:        "empty",
			expectedKey: "||0",
		},
		{
			desc: "listener with port, protocol and hostname",
			listener: gatev1.Listener{
				Port:     443,
				Protocol: gatev1.HTTPSProtocolType,
				Hostname: hostnamePtr("www.example.com"),
			},
			expectedKey: "HTTPS|www.example.com|443",
		},
		{
			desc: "listener with port, protocol and nil hostname",
			listener: gatev1.Listener{
				Port:     443,
				Protocol: gatev1.HTTPSProtocolType,
			},
			expectedKey: "HTTPS||443",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expectedKey, makeListenerKey(test.listener))
		})
	}
}

func Test_referenceGrantMatchesFrom(t *testing.T) {
	testCases := []struct {
		desc           string
		referenceGrant gatev1beta1.ReferenceGrant
		group          string
		kind           string
		namespace      string
		expectedResult bool
	}{
		{
			desc: "matches",
			referenceGrant: gatev1beta1.ReferenceGrant{
				Spec: gatev1beta1.ReferenceGrantSpec{
					From: []gatev1beta1.ReferenceGrantFrom{
						{
							Group:     "correct-group",
							Kind:      "correct-kind",
							Namespace: "correct-namespace",
						},
					},
				},
			},
			group:          "correct-group",
			kind:           "correct-kind",
			namespace:      "correct-namespace",
			expectedResult: true,
		},
		{
			desc: "empty group matches core",
			referenceGrant: gatev1beta1.ReferenceGrant{
				Spec: gatev1beta1.ReferenceGrantSpec{
					From: []gatev1beta1.ReferenceGrantFrom{
						{
							Group:     "",
							Kind:      "correct-kind",
							Namespace: "correct-namespace",
						},
					},
				},
			},
			group:          "core",
			kind:           "correct-kind",
			namespace:      "correct-namespace",
			expectedResult: true,
		},
		{
			desc: "wrong group",
			referenceGrant: gatev1beta1.ReferenceGrant{
				Spec: gatev1beta1.ReferenceGrantSpec{
					From: []gatev1beta1.ReferenceGrantFrom{
						{
							Group:     "wrong-group",
							Kind:      "correct-kind",
							Namespace: "correct-namespace",
						},
					},
				},
			},
			group:          "correct-group",
			kind:           "correct-kind",
			namespace:      "correct-namespace",
			expectedResult: false,
		},
		{
			desc: "wrong kind",
			referenceGrant: gatev1beta1.ReferenceGrant{
				Spec: gatev1beta1.ReferenceGrantSpec{
					From: []gatev1beta1.ReferenceGrantFrom{
						{
							Group:     "correct-group",
							Kind:      "wrong-kind",
							Namespace: "correct-namespace",
						},
					},
				},
			},
			group:          "correct-group",
			kind:           "correct-kind",
			namespace:      "correct-namespace",
			expectedResult: false,
		},
		{
			desc: "wrong namespace",
			referenceGrant: gatev1beta1.ReferenceGrant{
				Spec: gatev1beta1.ReferenceGrantSpec{
					From: []gatev1beta1.ReferenceGrantFrom{
						{
							Group:     "correct-group",
							Kind:      "correct-kind",
							Namespace: "wrong-namespace",
						},
					},
				},
			},
			group:          "correct-group",
			kind:           "correct-kind",
			namespace:      "correct-namespace",
			expectedResult: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expectedResult, referenceGrantMatchesFrom(&test.referenceGrant, test.group, test.kind, test.namespace))
		})
	}
}

func Test_referenceGrantMatchesTo(t *testing.T) {
	testCases := []struct {
		desc           string
		referenceGrant gatev1beta1.ReferenceGrant
		group          string
		kind           string
		name           string
		expectedResult bool
	}{
		{
			desc: "matches",
			referenceGrant: gatev1beta1.ReferenceGrant{
				Spec: gatev1beta1.ReferenceGrantSpec{
					To: []gatev1beta1.ReferenceGrantTo{
						{
							Group: "correct-group",
							Kind:  "correct-kind",
							Name:  objectNamePtr("correct-name"),
						},
					},
				},
			},
			group:          "correct-group",
			kind:           "correct-kind",
			name:           "correct-name",
			expectedResult: true,
		},
		{
			desc: "matches without name",
			referenceGrant: gatev1beta1.ReferenceGrant{
				Spec: gatev1beta1.ReferenceGrantSpec{
					To: []gatev1beta1.ReferenceGrantTo{
						{
							Group: "correct-group",
							Kind:  "correct-kind",
							Name:  nil,
						},
					},
				},
			},
			group:          "correct-group",
			kind:           "correct-kind",
			name:           "some-name",
			expectedResult: true,
		},
		{
			desc: "empty group matches core",
			referenceGrant: gatev1beta1.ReferenceGrant{
				Spec: gatev1beta1.ReferenceGrantSpec{
					To: []gatev1beta1.ReferenceGrantTo{
						{
							Group: "",
							Kind:  "correct-kind",
							Name:  objectNamePtr("correct-name"),
						},
					},
				},
			},
			group:          "core",
			kind:           "correct-kind",
			name:           "correct-name",
			expectedResult: true,
		},
		{
			desc: "wrong group",
			referenceGrant: gatev1beta1.ReferenceGrant{
				Spec: gatev1beta1.ReferenceGrantSpec{
					To: []gatev1beta1.ReferenceGrantTo{
						{
							Group: "wrong-group",
							Kind:  "correct-kind",
							Name:  objectNamePtr("correct-name"),
						},
					},
				},
			},
			group:          "correct-group",
			kind:           "correct-kind",
			name:           "correct-namespace",
			expectedResult: false,
		},
		{
			desc: "wrong kind",
			referenceGrant: gatev1beta1.ReferenceGrant{
				Spec: gatev1beta1.ReferenceGrantSpec{
					To: []gatev1beta1.ReferenceGrantTo{
						{
							Group: "correct-group",
							Kind:  "wrong-kind",
							Name:  objectNamePtr("correct-name"),
						},
					},
				},
			},
			group:          "correct-group",
			kind:           "correct-kind",
			name:           "correct-name",
			expectedResult: false,
		},
		{
			desc: "wrong name",
			referenceGrant: gatev1beta1.ReferenceGrant{
				Spec: gatev1beta1.ReferenceGrantSpec{
					To: []gatev1beta1.ReferenceGrantTo{
						{
							Group: "correct-group",
							Kind:  "correct-kind",
							Name:  objectNamePtr("wrong-name"),
						},
					},
				},
			},
			group:          "correct-group",
			kind:           "correct-kind",
			name:           "correct-name",
			expectedResult: false,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.expectedResult, referenceGrantMatchesTo(&test.referenceGrant, test.group, test.kind, test.name))
		})
	}
}

func Test_gatewayAddresses(t *testing.T) {
	testCases := []struct {
		desc          string
		statusAddress *StatusAddress
		paths         []string
		wantErr       require.ErrorAssertionFunc
		want          []gatev1.GatewayStatusAddress
	}{
		{
			desc:    "nothing",
			wantErr: require.NoError,
		},
		{
			desc:          "empty configuration",
			statusAddress: &StatusAddress{},
			wantErr:       require.Error,
		},
		{
			desc: "IP address",
			statusAddress: &StatusAddress{
				IP: "1.2.3.4",
			},
			wantErr: require.NoError,
			want: []gatev1.GatewayStatusAddress{
				{
					Type:  ptr.To(gatev1.IPAddressType),
					Value: "1.2.3.4",
				},
			},
		},
		{
			desc: "hostname address",
			statusAddress: &StatusAddress{
				Hostname: "foo.bar",
			},
			wantErr: require.NoError,
			want: []gatev1.GatewayStatusAddress{
				{
					Type:  ptr.To(gatev1.HostnameAddressType),
					Value: "foo.bar",
				},
			},
		},
		{
			desc: "service",
			statusAddress: &StatusAddress{
				Service: ServiceRef{
					Name:      "status-address",
					Namespace: "default",
				},
			},
			paths:   []string{"services.yml"},
			wantErr: require.NoError,
			want: []gatev1.GatewayStatusAddress{
				{
					Type:  ptr.To(gatev1.HostnameAddressType),
					Value: "foo.bar",
				},
				{
					Type:  ptr.To(gatev1.IPAddressType),
					Value: "1.2.3.4",
				},
			},
		},
		{
			desc: "missing service",
			statusAddress: &StatusAddress{
				Service: ServiceRef{
					Name:      "status-address2",
					Namespace: "default",
				},
			},
			wantErr: require.Error,
		},
		{
			desc: "service without load-balancer status",
			statusAddress: &StatusAddress{
				Service: ServiceRef{
					Name:      "whoamitcp-bar",
					Namespace: "bar",
				},
			},
			paths:   []string{"services.yml"},
			wantErr: require.NoError,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{StatusAddress: test.statusAddress}

			got, err := p.gatewayAddresses(newClientMock(test.paths...))
			test.wantErr(t, err)

			assert.Equal(t, test.want, got)
		})
	}
}

func hostnamePtr(hostname gatev1.Hostname) *gatev1.Hostname {
	return &hostname
}

func groupPtr(group gatev1.Group) *gatev1.Group {
	return &group
}

func sectionNamePtr(sectionName gatev1.SectionName) *gatev1.SectionName {
	return &sectionName
}

func namespacePtr(namespace gatev1.Namespace) *gatev1.Namespace {
	return &namespace
}

func kindPtr(kind gatev1.Kind) *gatev1.Kind {
	return &kind
}

func pathMatchTypePtr(p gatev1.PathMatchType) *gatev1.PathMatchType { return &p }

func headerMatchTypePtr(h gatev1.HeaderMatchType) *gatev1.HeaderMatchType { return &h }

func objectNamePtr(objectName gatev1.ObjectName) *gatev1.ObjectName {
	return &objectName
}
