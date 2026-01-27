package healthcheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ptypes "github.com/traefik/paerser/types"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const delta float64 = 1e-10

func pointer[T any](v T) *T { return &v }

func TestNewServiceHealthChecker_durations(t *testing.T) {
	testCases := []struct {
		desc        string
		config      *dynamic.ServerHealthCheck
		expInterval time.Duration
		expTimeout  time.Duration
	}{
		{
			desc:        "default values",
			config:      &dynamic.ServerHealthCheck{},
			expInterval: time.Duration(dynamic.DefaultHealthCheckInterval),
			expTimeout:  time.Duration(dynamic.DefaultHealthCheckTimeout),
		},
		{
			desc: "out of range values",
			config: &dynamic.ServerHealthCheck{
				Interval: ptypes.Duration(-time.Second),
				Timeout:  ptypes.Duration(-time.Second),
			},
			expInterval: time.Duration(dynamic.DefaultHealthCheckInterval),
			expTimeout:  time.Duration(dynamic.DefaultHealthCheckTimeout),
		},
		{
			desc: "custom durations",
			config: &dynamic.ServerHealthCheck{
				Interval: ptypes.Duration(time.Second * 10),
				Timeout:  ptypes.Duration(time.Second * 5),
			},
			expInterval: time.Second * 10,
			expTimeout:  time.Second * 5,
		},
		{
			desc: "interval shorter than timeout",
			config: &dynamic.ServerHealthCheck{
				Interval: ptypes.Duration(time.Second),
				Timeout:  ptypes.Duration(time.Second * 5),
			},
			expInterval: time.Second,
			expTimeout:  time.Second * 5,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			healthChecker := NewServiceHealthChecker(t.Context(), nil, test.config, nil, nil, http.DefaultTransport, nil, "")
			assert.Equal(t, test.expInterval, healthChecker.interval)
			assert.Equal(t, test.expTimeout, healthChecker.timeout)
		})
	}
}

