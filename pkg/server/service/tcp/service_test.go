package tcp

import (
	"context"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestManager_BuildTCP(t *testing.T) {
	testCases := []struct {
		desc          string
		serviceName   string
		configs       map[string]*config.TCPService
		expectedError string
	}{
		{
			desc:          "without configuration",
			serviceName:   "test",
			configs:       nil,
			expectedError: `the service "test" does not exits`,
		},
		{
			desc:        "missing lb configuration",
			serviceName: "test",
			configs: map[string]*config.TCPService{
				"test": {},
			},
			expectedError: `the service "test" doesn't have any TCP load balancer`,
		},
		{
			desc:        "no such host",
			serviceName: "test",
			configs: map[string]*config.TCPService{
				"test": {
					LoadBalancer: &config.TCPLoadBalancerService{
						Servers: []config.TCPServer{
							{Address: "test:31"},
						},
					},
				},
			},
		},
		{
			desc:        "invalid IP address",
			serviceName: "test",
			configs: map[string]*config.TCPService{
				"test": {
					LoadBalancer: &config.TCPLoadBalancerService{
						Servers: []config.TCPServer{
							{Address: "foobar"},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			manager := NewManager(test.configs)

			handler, err := manager.BuildTCP(context.Background(), test.serviceName)

			if test.expectedError != "" {
				if err == nil {
					require.Error(t, err)
				} else {
					require.EqualError(t, err, test.expectedError)
					require.Nil(t, handler)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, handler)
			}
		})
	}
}
