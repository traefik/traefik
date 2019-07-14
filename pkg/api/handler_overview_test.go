package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/mux"
	"github.com/containous/traefik/pkg/config/dynamic"
	"github.com/containous/traefik/pkg/config/static"
	"github.com/containous/traefik/pkg/provider/docker"
	"github.com/containous/traefik/pkg/provider/file"
	"github.com/containous/traefik/pkg/provider/kubernetes/crd"
	"github.com/containous/traefik/pkg/provider/kubernetes/ingress"
	"github.com/containous/traefik/pkg/provider/marathon"
	"github.com/containous/traefik/pkg/provider/rancher"
	"github.com/containous/traefik/pkg/provider/rest"
	"github.com/containous/traefik/pkg/tracing/jaeger"
	"github.com/containous/traefik/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Overview(t *testing.T) {
	type expected struct {
		statusCode int
		jsonFile   string
	}

	testCases := []struct {
		desc       string
		path       string
		confStatic static.Configuration
		confDyn    dynamic.RuntimeConfiguration
		expected   expected
	}{
		{
			desc:       "without data in the dynamic configuration",
			path:       "/api/overview",
			confStatic: static.Configuration{API: &static.API{}, Global: &static.Global{}},
			confDyn:    dynamic.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/overview-empty.json",
			},
		},
		{
			desc:       "with data in the dynamic configuration",
			path:       "/api/overview",
			confStatic: static.Configuration{API: &static.API{}, Global: &static.Global{}},
			confDyn: dynamic.RuntimeConfiguration{
				Services: map[string]*dynamic.ServiceInfo{
					"foo-service@myprovider": {
						Service: &dynamic.Service{
							LoadBalancer: &dynamic.LoadBalancerService{
								Servers: []dynamic.Server{{URL: "http://127.0.0.1"}},
							},
						},
						Status: dynamic.RuntimeStatusEnabled,
					},
					"bar-service@myprovider": {
						Service: &dynamic.Service{
							LoadBalancer: &dynamic.LoadBalancerService{
								Servers: []dynamic.Server{{URL: "http://127.0.0.1"}},
							},
						},
						Status: dynamic.RuntimeStatusWarning,
					},
					"fii-service@myprovider": {
						Service: &dynamic.Service{
							LoadBalancer: &dynamic.LoadBalancerService{
								Servers: []dynamic.Server{{URL: "http://127.0.0.1"}},
							},
						},
						Status: dynamic.RuntimeStatusDisabled,
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
						Err: []string{"error"},
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
						Status: dynamic.RuntimeStatusEnabled,
					},
					"test@myprovider": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							Middlewares: []string{"addPrefixTest", "auth"},
						},
						Status: dynamic.RuntimeStatusWarning,
					},
					"foo@myprovider": {
						Router: &dynamic.Router{
							EntryPoints: []string{"web"},
							Service:     "foo-service@myprovider",
							Rule:        "Host(`foo.bar.other`)",
							Middlewares: []string{"addPrefixTest", "auth"},
						},
						Status: dynamic.RuntimeStatusDisabled,
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
				jsonFile:   "testdata/overview-dynamic.json",
			},
		},
		{
			desc: "with providers",
			path: "/api/overview",
			confStatic: static.Configuration{
				Global: &static.Global{},
				API:    &static.API{},
				Providers: &static.Providers{
					Docker:            &docker.Provider{},
					File:              &file.Provider{},
					Marathon:          &marathon.Provider{},
					KubernetesIngress: &ingress.Provider{},
					KubernetesCRD:     &crd.Provider{},
					Rest:              &rest.Provider{},
					Rancher:           &rancher.Provider{},
				},
			},
			confDyn: dynamic.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/overview-providers.json",
			},
		},
		{
			desc: "with features",
			path: "/api/overview",
			confStatic: static.Configuration{
				Global: &static.Global{},
				API:    &static.API{},
				Metrics: &types.Metrics{
					Prometheus: &types.Prometheus{},
				},
				Tracing: &static.Tracing{
					Jaeger: &jaeger.Config{},
				},
			},
			confDyn: dynamic.RuntimeConfiguration{},
			expected: expected{
				statusCode: http.StatusOK,
				jsonFile:   "testdata/overview-features.json",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			handler := New(test.confStatic, &test.confDyn)
			router := mux.NewRouter()
			handler.Append(router)

			server := httptest.NewServer(router)

			resp, err := http.DefaultClient.Get(server.URL + test.path)
			require.NoError(t, err)

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
