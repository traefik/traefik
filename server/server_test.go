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
	"github.com/containous/traefik/healthcheck"
	"github.com/containous/traefik/types"
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

			gotOpts := parseHealthCheckOptions(lb, "backend", test.hc, HealthCheckConfig{Interval: flaeg.Duration(globalInterval)})
			if !reflect.DeepEqual(gotOpts, test.wantOpts) {
				t.Errorf("got health check options %+v, want %+v", gotOpts, test.wantOpts)
			}
		})
	}
}

func TestBuildDefaultHTTPRouterSetupNotFoundHandler(t *testing.T) {
	tests := []struct {
		name                   string
		globalConfig           GlobalConfiguration
		wantNotFoundStatusCode int
	}{
		{
			name:                   "default",
			globalConfig:           GlobalConfiguration{},
			wantNotFoundStatusCode: http.StatusNotFound,
		},
		{
			name:                   "sane config",
			globalConfig:           GlobalConfiguration{NoRouteResponseCode: http.StatusBadGateway},
			wantNotFoundStatusCode: http.StatusBadGateway,
		},
		{
			name:                   "invalid config",
			globalConfig:           GlobalConfiguration{NoRouteResponseCode: 1000},
			wantNotFoundStatusCode: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "http://localhost/notknown", nil)

			router := buildDefaultHTTPRouter(test.globalConfig)
			router.ServeHTTP(recorder, req)

			if recorder.Code != test.wantNotFoundStatusCode {
				t.Errorf("got wrong status status code %v, want %v", recorder.Code, test.wantNotFoundStatusCode)
			}
		})
	}

}

func TestNewNotFounderHandler(t *testing.T) {
	tests := []struct {
		name                string
		noRouteResponseCode int
		wantStatusCode      int
		wantBody            string
	}{
		{
			name:                "not found",
			noRouteResponseCode: http.StatusNotFound,
			wantStatusCode:      http.StatusNotFound,
			wantBody:            "Not Found",
		},
		{
			name:                "bad gateway",
			noRouteResponseCode: http.StatusBadGateway,
			wantStatusCode:      http.StatusBadGateway,
			wantBody:            "Bad Gateway",
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			notFoundHandler := newNotFounderHandler(test.noRouteResponseCode)
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "http://localhost/notknown", nil)

			notFoundHandler.ServeHTTP(recorder, req)

			if recorder.Code != test.wantStatusCode {
				t.Errorf("got wrong status status code %v, want %v", recorder.Code, test.wantStatusCode)
			}
			if recorder.Body.String() != test.wantBody {
				t.Errorf("got wrong body %v, want %v", recorder.Body.String(), test.wantBody)
			}
		})
	}
}
