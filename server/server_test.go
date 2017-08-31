package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/mux"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/healthcheck"
	"github.com/containous/traefik/metrics"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/negroni"
	"github.com/vulcand/oxy/roundrobin"
)

type testLoadBalancer struct{}

func (lb *testLoadBalancer) RemoveServer(u *url.URL) error {
	return nil
}

func (lb *testLoadBalancer) UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error {
	return nil
}

func (lb *testLoadBalancer) Servers() []*url.URL {
	return []*url.URL{}
}

func TestPrepareServerTimeouts(t *testing.T) {
	tests := []struct {
		desc             string
		globalConfig     configuration.GlobalConfiguration
		wantIdleTimeout  time.Duration
		wantReadTimeout  time.Duration
		wantWriteTimeout time.Duration
	}{
		{
			desc: "full configuration",
			globalConfig: configuration.GlobalConfiguration{
				RespondingTimeouts: &configuration.RespondingTimeouts{
					IdleTimeout:  flaeg.Duration(10 * time.Second),
					ReadTimeout:  flaeg.Duration(12 * time.Second),
					WriteTimeout: flaeg.Duration(14 * time.Second),
				},
			},
			wantIdleTimeout:  time.Duration(10 * time.Second),
			wantReadTimeout:  time.Duration(12 * time.Second),
			wantWriteTimeout: time.Duration(14 * time.Second),
		},
		{
			desc:             "using defaults",
			globalConfig:     configuration.GlobalConfiguration{},
			wantIdleTimeout:  time.Duration(180 * time.Second),
			wantReadTimeout:  time.Duration(0 * time.Second),
			wantWriteTimeout: time.Duration(0 * time.Second),
		},
		{
			desc: "deprecated IdleTimeout configured",
			globalConfig: configuration.GlobalConfiguration{
				IdleTimeout: flaeg.Duration(45 * time.Second),
			},
			wantIdleTimeout:  time.Duration(45 * time.Second),
			wantReadTimeout:  time.Duration(0 * time.Second),
			wantWriteTimeout: time.Duration(0 * time.Second),
		},
		{
			desc: "deprecated and new IdleTimeout configured",
			globalConfig: configuration.GlobalConfiguration{
				IdleTimeout: flaeg.Duration(45 * time.Second),
				RespondingTimeouts: &configuration.RespondingTimeouts{
					IdleTimeout: flaeg.Duration(80 * time.Second),
				},
			},
			wantIdleTimeout:  time.Duration(80 * time.Second),
			wantReadTimeout:  time.Duration(0 * time.Second),
			wantWriteTimeout: time.Duration(0 * time.Second),
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			entryPointName := "http"
			entryPoint := &configuration.EntryPoint{Address: "localhost:8080"}
			router := middlewares.NewHandlerSwitcher(mux.NewRouter())

			srv := NewServer(test.globalConfig)
			httpServer, _, err := srv.prepareServer(entryPointName, entryPoint, router)
			if err != nil {
				t.Fatalf("Unexpected error when preparing srv: %s", err)
			}

			if httpServer.IdleTimeout != test.wantIdleTimeout {
				t.Errorf("Got %s as IdleTimeout, want %s", httpServer.IdleTimeout, test.wantIdleTimeout)
			}
			if httpServer.ReadTimeout != test.wantReadTimeout {
				t.Errorf("Got %s as ReadTimeout, want %s", httpServer.ReadTimeout, test.wantReadTimeout)
			}
			if httpServer.WriteTimeout != test.wantWriteTimeout {
				t.Errorf("Got %s as WriteTimeout, want %s", httpServer.WriteTimeout, test.wantWriteTimeout)
			}
		})
	}
}

