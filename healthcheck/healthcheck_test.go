package healthcheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/containous/traefik/testhelpers"
	"github.com/vulcand/oxy/roundrobin"
)

const healthCheckInterval = 100 * time.Millisecond

type testHandler struct {
	done           func()
	healthSequence []bool
}

func TestSetBackendsConfiguration(t *testing.T) {
	tests := []struct {
		desc                   string
		startHealthy           bool
		healthSequence         []bool
		wantNumRemovedServers  int
		wantNumUpsertedServers int
	}{
		{
			desc:                   "healthy server staying healthy",
			startHealthy:           true,
			healthSequence:         []bool{true},
			wantNumRemovedServers:  0,
			wantNumUpsertedServers: 0,
		},
		{
			desc:                   "healthy server becoming sick",
			startHealthy:           true,
			healthSequence:         []bool{false},
			wantNumRemovedServers:  1,
			wantNumUpsertedServers: 0,
		},
		{
			desc:                   "sick server becoming healthy",
			startHealthy:           false,
			healthSequence:         []bool{true},
			wantNumRemovedServers:  0,
			wantNumUpsertedServers: 1,
		},
		{
			desc:                   "sick server staying sick",
			startHealthy:           false,
			healthSequence:         []bool{false},
			wantNumRemovedServers:  0,
			wantNumUpsertedServers: 0,
		},
		{
			desc:                   "healthy server toggling to sick and back to healthy",
			startHealthy:           true,
			healthSequence:         []bool{false, true},
			wantNumRemovedServers:  1,
			wantNumUpsertedServers: 1,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			// The context is passed to the health check and canonically cancelled by
			// the test server once all expected requests have been received.
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ts := newTestServer(cancel, test.healthSequence)
			defer ts.Close()

			lb := &testLoadBalancer{RWMutex: &sync.RWMutex{}}
			backend := NewBackendHealthCheck(Options{
				Path:     "/path",
				Interval: healthCheckInterval,
				LB:       lb,
			})
			serverURL := testhelpers.MustParseURL(ts.URL)
			if test.startHealthy {
				lb.servers = append(lb.servers, serverURL)
			} else {
				backend.disabledURLs = append(backend.disabledURLs, serverURL)
			}

			check := HealthCheck{
				Backends: make(map[string]*BackendHealthCheck),
			}
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				check.execute(ctx, "id", backend)
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
			if lb.numRemovedServers != test.wantNumRemovedServers {
				t.Errorf("got %d removed servers, wanted %d", lb.numRemovedServers, test.wantNumRemovedServers)
			}

			if lb.numUpsertedServers != test.wantNumUpsertedServers {
				t.Errorf("got %d upserted servers, wanted %d", lb.numUpsertedServers, test.wantNumUpsertedServers)
			}
		})
	}
}

func TestNewRequest(t *testing.T) {
	tests := []struct {
		desc     string
		host     string
		port     int
		path     string
		expected string
	}{
		{
			desc:     "no port override",
			host:     "backend1:80",
			port:     0,
			path:     "/test",
			expected: "http://backend1:80/test",
		},
		{
			desc:     "port override",
			host:     "backend2:80",
			port:     8080,
			path:     "/test",
			expected: "http://backend2:8080/test",
		},
		{
			desc:     "no port override with no port in host",
			host:     "backend1",
			port:     0,
			path:     "/health",
			expected: "http://backend1/health",
		},
		{
			desc:     "port override with no port in host",
			host:     "backend2",
			port:     8080,
			path:     "/health",
			expected: "http://backend2:8080/health",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			backend := NewBackendHealthCheck(
				Options{
					Path: test.path,
					Port: test.port,
				})

			u := &url.URL{
				Scheme: "http",
				Host:   test.host,
			}

			req, err := backend.newRequest(u)
			if err != nil {
				t.Fatalf("failed to create new backend request: %s", err)
			}

			actual := req.URL.String()
			if actual != test.expected {
				t.Fatalf("got %s for healthcheck URL, wanted %s", actual, test.expected)
			}
		})
	}
}

type testLoadBalancer struct {
	// RWMutex needed due to parallel test execution: Both the system-under-test
	// and the test assertions reference the counters.
	*sync.RWMutex
	numRemovedServers  int
	numUpsertedServers int
	servers            []*url.URL
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
	return nil
}

func (lb *testLoadBalancer) Servers() []*url.URL {
	return lb.servers
}

func (lb *testLoadBalancer) removeServer(u *url.URL) {
	var i int
	var serverURL *url.URL
	for i, serverURL = range lb.servers {
		if *serverURL == *u {
			break
		}
	}

	lb.servers = append(lb.servers[:i], lb.servers[i+1:]...)
}

func newTestServer(done func(), healthSequence []bool) *httptest.Server {
	handler := &testHandler{
		done:           done,
		healthSequence: healthSequence,
	}
	return httptest.NewServer(handler)
}

// ServeHTTP returns 200 or 503 HTTP response codes depending on whether the
// current request is marked as healthy or not.
// It calls the given 'done' function once all request health indicators have
// been depleted.
func (th *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(th.healthSequence) == 0 {
		panic("received unexpected request")
	}

	healthy := th.healthSequence[0]
	if healthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	th.healthSequence = th.healthSequence[1:]
	if len(th.healthSequence) == 0 {
		th.done()
	}
}
