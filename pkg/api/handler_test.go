package api

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/mux"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/config/static"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var updateExpected = flag.Bool("update_expected", false, "Update expected files in testdata")

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

			rtConf := &test.conf
			handler := New(static.Configuration{API: &static.API{}, Global: &static.Global{}}, rtConf)
			router := mux.NewRouter()
			handler.Append(router)
			rtConf.PopulateUsedBy()

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