func TestServiceHealthChecker_newRequest(t *testing.T) {
	testCases := []struct {
		desc        string
		targetURL   string
		config      dynamic.ServerHealthCheck
		expTarget   string
		expError    bool
		expHostname string
		expHeader   string
		expMethod   string
	}{
		{
			desc:      "no port override",
			targetURL: "http://backend1:80",
			config: dynamic.ServerHealthCheck{
				Path: "/test",
				Port: 0,
			},
			expError:    false,
			expTarget:   "http://backend1:80/test",
			expHostname: "backend1:80",
			expMethod:   http.MethodGet,
		},
		{
			desc:      "port override",
			targetURL: "http://backend2:80",
			config: dynamic.ServerHealthCheck{
				Path: "/test",
				Port: 8080,
			},
			expError:    false,
			expTarget:   "http://backend2:8080/test",
			expHostname: "backend2:8080",
			expMethod:   http.MethodGet,
		},
		{
			desc:      "no port override with no port in server URL",
			targetURL: "http://backend1",
			config: dynamic.ServerHealthCheck{
				Path: "/health",
				Port: 0,
			},
			expError:    false,
			expTarget:   "http://backend1/health",
			expHostname: "backend1",
			expMethod:   http.MethodGet,
		},
		{
			desc:      "port override with no port in server URL",
			targetURL: "http://backend2",
			config: dynamic.ServerHealthCheck{
				Path: "/health",
				Port: 8080,
			},
			expError:    false,
			expTarget:   "http://backend2:8080/health",
			expHostname: "backend2:8080",
			expMethod:   http.MethodGet,
		},
		{
			desc:      "scheme override",
			targetURL: "https://backend1:80",
			config: dynamic.ServerHealthCheck{
				Scheme: "http",
				Path:   "/test",
				Port:   0,
			},
			expError:    false,
			expTarget:   "http://backend1:80/test",
			expHostname: "backend1:80",
			expMethod:   http.MethodGet,
		},
		{
			desc:      "path with param",
			targetURL: "http://backend1:80",
			config: dynamic.ServerHealthCheck{
				Path: "/health?powpow=do",
				Port: 0,
			},
			expError:    false,
			expTarget:   "http://backend1:80/health?powpow=do",
			expHostname: "backend1:80",
			expMethod:   http.MethodGet,
		},
		{
			desc:      "path with params",
			targetURL: "http://backend1:80",
			config: dynamic.ServerHealthCheck{
				Path: "/health?powpow=do&do=powpow",
				Port: 0,
			},
			expError:    false,
			expTarget:   "http://backend1:80/health?powpow=do&do=powpow",
			expHostname: "backend1:80",
			expMethod:   http.MethodGet,
		},
		{
			desc:      "path with invalid path",
			targetURL: "http://backend1:80",
			config: dynamic.ServerHealthCheck{
				Path: ":",
				Port: 0,
			},
			expError:    true,
			expTarget:   "",
			expHostname: "backend1:80",
			expMethod:   http.MethodGet,
		},
		{
			desc:      "override hostname",
			targetURL: "http://backend1:80",
			config: dynamic.ServerHealthCheck{
				Hostname: "myhost",
				Path:     "/",
			},
			expTarget:   "http://backend1:80/",
			expHostname: "myhost",
			expHeader:   "",
			expMethod:   http.MethodGet,
		},
		{
			desc:      "not override hostname",
			targetURL: "http://backend1:80",
			config: dynamic.ServerHealthCheck{
				Hostname: "",
				Path:     "/",
			},
			expTarget:   "http://backend1:80/",
			expHostname: "backend1:80",
			expHeader:   "",
			expMethod:   http.MethodGet,
		},
		{
			desc:      "custom header",
			targetURL: "http://backend1:80",
			config: dynamic.ServerHealthCheck{
				Headers:  map[string]string{"Custom-Header": "foo"},
				Hostname: "",
				Path:     "/",
			},
			expTarget:   "http://backend1:80/",
			expHostname: "backend1:80",
			expHeader:   "foo",
			expMethod:   http.MethodGet,
		},
		{
			desc:      "custom header with hostname override",
			targetURL: "http://backend1:80",
			config: dynamic.ServerHealthCheck{
				Headers:  map[string]string{"Custom-Header": "foo"},
				Hostname: "myhost",
				Path:     "/",
			},
			expTarget:   "http://backend1:80/",
			expHostname: "myhost",
			expHeader:   "foo",
			expMethod:   http.MethodGet,
		},
		{
			desc:      "custom method",
			targetURL: "http://backend1:80",
			config: dynamic.ServerHealthCheck{
				Path:   "/",
				Method: http.MethodHead,
			},
			expTarget:   "http://backend1:80/",
			expHostname: "backend1:80",
			expMethod:   http.MethodHead,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			shc := ServiceHealthChecker{config: &test.config}

			u := testhelpers.MustParseURL(test.targetURL)
			req, err := shc.newRequest(t.Context(), u)

			if test.expError {
				require.Error(t, err)
				assert.Nil(t, req)
			} else {
				require.NoError(t, err, "failed to create new request")
				require.NotNil(t, req)

				assert.Equal(t, test.expTarget, req.URL.String())
				assert.Equal(t, test.expHeader, req.Header.Get("Custom-Header"))
				assert.Equal(t, test.expHostname, req.Host)
				assert.Equal(t, test.expMethod, req.Method)
			}
		})
	}
}

func TestServiceHealthChecker_checkHealthHTTP_NotFollowingRedirects(t *testing.T) {
	redirectServerCalled := false
	redirectTestServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		redirectServerCalled = true
	}))
	defer redirectTestServer.Close()

	ctx, cancel := context.WithTimeout(t.Context(), time.Duration(dynamic.DefaultHealthCheckTimeout))
	defer cancel()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("location", redirectTestServer.URL)
		rw.WriteHeader(http.StatusSeeOther)
	}))
	defer server.Close()

	config := &dynamic.ServerHealthCheck{
		Path:            "/path",
		FollowRedirects: pointer(false),
		Interval:        dynamic.DefaultHealthCheckInterval,
		Timeout:         dynamic.DefaultHealthCheckTimeout,
	}
	healthChecker := NewServiceHealthChecker(ctx, nil, config, nil, nil, http.DefaultTransport, nil, "")

	err := healthChecker.checkHealthHTTP(ctx, testhelpers.MustParseURL(server.URL))
	require.NoError(t, err)

	assert.False(t, redirectServerCalled, "HTTP redirect must not be followed")
}

