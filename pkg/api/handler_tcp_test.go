package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/config/static"
)

func TestHandler_TCP(t *testing.T) {
	type expected struct {
		statusCode int
		nextPage   string
		jsonFile   string
	}

	testCases := []struct {
		desc     string
		path     string
		conf     runtime.Configuration
		expected expected
	}{
		{
			desc: "all TCP routers, but no config",
			path: "/api/tcp/routers",
			conf: runtime.Configuration{},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcprouters-empty.json",
			},
		},
		{
			desc: "all TCP routers",
			path: "/api/tcp/routers",
			conf: runtime.Configuration{
				TCPRouters: map[string]*runtime.TCPRouterInfo{
					"test@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: false,
							},
						},
						Status: runtime.StatusEnabled,
					},
					"bar@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
						Status: runtime.StatusWarning,
					},
					"foo@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
						Status: runtime.StatusDisabled,
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcprouters.json",
			},
		},
		{
			desc: "all TCP routers, pagination, 1 res per page, want page 2",
			path: "/api/tcp/routers?page=2&per_page=1",
			conf: runtime.Configuration{
				TCPRouters: map[string]*runtime.TCPRouterInfo{
					"bar@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
					},
					"baz@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`toto.bar`)",
						},
					},
					"test@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
						},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "3",
				jsonFile:   "testdata/tcprouters-page2.json",
			},
		},
		{
			desc: "TCP routers filtered by status",
			path: "/api/tcp/routers?status=enabled",
			conf: runtime.Configuration{
				TCPRouters: map[string]*runtime.TCPRouterInfo{
					"test@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: false,
							},
						},
						Status: runtime.StatusEnabled,
					},
					"bar@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
						Status: runtime.StatusWarning,
					},
					"foo@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
						Status: runtime.StatusDisabled,
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcprouters-filtered-status.json",
			},
		},
		{
			desc: "TCP routers filtered by search",
			path: "/api/tcp/routers?search=bar@my",
			conf: runtime.Configuration{
				TCPRouters: map[string]*runtime.TCPRouterInfo{
					"test@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: false,
							},
						},
						Status: runtime.StatusEnabled,
					},
					"bar@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
						Status: runtime.StatusWarning,
					},
					"foo@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
						Status: runtime.StatusDisabled,
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcprouters-filtered-search.json",
			},
		},
		{
			desc: "TCP routers filtered by service",
			path: "/api/tcp/routers?serviceName=foo-service@myprovider",
			conf: runtime.Configuration{
				TCPRouters: map[string]*runtime.TCPRouterInfo{
					"test@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: false,
							},
						},
						Status: runtime.StatusEnabled,
					},
					"bar@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service",
							Rule:        "Host(`foo.bar`)",
						},
						Status: runtime.StatusWarning,
					},
					"foo@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "bar-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
						Status: runtime.StatusDisabled,
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcprouters-filtered-serviceName.json",
			},
		},
		{
			desc: "TCP routers filtered by middleware",
			path: "/api/tcp/routers?middlewareName=auth",
			conf: runtime.Configuration{
				TCPRouters: map[string]*runtime.TCPRouterInfo{
					"test@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							Middlewares: []string{"inflightconn@myprovider"},
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: false,
							},
						},
						Status: runtime.StatusEnabled,
					},
					"bar@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "inflightconn@myprovider"},
						},
						Status: runtime.StatusWarning,
					},
					"foo@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "bar-service@myprovider",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"inflightconn@myprovider", "auth"},
						},
						Status: runtime.StatusDisabled,
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcprouters-filtered-middlewareName.json",
			},
		},
		{
			desc: "one TCP router by id",
			path: "/api/tcp/routers/bar@myprovider",
			conf: runtime.Configuration{
				TCPRouters: map[string]*runtime.TCPRouterInfo{
					"bar@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/tcprouter-bar.json",
			},
		},
		{
			desc: "one TCP router by id containing slash",
			path: "/api/tcp/routers/" + url.PathEscape("foo / bar@myprovider"),
			conf: runtime.Configuration{
				TCPRouters: map[string]*runtime.TCPRouterInfo{
					"foo / bar@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/tcprouter-foo-slash-bar.json",
			},
		},
		{
			desc: "one TCP router by id, that does not exist",
			path: "/api/tcp/routers/foo@myprovider",
			conf: runtime.Configuration{
				TCPRouters: map[string]*runtime.TCPRouterInfo{
					"bar@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "one TCP router by id, but no config",
			path: "/api/tcp/routers/bar@myprovider",
			conf: runtime.Configuration{},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "all tcp services, but no config",
			path: "/api/tcp/services",
			conf: runtime.Configuration{},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcpservices-empty.json",
			},
		},
		{
			desc: "all tcp services",
			path: "/api/tcp/services",
			conf: runtime.Configuration{
				TCPServices: map[string]*runtime.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider", "test@myprovider"},
						Status: runtime.StatusEnabled,
					},
					"baz@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.2:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider"},
						Status: runtime.StatusWarning,
					},
					"foz@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.2:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider"},
						Status: runtime.StatusDisabled,
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcpservices.json",
			},
		},
		{
			desc: "tcp services filtered by status",
			path: "/api/tcp/services?status=enabled",
			conf: runtime.Configuration{
				TCPServices: map[string]*runtime.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider", "test@myprovider"},
						Status: runtime.StatusEnabled,
					},
					"baz@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.2:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider"},
						Status: runtime.StatusWarning,
					},
					"foz@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.2:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider"},
						Status: runtime.StatusDisabled,
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcpservices-filtered-status.json",
			},
		},
		{
			desc: "tcp services filtered by search",
			path: "/api/tcp/services?search=baz@my",
			conf: runtime.Configuration{
				TCPServices: map[string]*runtime.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider", "test@myprovider"},
						Status: runtime.StatusEnabled,
					},
					"baz@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.2:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider"},
						Status: runtime.StatusWarning,
					},
					"foz@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.2:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider"},
						Status: runtime.StatusDisabled,
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcpservices-filtered-search.json",
			},
		},
		{
			desc: "all tcp services, 1 res per page, want page 2",
			path: "/api/tcp/services?page=2&per_page=1",
			conf: runtime.Configuration{
				TCPServices: map[string]*runtime.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider", "test@myprovider"},
					},
					"baz@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.2:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider"},
					},
					"test@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.3:2345",
									},
								},
							},
						},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "3",
				jsonFile:   "testdata/tcpservices-page2.json",
			},
		},
		{
			desc: "one tcp service by id",
			path: "/api/tcp/services/bar@myprovider",
			conf: runtime.Configuration{
				TCPServices: map[string]*runtime.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider", "test@myprovider"},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/tcpservice-bar.json",
			},
		},
		{
			desc: "one tcp service by id containing slash",
			path: "/api/tcp/services/" + url.PathEscape("foo / bar@myprovider"),
			conf: runtime.Configuration{
				TCPServices: map[string]*runtime.TCPServiceInfo{
					"foo / bar@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider", "test@myprovider"},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/tcpservice-foo-slash-bar.json",
			},
		},
		{
			desc: "one tcp service by id, that does not exist",
			path: "/api/tcp/services/nono@myprovider",
			conf: runtime.Configuration{
				TCPServices: map[string]*runtime.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPServersLoadBalancer{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider", "test@myprovider"},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "one tcp service by id, but no config",
			path: "/api/tcp/services/foo@myprovider",
			conf: runtime.Configuration{},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "all middlewares",
			path: "/api/tcp/middlewares",
			conf: runtime.Configuration{
				TCPMiddlewares: map[string]*runtime.TCPMiddlewareInfo{
					"ipallowlist1@myprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						UsedBy: []string{"bar@myprovider", "test@myprovider"},
					},
					"ipallowlist2@myprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.2/32"},
							},
						},
						UsedBy: []string{"test@myprovider"},
					},
					"ipallowlist1@anotherprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						UsedBy: []string{"bar@myprovider"},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcpmiddlewares.json",
			},
		},
		{
			desc: "middlewares filtered by status",
			path: "/api/tcp/middlewares?status=enabled",
			conf: runtime.Configuration{
				TCPMiddlewares: map[string]*runtime.TCPMiddlewareInfo{
					"ipallowlist@myprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						UsedBy: []string{"bar@myprovider", "test@myprovider"},
						Status: runtime.StatusEnabled,
					},
					"ipallowlist2@myprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.2/32"},
							},
						},
						UsedBy: []string{"test@myprovider"},
						Status: runtime.StatusDisabled,
					},
					"ipallowlist@anotherprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						UsedBy: []string{"bar@myprovider"},
						Status: runtime.StatusEnabled,
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcpmiddlewares-filtered-status.json",
			},
		},
		{
			desc: "middlewares filtered by search",
			path: "/api/tcp/middlewares?search=ipallowlist",
			conf: runtime.Configuration{
				TCPMiddlewares: map[string]*runtime.TCPMiddlewareInfo{
					"bad@myprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						UsedBy: []string{"bar@myprovider", "test@myprovider"},
						Status: runtime.StatusEnabled,
					},
					"ipallowlist@myprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						UsedBy: []string{"test@myprovider"},
						Status: runtime.StatusDisabled,
					},
					"ipallowlist@anotherprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						UsedBy: []string{"bar@myprovider"},
						Status: runtime.StatusEnabled,
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcpmiddlewares-filtered-search.json",
			},
		},
		{
			desc: "all middlewares, 1 res per page, want page 2",
			path: "/api/tcp/middlewares?page=2&per_page=1",
			conf: runtime.Configuration{
				TCPMiddlewares: map[string]*runtime.TCPMiddlewareInfo{
					"ipallowlist@myprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						UsedBy: []string{"bar@myprovider", "test@myprovider"},
					},
					"ipallowlist2@myprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.2/32"},
							},
						},
						UsedBy: []string{"test@myprovider"},
					},
					"ipallowlist@anotherprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						UsedBy: []string{"bar@myprovider"},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "3",
				jsonFile:   "testdata/tcpmiddlewares-page2.json",
			},
		},
		{
			desc: "one middleware by id",
			path: "/api/tcp/middlewares/ipallowlist@myprovider",
			conf: runtime.Configuration{
				TCPMiddlewares: map[string]*runtime.TCPMiddlewareInfo{
					"ipallowlist@myprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						UsedBy: []string{"bar@myprovider", "test@myprovider"},
					},
					"ipallowlist2@myprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.2/32"},
							},
						},
						UsedBy: []string{"test@myprovider"},
					},
					"ipallowlist@anotherprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						UsedBy: []string{"bar@myprovider"},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/tcpmiddleware-ipallowlist.json",
			},
		},
		{
			desc: "one middleware by id containing slash",
			path: "/api/tcp/middlewares/" + url.PathEscape("foo / bar@myprovider"),
			conf: runtime.Configuration{
				TCPMiddlewares: map[string]*runtime.TCPMiddlewareInfo{
					"foo / bar@myprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPWhiteList: &dynamic.TCPIPWhiteList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						UsedBy: []string{"bar@myprovider", "test@myprovider"},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/tcpmiddleware-foo-slash-bar.json",
			},
		},
		{
			desc: "one middleware by id, that does not exist",
			path: "/api/tcp/middlewares/foo@myprovider",
			conf: runtime.Configuration{
				TCPMiddlewares: map[string]*runtime.TCPMiddlewareInfo{
					"ipallowlist@myprovider": {
						TCPMiddleware: &dynamic.TCPMiddleware{
							IPAllowList: &dynamic.TCPIPAllowList{
								SourceRange: []string{"127.0.0.1/32"},
							},
						},
						UsedBy: []string{"bar@myprovider", "test@myprovider"},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "one middleware by id, but no config",
			path: "/api/tcp/middlewares/foo@myprovider",
			conf: runtime.Configuration{},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rtConf := &test.conf
			// To lazily initialize the Statuses.
			rtConf.PopulateUsedBy()
			rtConf.GetTCPRoutersByEntryPoints(t.Context(), []string{"web"})

			handler := New(static.Configuration{API: &static.API{}, Global: &static.Global{}}, rtConf)
			server := httptest.NewServer(handler.createRouter())

			resp, err := http.DefaultClient.Get(server.URL + test.path)
			require.NoError(t, err)

			assert.Equal(t, test.expected.nextPage, resp.Header.Get(nextPageHeader))

			require.Equal(t, test.expected.statusCode, resp.StatusCode)

			if test.expected.jsonFile == "" {
				return
			}

			assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

			contents, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			err = resp.Body.Close()
			require.NoError(t, err)

			if *updateExpected {
				var results interface{}
				err := json.Unmarshal(contents, &results)
				require.NoError(t, err)

				newJSON, err := json.MarshalIndent(results, "", "\t")
				require.NoError(t, err)

				err = os.WriteFile(test.expected.jsonFile, newJSON, 0o644)
				require.NoError(t, err)
			}

			data, err := os.ReadFile(test.expected.jsonFile)
			require.NoError(t, err)
			assert.JSONEq(t, string(data), string(contents))
		})
	}
}
