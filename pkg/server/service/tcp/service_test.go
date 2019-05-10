package tcp

import (
	"context"
	"errors"
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
		expectedError error
	}{
		{
			desc:          "without configuration",
			serviceName:   "test",
			configs:       nil,
			expectedError: errors.New(`the service "test" does not exist`),
		},
		{
			desc:        "missing lb configuration",
			serviceName: "test",
			configs: map[string]*config.TCPServiceInfo{
				"test": {
					TCPService: &config.TCPService{},
				},
			},
			expectedError: errors.New(`the service "test" doesn't have any TCP load balancer`),
		},
		{
			desc:        "no such host",
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
			desc:        "invalid IP address",
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
			expectedError: errors.New(`in service "test": address foobar: missing port in address`),
		},
		{
			desc:        "Simple service name",
			serviceName: "serviceName",
			configs: map[string]*config.TCPServiceInfo{
				"serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{Method: "wrr"},
					},
				},
			},
		},
		{
			desc:        "Service name with provider",
			serviceName: "provider-1.serviceName",
			configs: map[string]*config.TCPServiceInfo{
				"provider-1.serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{Method: "wrr"},
					},
				},
			},
		},
		{
			desc:        "Service name with provider in context",
			serviceName: "serviceName",
			configs: map[string]*config.TCPServiceInfo{
				"provider-1.serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{Method: "wrr"},
					},
				},
			},
			providerName: "provider-1",
		},
		{
			desc:        "Server with correct host:port as address",
			serviceName: "serviceName",
			configs: map[string]*config.TCPServiceInfo{
				"provider-1.serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
								{
									Address: "foobar.com:80",
								},
							},
							Method: "wrr",
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
				"provider-1.serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
								{
									Address: "192.168.0.12:80",
								},
							},
							Method: "wrr",
						},
					},
				},
			},
			providerName: "provider-1",
		},
		{
			desc:        "Server address, hostname but missing port",
			serviceName: "serviceName",
			configs: map[string]*config.TCPServiceInfo{
				"provider-1.serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
								{
									Address: "foobar.com",
								},
							},
							Method: "wrr",
						},
					},
				},
			},
			providerName:  "provider-1",
			expectedError: errors.New(`in service "provider-1.serviceName": address foobar.com: missing port in address`),
		},
		{
			desc:        "Server address, ip but missing port",
			serviceName: "serviceName",
			configs: map[string]*config.TCPServiceInfo{
				"provider-1.serviceName": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
								{
									Address: "192.168.0.12",
								},
							},
							Method: "wrr",
						},
					},
				},
			},
			providerName:  "provider-1",
			expectedError: errors.New(`in service "provider-1.serviceName": address 192.168.0.12: missing port in address`),
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
				ctx = internal.AddProviderInContext(ctx, test.providerName+".foobar")
			}

			handler, err := manager.BuildTCP(ctx, test.serviceName)

			assert.Equal(t, test.expectedError, err)
			if test.expectedError != nil {
				require.Nil(t, handler)
			} else {
				require.NotNil(t, handler)
			}

			assert.Equal(t, test.expectedError, err)
		})
	}
}