func TestServerMultipleFrontendRules(t *testing.T) {
	cases := []struct {
		expression  string
		requestURL  string
		expectedURL string
	}{
		{
			expression:  "Host:foo.bar",
			requestURL:  "http://foo.bar",
			expectedURL: "http://foo.bar",
		},
		{
			expression:  "PathPrefix:/management;ReplacePath:/health",
			requestURL:  "http://foo.bar/management",
			expectedURL: "http://foo.bar/health",
		},
		{
			expression:  "Host:foo.bar;AddPrefix:/blah",
			requestURL:  "http://foo.bar/baz",
			expectedURL: "http://foo.bar/blah/baz",
		},
		{
			expression:  "PathPrefixStripRegex:/one/{two}/{three:[0-9]+}",
			requestURL:  "http://foo.bar/one/some/12345/four",
			expectedURL: "http://foo.bar/four",
		},
		{
			expression:  "PathPrefixStripRegex:/one/{two}/{three:[0-9]+};AddPrefix:/zero",
			requestURL:  "http://foo.bar/one/some/12345/four",
			expectedURL: "http://foo.bar/zero/four",
		},
		{
			expression:  "AddPrefix:/blah;ReplacePath:/baz",
			requestURL:  "http://foo.bar/hello",
			expectedURL: "http://foo.bar/baz",
		},
		{
			expression:  "PathPrefixStrip:/management;ReplacePath:/health",
			requestURL:  "http://foo.bar/management",
			expectedURL: "http://foo.bar/health",
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.expression, func(t *testing.T) {
			t.Parallel()

			router := mux.NewRouter()
			route := router.NewRoute()
			serverRoute := &serverRoute{route: route}
			rules := &Rules{route: serverRoute}

			expression := test.expression
			routeResult, err := rules.Parse(expression)

			if err != nil {
				t.Fatalf("Error while building route for %s: %+v", expression, err)
			}

			request := testhelpers.MustNewRequest(http.MethodGet, test.requestURL, nil)
			routeMatch := routeResult.Match(request, &mux.RouteMatch{Route: routeResult})

			if !routeMatch {
				t.Fatalf("Rule %s doesn't match", expression)
			}

			server := new(Server)

			server.wireFrontendBackend(serverRoute, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.String() != test.expectedURL {
					t.Fatalf("got URL %s, expected %s", r.URL.String(), test.expectedURL)
				}
			}))
			serverRoute.route.GetHandler().ServeHTTP(nil, request)
		})
	}
}

func TestServerLoadConfigHealthCheckOptions(t *testing.T) {
	healthChecks := []*types.HealthCheck{
		nil,
		{
			Path: "/path",
		},
	}

	for _, lbMethod := range []string{"Wrr", "Drr"} {
		for _, healthCheck := range healthChecks {
			t.Run(fmt.Sprintf("%s/hc=%t", lbMethod, healthCheck != nil), func(t *testing.T) {
				globalConfig := configuration.GlobalConfiguration{
					EntryPoints: configuration.EntryPoints{
						"http": &configuration.EntryPoint{},
					},
					HealthCheck: &configuration.HealthCheckConfig{Interval: flaeg.Duration(5 * time.Second)},
				}

				dynamicConfigs := types.Configurations{
					"config": &types.Configuration{
						Frontends: map[string]*types.Frontend{
							"frontend": {
								EntryPoints: []string{"http"},
								Backend:     "backend",
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
									Method: lbMethod,
								},
								HealthCheck: healthCheck,
							},
						},
					},
				}

				srv := NewServer(globalConfig)
				if _, err := srv.loadConfig(dynamicConfigs, globalConfig); err != nil {
					t.Fatalf("got error: %s", err)
				}

				wantNumHealthCheckBackends := 0
				if healthCheck != nil {
					wantNumHealthCheckBackends = 1
				}
				gotNumHealthCheckBackends := len(healthcheck.GetHealthCheck().Backends)
				if gotNumHealthCheckBackends != wantNumHealthCheckBackends {
					t.Errorf("got %d health check backends, want %d", gotNumHealthCheckBackends, wantNumHealthCheckBackends)
				}
			})
		}
	}
}

