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
	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
	"github.com/vulcand/oxy/v2/roundrobin"
)

const (
	healthCheckInterval = 200 * time.Millisecond
	healthCheckTimeout  = 100 * time.Millisecond
)

type testHandler struct {
	done           func()
	healthSequence []int
}

func TestSetBackendsConfiguration(t *testing.T) {
	testCases := []struct {
		desc                       string
		startHealthy               bool
		healthSequence             []int
		expectedNumRemovedServers  int
		expectedNumUpsertedServers int
		expectedGaugeValue         float64
	}{
		{
			desc:                       "healthy server staying healthy",
			startHealthy:               true,
			healthSequence:             []int{http.StatusOK},
			expectedNumRemovedServers:  0,
			expectedNumUpsertedServers: 0,
			expectedGaugeValue:         1,
		},
		{
			desc:                       "healthy server staying healthy (StatusNoContent)",
			startHealthy:               true,
			healthSequence:             []int{http.StatusNoContent},
			expectedNumRemovedServers:  0,
			expectedNumUpsertedServers: 0,
			expectedGaugeValue:         1,
		},
		{
			desc:                       "healthy server staying healthy (StatusPermanentRedirect)",
			startHealthy:               true,
			healthSequence:             []int{http.StatusPermanentRedirect},
			expectedNumRemovedServers:  0,
			expectedNumUpsertedServers: 0,
			expectedGaugeValue:         1,
		},
		{
			desc:                       "healthy server becoming sick",
			startHealthy:               true,
			healthSequence:             []int{http.StatusServiceUnavailable},
			expectedNumRemovedServers:  1,
			expectedNumUpsertedServers: 0,
			expectedGaugeValue:         0,
		},
		{
			desc:                       "sick server becoming healthy",
			startHealthy:               false,
			healthSequence:             []int{http.StatusOK},
			expectedNumRemovedServers:  0,
			expectedNumUpsertedServers: 1,
			expectedGaugeValue:         1,
		},
		{
			desc:                       "sick server staying sick",
			startHealthy:               false,
			healthSequence:             []int{http.StatusServiceUnavailable},
			expectedNumRemovedServers:  0,
			expectedNumUpsertedServers: 0,
			expectedGaugeValue:         0,
		},
		{
			desc:                       "healthy server toggling to sick and back to healthy",
			startHealthy:               true,
			healthSequence:             []int{http.StatusServiceUnavailable, http.StatusOK},
			expectedNumRemovedServers:  1,
			expectedNumUpsertedServers: 1,
			expectedGaugeValue:         1,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			// The context is passed to the health check and canonically canceled by
			// the test server once all expected requests have been received.
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ts := newTestServer(cancel, test.healthSequence)
			defer ts.Close()

			lb := &testLoadBalancer{RWMutex: &sync.RWMutex{}}
			backend := NewBackendConfig(Options{
				Path:     "/path",
				Interval: healthCheckInterval,
				Timeout:  healthCheckTimeout,
				LB:       lb,
			}, "backendName")

			serverURL := testhelpers.MustParseURL(ts.URL)
			if test.startHealthy {
				lb.servers = append(lb.servers, serverURL)
			} else {
				backend.disabledURLs = append(backend.disabledURLs, backendURL{url: serverURL, weight: 1})
			}

			collectingMetrics := &testhelpers.CollectingGauge{}
			check := HealthCheck{
				Backends: make(map[string]*BackendConfig),
				metrics:  metricsHealthcheck{serverUpGauge: collectingMetrics},
			}

			wg := sync.WaitGroup{}
			wg.Add(1)

			go func() {
				check.execute(ctx, backend)
				wg.Done()
			}()

			// Make test timeout dependent on number of expected requests, health
			// check interval, and a safety margin.
			timeout := time.Duration(len(test.healthSequence)*int(healthCheckInterval) + 500)
			select {
			case <-time.After(timeout):
				t.Fatal("test did not complete in time")
			case <-ctx.Done():
				wg.Wait()
			}

			lb.Lock()
			defer lb.Unlock()

			assert.Equal(t, test.expectedNumRemovedServers, lb.numRemovedServers, "removed servers")
			assert.Equal(t, test.expectedNumUpsertedServers, lb.numUpsertedServers, "upserted servers")
			assert.Equal(t, test.expectedGaugeValue, collectingMetrics.GaugeValue, "ServerUp Gauge")
		})
	}
}

