package udp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/server/provider"
)

func TestManager_BuildUDP(t *testing.T) {
	testCases := []struct {
		desc          string
		serviceName   string
		configs       map[string]*runtime.UDPServiceInfo
		providerName  string
		expectedError string
	}{
		{
			desc:          "without configuration",
			serviceName:   "test",
			configs:       nil,
			expectedError: `the UDP service "test" does not exist`,
		},
		{
			desc:        "missing lb configuration",
			serviceName: "test",
			configs: map[string]*runtime.UDPServiceInfo{
				"test": {
					UDPService: &dynamic.UDPService{},
				},
			},
			expectedError: `the UDP service "test" does not have any type defined`,
		},
		{
			desc:        "no such host, server is skipped, error is logged",
			serviceName: "test",
			configs: map[string]*runtime.UDPServiceInfo{
				"test": {
					UDPService: &dynamic.UDPService{
						LoadBalancer: &dynamic.UDPServersLoadBalancer{
							Servers: []dynamic.UDPServer{
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
			configs: map[string]*runtime.UDPServiceInfo{
				"test": {
					UDPService: &dynamic.UDPService{
						LoadBalancer: &dynamic.UDPServersLoadBalancer{
							Servers: []dynamic.UDPServer{
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
			configs: map[string]*runtime.UDPServiceInfo{
				"serviceName": {
					UDPService: &dynamic.UDPService{
						LoadBalancer: &dynamic.UDPServersLoadBalancer{},
					},
				},
			},
		},
		{
			desc:        "Service name with provider",
			serviceName: "serviceName@provider-1",
			configs: map[string]*runtime.UDPServiceInfo{
				"serviceName@provider-1": {
					UDPService: &dynamic.UDPService{
						LoadBalancer: &dynamic.UDPServersLoadBalancer{},
					},
				},
			},
		},
		{
			desc:        "Service name with provider in context",
			serviceName: "serviceName",
			configs: map[string]*runtime.UDPServiceInfo{
				"serviceName@provider-1": {
					UDPService: &dynamic.UDPService{
						LoadBalancer: &dynamic.UDPServersLoadBalancer{},
					},
				},
			},
			providerName: "provider-1",
		},
		{
			desc:        "Server with correct host:port as address",
			serviceName: "serviceName",
			configs: map[string]*runtime.UDPServiceInfo{
				"serviceName@provider-1": {
					UDPService: &dynamic.UDPService{
						LoadBalancer: &dynamic.UDPServersLoadBalancer{
							Servers: []dynamic.UDPServer{
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
			configs: map[string]*runtime.UDPServiceInfo{
				"serviceName@provider-1": {
					UDPService: &dynamic.UDPService{
						LoadBalancer: &dynamic.UDPServersLoadBalancer{
							Servers: []dynamic.UDPServer{
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
			desc:        "missing port in address with hostname, server is skipped, error is logged",
			serviceName: "serviceName",
			configs: map[string]*runtime.UDPServiceInfo{
				"serviceName@provider-1": {
					UDPService: &dynamic.UDPService{
						LoadBalancer: &dynamic.UDPServersLoadBalancer{
							Servers: []dynamic.UDPServer{
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
			configs: map[string]*runtime.UDPServiceInfo{
				"serviceName@provider-1": {
					UDPService: &dynamic.UDPService{
						LoadBalancer: &dynamic.UDPServersLoadBalancer{
							Servers: []dynamic.UDPServer{
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			manager := NewManager(&runtime.Configuration{
				UDPServices: test.configs,
			})

			ctx := t.Context()
			if len(test.providerName) > 0 {
				ctx = provider.AddInContext(ctx, "foobar@"+test.providerName)
			}

			handler, err := manager.BuildUDP(ctx, test.serviceName)

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
