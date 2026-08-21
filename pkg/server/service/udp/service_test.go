package udp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/server/provider"
)

func TestManager_BuildUDP(t *testing.T) {
	negativeWeight := -2

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
			expectedError: `the udp service "test" does not exist`,
		},
		{
			desc:        "missing lb configuration",
			serviceName: "test",
			configs: map[string]*runtime.UDPServiceInfo{
				"test": {
					UDPService: &dynamic.UDPService{},
				},
			},
			expectedError: `the udp service "test" does not have any type defined`,
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
		{
			desc:        "negative weight in a weighted service",
			serviceName: "test",
			configs: map[string]*runtime.UDPServiceInfo{
				"test": {
					UDPService: &dynamic.UDPService{
						Weighted: &dynamic.UDPWeightedRoundRobin{
							Services: []dynamic.UDPWRRService{
								{Name: "child", Weight: &negativeWeight},
							},
						},
					},
				},
			},
			expectedError: `invalid negative weight -2 for udp service "child"`,
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

func TestManager_BuildUDP_WeightedChildErrorIsReportedOnParent(t *testing.T) {
	conf := &runtime.UDPServiceInfo{
		UDPService: &dynamic.UDPService{
			Weighted: &dynamic.UDPWeightedRoundRobin{
				Services: []dynamic.UDPWRRService{
					{Name: "child"},
				},
			},
		},
	}

	manager := NewManager(&runtime.Configuration{
		UDPServices: map[string]*runtime.UDPServiceInfo{"parent": conf},
	})

	handler, err := manager.BuildUDP(t.Context(), "parent")
	require.Error(t, err)
	require.Nil(t, handler)

	assert.Equal(t, runtime.StatusDisabled, conf.Status)
	assert.Equal(t, []string{`the udp service "child" does not exist`}, conf.Err)
}