func TestNewRequest(t *testing.T) {
	type expected struct {
		err   bool
		value string
	}

	testCases := []struct {
		desc      string
		serverURL string
		options   Options
		expected  expected
	}{
		{
			desc:      "no port override",
			serverURL: "http://backend1:80",
			options: Options{
				Path: "/test",
				Port: 0,
			},
			expected: expected{
				err:   false,
				value: "http://backend1:80/test",
			},
		},
		{
			desc:      "port override",
			serverURL: "http://backend2:80",
			options: Options{
				Path: "/test",
				Port: 8080,
			},
			expected: expected{
				err:   false,
				value: "http://backend2:8080/test",
			},
		},
		{
			desc:      "no port override with no port in server URL",
			serverURL: "http://backend1",
			options: Options{
				Path: "/health",
				Port: 0,
			},
			expected: expected{
				err:   false,
				value: "http://backend1/health",
			},
		},
		{
			desc:      "port override with no port in server URL",
			serverURL: "http://backend2",
			options: Options{
				Path: "/health",
				Port: 8080,
			},
			expected: expected{
				err:   false,
				value: "http://backend2:8080/health",
			},
		},
		{
			desc:      "scheme override",
			serverURL: "https://backend1:80",
			options: Options{
				Scheme: "http",
				Path:   "/test",
				Port:   0,
			},
			expected: expected{
				err:   false,
				value: "http://backend1:80/test",
			},
		},
		{
			desc:      "path with param",
			serverURL: "http://backend1:80",
			options: Options{
				Path: "/health?powpow=do",
				Port: 0,
			},
			expected: expected{
				err:   false,
				value: "http://backend1:80/health?powpow=do",
			},
		},
		{
			desc:      "path with params",
			serverURL: "http://backend1:80",
			options: Options{
				Path: "/health?powpow=do&do=powpow",
				Port: 0,
			},
			expected: expected{
				err:   false,
				value: "http://backend1:80/health?powpow=do&do=powpow",
			},
		},
		{
			desc:      "path with invalid path",
			serverURL: "http://backend1:80",
			options: Options{
				Path: ":",
				Port: 0,
			},
			expected: expected{
				err:   true,
				value: "",
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			backend := NewBackendConfig(test.options, "backendName")

			u := testhelpers.MustParseURL(test.serverURL)

			req, err := backend.newRequest(u)

			if test.expected.err {
				require.Error(t, err)
				assert.Nil(t, nil)
			} else {
				require.NoError(t, err, "failed to create new backend request")
				require.NotNil(t, req)
				assert.Equal(t, test.expected.value, req.URL.String())
			}
		})
	}
}

