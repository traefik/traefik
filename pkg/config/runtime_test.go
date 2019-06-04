package config_test

import (
	"context"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// all the Routers/Middlewares/Services are considered fully qualified
func TestPopulateUsedby(t *testing.T) {
	testCases := []struct {
		desc     string
		conf     *config.RuntimeConfiguration
		expected config.RuntimeConfiguration
	}{
		{
			desc:     "nil config",
			conf:     nil,
			expected: config.RuntimeConfiguration{},
		},
		{
			desc: "One service used by two routers",
			conf: &config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"myprovider.foo": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`bar.foo`)",
						},
					},
					"myprovider.bar": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
						},
					},
				},
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						Service: &config.Service{
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{URL: "http://127.0.0.1:8085"},
									{URL: "http://127.0.0.1:8086"},
								},
								HealthCheck: &config.HealthCheck{
									Interval: "500ms",
									Path:     "/health",
								},
							},
						},
					},
				},
			},
			expected: config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"myprovider.foo": {},
					"myprovider.bar": {},
				},
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						UsedBy: []string{"myprovider.bar", "myprovider.foo"},
					},
				},
			},
		},
		{
			desc: "One service used by two routers, but one router with wrong rule",
			conf: &config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						Service: &config.Service{
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{URL: "http://127.0.0.1"},
								},
							},
						},
					},
				},
				Routers: map[string]*config.RouterInfo{
					"myprovider.foo": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "WrongRule(`bar.foo`)",
						},
					},
					"myprovider.bar": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
						},
					},
				},
			},
			expected: config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"myprovider.foo": {},
					"myprovider.bar": {},
				},
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						UsedBy: []string{"myprovider.bar", "myprovider.foo"},
					},
				},
			},
		},
		{
			desc: "Broken Service used by one Router",
			conf: &config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						Service: &config.Service{
							LoadBalancer: nil,
						},
					},
				},
				Routers: map[string]*config.RouterInfo{
					"myprovider.bar": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
						},
					},
				},
			},
			expected: config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"myprovider.bar": {},
				},
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						UsedBy: []string{"myprovider.bar"},
					},
				},
			},
		},
		{
			desc: "2 different Services each used by a disctinct router.",
			conf: &config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						Service: &config.Service{
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:8085",
									},
									{
										URL: "http://127.0.0.1:8086",
									},
								},
								HealthCheck: &config.HealthCheck{
									Interval: "500ms",
									Path:     "/health",
								},
							},
						},
					},
					"myprovider.bar-service": {
						Service: &config.Service{
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:8087",
									},
									{
										URL: "http://127.0.0.1:8088",
									},
								},
								HealthCheck: &config.HealthCheck{
									Interval: "500ms",
									Path:     "/health",
								},
							},
						},
					},
				},
				Routers: map[string]*config.RouterInfo{
					"myprovider.foo": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`bar.foo`)",
						},
					},
					"myprovider.bar": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.bar-service",
							Rule:        "Host(`foo.bar`)",
						},
					},
				},
			},
			expected: config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"myprovider.bar": {},
					"myprovider.foo": {},
				},
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						UsedBy: []string{"myprovider.foo"},
					},
					"myprovider.bar-service": {
						UsedBy: []string{"myprovider.bar"},
					},
				},
			},
		},
		{
			desc: "2 middlewares both used by 2 Routers",
			conf: &config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						Service: &config.Service{
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1",
									},
								},
							},
						},
					},
				},
				Middlewares: map[string]*config.MiddlewareInfo{
					"myprovider.auth": {
						Middleware: &config.Middleware{
							BasicAuth: &config.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
					},
					"myprovider.addPrefixTest": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/toto",
							},
						},
					},
				},
				Routers: map[string]*config.RouterInfo{
					"myprovider.bar": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "addPrefixTest"},
						},
					},
					"myprovider.test": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar.other`)",
							Middlewares: []string{"addPrefixTest", "auth"},
						},
					},
				},
			},
			expected: config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"myprovider.bar":  {},
					"myprovider.test": {},
				},
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						UsedBy: []string{"myprovider.bar", "myprovider.test"},
					},
				},
				Middlewares: map[string]*config.MiddlewareInfo{
					"myprovider.auth": {
						UsedBy: []string{"myprovider.bar", "myprovider.test"},
					},
					"myprovider.addPrefixTest": {
						UsedBy: []string{"myprovider.bar", "myprovider.test"},
					},
				},
			},
		},
		{
			desc: "Unknown middleware is not used by the Router",
			conf: &config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						Service: &config.Service{
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1",
									},
								},
							},
						},
					},
				},
				Middlewares: map[string]*config.MiddlewareInfo{
					"myprovider.auth": {
						Middleware: &config.Middleware{
							BasicAuth: &config.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
					},
				},
				Routers: map[string]*config.RouterInfo{
					"myprovider.bar": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"unknown"},
						},
					},
				},
			},
			expected: config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						UsedBy: []string{"myprovider.bar"},
					},
				},
			},
		},
		{
			desc: "Broken middleware is used by Router",
			conf: &config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						Service: &config.Service{
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1",
									},
								},
							},
						},
					},
				},
				Middlewares: map[string]*config.MiddlewareInfo{
					"myprovider.auth": {
						Middleware: &config.Middleware{
							BasicAuth: &config.BasicAuth{
								Users: []string{"badConf"},
							},
						},
					},
				},
				Routers: map[string]*config.RouterInfo{
					"myprovider.bar": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"myprovider.auth"},
						},
					},
				},
			},
			expected: config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"myprovider.bar": {},
				},
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						UsedBy: []string{"myprovider.bar"},
					},
				},
				Middlewares: map[string]*config.MiddlewareInfo{
					"myprovider.auth": {
						UsedBy: []string{"myprovider.bar"},
					},
				},
			},
		},
		{
			desc: "2 middlewares from 2 disctinct providers both used by 2 Routers",
			conf: &config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						Service: &config.Service{
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1",
									},
								},
							},
						},
					},
				},
				Middlewares: map[string]*config.MiddlewareInfo{
					"myprovider.auth": {
						Middleware: &config.Middleware{
							BasicAuth: &config.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
					},
					"myprovider.addPrefixTest": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/titi",
							},
						},
					},
					"anotherprovider.addPrefixTest": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/toto",
							},
						},
					},
				},
				Routers: map[string]*config.RouterInfo{
					"myprovider.bar": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "anotherprovider.addPrefixTest"},
						},
					},
					"myprovider.test": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar.other`)",
							Middlewares: []string{"addPrefixTest", "auth"},
						},
					},
				},
			},
			expected: config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"myprovider.bar":  {},
					"myprovider.test": {},
				},
				Services: map[string]*config.ServiceInfo{
					"myprovider.foo-service": {
						UsedBy: []string{"myprovider.bar", "myprovider.test"},
					},
				},
				Middlewares: map[string]*config.MiddlewareInfo{
					"myprovider.auth": {
						UsedBy: []string{"myprovider.bar", "myprovider.test"},
					},
					"myprovider.addPrefixTest": {
						UsedBy: []string{"myprovider.test"},
					},
					"anotherprovider.addPrefixTest": {
						UsedBy: []string{"myprovider.bar"},
					},
				},
			},
		},

		// TCP tests from hereon
		{
			desc: "TCP, One service used by two routers",
			conf: &config.RuntimeConfiguration{
				TCPRouters: map[string]*config.TCPRouterInfo{
					"myprovider.foo": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`bar.foo`)",
						},
					},
					"myprovider.bar": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
						},
					},
				},
				TCPServices: map[string]*config.TCPServiceInfo{
					"myprovider.foo-service": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1",
										Port:    "8085",
									},
									{
										Address: "127.0.0.1",
										Port:    "8086",
									},
								},
							},
						},
					},
				},
			},
			expected: config.RuntimeConfiguration{
				TCPRouters: map[string]*config.TCPRouterInfo{
					"myprovider.foo": {},
					"myprovider.bar": {},
				},
				TCPServices: map[string]*config.TCPServiceInfo{
					"myprovider.foo-service": {
						UsedBy: []string{"myprovider.bar", "myprovider.foo"},
					},
				},
			},
		},
		{
			desc: "TCP, One service used by two routers, but one router with wrong rule",
			conf: &config.RuntimeConfiguration{
				TCPServices: map[string]*config.TCPServiceInfo{
					"myprovider.foo-service": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1",
									},
								},
							},
						},
					},
				},
				TCPRouters: map[string]*config.TCPRouterInfo{
					"myprovider.foo": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "WrongRule(`bar.foo`)",
						},
					},
					"myprovider.bar": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
						},
					},
				},
			},
			expected: config.RuntimeConfiguration{
				TCPRouters: map[string]*config.TCPRouterInfo{
					"myprovider.foo": {},
					"myprovider.bar": {},
				},
				TCPServices: map[string]*config.TCPServiceInfo{
					"myprovider.foo-service": {
						UsedBy: []string{"myprovider.bar", "myprovider.foo"},
					},
				},
			},
		},
		{
			desc: "TCP, Broken Service used by one Router",
			conf: &config.RuntimeConfiguration{
				TCPServices: map[string]*config.TCPServiceInfo{
					"myprovider.foo-service": {
						TCPService: &config.TCPService{
							LoadBalancer: nil,
						},
					},
				},
				TCPRouters: map[string]*config.TCPRouterInfo{
					"myprovider.bar": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
						},
					},
				},
			},
			expected: config.RuntimeConfiguration{
				TCPRouters: map[string]*config.TCPRouterInfo{
					"myprovider.bar": {},
				},
				TCPServices: map[string]*config.TCPServiceInfo{
					"myprovider.foo-service": {
						UsedBy: []string{"myprovider.bar"},
					},
				},
			},
		},
		{
			desc: "TCP, 2 different Services each used by a disctinct router.",
			conf: &config.RuntimeConfiguration{
				TCPServices: map[string]*config.TCPServiceInfo{
					"myprovider.foo-service": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1",
										Port:    "8085",
									},
									{
										Address: "127.0.0.1",
										Port:    "8086",
									},
								},
							},
						},
					},
					"myprovider.bar-service": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1",
										Port:    "8087",
									},
									{
										Address: "127.0.0.1",
										Port:    "8088",
									},
								},
							},
						},
					},
				},
				TCPRouters: map[string]*config.TCPRouterInfo{
					"myprovider.foo": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`bar.foo`)",
						},
					},
					"myprovider.bar": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.bar-service",
							Rule:        "Host(`foo.bar`)",
						},
					},
				},
			},
			expected: config.RuntimeConfiguration{
				TCPRouters: map[string]*config.TCPRouterInfo{
					"myprovider.bar": {},
					"myprovider.foo": {},
				},
				TCPServices: map[string]*config.TCPServiceInfo{
					"myprovider.foo-service": {
						UsedBy: []string{"myprovider.foo"},
					},
					"myprovider.bar-service": {
						UsedBy: []string{"myprovider.bar"},
					},
				},
			},
		},
	}
	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			runtimeConf := test.conf
			runtimeConf.PopulateUsedBy()

			for key, expectedService := range test.expected.Services {
				require.NotNil(t, runtimeConf.Services[key])
				assert.Equal(t, expectedService.UsedBy, runtimeConf.Services[key].UsedBy)
			}

			for key, expectedMiddleware := range test.expected.Middlewares {
				require.NotNil(t, runtimeConf.Middlewares[key])
				assert.Equal(t, expectedMiddleware.UsedBy, runtimeConf.Middlewares[key].UsedBy)
			}

			for key, expectedTCPService := range test.expected.TCPServices {
				require.NotNil(t, runtimeConf.TCPServices[key])
				assert.Equal(t, expectedTCPService.UsedBy, runtimeConf.TCPServices[key].UsedBy)
			}
		})
	}

}

