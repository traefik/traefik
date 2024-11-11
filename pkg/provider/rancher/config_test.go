package rancher

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/tls"
	"github.com/traefik/traefik/v2/pkg/types"
)

func pointer[T any](v T) *T { return &v }

func Test_buildConfiguration(t *testing.T) {
	testCases := []struct {
		desc        string
		containers  []rancherData
		constraints string
		expected    *dynamic.Configuration
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test1": {
							Service:     "Test1",
							Rule:        "Host(`Test1.traefik.wtf`)",
							DefaultRule: true,
						},
						"Test2": {
							Service:     "Test2",
							Rule:        "Host(`Test2.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
							},
						},
						"Test2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: pointer(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test1": {
							Service:     "Test1",
							Rule:        "Host(`Test1.traefik.wtf`)",
							DefaultRule: true,
						},
						"Test2": {
							Service:     "Test2",
							Rule:        "Host(`Test2.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: pointer(true),
							},
						},
						"Test2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://128.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Router1": {
							Service: "Service1",
							Rule:    "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Router1": {
							Service: "Test",
							Rule:    "Host(`foo.com`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							Middlewares: []string{"Middleware1"},
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"Middleware1": {
							BasicAuth: &dynamic.BasicAuth{
								Users: []string{
									"test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/",
									"test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
								},
							},
						},
					},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "Middlewares used in TCP router",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.Test.rule":                               "HostSNI(`foo.bar`)",
						"traefik.tcp.middlewares.Middleware1.ipwhitelist.sourcerange": "foobar, fiibar",
						"traefik.tcp.routers.Test.middlewares":                        "Middleware1",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"Test": {
							Service:     "Test",
							Rule:        "HostSNI(`foo.bar`)",
							Middlewares: []string{"Middleware1"},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{
						"Middleware1": {
							IPWhiteList: &dynamic.TCPIPWhiteList{
								SourceRange: []string{"foobar", "fiibar"},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"Test": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:80",
									},
								},
								TerminationDelay: pointer(100),
							},
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {
							Service: "Test",
							Rule:    "HostSNI(`foo.bar`)",
							TLS:     &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"Test": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:80",
									},
								},
								TerminationDelay: pointer(100),
							},
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "udp with label",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.udp.routers.foo.entrypoints": "mydns",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"foo": {
							Service:     "Test",
							EntryPoints: []string{"mydns"},
						},
					},
					Services: map[string]*dynamic.UDPService{
						"Test": {
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
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"Test": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:80",
									},
								},
								TerminationDelay: pointer(100),
							},
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {
							Service: "foo",
							Rule:    "HostSNI(`foo.bar`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"foo": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:8080",
									},
								},
								TerminationDelay: pointer(100),
							},
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "udp with label and port",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.udp.routers.foo.entrypoints":               "mydns",
						"traefik.udp.services.foo.loadbalancer.server.port": "8080",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"foo": {
							Service:     "foo",
							EntryPoints: []string{"mydns"},
						},
					},
					Services: map[string]*dynamic.UDPService{
						"foo": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "127.0.0.1:8080",
									},
								},
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {
							Service: "foo",
							Rule:    "HostSNI(`foo.bar`)",
							TLS:     &dynamic.RouterTCPTLSConfig{},
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"foo": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:8080",
									},
									{
										Address: "127.0.0.2:8080",
									},
								},
								TerminationDelay: pointer(100),
							},
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: pointer(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "udp with label and port and http service",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.udp.routers.foo.entrypoints":                        "mydns",
						"traefik.udp.services.foo.loadbalancer.server.port":          "8080",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1", "127.0.0.2"},
					Health:     "",
					State:      "",
				},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"foo": {
							Service:     "foo",
							EntryPoints: []string{"mydns"},
						},
					},
					Services: map[string]*dynamic.UDPService{
						"foo": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
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
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: pointer(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
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
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"foo": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:8080",
									},
								},
								TerminationDelay: pointer(100),
							},
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "udp with label for tcp service",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.udp.services.foo.loadbalancer.server.port": "8080",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &dynamic.Configuration{
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{
						"foo": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{
										Address: "127.0.0.1:8080",
									},
								},
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "tcp with label for tcp service, with termination delay",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.services.foo.loadbalancer.server.port":      "8080",
						"traefik.tcp.services.foo.loadbalancer.terminationdelay": "200",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"foo": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:8080",
									},
								},
								TerminationDelay: pointer(200),
							},
						},
					},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one service with default generated certificate labels",
			containers: []rancherData{
				{
					Name: "Test",
					Labels: map[string]string{
						"traefik.tls.stores.default.defaultgeneratedcert.resolver":    "foobar",
						"traefik.tls.stores.default.defaultgeneratedcert.domain.main": "foobar",
						"traefik.tls.stores.default.defaultgeneratedcert.domain.sans": "foobar, fiibar",
					},
					Port:       "80/tcp",
					Containers: []string{"127.0.0.1"},
					Health:     "",
					State:      "",
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{
						"default": {
							DefaultGeneratedCert: &tls.GeneratedCert{
								Resolver: "foobar",
								Domain: &types.Domain{
									Main: "foobar",
									SANs: []string{"foobar", "fiibar"},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
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

			for i := range len(test.containers) {
				var err error
				test.containers[i].ExtraConf, err = p.getConfiguration(test.containers[i])
				require.NoError(t, err)
			}

			configuration := p.buildConfiguration(context.Background(), test.containers)

			assert.Equal(t, test.expected, configuration)
		})
	}
}