func TestServiceHealthChecker_Launch(t *testing.T) {
	testCases := []struct {
		desc                  string
		mode                  string
		status                int
		server                StartTestServer
		expNumRemovedServers  int
		expNumUpsertedServers int
		expGaugeValue         float64
		targetStatus          string
	}{
		{
			desc:                  "healthy server staying healthy",
			server:                newHTTPServer(http.StatusOK),
			expNumRemovedServers:  0,
			expNumUpsertedServers: 1,
			expGaugeValue:         1,
			targetStatus:          runtime.StatusUp,
		},
		{
			desc:                  "healthy server staying healthy, with custom code status check",
			server:                newHTTPServer(http.StatusNotFound),
			status:                http.StatusNotFound,
			expNumRemovedServers:  0,
			expNumUpsertedServers: 1,
			expGaugeValue:         1,
			targetStatus:          runtime.StatusUp,
		},
		{
			desc:                  "healthy server staying healthy (StatusNoContent)",
			server:                newHTTPServer(http.StatusNoContent),
			expNumRemovedServers:  0,
			expNumUpsertedServers: 1,
			expGaugeValue:         1,
			targetStatus:          runtime.StatusUp,
		},
		{
			desc:                  "healthy server staying healthy (StatusPermanentRedirect)",
			server:                newHTTPServer(http.StatusPermanentRedirect),
			expNumRemovedServers:  0,
			expNumUpsertedServers: 1,
			expGaugeValue:         1,
			targetStatus:          runtime.StatusUp,
		},
		{
			desc:                  "healthy server becoming sick",
			server:                newHTTPServer(http.StatusServiceUnavailable),
			expNumRemovedServers:  1,
			expNumUpsertedServers: 0,
			expGaugeValue:         0,
			targetStatus:          runtime.StatusDown,
		},
		{
			desc:                  "healthy server becoming sick, with custom code status check",
			server:                newHTTPServer(http.StatusOK),
			status:                http.StatusServiceUnavailable,
			expNumRemovedServers:  1,
			expNumUpsertedServers: 0,
			expGaugeValue:         0,
			targetStatus:          runtime.StatusDown,
		},
		{
			desc:                  "healthy server toggling to sick and back to healthy",
			server:                newHTTPServer(http.StatusServiceUnavailable, http.StatusOK),
			expNumRemovedServers:  1,
			expNumUpsertedServers: 1,
			expGaugeValue:         1,
			targetStatus:          runtime.StatusUp,
		},
		{
			desc:                  "healthy server toggling to healthy and go to sick",
			server:                newHTTPServer(http.StatusOK, http.StatusServiceUnavailable),
			expNumRemovedServers:  1,
			expNumUpsertedServers: 1,
			expGaugeValue:         0,
			targetStatus:          runtime.StatusDown,
		},
		{
			desc:                  "healthy grpc server staying healthy",
			mode:                  "grpc",
			server:                newGRPCServer(healthpb.HealthCheckResponse_SERVING),
			expNumRemovedServers:  0,
			expNumUpsertedServers: 1,
			expGaugeValue:         1,
			targetStatus:          runtime.StatusUp,
		},
		{
			desc:                  "healthy grpc server becoming sick",
			mode:                  "grpc",
			server:                newGRPCServer(healthpb.HealthCheckResponse_NOT_SERVING),
			expNumRemovedServers:  1,
			expNumUpsertedServers: 0,
			expGaugeValue:         0,
			targetStatus:          runtime.StatusDown,
		},
		{
			desc:                  "healthy grpc server toggling to sick and back to healthy",
			mode:                  "grpc",
			server:                newGRPCServer(healthpb.HealthCheckResponse_NOT_SERVING, healthpb.HealthCheckResponse_SERVING),
			expNumRemovedServers:  1,
			expNumUpsertedServers: 1,
			expGaugeValue:         1,
			targetStatus:          runtime.StatusUp,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			// The context is passed to the health check and
			// canonically canceled by the test server once all expected requests have been received.
			ctx, cancel := context.WithCancel(t.Context())
			t.Cleanup(cancel)

			targetURL, timeout := test.server.Start(t, cancel)

			lb := &testLoadBalancer{RWMutex: &sync.RWMutex{}}

			config := &dynamic.ServerHealthCheck{
				Mode:              test.mode,
				Status:            test.status,
				Path:              "/path",
				Interval:          ptypes.Duration(500 * time.Millisecond),
				UnhealthyInterval: pointer(ptypes.Duration(500 * time.Millisecond)),
				Timeout:           ptypes.Duration(499 * time.Millisecond),
			}

			gauge := &testhelpers.CollectingGauge{}
			serviceInfo := &runtime.ServiceInfo{}
			hc := NewServiceHealthChecker(ctx, &MetricsMock{gauge}, config, lb, serviceInfo, http.DefaultTransport, map[string]*url.URL{"test": targetURL}, "foobar")

			wg := sync.WaitGroup{}
			wg.Add(1)

			go func() {
				hc.Launch(ctx)
				wg.Done()
			}()

			select {
			case <-time.After(timeout):
				t.Fatal("test did not complete in time")
			case <-ctx.Done():
				wg.Wait()
			}

			lb.Lock()
			defer lb.Unlock()

			assert.Equal(t, test.expNumRemovedServers, lb.numRemovedServers, "removed servers")
			assert.Equal(t, test.expNumUpsertedServers, lb.numUpsertedServers, "upserted servers")
			assert.InDelta(t, test.expGaugeValue, gauge.GaugeValue, delta, "ServerUp Gauge")
			assert.Equal(t, []string{"service", "foobar", "url", targetURL.String()}, gauge.LastLabelValues)
			assert.Equal(t, map[string]string{targetURL.String(): test.targetStatus}, serviceInfo.GetAllStatus())
		})
	}
}