func TestServerParseHealthCheckOptions(t *testing.T) {
	lb := &testLoadBalancer{}
	globalInterval := 15 * time.Second

	tests := []struct {
		desc     string
		hc       *types.HealthCheck
		wantOpts *healthcheck.Options
	}{
		{
			desc:     "nil health check",
			hc:       nil,
			wantOpts: nil,
		},
		{
			desc: "empty path",
			hc: &types.HealthCheck{
				Path: "",
			},
			wantOpts: nil,
		},
		{
			desc: "unparseable interval",
			hc: &types.HealthCheck{
				Path:     "/path",
				Interval: "unparseable",
			},
			wantOpts: &healthcheck.Options{
				Path:     "/path",
				Interval: globalInterval,
				LB:       lb,
			},
		},
		{
			desc: "sub-zero interval",
			hc: &types.HealthCheck{
				Path:     "/path",
				Interval: "-42s",
			},
			wantOpts: &healthcheck.Options{
				Path:     "/path",
				Interval: globalInterval,
				LB:       lb,
			},
		},
		{
			desc: "parseable interval",
			hc: &types.HealthCheck{
				Path:     "/path",
				Interval: "5m",
			},
			wantOpts: &healthcheck.Options{
				Path:     "/path",
				Interval: 5 * time.Minute,
				LB:       lb,
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			gotOpts := parseHealthCheckOptions(lb, "backend", test.hc, &configuration.HealthCheckConfig{Interval: flaeg.Duration(globalInterval)})
			if !reflect.DeepEqual(gotOpts, test.wantOpts) {
				t.Errorf("got health check options %+v, want %+v", gotOpts, test.wantOpts)
			}
		})
	}
}

func TestNewServerWithWhitelistSourceRange(t *testing.T) {
	cases := []struct {
		desc                 string
		whitelistStrings     []string
		middlewareConfigured bool
		errMessage           string
	}{
		{
			desc:                 "no whitelists configued",
			whitelistStrings:     nil,
			middlewareConfigured: false,
			errMessage:           "",
		}, {
			desc: "whitelists configued",
			whitelistStrings: []string{
				"1.2.3.4/24",
				"fe80::/16",
			},
			middlewareConfigured: true,
			errMessage:           "",
		}, {
			desc: "invalid whitelists configued",
			whitelistStrings: []string{
				"foo",
			},
			middlewareConfigured: false,
			errMessage:           "parsing CIDR whitelist <nil>: invalid CIDR address: foo",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			middleware, err := configureIPWhitelistMiddleware(tc.whitelistStrings)

			if tc.errMessage != "" {
				require.EqualError(t, err, tc.errMessage)
			} else {
				assert.NoError(t, err)

				if tc.middlewareConfigured {
					require.NotNil(t, middleware, "not expected middleware to be configured")
				} else {
					require.Nil(t, middleware, "expected middleware to be configured")
				}
			}
		})
	}
}

