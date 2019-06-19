package tcp

import (
	"context"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/server/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_BuildTCP(t *testing.T) {
	testCases := []struct {
		desc          string
		serviceName   string
		configs       map[string]*config.TCPServiceInfo
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
			configs: map[string]*config.TCPServiceInfo{
				"test": {
					TCPService: &config.TCPService{},
				},
			},
			expectedError: `the service "test" doesn't have any TCP load balancer`,
		},
		{
			desc:        "no such host, server is skipped, error is logged",
			serviceName: "test",
			configs: map[string]*config.TCPServiceInfo{
				"test": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
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
			configs: map[string]*config.TCPServiceInfo{
				"test": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
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
			configs: map[string]*config.TCPServiceInfo{
				"serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{},
					},
				},
			},
		},
		{
			desc:        "Service name with provider",
			serviceName: "provider-1@serviceName",
			configs: map[string]*config.TCPServiceInfo{
				"provider-1@serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{},
					},
				},
			},
		},
		{
			desc:        "Service name with provider in context",
			serviceName: "serviceName",
			configs: map[string]*config.TCPServiceInfo{
				"provider-1@serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{},
					},
				},
			},
			providerName: "provider-1",
		},
		{
			desc:        "Server with correct host:port as address",
			serviceName: "serviceName",
			configs: map[string]*config.TCPServiceInfo{
				"provider-1@serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
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
			configs: map[string]*config.TCPServiceInfo{
				"provider-1@serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
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
			configs: map[string]*config.TCPServiceInfo{
				"provider-1@serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
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
			configs: map[string]*config.TCPServiceInfo{
				"provider-1@serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
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
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			manager := NewManager(&config.RuntimeConfiguration{
				TCPServices: test.configs,
			})

			ctx := context.Background()
			if len(test.providerName) > 0 {
				ctx = internal.AddProviderInContext(ctx, test.providerName+"@foobar")
			}

			handler, err := manager.BuildTCP(ctx, test.serviceName)

			if test.expectedError != "" {
				assert.EqualError(t, err, test.expectedError)
				require.Nil(t, handler)
			} else {
				assert.Nil(t, err)
				require.NotNil(t, handler)
			}
		})
	}
}
