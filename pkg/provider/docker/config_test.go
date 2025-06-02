package docker

import (
	"strconv"
	"testing"
	"time"

	docker "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/types"
)

func TestDynConfBuilder_DefaultRule(t *testing.T) {
	testCases := []struct {
		desc        string
		containers  []dockerData
		defaultRule string
		expected    *dynamic.Configuration
	}{
		{
			desc: "default rule with no variable",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			defaultRule: "Host(`foo.bar`)",
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
							Rule:        "Host(`foo.bar`)",
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
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "default rule with service name",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			defaultRule: "Host(`{{ .Name }}.foo.bar`)",
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
							Rule:        "Host(`Test.foo.bar`)",
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
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.domain": "foo.bar",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			defaultRule: `Host("{{ .Name }}.{{ index .Labels "traefik.domain" }}")`,
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
							Rule:        `Host("Test.foo.bar")`,
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
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			defaultRule: `Host("{{ .Toto }}")`,
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
										URL: "http://127.0.0.1:80",
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
			desc: "undefined rule",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			defaultRule: ``,
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
										URL: "http://127.0.0.1:80",
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
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			defaultRule: DefaultTemplateRule,
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
										URL: "http://127.0.0.1:80",
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
			t.Parallel()

			p := Provider{
				Shared: Shared{
					ExposedByDefault: true,
					DefaultRule:      test.defaultRule,
				},
			}
			require.NoError(t, p.Init())

			builder := NewDynConfBuilder(p.Shared, nil, false)

			for i := range len(test.containers) {
				var err error
				test.containers[i].ExtraConf, err = builder.extractLabels(test.containers[i])
				require.NoError(t, err)
			}

			configuration := builder.build(t.Context(), test.containers)

			assert.Equal(t, test.expected, configuration)
		})
	}
}