func TestServerLoadConfigEmptyBasicAuth(t *testing.T) {
	globalConfig := configuration.GlobalConfiguration{
		EntryPoints: configuration.EntryPoints{
			"http": &configuration.EntryPoint{},
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

	srv := NewServer(globalConfig)
	if _, err := srv.loadConfig(dynamicConfigs, globalConfig); err != nil {
		t.Fatalf("got error: %s", err)
	}
}

func TestConfigureBackends(t *testing.T) {
	validMethod := "Drr"
	defaultMethod := "wrr"

	tests := []struct {
		desc       string
		lb         *types.LoadBalancer
		wantMethod string
		wantSticky bool
	}{
		{
			desc: "valid load balancer method with sticky enabled",
			lb: &types.LoadBalancer{
				Method: validMethod,
				Sticky: true,
			},
			wantMethod: validMethod,
			wantSticky: true,
		},
		{
			desc: "valid load balancer method with sticky disabled",
			lb: &types.LoadBalancer{
				Method: validMethod,
				Sticky: false,
			},
			wantMethod: validMethod,
			wantSticky: false,
		},
		{
			desc: "invalid load balancer method with sticky enabled",
			lb: &types.LoadBalancer{
				Method: "Invalid",
				Sticky: true,
			},
			wantMethod: defaultMethod,
			wantSticky: true,
		},
		{
			desc: "invalid load balancer method with sticky disabled",
			lb: &types.LoadBalancer{
				Method: "Invalid",
				Sticky: false,
			},
			wantMethod: defaultMethod,
			wantSticky: false,
		},
		{
			desc:       "missing load balancer",
			lb:         nil,
			wantMethod: defaultMethod,
			wantSticky: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			backend := &types.Backend{
				LoadBalancer: test.lb,
			}

			srv := Server{}
			srv.configureBackends(map[string]*types.Backend{
				"backend": backend,
			})

			wantLB := types.LoadBalancer{
				Method: test.wantMethod,
				Sticky: test.wantSticky,
			}
			if !reflect.DeepEqual(*backend.LoadBalancer, wantLB) {
				t.Errorf("got backend load-balancer\n%v\nwant\n%v\n", spew.Sdump(backend.LoadBalancer), spew.Sdump(wantLB))
			}
		})
	}
}

func TestServerEntrypointWhitelistConfig(t *testing.T) {
	tests := []struct {
		desc           string
		entrypoint     *configuration.EntryPoint
		wantMiddleware bool
	}{
		{
			desc: "no whitelist middleware if no config on entrypoint",
			entrypoint: &configuration.EntryPoint{
				Address: ":8080",
			},
			wantMiddleware: false,
		},
		{
			desc: "whitelist middleware should be added if configured on entrypoint",
			entrypoint: &configuration.EntryPoint{
				Address: ":8080",
				WhitelistSourceRange: []string{
					"127.0.0.1/32",
				},
			},
			wantMiddleware: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			srv := Server{
				globalConfiguration: configuration.GlobalConfiguration{
					EntryPoints: map[string]*configuration.EntryPoint{
						"test": test.entrypoint,
					},
				},
				metricsRegistry: metrics.NewVoidRegistry(),
			}

			srv.serverEntryPoints = srv.buildEntryPoints(srv.globalConfiguration)
			srvEntryPoint := srv.setupServerEntryPoint("test", srv.serverEntryPoints["test"])
			handler := srvEntryPoint.httpServer.Handler.(*negroni.Negroni)
			found := false
			for _, handler := range handler.Handlers() {
				if reflect.TypeOf(handler) == reflect.TypeOf((*middlewares.IPWhitelister)(nil)) {
					found = true
				}
			}

			if found && !test.wantMiddleware {
				t.Errorf("ip whitelist middleware was installed even though it should not")
			}

			if !found && test.wantMiddleware {
				t.Errorf("ip whitelist middleware was not installed even though it should have")
			}
		})
	}
}