func TestDifferentIntervals(t *testing.T) {
	// The context is passed to the health check and
	// canonically canceled by the test server once all expected requests have been received.
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	healthyURL := testhelpers.MustParseURL(healthyServer.URL)

	unhealthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	unhealthyURL := testhelpers.MustParseURL(unhealthyServer.URL)

	lb := &testLoadBalancer{RWMutex: &sync.RWMutex{}}

	config := &dynamic.ServerHealthCheck{
		Mode:              "http",
		Path:              "/path",
		Interval:          ptypes.Duration(500 * time.Millisecond),
		UnhealthyInterval: pointer(ptypes.Duration(50 * time.Millisecond)),
		Timeout:           ptypes.Duration(499 * time.Millisecond),
	}

	gauge := &testhelpers.CollectingGauge{}
	serviceInfo := &runtime.ServiceInfo{}
	hc := NewServiceHealthChecker(ctx, &MetricsMock{gauge}, config, lb, serviceInfo, http.DefaultTransport, map[string]*url.URL{"healthy": healthyURL, "unhealthy": unhealthyURL}, "foobar")

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		hc.Launch(ctx)
		wg.Done()
	}()

	select {
	case <-time.After(2 * time.Second):
		break
	case <-ctx.Done():
		wg.Wait()
	}

	lb.Lock()
	defer lb.Unlock()

	assert.Greater(t, lb.numRemovedServers, lb.numUpsertedServers, "removed servers greater than upserted servers")
}