func TestDynConfBuilder_build(t *testing.T) {
	testCases := []struct {
		desc               string
		containers         []dockerData
		useBindPortIP      bool
		constraints        string
		expected           *dynamic.Configuration
		allowEmptyServices bool
	}{
		{
			desc: "invalid HTTP service definition",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.test": "",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
			desc: "invalid TCP service definition",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tcp.services.test": "",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
			desc: "invalid UDP service definition",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.udp.services.test": "",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/udp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
			desc: "one container no label",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
							Rule:        "Host(`Test.traefik.wtf`)",
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
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "two containers no label",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ServiceName: "Test2",
					Name:        "Test2",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
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
							Rule:        "Host(`Test.traefik.wtf`)",
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
						"Test2": {
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
			desc: "two containers with same service name no label",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
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
							Rule:        "Host(`Test.traefik.wtf`)",
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
			desc: "one container with label (not on server)",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
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
										URL: "http://127.0.0.1:80",
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
			desc: "one container with labels",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
						"traefik.http.routers.Router1.rule":                          "Host(`foo.com`)",
						"traefik.http.routers.Router1.service":                       "Service1",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
										URL: "http://127.0.0.1:80",
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
			desc: "one container with rule label",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
			desc: "one container with rule label and one service",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule":                          "Host(`foo.com`)",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
										URL: "http://127.0.0.1:80",
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
			desc: "one container with rule label and two services",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule":                          "Host(`foo.com`)",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
						"traefik.http.services.Service2.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
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
						"Service2": {
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
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one router, one specified but undefined service -> specified one is assigned, but automatic is created instead",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule":    "Host(`foo.com`)",
						"traefik.http.routers.Router1.service": "Service1",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
						"Router1": {
							Service: "Service1",
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
										URL: "http://127.0.0.1:80",
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
			desc: "two containers with same service name and different passhostheader",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "false",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
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
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
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
			desc: "three containers with same service name and different passhostheader",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "false",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
				{
					ID:          "3",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
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
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
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
			desc: "two containers with same service name and same LB methods",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
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
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
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
										URL: "http://127.0.0.1:80",
									},
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
			desc: "one container with InFlightReq in label (default value)",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
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
			desc: "two containers with two identical middlewares",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
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
							Rule:        "Host(`Test.traefik.wtf`)",
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
										URL: "http://127.0.0.1:80",
									},
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
			desc: "two containers with two different middlewares with same name",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "41",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
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
							Rule:        "Host(`Test.traefik.wtf`)",
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
			desc: "three containers with different middlewares with same name",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "41",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
				{
					ID:          "3",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "40",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.3",
							},
						},
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
							Rule:        "Host(`Test.traefik.wtf`)",
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
									{
										URL: "http://127.0.0.2:80",
									},
									{
										URL: "http://127.0.0.3:80",
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
			desc: "two containers with two different routers with same name",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`bar.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
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
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
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
			desc: "three containers with different routers with same name",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`bar.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
				},
				{
					ID:          "3",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foobar.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.3",
							},
						},
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
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy: dynamic.BalancerStrategyWRR,
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
			desc: "two containers with two identical routers",
			containers: []dockerData{
				{
					ID:          "1",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
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
										URL: "http://127.0.0.1:80",
									},
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
			desc: "two containers with two identical router rules and different service names",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ServiceName: "Test2",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.routers.Router1.rule": "Host(`foo.com`)",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
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
					Routers:     map[string]*dynamic.Router{},
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
						"Test2": {
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
			desc: "one container with bad label",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.wrong.label": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
							Rule:        "Host(`Test.traefik.wtf`)",
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
					},
					ServersTransports: map[string]*dynamic.ServersTransport{},
				},
				TLS: &dynamic.TLSConfiguration{
					Stores: map[string]tls.Store{},
				},
			},
		},
		{
			desc: "one container with label port",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.scheme": "h2c",
						"traefik.http.services.Service1.LoadBalancer.server.port":   "8080",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
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
			desc: "one container with label port on two services",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.port": "",
						"traefik.http.services.Service2.LoadBalancer.server.port": "8080",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
					Routers:     map[string]*dynamic.Router{},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Service1": {
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
			desc: "one container with label url",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.url": "http://1.2.3.4:5678",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("4567/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
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
			desc: "one container with label url and preserve path",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.url":          "http://1.2.3.4:5678",
						"traefik.http.services.Service1.LoadBalancer.server.preservepath": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("4567/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
							Service:     "Service1",
							Rule:        "Host(`Test.traefik.wtf`)",
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
			desc: "one container with label url and port",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.url":  "http://1.2.3.4:5678",
						"traefik.http.services.Service1.LoadBalancer.server.port": "1234",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("4567/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
			desc: "one container with label url and scheme",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Service1.LoadBalancer.server.url":    "http://1.2.3.4:5678",
						"traefik.http.services.Service1.LoadBalancer.server.scheme": "https",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("4567/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
			desc: "one container without port",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels:      map[string]string{},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
			desc: "one container without port with middleware",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.inflightreq.amount": "42",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
			desc: "one container with traefik.enable false",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.enable": "false",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
			desc: "one unhealthy HTTP container",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Health:      docker.Unhealthy,
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
			desc:               "one unhealthy HTTP container with allowEmptyServices",
			allowEmptyServices: true,
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Health:      docker.Unhealthy,
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
							Rule:        "Host(`Test.traefik.wtf`)",
							DefaultRule: true,
						},
					},
					Middlewares: map[string]*dynamic.Middleware{},
					Services: map[string]*dynamic.Service{
						"Test": {
							LoadBalancer: &dynamic.ServersLoadBalancer{
								Strategy:       dynamic.BalancerStrategyWRR,
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
			desc: "one unhealthy TCP container",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Health:      docker.Unhealthy,
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule": "HostSNI(`foo.bar`)",
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
			desc:               "one unhealthy TCP container with allowEmptyServices",
			allowEmptyServices: true,
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Health:      docker.Unhealthy,
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule": "HostSNI(`foo.bar`)",
					},
				},
			},
			expected: &dynamic.Configuration{
				TCP: &dynamic.TCPConfiguration{
					Routers: map[string]*dynamic.TCPRouter{
						"foo": {
							Service: "Test",
							Rule:    "HostSNI(`foo.bar`)",
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
			desc: "one unhealthy UDP container",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Health:      docker.Unhealthy,
					Labels: map[string]string{
						"traefik.udp.routers.foo": "true",
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
			desc:               "one unhealthy UDP container with allowEmptyServices",
			allowEmptyServices: true,
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.udp.routers.foo": "true",
					},
					Health: docker.Unhealthy,
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
						"foo": {
							Service: "Test",
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
		{
			desc: "one container with non matching constraints",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tags": "foo",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			constraints: `Label("traefik.tags", "bar")`,
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
			desc: "one container with matching constraints",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tags": "foo",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			constraints: `Label("traefik.tags", "foo")`,
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
							Rule:        "Host(`Test.traefik.wtf`)",
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
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.middlewares.Middleware1.basicauth.users": "test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/,test2:$apr1$d9hr9HBB$4HxwgUir3HP4EsggP/QNo0",
						"traefik.http.routers.Test.middlewares":                "Middleware1",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
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
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.Test.rule":                               "HostSNI(`foo.bar`)",
						"traefik.tcp.middlewares.Middleware1.ipallowlist.sourcerange": "foobar, fiibar",
						"traefik.tcp.routers.Test.middlewares":                        "Middleware1",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
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
			desc: "tcp with label",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule": "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls":  "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
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
			desc: "udp with label",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.udp.routers.foo.entrypoints": "mydns",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
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
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.tls": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
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
			desc: "tcp with label and port",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tcp.routers.foo.rule":                      "HostSNI(`foo.bar`)",
						"traefik.tcp.routers.foo.tls.options":               "foo",
						"traefik.tcp.services.foo.loadbalancer.server.port": "8080",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
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
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.udp.routers.foo.entrypoints":               "mydns",
						"traefik.udp.services.foo.loadbalancer.server.port": "8080",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
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
			desc: "udp with label and port and http service",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.udp.routers.foo.entrypoints":                        "mydns",
						"traefik.udp.services.foo.loadbalancer.server.port":          "8080",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
				{
					ID:          "2",
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.udp.routers.foo.entrypoints":                        "mydns",
						"traefik.udp.services.foo.loadbalancer.server.port":          "8080",
						"traefik.http.services.Service1.loadbalancer.passhostheader": "true",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.2",
							},
						},
					},
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
					Routers:           map[string]*dynamic.TCPRouter{},
					Middlewares:       map[string]*dynamic.TCPMiddleware{},
					Services:          map[string]*dynamic.TCPService{},
					ServersTransports: map[string]*dynamic.TCPServersTransport{},
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:80",
									},
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
			desc: "udp with label for tcp service",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.udp.services.foo.loadbalancer.server.port": "8080",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
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
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tcp.services.foo.loadbalancer.server.port": "8080",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("80/tcp"): []nat.PortBinding{},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
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
			desc: "useBindPortIP with LblPort |	ExtIp:ExtPort:LblPort => ExtIp:ExtPort",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.http.services.Test.loadbalancer.server.port": "80",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("79/tcp"): []nat.PortBinding{{
								HostIP:   "192.168.0.1",
								HostPort: "8080",
							}},
							nat.Port("80/tcp"): []nat.PortBinding{{
								HostIP:   "192.168.0.1",
								HostPort: "8081",
							}},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			useBindPortIP: true,
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
							Rule:        "Host(`Test.traefik.wtf`)",
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
										URL: "http://192.168.0.1:8081",
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
			desc: "one container with default generated certificate labels",
			containers: []dockerData{
				{
					ServiceName: "Test",
					Name:        "Test",
					Labels: map[string]string{
						"traefik.tls.stores.default.defaultgeneratedcert.resolver":    "foobar",
						"traefik.tls.stores.default.defaultgeneratedcert.domain.main": "foobar",
						"traefik.tls.stores.default.defaultgeneratedcert.domain.sans": "foobar, fiibar",
					},
					NetworkSettings: networkSettings{
						Ports: nat.PortMap{
							nat.Port("79/tcp"): []nat.PortBinding{{
								HostIP:   "192.168.0.1",
								HostPort: "8080",
							}},
							nat.Port("80/tcp"): []nat.PortBinding{{
								HostIP:   "192.168.0.1",
								HostPort: "8081",
							}},
						},
						Networks: map[string]*networkData{
							"bridge": {
								Name: "bridge",
								Addr: "127.0.0.1",
							},
						},
					},
				},
			},
			expected: &dynamic.Configuration{
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
								Strategy: dynamic.BalancerStrategyWRR,
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1:79",
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
				Shared: Shared{
					AllowEmptyServices: test.allowEmptyServices,
					ExposedByDefault:   true,
					UseBindPortIP:      test.useBindPortIP,
					DefaultRule:        "Host(`{{ normalize .Name }}.traefik.wtf`)",
				},
			}
			p.Constraints = test.constraints

			require.NoError(t, p.Init())

			builder := NewDynConfBuilder(p.Shared, nil, false)

			for i := range len(test.containers) {
				var err error
				test.containers[i].ExtraConf, err = builder.extractLabels(test.containers[i])
				require.NoError(t, err)
			}

			configuration := builder.build(t.Context(), test.containers)

			assert.Equal(t, test.expected, configuration)
		})
	}
}

