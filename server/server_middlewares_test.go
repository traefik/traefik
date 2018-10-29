package server

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/containous/mux"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/metrics"
	"github.com/containous/traefik/middlewares"
	th "github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"
)

func TestServerEntryPointWhitelistConfig(t *testing.T) {
	testCases := []struct {
		desc             string
		entrypoint       *configuration.EntryPoint
		expectMiddleware bool
	}{
		{
			desc: "no whitelist middleware if no config on entrypoint",
			entrypoint: &configuration.EntryPoint{
				Address:          ":0",
				ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true},
			},
			expectMiddleware: false,
		},
		{
			desc: "whitelist middleware should be added if configured on entrypoint",
			entrypoint: &configuration.EntryPoint{
				Address: ":0",
				WhitelistSourceRange: []string{
					"127.0.0.1/32",
				},
				ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true},
			},
			expectMiddleware: true,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			srv := Server{
				globalConfiguration: configuration.GlobalConfiguration{},
				metricsRegistry:     metrics.NewVoidRegistry(),
				entryPoints: map[string]EntryPoint{
					"test": {
						Configuration: test.entrypoint,
					},
				},
			}

			srv.serverEntryPoints = srv.buildServerEntryPoints()
			srvEntryPoint := srv.setupServerEntryPoint("test", srv.serverEntryPoints["test"])
			handler := srvEntryPoint.httpServer.Handler.(*mux.Router).NotFoundHandler.(*negroni.Negroni)

			found := false
			for _, handler := range handler.Handlers() {
				if reflect.TypeOf(handler) == reflect.TypeOf((*middlewares.IPWhiteLister)(nil)) {
					found = true
				}
			}

			if found && !test.expectMiddleware {
				t.Error("ip whitelist middleware was installed even though it should not")
			}

			if !found && test.expectMiddleware {
				t.Error("ip whitelist middleware was not installed even though it should have")
			}
		})
	}
}

func TestBuildIPWhiteLister(t *testing.T) {
	testCases := []struct {
		desc                 string
		whitelistSourceRange []string
		whiteList            *types.WhiteList
		middlewareConfigured bool
		errMessage           string
	}{
		{
			desc:                 "no whitelists configured",
			whitelistSourceRange: nil,
			middlewareConfigured: false,
			errMessage:           "",
		},
		{
			desc: "whitelists configured (deprecated)",
			whitelistSourceRange: []string{
				"1.2.3.4/24",
				"fe80::/16",
			},
			middlewareConfigured: true,
			errMessage:           "",
		},
		{
			desc: "invalid whitelists configured (deprecated)",
			whitelistSourceRange: []string{
				"foo",
			},
			middlewareConfigured: false,
			errMessage:           "parsing CIDR whitelist [foo]: parsing CIDR white list <nil>: invalid CIDR address: foo",
		},
		{
			desc: "whitelists configured",
			whiteList: &types.WhiteList{
				SourceRange: []string{
					"1.2.3.4/24",
					"fe80::/16",
				},
				UseXForwardedFor: false,
			},
			middlewareConfigured: true,
			errMessage:           "",
		},
		{
			desc: "invalid whitelists configured (deprecated)",
			whiteList: &types.WhiteList{
				SourceRange: []string{
					"foo",
				},
				UseXForwardedFor: false,
			},
			middlewareConfigured: false,
			errMessage:           "parsing CIDR whitelist [foo]: parsing CIDR white list <nil>: invalid CIDR address: foo",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			middleware, err := buildIPWhiteLister(test.whiteList, test.whitelistSourceRange)

			if test.errMessage != "" {
				require.EqualError(t, err, test.errMessage)
			} else {
				assert.NoError(t, err)

				if test.middlewareConfigured {
					require.NotNil(t, middleware, "not expected middleware to be configured")
				} else {
					require.Nil(t, middleware, "expected middleware to be configured")
				}
			}
		})
	}
}

func TestBuildRedirectHandler(t *testing.T) {
	srv := Server{
		globalConfiguration: configuration.GlobalConfiguration{},
		entryPoints: map[string]EntryPoint{
			"http":  {Configuration: &configuration.EntryPoint{Address: ":80"}},
			"https": {Configuration: &configuration.EntryPoint{Address: ":443", TLS: &tls.TLS{}}},
		},
	}

	testCases := []struct {
		desc              string
		srcEntryPointName string
		url               string
		entryPoint        *configuration.EntryPoint
		redirect          *types.Redirect
		expectedURL       string
	}{
		{
			desc:              "redirect regex",
			srcEntryPointName: "http",
			url:               "http://foo.com",
			redirect: &types.Redirect{
				Regex:       `^(?:http?:\/\/)(foo)(\.com)$`,
				Replacement: "https://$1{{\"bar\"}}$2",
			},
			entryPoint: &configuration.EntryPoint{
				Address: ":80",
				Redirect: &types.Redirect{
					Regex:       `^(?:http?:\/\/)(foo)(\.com)$`,
					Replacement: "https://$1{{\"bar\"}}$2",
				},
			},
			expectedURL: "https://foobar.com",
		},
		{
			desc:              "redirect entry point",
			srcEntryPointName: "http",
			url:               "http://foo:80",
			redirect: &types.Redirect{
				EntryPoint: "https",
			},
			entryPoint: &configuration.EntryPoint{
				Address: ":80",
				Redirect: &types.Redirect{
					EntryPoint: "https",
				},
			},
			expectedURL: "https://foo:443",
		},
		{
			desc:              "redirect entry point with regex (ignored)",
			srcEntryPointName: "http",
			url:               "http://foo.com:80",
			redirect: &types.Redirect{
				EntryPoint:  "https",
				Regex:       `^(?:http?:\/\/)(foo)(\.com)$`,
				Replacement: "https://$1{{\"bar\"}}$2",
			},
			entryPoint: &configuration.EntryPoint{
				Address: ":80",
				Redirect: &types.Redirect{
					EntryPoint:  "https",
					Regex:       `^(?:http?:\/\/)(foo)(\.com)$`,
					Replacement: "https://$1{{\"bar\"}}$2",
				},
			},
			expectedURL: "https://foo.com:443",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			rewrite, err := srv.buildRedirectHandler(test.srcEntryPointName, test.redirect)
			require.NoError(t, err)

			req := th.MustNewRequest(http.MethodGet, test.url, nil)
			recorder := httptest.NewRecorder()

			rewrite.ServeHTTP(recorder, req, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Add("Location", "fail")
			}))

			location, err := recorder.Result().Location()
			require.NoError(t, err)
			assert.Equal(t, test.expectedURL, location.String())
		})
	}
}

func TestServerGenericFrontendAuthFail(t *testing.T) {
	globalConfig := configuration.GlobalConfiguration{
		EntryPoints: configuration.EntryPoints{
			"http": &configuration.EntryPoint{ForwardedHeaders: &configuration.ForwardedHeaders{Insecure: true}},
		},
	}

	dynamicConfigs := types.Configurations{
		"config": &types.Configuration{
			Frontends: map[string]*types.Frontend{
				"frontend": {
					EntryPoints: []string{"http"},
					Backend:     "backend",
					BasicAuth:   []string{""},
				},
			},
			Backends: map[string]*types.Backend{
				"backend": {
					Servers: map[string]types.Server{
						"server": {
							URL: "http://localhost",
						},
					},
					LoadBalancer: &types.LoadBalancer{
						Method: "Wrr",
					},
				},
			},
		},
	}

	srv := NewServer(globalConfig, nil, nil)

	_ = srv.loadConfig(dynamicConfigs, globalConfig)
}
