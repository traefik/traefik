package gateway

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/provider"
	traefikv1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1alpha1"
	"github.com/traefik/traefik/v3/pkg/provider/kubernetes/k8s"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
	"google.golang.org/grpc/codes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	kscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	gatev1 "sigs.k8s.io/gateway-api/apis/v1"
	gatev1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatev1alpha3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	gatev1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	gatefake "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned/fake"
)

var _ provider.Provider = (*Provider)(nil)

func init() {
	// required by k8s.MustParseYaml
	if err := gatev1.AddToScheme(kscheme.Scheme); err != nil {
		panic(err)
	}
	if err := gatev1beta1.AddToScheme(kscheme.Scheme); err != nil {
		panic(err)
	}
	if err := gatev1alpha2.AddToScheme(kscheme.Scheme); err != nil {
		panic(err)
	}
	if err := gatev1alpha3.AddToScheme(kscheme.Scheme); err != nil {
		panic(err)
	}
}

func TestGatewayClassLabelSelector(t *testing.T) {
	k8sObjects, gwObjects := readResources(t, []string{"gatewayclass_labelselector.yaml"})

	kubeClient := kubefake.NewSimpleClientset(k8sObjects...)
	gwClient := newGatewaySimpleClientSet(t, gwObjects...)

	client := newClientImpl(kubeClient, gwClient)

	// This is initialized by the Provider init method but this cannot be called in a unit test.
	client.labelSelector = "name=traefik-internal"

	eventCh, err := client.WatchAll(nil, make(chan struct{}))
	require.NoError(t, err)

	if len(k8sObjects) > 0 || len(gwObjects) > 0 {
		// just wait for the first event
		<-eventCh
	}

	p := Provider{
		EntryPoints:   map[string]Entrypoint{"http": {Address: ":9080"}},
		StatusAddress: &StatusAddress{IP: "1.2.3.4"},
		client:        client,
	}

	_ = p.loadConfigurationFromGateways(t.Context())

	gw, err := gwClient.GatewayV1().Gateways("default").Get(t.Context(), "traefik-external", metav1.GetOptions{})
	require.NoError(t, err)

	assert.Empty(t, gw.Status.Addresses)

	gw, err = gwClient.GatewayV1().Gateways("default").Get(t.Context(), "traefik-internal", metav1.GetOptions{})
	require.NoError(t, err)
	require.Len(t, gw.Status.Addresses, 1)
	require.NotNil(t, gw.Status.Addresses[0].Type)

	assert.Equal(t, gatev1.IPAddressType, *gw.Status.Addresses[0].Type)
	assert.Equal(t, "1.2.3.4", gw.Status.Addresses[0].Value)
}