func TestServerResponseEmptyBackend(t *testing.T) {
	const requestPath = "/path"
	const routeRule = "Path:" + requestPath

	testCases := []struct {
		desc           string
		dynamicConfig  func(testServerURL string) *types.Configuration
		wantStatusCode int
	}{
		{
			desc: "Ok",
			dynamicConfig: func(testServerURL string) *types.Configuration {
				return buildDynamicConfig(
					withFrontend("frontend", buildFrontend(withRoute(requestPath, routeRule))),
					withBackend("backend", buildBackend(withServer("testServer", testServerURL))),
				)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			desc: "No Frontend",
			dynamicConfig: func(testServerURL string) *types.Configuration {
				return buildDynamicConfig()
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			desc: "Empty Backend LB-Drr",
			dynamicConfig: func(testServerURL string) *types.Configuration {
				return buildDynamicConfig(
					withFrontend("frontend", buildFrontend(withRoute(requestPath, routeRule))),
					withBackend("backend", buildBackend(withLoadBalancer("Drr", false))),
				)
			},
			wantStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Drr Sticky",
			dynamicConfig: func(testServerURL string) *types.Configuration {
				return buildDynamicConfig(
					withFrontend("frontend", buildFrontend(withRoute(requestPath, routeRule))),
					withBackend("backend", buildBackend(withLoadBalancer("Drr", true))),
				)
			},
			wantStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Wrr",
			dynamicConfig: func(testServerURL string) *types.Configuration {
				return buildDynamicConfig(
					withFrontend("frontend", buildFrontend(withRoute(requestPath, routeRule))),
					withBackend("backend", buildBackend(withLoadBalancer("Wrr", false))),
				)
			},
			wantStatusCode: http.StatusServiceUnavailable,
		},
		{
			desc: "Empty Backend LB-Wrr Sticky",
			dynamicConfig: func(testServerURL string) *types.Configuration {
				return buildDynamicConfig(
					withFrontend("frontend", buildFrontend(withRoute(requestPath, routeRule))),
					withBackend("backend", buildBackend(withLoadBalancer("Wrr", true))),
				)
			},
			wantStatusCode: http.StatusServiceUnavailable,
		},
	}

	for _, test := range testCases {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			testServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
			}))
			defer testServer.Close()

			globalConfig := configuration.GlobalConfiguration{
				EntryPoints: configuration.EntryPoints{
					"http": &configuration.EntryPoint{},
				},
			}
			dynamicConfigs := types.Configurations{"config": test.dynamicConfig(testServer.URL)}

			srv := NewServer(globalConfig)
			entryPoints, err := srv.loadConfig(dynamicConfigs, globalConfig)
			if err != nil {
				t.Fatalf("error loading config: %s", err)
			}

			responseRecorder := &httptest.ResponseRecorder{}
			request := httptest.NewRequest(http.MethodGet, testServer.URL+requestPath, nil)

			entryPoints["http"].httpRouter.ServeHTTP(responseRecorder, request)

			if responseRecorder.Result().StatusCode != test.wantStatusCode {
				t.Errorf("got status code %d, want %d", responseRecorder.Result().StatusCode, test.wantStatusCode)
			}
		})
	}
}

func buildDynamicConfig(dynamicConfigBuilders ...func(*types.Configuration)) *types.Configuration {
	config := &types.Configuration{
		Frontends: make(map[string]*types.Frontend),
		Backends:  make(map[string]*types.Backend),
	}
	for _, build := range dynamicConfigBuilders {
		build(config)
	}
	return config
}

func withFrontend(frontendName string, frontend *types.Frontend) func(*types.Configuration) {
	return func(config *types.Configuration) {
		config.Frontends[frontendName] = frontend
	}
}

func withBackend(backendName string, backend *types.Backend) func(*types.Configuration) {
	return func(config *types.Configuration) {
		config.Backends[backendName] = backend
	}
}

func buildFrontend(frontendBuilders ...func(*types.Frontend)) *types.Frontend {
	fe := &types.Frontend{
		EntryPoints: []string{"http"},
		Backend:     "backend",
		Routes:      make(map[string]types.Route),
	}
	for _, build := range frontendBuilders {
		build(fe)
	}
	return fe
}

func withRoute(routeName, rule string) func(*types.Frontend) {
	return func(fe *types.Frontend) {
		fe.Routes[routeName] = types.Route{Rule: rule}
	}
}

func buildBackend(backendBuilders ...func(*types.Backend)) *types.Backend {
	be := &types.Backend{
		Servers:      make(map[string]types.Server),
		LoadBalancer: &types.LoadBalancer{Method: "Wrr"},
	}
	for _, build := range backendBuilders {
		build(be)
	}
	return be
}

func withServer(name, url string) func(backend *types.Backend) {
	return func(be *types.Backend) {
		be.Servers[name] = types.Server{URL: url}
	}
}

func withLoadBalancer(method string, sticky bool) func(*types.Backend) {
	return func(be *types.Backend) {
		be.LoadBalancer = &types.LoadBalancer{Method: method, Sticky: sticky}
	}
}
