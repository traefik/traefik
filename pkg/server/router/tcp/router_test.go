package tcp

import (
	"context"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/server/service/tcp"
	"github.com/stretchr/testify/assert"
)

func TestRuntimeConfiguration(t *testing.T) {
	testCases := []struct {
		desc          string
		serviceConfig map[string]*config.TCPServiceInfo
		routerConfig  map[string]*config.TCPRouterInfo
		expectedError int
	}{
		{
			desc: "No error",
			serviceConfig: map[string]*config.TCPServiceInfo{
				"foo-service": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
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
			routerConfig: map[string]*config.TCPRouterInfo{
				"foo": {
					TCPRouter: &config.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`bar.foo`)",
					},
				},
				"bar": {
					TCPRouter: &config.TCPRouter{

						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
					},
				},
			},
			expectedError: 0,
		},
		{
			desc: "One router with wrong rule",
			serviceConfig: map[string]*config.TCPServiceInfo{
				"foo-service": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
								{
									Address: "127.0.0.1:80",
								},
							},
						},
					},
				},
			},
			routerConfig: map[string]*config.TCPRouterInfo{
				"foo": {
					TCPRouter: &config.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "WrongRule(`bar.foo`)",
					},
				},

				"bar": {
					TCPRouter: &config.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
					},
				},
			},
			expectedError: 1,
		},
		{
			desc: "All router with wrong rule",
			serviceConfig: map[string]*config.TCPServiceInfo{
				"foo-service": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
								{
									Address: "127.0.0.1:80",
								},
							},
						},
					},
				},
			},
			routerConfig: map[string]*config.TCPRouterInfo{
				"foo": {
					TCPRouter: &config.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "WrongRule(`bar.foo`)",
					},
				},
				"bar": {
					TCPRouter: &config.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "WrongRule(`foo.bar`)",
					},
				},
			},
			expectedError: 2,
		},
		{
			desc: "Router with unknown service",
			serviceConfig: map[string]*config.TCPServiceInfo{
				"foo-service": {
					TCPService: &config.TCPService{
						LoadBalancer: &config.TCPLoadBalancerService{
							Servers: []config.TCPServer{
								{
									Address: "127.0.0.1:80",
								},
							},
						},
					},
				},
			},
			routerConfig: map[string]*config.TCPRouterInfo{
				"foo": {
					TCPRouter: &config.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "wrong-service",
						Rule:        "HostSNI(`bar.foo`)",
					},
				},
				"bar": {
					TCPRouter: &config.TCPRouter{

						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
					},
				},
			},
			expectedError: 1,
		},
		{
			desc: "Router with broken service",
			serviceConfig: map[string]*config.TCPServiceInfo{
				"foo-service": {
					TCPService: &config.TCPService{
						LoadBalancer: nil,
					},
				},
			},
			routerConfig: map[string]*config.TCPRouterInfo{
				"bar": {
					TCPRouter: &config.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
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

			conf := &config.RuntimeConfiguration{
				TCPServices: test.serviceConfig,
				TCPRouters:  test.routerConfig,
			}
			serviceManager := tcp.NewManager(conf)
			routerManager := NewManager(conf, serviceManager,
				nil, nil, nil)

			_ = routerManager.BuildHandlers(context.Background(), entryPoints)

			// even though conf was passed by argument to the manager builders above,
			// it's ok to use it as the result we check, because everything worth checking
			// can be accessed by pointers in it.
			var allErrors int
			for _, v := range conf.TCPServices {
				if v.Err != nil {
					allErrors++
				}
			}
			for _, v := range conf.TCPRouters {
				if v.Err != "" {
					allErrors++
				}
			}
			assert.Equal(t, test.expectedError, allErrors)
		})
	}

}
