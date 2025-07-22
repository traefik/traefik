package tcp

import (
	"crypto/tls"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	tcpmiddleware "github.com/traefik/traefik/v3/pkg/server/middleware/tcp"
	"github.com/traefik/traefik/v3/pkg/server/service/tcp"
	tcp2 "github.com/traefik/traefik/v3/pkg/tcp"
	traefiktls "github.com/traefik/traefik/v3/pkg/tls"
)

func TestRuntimeConfiguration(t *testing.T) {
	testCases := []struct {
		desc              string
		httpServiceConfig map[string]*runtime.ServiceInfo
		httpRouterConfig  map[string]*runtime.RouterInfo
		tcpServiceConfig  map[string]*runtime.TCPServiceInfo
		tcpRouterConfig   map[string]*runtime.TCPRouterInfo
		expectedError     int
	}{
		{
			desc: "No error",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
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
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"foo": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`bar.foo`)",
						TLS: &dynamic.RouterTCPTLSConfig{
							Passthrough: false,
							Options:     "foo",
						},
					},
				},
				"bar": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
						TLS: &dynamic.RouterTCPTLSConfig{
							Passthrough: false,
							Options:     "bar",
						},
					},
				},
			},
			expectedError: 0,
		},
		{
			desc: "Non-ASCII domain error",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Port:    "8085",
									Address: "127.0.0.1:8085",
								},
							},
						},
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"foo": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`bàr.foo`)",
						TLS: &dynamic.RouterTCPTLSConfig{
							Passthrough: false,
							Options:     "foo",
						},
					},
				},
			},
			expectedError: 1,
		},
		{
			desc: "HTTP routers with same domain but different TLS options",
			httpServiceConfig: map[string]*runtime.ServiceInfo{
				"foo-service": {
					Service: &dynamic.Service{
						LoadBalancer: &dynamic.ServersLoadBalancer{
							Servers: []dynamic.Server{
								{
									Port: "8085",
									URL:  "127.0.0.1:8085",
								},
								{
									URL:  "127.0.0.1:8086",
									Port: "8086",
								},
							},
						},
					},
				},
			},
			httpRouterConfig: map[string]*runtime.RouterInfo{
				"foo": {
					Router: &dynamic.Router{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "Host(`bar.foo`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "foo",
						},
					},
				},
				"bar": {
					Router: &dynamic.Router{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "Host(`bar.foo`) && PathPrefix(`/path`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "bar",
						},
					},
				},
			},
			expectedError: 2,
		},
		{
			desc: "One router with wrong rule",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "127.0.0.1:80",
								},
							},
						},
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"foo": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "WrongRule(`bar.foo`)",
					},
				},

				"bar": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
						TLS:         &dynamic.RouterTCPTLSConfig{},
					},
				},
			},
			expectedError: 1,
		},
		{
			desc: "All router with wrong rule",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "127.0.0.1:80",
								},
							},
						},
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"foo": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "WrongRule(`bar.foo`)",
					},
				},
				"bar": {
					TCPRouter: &dynamic.TCPRouter{
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
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "127.0.0.1:80",
								},
							},
						},
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"foo": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "wrong-service",
						Rule:        "HostSNI(`bar.foo`)",
						TLS:         &dynamic.RouterTCPTLSConfig{},
					},
				},
				"bar": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
						TLS:         &dynamic.RouterTCPTLSConfig{},
					},
				},
			},
			expectedError: 1,
		},
		{
			desc: "Router with broken service",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: nil,
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"bar": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
						TLS:         &dynamic.RouterTCPTLSConfig{},
					},
				},
			},
			expectedError: 2,
		},
		{
			desc: "Router with priority exceeding the max user-defined priority",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
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
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"bar": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`foo.bar`)",
						TLS:         &dynamic.RouterTCPTLSConfig{},
						Priority:    math.MaxInt,
					},
				},
			},
			expectedError: 1,
		},
		{
			desc: "Router with HostSNI but no TLS",
			tcpServiceConfig: map[string]*runtime.TCPServiceInfo{
				"foo-service": {
					TCPService: &dynamic.TCPService{
						LoadBalancer: &dynamic.TCPServersLoadBalancer{
							Servers: []dynamic.TCPServer{
								{
									Address: "127.0.0.1:80",
								},
							},
						},
					},
				},
			},
			tcpRouterConfig: map[string]*runtime.TCPRouterInfo{
				"foo": {
					TCPRouter: &dynamic.TCPRouter{
						EntryPoints: []string{"web"},
						Service:     "foo-service",
						Rule:        "HostSNI(`bar.foo`)",
					},
				},
			},
			expectedError: 1,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			entryPoints := []string{"web"}

			conf := &runtime.Configuration{
				Services:    test.httpServiceConfig,
				Routers:     test.httpRouterConfig,
				TCPServices: test.tcpServiceConfig,
				TCPRouters:  test.tcpRouterConfig,
			}
			dialerManager := tcp2.NewDialerManager(nil)
			dialerManager.Update(map[string]*dynamic.TCPServersTransport{"default@internal": {}})
			serviceManager := tcp.NewManager(conf, dialerManager)
			tlsManager := traefiktls.NewManager(nil)
			tlsManager.UpdateConfigs(
				t.Context(),
				map[string]traefiktls.Store{},
				map[string]traefiktls.Options{
					"default": {
						MinVersion: "VersionTLS10",
					},
					"foo": {
						MinVersion: "VersionTLS12",
					},
					"bar": {
						MinVersion: "VersionTLS11",
					},
				},
				[]*traefiktls.CertAndStores{})

			middlewaresBuilder := tcpmiddleware.NewBuilder(conf.TCPMiddlewares)

			routerManager := NewManager(conf, serviceManager, middlewaresBuilder,
				nil, nil, tlsManager)

			_ = routerManager.BuildHandlers(t.Context(), entryPoints)

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
				if len(v.Err) > 0 {
					allErrors++
				}
			}
			for _, v := range conf.Services {
				if v.Err != nil {
					allErrors++
				}
			}
			for _, v := range conf.Routers {
				if len(v.Err) > 0 {
					allErrors++
				}
			}
			assert.Equal(t, test.expectedError, allErrors)
		})
	}
}

