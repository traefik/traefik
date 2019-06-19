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
					"myprovider.test": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar.other`)",
							TLS: &config.RouterTCPTLSConfig{
								Passthrough: false,
							},
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
					"myprovider.bar": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
						},
					},
					"myprovider.baz": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`toto.bar`)",
						},
					},
					"myprovider.test": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
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
			path: "/api/tcp/routers/myprovider.bar",
			conf: config.RuntimeConfiguration{
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
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/tcprouter-bar.json",
			},
		},
		{
			desc: "one TCP router by id, that does not exist",
			path: "/api/tcp/routers/myprovider.foo",
			conf: config.RuntimeConfiguration{
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
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "one TCP router by id, but no config",
			path: "/api/tcp/routers/myprovider.bar",
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
					"myprovider.bar": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1:2345",
									},
								},
							},
						},
						UsedBy: []string{"myprovider.foo", "myprovider.test"},
					},
					"myprovider.baz": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.2:2345",
									},
								},
							},
						},
						UsedBy: []string{"myprovider.foo"},
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
					"myprovider.bar": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1:2345",
									},
								},
							},
						},
						UsedBy: []string{"myprovider.foo", "myprovider.test"},
					},
					"myprovider.baz": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.2:2345",
									},
								},
							},
						},
						UsedBy: []string{"myprovider.foo"},
					},
					"myprovider.test": {
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
			path: "/api/tcp/services/myprovider.bar",
			conf: config.RuntimeConfiguration{
				TCPServices: map[string]*config.TCPServiceInfo{
					"myprovider.bar": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1:2345",
									},
								},
							},
						},
						UsedBy: []string{"myprovider.foo", "myprovider.test"},
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
			path: "/api/tcp/services/myprovider.nono",
			conf: config.RuntimeConfiguration{
				TCPServices: map[string]*config.TCPServiceInfo{
					"myprovider.bar": {
						TCPService: &config.TCPService{
							LoadBalancer: &config.TCPLoadBalancerService{
								Servers: []config.TCPServer{
									{
										Address: "127.0.0.1:2345",
									},
								},
							},
						},
						UsedBy: []string{"myprovider.foo", "myprovider.test"},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "one tcp service by id, but no config",
			path: "/api/tcp/services/myprovider.foo",
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
					"myprovider.test": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar.other`)",
							Middlewares: []string{"addPrefixTest", "auth"},
						},
					},
					"myprovider.bar": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "anotherprovider.addPrefixTest"},
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
					"myprovider.bar": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "anotherprovider.addPrefixTest"},
						},
					},
					"myprovider.baz": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`toto.bar`)",
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
			path: "/api/http/routers/myprovider.bar",
			conf: config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"myprovider.bar": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "anotherprovider.addPrefixTest"},
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
			path: "/api/http/routers/myprovider.foo",
			conf: config.RuntimeConfiguration{
				Routers: map[string]*config.RouterInfo{
					"myprovider.bar": {
						Router: &config.Router{
							EntryPoints: []string{"web"},
							Service:     "myprovider.foo-service",
							Rule:        "Host(`foo.bar`)",
							Middlewares: []string{"auth", "anotherprovider.addPrefixTest"},
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
			path: "/api/http/routers/myprovider.foo",
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
					"myprovider.bar": func() *config.ServiceInfo {
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
							UsedBy: []string{"myprovider.foo", "myprovider.test"},
						}
						si.UpdateStatus("http://127.0.0.1", "UP")
						return si
					}(),
					"myprovider.baz": func() *config.ServiceInfo {
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
							UsedBy: []string{"myprovider.foo"},
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
					"myprovider.bar": func() *config.ServiceInfo {
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
							UsedBy: []string{"myprovider.foo", "myprovider.test"},
						}
						si.UpdateStatus("http://127.0.0.1", "UP")
						return si
					}(),
					"myprovider.baz": func() *config.ServiceInfo {
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
							UsedBy: []string{"myprovider.foo"},
						}
						si.UpdateStatus("http://127.0.0.2", "UP")
						return si
					}(),
					"myprovider.test": func() *config.ServiceInfo {
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
							UsedBy: []string{"myprovider.foo", "myprovider.test"},
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
			path: "/api/http/services/myprovider.bar",
			conf: config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"myprovider.bar": func() *config.ServiceInfo {
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
							UsedBy: []string{"myprovider.foo", "myprovider.test"},
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
			path: "/api/http/services/myprovider.nono",
			conf: config.RuntimeConfiguration{
				Services: map[string]*config.ServiceInfo{
					"myprovider.bar": func() *config.ServiceInfo {
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
							UsedBy: []string{"myprovider.foo", "myprovider.test"},
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
			path: "/api/http/services/myprovider.foo",
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
					"myprovider.auth": {
						Middleware: &config.Middleware{
							BasicAuth: &config.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
						UsedBy: []string{"myprovider.bar", "myprovider.test"},
					},
					"myprovider.addPrefixTest": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/titi",
							},
						},
						UsedBy: []string{"myprovider.test"},
					},
					"anotherprovider.addPrefixTest": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/toto",
							},
						},
						UsedBy: []string{"myprovider.bar"},
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
					"myprovider.auth": {
						Middleware: &config.Middleware{
							BasicAuth: &config.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
						UsedBy: []string{"myprovider.bar", "myprovider.test"},
					},
					"myprovider.addPrefixTest": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/titi",
							},
						},
						UsedBy: []string{"myprovider.test"},
					},
					"anotherprovider.addPrefixTest": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/toto",
							},
						},
						UsedBy: []string{"myprovider.bar"},
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
			path: "/api/http/middlewares/myprovider.auth",
			conf: config.RuntimeConfiguration{
				Middlewares: map[string]*config.MiddlewareInfo{
					"myprovider.auth": {
						Middleware: &config.Middleware{
							BasicAuth: &config.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
						UsedBy: []string{"myprovider.bar", "myprovider.test"},
					},
					"myprovider.addPrefixTest": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/titi",
							},
						},
						UsedBy: []string{"myprovider.test"},
					},
					"anotherprovider.addPrefixTest": {
						Middleware: &config.Middleware{
							AddPrefix: &config.AddPrefix{
								Prefix: "/toto",
							},
						},
						UsedBy: []string{"myprovider.bar"},
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
			path: "/api/http/middlewares/myprovider.foo",
			conf: config.RuntimeConfiguration{
				Middlewares: map[string]*config.MiddlewareInfo{
					"myprovider.auth": {
						Middleware: &config.Middleware{
							BasicAuth: &config.BasicAuth{
								Users: []string{"admin:admin"},
							},
						},
						UsedBy: []string{"myprovider.bar", "myprovider.test"},
					},
				},
			},
			expected: expected{
				statusCode: http.StatusNotFound,
			},
		},
		{
			desc: "one middleware by id, but no config",
			path: "/api/http/middlewares/myprovider.foo",
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
				TCPServices: map[string]*config.TCPServiceInfo{
					"myprovider.tcpfoo-service": {
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
					"myprovider.tcpbar": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.tcpfoo-service",
							Rule:        "HostSNI(`foo.bar`)",
						},
					},
					"myprovider.tcptest": {
						TCPRouter: &config.TCPRouter{
							EntryPoints: []string{"web"},
							Service:     "myprovider.tcpfoo-service",
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
		routers[fmt.Sprintf("myprovider.bar%2d", i)] = &config.RouterInfo{
			Router: &config.Router{
				EntryPoints: []string{"web"},
				Service:     "myprovider.foo-service",
				Rule:        "Host(`foo.bar" + strconv.Itoa(i) + "`)",
			},
		}
	}
	return routers
}
