package server

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/containous/flaeg"
	"github.com/containous/mux"
	"github.com/containous/traefik/healthcheck"
	"github.com/containous/traefik/testhelpers"
	"github.com/containous/traefik/types"
	"github.com/davecgh/go-spew/spew"
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
				globalConfig := GlobalConfiguration{
					EntryPoints: EntryPoints{
						"http": &EntryPoint{},
					},
					HealthCheck: &HealthCheckConfig{Interval: flaeg.Duration(5 * time.Second)},
				}

				dynamicConfigs := configs{
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

			gotOpts := parseHealthCheckOptions(lb, "backend", test.hc, &HealthCheckConfig{Interval: flaeg.Duration(globalInterval)})
			if !reflect.DeepEqual(gotOpts, test.wantOpts) {
				t.Errorf("got health check options %+v, want %+v", gotOpts, test.wantOpts)
			}
		})
	}
}

func TestServerLoadConfigEmptyBasicAuth(t *testing.T) {
	globalConfig := GlobalConfiguration{
		EntryPoints: EntryPoints{
			"http": &EntryPoint{},
		},
	}

	dynamicConfigs := configs{
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
