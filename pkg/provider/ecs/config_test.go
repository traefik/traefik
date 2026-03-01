package ecs

import (
	"testing"
	"time"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

func pointer[T any](v T) *T { return &v }

func TestDefaultRule(t *testing.T) {
	testCases := []struct {
		desc        string
		instances   []ecsInstance
		defaultRule string
		expected    *dynamic.Configuration
	}{
		{
			desc: "default rule with no variable",
			instances: []ecsInstance{
				instance(
					name("Test"),
					id("1"),
					labels(map[string]string{}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("10.0.0.1"),
						mPorts(
							mPort(0, 1337, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			defaultRule: "Host(`foo.bar`)",
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`foo.bar`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://10.0.0.1:1337",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "default rule with service name",
			instances: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			defaultRule: "Host(`{{ .Name }}.foo.bar`)",
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.foo.bar`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "default rule with label",
			instances: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.domain": "foo.bar",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			defaultRule: `Host("{{ .Name }}.{{ index .Labels "traefik.domain" }}")`,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        `Host("Test.foo.bar")`,
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "invalid rule",
			instances: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			defaultRule: `Host("{{ .Toto }}")`,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "undefined rule",
			instances: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			defaultRule: ``,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "default template rule",
			instances: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			defaultRule: DefaultTemplateRule,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{
				ExposedByDefault: true,
				DefaultRule:      test.defaultRule,
				defaultRuleTpl:   nil,
			}

			err := p.Init()
			require.NoError(t, err)

			for i := range len(test.instances) {
				var err error
				test.instances[i].ExtraConf, err = p.getConfiguration(test.instances[i])
				require.NoError(t, err)
			}

			configuration := p.buildConfiguration(t.Context(), test.instances)

			assert.Equal(t, test.expected, configuration)
		})
	}
}

func Test_buildConfiguration(t *testing.T) {
	testCases := []struct {
		desc        string
		containers  []ecsInstance
		constraints string
		expected    *dynamic.Configuration
	}{
		{
			desc: "invalid HTTP service definition",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.services.test": "",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "invalid TCP service definition",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.tcp.services.test": "",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "invalid UDP service definition",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.udp.services.test": "",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container no label",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "two containers no label",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test2"),
					labels(map[string]string{}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.2"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
						"Test2": {
							Service:     "Test2",
							Rule:        "Host(`Test2.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"Test2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.2:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "two containers with same service name no label",
			containers: []ecsInstance{
				instance(
					id("1"),
					name("Test"),
					labels(map[string]string{}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					id("2"),
					name("Test"),
					labels(map[string]string{}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.2"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with label (not on server)",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with labels",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
						"traefik.http.routers.Router1.rule":                          "Host(`foo.com`)",
						"traefik.http.routers.Router1.service":                       "Service1",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Router1": {
							Service: "Service1",
							Rule:    "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with rule label",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
					},
					Routers: map[string]*dynamic.Router{
						"Router1": {
							Service: "Test",
							Rule:    "Host(`foo.com`)",
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with rule label and one service",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.routers.Router1.rule":                          "Host(`foo.com`)",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Router1": {
							Service: "Service1",
							Rule:    "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with rule label and two services",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.routers.Router1.rule":                          "Host(`foo.com`)",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
						"traefik.http.services.Service2.loadbalancer.passhostheader": "true",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"Service2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one router, one specified but undefined service -> specified one is assigned, but automatic is created instead",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.routers.Router1.rule":    "Host(`foo.com`)",
						"traefik.http.routers.Router1.service": "Service1",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Router1": {
							Service: "Service1",
							Rule:    "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "two containers with same service name and different passhostheader",
			containers: []ecsInstance{
				instance(
					name("Test"),
					id("1"),
					labels(map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test"),
					id("2"),
					labels(map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "false",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "three containers with same service name and different passhostheader",
			containers: []ecsInstance{
				instance(
					name("Test"),
					id("1"),
					labels(map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "false",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test"),
					id("2"),
					labels(map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test"),
					id("3"),
					labels(map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "two containers with same service name and same LB methods",
			containers: []ecsInstance{
				instance(
					name("Test"),
					id("1"),
					labels(map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test"),
					id("2"),
					labels(map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.2"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with InFlightReq in label (default value)",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"Middleware1": {
							InFlightReq: &dynamic.InFlightReq{
								Amount: 42,
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "two containers with two identical middlewares",
			containers: []ecsInstance{
				instance(
					name("Test"),
					id("1"),
					labels(map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test"),
					id("2"),
					labels(map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.2"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"Middleware1": {
							InFlightReq: &dynamic.InFlightReq{
								Amount: 42,
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "two containers with two different middlewares with same name",
			containers: []ecsInstance{
				instance(
					name("Test"),
					id("1"),
					labels(map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test"),
					id("2"),
					labels(map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "41",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.2"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "three containers with different middlewares with same name",
			containers: []ecsInstance{
				instance(
					name("Test"),
					id("1"),
					labels(map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test"),
					id("2"),
					labels(map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "41",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.2"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test"),
					id("3"),
					labels(map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "40",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.3"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
									},
									{
										URL: "http://127.0.0.3:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "two containers with two different routers with same name",
			containers: []ecsInstance{
				instance(
					name("Test"),
					id("1"),
					labels(map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test"),
					id("2"),
					labels(map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`bar.com`)",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.2"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "three containers with different routers with same name",
			containers: []ecsInstance{
				instance(
					name("Test"),
					id("1"),
					labels(map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test"),
					id("2"),
					labels(map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`bar.com`)",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.2"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test"),
					id("3"),
					labels(map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foobar.com`)",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.3"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
									},
									{
										URL: "http://127.0.0.3:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "two containers with two identical routers",
			containers: []ecsInstance{
				instance(
					name("Test"),
					id("1"),
					labels(map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),

				instance(
					name("Test"),
					id("2"),
					labels(map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.2"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Router1": {
							Service: "Test",
							Rule:    "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "two containers with two identical router rules and different service names",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test2"),
					labels(map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.2"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"Test2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.2:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with bad label",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.wrong.label": "42",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with label port",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.scheme": "h2c",
						"traefik.http.services.Service1.LoadBalancer.server.port":   "80",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 8080, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "h2c://127.0.0.1:8080",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with label url",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.url": "http://1.2.3.4:5678",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 8080, "tcp"),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://1.2.3.4:5678",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with label url and preserve path",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.url":          "http://1.2.3.4:5678",
						"traefik.http.services.Service1.LoadBalancer.server.preservepath": "true",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 8080, "tcp"),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL:          "http://1.2.3.4:5678",
										PreservePath: true,
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with label url and port",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.url":  "http://1.2.3.4:5678",
						"traefik.http.services.Service1.LoadBalancer.server.port": "1234",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 8080, "tcp"),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with label url and scheme",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.url":    "http://1.2.3.4:5678",
						"traefik.http.services.Service1.LoadBalancer.server.scheme": "https",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 8080, "tcp"),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with label port not exposed by container",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.scheme": "h2c",
						"traefik.http.services.Service1.LoadBalancer.server.port":   "8040",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 8080, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "h2c://127.0.0.1:8040",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with label and multiple ports",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.routers.Test.rule":                          "Host(`Test.traefik.wtf`)",
						"traefik.http.routers.Test.service":                       "Service1",
						"traefik.http.services.Service1.LoadBalancer.server.port": "4445",
						"traefik.http.routers.Test2.rule":                         "Host(`Test.traefik.local`)",
						"traefik.http.routers.Test2.service":                      "Service2",
						"traefik.http.services.Service2.LoadBalancer.server.port": "4444",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(4444, 32123, ecstypes.TransportProtocolTcp),
							mPort(4445, 32124, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
						"Test2": {
							Service: "Service2",
							Rule:    "Host(`Test.traefik.local`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:32124",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"Service2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:32123",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with label port on two services",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.port": "",
						"traefik.http.services.Service2.LoadBalancer.server.port": "8080",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"Service2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:8080",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container without port",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container without port with middleware",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with traefik.enable false",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.enable": "false",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container not healthy",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.enable": "false",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mHealthStatus(ecstypes.HealthStatusUnhealthy),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container not running",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.enable": "false",
					}),
					iMachine(
						mState(ec2types.InstanceStateNamePending),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with non matching constraints",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.tags": "foo",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			constraints: `Label("traefik.tags", "bar")`,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with matching constraints",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.tags": "foo",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			constraints: `Label("traefik.tags", "foo")`,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "Middlewares used in router",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.http.middlewares.Middleware1.basicauth.users": "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						"traefik.http.routers.Test.middlewares":                "Middleware1",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							Middlewares: []string{"Middleware1"},
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"Middleware1": {
							BasicAuth: &dynamic.BasicAuth{
								Users: []string{
									"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
									"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
								},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "Middlewares used in TCP router",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.tcp.routers.Test.rule":                               "HostSNI(`foo.bar`)",
						"traefik.tcp.middlewares.Middleware1.ipallowlist.sourcerange": "foobar, fiibar",
						"traefik.tcp.routers.Test.middlewares":                        "Middleware1",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"Test": {
							Service:     "Test",
							Rule:        "HostSNI(`foo.bar`)",
							Middlewares: []string{"Middleware1"},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"Middleware1": {
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"foobar", "fiibar"},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"Test": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:80",
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "tcp with label",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.tcp.routers.foo.rule": "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls":  "true",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {
							Service: "Test",
							Rule:    "HostSNI(`foo.bar`)",
							TLS:     &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"Test": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:80",
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "udp with label",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.udp.routers.foo.entrypoints": "mydns",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolUdp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"foo": {
							Service:     "Test",
							EntryPoints: []string{"mydns"},
						},
					},
					Services: map[string]*dynamic.UDPService{
						"Test": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "127.0.0.1:80",
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "tcp with label without rule",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.tcp.routers.foo.tls": "true",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"Test": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:80",
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "tcp with label and port",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.tcp.routers.foo.rule":                      "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls.options":               "foo",
						"traefik.tcp.services.foo.loadbalancer.server.port": "80",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 8080, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {
							Service: "foo",
							Rule:    "HostSNI(`foo.bar`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "foo",
							},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"foo": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:8080",
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "udp with label and port",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.udp.routers.foo.entrypoints":               "mydns",
						"traefik.udp.services.foo.loadbalancer.server.port": "80",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 8080, ecstypes.TransportProtocolUdp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"foo": {
							Service:     "foo",
							EntryPoints: []string{"mydns"},
						},
					},
					Services: map[string]*dynamic.UDPService{
						"foo": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "127.0.0.1:8080",
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "udp with label and port and http service",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.udp.routers.foo.entrypoints":                        "mydns",
						"traefik.udp.services.foo.loadbalancer.server.port":          "8080",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
				instance(
					name("Test"),
					id("2"),
					labels(map[string]string{
						"traefik.udp.routers.foo.entrypoints":                        "mydns",
						"traefik.udp.services.foo.loadbalancer.server.port":          "8080",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.2"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"foo": {
							Service:     "foo",
							EntryPoints: []string{"mydns"},
						},
					},
					Services: map[string]*dynamic.UDPService{
						"foo": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "127.0.0.1:8080",
									},
									{
										Address: "127.0.0.2:8080",
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
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
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
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "udp with label for tcp service",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.udp.services.foo.loadbalancer.server.port": "8080",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{
						"foo": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "127.0.0.1:8080",
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			// TODO: replace or delete?
			desc: "tcp with label for tcp service, with termination delay",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.tcp.services.foo.loadbalancer.server.port": "80",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(80, 8080, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"foo": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:8080",
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with default generated certificate",
			containers: []ecsInstance{
				instance(
					name("Test"),
					labels(map[string]string{
						"traefik.tls.stores.default.defaultgeneratedcert.resolver":    "foobar",
						"traefik.tls.stores.default.defaultgeneratedcert.domain.main": "foobar",
						"traefik.tls.stores.default.defaultgeneratedcert.domain.sans": "foobar, fiibar",
					}),
					iMachine(
						mState(ec2types.InstanceStateNameRunning),
						mPrivateIP("127.0.0.1"),
						mPorts(
							mPort(0, 80, ecstypes.TransportProtocolTcp),
						),
					),
				),
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
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
					Stores: map[string]tls.Store{
						"default": {
							DefaultGeneratedCert: &tls.GeneratedCert{
								Resolver: "foobar",
								Domain: &types.Domain{
									Main: "foobar",
									SANs: []string{"foobar", "fiibar"},
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

			p := Provider{
				ExposedByDefault: true,
				DefaultRule:      "Host(`{{ normalize .Name }}.traefik.wtf`)",
			}
			p.Constraints = test.constraints

			err := p.Init()
			require.NoError(t, err)

			for i := range len(test.containers) {
				var err error
				test.containers[i].ExtraConf, err = p.getConfiguration(test.containers[i])
				require.NoError(t, err)
			}

			configuration := p.buildConfiguration(t.Context(), test.containers)

			assert.Equal(t, test.expected, configuration)
		})
	}
}