func TestDynConfBuilder_getIPPort_docker(t *testing.T) {
	type expected struct {
		ip    string
		port  string
		error bool
	}

	testCases := []struct {
		desc       string
		container  docker.ContainerJSON
		serverPort string
		expected   expected
	}{
		{
			desc: "label traefik.port not set, no binding, falling back on the container's IP/Port",
			container: containerJSON(
				ports(nat.PortMap{
					"8080/tcp": {},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			expected: expected{
				ip:   "10.11.12.13",
				port: "8080",
			},
		},
		{
			desc: "label traefik.port not set, single binding with port only, falling back on the container's IP/Port",
			container: containerJSON(
				withNetwork("testnet", ipv4("10.11.12.13")),
				ports(nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostPort: "8082",
						},
					},
				}),
			),
			expected: expected{
				ip:   "10.11.12.13",
				port: "80",
			},
		},
		{
			desc: "label traefik.port not set, binding with ip:port should create a route to the bound ip:port",
			container: containerJSON(
				ports(nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "1.2.3.4",
							HostPort: "8081",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			expected: expected{
				ip:   "1.2.3.4",
				port: "8081",
			},
		},
		{
			desc:       "label traefik.port set, no binding, falling back on the container's IP/traefik.port",
			container:  containerJSON(withNetwork("testnet", ipv4("10.11.12.13"))),
			serverPort: "80",
			expected: expected{
				ip:   "10.11.12.13",
				port: "80",
			},
		},
		{
			desc: "label traefik.port set, single binding with ip:port for the label, creates the route",
			container: containerJSON(
				ports(nat.PortMap{
					"443/tcp": []nat.PortBinding{
						{
							HostIP:   "5.6.7.8",
							HostPort: "8082",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			serverPort: "443",
			expected: expected{
				ip:   "5.6.7.8",
				port: "8082",
			},
		},
		{
			desc: "label traefik.port set, no binding on the corresponding port, falling back on the container's IP/label.port",
			container: containerJSON(
				ports(nat.PortMap{
					"443/tcp": []nat.PortBinding{
						{
							HostIP:   "5.6.7.8",
							HostPort: "8082",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			serverPort: "80",
			expected: expected{
				ip:   "10.11.12.13",
				port: "80",
			},
		},
		{
			desc: "label traefik.port set, multiple bindings on different ports, uses the label to select the correct (first) binding",
			container: containerJSON(
				ports(nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "1.2.3.4",
							HostPort: "8081",
						},
					},
					"443/tcp": []nat.PortBinding{
						{
							HostIP:   "5.6.7.8",
							HostPort: "8082",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			serverPort: "80",
			expected: expected{
				ip:   "1.2.3.4",
				port: "8081",
			},
		},
		{
			desc: "label traefik.port set, multiple bindings on different ports, uses the label to select the correct (second) binding",
			container: containerJSON(
				ports(nat.PortMap{
					"80/tcp": []nat.PortBinding{
						{
							HostIP:   "1.2.3.4",
							HostPort: "8081",
						},
					},
					"443/tcp": []nat.PortBinding{
						{
							HostIP:   "5.6.7.8",
							HostPort: "8082",
						},
					},
				}),
				withNetwork("testnet", ipv4("10.11.12.13"))),
			serverPort: "443",
			expected: expected{
				ip:   "5.6.7.8",
				port: "8082",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			dData := parseContainer(test.container)

			builder := NewDynConfBuilder(Shared{
				Network:       "testnet",
				UseBindPortIP: true,
			}, nil, false)

			actualIP, actualPort, actualError := builder.getIPPort(t.Context(), dData, test.serverPort)
			if test.expected.error {
				require.Error(t, actualError)
			} else {
				require.NoError(t, actualError)
			}
			assert.Equal(t, test.expected.ip, actualIP)
			assert.Equal(t, test.expected.port, actualPort)
		})
	}
}

func TestDynConfBuilder_getIPAddress_docker(t *testing.T) {
	testCases := []struct {
		desc      string
		container docker.ContainerJSON
		network   string
		expected  string
	}{
		{
			desc:      "one network, no network label",
			container: containerJSON(withNetwork("testnet", ipv4("10.11.12.13"))),
			expected:  "10.11.12.13",
		},
		{
			desc: "one network, network label",
			container: containerJSON(
				withNetwork("testnet", ipv4("10.11.12.13")),
			),
			network:  "testnet",
			expected: "10.11.12.13",
		},
		{
			desc: "one ipv6 network, network label",
			container: containerJSON(
				withNetwork("testnet", ipv6("fd00:1:2:3:4::")),
			),
			network:  "testnet",
			expected: "fd00:1:2:3:4::",
		},
		{
			desc: "two network ipv4 + ipv6, network label",
			container: containerJSON(
				withNetwork("testnet", ipv4("10.11.12.13"), ipv6("fd00:1:2:3:4::")),
			),
			network:  "testnet",
			expected: "10.11.12.13",
		},
		{
			desc: "two networks, network label",
			container: containerJSON(
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("testnet2", ipv4("10.11.12.14")),
			),
			network:  "testnet2",
			expected: "10.11.12.14",
		},
		{
			desc: "two networks, no network label, mode host",
			container: containerJSON(
				networkMode("host"),
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("testnet2", ipv4("10.11.12.14")),
			),
			expected: "127.0.0.1",
		},
		{
			desc: "two networks, no network label, mode host, use provider network",
			container: containerJSON(
				networkMode("host"),
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("webnet", ipv4("10.11.12.14")),
			),
			expected: "10.11.12.14",
		},
		{
			desc: "two networks, network label",
			container: containerJSON(
				withNetwork("testnet", ipv4("10.11.12.13")),
				withNetwork("webnet", ipv4("10.11.12.14")),
			),
			network:  "testnet",
			expected: "10.11.12.13",
		},
		{
			desc: "no network, no network label, mode host",
			container: containerJSON(
				networkMode("host"),
			),
			expected: "127.0.0.1",
		},
		{
			desc: "no network, no network label, mode host, node IP",
			container: containerJSON(
				networkMode("host"),
				nodeIP("10.0.0.5"),
			),
			expected: "10.0.0.5",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			conf := Shared{
				Network: "webnet",
			}

			dData := parseContainer(test.container)

			dData.ExtraConf.Network = conf.Network
			if len(test.network) > 0 {
				dData.ExtraConf.Network = test.network
			}

			builder := NewDynConfBuilder(conf, nil, false)

			actual := builder.getIPAddress(t.Context(), dData)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestDynConfBuilder_getIPAddress_swarm(t *testing.T) {
	testCases := []struct {
		service  swarm.Service
		expected string
		networks map[string]*network.Summary
	}{
		{
			service:  swarmService(withEndpointSpec(modeDNSRR)),
			expected: "",
			networks: map[string]*network.Summary{},
		},
		{
			service: swarmService(
				withEndpointSpec(modeVIP),
				withEndpoint(virtualIP("1", "10.11.12.13/24")),
			),
			expected: "10.11.12.13",
			networks: map[string]*network.Summary{
				"1": {
					Name: "foo",
				},
			},
		},
		{
			service: swarmService(
				serviceLabels(map[string]string{
					"traefik.swarm.network": "barnet",
				}),
				withEndpointSpec(modeVIP),
				withEndpoint(
					virtualIP("1", "10.11.12.13/24"),
					virtualIP("2", "10.11.12.99/24"),
				),
			),
			expected: "10.11.12.99",
			networks: map[string]*network.Summary{
				"1": {
					Name: "foonet",
				},
				"2": {
					Name: "barnet",
				},
			},
		},
	}

	for serviceID, test := range testCases {
		t.Run(strconv.Itoa(serviceID), func(t *testing.T) {
			t.Parallel()

			var p SwarmProvider
			require.NoError(t, p.Init())

			dData, err := p.parseService(t.Context(), test.service, test.networks)
			require.NoError(t, err)

			builder := NewDynConfBuilder(p.Shared, nil, false)
			actual := builder.getIPAddress(t.Context(), dData)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func pointer[T any](v T) *T { return &v }
