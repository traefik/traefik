package api

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/mux"
	"github.com/containous/traefik/config"
	"github.com/containous/traefik/safe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Configuration(t *testing.T) {
	type expected struct {
		statusCode int
		body       string
	}

	testCases := []struct {
		desc          string
		path          string
		configuration config.Configurations
		expected      expected
	}{
		{
			desc: "Get all the providers",
			path: "/api/providers",
			configuration: config.Configurations{
				"foo": {
					HTTP: &config.HTTPConfiguration{
						Routers: map[string]*config.Router{
							"bar": {EntryPoints: []string{"foo", "bar"}},
						},
					},
				},
			},
			expected: expected{statusCode: http.StatusOK, body: `[{"id":"foo","path":"/api/providers/foo"}]`},
		},
		{
			desc: "Get a provider",
			path: "/api/providers/foo",
			configuration: config.Configurations{
				"foo": {
					HTTP: &config.HTTPConfiguration{
						Routers: map[string]*config.Router{
							"bar": {EntryPoints: []string{"foo", "bar"}},
						},
						Middlewares: map[string]*config.Middleware{
							"bar": {
								AddPrefix: &config.AddPrefix{Prefix: "bar"},
							},
						},
						Services: map[string]*config.Service{
							"foo": {
								LoadBalancer: &config.LoadBalancerService{
									Method: "wrr",
								},
							},
						},
					},
				},
			},
			expected: expected{statusCode: http.StatusOK, body: `{"routers":[{"id":"bar","path":"/api/providers/foo/routers"}],"middlewares":[{"id":"bar","path":"/api/providers/foo/middlewares"}],"services":[{"id":"foo","path":"/api/providers/foo/services"}]}`},
		},
		{
			desc:          "Provider not found",
			path:          "/api/providers/foo",
			configuration: config.Configurations{},
			expected:      expected{statusCode: http.StatusNotFound, body: "404 page not found\n"},
		},
		{
			desc: "Get all routers",
			path: "/api/providers/foo/routers",
			configuration: config.Configurations{
				"foo": {
					HTTP: &config.HTTPConfiguration{
						Routers: map[string]*config.Router{
							"bar": {EntryPoints: []string{"foo", "bar"}},
						},
					},
				},
			},
			expected: expected{statusCode: http.StatusOK, body: `[{"entryPoints":["foo","bar"],"id":"bar"}]`},
		},
		{
			desc: "Get a router",
			path: "/api/providers/foo/routers/bar",
			configuration: config.Configurations{
				"foo": {
					HTTP: &config.HTTPConfiguration{
						Routers: map[string]*config.Router{
							"bar": {EntryPoints: []string{"foo", "bar"}},
						},
					},
				},
			},
			expected: expected{statusCode: http.StatusOK, body: `{"entryPoints":["foo","bar"]}`},
		},
		{
			desc: "Router not found",
			path: "/api/providers/foo/routers/bar",
			configuration: config.Configurations{
				"foo": {},
			},
			expected: expected{statusCode: http.StatusNotFound, body: "404 page not found\n"},
		},
		{
			desc: "Get all services",
			path: "/api/providers/foo/services",
			configuration: config.Configurations{
				"foo": {
					HTTP: &config.HTTPConfiguration{
						Services: map[string]*config.Service{
							"foo": {
								LoadBalancer: &config.LoadBalancerService{
									Method: "wrr",
								},
							},
						},
					},
				},
			},
			expected: expected{statusCode: http.StatusOK, body: `[{"loadbalancer":{"method":"wrr","passHostHeader":false},"id":"foo"}]`},
		},
		{
			desc: "Get a service",
			path: "/api/providers/foo/services/foo",
			configuration: config.Configurations{
				"foo": {
					HTTP: &config.HTTPConfiguration{
						Services: map[string]*config.Service{
							"foo": {
								LoadBalancer: &config.LoadBalancerService{
									Method: "wrr",
								},
							},
						},
					},
				},
			},
			expected: expected{statusCode: http.StatusOK, body: `{"loadbalancer":{"method":"wrr","passHostHeader":false}}`},
		},
		{
			desc: "Service not found",
			path: "/api/providers/foo/services/bar",
			configuration: config.Configurations{
				"foo": {},
			},
			expected: expected{statusCode: http.StatusNotFound, body: "404 page not found\n"},
		},
		{
			desc: "Get all middlewares",
			path: "/api/providers/foo/middlewares",
			configuration: config.Configurations{
				"foo": {
					HTTP: &config.HTTPConfiguration{
						Middlewares: map[string]*config.Middleware{
							"bar": {
								AddPrefix: &config.AddPrefix{Prefix: "bar"},
							},
						},
					},
				},
			},
			expected: expected{statusCode: http.StatusOK, body: `[{"addPrefix":{"prefix":"bar"},"id":"bar"}]`},
		},
		{
			desc: "Get a middleware",
			path: "/api/providers/foo/middlewares/bar",
			configuration: config.Configurations{
				"foo": {
					HTTP: &config.HTTPConfiguration{
						Middlewares: map[string]*config.Middleware{
							"bar": {
								AddPrefix: &config.AddPrefix{Prefix: "bar"},
							},
						},
					},
				},
			},
			expected: expected{statusCode: http.StatusOK, body: `{"addPrefix":{"prefix":"bar"}}`},
		},
		{
			desc: "Middleware not found",
			path: "/api/providers/foo/middlewares/bar",
			configuration: config.Configurations{
				"foo": {},
			},
			expected: expected{statusCode: http.StatusNotFound, body: "404 page not found\n"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			currentConfiguration := &safe.Safe{}
			currentConfiguration.Set(test.configuration)

			handler := Handler{
				CurrentConfigurations: currentConfiguration,
			}

			router := mux.NewRouter()
			handler.Append(router)

			server := httptest.NewServer(router)

			resp, err := http.DefaultClient.Get(server.URL + test.path)
			require.NoError(t, err)

			assert.Equal(t, test.expected.statusCode, resp.StatusCode)

			content, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			err = resp.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, test.expected.body, string(content))
		})
	}
}