func TestDomainFronting(t *testing.T) {
	tlsOptionsBase := map[string]traefiktls.Options{
		"default": {
			MinVersion: "VersionTLS10",
		},
		"host1@file": {
			MinVersion: "VersionTLS12",
		},
		"host1@crd": {
			MinVersion: "VersionTLS12",
		},
	}

	entryPoints := []string{"web"}

	tests := []struct {
		desc           string
		routers        map[string]*runtime.RouterInfo
		tlsOptions     map[string]traefiktls.Options
		host           string
		ServerName     string
		expectedStatus int
	}{
		{
			desc: "Request is misdirected when TLS options are different",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
				"router-2@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host2.local`)",
						TLS:         &dynamic.RouterTLSConfig{},
					},
				},
			},
			tlsOptions:     tlsOptionsBase,
			host:           "host1.local",
			ServerName:     "host2.local",
			expectedStatus: http.StatusMisdirectedRequest,
		},
		{
			desc: "Request is OK when TLS options are the same",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
				"router-2@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host2.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
			},
			tlsOptions:     tlsOptionsBase,
			host:           "host1.local",
			ServerName:     "host2.local",
			expectedStatus: http.StatusOK,
		},
		{
			desc: "Default TLS options is used when options are ambiguous for the same host",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
				"router-2@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`) && PathPrefix(`/foo`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "default",
						},
					},
				},
				"router-3@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host2.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
			},
			tlsOptions:     tlsOptionsBase,
			host:           "host1.local",
			ServerName:     "host2.local",
			expectedStatus: http.StatusMisdirectedRequest,
		},
		{
			desc: "Default TLS options should not be used when options are the same for the same host",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
				"router-2@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`) && PathPrefix(`/bar`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
				"router-3@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host2.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
			},
			tlsOptions:     tlsOptionsBase,
			host:           "host1.local",
			ServerName:     "host2.local",
			expectedStatus: http.StatusOK,
		},
		{
			desc: "Request is misdirected when TLS options have the same name but from different providers",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
				"router-2@crd": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host2.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1",
						},
					},
				},
			},
			tlsOptions:     tlsOptionsBase,
			host:           "host1.local",
			ServerName:     "host2.local",
			expectedStatus: http.StatusMisdirectedRequest,
		},
		{
			desc: "Request is OK when TLS options reference from a different provider is the same",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1@crd",
						},
					},
				},
				"router-2@crd": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host2.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1@crd",
						},
					},
				},
			},
			tlsOptions:     tlsOptionsBase,
			host:           "host1.local",
			ServerName:     "host2.local",
			expectedStatus: http.StatusOK,
		},
		{
			desc: "Request is misdirected when server name is empty and the host name is an FQDN, but router's rule is not",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1@file",
						},
					},
				},
			},
			tlsOptions: map[string]traefiktls.Options{
				"default": {
					MinVersion: "VersionTLS13",
				},
				"host1@file": {
					MinVersion: "VersionTLS12",
				},
			},
			host:           "host1.local.",
			expectedStatus: http.StatusMisdirectedRequest,
		},
		{
			desc: "Request is misdirected when server name is empty and the host name is not FQDN, but router's rule is",
			routers: map[string]*runtime.RouterInfo{
				"router-1@file": {
					Router: &dynamic.Router{
						EntryPoints: entryPoints,
						Rule:        "Host(`host1.local.`)",
						TLS: &dynamic.RouterTLSConfig{
							Options: "host1@file",
						},
					},
				},
			},
			tlsOptions: map[string]traefiktls.Options{
				"default": {
					MinVersion: "VersionTLS13",
				},
				"host1@file": {
					MinVersion: "VersionTLS12",
				},
			},
			host:           "host1.local",
			expectedStatus: http.StatusMisdirectedRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			conf := &runtime.Configuration{
				Routers: test.routers,
			}

			serviceManager := tcp.NewManager(conf, tcp2.NewDialerManager(nil))

			tlsManager := traefiktls.NewManager(nil)
			tlsManager.UpdateConfigs(t.Context(), map[string]traefiktls.Store{}, test.tlsOptions, []*traefiktls.CertAndStores{})

			httpsHandler := map[string]http.Handler{
				"web": http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {}),
			}

			middlewaresBuilder := tcpmiddleware.NewBuilder(conf.TCPMiddlewares)

			routerManager := NewManager(conf, serviceManager, middlewaresBuilder, nil, httpsHandler, tlsManager)

			routers := routerManager.BuildHandlers(t.Context(), entryPoints)

			router, ok := routers["web"]
			require.True(t, ok)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = test.host
			req.TLS = &tls.ConnectionState{
				ServerName: test.ServerName,
			}

			rw := httptest.NewRecorder()

			router.GetHTTPSHandler().ServeHTTP(rw, req)

			assert.Equal(t, test.expectedStatus, rw.Code)
		})
	}
}
