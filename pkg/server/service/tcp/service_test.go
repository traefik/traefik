package tcp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/server/provider"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

func TestManager_BuildTCP(t *testing.T) {
	testCases := []struct {
		desc          string
		serviceName   string
		configs       map[string]*runtime.TCPServiceInfo
		stConfigs     map[string]*dynamic.TCPServersTransport
		providerName  string
		expectedError string
	}{
		{
			desc:          "without configuration",
			serviceName:   "test",
			configs:       nil,
			expectedError: `the service "test" does not exist`,
		},
		{
			desc:        "missing lb configuration",
			serviceName: "test",
			configs: map[string]*runtime.TCPServiceInfo{
				"test": {
					TCPService: &dynamic.TCPService{},
				},
			},
			expectedError: `the service "test" does not have any type defined`,
		},
		{
			desc:        "no such host, server is skipped, error is logged",
			serviceName: "test",
			stConfigs:   map[string]*dynamic.TCPServersTransport{"default@internal": {}},
			configs: map[string]*runtime.TCPServiceInfo{
				"test": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{Address: "test:31"},
							},
						},
					},
				},
			},
		},
		{
			desc:        "invalid IP address, server is skipped, error is logged",
			serviceName: "test",
			configs: map[string]*runtime.TCPServiceInfo{
				"test": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{Address: "foobar"},
							},
						},
					},
				},
			},
		},
		{
			desc:        "Simple service name",
			serviceName: "serviceName",
			configs: map[string]*runtime.TCPServiceInfo{
				"serviceName": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{},
					},
				},
			},
		},
		{
			desc:        "Service name with provider",
			serviceName: "serviceName@provider-1",
			configs: map[string]*runtime.TCPServiceInfo{
				"serviceName@provider-1": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{},
					},
				},
			},
		},
		{
			desc:        "Service name with provider in context",
			serviceName: "serviceName",
			configs: map[string]*runtime.TCPServiceInfo{
				"serviceName@provider-1": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{},
					},
				},
			},
			providerName: "provider-1",
		},
		{
			desc:        "Server with correct host:port as address",
			serviceName: "serviceName",
			stConfigs:   map[string]*dynamic.TCPServersTransport{"default@internal": {}},
			configs: map[string]*runtime.TCPServiceInfo{
				"serviceName@provider-1": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "foobar.com:80",
								},
							},
						},
					},
				},
			},
			providerName: "provider-1",
		},
		{
			desc:        "Server with correct ip:port as address",
			serviceName: "serviceName",
			stConfigs:   map[string]*dynamic.TCPServersTransport{"default@internal": {}},
			configs: map[string]*runtime.TCPServiceInfo{
				"serviceName@provider-1": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "192.168.0.12:80",
								},
							},
						},
					},
				},
			},
			providerName: "provider-1",
		},
		{
			desc:        "empty server address, server is skipped, error is logged",
			serviceName: "serviceName",
			configs: map[string]*runtime.TCPServiceInfo{
				"serviceName@provider-1": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "",
								},
							},
						},
					},
				},
			},
			providerName: "provider-1",
		},
		{
			desc:        "missing port in address with hostname, server is skipped, error is logged",
			serviceName: "serviceName",
			configs: map[string]*runtime.TCPServiceInfo{
				"serviceName@provider-1": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "foobar.com",
								},
							},
						},
					},
				},
			},
			providerName: "provider-1",
		},
		{
			desc:        "missing port in address with ip, server is skipped, error is logged",
			serviceName: "serviceName",
			configs: map[string]*runtime.TCPServiceInfo{
				"serviceName@provider-1": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "192.168.0.12",
								},
							},
						},
					},
				},
			},
			providerName: "provider-1",
		},
		{
			desc:        "user defined serversTransport reference",
			serviceName: "serviceName",
			stConfigs:   map[string]*dynamic.TCPServersTransport{"myServersTransport@provider-1": {}},
			configs: map[string]*runtime.TCPServiceInfo{
				"serviceName@provider-1": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "192.168.0.12:80",
								},
							},
							ServersTransport: "myServersTransport@provider-1",
						},
					},
				},
			},
			providerName: "provider-1",
		},
		{
			desc:        "user defined serversTransport reference not found",
			serviceName: "serviceName",
			configs: map[string]*runtime.TCPServiceInfo{
				"serviceName@provider-1": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "192.168.0.12:80",
								},
							},
							ServersTransport: "myServersTransport@provider-1",
						},
					},
				},
			},
			providerName:  "provider-1",
			expectedError: "TCP dialer not found myServersTransport@provider-1",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dialerManager := tcp.NewDialerManager(nil)
			if test.stConfigs != nil {
				dialerManager.Update(test.stConfigs)
			}

			manager := NewManager(&runtime.Configuration{
				TCPServices: test.configs,
			}, dialerManager)

			ctx := context.Background()
			if len(test.providerName) > 0 {
				ctx = provider.AddInContext(ctx, "foobar@"+test.providerName)
			}

			handler, err := manager.BuildTCP(ctx, test.serviceName)

			if test.expectedError != "" {
				assert.EqualError(t, err, test.expectedError)
				require.Nil(t, handler)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, handler)
			}
		})
	}
}