func TestGetTCPRoutersByEntrypoints(t *testing.T) {
	testCases := []struct {
		desc        string
		conf        config.Configuration
		entryPoints []string
		expected    map[string]map[string]*config.TCPRouterInfo
	}{
		{
			desc:        "Empty Configuration without entrypoint",
			conf:        config.Configuration{},
			entryPoints: []string{""},
			expected:    map[string]map[string]*config.TCPRouterInfo{},
		},
		{
			desc:        "Empty Configuration with unknown entrypoints",
			conf:        config.Configuration{},
			entryPoints: []string{"foo"},
			expected:    map[string]map[string]*config.TCPRouterInfo{},
		},
		{
			desc: "Valid configuration with an unknown entrypoint",
			conf: config.Configuration{
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`bar.foo`)",
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "HostSNI(`bar.foo`)",
						},
					},
				},
			},
			entryPoints: []string{"foo"},
			expected:    map[string]map[string]*config.TCPRouterInfo{},
		},
		{
			desc: "Valid configuration with a known entrypoint",
			conf: config.Configuration{
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`bar.foo`)",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "myprovider.bar-service",
							Rule:        "Host(`foo.bar`)",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "Host(`bar.foobar`)",
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "HostSNI(`bar.foo`)",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "myprovider.bar-service",
							Rule:        "HostSNI(`foo.bar`)",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "HostSNI(`bar.foobar`)",
						},
					},
				},
			},
			entryPoints: []string{"web"},
			expected: map[string]map[string]*config.TCPRouterInfo{
				"web": {
					"foo": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "HostSNI(`bar.foo`)",
						},
					},
					"foobar": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "HostSNI(`bar.foobar`)",
						},
					},
				},
			},
		},
		{
			desc: "Valid configuration with multiple known entrypoints",
			conf: config.Configuration{
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`bar.foo`)",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "myprovider.bar-service",
							Rule:        "Host(`foo.bar`)",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "Host(`bar.foobar`)",
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "HostSNI(`bar.foo`)",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "myprovider.bar-service",
							Rule:        "HostSNI(`foo.bar`)",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "HostSNI(`bar.foobar`)",
						},
					},
				},
			},
			entryPoints: []string{"web", "webs"},
			expected: map[string]map[string]*config.TCPRouterInfo{
				"web": {
					"foo": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "HostSNI(`bar.foo`)",
						},
					},
					"foobar": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "HostSNI(`bar.foobar`)",
						},
					},
				},
				"webs": {
					"bar": {
						TCPRouter: &config.TCPRouter{

							EntryPoints: []string{"webs"},
							Service:     "myprovider.bar-service",
							Rule:        "HostSNI(`foo.bar`)",
						},
					},
					"foobar": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "HostSNI(`bar.foobar`)",
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
			runtimeConfig := config.NewRuntimeConfig(test.conf)
			actual := runtimeConfig.GetTCPRoutersByEntrypoints(context.Background(), test.entryPoints)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestGetRoutersByEntrypoints(t *testing.T) {
	testCases := []struct {
		desc        string
		conf        config.Configuration
		entryPoints []string
		expected    map[string]map[string]*config.RouterInfo
	}{
		{
			desc:        "Empty Configuration without entrypoint",
			conf:        config.Configuration{},
			entryPoints: []string{""},
			expected:    map[string]map[string]*config.RouterInfo{},
		},
		{
			desc:        "Empty Configuration with unknown entrypoints",
			conf:        config.Configuration{},
			entryPoints: []string{"foo"},
			expected:    map[string]map[string]*config.RouterInfo{},
		},
		{
			desc: "Valid configuration with an unknown entrypoint",
			conf: config.Configuration{
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`bar.foo`)",
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "HostSNI(`bar.foo`)",
						},
					},
				},
			},
			entryPoints: []string{"foo"},
			expected:    map[string]map[string]*config.RouterInfo{},
		},
		{
			desc: "Valid configuration with a known entrypoint",
			conf: config.Configuration{
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`bar.foo`)",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "myprovider.bar-service",
							Rule:        "Host(`foo.bar`)",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "Host(`bar.foobar`)",
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "HostSNI(`bar.foo`)",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "myprovider.bar-service",
							Rule:        "HostSNI(`foo.bar`)",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "HostSNI(`bar.foobar`)",
						},
					},
				},
			},
			entryPoints: []string{"web"},
			expected: map[string]map[string]*config.RouterInfo{
				"web": {
					"foo": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`bar.foo`)",
						},
					},
					"foobar": {
						Router: &config.Router{
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "Host(`bar.foobar`)",
						},
					},
				},
			},
		},
		{
			desc: "Valid configuration with multiple known entrypoints",
			conf: config.Configuration{
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`bar.foo`)",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "myprovider.bar-service",
							Rule:        "Host(`foo.bar`)",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "Host(`bar.foobar`)",
						},
					},
				},
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"foo": {
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "HostSNI(`bar.foo`)",
						},
						"bar": {
							EntryPoints: []string{"webs"},
							Service:     "myprovider.bar-service",
							Rule:        "HostSNI(`foo.bar`)",
						},
						"foobar": {
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "HostSNI(`bar.foobar`)",
						},
					},
				},
			},
			entryPoints: []string{"web", "webs"},
			expected: map[string]map[string]*config.RouterInfo{
				"web": {
					"foo": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`bar.foo`)",
						},
					},
					"foobar": {
						Router: &config.Router{
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "Host(`bar.foobar`)",
						},
					},
				},
				"webs": {
					"bar": {
						Router: &config.Router{

							EntryPoints: []string{"webs"},
							Service:     "myprovider.bar-service",
							Rule:        "Host(`foo.bar`)",
						},
					},
					"foobar": {
						Router: &config.Router{
							EntryPoints: []string{"web", "webs"},
							Service:     "myprovider.foobar-service",
							Rule:        "Host(`bar.foobar`)",
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
			runtimeConfig := config.NewRuntimeConfig(test.conf)
			actual := runtimeConfig.GetRoutersByEntrypoints(context.Background(), test.entryPoints, false)
			assert.Equal(t, test.expected, actual)
		})
	}
}