func TestRequestOptions(t *testing.T) {
	testCases := []struct {
		desc             string
		serverURL        string
		options          Options
		expectedHostname string
		expectedHeader   string
		expectedMethod   string
	}{
		{
			desc:      "override hostname",
			serverURL: "http://backend1:80",
			options: Options{
				Hostname: "myhost",
				Path:     "/",
			},
			expectedHostname: "myhost",
			expectedHeader:   "",
			expectedMethod:   http.MethodGet,
		},
		{
			desc:      "not override hostname",
			serverURL: "http://backend1:80",
			options: Options{
				Hostname: "",
				Path:     "/",
			},
			expectedHostname: "backend1:80",
			expectedHeader:   "",
			expectedMethod:   http.MethodGet,
		},
		{
			desc:      "custom header",
			serverURL: "http://backend1:80",
			options: Options{
				Headers:  map[string]string{"Custom-Header": "foo"},
				Hostname: "",
				Path:     "/",
			},
			expectedHostname: "backend1:80",
			expectedHeader:   "foo",
			expectedMethod:   http.MethodGet,
		},
		{
			desc:      "custom header with hostname override",
			serverURL: "http://backend1:80",
			options: Options{
				Headers:  map[string]string{"Custom-Header": "foo"},
				Hostname: "myhost",
				Path:     "/",
			},
			expectedHostname: "myhost",
			expectedHeader:   "foo",
			expectedMethod:   http.MethodGet,
		},
		{
			desc:      "custom method",
			serverURL: "http://backend1:80",
			options: Options{
				Path:   "/",
				Method: http.MethodHead,
			},
			expectedHostname: "backend1:80",
			expectedMethod:   http.MethodHead,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			backend := NewBackendConfig(test.options, "backendName")

			u, err := url.Parse(test.serverURL)
			require.NoError(t, err)

			req, err := backend.newRequest(u)
			require.NoError(t, err, "failed to create new backend request")

			req = backend.setRequestOptions(req)

			assert.Equal(t, "http://backend1:80/", req.URL.String())
			assert.Equal(t, test.expectedHostname, req.Host)
			assert.Equal(t, test.expectedHeader, req.Header.Get("Custom-Header"))
			assert.Equal(t, test.expectedMethod, req.Method)
		})
	}
}

func TestBalancers_Servers(t *testing.T) {
	server1, err := url.Parse("http://foo.com")
	require.NoError(t, err)

	balancer1, err := roundrobin.New(nil)
	require.NoError(t, err)

	err = balancer1.UpsertServer(server1)
	require.NoError(t, err)

	server2, err := url.Parse("http://foo.com")
	require.NoError(t, err)

	balancer2, err := roundrobin.New(nil)
	require.NoError(t, err)

	err = balancer2.UpsertServer(server2)
	require.NoError(t, err)

	balancers := Balancers([]Balancer{balancer1, balancer2})

	want, err := url.Parse("http://foo.com")
	require.NoError(t, err)

	assert.Equal(t, 1, len(balancers.Servers()))
	assert.Equal(t, want, balancers.Servers()[0])
}

func TestBalancers_UpsertServer(t *testing.T) {
	balancer1, err := roundrobin.New(nil)
	require.NoError(t, err)

	balancer2, err := roundrobin.New(nil)
	require.NoError(t, err)

	want, err := url.Parse("http://foo.com")
	require.NoError(t, err)

	balancers := Balancers([]Balancer{balancer1, balancer2})

	err = balancers.UpsertServer(want)
	require.NoError(t, err)

	assert.Equal(t, 1, len(balancer1.Servers()))
	assert.Equal(t, want, balancer1.Servers()[0])

	assert.Equal(t, 1, len(balancer2.Servers()))
	assert.Equal(t, want, balancer2.Servers()[0])
}

func TestBalancers_RemoveServer(t *testing.T) {
	server, err := url.Parse("http://foo.com")
	require.NoError(t, err)

	balancer1, err := roundrobin.New(nil)
	require.NoError(t, err)

	err = balancer1.UpsertServer(server)
	require.NoError(t, err)

	balancer2, err := roundrobin.New(nil)
	require.NoError(t, err)

	err = balancer2.UpsertServer(server)
	require.NoError(t, err)

	balancers := Balancers([]Balancer{balancer1, balancer2})

	err = balancers.RemoveServer(server)
	require.NoError(t, err)

	assert.Equal(t, 0, len(balancer1.Servers()))
	assert.Equal(t, 0, len(balancer2.Servers()))
}

type testLoadBalancer struct {
	// RWMutex needed due to parallel test execution: Both the system-under-test
	// and the test assertions reference the counters.
	*sync.RWMutex
	numRemovedServers  int
	numUpsertedServers int
	servers            []*url.URL
	// options is just to make sure that LBStatusUpdater forwards options on Upsert to its BalancerHandler
	options []roundrobin.ServerOption
}

func (lb *testLoadBalancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// noop
}

func (lb *testLoadBalancer) RemoveServer(u *url.URL) error {
	lb.Lock()
	defer lb.Unlock()
	lb.numRemovedServers++
	lb.removeServer(u)
	return nil
}

