package api

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/containous/mux"
	"github.com/containous/traefik/pkg/config/dynamic"
	"github.com/containous/traefik/pkg/config/static"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var updateExpected = flag.Bool("update_expected", false, "Update expected files in testdata")

func TestHandler_TCP(t *testing.T) {
	type expected struct {
		statusCode int
		nextPage   string
		jsonFile   string
	}

	testCases := []struct {
		desc     string
		path     string
		conf     dynamic.RuntimeConfiguration
		expected expected
	}{
		{
			desc: "all TCP routers, but no config",
			path: "/api/tcp/routers",
			conf: dynamic.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcprouters-empty.json",
			},
		},
		{
			desc: "all TCP routers",
			path: "/api/tcp/routers",
			conf: dynamic.RuntimeConfiguration{
				TCPRouters: map[string]*dynamic.TCPRouterInfo{
					"test@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							TLS: &dynamic.RouterTCPTLSConfig{
								Passthrough: false,
							},
						},
					},
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
				nextPage:   "1",
				jsonFile:   "testdata/tcprouters.json",
			},
		},
		{
			desc: "all TCP routers, pagination, 1 res per page, want page 2",
			path: "/api/tcp/routers?page=2&per_page=1",
			conf: dynamic.RuntimeConfiguration{
				TCPRouters: map[string]*dynamic.TCPRouterInfo{
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
			desc: "one TCP router by id",
			path: "/api/tcp/routers/bar@myprovider",
			conf: dynamic.RuntimeConfiguration{
				TCPRouters: map[string]*dynamic.TCPRouterInfo{
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
			desc: "one TCP router by id, that does not exist",
			path: "/api/tcp/routers/foo@myprovider",
			conf: dynamic.RuntimeConfiguration{
				TCPRouters: map[string]*dynamic.TCPRouterInfo{
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
			conf: dynamic.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "all tcp services, but no config",
			path: "/api/tcp/services",
			conf: dynamic.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcpservices-empty.json",
			},
		},
		{
			desc: "all tcp services",
			path: "/api/tcp/services",
			conf: dynamic.RuntimeConfiguration{
				TCPServices: map[string]*dynamic.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPLoadBalancerService{
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
							LoadBalancer: &dynamic.TCPLoadBalancerService{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.2:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider"},
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
			desc: "all tcp services, 1 res per page, want page 2",
			path: "/api/tcp/services?page=2&per_page=1",
			conf: dynamic.RuntimeConfiguration{
				TCPServices: map[string]*dynamic.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPLoadBalancerService{
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
							LoadBalancer: &dynamic.TCPLoadBalancerService{
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
							LoadBalancer: &dynamic.TCPLoadBalancerService{
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
			conf: dynamic.RuntimeConfiguration{
				TCPServices: map[string]*dynamic.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPLoadBalancerService{
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
			desc: "one tcp service by id, that does not exist",
			path: "/api/tcp/services/nono@myprovider",
			conf: dynamic.RuntimeConfiguration{
				TCPServices: map[string]*dynamic.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPLoadBalancerService{
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
			conf: dynamic.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rtConf := &test.conf
			handler := New(static.Configuration{API: &static.API{}, Global: &static.Global{}}, rtConf)
			router := mux.NewRouter()
			handler.Append(router)

			server := httptest.NewServer(router)

			resp, err := http.DefaultClient.Get(server.URL + test.path)
			require.NoError(t, err)

			assert.Equal(t, test.expected.nextPage, resp.Header.Get(nextPageHeader))

			require.Equal(t, test.expected.statusCode, resp.StatusCode)

			if test.expected.jsonFile == "" {
				return
			}

			assert.Equal(t, resp.Header.Get("Content-Type"), "application/json")

			contents, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			err = resp.Body.Close()
			require.NoError(t, err)

			if *updateExpected {
				var results interface{}
				err := json.Unmarshal(contents, &results)
				require.NoError(t, err)

				newJSON, err := json.MarshalIndent(results, "", "\t")
				require.NoError(t, err)

				err = ioutil.WriteFile(test.expected.jsonFile, newJSON, 0644)
				require.NoError(t, err)
			}

			data, err := ioutil.ReadFile(test.expected.jsonFile)
			require.NoError(t, err)
			assert.JSONEq(t, string(data), string(contents))
		})
	}
}

func TestHandler_HTTP(t *testing.T) {
	type expected struct {
		statusCode int
		nextPage   string
		jsonFile   string
	}

	testCases := []struct {
		desc     string
		path     string
		conf     dynamic.RuntimeConfiguration
		expected expected
	}{
		{
			desc: "all routers, but no config",
			path: "/api/http/routers",
			conf: dynamic.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/routers-empty.json",
			},
		},
		{
			desc: "all routers",
			path: "/api/http/routers",
			conf: dynamic.RuntimeConfiguration{
				Routers: map[string]*dynamic.RouterInfo{
					"test@myprovider": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							Middlewares: []string{"addPrefixTest", "auth"},
						},
					},
					"bar@myprovider": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "addPrefixTest@anotherprovider"},
						},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/routers.json",
			},
		},
		{
			desc: "all routers, pagination, 1 res per page, want page 2",
			path: "/api/http/routers?page=2&per_page=1",
			conf: dynamic.RuntimeConfiguration{
				Routers: map[string]*dynamic.RouterInfo{
					"bar@myprovider": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "addPrefixTest@anotherprovider"},
						},
					},
					"baz@myprovider": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`toto.bar`)",
						},
					},
					"test@myprovider": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							Middlewares: []string{"addPrefixTest", "auth"},
						},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "3",
				jsonFile:   "testdata/routers-page2.json",
			},
		},
		{
			desc: "all routers, pagination, 19 results overall, 7 res per page, want page 3",
			path: "/api/http/routers?page=3&per_page=7",
			conf: dynamic.RuntimeConfiguration{
				Routers: generateHTTPRouters(19),
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/routers-many-lastpage.json",
			},
		},
		{
			desc: "all routers, pagination, 5 results overall, 10 res per page, want page 2",
			path: "/api/http/routers?page=2&per_page=10",
			conf: dynamic.RuntimeConfiguration{
				Routers: generateHTTPRouters(5),
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			desc: "all routers, pagination, 10 results overall, 10 res per page, want page 2",
			path: "/api/http/routers?page=2&per_page=10",
			conf: dynamic.RuntimeConfiguration{
				Routers: generateHTTPRouters(10),
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			desc: "one router by id",
			path: "/api/http/routers/bar@myprovider",
			conf: dynamic.RuntimeConfiguration{
				Routers: map[string]*dynamic.RouterInfo{
					"bar@myprovider": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "addPrefixTest@anotherprovider"},
						},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/router-bar.json",
			},
		},
		{
			desc: "one router by id, that does not exist",
			path: "/api/http/routers/foo@myprovider",
			conf: dynamic.RuntimeConfiguration{
				Routers: map[string]*dynamic.RouterInfo{
					"bar@myprovider": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "addPrefixTest@anotherprovider"},
						},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "one router by id, but no config",
			path: "/api/http/routers/foo@myprovider",
			conf: dynamic.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "all services, but no config",
			path: "/api/http/services",
			conf: dynamic.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/services-empty.json",
			},
		},
		{
			desc: "all services",
			path: "/api/http/services",
			conf: dynamic.RuntimeConfiguration{
				Services: map[string]*dynamic.ServiceInfo{
					"bar@myprovider": func() *dynamic.ServiceInfo {
						si := &dynamic.ServiceInfo{
							Service: &dynamic.Service{
								LoadBalancer: &dynamic.LoadBalancerService{
									Servers: []dynamic.Server{
										{
											URL: "http://127.0.0.1",
										},
									},
								},
							},
							UsedBy: []string{"foo@myprovider", "test@myprovider"},
						}
						si.UpdateStatus("http://127.0.0.1", "UP")
						return si
					}(),
					"baz@myprovider": func() *dynamic.ServiceInfo {
						si := &dynamic.ServiceInfo{
							Service: &dynamic.Service{
								LoadBalancer: &dynamic.LoadBalancerService{
									Servers: []dynamic.Server{
										{
											URL: "http://127.0.0.2",
										},
									},
								},
							},
							UsedBy: []string{"foo@myprovider"},
						}
						si.UpdateStatus("http://127.0.0.2", "UP")
						return si
					}(),
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/services.json",
			},
		},
		{
			desc: "all services, 1 res per page, want page 2",
			path: "/api/http/services?page=2&per_page=1",
			conf: dynamic.RuntimeConfiguration{
				Services: map[string]*dynamic.ServiceInfo{
					"bar@myprovider": func() *dynamic.ServiceInfo {
						si := &dynamic.ServiceInfo{
							Service: &dynamic.Service{
								LoadBalancer: &dynamic.LoadBalancerService{
									Servers: []dynamic.Server{
										{
											URL: "http://127.0.0.1",
										},
									},
								},
							},
							UsedBy: []string{"foo@myprovider", "test@myprovider"},
						}
						si.UpdateStatus("http://127.0.0.1", "UP")
						return si
					}(),
					"baz@myprovider": func() *dynamic.ServiceInfo {
						si := &dynamic.ServiceInfo{
							Service: &dynamic.Service{
								LoadBalancer: &dynamic.LoadBalancerService{
									Servers: []dynamic.Server{
										{
											URL: "http://127.0.0.2",
										},
									},
								},
							},
							UsedBy: []string{"foo@myprovider"},
						}
						si.UpdateStatus("http://127.0.0.2", "UP")
						return si
					}(),
					"test@myprovider": func() *dynamic.ServiceInfo {
						si := &dynamic.ServiceInfo{
							Service: &dynamic.Service{
								LoadBalancer: &dynamic.LoadBalancerService{
									Servers: []dynamic.Server{
										{
											URL: "http://127.0.0.3",
										},
									},
								},
							},
							UsedBy: []string{"foo@myprovider", "test@myprovider"},
						}
						si.UpdateStatus("http://127.0.0.4", "UP")
						return si
					}(),
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "3",
				jsonFile:   "testdata/services-page2.json",
			},
		},
		{
			desc: "one service by id",
			path: "/api/http/services/bar@myprovider",
			conf: dynamic.RuntimeConfiguration{
				Services: map[string]*dynamic.ServiceInfo{
					"bar@myprovider": func() *dynamic.ServiceInfo {
						si := &dynamic.ServiceInfo{
							Service: &dynamic.Service{
								LoadBalancer: &dynamic.LoadBalancerService{
									Servers: []dynamic.Server{
										{
											URL: "http://127.0.0.1",
										},
									},
								},
							},
							UsedBy: []string{"foo@myprovider", "test@myprovider"},
						}
						si.UpdateStatus("http://127.0.0.1", "UP")
						return si
					}(),
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/service-bar.json",
			},
		},
		{
			desc: "one service by id, that does not exist",
			path: "/api/http/services/nono@myprovider",
			conf: dynamic.RuntimeConfiguration{
				Services: map[string]*dynamic.ServiceInfo{
					"bar@myprovider": func() *dynamic.ServiceInfo {
						si := &dynamic.ServiceInfo{
							Service: &dynamic.Service{
								LoadBalancer: &dynamic.LoadBalancerService{
									Servers: []dynamic.Server{
										{
											URL: "http://127.0.0.1",
										},
									},
								},
							},
							UsedBy: []string{"foo@myprovider", "test@myprovider"},
						}
						si.UpdateStatus("http://127.0.0.1", "UP")
						return si
					}(),
				},
			},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "one service by id, but no config",
			path: "/api/http/services/foo@myprovider",
			conf: dynamic.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "all middlewares, but no config",
			path: "/api/http/middlewares",
			conf: dynamic.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/middlewares-empty.json",
			},
		},
		{
			desc: "all middlewares",
			path: "/api/http/middlewares",
			conf: dynamic.RuntimeConfiguration{
				Middlewares: map[string]*dynamic.MiddlewareInfo{
					"auth@myprovider": {
						Middleware: &dynamic.Middleware{
							BasicAuth: &dynamic.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
						UsedBy: []string{"bar@myprovider", "test@myprovider"},
					},
					"addPrefixTest@myprovider": {
						Middleware: &dynamic.Middleware{
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/titi",
							},
						},
						UsedBy: []string{"test@myprovider"},
					},
					"addPrefixTest@anotherprovider": {
						Middleware: &dynamic.Middleware{
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/toto",
							},
						},
						UsedBy: []string{"bar@myprovider"},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/middlewares.json",
			},
		},
		{
			desc: "all middlewares, 1 res per page, want page 2",
			path: "/api/http/middlewares?page=2&per_page=1",
			conf: dynamic.RuntimeConfiguration{
				Middlewares: map[string]*dynamic.MiddlewareInfo{
					"auth@myprovider": {
						Middleware: &dynamic.Middleware{
							BasicAuth: &dynamic.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
						UsedBy: []string{"bar@myprovider", "test@myprovider"},
					},
					"addPrefixTest@myprovider": {
						Middleware: &dynamic.Middleware{
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/titi",
							},
						},
						UsedBy: []string{"test@myprovider"},
					},
					"addPrefixTest@anotherprovider": {
						Middleware: &dynamic.Middleware{
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/toto",
							},
						},
						UsedBy: []string{"bar@myprovider"},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "3",
				jsonFile:   "testdata/middlewares-page2.json",
			},
		},
		{
			desc: "one middleware by id",
			path: "/api/http/middlewares/auth@myprovider",
			conf: dynamic.RuntimeConfiguration{
				Middlewares: map[string]*dynamic.MiddlewareInfo{
					"auth@myprovider": {
						Middleware: &dynamic.Middleware{
							BasicAuth: &dynamic.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
						UsedBy: []string{"bar@myprovider", "test@myprovider"},
					},
					"addPrefixTest@myprovider": {
						Middleware: &dynamic.Middleware{
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/titi",
							},
						},
						UsedBy: []string{"test@myprovider"},
					},
					"addPrefixTest@anotherprovider": {
						Middleware: &dynamic.Middleware{
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/toto",
							},
						},
						UsedBy: []string{"bar@myprovider"},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/middleware-auth.json",
			},
		},
		{
			desc: "one middleware by id, that does not exist",
			path: "/api/http/middlewares/foo@myprovider",
			conf: dynamic.RuntimeConfiguration{
				Middlewares: map[string]*dynamic.MiddlewareInfo{
					"auth@myprovider": {
						Middleware: &dynamic.Middleware{
							BasicAuth: &dynamic.BasicAuth{
								Users: []string{"admin:admin"},
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
			path: "/api/http/middlewares/foo@myprovider",
			conf: dynamic.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rtConf := &test.conf
			handler := New(static.Configuration{API: &static.API{}, Global: &static.Global{}}, rtConf)
			router := mux.NewRouter()
			handler.Append(router)

			server := httptest.NewServer(router)

			resp, err := http.DefaultClient.Get(server.URL + test.path)
			require.NoError(t, err)

			require.Equal(t, test.expected.statusCode, resp.StatusCode)

			assert.Equal(t, test.expected.nextPage, resp.Header.Get(nextPageHeader))

			if test.expected.jsonFile == "" {
				return
			}

			assert.Equal(t, resp.Header.Get("Content-Type"), "application/json")
			contents, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			err = resp.Body.Close()
			require.NoError(t, err)

			if *updateExpected {
				var results interface{}
				err := json.Unmarshal(contents, &results)
				require.NoError(t, err)

				newJSON, err := json.MarshalIndent(results, "", "\t")
				require.NoError(t, err)

				err = ioutil.WriteFile(test.expected.jsonFile, newJSON, 0644)
				require.NoError(t, err)
			}

			data, err := ioutil.ReadFile(test.expected.jsonFile)
			require.NoError(t, err)
			assert.JSONEq(t, string(data), string(contents))
		})
	}
}

func TestHandler_EntryPoints(t *testing.T) {
	type expected struct {
		statusCode int
		nextPage   string
		jsonFile   string
	}

	testCases := []struct {
		desc     string
		path     string
		conf     static.Configuration
		expected expected
	}{
		{
			desc: "all entry points, but no config",
			path: "/api/entrypoints",
			conf: static.Configuration{API: &static.API{}, Global: &static.Global{}},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/entrypoints-empty.json",
			},
		},
		{
			desc: "all entry points",
			path: "/api/entrypoints",
			conf: static.Configuration{
				Global: &static.Global{},
				API:    &static.API{},
				EntryPoints: map[string]*static.EntryPoint{
					"web": {
						Address: ":80",
						Transport: &static.EntryPointsTransport{
							LifeCycle: &static.LifeCycle{
								RequestAcceptGraceTimeout: 1,
								GraceTimeOut:              2,
							},
							RespondingTimeouts: &static.RespondingTimeouts{
								ReadTimeout:  3,
								WriteTimeout: 4,
								IdleTimeout:  5,
							},
						},
						ProxyProtocol: &static.ProxyProtocol{
							Insecure:   true,
							TrustedIPs: []string{"192.168.1.1", "192.168.1.2"},
						},
						ForwardedHeaders: &static.ForwardedHeaders{
							Insecure:   true,
							TrustedIPs: []string{"192.168.1.3", "192.168.1.4"},
						},
					},
					"web-secure": {
						Address: ":443",
						Transport: &static.EntryPointsTransport{
							LifeCycle: &static.LifeCycle{
								RequestAcceptGraceTimeout: 10,
								GraceTimeOut:              20,
							},
							RespondingTimeouts: &static.RespondingTimeouts{
								ReadTimeout:  30,
								WriteTimeout: 40,
								IdleTimeout:  50,
							},
						},
						ProxyProtocol: &static.ProxyProtocol{
							Insecure:   true,
							TrustedIPs: []string{"192.168.1.10", "192.168.1.20"},
						},
						ForwardedHeaders: &static.ForwardedHeaders{
							Insecure:   true,
							TrustedIPs: []string{"192.168.1.30", "192.168.1.40"},
						},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/entrypoints.json",
			},
		},
		{
			desc: "all entry points, pagination, 1 res per page, want page 2",
			path: "/api/entrypoints?page=2&per_page=1",
			conf: static.Configuration{
				Global: &static.Global{},
				API:    &static.API{},
				EntryPoints: map[string]*static.EntryPoint{
					"web1": {Address: ":81"},
					"web2": {Address: ":82"},
					"web3": {Address: ":83"},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "3",
				jsonFile:   "testdata/entrypoints-page2.json",
			},
		},
		{
			desc: "all entry points, pagination, 19 results overall, 7 res per page, want page 3",
			path: "/api/entrypoints?page=3&per_page=7",
			conf: static.Configuration{
				Global:      &static.Global{},
				API:         &static.API{},
				EntryPoints: generateEntryPoints(19),
			},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/entrypoints-many-lastpage.json",
			},
		},
		{
			desc: "all entry points, pagination, 5 results overall, 10 res per page, want page 2",
			path: "/api/entrypoints?page=2&per_page=10",
			conf: static.Configuration{
				Global:      &static.Global{},
				API:         &static.API{},
				EntryPoints: generateEntryPoints(5),
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			desc: "all entry points, pagination, 10 results overall, 10 res per page, want page 2",
			path: "/api/entrypoints?page=2&per_page=10",
			conf: static.Configuration{
				Global:      &static.Global{},
				API:         &static.API{},
				EntryPoints: generateEntryPoints(10),
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			desc: "one entry point by id",
			path: "/api/entrypoints/bar",
			conf: static.Configuration{
				Global: &static.Global{},
				API:    &static.API{},
				EntryPoints: map[string]*static.EntryPoint{
					"bar": {Address: ":81"},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/entrypoint-bar.json",
			},
		},
		{
			desc: "one entry point by id, that does not exist",
			path: "/api/entrypoints/foo",
			conf: static.Configuration{
				Global: &static.Global{},
				API:    &static.API{},
				EntryPoints: map[string]*static.EntryPoint{
					"bar": {Address: ":81"},
				},
			},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "one entry point by id, but no config",
			path: "/api/entrypoints/foo",
			conf: static.Configuration{API: &static.API{}, Global: &static.Global{}},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := New(test.conf, &dynamic.RuntimeConfiguration{})
			router := mux.NewRouter()
			handler.Append(router)

			server := httptest.NewServer(router)

			resp, err := http.DefaultClient.Get(server.URL + test.path)
			require.NoError(t, err)

			require.Equal(t, test.expected.statusCode, resp.StatusCode)

			assert.Equal(t, test.expected.nextPage, resp.Header.Get(nextPageHeader))

			if test.expected.jsonFile == "" {
				return
			}

			assert.Equal(t, resp.Header.Get("Content-Type"), "application/json")
			contents, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			err = resp.Body.Close()
			require.NoError(t, err)

			if *updateExpected {
				var results interface{}
				err := json.Unmarshal(contents, &results)
				require.NoError(t, err)

				newJSON, err := json.MarshalIndent(results, "", "\t")
				require.NoError(t, err)

				err = ioutil.WriteFile(test.expected.jsonFile, newJSON, 0644)
				require.NoError(t, err)
			}

			data, err := ioutil.ReadFile(test.expected.jsonFile)
			require.NoError(t, err)
			assert.JSONEq(t, string(data), string(contents))
		})
	}
}

func TestHandler_RawData(t *testing.T) {
	type expected struct {
		statusCode int
		json       string
	}

	testCases := []struct {
		desc     string
		path     string
		conf     dynamic.RuntimeConfiguration
		expected expected
	}{
		{
			desc: "Get rawdata",
			path: "/api/rawdata",
			conf: dynamic.RuntimeConfiguration{
				Services: map[string]*dynamic.ServiceInfo{
					"foo-service@myprovider": {
						Service: &dynamic.Service{
							LoadBalancer: &dynamic.LoadBalancerService{
								Servers: []dynamic.Server{
									{
										URL: "http://127.0.0.1",
									},
								},
							},
						},
					},
				},
				Middlewares: map[string]*dynamic.MiddlewareInfo{
					"auth@myprovider": {
						Middleware: &dynamic.Middleware{
							BasicAuth: &dynamic.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
					},
					"addPrefixTest@myprovider": {
						Middleware: &dynamic.Middleware{
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/titi",
							},
						},
					},
					"addPrefixTest@anotherprovider": {
						Middleware: &dynamic.Middleware{
							AddPrefix: &dynamic.AddPrefix{
								Prefix: "/toto",
							},
						},
					},
				},
				Routers: map[string]*dynamic.RouterInfo{
					"bar@myprovider": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "addPrefixTest@anotherprovider"},
						},
					},
					"test@myprovider": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							Middlewares: []string{"addPrefixTest", "auth"},
						},
					},
				},
				TCPServices: map[string]*dynamic.TCPServiceInfo{
					"tcpfoo-service@myprovider": {
						TCPService: &dynamic.TCPService{
							LoadBalancer: &dynamic.TCPLoadBalancerService{
								Servers: []dynamic.TCPServer{
									{
										Address: "127.0.0.1",
									},
								},
							},
						},
					},
				},
				TCPRouters: map[string]*dynamic.TCPRouterInfo{
					"tcpbar@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "tcpfoo-service@myprovider",
							Rule:        "HostSNI(`foo.bar`)",
						},
					},
					"tcptest@myprovider": {
						TCPRouter: &dynamic.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "tcpfoo-service@myprovider",
							Rule:        "HostSNI(`foo.bar.other`)",
						},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusOK,
				json:       "testdata/getrawdata.json",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			// TODO: server status

			rtConf := &test.conf

			rtConf.PopulateUsedBy()
			handler := New(static.Configuration{API: &static.API{}, Global: &static.Global{}}, rtConf)
			router := mux.NewRouter()
			handler.Append(router)

			server := httptest.NewServer(router)

			resp, err := http.DefaultClient.Get(server.URL + test.path)
			require.NoError(t, err)

			assert.Equal(t, test.expected.statusCode, resp.StatusCode)
			assert.Equal(t, resp.Header.Get("Content-Type"), "application/json")

			contents, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			err = resp.Body.Close()
			require.NoError(t, err)

			if test.expected.json == "" {
				return
			}
			if *updateExpected {
				var rtRepr RunTimeRepresentation
				err := json.Unmarshal(contents, &rtRepr)
				require.NoError(t, err)

				newJSON, err := json.MarshalIndent(rtRepr, "", "\t")
				require.NoError(t, err)

				err = ioutil.WriteFile(test.expected.json, newJSON, 0644)
				require.NoError(t, err)
			}

			data, err := ioutil.ReadFile(test.expected.json)
			require.NoError(t, err)
			assert.JSONEq(t, string(data), string(contents))
		})
	}
}

func generateHTTPRouters(nbRouters int) map[string]*dynamic.RouterInfo {
	routers := make(map[string]*dynamic.RouterInfo, nbRouters)
	for i := 0; i < nbRouters; i++ {
		routers[fmt.Sprintf("bar%2d@myprovider", i)] = &dynamic.RouterInfo{
			Router: &dynamic.Router{
				EntryPoints: []string{"web"},
				Service:     "foo-service@myprovider",
				Rule:        "Host(`foo.bar" + strconv.Itoa(i) + "`)",
			},
		}
	}
	return routers
}

func generateEntryPoints(nb int) map[string]*static.EntryPoint {
	eps := make(map[string]*static.EntryPoint, nb)
	for i := 0; i < nb; i++ {
		eps[fmt.Sprintf("ep%2d", i)] = &static.EntryPoint{
			Address: ":" + strconv.Itoa(i),
		}
	}

	return eps
}
