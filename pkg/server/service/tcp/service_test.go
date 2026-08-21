package tcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/server/provider"
)

func TestManager_BuildTCP(t *testing.T) {
	negativeWeight := -2

	testCases := []struct {
		desc          string
		serviceName   string
		configs       map[string]*runtime.TCPServiceInfo
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
			desc:        "negative weight in a weighted service",
			serviceName: "test",
			configs: map[string]*runtime.TCPServiceInfo{
				"test": {
					TCPService: &dynamic.TCPService{
						Weighted: &dynamic.TCPWeightedRoundRobin{
							Services: []dynamic.TCPWRRService{
								{Name: "child", Weight: &negativeWeight},
							},
						},
					},
				},
			},
			expectedError: `invalid negative weight -2 for service "child"`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			manager := NewManager(&runtime.Configuration{
				TCPServices: test.configs,
			})

			ctx := t.Context()
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

func TestManager_BuildTCP_WeightedChildErrorIsReportedOnParent(t *testing.T) {
	conf := &runtime.TCPServiceInfo{
		TCPService: &dynamic.TCPService{
			Weighted: &dynamic.TCPWeightedRoundRobin{
				Services: []dynamic.TCPWRRService{
					{Name: "child"},
				},
			},
		},
	}

	manager := NewManager(&runtime.Configuration{
		TCPServices: map[string]*runtime.TCPServiceInfo{"parent": conf},
	})

	handler, err := manager.BuildTCP(t.Context(), "parent")
	require.Error(t, err)
	require.Nil(t, handler)

	assert.Equal(t, runtime.StatusDisabled, conf.Status)
	assert.Equal(t, []string{`the service "child" does not exist`}, conf.Err)
}
