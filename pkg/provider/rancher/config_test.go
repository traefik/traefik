package rancher

import (
	"context"
	"testing"

	"github.com/containous/traefik/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_buildConfiguration(t *testing.T) {
	testCases := []struct {
		desc        string
		containers  []rancherData
		constraints string
		expected    *config.Configuration
	}{
		{
			desc: "one service no label",
			containers: []rancherData{
				{
					Name:       "Test",
					Labels:     map[string]string{},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "two services no label",
			containers: []rancherData{
				{
					Name:       "Test1",
					Labels:     map[string]string{},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
				{
					Name:       "Test2",
					Labels:     map[string]string{},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.2"},
					Health:     "",
					State:      "",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test1": {
							Service: "Test1",
							Rule:    "Host(`Test1.traefik.wtf`)",
						},
						"Test2": {
							Service: "Test2",
							Rule:    "Host(`Test2.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
						},
						"Test2": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "two services no label multiple containers",
			containers: []rancherData{
				{
					Name:       "Test1",
					Labels:     map[string]string{},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1", "127.0.0.2"},
					Health:     "",
					State:      "",
				},
				{
					Name:       "Test2",
					Labels:     map[string]string{},
					Port:       "80/tcp",
					Containers: []string{"128.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test1": {
							Service: "Test1",
							Rule:    "Host(`Test1.traefik.wtf`)",
						},
						"Test2": {
							Service: "Test2",
							Rule:    "Host(`Test2.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: true,
							},
						},
						"Test2": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://128.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "one service some labels",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
						"traefik.http.routers.Router1.rule":                          "Host(`foo.com`)",
						"traefik.http.routers.Router1.service":                       "Service1",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Router1": {
							Service: "Service1",
							Rule:    "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "one service which is unhealthy",
			containers: []rancherData{
				{
					Name:       "Test",
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "broken",
					State:      "",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "one service which is upgrading",
			containers: []rancherData{
				{
					Name:       "Test",
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "upgradefailed",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "one service with rule label and has a host exposed port",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					Port:       "12345:80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Router1": {
							Service: "Test",
							Rule:    "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "one service with non matching constraints",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					Port:       "12345:80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			constraints: `Label("traefik.tags", "bar")`,
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "one service with matching constraints",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.tags": "foo",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			constraints: `Label("traefik.tags", "foo")`,
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "Middlewares used in router",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.basicauth.users": "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						"traefik.http.routers.Test.middlewares":                "Middleware1",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							Middlewares: []string{"Middleware1"},
						},
					},
					Middlewares: map[string]*config.Middleware{
						"Middleware1": {
							BasicAuth: &config.BasicAuth{
								Users: []string{
									"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
									"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
								},
							},
						},
					},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "Port in labels",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.services.Test.loadbalancer.server.port": "80",
					},
					Port:       "",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers:  map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Test": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "tcp with label",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule": "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls":  "true",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"foo": {
							Service: "Test",
							Rule:    "HostSNI(`foo.bar`)",
							TLS:     &config.RouterTCPTLSConfig{},
						},
					},
					Services: map[string]*config.TCPService{
						"Test": {
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
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "tcp with label without rule",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.tls": "true",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{
						"Test": {
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
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "tcp with label and port",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule":                      "HostSNI(`foo.bar`)",
						"traefik.tcp.services.foo.loadbalancer.server.port": "8080",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"foo": {
							Service: "foo",
							Rule:    "HostSNI(`foo.bar`)",
						},
					},
					Services: map[string]*config.TCPService{
						"foo": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1:8080",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
		{
			desc: "tcp with label and port and http service",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule":                               "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls":                                "true",
						"traefik.tcp.services.foo.loadbalancer.server.port":          "8080",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1", "127.0.0.2"},
					Health:     "",
					State:      "",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{
						"foo": {
							Service: "foo",
							Rule:    "HostSNI(`foo.bar`)",
							TLS:     &config.RouterTCPTLSConfig{},
						},
					},
					Services: map[string]*config.TCPService{
						"foo": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1:8080",
									},
									{
										Address: "127.0.0.2:8080",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers: map[string]*config.Router{
						"Test": {
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*config.Middleware{},
					Services: map[string]*config.Service{
						"Service1": {
							LoadBalancer: &config.LoadBalancerService{
								Servers: []config.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: true,
							},
						},
					},
				},
			},
		},
		{
			desc: "tcp with label for tcp service",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.services.foo.loadbalancer.server.port": "8080",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &config.Configuration{
				TCP: &config.TCPConfiguration{
					Routers: map[string]*config.TCPRouter{},
					Services: map[string]*config.TCPService{
						"foo": {
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1:8080",
									},
								},
							},
						},
					},
				},
				HTTP: &config.HTTPConfiguration{
					Routers:     map[string]*config.Router{},
					Middlewares: map[string]*config.Middleware{},
					Services:    map[string]*config.Service{},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{
				ExposedByDefault:          true,
				DefaultRule:               "Host(`{{ normalize .Name }}.traefik.wtf`)",
				EnableServiceHealthFilter: true,
			}

			p.Constraints = test.constraints

			err := p.Init()
			require.NoError(t, err)

			for i := 0; i < len(test.containers); i++ {
				var err error
				test.containers[i].ExtraConf, err = p.getConfiguration(test.containers[i])
				require.NoError(t, err)
			}

			configuration := p.buildConfiguration(context.Background(), test.containers)

			assert.Equal(t, test.expected, configuration)
		})
	}
}