func TestLoadHTTPRoutes(t *testing.T) {
	testCases := []struct {
		desc                string
		ingressClass        string
		paths               []string
		expected            *dynamic.Configuration
		entryPoints         map[string]Entrypoint
		experimentalChannel bool
		nativeLB            bool
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
			desc: "Router with service in error caused by wrong TargetPort",
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
					Routers: map[string]*dynamic.Router{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-9000",
										Weight: ptr.To(1),
										Status: ptr.To(500),
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "api@internal",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-websecure-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"websecure"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-websecure-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-websecure-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-66e726cd8903b49727ae": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-66e726cd8903b49727ae-wrr",
							Rule:        "(Host(`foo.com`) || Host(`bar.com`)) && PathPrefix(`/`)",
							Priority:    9,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-66e726cd8903b49727ae-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-baa117c0219e3878749f": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-baa117c0219e3878749f-wrr",
							Rule:        "(Host(`foo.com`) || HostRegexp(`^[a-z0-9-\\.]+\\.bar\\.com$`)) && PathPrefix(`/`)",
							Priority:    11,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-baa117c0219e3878749f-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-45eba2eaf40ac792e036": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-45eba2eaf40ac792e036-wrr",
							Rule:        "(Host(`foo.com`) || HostRegexp(`^[a-z0-9-\\.]+\\.foo\\.com$`)) && PathPrefix(`/`)",
							Priority:    11,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-45eba2eaf40ac792e036-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100009,
							RuleSyntax:  "default",
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
						},
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-1-d737b4933fa88e68ab8a": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && Path(`/bir`)",
							Priority:    100008,
							RuleSyntax:  "default",
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-1-d737b4933fa88e68ab8a-wrr",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-1-d737b4933fa88e68ab8a-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami2-http-8080",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami2-http-8080": {
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
									{
										Name:   "default-whoami2-http-8080",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami2-http-8080": {
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
						"httproute-default-http-app-1-gw-default-my-gateway-http-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-http-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
						"httproute-default-http-app-1-gw-default-my-gateway-https-ep-websecure-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"websecure"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-https-ep-websecure-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-http-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"httproute-default-http-app-1-gw-default-my-gateway-https-ep-websecure-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
						"httproute-default-http-app-1-gw-default-my-gateway-ep-websecure-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"websecure"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-websecure-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"httproute-default-http-app-1-gw-default-my-gateway-ep-websecure-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-6cf37fa71907768d925c": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && (Path(`/bar`) || PathPrefix(`/bar/`)) && Header(`my-header`,`foo`) && Header(`my-header2`,`bar`)",
							Priority:    10610,
							RuleSyntax:  "default",
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-6cf37fa71907768d925c-wrr",
						},
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-2-d23f7039bc8036fb918c": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && PathRegexp(`^/buzz/[0-9]+$`)",
							Priority:    11408,
							RuleSyntax:  "default",
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-2-d23f7039bc8036fb918c-wrr",
						},
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-1-aaba0f24fd26e1ca2276": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && Path(`/bar`) && Header(`my-header`,`bar`)",
							Priority:    100109,
							RuleSyntax:  "default",
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-1-aaba0f24fd26e1ca2276-wrr",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-6cf37fa71907768d925c-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-2-d23f7039bc8036fb918c-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-1-aaba0f24fd26e1ca2276-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
			desc:  "Simple HTTPRoute, with method matching",
			paths: []string{"services.yml", "httproute/with_method_matching.yml"},
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-74ad70a7cf090becdd3c": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && (Path(`/foo`) || PathPrefix(`/foo/`)) && Method(`GET`)",
							Priority:    11408,
							RuleSyntax:  "default",
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-74ad70a7cf090becdd3c-wrr",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-74ad70a7cf090becdd3c-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
			desc:  "Simple HTTPRoute, with query param matching",
			paths: []string{"services.yml", "httproute/with_query_param_matching.yml"},
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-bb7b03c9610e982fd627": {
							EntryPoints: []string{"web"},
							Rule:        "Host(`foo.com`) && (Path(`/foo`) || PathPrefix(`/foo/`)) && Query(`foo`,`bar`) && QueryRegexp(`baz`,`buz`)",
							Priority:    10428,
							RuleSyntax:  "default",
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-bb7b03c9610e982fd627-wrr",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-bb7b03c9610e982fd627-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
						"httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-efde1997778109a1f6eb": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-efde1997778109a1f6eb-wrr",
							Rule:        "Host(`foo.com`) && Path(`/foo`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-efde1997778109a1f6eb-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
						"httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-efde1997778109a1f6eb": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-efde1997778109a1f6eb-wrr",
							Rule:        "Host(`foo.com`) && Path(`/foo`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
						"httproute-bar-http-app-bar-gw-default-my-gateway-ep-web-0-66f5c78d03d948e36597": {
							EntryPoints: []string{"web"},
							Service:     "httproute-bar-http-app-bar-gw-default-my-gateway-ep-web-0-66f5c78d03d948e36597-wrr",
							Rule:        "Host(`bar.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-efde1997778109a1f6eb-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"httproute-bar-http-app-bar-gw-default-my-gateway-ep-web-0-66f5c78d03d948e36597-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "bar-whoami-bar-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"bar-whoami-bar-http-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
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
						"httproute-bar-http-app-bar-gw-default-my-gateway-ep-web-0-66f5c78d03d948e36597": {
							EntryPoints: []string{"web"},
							Service:     "httproute-bar-http-app-bar-gw-default-my-gateway-ep-web-0-66f5c78d03d948e36597-wrr",
							Rule:        "Host(`bar.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-bar-http-app-bar-gw-default-my-gateway-ep-web-0-66f5c78d03d948e36597-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "bar-whoami-bar-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"bar-whoami-bar-http-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-wrr",
							Rule:        "Host(`example.org`) && PathPrefix(`/`)",
							Priority:    13,
							RuleSyntax:  "default",
							Middlewares: []string{"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-requestheadermodifier-0"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-requestheadermodifier-0": {
							RequestHeaderModifier: &dynamic.HeaderModifier{
								Set:    map[string]string{"X-Foo": "Bar"},
								Add:    map[string]string{"X-Bar": "Foo"},
								Remove: []string{"X-Baz"},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
			desc:  "Simple HTTPRoute, response header modifier",
			paths: []string{"services.yml", "httproute/filter_response_header_modifier.yml"},
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-wrr",
							Rule:        "Host(`example.org`) && PathPrefix(`/`)",
							Priority:    13,
							RuleSyntax:  "default",
							Middlewares: []string{"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-responseheadermodifier-0"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-responseheadermodifier-0": {
							ResponseHeaderModifier: &dynamic.HeaderModifier{
								Set:    map[string]string{"X-Foo": "Bar"},
								Add:    map[string]string{"X-Bar": "Foo"},
								Remove: []string{"X-Baz"},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-wrr",
							Rule:        "Host(`example.org`) && PathPrefix(`/`)",
							Priority:    13,
							RuleSyntax:  "default",
							Middlewares: []string{"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-requestredirect-0"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-requestredirect-0": {
							RequestRedirect: &dynamic.RequestRedirect{
								Scheme:     ptr.To("https"),
								Port:       ptr.To(""),
								StatusCode: http.StatusMovedPermanently,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{},
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-wrr",
							Rule:        "Host(`example.org`) && PathPrefix(`/`)",
							Priority:    13,
							RuleSyntax:  "default",
							Middlewares: []string{"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-requestredirect-0"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-requestredirect-0": {
							RequestRedirect: &dynamic.RequestRedirect{
								Hostname:   ptr.To("example.com"),
								Port:       ptr.To("443"),
								StatusCode: http.StatusFound,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-364ce6ec04c3d49b19c4-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Simple HTTPRoute URL rewrite FullPath",
			paths: []string{"services.yml", "httproute/filter_url_rewrite_fullpath.yml"},
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8-wrr",
							Rule:        "Host(`example.com`) && (Path(`/foo`) || PathPrefix(`/foo/`))",
							RuleSyntax:  "default",
							Priority:    10412,
							Middlewares: []string{"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8-urlrewrite-0"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8-urlrewrite-0": {
							URLRewrite: &dynamic.URLRewrite{
								Path: ptr.To("/bar"),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
			desc:  "Simple HTTPRoute URL rewrite Hostname",
			paths: []string{"services.yml", "httproute/filter_url_rewrite_hostname.yml"},
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8-wrr",
							Rule:        "Host(`example.com`) && (Path(`/foo`) || PathPrefix(`/foo/`))",
							RuleSyntax:  "default",
							Priority:    10412,
							Middlewares: []string{"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8-urlrewrite-0"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8-urlrewrite-0": {
							URLRewrite: &dynamic.URLRewrite{
								Hostname: ptr.To("www.foo.bar"),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
			desc:  "Simple HTTPRoute URL rewrite Combined",
			paths: []string{"services.yml", "httproute/filter_url_rewrite_combined.yml"},
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8-wrr",
							Rule:        "Host(`example.com`) && (Path(`/foo`) || PathPrefix(`/foo/`))",
							RuleSyntax:  "default",
							Priority:    10412,
							Middlewares: []string{"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8-urlrewrite-0"},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8-urlrewrite-0": {
							URLRewrite: &dynamic.URLRewrite{
								Hostname:   ptr.To("www.foo.bar"),
								Path:       ptr.To("/xyz"),
								PathPrefix: ptr.To("/foo"),
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-7f90cf546b15efadf2f8-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: func(i int) *int { return &i }(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
			desc:  "Simple HTTPRoute and BackendTLSPolicy, experimental channel disabled",
			paths: []string{"services.yml", "httproute/with_backend_tls_policy.yml"},
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
			desc:  "Simple HTTPRoute and BackendTLSPolicy with CA certificate, experimental channel enabled",
			paths: []string{"services.yml", "httproute/with_backend_tls_policy.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
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
					Routers: map[string]*dynamic.Router{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.1:80",
									},
									{
										URL: "https://10.10.0.2:80",
									},
								},
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "default-whoami-http-80",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-whoami-http-80": {
							ServerName: "whoami",
							RootCAs: []types.FileOrContent{
								"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=",
								"LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0=",
							},
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:  "Simple HTTPRoute and BackendTLSPolicy with System CA, experimental channel enabled",
			paths: []string{"services.yml", "httproute/with_backend_tls_policy_system.yml"},
			entryPoints: map[string]Entrypoint{"web": {
				Address: ":80",
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
					Routers: map[string]*dynamic.Router{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.1:80",
									},
									{
										URL: "https://10.10.0.2:80",
									},
								},
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
								ServersTransport: "default-whoami-http-80",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"default-whoami-http-80": {
							ServerName: "whoami",
						},
					},
				},
				TLS: &dynamic.TLSConfiguration{},
			},
		},
		{
			desc:     "Simple HTTPRoute with NativeLBByDefault enabled",
			paths:    []string{"services.yml", "httproute/simple.yml"},
			nativeLB: true,
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.10.1:80",
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
			desc:  "Simple HTTPRoute with NativeLB annotation",
			paths: []string{"services.yml", "httproute/simple_nativelb.yml"},
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-native-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-native-http-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.10.1:80",
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
			desc:     "Simple HTTPRoute with NativeLBByDefault enabled but service has disabled nativelb",
			paths:    []string{"services.yml", "httproute/simple_nativelb_disabled.yml"},
			nativeLB: true,
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-native-disabled-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-native-disabled-http-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.20:80",
									},
									{
										URL: "http://10.10.0.21:80",
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

			k8sObjects, gwObjects := readResources(t, test.paths)

			kubeClient := kubefake.NewSimpleClientset(k8sObjects...)
			gwClient := newGatewaySimpleClientSet(t, gwObjects...)

			client := newClientImpl(kubeClient, gwClient)
			client.experimentalChannel = test.experimentalChannel

			eventCh, err := client.WatchAll(nil, make(chan struct{}))
			require.NoError(t, err)

			if len(k8sObjects) > 0 || len(gwObjects) > 0 {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				EntryPoints:         test.entryPoints,
				ExperimentalChannel: test.experimentalChannel,
				NativeLBByDefault:   test.nativeLB,
				client:              client,
			}

			conf := p.loadConfigurationFromGateways(t.Context())
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "whoami",
										Weight: ptr.To(1),
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
					return name, &dynamic.Service{LoadBalancer: &dynamic.ServersLoadBalancer{Strategy: dynamic.BalancerStrategyWRR, Servers: []dynamic.Server{{URL: "foobar"}}}}, nil
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "whoami",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"whoami": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
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
					Routers: map[string]*dynamic.Router{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http",
										Weight: ptr.To(1),
										Status: ptr.To(500),
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
					Routers: map[string]*dynamic.Router{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http",
										Weight: ptr.To(1),
										Status: ptr.To(500),
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "service@file",
										Weight: ptr.To(1),
									},
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
			desc:  "Simple HTTPRoute, with appProtocol service",
			paths: []string{"services.yml", "httproute/with_app_protocol_service.yml"},
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
						"httproute-default-http-multi-protocols-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-multi-protocols-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-multi-protocols-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-h2c-http-80",
										Weight: ptr.To(1),
									},
									{
										Name:   "default-whoami-ws-http-80",
										Weight: ptr.To(1),
									},
									{
										Name:   "default-whoami-wss-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-h2c-http-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "h2c://10.10.0.13:80",
									},
								},
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami-ws-http-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.10.0.14:80",
									},
								},
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"default-whoami-wss-http-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "https://10.10.0.15:80",
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

			k8sObjects, gwObjects := readResources(t, test.paths)

			kubeClient := kubefake.NewSimpleClientset(k8sObjects...)
			gwClient := newGatewaySimpleClientSet(t, gwObjects...)

			client := newClientImpl(kubeClient, gwClient)

			eventCh, err := client.WatchAll(nil, make(chan struct{}))
			require.NoError(t, err)

			if len(k8sObjects) > 0 || len(gwObjects) > 0 {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				EntryPoints: test.entryPoints,
				client:      client,
			}

			for group, kindFuncs := range test.groupKindBackendFuncs {
				for kind, backendFunc := range kindFuncs {
					p.RegisterBackendFuncs(group, kind, backendFunc)
				}
			}
			conf := p.loadConfigurationFromGateways(t.Context())
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
			desc: "ExtensionRef filter",
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-my-first-middleware",
								"default-my-second-middleware",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
			desc: "ExtensionRef filter with middleware creation",
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-my-first-middleware",
								"default-my-second-middleware",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-my-first-middleware":  {Headers: &dynamic.Headers{CustomRequestHeaders: map[string]string{"Test-Header": "Test"}}},
						"default-my-second-middleware": {Headers: &dynamic.Headers{CustomRequestHeaders: map[string]string{"Test-Header": "Test"}}},
					},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
			desc: "Unknown ExtensionRef filter",
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-err-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-err-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "invalid-httproute-filter",
										Weight: ptr.To(1),
										Status: ptr.To(500),
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
					Routers: map[string]*dynamic.Router{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-err-wrr",
							Rule:        "Host(`foo.com`) && Path(`/bar`)",
							Priority:    100008,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-1c0cf64bde37d9d0df06-err-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "invalid-httproute-filter",
										Weight: ptr.To(1),
										Status: ptr.To(500),
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.expected == nil {
				return
			}

			k8sObjects, gwObjects := readResources(t, []string{"services.yml", "httproute/filter_extension_ref.yml"})

			kubeClient := kubefake.NewSimpleClientset(k8sObjects...)
			gwClient := newGatewaySimpleClientSet(t, gwObjects...)

			client := newClientImpl(kubeClient, gwClient)

			eventCh, err := client.WatchAll(nil, make(chan struct{}))
			require.NoError(t, err)

			if len(k8sObjects) > 0 || len(gwObjects) > 0 {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				EntryPoints: test.entryPoints,
				client:      client,
			}

			for group, kindFuncs := range test.groupKindFilterFuncs {
				for kind, filterFunc := range kindFuncs {
					p.RegisterFilterFuncs(group, kind, filterFunc)
				}
			}
			conf := p.loadConfigurationFromGateways(t.Context())
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadGRPCRoutes_filterExtensionRef(t *testing.T) {
	testCases := []struct {
		desc                 string
		groupKindFilterFuncs map[string]map[string]BuildFilterFunc
		expected             *dynamic.Configuration
		entryPoints          map[string]Entrypoint
	}{
		{
			desc: "ExtensionRef filter",
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
						"grpcroute-default-grpc-app-1-gw-default-my-gateway-ep-web-0-74471866db6e94e08d00": {
							EntryPoints: []string{"web"},
							Service:     "grpcroute-default-grpc-app-1-gw-default-my-gateway-ep-web-0-74471866db6e94e08d00-wrr",
							Rule:        "Host(`foo.com`) && PathPrefix(`/`)",
							Priority:    22,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-my-first-middleware",
								"default-my-second-middleware",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"grpcroute-default-grpc-app-1-gw-default-my-gateway-ep-web-0-74471866db6e94e08d00-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80-grpc",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-80-grpc": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "h2c://10.10.0.1:80",
									},
									{
										URL: "h2c://10.10.0.2:80",
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
			desc: "ExtensionRef filter with middleware creation",
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
						"grpcroute-default-grpc-app-1-gw-default-my-gateway-ep-web-0-74471866db6e94e08d00": {
							EntryPoints: []string{"web"},
							Service:     "grpcroute-default-grpc-app-1-gw-default-my-gateway-ep-web-0-74471866db6e94e08d00-wrr",
							Rule:        "Host(`foo.com`) && PathPrefix(`/`)",
							Priority:    22,
							RuleSyntax:  "default",
							Middlewares: []string{
								"default-my-first-middleware",
								"default-my-second-middleware",
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"default-my-first-middleware":  {Headers: &dynamic.Headers{CustomRequestHeaders: map[string]string{"Test-Header": "Test"}}},
						"default-my-second-middleware": {Headers: &dynamic.Headers{CustomRequestHeaders: map[string]string{"Test-Header": "Test"}}},
					},
					Services: map[string]*dynamic.Service{
						"grpcroute-default-grpc-app-1-gw-default-my-gateway-ep-web-0-74471866db6e94e08d00-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-80-grpc",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-80-grpc": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "h2c://10.10.0.1:80",
									},
									{
										URL: "h2c://10.10.0.2:80",
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
			desc: "Unknown ExtensionRef filter",
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
						"grpcroute-default-grpc-app-1-gw-default-my-gateway-ep-web-0-74471866db6e94e08d00": {
							EntryPoints: []string{"web"},
							Service:     "grpcroute-default-grpc-app-1-gw-default-my-gateway-ep-web-0-74471866db6e94e08d00-err-wrr",
							Rule:        "Host(`foo.com`) && PathPrefix(`/`)",
							Priority:    22,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"grpcroute-default-grpc-app-1-gw-default-my-gateway-ep-web-0-74471866db6e94e08d00-err-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "invalid-grpcroute-filter",
										Weight: ptr.To(1),
										GRPCStatus: &dynamic.GRPCStatus{
											Code: codes.Unavailable,
											Msg:  "Service Unavailable",
										},
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
					Routers: map[string]*dynamic.Router{
						"grpcroute-default-grpc-app-1-gw-default-my-gateway-ep-web-0-74471866db6e94e08d00": {
							EntryPoints: []string{"web"},
							Service:     "grpcroute-default-grpc-app-1-gw-default-my-gateway-ep-web-0-74471866db6e94e08d00-err-wrr",
							Rule:        "Host(`foo.com`) && PathPrefix(`/`)",
							Priority:    22,
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"grpcroute-default-grpc-app-1-gw-default-my-gateway-ep-web-0-74471866db6e94e08d00-err-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "invalid-grpcroute-filter",
										Weight: ptr.To(1),
										GRPCStatus: &dynamic.GRPCStatus{
											Code: codes.Unavailable,
											Msg:  "Service Unavailable",
										},
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.expected == nil {
				return
			}

			k8sObjects, gwObjects := readResources(t, []string{"services.yml", "grpcroute/filter_extension_ref.yml"})

			kubeClient := kubefake.NewSimpleClientset(k8sObjects...)
			gwClient := newGatewaySimpleClientSet(t, gwObjects...)

			client := newClientImpl(kubeClient, gwClient)

			eventCh, err := client.WatchAll(nil, make(chan struct{}))
			require.NoError(t, err)

			if len(k8sObjects) > 0 || len(gwObjects) > 0 {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				EntryPoints: test.entryPoints,
				client:      client,
			}

			for group, kindFuncs := range test.groupKindFilterFuncs {
				for kind, filterFunc := range kindFuncs {
					p.RegisterFilterFuncs(group, kind, filterFunc)
				}
			}
			conf := p.loadConfigurationFromGateways(t.Context())
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
		nativeLB     bool
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
			desc: "Router with service in error caused by wrong TargetPort",
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
					Routers: map[string]*dynamic.TCPRouter{
						"tcproute-default-TCP-app-1-gw-default-my-gateway-ep-TCP-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"TCP"},
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							Service:     "tcproute-default-TCP-app-1-gw-default-my-gateway-ep-TCP-0-e3b0c44298fc1c149afb-wrr",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-TCP-app-1-gw-default-my-gateway-ep-TCP-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{{
									Name:   "tcproute-default-TCP-app-1-gw-default-my-gateway-ep-TCP-0-e3b0c44298fc1c149afb-err-lb",
									Weight: ptr.To(1),
								}},
							},
						},
						"tcproute-default-TCP-app-1-gw-default-my-gateway-ep-TCP-0-e3b0c44298fc1c149afb-err-lb": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{},
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
						"tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-1-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp-1"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-1-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
						"tcproute-default-tcp-app-2-gw-default-my-tcp-gateway-ep-tcp-2-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp-2"},
							Service:     "tcproute-default-tcp-app-2-gw-default-my-tcp-gateway-ep-tcp-2-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-1-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tcproute-default-tcp-app-2-gw-default-my-tcp-gateway-ep-tcp-2-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-10000",
										Weight: ptr.To(1),
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
						"tcproute-default-tcp-app-gw-default-my-tcp-gateway-ep-tcp-1-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp-1"},
							Service:     "tcproute-default-tcp-app-gw-default-my-tcp-gateway-ep-tcp-1-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
						"tcproute-default-tcp-app-gw-default-my-tcp-gateway-ep-tcp-1-1-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp-1"},
							Service:     "tcproute-default-tcp-app-gw-default-my-tcp-gateway-ep-tcp-1-1-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-gw-default-my-tcp-gateway-ep-tcp-1-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tcproute-default-tcp-app-gw-default-my-tcp-gateway-ep-tcp-1-1-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-10000",
										Weight: ptr.To(1),
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
						"tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "service@file",
										Weight: ptr.To(1),
									},
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{{
									Name:   "default-whoamitcp-9000",
									Weight: ptr.To(1),
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
						"tcproute-default-tcp-app-default-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-default-tcp-app-default-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-default-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"tcproute-default-tcp-app-default-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-default-tcp-app-default-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
						"tcproute-bar-tcp-app-bar-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-bar-tcp-app-bar-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-default-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tcproute-bar-tcp-app-bar-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: ptr.To(1),
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
						"tcproute-bar-tcp-app-bar-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-bar-tcp-app-bar-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-bar-tcp-app-bar-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: ptr.To(1),
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
			desc:     "Simple TCPRoute with NativeLBByDefault",
			paths:    []string{"services.yml", "tcproute/simple.yml"},
			nativeLB: true,
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
						"tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.10.1:9000",
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
			desc:  "Simple TCPRoute with NativeLB annotation",
			paths: []string{"services.yml", "tcproute/simple_nativelb.yml"},
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
						"tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-native-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoamitcp-native-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.10.1:9000",
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
			desc:     "Simple TCPRoute with NativeLBByDefault enabled but service has disabled nativelb",
			paths:    []string{"services.yml", "tcproute/simple_nativelb_disabled.yml"},
			nativeLB: true,
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
						"tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-1-gw-default-my-tcp-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-native-disabled-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoamitcp-native-disabled-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.0.30:9000",
									},
									{
										Address: "10.10.0.31:9000",
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

			k8sObjects, gwObjects := readResources(t, test.paths)

			kubeClient := kubefake.NewSimpleClientset(k8sObjects...)
			gwClient := newGatewaySimpleClientSet(t, gwObjects...)

			client := newClientImpl(kubeClient, gwClient)
			client.experimentalChannel = true

			eventCh, err := client.WatchAll(nil, make(chan struct{}))
			require.NoError(t, err)

			if len(k8sObjects) > 0 || len(gwObjects) > 0 {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				EntryPoints:         test.entryPoints,
				NativeLBByDefault:   test.nativeLB,
				ExperimentalChannel: true,
				client:              client,
			}

			conf := p.loadConfigurationFromGateways(t.Context())
			assert.Equal(t, test.expected, conf)
		})
	}
}

func TestLoadTLSRoutes(t *testing.T) {
	testCases := []struct {
		desc         string
		ingressClass string
		paths        []string
		entryPoints  map[string]Entrypoint
		nativeLB     bool
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
			desc: "Router with service in error caused by wrong TargetPort",
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
					Routers: map[string]*dynamic.TCPRouter{
						"tlsroute-default-tls-app-1-gw-default-my-gateway-ep-TCP-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"TCP"},
							Priority:    0,
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							Service:     "tlsroute-default-tls-app-1-gw-default-my-gateway-ep-TCP-0-e3b0c44298fc1c149afb-wrr",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tlsroute-default-tls-app-1-gw-default-my-gateway-ep-TCP-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{{
									Name:   "tlsroute-default-tls-app-1-gw-default-my-gateway-ep-TCP-0-e3b0c44298fc1c149afb-err-lb",
									Weight: ptr.To(1),
								}},
							},
						},
						"tlsroute-default-tls-app-1-gw-default-my-gateway-ep-TCP-0-e3b0c44298fc1c149afb-err-lb": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{},
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
						"tcproute-default-tcp-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"tcproute-default-tcp-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"tlsroute-default-tls-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tlsroute-default-tls-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Priority:    15,
							Rule:        "HostSNI(`foo.example.com`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tlsroute-default-tls-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"tcproute-default-tcp-app-1-gw-default-my-tls-gateway-ep-tls-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-tls-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
						"tlsroute-default-tls-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tlsroute-default-tls-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Priority:    0,
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-1-gw-default-my-tls-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tlsroute-default-tls-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-10000",
										Weight: ptr.To(1),
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
						"tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "service@file",
										Weight: ptr.To(1),
									},
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr",
							Priority:    15,
							Rule:        "HostSNI(`foo.example.com`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr",
							Priority:    15,
							Rule:        "HostSNI(`foo.example.com`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr",
							Priority:    15,
							Rule:        "HostSNI(`foo.example.com`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr",
							Priority:    15,
							Rule:        "HostSNI(`foo.example.com`) || HostSNI(`bar.example.com`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"tlsroute-default-tls-app-default-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "tlsroute-default-tls-app-default-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr",
							Priority:    11,
							Rule:        "HostSNI(`foo.default`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tlsroute-default-tls-app-default-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"tlsroute-default-tls-app-default-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "tlsroute-default-tls-app-default-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr",
							Priority:    11,
							Rule:        "HostSNI(`foo.default`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
						"tlsroute-bar-tls-app-bar-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "tlsroute-bar-tls-app-bar-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr",
							Priority:    7,
							Rule:        "HostSNI(`foo.bar`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tlsroute-default-tls-app-default-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tlsroute-bar-tls-app-bar-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: ptr.To(1),
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
						"tlsroute-bar-tls-app-bar-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "tlsroute-bar-tls-app-bar-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr",
							Priority:    7,
							Rule:        "HostSNI(`foo.bar`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tlsroute-bar-tls-app-bar-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: ptr.To(1),
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
						"tlsroute-default-tls-app-gw-default-my-gateway-ep-tcp-1-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp-1"},
							Service:     "tlsroute-default-tls-app-gw-default-my-gateway-ep-tcp-1-0-e3b0c44298fc1c149afb-wrr",
							Priority:    0,
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
						"tlsroute-default-tls-app-gw-default-my-gateway-ep-tcp-1-1-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp-1"},
							Service:     "tlsroute-default-tls-app-gw-default-my-gateway-ep-tcp-1-1-e3b0c44298fc1c149afb-wrr",
							Priority:    0,
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tlsroute-default-tls-app-gw-default-my-gateway-ep-tcp-1-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tlsroute-default-tls-app-gw-default-my-gateway-ep-tcp-1-1-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-10000",
										Weight: ptr.To(1),
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
			desc:     "Simple TLSRoute with NativeLBByDefault",
			paths:    []string{"services.yml", "tlsroute/simple_TLS_to_TLSRoute.yml"},
			nativeLB: true,
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
						"tlsroute-default-tls-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tlsroute-default-tls-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Priority:    15,
							Rule:        "HostSNI(`foo.example.com`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tlsroute-default-tls-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoamitcp-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.10.1:9000",
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
			desc:  "Simple TLSRoute with NativeLB annotation",
			paths: []string{"services.yml", "tlsroute/simple_nativelb.yml"},
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
						"tlsroute-default-tls-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tlsroute-default-tls-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Priority:    15,
							Rule:        "HostSNI(`foo.example.com`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tlsroute-default-tls-app-1-gw-default-my-tls-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-native-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoamitcp-native-9000": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "10.10.10.1:9000",
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

			k8sObjects, gwObjects := readResources(t, test.paths)

			kubeClient := kubefake.NewSimpleClientset(k8sObjects...)
			gwClient := newGatewaySimpleClientSet(t, gwObjects...)

			client := newClientImpl(kubeClient, gwClient)
			client.experimentalChannel = true

			eventCh, err := client.WatchAll(nil, make(chan struct{}))
			require.NoError(t, err)

			if len(k8sObjects) > 0 || len(gwObjects) > 0 {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				EntryPoints:         test.entryPoints,
				NativeLBByDefault:   test.nativeLB,
				ExperimentalChannel: true,
				client:              client,
			}

			conf := p.loadConfigurationFromGateways(t.Context())
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
						"tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
						"tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-1"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
						"tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-2-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-2"},
							Service:     "tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-2-0-e3b0c44298fc1c149afb-wrr",
							Priority:    24,
							Rule:        "HostSNI(`pass.tls.foo.example.com`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tlsroute-default-tls-app-1-gw-default-my-gateway-ep-tls-2-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							Priority:    2,
							RuleSyntax:  "default",
						},
						"httproute-default-http-app-1-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd": {
							EntryPoints: []string{"websecure"},
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							Priority:    2,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"httproute-default-http-app-1-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
						"tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
						"tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-1"},
							Service:     "tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
						"tlsroute-default-tls-app-default-gw-default-my-gateway-ep-tls-2-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-2"},
							Service:     "tlsroute-default-tls-app-default-gw-default-my-gateway-ep-tls-2-0-e3b0c44298fc1c149afb-wrr",
							Priority:    24,
							Rule:        "HostSNI(`pass.tls.foo.example.com`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tlsroute-default-tls-app-default-gw-default-my-gateway-ep-tls-2-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							Priority:    2,
							RuleSyntax:  "default",
						},
						"httproute-default-http-app-default-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd": {
							EntryPoints: []string{"websecure"},
							Service:     "httproute-default-http-app-default-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							Priority:    2,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"httproute-default-http-app-default-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
						"tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
						"tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-1"},
							Service:     "tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
						"tlsroute-default-tls-app-default-gw-default-my-gateway-ep-tls-2-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-2"},
							Service:     "tlsroute-default-tls-app-default-gw-default-my-gateway-ep-tls-2-0-e3b0c44298fc1c149afb-wrr",
							Priority:    24,
							Rule:        "HostSNI(`pass.tls.foo.example.com`)",
							RuleSyntax:  "default",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: true,
							},
						},
						"tcproute-bar-tcp-app-bar-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-bar-tcp-app-bar-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
						"tcproute-bar-tcp-app-bar-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-1"},
							Service:     "tcproute-bar-tcp-app-bar-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tlsroute-default-tls-app-default-gw-default-my-gateway-ep-tls-2-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"tcproute-bar-tcp-app-bar-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tcproute-bar-tcp-app-bar-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							Priority:    2,
							RuleSyntax:  "default",
						},
						"httproute-default-http-app-default-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd": {
							EntryPoints: []string{"websecure"},
							Service:     "httproute-default-http-app-default-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							Priority:    2,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
						},
						"httproute-bar-http-app-bar-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd": {
							EntryPoints: []string{"web"},
							Service:     "httproute-bar-http-app-bar-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							Priority:    2,
							RuleSyntax:  "default",
						},
						"httproute-bar-http-app-bar-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd": {
							EntryPoints: []string{"websecure"},
							Service:     "httproute-bar-http-app-bar-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							Priority:    2,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"httproute-default-http-app-default-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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
								PassHostHeader: ptr.To(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"bar-whoami-bar-http-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
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
						"httproute-bar-http-app-bar-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "bar-whoami-bar-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"httproute-bar-http-app-bar-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "bar-whoami-bar-http-80",
										Weight: ptr.To(1),
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
						"tcproute-bar-tcp-app-bar-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-bar-tcp-app-bar-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
						"tcproute-bar-tcp-app-bar-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-1"},
							Service:     "tcproute-bar-tcp-app-bar-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
						"tlsroute-bar-tls-app-bar-gw-default-my-gateway-ep-tls-2-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls-2"},
							Service:     "tlsroute-bar-tls-app-bar-gw-default-my-gateway-ep-tls-2-0-e3b0c44298fc1c149afb-wrr",
							Priority:    24,
							Rule:        "HostSNI(`pass.tls.foo.example.com`)",
							RuleSyntax:  "default",
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
						"tcproute-bar-tcp-app-bar-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tcproute-bar-tcp-app-bar-gw-default-my-gateway-ep-tls-1-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tlsroute-bar-tls-app-bar-gw-default-my-gateway-ep-tls-2-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "bar-whoamitcp-bar-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"httproute-bar-http-app-bar-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd": {
							EntryPoints: []string{"web"},
							Service:     "httproute-bar-http-app-bar-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							Priority:    2,
							RuleSyntax:  "default",
						},
						"httproute-bar-http-app-bar-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd": {
							EntryPoints: []string{"websecure"},
							Service:     "httproute-bar-http-app-bar-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							Priority:    2,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"bar-whoami-bar-http-80": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
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
						"httproute-bar-http-app-bar-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "bar-whoami-bar-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"httproute-bar-http-app-bar-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "bar-whoami-bar-http-80",
										Weight: ptr.To(1),
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
						"tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tcp"},
							Service:     "tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
						},
						"tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tcp-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"tcproute-default-tcp-app-default-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{
									{
										Name:   "default-whoamitcp-9000",
										Weight: ptr.To(1),
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
						"httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd": {
							EntryPoints: []string{"web"},
							Service:     "httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							Priority:    2,
							RuleSyntax:  "default",
						},
						"httproute-default-http-app-default-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd": {
							EntryPoints: []string{"websecure"},
							Service:     "httproute-default-http-app-default-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd-wrr",
							Rule:        "PathPrefix(`/`)",
							Priority:    2,
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-default-gw-default-my-gateway-ep-web-0-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"httproute-default-http-app-default-gw-default-my-gateway-ep-websecure-0-a431b128267aabc954fd-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-http-80",
										Weight: ptr.To(1),
									},
								},
							},
						},
						"default-whoami-http-80": {
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

			k8sObjects, gwObjects := readResources(t, test.paths)

			kubeClient := kubefake.NewSimpleClientset(k8sObjects...)
			gwClient := newGatewaySimpleClientSet(t, gwObjects...)

			client := newClientImpl(kubeClient, gwClient)
			client.experimentalChannel = test.experimentalChannel

			eventCh, err := client.WatchAll(nil, make(chan struct{}))
			require.NoError(t, err)

			if len(k8sObjects) > 0 || len(gwObjects) > 0 {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				EntryPoints:         test.entryPoints,
				ExperimentalChannel: test.experimentalChannel,
				client:              client,
			}

			conf := p.loadConfigurationFromGateways(t.Context())
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
			desc:  "Empty because ReferenceGrant spec.from does not match secret",
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
			desc:  "Empty because ReferenceGrant spec.to does not match secret",
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
						"tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb": {
							EntryPoints: []string{"tls"},
							Service:     "tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr",
							Rule:        "HostSNI(`*`)",
							RuleSyntax:  "default",
							TLS:         &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"tcproute-default-tcp-app-1-gw-default-my-gateway-ep-tls-0-e3b0c44298fc1c149afb-wrr": {
							Weighted: &dynamic.TCPWeightedRoundRobin{
								Services: []dynamic.TCPWRRService{{
									Name:   "default-whoamitcp-9000",
									Weight: ptr.To(1),
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
			desc:  "Empty because ReferenceGrant for Service is missing",
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
			desc:  "Empty because ReferenceGrant spec.from does not match service",
			paths: []string{"services.yml", "referencegrant/for_service_not_matching_from.yml"},
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
			desc:  "Empty because ReferenceGrant spec.to does not match service",
			paths: []string{"services.yml", "referencegrant/for_service_not_matching_to.yml"},
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
			desc:  "For Service",
			paths: []string{"services.yml", "referencegrant/for_service.yml"},
			entryPoints: map[string]Entrypoint{
				"http": {Address: ":80"},
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
						"httproute-default-http-app-1-gw-default-my-gateway-ep-http-0-d40286ed9f4652ca2108": {
							EntryPoints: []string{"http"},
							Rule:        "Host(`foo.example.com`) && PathPrefix(`/`)",
							Service:     "httproute-default-http-app-1-gw-default-my-gateway-ep-http-0-d40286ed9f4652ca2108-wrr",
							RuleSyntax:  "default",
							Priority:    17,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"httproute-default-http-app-1-gw-default-my-gateway-ep-http-0-d40286ed9f4652ca2108-wrr": {
							Weighted: &dynamic.WeightedRoundRobin{
								Services: []dynamic.WRRService{
									{
										Name:   "default-whoami-bar-http-80",
										Weight: ptr.To(1),
										Status: ptr.To(500),
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			if test.expected == nil {
				return
			}

			k8sObjects, gwObjects := readResources(t, test.paths)

			kubeClient := kubefake.NewSimpleClientset(k8sObjects...)
			gwClient := newGatewaySimpleClientSet(t, gwObjects...)

			client := newClientImpl(kubeClient, gwClient)
			client.experimentalChannel = test.experimentalChannel

			eventCh, err := client.WatchAll(nil, make(chan struct{}))
			require.NoError(t, err)

			if len(k8sObjects) > 0 || len(gwObjects) > 0 {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				EntryPoints:         test.entryPoints,
				ExperimentalChannel: test.experimentalChannel,
				client:              client,
			}

			conf := p.loadConfigurationFromGateways(t.Context())
			assert.Equal(t, test.expected, conf)
		})
	}
}

func Test_matchingGatewayListener(t *testing.T) {
	testCases := []struct {
		desc           string
		gwListeners    []gatewayListener
		parentRefs     []gatev1.ParentReference
		routeNamespace string
		wantLen        int
	}{
		{
			desc: "Unsupported group",
			gwListeners: []gatewayListener{{
				Name:        "foo",
				GWName:      "gateway",
				GWNamespace: "default",
			}},
			parentRefs: []gatev1.ParentReference{{
				Group: ptr.To(gatev1.Group("foo")),
			}},
			wantLen: 0,
		},
		{
			desc: "Unsupported kind",
			gwListeners: []gatewayListener{{
				Name:        "foo",
				GWName:      "gateway",
				GWNamespace: "default",
			}},
			parentRefs: []gatev1.ParentReference{{
				Group: ptr.To(gatev1.Group(gatev1.GroupName)),
				Kind:  ptr.To(gatev1.Kind("foo")),
			}},
			wantLen: 0,
		},
		{
			desc: "Namespace does not match the listener",
			gwListeners: []gatewayListener{{
				Name:        "foo",
				GWName:      "gateway",
				GWNamespace: "default",
			}},
			parentRefs: []gatev1.ParentReference{{
				Namespace: ptr.To(gatev1.Namespace("foo")),
				Group:     ptr.To(gatev1.Group(gatev1.GroupName)),
				Kind:      ptr.To(gatev1.Kind("Gateway")),
			}},
			wantLen: 0,
		},
		{
			desc: "Route namespace defaulting does not match the listener",
			gwListeners: []gatewayListener{{
				Name:        "foo",
				GWName:      "gateway",
				GWNamespace: "default",
			}},
			routeNamespace: "foo",
			parentRefs: []gatev1.ParentReference{{
				Group: ptr.To(gatev1.Group(gatev1.GroupName)),
				Kind:  ptr.To(gatev1.Kind("Gateway")),
			}},
			wantLen: 0,
		},
		{
			desc: "Name does not match the listener",
			gwListeners: []gatewayListener{{
				GWName:      "gateway",
				GWNamespace: "default",
			}},
			parentRefs: []gatev1.ParentReference{{
				Namespace: ptr.To(gatev1.Namespace("default")),
				Name:      "foo",
				Group:     ptr.To(gatev1.Group(gatev1.GroupName)),
				Kind:      ptr.To(gatev1.Kind("Gateway")),
			}},
			wantLen: 0,
		},
		{
			desc: "Match",
			gwListeners: []gatewayListener{{
				GWName:      "gateway",
				GWNamespace: "default",
			}},
			parentRefs: []gatev1.ParentReference{{
				Name:      "gateway",
				Namespace: ptr.To(gatev1.Namespace("default")),
				Group:     ptr.To(gatev1.Group(gatev1.GroupName)),
				Kind:      ptr.To(gatev1.Kind("Gateway")),
			}},
			wantLen: 1,
		},
		{
			desc: "Match with route namespace defaulting",
			gwListeners: []gatewayListener{{
				GWName:      "gateway",
				GWNamespace: "default",
			}},
			routeNamespace: "default",
			parentRefs: []gatev1.ParentReference{{
				Name:  "gateway",
				Group: ptr.To(gatev1.Group(gatev1.GroupName)),
				Kind:  ptr.To(gatev1.Kind("Gateway")),
			}},
			wantLen: 1,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			listeners := matchingGatewayListeners(test.gwListeners, test.routeNamespace, test.parentRefs)
			assert.Len(t, listeners, test.wantLen)
		})
	}
}

func Test_matchListener(t *testing.T) {
	testCases := []struct {
		desc       string
		gwListener gatewayListener
		parentRef  gatev1.ParentReference
		wantMatch  bool
	}{
		{
			desc: "Section do not match",
			gwListener: gatewayListener{
				Name: "foo",
				Port: gatev1.PortNumber(80),
			},
			parentRef: gatev1.ParentReference{
				SectionName: ptr.To(gatev1.SectionName("bar")),
				Port:        ptr.To(gatev1.PortNumber(80)),
			},
		},
		{
			desc: "Section matches",
			gwListener: gatewayListener{
				Name: "foo",
				Port: gatev1.PortNumber(80),
			},
			parentRef: gatev1.ParentReference{
				SectionName: ptr.To(gatev1.SectionName("foo")),
				Port:        ptr.To(gatev1.PortNumber(80)),
			},
			wantMatch: true,
		},
		{
			desc: "Port do not match",
			gwListener: gatewayListener{
				Name: "foo",
				Port: gatev1.PortNumber(90),
			},
			parentRef: gatev1.ParentReference{
				SectionName: ptr.To(gatev1.SectionName("foo")),
				Port:        ptr.To(gatev1.PortNumber(80)),
			},
		},
		{
			desc: "Port matches",
			gwListener: gatewayListener{
				Name: "foo",
				Port: gatev1.PortNumber(80),
			},
			parentRef: gatev1.ParentReference{
				SectionName: ptr.To(gatev1.SectionName("foo")),
				Port:        ptr.To(gatev1.PortNumber(80)),
			},
			wantMatch: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			gotMatch := matchListener(test.gwListener, test.parentRef)
			assert.Equal(t, test.wantMatch, gotMatch)
		})
	}
}

func Test_allowRoute(t *testing.T) {
	testCases := []struct {
		desc           string
		gwListener     gatewayListener
		routeNamespace string
		routeKind      string
		wantAllow      bool
	}{
		{
			desc: "Not allowed Kind",
			gwListener: gatewayListener{
				Name:        "foo",
				GWName:      "gateway",
				GWNamespace: "default",
				AllowedRouteKinds: []string{
					"foo",
					"bar",
				},
			},
			routeKind: "baz",
			wantAllow: false,
		},
		{
			desc: "Allowed Kind",
			gwListener: gatewayListener{
				Name:        "foo",
				GWName:      "gateway",
				GWNamespace: "default",
				AllowedRouteKinds: []string{
					"foo",
					"bar",
				},
				AllowedNamespaces: []string{
					corev1.NamespaceAll,
				},
			},
			routeKind: "bar",
			wantAllow: true,
		},
		{
			desc: "Not allowed namespace",
			gwListener: gatewayListener{
				Name:        "foo",
				GWName:      "gateway",
				GWNamespace: "default",
				AllowedRouteKinds: []string{
					"foo",
				},
				AllowedNamespaces: []string{
					"foo",
					"bar",
				},
			},
			routeKind:      "foo",
			routeNamespace: "baz",
			wantAllow:      false,
		},
		{
			desc: "Allowed namespace",
			gwListener: gatewayListener{
				Name:        "foo",
				GWName:      "gateway",
				GWNamespace: "default",
				AllowedRouteKinds: []string{
					"foo",
				},
				AllowedNamespaces: []string{
					"foo",
					"bar",
				},
			},
			routeKind:      "foo",
			routeNamespace: "foo",
			wantAllow:      true,
		},
		{
			desc: "Allowed namespace",
			gwListener: gatewayListener{
				Name:        "foo",
				GWName:      "gateway",
				GWNamespace: "default",
				AllowedRouteKinds: []string{
					"foo",
				},
				AllowedNamespaces: []string{
					corev1.NamespaceAll,
					"bar",
				},
			},
			routeKind:      "foo",
			routeNamespace: "foo",
			wantAllow:      true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			gotAllow := allowRoute(test.gwListener, test.routeNamespace, test.routeKind)
			assert.Equal(t, test.wantAllow, gotAllow)
		})
	}
}

func Test_findMatchingHostnames(t *testing.T) {
	testCases := []struct {
		desc             string
		listenerHostname *gatev1.Hostname
		routeHostnames   []gatev1.Hostname
		want             []gatev1.Hostname
		wantOk           bool
	}{
		{
			desc:   "Empty",
			wantOk: true,
		},
		{
			desc:             "Only listener hostname",
			listenerHostname: ptr.To(gatev1.Hostname("foo.com")),
			want:             []gatev1.Hostname{"foo.com"},
			wantOk:           true,
		},
		{
			desc:           "Only Route hostname",
			routeHostnames: []gatev1.Hostname{"foo.com"},
			want:           []gatev1.Hostname{"foo.com"},
			wantOk:         true,
		},
		{
			desc:             "Matching hostname",
			listenerHostname: ptr.To(gatev1.Hostname("foo.com")),
			routeHostnames:   []gatev1.Hostname{"foo.com"},
			want:             []gatev1.Hostname{"foo.com"},
			wantOk:           true,
		},
		{
			desc:             "Matching hostname with wildcard",
			listenerHostname: ptr.To(gatev1.Hostname("*.foo.com")),
			routeHostnames:   []gatev1.Hostname{"*.foo.com"},
			want:             []gatev1.Hostname{"*.foo.com"},
			wantOk:           true,
		},
		{
			desc:             "Matching subdomain with listener wildcard",
			listenerHostname: ptr.To(gatev1.Hostname("*.foo.com")),
			routeHostnames:   []gatev1.Hostname{"bar.foo.com"},
			want:             []gatev1.Hostname{"bar.foo.com"},
			wantOk:           true,
		},
		{
			desc:             "Matching subsubdomain with listener wildcard",
			listenerHostname: ptr.To(gatev1.Hostname("*.foo.com")),
			routeHostnames:   []gatev1.Hostname{"baz.bar.foo.com"},
			want:             []gatev1.Hostname{"baz.bar.foo.com"},
			wantOk:           true,
		},
		{
			desc:             "Matching subdomain with route hostname wildcard",
			listenerHostname: ptr.To(gatev1.Hostname("bar.foo.com")),
			routeHostnames:   []gatev1.Hostname{"*.foo.com"},
			want:             []gatev1.Hostname{"bar.foo.com"},
			wantOk:           true,
		},
		{
			desc:             "Matching subsubdomain with route hostname wildcard",
			listenerHostname: ptr.To(gatev1.Hostname("baz.bar.foo.com")),
			routeHostnames:   []gatev1.Hostname{"*.foo.com"},
			want:             []gatev1.Hostname{"baz.bar.foo.com"},
			wantOk:           true,
		},
		{
			desc:             "Non matching root domain with listener wildcard",
			listenerHostname: ptr.To(gatev1.Hostname("*.foo.com")),
			routeHostnames:   []gatev1.Hostname{"foo.com"},
		},
		{
			desc:             "Non matching root domain with route hostname wildcard",
			listenerHostname: ptr.To(gatev1.Hostname("foo.com")),
			routeHostnames:   []gatev1.Hostname{"*.foo.com"},
		},
		{
			desc:             "Multiple route hostnames with one matching route hostname",
			listenerHostname: ptr.To(gatev1.Hostname("*.foo.com")),
			routeHostnames:   []gatev1.Hostname{"bar.com", "test.foo.com", "test.buz.com"},
			want:             []gatev1.Hostname{"test.foo.com"},
			wantOk:           true,
		},
		{
			desc:             "Multiple route hostnames with non matching route hostname",
			listenerHostname: ptr.To(gatev1.Hostname("*.fuz.com")),
			routeHostnames:   []gatev1.Hostname{"bar.com", "test.foo.com", "test.buz.com"},
		},
		{
			desc:             "Multiple route hostnames with multiple matching route hostnames",
			listenerHostname: ptr.To(gatev1.Hostname("*.foo.com")),
			routeHostnames:   []gatev1.Hostname{"toto.foo.com", "test.foo.com", "test.buz.com"},
			want:             []gatev1.Hostname{"toto.foo.com", "test.foo.com"},
			wantOk:           true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got, ok := findMatchingHostnames(test.listenerHostname, test.routeHostnames)
			assert.Equal(t, test.wantOk, ok)
			assert.Equal(t, test.want, got)
		})
	}
}

func Test_allowedRouteKinds(t *testing.T) {
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
				{Kind: kindTLSRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
			},
			wantKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
			},
		},
		{
			desc: "AllowedRoutes with unsupported Group",
			listener: gatev1.Listener{
				AllowedRoutes: &gatev1.AllowedRoutes{
					Kinds: []gatev1.RouteGroupKind{{
						Kind: kindTLSRoute, Group: ptr.To(gatev1.Group("foo")),
					}},
				},
			},
			supportedRouteKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
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
				{Kind: kindTLSRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
			},
			wantErr: true,
		},
		{
			desc: "AllowedRoutes with unsupported Kind",
			listener: gatev1.Listener{
				AllowedRoutes: &gatev1.AllowedRoutes{
					Kinds: []gatev1.RouteGroupKind{{
						Kind: "foo", Group: ptr.To(gatev1.Group(gatev1.GroupName)),
					}},
				},
			},
			supportedRouteKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
			},
			wantErr: true,
		},
		{
			desc: "Supported AllowedRoutes",
			listener: gatev1.Listener{
				AllowedRoutes: &gatev1.AllowedRoutes{
					Kinds: []gatev1.RouteGroupKind{{
						Kind: kindTLSRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName)),
					}},
				},
			},
			supportedRouteKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
			},
			wantKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
			},
		},
		{
			desc: "Supported AllowedRoutes with duplicates",
			listener: gatev1.Listener{
				AllowedRoutes: &gatev1.AllowedRoutes{
					Kinds: []gatev1.RouteGroupKind{
						{Kind: kindTLSRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
						{Kind: kindTCPRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
						{Kind: kindTLSRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
						{Kind: kindTCPRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
					},
				},
			},
			supportedRouteKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
				{Kind: kindTCPRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
			},
			wantKinds: []gatev1.RouteGroupKind{
				{Kind: kindTLSRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
				{Kind: kindTCPRoute, Group: ptr.To(gatev1.Group(gatev1.GroupName))},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got, conditions := allowedRouteKinds(&gatev1.Gateway{}, test.listener, test.supportedRouteKinds)
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
				Hostname: ptr.To(gatev1.Hostname("www.example.com")),
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
							Name:  ptr.To(gatev1.ObjectName("correct-name")),
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
							Name:  ptr.To(gatev1.ObjectName("correct-name")),
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
							Name:  ptr.To(gatev1.ObjectName("correct-name")),
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
							Name:  ptr.To(gatev1.ObjectName("correct-name")),
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
							Name:  ptr.To(gatev1.ObjectName("wrong-name")),
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

			k8sObjects, gwObjects := readResources(t, test.paths)

			kubeClient := kubefake.NewSimpleClientset(k8sObjects...)
			gwClient := newGatewaySimpleClientSet(t, gwObjects...)

			client := newClientImpl(kubeClient, gwClient)

			eventCh, err := client.WatchAll(nil, make(chan struct{}))
			require.NoError(t, err)

			if len(k8sObjects) > 0 || len(gwObjects) > 0 {
				// just wait for the first event
				<-eventCh
			}

			p := Provider{
				StatusAddress: test.statusAddress,
				client:        client,
			}

			got, err := p.gatewayAddresses()
			test.wantErr(t, err)

			assert.Equal(t, test.want, got)
		})
	}
}

func Test_upsertRouteConditionResolvedRefs(t *testing.T) {
	testCases := []struct {
		desc           string
		conditions     []metav1.Condition
		condition      metav1.Condition
		wantConditions []metav1.Condition
	}{
		{
			desc: "True to False",
			conditions: []metav1.Condition{
				{
					Type:    "foo",
					Status:  "bar",
					Reason:  "baz",
					Message: "foobarbaz",
				},
				{
					Type:    string(gatev1.RouteConditionResolvedRefs),
					Status:  metav1.ConditionTrue,
					Reason:  "foo",
					Message: "foo",
				},
			},
			condition: metav1.Condition{
				Type:    string(gatev1.RouteConditionResolvedRefs),
				Status:  metav1.ConditionFalse,
				Reason:  "bar",
				Message: "bar",
			},
			wantConditions: []metav1.Condition{
				{
					Type:    "foo",
					Status:  "bar",
					Reason:  "baz",
					Message: "foobarbaz",
				},
				{
					Type:    string(gatev1.RouteConditionResolvedRefs),
					Status:  metav1.ConditionFalse,
					Reason:  "bar",
					Message: "bar",
				},
			},
		},
		{
			desc: "False to False",
			conditions: []metav1.Condition{
				{
					Type:    "foo",
					Status:  "bar",
					Reason:  "baz",
					Message: "foobarbaz",
				},
				{
					Type:    string(gatev1.RouteConditionResolvedRefs),
					Status:  metav1.ConditionFalse,
					Reason:  "foo",
					Message: "foo",
				},
			},
			condition: metav1.Condition{
				Type:    string(gatev1.RouteConditionResolvedRefs),
				Status:  metav1.ConditionFalse,
				Reason:  "bar",
				Message: "bar",
			},
			wantConditions: []metav1.Condition{
				{
					Type:    "foo",
					Status:  "bar",
					Reason:  "baz",
					Message: "foobarbaz",
				},
				{
					Type:    string(gatev1.RouteConditionResolvedRefs),
					Status:  metav1.ConditionFalse,
					Reason:  "bar",
					Message: "bar",
				},
			},
		},
		{
			desc: "False to True: no upsert",
			conditions: []metav1.Condition{
				{
					Type:    "foo",
					Status:  "bar",
					Reason:  "baz",
					Message: "foobarbaz",
				},
				{
					Type:    string(gatev1.RouteConditionResolvedRefs),
					Status:  metav1.ConditionFalse,
					Reason:  "foo",
					Message: "foo",
				},
			},
			condition: metav1.Condition{
				Type:    string(gatev1.RouteConditionResolvedRefs),
				Status:  metav1.ConditionTrue,
				Reason:  "bar",
				Message: "bar",
			},
			wantConditions: []metav1.Condition{
				{
					Type:    "foo",
					Status:  "bar",
					Reason:  "baz",
					Message: "foobarbaz",
				},
				{
					Type:    string(gatev1.RouteConditionResolvedRefs),
					Status:  metav1.ConditionFalse,
					Reason:  "foo",
					Message: "foo",
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := upsertRouteConditionResolvedRefs(test.conditions, test.condition)
			assert.Equal(t, test.wantConditions, got)
		})
	}
}

// We cannot use the gateway-api fake.NewSimpleClientset due to Gateway being pluralized as "gatewaies" instead of "gateways".
func newGatewaySimpleClientSet(t *testing.T, objects ...runtime.Object) *gatefake.Clientset {
	t.Helper()

	client := gatefake.NewSimpleClientset(objects...)
	for _, object := range objects {
		gateway, ok := object.(*gatev1.Gateway)
		if !ok {
			continue
		}

		_, err := client.GatewayV1().Gateways(gateway.Namespace).Create(t.Context(), gateway, metav1.CreateOptions{})
		require.NoError(t, err)
	}

	return client
}

func readResources(t *testing.T, paths []string) ([]runtime.Object, []runtime.Object) {
	t.Helper()

	var k8sObjects []runtime.Object
	var gwObjects []runtime.Object
	for _, path := range paths {
		yamlContent, err := os.ReadFile(filepath.FromSlash("./fixtures/" + path))
		if err != nil {
			panic(err)
		}

		objects := k8s.MustParseYaml(yamlContent)
		for _, obj := range objects {
			switch obj.GetObjectKind().GroupVersionKind().Group {
			case "gateway.networking.k8s.io":
				gwObjects = append(gwObjects, obj)
			default:
				k8sObjects = append(k8sObjects, obj)
			}
		}
	}

	return k8sObjects, gwObjects
}