func TestServiceHealthChecker_FailsThreshold(t *testing.T) {
	// Test that with failsThreshold=3, server stays healthy until 3 consecutive failures.
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	// Create a counter to track request sequence
	requestCount := 0
	requestCountMu := sync.Mutex{}

	// Server returns: OK, FAIL, OK, FAIL, FAIL, FAIL (should go down after 3rd consecutive fail)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestCountMu.Lock()
		count := requestCount
		requestCount++
		requestCountMu.Unlock()

		switch count {
		case 0: // OK
			w.WriteHeader(http.StatusOK)
		case 1: // FAIL (1st fail, resets on next OK)
			w.WriteHeader(http.StatusServiceUnavailable)
		case 2: // OK (resets fail counter)
			w.WriteHeader(http.StatusOK)
		case 3, 4, 5: // FAIL x3 (consecutive, should trigger down)
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			cancel()
			time.Sleep(100 * time.Millisecond)
		}
	}))
	t.Cleanup(server.Close)

	serverURL := testhelpers.MustParseURL(server.URL)

	lb := &testLoadBalancer{RWMutex: &sync.RWMutex{}}

	config := &dynamic.ServerHealthCheck{
		Mode:            "http",
		Path:            "/path",
		Interval:        ptypes.Duration(100 * time.Millisecond),
		Timeout:         ptypes.Duration(99 * time.Millisecond),
		FailsThreshold:  3,
		PassesThreshold: 1,
	}

	gauge := &testhelpers.CollectingGauge{}
	serviceInfo := &runtime.ServiceInfo{}
	hc := NewServiceHealthChecker(ctx, &MetricsMock{gauge}, config, lb, serviceInfo, http.DefaultTransport, map[string]*url.URL{"test": serverURL}, "foobar")

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		hc.Launch(ctx)
		wg.Done()
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("test did not complete in time")
	case <-ctx.Done():
		wg.Wait()
	}

	lb.Lock()
	defer lb.Unlock()

	// With failsThreshold=3:
	// - Request 0 (OK): healthy, upserted
	// - Request 1 (FAIL): still healthy (only 1 fail), upserted
	// - Request 2 (OK): healthy (fail counter reset), upserted
	// - Request 3 (FAIL): still healthy (only 1 fail), upserted
	// - Request 4 (FAIL): still healthy (only 2 fails), upserted
	// - Request 5 (FAIL): NOW unhealthy (3 fails), removed
	// So: 5 upserts, 1 remove
	assert.Equal(t, 1, lb.numRemovedServers, "removed servers")
	assert.Equal(t, 5, lb.numUpsertedServers, "upserted servers")
}

func TestServiceHealthChecker_PassesThreshold(t *testing.T) {
	// Test that with passesThreshold=2, unhealthy server needs 2 consecutive successes to recover.
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	requestCount := 0
	requestCountMu := sync.Mutex{}

	// Server returns: FAIL, OK, FAIL, OK, OK (should recover after 2nd consecutive OK)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requestCountMu.Lock()
		count := requestCount
		requestCount++
		requestCountMu.Unlock()

		switch count {
		case 0: // FAIL (goes down immediately with failsThreshold=1)
			w.WriteHeader(http.StatusServiceUnavailable)
		case 1: // OK (1st pass, not enough)
			w.WriteHeader(http.StatusOK)
		case 2: // FAIL (resets pass counter)
			w.WriteHeader(http.StatusServiceUnavailable)
		case 3, 4: // OK x2 (should recover)
			w.WriteHeader(http.StatusOK)
		default:
			cancel()
			time.Sleep(100 * time.Millisecond)
		}
	}))
	t.Cleanup(server.Close)

	serverURL := testhelpers.MustParseURL(server.URL)

	lb := &testLoadBalancer{RWMutex: &sync.RWMutex{}}

	config := &dynamic.ServerHealthCheck{
		Mode:              "http",
		Path:              "/path",
		Interval:          ptypes.Duration(100 * time.Millisecond),
		UnhealthyInterval: pointer(ptypes.Duration(100 * time.Millisecond)),
		Timeout:           ptypes.Duration(99 * time.Millisecond),
		FailsThreshold:    1,
		PassesThreshold:   2,
	}

	gauge := &testhelpers.CollectingGauge{}
	serviceInfo := &runtime.ServiceInfo{}
	hc := NewServiceHealthChecker(ctx, &MetricsMock{gauge}, config, lb, serviceInfo, http.DefaultTransport, map[string]*url.URL{"test": serverURL}, "foobar")

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		hc.Launch(ctx)
		wg.Done()
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("test did not complete in time")
	case <-ctx.Done():
		wg.Wait()
	}

	lb.Lock()
	defer lb.Unlock()

	// With passesThreshold=2, failsThreshold=1:
	// - Request 0 (FAIL): unhealthy, removed
	// - Request 1 (OK): still unhealthy (only 1 pass), removed
	// - Request 2 (FAIL): still unhealthy (pass counter reset), removed
	// - Request 3 (OK): still unhealthy (only 1 pass), removed
	// - Request 4 (OK): NOW healthy (2 passes), upserted
	// So: 4 removes, 1 upsert
	assert.Equal(t, 4, lb.numRemovedServers, "removed servers")
	assert.Equal(t, 1, lb.numUpsertedServers, "upserted servers")
}
