package consulcatalog

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/tls"
)

func Int(v int) *int    { return &v }
func Bool(v bool) *bool { return &v }

func TestDefaultRule(t *testing.T) {
	testCases := []struct {
		desc        string
		items       []itemData
		defaultRule string
		expected    *dynamic.Configuration
	}{
		{
			desc: "default rule with no variable",
			items: []itemData{
				{
					ID:      "id",
					Node:    "Node1",
					Name:    "Test",
					Address: "127.0.0.1",
					Port:    "80",
					Labels:  nil,
					Status:  api.HealthPassing,
				},
			},
			defaultRule: "Host(`foo.bar`)",
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
							Service: "Test",
							Rule:    "Host(`foo.bar`)",
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
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "default rule with label",
			items: []itemData{
				{
					ID:      "id",
					Node:    "Node1",
					Name:    "Test",
					Address: "127.0.0.1",
					Port:    "80",
					Labels: map[string]string{
						"traefik.domain": "foo.bar",
					},
					Status: api.HealthPassing,
				},
			},
			defaultRule: `Host("{{ .Name }}.{{ index .Labels "traefik.domain" }}")`,
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
							Service: "Test",
							Rule:    `Host("Test.foo.bar")`,
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
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "invalid rule",
			items: []itemData{
				{
					ID:      "Test",
					Node:    "Node1",
					Name:    "Test",
					Labels:  map[string]string{},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
			},
			defaultRule: `Host("{{ .Toto }}")`,
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
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "undefined rule",
			items: []itemData{
				{
					ID:      "Test",
					Node:    "Node1",
					Name:    "Test",
					Labels:  map[string]string{},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
			},
			defaultRule: ``,
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
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "default template rule",
			items: []itemData{
				{
					ID:      "Test",
					Node:    "Node1",
					Name:    "Test",
					Labels:  map[string]string{},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
			},
			defaultRule: defaultTemplateRule,
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
							Service: "Test",
							Rule:    "Host(`Test`)",
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
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{
				Configuration: Configuration{
					ExposedByDefault: true,
					DefaultRule:      test.defaultRule,
				},
			}

			err := p.Init()
			require.NoError(t, err)

			for i := 0; i < len(test.items); i++ {
				var err error
				test.items[i].ExtraConf, err = p.getExtraConf(test.items[i].Labels)
				require.NoError(t, err)
			}

			configuration := p.buildConfiguration(context.Background(), test.items, nil)

			assert.Equal(t, test.expected, configuration)
		})
	}
}

func Test_buildConfiguration(t *testing.T) {
	testCases := []struct {
		desc         string
		items        []itemData
		constraints  string
		ConnectAware bool
		expected     *dynamic.Configuration
	}{
		{
			desc: "one container no label",
			items: []itemData{
				{
					ID:      "Test",
					Node:    "Node1",
					Name:    "dev/Test",
					Labels:  map[string]string{},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
						"dev-Test": {
							Service: "dev-Test",
							Rule:    "Host(`dev-Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"dev-Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:         "one connect container",
			ConnectAware: true,
			items: []itemData{
				{
					ID:         "Test",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "dev/Test",
					Namespace:  "ns",
					Address:    "127.0.0.1",
					Port:       "443",
					Status:     api.HealthPassing,
					Labels: map[string]string{
						"traefik.consulcatalog.connect": "true",
					},
					Tags: nil,
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Services:    map[string]*dynamic.TCPService{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"dev-Test": {
							Service: "dev-Test",
							Rule:    "Host(`dev-Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"dev-Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "https://127.0.0.1:443",
									},
								},
								PassHostHeader:   Bool(true),
								ServersTransport: "tls-ns-dc1-dev-Test",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"tls-ns-dc1-dev-Test": {
							ServerName:         "ns-dc1-dev/Test",
							InsecureSkipVerify: true,
							RootCAs: []tls.FileOrContent{
								"root",
							},
							Certificates: []tls.Certificate{
								{
									CertFile: "cert",
									KeyFile:  "key",
								},
							},
							PeerCertURI: "spiffe:///ns/ns/dc/dc1/svc/dev/Test",
						},
					},
				},
			},
		},
		{
			desc:         "two connect containers on same service",
			ConnectAware: true,
			items: []itemData{
				{
					ID:         "Test1",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "dev/Test",
					Namespace:  "ns",
					Address:    "127.0.0.1",
					Port:       "443",
					Status:     api.HealthPassing,
					Labels: map[string]string{
						"traefik.consulcatalog.connect": "true",
					},
					Tags: nil,
				},
				{
					ID:         "Test2",
					Node:       "Node2",
					Datacenter: "dc1",
					Name:       "dev/Test",
					Namespace:  "ns",
					Address:    "127.0.0.2",
					Port:       "444",
					Status:     api.HealthPassing,
					Labels: map[string]string{
						"traefik.consulcatalog.connect": "true",
					},
					Tags: nil,
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Services:    map[string]*dynamic.TCPService{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"dev-Test": {
							Service: "dev-Test",
							Rule:    "Host(`dev-Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"dev-Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "https://127.0.0.1:443",
									},
									{
										URL: "https://127.0.0.2:444",
									},
								},
								PassHostHeader:   Bool(true),
								ServersTransport: "tls-ns-dc1-dev-Test",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"tls-ns-dc1-dev-Test": {
							ServerName:         "ns-dc1-dev/Test",
							InsecureSkipVerify: true,
							RootCAs: []tls.FileOrContent{
								"root",
							},
							Certificates: []tls.Certificate{
								{
									CertFile: "cert",
									KeyFile:  "key",
								},
							},
							PeerCertURI: "spiffe:///ns/ns/dc/dc1/svc/dev/Test",
						},
					},
				},
			},
		},
		{
			desc: "two containers no label",
			items: []itemData{
				{
					ID:      "Test",
					Node:    "Node1",
					Name:    "Test",
					Labels:  map[string]string{},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:      "Test2",
					Node:    "Node1",
					Name:    "Test2",
					Labels:  map[string]string{},
					Address: "127.0.0.2",
					Port:    "80",
					Status:  api.HealthPassing,
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
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
						"Test2": {
							Service: "Test2",
							Rule:    "Host(`Test2.traefik.wtf`)",
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
								PassHostHeader: Bool(true),
							},
						},
						"Test2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "two containers with same service name no label",
			items: []itemData{
				{
					ID:      "1",
					Node:    "Node1",
					Name:    "Test",
					Labels:  map[string]string{},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:      "2",
					Node:    "Node1",
					Name:    "Test",
					Labels:  map[string]string{},
					Address: "127.0.0.2",
					Port:    "80",
					Status:  api.HealthPassing,
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
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
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
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "two containers with same service name & id no label on same node",
			items: []itemData{
				{
					ID:      "1",
					Node:    "Node1",
					Name:    "Test",
					Labels:  map[string]string{},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:      "1",
					Node:    "Node1",
					Name:    "Test",
					Labels:  map[string]string{},
					Address: "127.0.0.2",
					Port:    "80",
					Status:  api.HealthPassing,
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
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "two containers with same service name & id no label on different nodes",
			items: []itemData{
				{
					ID:      "1",
					Node:    "Node1",
					Name:    "Test",
					Labels:  map[string]string{},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:      "1",
					Node:    "Node2",
					Name:    "Test",
					Labels:  map[string]string{},
					Address: "127.0.0.2",
					Port:    "80",
					Status:  api.HealthPassing,
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
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
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
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "one container with label (not on server)",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
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
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "one container with labels",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
						"traefik.http.routers.Router1.rule":                          "Host(`foo.com`)",
						"traefik.http.routers.Router1.service":                       "Service1",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "one container with rule label",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					Routers: map[string]*dynamic.Router{
						"Router1": {
							Service: "Test",
							Rule:    "Host(`foo.com`)",
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "one container with rule label and one service",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule":                          "Host(`foo.com`)",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "one container with rule label and two services",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule":                          "Host(`foo.com`)",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
						"traefik.http.services.Service2.loadbalancer.passhostheader": "true",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
						"Service2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "two containers with same service name and different passhostheader",
			items: []itemData{
				{
					ID:   "1",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:   "2",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "false",
					}, Address: "127.0.0.2",
					Port:   "80",
					Status: api.HealthPassing,
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
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "three containers with same service name and different passhostheader",
			items: []itemData{
				{
					ID:   "1",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "false",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:   "2",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:   "3",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					Address: "127.0.0.2",
					Port:    "80",
					Status:  api.HealthPassing,
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
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "two containers with same service name and same LB methods",
			items: []itemData{
				{
					ID:   "1",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:   "2",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					Address: "127.0.0.2",
					Port:    "80",
					Status:  api.HealthPassing,
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
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
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
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "one container with InFlightReq in label (default value)",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
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
								PassHostHeader: Bool(true),
							},
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"Middleware1": {
							InFlightReq: &dynamic.InFlightReq{
								Amount: 42,
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "two services with two identical middlewares",
			items: []itemData{
				{
					ID:   "1",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:   "2",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					}, Address: "127.0.0.2",
					Port:   "80",
					Status: api.HealthPassing,
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
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{
						"Middleware1": {
							InFlightReq: &dynamic.InFlightReq{
								Amount: 42,
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
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "two containers with two different middlewares with same name",
			items: []itemData{
				{
					ID:   "1",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:   "2",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "41",
					}, Address: "127.0.0.2",
					Port:   "80",
					Status: api.HealthPassing,
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
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
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
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "three containers with different middlewares with same name",
			items: []itemData{
				{
					ID:   "1",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID: "2",

					Name: "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "41",
					},
					Address: "127.0.0.2",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID: "3",

					Name: "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "40",
					},
					Address: "127.0.0.3",
					Port:    "80",
					Status:  api.HealthPassing,
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
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
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
									{
										URL: "http://127.0.0.2:80",
									},
									{
										URL: "http://127.0.0.3:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "two containers with two different routers with same name",
			items: []itemData{
				{
					ID: "1",

					Name: "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID: "2",

					Name: "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`bar.com`)",
					}, Address: "127.0.0.2",
					Port:   "80",
					Status: api.HealthPassing,
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
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "three containers with different routers with same name",
			items: []itemData{
				{
					ID: "1",

					Name: "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID: "2",

					Name: "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`bar.com`)",
					}, Address: "127.0.0.2",
					Port:   "80",
					Status: api.HealthPassing,
				},
				{
					ID: "3",

					Name: "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foobar.com`)",
					},
					Address: "127.0.0.3",
					Port:    "80",
					Status:  api.HealthPassing,
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
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
									{
										URL: "http://127.0.0.2:80",
									},
									{
										URL: "http://127.0.0.3:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "two containers with two identical routers",
			items: []itemData{
				{
					ID: "1",

					Name: "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID: "2",

					Name: "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					Address: "127.0.0.2",
					Port:    "80",
					Status:  api.HealthPassing,
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
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "one container with bad label",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.wrong.label": "42",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
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
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "one container with label port",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.scheme": "h2c",
						"traefik.http.services.Service1.LoadBalancer.server.port":   "8080",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "h2c://127.0.0.1:8080",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "one container with label port on two services",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.port": "",
						"traefik.http.services.Service2.LoadBalancer.server.port": "8080",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
						"Service2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:8080",
									},
								},
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "one container without port",
			items: []itemData{
				{
					ID:      "Test",
					Node:    "Node1",
					Name:    "Test",
					Labels:  map[string]string{},
					Address: "127.0.0.2",
					Status:  api.HealthPassing,
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
			},
		},
		{
			desc: "one container without port with middleware",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					},
					Address: "127.0.0.2",
					Status:  api.HealthPassing,
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
			},
		},
		{
			desc: "one container with traefik.enable false",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.enable": "false",
					},
					Address: "127.0.0.1",
					Port:    "80",
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
			},
		},
		{
			desc: "one container not healthy",
			items: []itemData{
				{
					ID:      "Test",
					Node:    "Node1",
					Name:    "Test",
					Labels:  map[string]string{},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthCritical,
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
			},
		},
		{
			desc: "one container with non matching constraints",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.tags": "foo",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
			},
			constraints: `Tag("traefik.tags=bar")`,
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
			},
		},
		{
			desc: "one container with matching constraints",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.tags": "foo",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
			},
			constraints: `Tag("traefik.tags=foo")`,
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
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
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
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "Middlewares used in router",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.basicauth.users": "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						"traefik.http.routers.Test.middlewares":                "Middleware1",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "Middlewares used in TCP router",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.Test.rule":                               "HostSNI(`foo.bar`)",
						"traefik.tcp.middlewares.Middleware1.ipwhitelist.sourcerange": "foobar, fiibar",
						"traefik.tcp.routers.Test.middlewares":                        "Middleware1",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
								TerminationDelay: Int(100),
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
			},
		},
		{
			desc: "tcp with label",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule": "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls":  "true",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
								TerminationDelay: Int(100),
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
			},
		},
		{
			desc: "udp with label",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.udp.routers.foo.entrypoints": "mydns",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
			},
		},
		{
			desc: "tcp with label without rule",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.tls": "true",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
								TerminationDelay: Int(100),
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
			},
		},
		{
			desc: "tcp with label and port",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule":                      "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls.options":               "foo",
						"traefik.tcp.services.foo.loadbalancer.server.port": "8080",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {
							Service: "foo",
							Rule:    "HostSNI(`foo.bar`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Options: "foo",
							},
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
								TerminationDelay: Int(100),
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
			},
		},
		{
			desc: "udp with label and port",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.udp.routers.foo.entrypoints":               "mydns",
						"traefik.udp.services.foo.loadbalancer.server.port": "80",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
			},
		},
		{
			desc: "tcp with label and port and http service",
			items: []itemData{
				{
					ID:   "1",
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule":                               "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls":                                "true",
						"traefik.tcp.services.foo.loadbalancer.server.port":          "80",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:   "2",
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule":                               "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls":                                "true",
						"traefik.tcp.services.foo.loadbalancer.server.port":          "80",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					Address: "127.0.0.2",
					Port:    "80",
					Status:  api.HealthPassing,
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
										Address: "127.0.0.1:80",
									},
									{
										Address: "127.0.0.2:80",
									},
								},
								TerminationDelay: Int(100),
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
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
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
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "udp with label and port and http service",
			items: []itemData{
				{
					ID:   "1",
					Name: "Test",
					Labels: map[string]string{
						"traefik.udp.routers.foo.entrypoints":                        "mydns",
						"traefik.udp.services.foo.loadbalancer.server.port":          "80",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:   "2",
					Name: "Test",
					Labels: map[string]string{
						"traefik.udp.routers.foo.entrypoints":                        "mydns",
						"traefik.udp.services.foo.loadbalancer.server.port":          "80",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					Address: "127.0.0.2",
					Port:    "80",
					Status:  api.HealthPassing,
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
										Address: "127.0.0.1:80",
									},
									{
										Address: "127.0.0.2:80",
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
							Service: "Service1",
							Rule:    "Host(`Test.traefik.wtf`)",
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
								PassHostHeader: Bool(true),
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc: "tcp with label for tcp service",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.services.foo.loadbalancer.server.port": "80",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
										Address: "127.0.0.1:80",
									},
								},
								TerminationDelay: Int(100),
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
			},
		},
		{
			desc: "udp with label for tcp service",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.udp.services.foo.loadbalancer.server.port": "80",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
			},
		},
		{
			desc: "tcp with label for tcp service, with termination delay",
			items: []itemData{
				{
					ID:   "Test",
					Name: "Test",
					Labels: map[string]string{
						"traefik.tcp.services.foo.loadbalancer.server.port":      "80",
						"traefik.tcp.services.foo.loadbalancer.terminationdelay": "200",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
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
										Address: "127.0.0.1:80",
									},
								},
								TerminationDelay: Int(200),
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
			},
		},
		{
			desc:         "two HTTP service instances with one canary",
			ConnectAware: true,
			items: []itemData{
				{
					ID:         "1",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "Test",
					Namespace:  "ns",
					Labels: map[string]string{
						"traefik.consulcatalog.connect": "true",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:         "2",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "Test",
					Namespace:  "ns",
					Labels: map[string]string{
						"traefik.consulcatalog.connect": "true",
						"traefik.consulcatalog.canary":  "true",
					},
					Address: "127.0.0.2",
					Port:    "80",
					Status:  api.HealthPassing,
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
							Service: "Test",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
						"Test-97077516270503695": {
							Service: "Test-97077516270503695",
							Rule:    "Host(`Test.traefik.wtf`)",
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "https://127.0.0.1:80",
									},
								},
								PassHostHeader:   Bool(true),
								ServersTransport: "tls-ns-dc1-Test",
							},
						},
						"Test-97077516270503695": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Servers: []dynamic.Server{
									{
										URL: "https://127.0.0.2:80",
									},
								},
								PassHostHeader:   Bool(true),
								ServersTransport: "tls-ns-dc1-Test",
							},
						},
					},
					ServersTransports: map[string]*dynamic.ServersTransport{
						"tls-ns-dc1-Test": {
							ServerName:         "ns-dc1-Test",
							InsecureSkipVerify: true,
							RootCAs: []tls.FileOrContent{
								"root",
							},
							Certificates: []tls.Certificate{
								{
									CertFile: "cert",
									KeyFile:  "key",
								},
							},
							PeerCertURI: "spiffe:///ns/ns/dc/dc1/svc/Test",
						},
					},
				},
			},
		},
		{
			desc:         "two TCP service instances with one canary",
			ConnectAware: true,
			items: []itemData{
				{
					ID:         "1",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "Test",
					Namespace:  "ns",
					Labels: map[string]string{
						"traefik.tcp.routers.test.rule": "HostSNI(`foobar`)",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:         "2",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "Test",
					Namespace:  "ns",
					Labels: map[string]string{
						"traefik.consulcatalog.canary":         "true",
						"traefik.tcp.routers.test-canary.rule": "HostSNI(`canary.foobar`)",
					},
					Address: "127.0.0.2",
					Port:    "80",
					Status:  api.HealthPassing,
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"test": {
							Service: "Test",
							Rule:    "HostSNI(`foobar`)",
						},
						"test-canary": {
							Service: "Test-17573747155436217342",
							Rule:    "HostSNI(`canary.foobar`)",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"Test": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{Address: "127.0.0.1:80"},
								},
								TerminationDelay: Int(100),
							},
						},
						"Test-17573747155436217342": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{Address: "127.0.0.2:80"},
								},
								TerminationDelay: Int(100),
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
			},
		},
		{
			desc:         "two UDP service instances with one canary",
			ConnectAware: true,
			items: []itemData{
				{
					ID:         "1",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "Test",
					Namespace:  "ns",
					Labels: map[string]string{
						"traefik.udp.routers.test.entrypoints": "udp",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
				{
					ID:         "2",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "Test",
					Namespace:  "ns",
					Labels: map[string]string{
						"traefik.consulcatalog.canary":                "true",
						"traefik.udp.routers.test-canary.entrypoints": "udp",
					},
					Address: "127.0.0.2",
					Port:    "80",
					Status:  api.HealthPassing,
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"test": {
							EntryPoints: []string{"udp"},
							Service:     "Test",
						},
						"test-canary": {
							EntryPoints: []string{"udp"},
							Service:     "Test-12825244908842506376",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"Test": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{Address: "127.0.0.1:80"},
								},
							},
						},
						"Test-12825244908842506376": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{Address: "127.0.0.2:80"},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
		{
			desc:         "UDP service with labels only",
			ConnectAware: true,
			items: []itemData{
				{
					ID:         "1",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "Test",
					Namespace:  "ns",
					Labels: map[string]string{
						"traefik.udp.routers.test-udp-label.service":                           "test-udp-label-service",
						"traefik.udp.routers.test-udp-label.entryPoints":                       "udp",
						"traefik.udp.services.test-udp-label-service.loadBalancer.server.port": "21116",
					},
					Address: "127.0.0.1",
					Port:    "80",
					Status:  api.HealthPassing,
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:     map[string]*dynamic.TCPRouter{},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services:    map[string]*dynamic.TCPService{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"test-udp-label": {
							EntryPoints: []string{"udp"},
							Service:     "test-udp-label-service",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"test-udp-label-service": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{
								Servers: []dynamic.UDPServer{
									{Address: "127.0.0.1:21116"},
								},
							},
						},
					},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers:           map[string]*dynamic.Router{},
					Middlewares:       map[string]*dynamic.Middleware{},
					Services:          map[string]*dynamic.Service{},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
			},
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			p := Provider{
				Configuration: Configuration{
					ExposedByDefault: true,
					DefaultRule:      "Host(`{{ normalize .Name }}.traefik.wtf`)",
					ConnectAware:     test.ConnectAware,
					Constraints:      test.constraints,
				},
			}

			err := p.Init()
			require.NoError(t, err)

			for i := 0; i < len(test.items); i++ {
				var err error
				test.items[i].ExtraConf, err = p.getExtraConf(test.items[i].Labels)
				require.NoError(t, err)

				var tags []string
				for k, v := range test.items[i].Labels {
					tags = append(tags, fmt.Sprintf("%s=%s", k, v))
				}
				test.items[i].Tags = tags
			}

			configuration := p.buildConfiguration(context.Background(), test.items, &connectCert{
				root: []string{"root"},
				leaf: keyPair{
					cert: "cert",
					key:  "key",
				},
			})

			assert.Equal(t, test.expected, configuration)
		})
	}
}

func TestNamespaces(t *testing.T) {
	testCases := []struct {
		desc               string
		namespace          string
		namespaces         []string
		expectedNamespaces []string
	}{
		{
			desc:               "no defined namespaces",
			expectedNamespaces: []string{""},
		},
		{
			desc:               "deprecated: use of defined namespace",
			namespace:          "test-ns",
			expectedNamespaces: []string{"test-ns"},
		},
		{
			desc:               "use of 1 defined namespaces",
			namespaces:         []string{"test-ns"},
			expectedNamespaces: []string{"test-ns"},
		},
		{
			desc:               "use of multiple defined namespaces",
			namespaces:         []string{"test-ns1", "test-ns2", "test-ns3", "test-ns4"},
			expectedNamespaces: []string{"test-ns1", "test-ns2", "test-ns3", "test-ns4"},
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			pb := &ProviderBuilder{
				Namespace:  test.namespace,
				Namespaces: test.namespaces,
			}

			assert.Equal(t, test.expectedNamespaces, extractNSFromProvider(pb.BuildProviders()))
		})
	}
}

func extractNSFromProvider(providers []*Provider) []string {
	res := make([]string, len(providers))
	for i, p := range providers {
		res[i] = p.namespace
	}
	return res
}