func (lb *testLoadBalancer) UpsertServer(u *url.URL, options ...roundrobin.ServerOption) error {
	lb.Lock()
	defer lb.Unlock()
	lb.numUpsertedServers++
	lb.servers = append(lb.servers, u)
	lb.options = append(lb.options, options...)
	return nil
}

func (lb *testLoadBalancer) Servers() []*url.URL {
	return lb.servers
}

func (lb *testLoadBalancer) Options() []roundrobin.ServerOption {
	return lb.options
}

func (lb *testLoadBalancer) removeServer(u *url.URL) {
	var i int
	var serverURL *url.URL
	found := false
	for i, serverURL = range lb.servers {
		if *serverURL == *u {
			found = true
			break
		}
	}
	if !found {
		return
	}

	lb.servers = append(lb.servers[:i], lb.servers[i+1:]...)
}

func newTestServer(done func(), healthSequence []int) *httptest.Server {
	handler := &testHandler{
		done:           done,
		healthSequence: healthSequence,
	}
	return httptest.NewServer(handler)
}

// ServeHTTP returns HTTP response codes following a status sequences.
// It calls the given 'done' function once all request health indicators have been depleted.
func (th *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(th.healthSequence) == 0 {
		panic("received unexpected request")
	}

	w.WriteHeader(th.healthSequence[0])

	th.healthSequence = th.healthSequence[1:]
	if len(th.healthSequence) == 0 {
		th.done()
	}
}

func TestLBStatusUpdater(t *testing.T) {
	lb := &testLoadBalancer{RWMutex: &sync.RWMutex{}}
	svInfo := &runtime.ServiceInfo{}
	lbsu := NewLBStatusUpdater(lb, svInfo, nil)
	newServer, err := url.Parse("http://foo.com")
	assert.NoError(t, err)
	err = lbsu.UpsertServer(newServer, roundrobin.Weight(1))
	assert.NoError(t, err)
	assert.Equal(t, len(lbsu.Servers()), 1)
	assert.Equal(t, len(lbsu.BalancerHandler.(*testLoadBalancer).Options()), 1)
	statuses := svInfo.GetAllStatus()
	assert.Equal(t, len(statuses), 1)
	for k, v := range statuses {
		assert.Equal(t, k, newServer.String())
		assert.Equal(t, v, serverUp)
		break
	}
	err = lbsu.RemoveServer(newServer)
	assert.NoError(t, err)
	assert.Equal(t, len(lbsu.Servers()), 0)
	statuses = svInfo.GetAllStatus()
	assert.Equal(t, len(statuses), 1)
	for k, v := range statuses {
		assert.Equal(t, k, newServer.String())
		assert.Equal(t, v, serverDown)
		break
	}
}

func TestNotFollowingRedirects(t *testing.T) {
	redirectServerCalled := false
	redirectTestServer := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		redirectServerCalled = true
	}))
	defer redirectTestServer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Add("location", redirectTestServer.URL)
		rw.WriteHeader(http.StatusSeeOther)
		cancel()
	}))
	defer server.Close()

	lb := &testLoadBalancer{
		RWMutex: &sync.RWMutex{},
		servers: []*url.URL{testhelpers.MustParseURL(server.URL)},
	}

	backend := NewBackendConfig(Options{
		Path:            "/path",
		Interval:        healthCheckInterval,
		Timeout:         healthCheckTimeout,
		LB:              lb,
		FollowRedirects: false,
	}, "backendName")

	collectingMetrics := &testhelpers.CollectingGauge{}
	check := HealthCheck{
		Backends: make(map[string]*BackendConfig),
		metrics:  metricsHealthcheck{serverUpGauge: collectingMetrics},
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		check.execute(ctx, backend)
		wg.Done()
	}()

	timeout := time.Duration(int(healthCheckInterval) + 500)
	select {
	case <-time.After(timeout):
		t.Fatal("test did not complete in time")
	case <-ctx.Done():
		wg.Wait()
	}

	assert.False(t, redirectServerCalled, "HTTP redirect must not be followed")
}
