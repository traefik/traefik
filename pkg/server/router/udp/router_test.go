package udp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/server/service/udp"
)

func TestRuntimeConfiguration(t *testing.T) {
	testCases := []struct {
		desc          string
		serviceConfig map[string]*runtime.UDPServiceInfo
		routerConfig  map[string]*runtime.UDPRouterInfo
		expectedError int
	}{
		{
			desc: "No error",
			serviceConfig: map[string]*runtime.UDPServiceInfo{
				"foo-service": {
					UDPService: &dynamic.UDPService{
						LoadBalancer: &dynamic.UDPServersLoadBalancer{
							Servers: []dynamic.UDPServer{
								{
									Port:    "8085",
									Address: "127.0.0.1:8085",
								},
								{
									Address: "127.0.0.1:8086",
									Port:    "8086",
								},
							},
						},
					},
				},
			},
			routerConfig: map[string]*runtime.UDPRouterInfo{
				"foo": {
					UDPRouter: &dynamic.UDPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
					},
				},
				"bar": {
					UDPRouter: &dynamic.UDPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
					},
				},
			},
			expectedError: 0,
		},
		{
			desc: "Router with unknown service",
			serviceConfig: map[string]*runtime.UDPServiceInfo{
				"foo-service": {
					UDPService: &dynamic.UDPService{
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
			routerConfig: map[string]*runtime.UDPRouterInfo{
				"foo": {
					UDPRouter: &dynamic.UDPRouter{
						EntryPoints: []string{"web"},
						Service:     "wrong-service",
					},
				},
				"bar": {
					UDPRouter: &dynamic.UDPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
					},
				},
			},
			expectedError: 1,
		},
		{
			desc: "Router with broken service",
			serviceConfig: map[string]*runtime.UDPServiceInfo{
				"foo-service": {
					UDPService: &dynamic.UDPService{
						LoadBalancer: nil,
					},
				},
			},
			routerConfig: map[string]*runtime.UDPRouterInfo{
				"bar": {
					UDPRouter: &dynamic.UDPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
					},
				},
			},
			expectedError: 2,
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			entryPoints := []string{"web"}

			conf := &runtime.Configuration{
				UDPServices: test.serviceConfig,
				UDPRouters:  test.routerConfig,
			}
			serviceManager := udp.NewManager(conf)
			routerManager := NewManager(conf, serviceManager)

			_ = routerManager.BuildHandlers(context.Background(), entryPoints)

			// even though conf was passed by argument to the manager builders above,
			// it's ok to use it as the result we check, because everything worth checking
			// can be accessed by pointers in it.
			var allErrors int
			for _, v := range conf.UDPServices {
				if v.Err != nil {
					allErrors++
				}
			}
			for _, v := range conf.UDPRouters {
				if len(v.Err) > 0 {
					allErrors++
				}
			}
			assert.Equal(t, test.expectedError, allErrors)
		})
	}
}
