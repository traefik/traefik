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
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/config/static"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var updateExpected = flag.Bool("update_expected", false, "Update expected files in testdata")

func TestHandlerTCP_API(t *testing.T) {
	type expected struct {
		statusCode int
		nextPage   string
		jsonFile   string
	}

	testCases := []struct {
		desc     string
		path     string
		conf     config.RuntimeConfiguration
		expected expected
	}{
		{
			desc: "all TCP routers, but no config",
			path: "/api/tcp/routers",
			conf: config.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcprouters-empty.json",
			},
		},
		{
			desc: "all TCP routers",
			path: "/api/tcp/routers",
			conf: config.RuntimeConfiguration{
				TCPRouters: map[string]*config.TCPRouterInfo{
					"test@myprovider": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							TLS: &config.RouterTCPTLSConfig{
								Passthrough: false,
							},
						},
					},
					"bar@myprovider": {
						TCPRouter: &config.TCPRouter{
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
			conf: config.RuntimeConfiguration{
				TCPRouters: map[string]*config.TCPRouterInfo{
					"bar@myprovider": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
						},
					},
					"baz@myprovider": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`toto.bar`)",
						},
					},
					"test@myprovider": {
						TCPRouter: &config.TCPRouter{
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
			conf: config.RuntimeConfiguration{
				TCPRouters: map[string]*config.TCPRouterInfo{
					"bar@myprovider": {
						TCPRouter: &config.TCPRouter{
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
			conf: config.RuntimeConfiguration{
				TCPRouters: map[string]*config.TCPRouterInfo{
					"bar@myprovider": {
						TCPRouter: &config.TCPRouter{
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
			conf: config.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "all tcp services, but no config",
			path: "/api/tcp/services",
			conf: config.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/tcpservices-empty.json",
			},
		},
		{
			desc: "all tcp services",
			path: "/api/tcp/services",
			conf: config.RuntimeConfiguration{
				TCPServices: map[string]*config.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider", "test@myprovider"},
					},
					"baz@myprovider": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
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
			conf: config.RuntimeConfiguration{
				TCPServices: map[string]*config.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider", "test@myprovider"},
					},
					"baz@myprovider": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.2:2345",
									},
								},
							},
						},
						UsedBy: []string{"foo@myprovider"},
					},
					"test@myprovider": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
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
			conf: config.RuntimeConfiguration{
				TCPServices: map[string]*config.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
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
			conf: config.RuntimeConfiguration{
				TCPServices: map[string]*config.TCPServiceInfo{
					"bar@myprovider": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
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
			conf: config.RuntimeConfiguration{},
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

func TestHandlerHTTP_API(t *testing.T) {
	type expected struct {
		statusCode int
		nextPage   string
		jsonFile   string
	}

	testCases := []struct {
		desc     string
		path     string
		conf     config.RuntimeConfiguration
		expected expected
	}{
		{
			desc: "all routers, but no config",
			path: "/api/http/routers",
			conf: config.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/routers-empty.json",
			},
		},
		{
			desc: "all routers",
			path: "/api/http/routers",
			conf: config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"test@myprovider": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							Middlewares: []string{"addPrefixTest", "auth"},
						},
					},
					"bar@myprovider": {
						Router: &config.Router{
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
			conf: config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"bar@myprovider": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "addPrefixTest@anotherprovider"},
						},
					},
					"baz@myprovider": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`toto.bar`)",
						},
					},
					"test@myprovider": {
						Router: &config.Router{
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
			conf: config.RuntimeConfiguration{
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
			conf: config.RuntimeConfiguration{
				Routers: generateHTTPRouters(5),
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			desc: "all routers, pagination, 10 results overall, 10 res per page, want page 2",
			path: "/api/http/routers?page=2&per_page=10",
			conf: config.RuntimeConfiguration{
				Routers: generateHTTPRouters(10),
			},
			expected: expected{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			desc: "one router by id",
			path: "/api/http/routers/bar@myprovider",
			conf: config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"bar@myprovider": {
						Router: &config.Router{
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
			conf: config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"bar@myprovider": {
						Router: &config.Router{
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
			conf: config.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "all services, but no config",
			path: "/api/http/services",
			conf: config.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/services-empty.json",
			},
		},
		{
			desc: "all services",
			path: "/api/http/services",
			conf: config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"bar@myprovider": func() *config.ServiceInfo {
						si := &config.ServiceInfo{
							Service: &config.Service{
								LoadBalancer: &config.LoadBalancerService{
									Servers: []config.Server{
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
					"baz@myprovider": func() *config.ServiceInfo {
						si := &config.ServiceInfo{
							Service: &config.Service{
								LoadBalancer: &config.LoadBalancerService{
									Servers: []config.Server{
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
			conf: config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"bar@myprovider": func() *config.ServiceInfo {
						si := &config.ServiceInfo{
							Service: &config.Service{
								LoadBalancer: &config.LoadBalancerService{
									Servers: []config.Server{
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
					"baz@myprovider": func() *config.ServiceInfo {
						si := &config.ServiceInfo{
							Service: &config.Service{
								LoadBalancer: &config.LoadBalancerService{
									Servers: []config.Server{
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
					"test@myprovider": func() *config.ServiceInfo {
						si := &config.ServiceInfo{
							Service: &config.Service{
								LoadBalancer: &config.LoadBalancerService{
									Servers: []config.Server{
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
			conf: config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"bar@myprovider": func() *config.ServiceInfo {
						si := &config.ServiceInfo{
							Service: &config.Service{
								LoadBalancer: &config.LoadBalancerService{
									Servers: []config.Server{
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
			conf: config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"bar@myprovider": func() *config.ServiceInfo {
						si := &config.ServiceInfo{
							Service: &config.Service{
								LoadBalancer: &config.LoadBalancerService{
									Servers: []config.Server{
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
			conf: config.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "all middlewares, but no config",
			path: "/api/http/middlewares",
			conf: config.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusOK,
				nextPage:   "1",
				jsonFile:   "testdata/middlewares-empty.json",
			},
		},
		{
			desc: "all middlewares",
			path: "/api/http/middlewares",
			conf: config.RuntimeConfiguration{
				Middlewares: map[string]*config.MiddlewareInfo{
					"auth@myprovider": {
						Middleware: &config.Middleware{
							BasicAuth: &config.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
						UsedBy: []string{"bar@myprovider", "test@myprovider"},
					},
					"addPrefixTest@myprovider": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/titi",
							},
						},
						UsedBy: []string{"test@myprovider"},
					},
					"addPrefixTest@anotherprovider": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
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
			conf: config.RuntimeConfiguration{
				Middlewares: map[string]*config.MiddlewareInfo{
					"auth@myprovider": {
						Middleware: &config.Middleware{
							BasicAuth: &config.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
						UsedBy: []string{"bar@myprovider", "test@myprovider"},
					},
					"addPrefixTest@myprovider": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/titi",
							},
						},
						UsedBy: []string{"test@myprovider"},
					},
					"addPrefixTest@anotherprovider": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
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
			conf: config.RuntimeConfiguration{
				Middlewares: map[string]*config.MiddlewareInfo{
					"auth@myprovider": {
						Middleware: &config.Middleware{
							BasicAuth: &config.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
						UsedBy: []string{"bar@myprovider", "test@myprovider"},
					},
					"addPrefixTest@myprovider": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/titi",
							},
						},
						UsedBy: []string{"test@myprovider"},
					},
					"addPrefixTest@anotherprovider": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
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
			conf: config.RuntimeConfiguration{
				Middlewares: map[string]*config.MiddlewareInfo{
					"auth@myprovider": {
						Middleware: &config.Middleware{
							BasicAuth: &config.BasicAuth{
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
			conf: config.RuntimeConfiguration{},
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

func TestHandler_Configuration(t *testing.T) {
	type expected struct {
		statusCode int
		json       string
	}

	testCases := []struct {
		desc     string
		path     string
		conf     config.RuntimeConfiguration
		expected expected
	}{
		{
			desc: "Get rawdata",
			path: "/api/rawdata",
			conf: config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"foo-service@myprovider": {
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
					"auth@myprovider": {
						Middleware: &config.Middleware{
							BasicAuth: &config.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
					},
					"addPrefixTest@myprovider": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/titi",
							},
						},
					},
					"addPrefixTest@anotherprovider": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/toto",
							},
						},
					},
				},
				Routers: map[string]*config.RouterInfo{
					"bar@myprovider": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "addPrefixTest@anotherprovider"},
						},
					},
					"test@myprovider": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							Middlewares: []string{"addPrefixTest", "auth"},
						},
					},
				},
				TCPServices: map[string]*config.TCPServiceInfo{
					"tcpfoo-service@myprovider": {
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
					"tcpbar@myprovider": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "tcpfoo-service@myprovider",
							Rule:        "HostSNI(`foo.bar`)",
						},
					},
					"tcptest@myprovider": {
						TCPRouter: &config.TCPRouter{
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

func generateHTTPRouters(nbRouters int) map[string]*config.RouterInfo {
	routers := make(map[string]*config.RouterInfo, nbRouters)
	for i := 0; i < nbRouters; i++ {
		routers[fmt.Sprintf("bar%2d@myprovider", i)] = &config.RouterInfo{
			Router: &config.Router{
				EntryPoints: []string{"web"},
				Service:     "foo-service@myprovider",
				Rule:        "Host(`foo.bar" + strconv.Itoa(i) + "`)",
			},
		}
	}
	return routers
}
