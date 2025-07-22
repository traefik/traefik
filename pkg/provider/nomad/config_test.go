package nomad

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

func Test_defaultRule(t *testing.T) {
	testCases := []struct {
		desc     string
		items    []item
		rule     string
		expected *dynamic.Configuration
	}{
		{
			desc: "default rule with no variable",
			items: []item{
				{
					ID:        "id",
					Node:      "node1",
					Name:      "Test",
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			rule: "Host(`example.com`)",
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`example.com`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "default rule with label",
			items: []item{
				{
					ID:      "id",
					Node:    "Node1",
					Name:    "Test",
					Address: "127.0.0.1",
					Tags: []string{
						"traefik.domain=example.com",
					},
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			rule: `Host("{{ .Name }}.{{ index .Labels "traefik.domain" }}")`,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        `Host("Test.example.com")`,
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "invalid rule",
			items: []item{
				{
					ID:        "id",
					Node:      "Node1",
					Name:      "Test",
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			rule: `Host"{{ .Invalid }}")`,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "default template rule",
			items: []item{
				{
					ID:        "id",
					Node:      "Node1",
					Name:      "Test",
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			rule: defaultTemplateRule,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p := new(Provider)
			p.SetDefaults()
			p.DefaultRule = test.rule
			err := p.Init()
			require.NoError(t, err)

			config := p.buildConfig(t.Context(), test.items)
			require.Equal(t, test.expected, config)
		})
	}
}

func Test_buildConfig(t *testing.T) {
	testCases := []struct {
		desc        string
		items       []item
		constraints string
		expected    *dynamic.Configuration
	}{
		{
			desc: "one service no tags",
			items: []item{
				{
					ID:        "id",
					Node:      "Node1",
					Name:      "dev/Test",
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"dev-Test": {
							Service:     "dev-Test",
							Rule:        "Host(`dev-Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"dev-Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "two services no tags",
			items: []item{
				{
					ID:        "id1",
					Node:      "Node1",
					Name:      "Test1",
					Address:   "192.168.1.101",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:        "id2",
					Node:      "Node2",
					Name:      "Test2",
					Address:   "192.168.1.102",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test1": {
							Service:     "Test1",
							Rule:        "Host(`Test1.traefik.test`)",
							DefaultRule: true,
						},
						"Test2": {
							Service:     "Test2",
							Rule:        "Host(`Test2.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://192.168.1.101:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"Test2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://192.168.1.102:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "two services with same name no label",
			items: []item{
				{
					ID:        "id1",
					Node:      "Node1",
					Name:      "Test",
					Tags:      []string{},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:        "id2",
					Node:      "Node2",
					Name:      "Test",
					Tags:      []string{},
					Address:   "127.0.0.2",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
									{
										URL: "http://127.0.0.2:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "two services same name and id no label same node",
			items: []item{
				{
					ID:        "id1",
					Node:      "Node1",
					Name:      "Test",
					Tags:      []string{},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:        "id1",
					Node:      "Node1",
					Name:      "Test",
					Tags:      []string{},
					Address:   "127.0.0.2",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.2:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "two services same service name and id no label on different nodes",
			items: []item{
				{
					ID:        "id1",
					Node:      "Node1",
					Name:      "Test",
					Tags:      []string{},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:        "id1",
					Node:      "Node2",
					Name:      "Test",
					Tags:      []string{},
					Address:   "127.0.0.2",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
									{
										URL: "http://127.0.0.2:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "one service with label (not on server)",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.loadbalancer.passhostheader=true",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "one service with labels",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.loadbalancer.passhostheader = true",
						"traefik.http.routers.Router1.rule = Host(`foo.com`)",
						"traefik.http.routers.Router1.service = Service1",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "empty service",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.enable=true",
					},
					Address:   "",
					Port:      -1,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "one service with rule label",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.routers.Router1.rule = Host(`foo.com`)",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one service with rule label and one traefik service",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.routers.Router1.rule = Host(`foo.com`)",
						"traefik.http.services.Service1.loadbalancer.passhostheader = true",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "one service with rule label and two traefik services",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.routers.Router1.rule = Host(`foo.com`)",
						"traefik.http.services.Service1.loadbalancer.passhostheader= true",
						"traefik.http.services.Service2.loadbalancer.passhostheader = true",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"Service2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "two services with same traefik service and different passhostheader",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.loadbalancer.passhostheader = true",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:   "id2",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.loadbalancer.passhostheader = false",
					},
					Address:   "127.0.0.2",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
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
			desc: "three services with same name and different passhostheader",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.loadbalancer.passhostheader = false",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:   "id2",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.loadbalancer.passhostheader = true",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:   "id3",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.loadbalancer.passhostheader = true",
					},
					Address:   "127.0.0.2",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
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
			desc: "two services with same name and same LB methods",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.loadbalancer.passhostheader = true",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:   "id2",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.loadbalancer.passhostheader = true",
					},
					Address:   "127.0.0.2",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
									{
										URL: "http://127.0.0.2:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "one service with InFlightReq in label (default value)",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount = 42",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "two services with same middleware",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount = 42",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:   "id2",
					Name: "Test",
					Tags: []string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount = 42",
					},
					Address:   "127.0.0.2",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
									{
										URL: "http://127.0.0.2:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "two services same name with different middleware",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount = 42",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:   "id2",
					Name: "Test",
					Tags: []string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount = 41",
					},
					Address:   "127.0.0.2",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
									{
										URL: "http://127.0.0.2:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "two services with different routers with same name",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.routers.Router1.rule = Host(`foo.com`)",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:   "id2",
					Name: "Test",
					Tags: []string{
						"traefik.http.routers.Router1.rule = Host(`bar.com`)",
					},
					Address:   "127.0.0.2",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
									{
										URL: "http://127.0.0.2:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "two services identical routers",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.routers.Router1.rule = Host(`foo.com`)",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:   "id2",
					Name: "Test",
					Tags: []string{
						"traefik.http.routers.Router1.rule = Host(`foo.com`)",
					},
					Address:   "127.0.0.2",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
									{
										URL: "http://127.0.0.2:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "one service with bad label",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.wrong.label = 42",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "one service with label port",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.LoadBalancer.server.scheme = h2c",
						"traefik.http.services.Service1.LoadBalancer.server.port = 8080",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "h2c://127.0.0.1:8080",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "one service with label url",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.LoadBalancer.server.url = http://1.2.3.4:5678",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://1.2.3.4:5678",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "one service with label url and preserve path",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.LoadBalancer.server.url = http://1.2.3.4:5678",
						"traefik.http.services.Service1.LoadBalancer.server.preservepath = true",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL:          "http://1.2.3.4:5678",
										PreservePath: true,
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "one service with label url and port",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.LoadBalancer.server.url = http://1.2.3.4:5678",
						"traefik.http.services.Service1.LoadBalancer.server.port = 1234",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "one service with label url and scheme",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.LoadBalancer.server.url = http://1.2.3.4:5678",
						"traefik.http.services.Service1.LoadBalancer.server.scheme = https",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "one service with label port on two services",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.services.Service1.LoadBalancer.server.port = ",
						"traefik.http.services.Service2.LoadBalancer.server.port = 8080",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"Service2": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:8080",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "one service without port",
			items: []item{
				{
					ID:        "id1",
					Node:      "Node1",
					Name:      "Test",
					Tags:      []string{},
					Address:   "127.0.0.2",
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "one service without port with middleware",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount = 42",
					},
					Address:   "127.0.0.2",
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "one service with traefik.enable false",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.enable=false",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: false},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "one service with non matching constraints",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.tags=foo",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			constraints: `Tag("traefik.tags=bar")`,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.tags=foo",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			constraints: `Tag("traefik.tags=foo")`,
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "middleware used in router",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.http.middlewares.Middleware1.basicauth.users = test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						"traefik.http.routers.Test.middlewares = Middleware1",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.test`)",
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "middleware used in tcp router",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.tcp.routers.Test.rule = HostSNI(`foo.bar`)",
						"traefik.tcp.middlewares.Middleware1.ipallowlist.sourcerange = foobar, fiibar",
						"traefik.tcp.routers.Test.middlewares = Middleware1",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
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
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"foobar", "fiibar"},
							},
						},
					},
					Services: map[string]*dynamic.TCPService{
						"Test": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:9999",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "tcp with tags",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.tcp.routers.foo.rule = HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls = true",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
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
										Address: "127.0.0.1:9999",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "udp with tags",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.udp.routers.foo.entrypoints = mydns",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
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
										Address: "127.0.0.1:9999",
									},
								},
							},
						},
					},
				},
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.tcp.routers.foo.tls = true",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
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
										Address: "127.0.0.1:9999",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "tcp with tags and port",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.tcp.routers.foo.rule = HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls.options = foo",
						"traefik.tcp.services.foo.loadbalancer.server.port = 80",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
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
										Address: "127.0.0.1:80",
									},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.udp.routers.foo.entrypoints = mydns",
						"traefik.udp.services.foo.loadbalancer.server.port = 80",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.tcp.routers.foo.rule = HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls = true",
						"traefik.tcp.services.foo.loadbalancer.server.port = 80",
						"traefik.http.services.Service1.loadbalancer.passhostheader = true",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:   "id2",
					Name: "Test",
					Tags: []string{
						"traefik.tcp.routers.foo.rule = HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls = true",
						"traefik.tcp.services.foo.loadbalancer.server.port = 80",
						"traefik.http.services.Service1.loadbalancer.passhostheader = true",
					},
					Address:   "127.0.0.2",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
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
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
									{
										URL: "http://127.0.0.2:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "udp with label and port and http services",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.udp.routers.foo.entrypoints = mydns",
						"traefik.udp.services.foo.loadbalancer.server.port = 80",
						"traefik.http.services.Service1.loadbalancer.passhostheader = true",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:   "id2",
					Name: "Test",
					Tags: []string{
						"traefik.udp.routers.foo.entrypoints = mydns",
						"traefik.udp.services.foo.loadbalancer.server.port = 80",
						"traefik.http.services.Service1.loadbalancer.passhostheader = true",
					},
					Address:   "127.0.0.2",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
									{
										URL: "http://127.0.0.2:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "tcp with tag for tcp service",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.tcp.services.foo.loadbalancer.server.port = 80",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
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
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.udp.services.foo.loadbalancer.server.port = 80",
					},
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			// TODO: replace or delete?
			desc: "tcp with label for tcp service, with termination delay",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.tcp.services.foo.loadbalancer.server.port = 80",
					},
					Address:   "127.0.0.1",
					Port:      80,
					ExtraConf: configuration{Enable: true},
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
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "two HTTP service instances with one canary",
			items: []item{
				{
					ID:         "1",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "Test",
					Namespace:  "ns",
					Tags:       []string{},
					Address:    "127.0.0.1",
					Port:       80,
					ExtraConf:  configuration{Enable: true},
				},
				{
					ID:         "2",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "Test",
					Namespace:  "ns",
					Tags: []string{
						"traefik.nomad.canary = true",
					},
					Address: "127.0.0.2",
					Port:    80,
					ExtraConf: configuration{
						Enable: true,
						Canary: true,
					},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
						"Test-1234154071633021619": {
							Service:     "Test-1234154071633021619",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
							},
						},
						"Test-1234154071633021619": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.2:80",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "two TCP service instances with one canary",
			items: []item{
				{
					ID:         "1",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "Test",
					Namespace:  "ns",
					Tags: []string{
						"traefik.tcp.routers.test.rule = HostSNI(`foobar`)",
					},
					Address:   "127.0.0.1",
					Port:      80,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:         "2",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "Test",
					Namespace:  "ns",
					Tags: []string{
						"traefik.nomad.canary = true",
						"traefik.tcp.routers.test-canary.rule = HostSNI(`canary.foobar`)",
					},
					Address: "127.0.0.2",
					Port:    80,
					ExtraConf: configuration{
						Enable: true,
						Canary: true,
					},
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
							Service: "Test-8769860286750522282",
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
							},
						},
						"Test-8769860286750522282": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{Address: "127.0.0.2:80"},
								},
							},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "two UDP service instances with one canary",
			items: []item{
				{
					ID:         "1",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "Test",
					Namespace:  "ns",
					Tags: []string{
						"traefik.udp.routers.test.entrypoints = udp",
					},
					Address:   "127.0.0.1",
					Port:      80,
					ExtraConf: configuration{Enable: true},
				},
				{
					ID:         "2",
					Node:       "Node1",
					Datacenter: "dc1",
					Name:       "Test",
					Namespace:  "ns",
					Tags: []string{
						"traefik.nomad.canary = true",
						"traefik.udp.routers.test-canary.entrypoints = udp",
					},
					Address: "127.0.0.2",
					Port:    80,
					ExtraConf: configuration{
						Enable: true,
						Canary: true,
					},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"test": {
							EntryPoints: []string{"udp"},
							Service:     "Test",
						},
						"test-canary": {
							EntryPoints: []string{"udp"},
							Service:     "Test-1611429260986126224",
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
						"Test-1611429260986126224": {
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
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one service with default generated certificate tags",
			items: []item{
				{
					ID:        "id",
					Node:      "Node1",
					Name:      "dev/Test",
					Address:   "127.0.0.1",
					Port:      9999,
					ExtraConf: configuration{Enable: true},
					Tags: []string{
						"traefik.tls.stores.default.defaultgeneratedcert.resolver = foobar",
						"traefik.tls.stores.default.defaultgeneratedcert.domain.main = foobar",
						"traefik.tls.stores.default.defaultgeneratedcert.domain.sans = foobar, fiibar",
					},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"dev-Test": {
							Service:     "dev-Test",
							Rule:        "Host(`dev-Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"dev-Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:9999",
									},
								},
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			p := new(Provider)
			p.SetDefaults()
			p.DefaultRule = "Host(`{{ normalize .Name }}.traefik.test`)"
			p.Constraints = test.constraints
			err := p.Init()
			require.NoError(t, err)

			c := p.buildConfig(t.Context(), test.items)
			require.Equal(t, test.expected, c)
		})
	}
}

func Test_buildConfigAllowEmptyServicesTrue(t *testing.T) {
	testCases := []struct {
		desc        string
		items       []item
		constraints string
		expected    *dynamic.Configuration
	}{
		{
			desc: "empty service http",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.enable=true",
					},
					Address:   "",
					Port:      -1,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers:  map[string]*dynamic.UDPRouter{},
					Services: map[string]*dynamic.UDPService{},
				},
				HTTP: &dynamic.HTTPConfiguration{
					Routers: map[string]*dynamic.Router{
						"Test": {
							Service:     "Test",
							Rule:        "Host(`Test.traefik.test`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
								Servers:        nil,
								PassHostHeader: pointer(true),
								ResponseForwarding: &dynamic.ResponseForwarding{
									FlushInterval: ptypes.Duration(100 * time.Millisecond),
								},
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
			desc: "empty service tcp",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.tcp.routers.test.rule = HostSNI(`foobar`)",
					},
					Address:   "",
					Port:      -1,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"test": {
							Rule:    "HostSNI(`foobar`)",
							Service: "Test",
						},
					},
					Middlewares: map[string]*dynamic.TCPMiddleware{},
					Services: map[string]*dynamic.TCPService{
						"Test": {
							LoadBalancer: &dynamic.TCPServersLoadBalancer{},
						},
					},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "empty service udp",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.udp.routers.test.entrypoints = udp",
					},
					Address:   "",
					Port:      -1,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
				},
				UDP: &dynamic.UDPConfiguration{
					Routers: map[string]*dynamic.UDPRouter{
						"test": {
							EntryPoints: []string{"udp"},
							Service:     "Test",
						},
					},
					Services: map[string]*dynamic.UDPService{
						"Test": {
							LoadBalancer: &dynamic.UDPServersLoadBalancer{},
						},
					},
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p := new(Provider)
			p.SetDefaults()
			p.AllowEmptyServices = true
			p.DefaultRule = "Host(`{{ normalize .Name }}.traefik.test`)"
			p.Constraints = test.constraints
			err := p.Init()
			require.NoError(t, err)

			c := p.buildConfig(t.Context(), test.items)
			require.Equal(t, test.expected, c)
		})
	}
}

func Test_buildConfigAllowEmptyServicesFalseDefault(t *testing.T) {
	testCases := []struct {
		desc        string
		items       []item
		constraints string
		expected    *dynamic.Configuration
	}{
		{
			desc: "empty service http",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.enable=true",
					},
					Address:   "",
					Port:      -1,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "empty service tcp",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.tcp.routers.test.rule = HostSNI(`foobar`)",
					},
					Address:   "",
					Port:      -1,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
			desc: "empty service udp",
			items: []item{
				{
					ID:   "id1",
					Name: "Test",
					Tags: []string{
						"traefik.udp.routers.test.entrypoints = udp",
					},
					Address:   "",
					Port:      -1,
					ExtraConf: configuration{Enable: true},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p := new(Provider)
			p.SetDefaults()
			p.DefaultRule = "Host(`{{ normalize .Name }}.traefik.test`)"
			p.Constraints = test.constraints
			err := p.Init()
			require.NoError(t, err)

			c := p.buildConfig(t.Context(), test.items)
			require.Equal(t, test.expected, c)
		})
	}
}

func Test_keepItem(t *testing.T) {
	testCases := []struct {
		name        string
		i           item
		constraints string
		exp         bool
	}{
		{
			name: "enable true",
			i:    item{ExtraConf: configuration{Enable: true}},
			exp:  true,
		},
		{
			name: "enable false",
			i:    item{ExtraConf: configuration{Enable: false}},
			exp:  false,
		},
		{
			name: "constraint matches",
			i: item{
				Tags:      []string{"traefik.tags=foo"},
				ExtraConf: configuration{Enable: true},
			},
			constraints: `Tag("traefik.tags=foo")`,
			exp:         true,
		},
		{
			name: "constraint not match",
			i: item{
				Tags:      []string{"traefik.tags=foo"},
				ExtraConf: configuration{Enable: true},
			},
			constraints: `Tag("traefik.tags=bar")`,
			exp:         false,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			p := new(Provider)
			p.SetDefaults()
			p.Constraints = test.constraints

			result := p.keepItem(t.Context(), test.i)
			require.Equal(t, test.exp, result)
		})
	}
}

func TestNamespaces(t *testing.T) {
	testCases := []struct {
		desc               string
		namespaces         []string
		expectedNamespaces []string
	}{
		{
			desc:               "no defined namespaces",
			expectedNamespaces: []string{""},
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
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			pb := &ProviderBuilder{
				Namespaces: test.namespaces,
			}

			assert.Equal(t, test.expectedNamespaces, extractNamespacesFromProvider(pb.BuildProviders()))
		})
	}
}

func extractNamespacesFromProvider(providers []*Provider) []string {
	res := make([]string, len(providers))
	for i, p := range providers {
		res[i] = p.namespace
	}
	return res
}

func pointer[T any](v T) *T { return &v }
