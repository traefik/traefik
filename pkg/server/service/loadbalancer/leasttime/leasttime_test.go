package leasttime

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type key string

const serviceName key = "serviceName"

func pointer[T any](v T) *T { return &v }

// responseRecorder tracks which servers handled requests.
type responseRecorder struct {
	*httptest.ResponseRecorder
	save map[string]int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	server := r.Header().Get("server")
	if server != "" {
		r.save[server]++
	}
	r.ResponseRecorder.WriteHeader(statusCode)
}

// TestBalancer tests basic server addition and least-time selection.
func TestBalancer(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(5 * time.Millisecond)
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(5 * time.Millisecond)
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 10 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// With least-time and equal response times, both servers should get some traffic.
	assert.Positive(t, recorder.save["first"])
	assert.Positive(t, recorder.save["second"])
	assert.Equal(t, 10, recorder.save["first"]+recorder.save["second"])
}

// TestBalancerNoService tests behavior when no servers are configured.
func TestBalancerNoService(t *testing.T) {
	balancer := New(nil, false)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

// TestBalancerNoServiceUp tests behavior when all servers are marked down.
func TestBalancerNoServiceUp(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}), pointer(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}), pointer(1), false)

	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "first", false)
	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "second", false)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

// TestBalancerOneServerDown tests that down servers are excluded from selection.
func TestBalancerOneServerDown(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}), pointer(1), false)
	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "second", false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 3 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	assert.Equal(t, 3, recorder.save["first"])
	assert.Equal(t, 0, recorder.save["second"])
}

// TestBalancerOneServerDownThenUp tests server status transitions.
func TestBalancerOneServerDownThenUp(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(5 * time.Millisecond)
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(5 * time.Millisecond)
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)
	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "second", false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 3 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Equal(t, 3, recorder.save["first"])
	assert.Equal(t, 0, recorder.save["second"])

	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "second", true)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 20 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	// Both servers should get some traffic.
	assert.Positive(t, recorder.save["first"])
	assert.Positive(t, recorder.save["second"])
	assert.Equal(t, 20, recorder.save["first"]+recorder.save["second"])
}

// TestBalancerAllServersZeroWeight tests that all zero-weight servers result in no available server.
func TestBalancerAllServersZeroWeight(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), pointer(0), false)
	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), pointer(0), false)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

// TestBalancerOneServerZeroWeight tests that zero-weight servers are ignored.
func TestBalancerOneServerZeroWeight(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), pointer(0), false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 3 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Only first server should receive traffic.
	assert.Equal(t, 3, recorder.save["first"])
	assert.Equal(t, 0, recorder.save["second"])
}

// TestBalancerPropagate tests status propagation to parent balancers.
func TestBalancerPropagate(t *testing.T) {
	balancer1 := New(nil, true)

	balancer1.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)
	balancer1.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer2 := New(nil, true)
	balancer2.Add("third", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "third")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)
	balancer2.Add("fourth", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "fourth")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	topBalancer := New(nil, true)
	topBalancer.Add("balancer1", balancer1, pointer(1), false)
	topBalancer.Add("balancer2", balancer2, pointer(1), false)
	err := balancer1.RegisterStatusUpdater(func(up bool) {
		topBalancer.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "balancer1", up)
	})
	assert.NoError(t, err)
	err = balancer2.RegisterStatusUpdater(func(up bool) {
		topBalancer.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "balancer2", up)
	})
	assert.NoError(t, err)

	// Set all children of balancer1 to down, should propagate to top.
	balancer1.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "first", false)
	balancer1.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "second", false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 4 {
		topBalancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Only balancer2 should receive traffic.
	assert.Equal(t, 0, recorder.save["first"])
	assert.Equal(t, 0, recorder.save["second"])
	assert.Equal(t, 4, recorder.save["third"]+recorder.save["fourth"])
}

// TestBalancerOneServerFenced tests that fenced servers are excluded from selection.
func TestBalancerOneServerFenced(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), true) // fenced

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 3 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Only first server should receive traffic.
	assert.Equal(t, 3, recorder.save["first"])
	assert.Equal(t, 0, recorder.save["second"])
}

// TestBalancerAllFencedServers tests that all fenced servers result in no available server.
func TestBalancerAllFencedServers(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), pointer(1), true)
	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), pointer(1), true)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

// TestBalancerRegisterStatusUpdaterWithoutHealthCheck tests error when registering updater without health check.
func TestBalancerRegisterStatusUpdaterWithoutHealthCheck(t *testing.T) {
	balancer := New(nil, false)

	err := balancer.RegisterStatusUpdater(func(up bool) {})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "healthCheck not enabled")
}

// TestBalancerSticky tests sticky session support.
func TestBalancerSticky(t *testing.T) {
	balancer := New(&dynamic.Sticky{
		Cookie: &dynamic.Cookie{
			Name: "test",
		},
	}, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// First request should set cookie.
	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	firstServer := recorder.Header().Get("server")
	assert.NotEmpty(t, firstServer)

	// Extract cookie from first response.
	cookies := recorder.Result().Cookies()
	assert.NotEmpty(t, cookies)

	// Second request with cookie should hit same server.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	recorder2 := httptest.NewRecorder()
	balancer.ServeHTTP(recorder2, req)
	secondServer := recorder2.Header().Get("server")

	assert.Equal(t, firstServer, secondServer)
}

// TestBalancerStickyFallback tests that sticky sessions fallback to least-time when sticky server is down.
func TestBalancerStickyFallback(t *testing.T) {
	balancer := New(&dynamic.Sticky{
		Cookie: &dynamic.Cookie{
			Name: "test",
		},
	}, false)

	balancer.Add("server1", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(50 * time.Millisecond)
		rw.Header().Set("server", "server1")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("server2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(50 * time.Millisecond)
		rw.Header().Set("server", "server2")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Make initial request to establish sticky session with server1.
	recorder1 := httptest.NewRecorder()
	balancer.ServeHTTP(recorder1, httptest.NewRequest(http.MethodGet, "/", nil))
	firstServer := recorder1.Header().Get("server")
	assert.NotEmpty(t, firstServer)

	// Extract cookie from first response.
	cookies := recorder1.Result().Cookies()
	assert.NotEmpty(t, cookies)

	// Mark the sticky server as DOWN
	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "test"), firstServer, false)

	// Request with sticky cookie should fallback to the other server
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	recorder2 := httptest.NewRecorder()
	balancer.ServeHTTP(recorder2, req2)
	fallbackServer := recorder2.Header().Get("server")
	assert.NotEqual(t, firstServer, fallbackServer)
	assert.NotEmpty(t, fallbackServer)

	// New sticky cookie should be written for the fallback server
	newCookies := recorder2.Result().Cookies()
	assert.NotEmpty(t, newCookies)

	// Verify sticky session persists with the fallback server
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range newCookies {
		req3.AddCookie(cookie)
	}
	recorder3 := httptest.NewRecorder()
	balancer.ServeHTTP(recorder3, req3)
	assert.Equal(t, fallbackServer, recorder3.Header().Get("server"))

	// Bring original server back UP
	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "test"), firstServer, true)

	// Request with fallback server cookie should still stick to fallback server
	req4 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range newCookies {
		req4.AddCookie(cookie)
	}
	recorder4 := httptest.NewRecorder()
	balancer.ServeHTTP(recorder4, req4)
	assert.Equal(t, fallbackServer, recorder4.Header().Get("server"))
}

// TestBalancerStickyFenced tests that sticky sessions persist to fenced servers (graceful shutdown)
// Fencing enables zero-downtime deployments: fenced servers reject NEW connections
// but continue serving EXISTING sticky sessions until they complete.
func TestBalancerStickyFenced(t *testing.T) {
	balancer := New(&dynamic.Sticky{
		Cookie: &dynamic.Cookie{
			Name: "test",
		},
	}, false)

	balancer.Add("server1", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "server1")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("server2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "server2")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Establish sticky session with any server.
	recorder1 := httptest.NewRecorder()
	balancer.ServeHTTP(recorder1, httptest.NewRequest(http.MethodGet, "/", nil))
	stickyServer := recorder1.Header().Get("server")
	assert.NotEmpty(t, stickyServer)

	cookies := recorder1.Result().Cookies()
	assert.NotEmpty(t, cookies)

	// Fence the sticky server (simulate graceful shutdown).
	balancer.handlersMu.Lock()
	balancer.fenced[stickyServer] = struct{}{}
	balancer.handlersMu.Unlock()

	// Existing sticky session should STILL work (graceful draining).
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	recorder2 := httptest.NewRecorder()
	balancer.ServeHTTP(recorder2, req)
	assert.Equal(t, stickyServer, recorder2.Header().Get("server"))

	// But NEW requests should NOT go to the fenced server.
	recorder3 := httptest.NewRecorder()
	balancer.ServeHTTP(recorder3, httptest.NewRequest(http.MethodGet, "/", nil))
	newServer := recorder3.Header().Get("server")
	assert.NotEqual(t, stickyServer, newServer)
	assert.NotEmpty(t, newServer)
}

// TestRingBufferBasic tests basic ring buffer functionality with few samples.
func TestRingBufferBasic(t *testing.T) {
	handler := &namedHandler{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}),
		name:    "test",
		weight:  1,
	}

	// Test cold start - no samples.
	avg := handler.getAvgResponseTime()
	assert.InDelta(t, 0.0, avg, 0)

	// Add one sample.
	handler.updateResponseTime(10 * time.Millisecond)
	avg = handler.getAvgResponseTime()
	assert.InDelta(t, 10.0, avg, 0)

	// Add more samples.
	handler.updateResponseTime(20 * time.Millisecond)
	handler.updateResponseTime(30 * time.Millisecond)
	avg = handler.getAvgResponseTime()
	assert.InDelta(t, 20.0, avg, 0) // (10 + 20 + 30) / 3 = 20
}

// TestRingBufferWraparound tests ring buffer behavior when it wraps around
func TestRingBufferWraparound(t *testing.T) {
	handler := &namedHandler{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}),
		name:    "test",
		weight:  1,
	}

	// Fill the buffer with 100 samples of 10ms each.
	for range sampleSize {
		handler.updateResponseTime(10 * time.Millisecond)
	}
	avg := handler.getAvgResponseTime()
	assert.InDelta(t, 10.0, avg, 0)

	// Add one more sample (should replace oldest).
	handler.updateResponseTime(20 * time.Millisecond)
	avg = handler.getAvgResponseTime()
	// Sum: 99*10 + 1*20 = 1010, avg = 1010/100 = 10.1
	assert.InDelta(t, 10.1, avg, 0)

	// Add 10 more samples of 30ms.
	for range 10 {
		handler.updateResponseTime(30 * time.Millisecond)
	}
	avg = handler.getAvgResponseTime()
	// Sum: 89*10 + 1*20 + 10*30 = 890 + 20 + 300 = 1210, avg = 1210/100 = 12.1
	assert.InDelta(t, 12.1, avg, 0)
}

// TestRingBufferLarge tests ring buffer with many samples (> 100).
func TestRingBufferLarge(t *testing.T) {
	handler := &namedHandler{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}),
		name:    "test",
		weight:  1,
	}

	// Add 150 samples.
	for i := range 150 {
		handler.updateResponseTime(time.Duration(i+1) * time.Millisecond)
	}

	// Should only track last 100 samples: 51, 52, ..., 150
	// Sum = (51 + 150) * 100 / 2 = 10050
	// Avg = 10050 / 100 = 100.5
	avg := handler.getAvgResponseTime()
	assert.InDelta(t, 100.5, avg, 0)
}

// TestInflightCounter tests inflight request tracking.
func TestInflightCounter(t *testing.T) {
	balancer := New(nil, false)

	var inflightAtRequest atomic.Int64

	balancer.Add("test", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		inflightAtRequest.Store(balancer.handlers[0].inflightCount.Load())
		rw.Header().Set("server", "test")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Check that inflight count is 0 initially.
	balancer.handlersMu.RLock()
	handler := balancer.handlers[0]
	balancer.handlersMu.RUnlock()
	assert.Equal(t, int64(0), handler.inflightCount.Load())

	// Make a request.
	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	// During request, inflight should have been 1.
	assert.Equal(t, int64(1), inflightAtRequest.Load())

	// After request completes, inflight should be back to 0.
	assert.Equal(t, int64(0), handler.inflightCount.Load())
}

// TestConcurrentResponseTimeUpdates tests thread safety of response time updates.
func TestConcurrentResponseTimeUpdates(t *testing.T) {
	handler := &namedHandler{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}),
		name:    "test",
		weight:  1,
	}

	// Concurrently update response times.
	var wg sync.WaitGroup
	numGoroutines := 10
	updatesPerGoroutine := 20

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for range updatesPerGoroutine {
				handler.updateResponseTime(time.Duration(id+1) * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Should have exactly 100 samples (buffer size).
	assert.Equal(t, sampleSize, handler.sampleCount)
}

// TestConcurrentInflightTracking tests thread safety of inflight counter.
func TestConcurrentInflightTracking(t *testing.T) {
	handler := &namedHandler{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			time.Sleep(10 * time.Millisecond)
			rw.WriteHeader(http.StatusOK)
		}),
		name:   "test",
		weight: 1,
	}

	var maxInflight atomic.Int64

	var wg sync.WaitGroup
	numRequests := 50

	for range numRequests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler.inflightCount.Add(1)
			defer handler.inflightCount.Add(-1)

			// Track maximum inflight count.
			current := handler.inflightCount.Load()
			for {
				maxLoad := maxInflight.Load()
				if current <= maxLoad || maxInflight.CompareAndSwap(maxLoad, current) {
					break
				}
			}

			time.Sleep(1 * time.Millisecond)
		}()
	}

	wg.Wait()

	// All requests completed, inflight should be 0.
	assert.Equal(t, int64(0), handler.inflightCount.Load())
	// Max inflight should be > 1 (concurrent requests).
	assert.Greater(t, maxInflight.Load(), int64(1))
}

// TestConcurrentRequestsRespectInflight tests that the load balancer dynamically
// adapts to inflight request counts during concurrent request processing.
func TestConcurrentRequestsRespectInflight(t *testing.T) {
	balancer := New(nil, false)

	// Use a channel to control when handlers start sleeping.
	// This ensures we can fill one server with inflight requests before routing new ones.
	blockChan := make(chan struct{})

	// Add two servers with equal response times and weights.
	balancer.Add("server1", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		<-blockChan // Wait for signal to proceed.
		time.Sleep(10 * time.Millisecond)
		rw.Header().Set("server", "server1")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("server2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		<-blockChan // Wait for signal to proceed.
		time.Sleep(10 * time.Millisecond)
		rw.Header().Set("server", "server2")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Pre-warm both servers to establish equal average response times.
	for i := range sampleSize {
		balancer.handlers[0].responseTimes[i] = 10.0
	}
	balancer.handlers[0].responseTimeSum = 10.0 * sampleSize
	balancer.handlers[0].sampleCount = sampleSize

	for i := range sampleSize {
		balancer.handlers[1].responseTimes[i] = 10.0
	}
	balancer.handlers[1].responseTimeSum = 10.0 * sampleSize
	balancer.handlers[1].sampleCount = sampleSize

	// Phase 1: Launch concurrent requests to server1 that will block.
	var wg sync.WaitGroup
	inflightRequests := 5

	for range inflightRequests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			recorder := httptest.NewRecorder()
			balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
		}()
	}

	// Wait for goroutines to start and increment inflight counters.
	// They will block on the channel, keeping inflight count high.
	time.Sleep(50 * time.Millisecond)

	// Verify inflight counts before making new requests.
	server1Inflight := balancer.handlers[0].inflightCount.Load()
	server2Inflight := balancer.handlers[1].inflightCount.Load()
	assert.Equal(t, int64(5), server1Inflight+server2Inflight)

	// Phase 2: Make new requests while the initial requests are blocked.
	// These should see the high inflight counts and route to the less-loaded server.
	var saveMu sync.Mutex
	save := map[string]int{}
	newRequests := 50

	// Launch new requests in background so they don't block.
	var newWg sync.WaitGroup
	for range newRequests {
		newWg.Add(1)
		go func() {
			defer newWg.Done()
			rec := httptest.NewRecorder()
			balancer.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			server := rec.Header().Get("server")
			if server != "" {
				saveMu.Lock()
				save[server]++
				saveMu.Unlock()
			}
		}()
	}

	// Wait for new requests to start and see the inflight counts.
	time.Sleep(50 * time.Millisecond)

	close(blockChan)

	wg.Wait()
	newWg.Wait()

	saveMu.Lock()
	total := save["server1"] + save["server2"]
	server1Count := save["server1"]
	server2Count := save["server2"]
	saveMu.Unlock()

	assert.Equal(t, newRequests, total)

	// With inflight tracking, load should naturally balance toward equal distribution.
	// We allow variance due to concurrent execution and race windows in server selection.
	assert.InDelta(t, 25.0, float64(server1Count), 5.0) // 20-30 requests
	assert.InDelta(t, 25.0, float64(server2Count), 5.0) // 20-30 requests
}

// TestTTFBMeasurement tests TTFB measurement accuracy.
func TestTTFBMeasurement(t *testing.T) {
	balancer := New(nil, false)

	// Add server with known delay.
	delay := 50 * time.Millisecond
	balancer.Add("slow", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(delay)
		rw.Header().Set("server", "slow")
		rw.WriteHeader(http.StatusOK)
		httptrace.ContextClientTrace(req.Context()).GotFirstResponseByte()
	}), pointer(1), false)

	// Make multiple requests to build average.
	for range 5 {
		recorder := httptest.NewRecorder()
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Check that average response time is approximately the delay.
	avg := balancer.handlers[0].getAvgResponseTime()

	// Allow 5ms tolerance for Go timing jitter and test environment variations.
	assert.InDelta(t, float64(delay.Milliseconds()), avg, 5.0)
}

// TestZeroSamplesReturnsZero tests that getAvgResponseTime returns 0 when no samples.
func TestZeroSamplesReturnsZero(t *testing.T) {
	handler := &namedHandler{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}),
		name:    "test",
		weight:  1,
	}

	avg := handler.getAvgResponseTime()
	assert.InDelta(t, 0.0, avg, 0)
}

// TestScoreCalculationWithWeights tests that weights are properly considered in score calculation.
func TestScoreCalculationWithWeights(t *testing.T) {
	balancer := New(nil, false)

	// Add two servers with same response time but different weights.
	// Server with higher weight should be preferred.
	balancer.Add("weighted", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(50 * time.Millisecond)
		rw.Header().Set("server", "weighted")
		rw.WriteHeader(http.StatusOK)
		httptrace.ContextClientTrace(req.Context()).GotFirstResponseByte()
	}), pointer(3), false) // Weight 3

	balancer.Add("normal", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(50 * time.Millisecond)
		rw.Header().Set("server", "normal")
		rw.WriteHeader(http.StatusOK)
		httptrace.ContextClientTrace(req.Context()).GotFirstResponseByte()
	}), pointer(1), false) // Weight 1

	// Make requests to build up response time averages.
	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 2 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Score for weighted: (50 × (1 + 0)) / 3 = 16.67
	// Score for normal: (50 × (1 + 0)) / 1 = 50
	// After warmup, weighted server has 3x better score (16.67 vs 50) and should receive nearly all requests.
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 10 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	assert.Equal(t, 10, recorder.save["weighted"])
	assert.Zero(t, recorder.save["normal"])
}

// TestScoreCalculationWithInflight tests that inflight requests are considered in score calculation.
func TestScoreCalculationWithInflight(t *testing.T) {
	balancer := New(nil, false)

	// We'll manually control the inflight counters to test the score calculation.
	// Add two servers with same response time.
	balancer.Add("server1", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(10 * time.Millisecond)
		rw.Header().Set("server", "server1")
		rw.WriteHeader(http.StatusOK)
		httptrace.ContextClientTrace(req.Context()).GotFirstResponseByte()
	}), pointer(1), false)

	balancer.Add("server2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(10 * time.Millisecond)
		rw.Header().Set("server", "server2")
		rw.WriteHeader(http.StatusOK)
		httptrace.ContextClientTrace(req.Context()).GotFirstResponseByte()
	}), pointer(1), false)

	// Build up response time averages for both servers.
	for range 2 {
		recorder := httptest.NewRecorder()
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Now manually set server1 to have high inflight count.
	balancer.handlers[0].inflightCount.Store(5)

	// Make requests - they should prefer server2 because:
	// Score for server1: (10 × (1 + 5)) / 1 = 60
	// Score for server2: (10 × (1 + 0)) / 1 = 10
	recorder := &responseRecorder{save: map[string]int{}}
	for range 5 {
		// Manually increment to simulate the ServeHTTP behavior.
		server, _ := balancer.nextServer()
		server.inflightCount.Add(1)

		if server.name == "server1" {
			recorder.save["server1"]++
		} else {
			recorder.save["server2"]++
		}
	}

	// Server2 should get all requests
	assert.Equal(t, 5, recorder.save["server2"])
	assert.Zero(t, recorder.save["server1"])
}

// TestScoreCalculationColdStart tests that new servers (0ms avg) get fair selection
func TestScoreCalculationColdStart(t *testing.T) {
	balancer := New(nil, false)

	// Add a warm server with established response time
	balancer.Add("warm", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(50 * time.Millisecond)
		rw.Header().Set("server", "warm")
		rw.WriteHeader(http.StatusOK)
		httptrace.ContextClientTrace(req.Context()).GotFirstResponseByte()
	}), pointer(1), false)

	// Warm up the first server
	for range 10 {
		recorder := httptest.NewRecorder()
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Now add a cold server (new, no response time data)
	balancer.Add("cold", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(10 * time.Millisecond) // Actually faster
		rw.Header().Set("server", "cold")
		rw.WriteHeader(http.StatusOK)
		httptrace.ContextClientTrace(req.Context()).GotFirstResponseByte()
	}), pointer(1), false)

	// Cold server should get selected because:
	// Score for warm: (50 × (1 + 0)) / 1 = 50
	// Score for cold: (0 × (1 + 0)) / 1 = 0
	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 20 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Cold server should get all or most requests initially due to 0ms average
	assert.Greater(t, recorder.save["cold"], 10)

	// After cold server builds up its average, it should continue to get more traffic
	// because it's actually faster (10ms vs 50ms)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 20 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Greater(t, recorder.save["cold"], recorder.save["warm"])
}

// TestFastServerGetsMoreTraffic verifies that servers with lower response times
// receive proportionally more traffic in steady state (after cold start).
// This tests the core selection bias of the least-time algorithm.
func TestFastServerGetsMoreTraffic(t *testing.T) {
	balancer := New(nil, false)

	// Add two servers with different static response times.
	balancer.Add("fast", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(20 * time.Millisecond)
		rw.Header().Set("server", "fast")
		rw.WriteHeader(http.StatusOK)
		httptrace.ContextClientTrace(req.Context()).GotFirstResponseByte()
	}), pointer(1), false)

	balancer.Add("slow", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(100 * time.Millisecond)
		rw.Header().Set("server", "slow")
		rw.WriteHeader(http.StatusOK)
		httptrace.ContextClientTrace(req.Context()).GotFirstResponseByte()
	}), pointer(1), false)

	// After just 1 request to each server, the algorithm identifies the fastest server
	// and routes nearly all subsequent traffic there (converges in ~2 requests).
	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 50 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	assert.Greater(t, recorder.save["fast"], recorder.save["slow"])
	assert.Greater(t, recorder.save["fast"], 48) // Expect ~96-98% to fast server (48-49/50).
}

// TestTrafficShiftsWhenPerformanceDegrades verifies that the load balancer
// adapts to changing server performance by shifting traffic away from degraded servers.
// This tests the adaptive behavior - the core value proposition of least-time load balancing.
func TestTrafficShiftsWhenPerformanceDegrades(t *testing.T) {
	balancer := New(nil, false)

	// Use atomic to dynamically control server1's response time.
	server1Delay := atomic.Int64{}
	server1Delay.Store(5) // Start with 5ms

	balancer.Add("server1", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Duration(server1Delay.Load()) * time.Millisecond)
		rw.Header().Set("server", "server1")
		rw.WriteHeader(http.StatusOK)
		httptrace.ContextClientTrace(req.Context()).GotFirstResponseByte()
	}), pointer(1), false)

	balancer.Add("server2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(5 * time.Millisecond) // Static 5ms
		rw.Header().Set("server", "server2")
		rw.WriteHeader(http.StatusOK)
		httptrace.ContextClientTrace(req.Context()).GotFirstResponseByte()
	}), pointer(1), false)

	// Pre-fill ring buffers to eliminate cold start effects and ensure deterministic equal performance state.
	for _, h := range balancer.handlers {
		for i := range sampleSize {
			h.responseTimes[i] = 5.0
		}
		h.responseTimeSum = 5.0 * sampleSize
		h.sampleCount = sampleSize
	}

	// Phase 1: Both servers perform equally (5ms each).
	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 50 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// With equal performance and pre-filled buffers, distribution should be balanced via WRR tie-breaking.
	total := recorder.save["server1"] + recorder.save["server2"]
	assert.Equal(t, 50, total)
	assert.InDelta(t, 25, recorder.save["server1"], 10) // 25 ± 10 requests
	assert.InDelta(t, 25, recorder.save["server2"], 10) // 25 ± 10 requests

	// Phase 2: server1 degrades (simulating GC pause, CPU spike, or network latency).
	server1Delay.Store(50) // Now 50ms (10x slower) - dramatic degradation for reliable detection

	// Make more requests to shift the moving average.
	// Ring buffer has 100 samples, need significant new samples to shift average.
	// server1's average will climb from ~5ms toward 50ms.
	recorder2 := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 60 {
		balancer.ServeHTTP(recorder2, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// server2 should get significantly more traffic
	// With 10x performance difference, server2 should dominate.
	total2 := recorder2.save["server1"] + recorder2.save["server2"]
	assert.Equal(t, 60, total2)
	assert.Greater(t, recorder2.save["server2"], 35) // At least ~60% (35/60)
	assert.Less(t, recorder2.save["server1"], 25)    // At most ~40% (25/60)
}

// TestMultipleServersWithSameScore tests WRR tie-breaking when multiple servers have identical scores.
// Uses nextServer() directly to avoid timing variations in the test.
func TestMultipleServersWithSameScore(t *testing.T) {
	balancer := New(nil, false)

	// Add three servers with identical response times and weights.
	balancer.Add("server1", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(5 * time.Millisecond)
		rw.Header().Set("server", "server1")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("server2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(5 * time.Millisecond)
		rw.Header().Set("server", "server2")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("server3", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(5 * time.Millisecond)
		rw.Header().Set("server", "server3")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Set all servers to identical response times to trigger tie-breaking.
	for _, h := range balancer.handlers {
		for i := range sampleSize {
			h.responseTimes[i] = 5.0
		}
		h.responseTimeSum = 5.0 * sampleSize
		h.sampleCount = sampleSize
	}

	// With all servers having identical scores, WRR tie-breaking should distribute fairly.
	// Test the selection logic directly without actual HTTP requests to avoid timing variations.
	counts := map[string]int{"server1": 0, "server2": 0, "server3": 0}
	for range 90 {
		server, err := balancer.nextServer()
		assert.NoError(t, err)
		counts[server.name]++
	}

	total := counts["server1"] + counts["server2"] + counts["server3"]
	assert.Equal(t, 90, total)

	// With WRR and 90 requests, each server should get ~30 requests (±1 due to initialization).
	assert.InDelta(t, 30, counts["server1"], 1)
	assert.InDelta(t, 30, counts["server2"], 1)
	assert.InDelta(t, 30, counts["server3"], 1)
}

// TestWRRTieBreakingWeightedDistribution tests weighted distribution among tied servers.
// Uses nextServer() directly to avoid timing variations in the test.
func TestWRRTieBreakingWeightedDistribution(t *testing.T) {
	balancer := New(nil, false)

	// Add two servers with different weights.
	// To create equal scores, response times must be proportional to weights.
	balancer.Add("weighted", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(15 * time.Millisecond) // 3x longer due to 3x weight
		rw.Header().Set("server", "weighted")
		rw.WriteHeader(http.StatusOK)
	}), pointer(3), false) // Weight 3

	balancer.Add("normal", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(5 * time.Millisecond)
		rw.Header().Set("server", "normal")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false) // Weight 1

	// Since response times is proportional to weights, both scores are equal, so WRR tie-breaking will apply.
	// weighted: score = (15 * 1) / 3 = 5
	// normal: score = (5 * 1) / 1 = 5
	for i := range sampleSize {
		balancer.handlers[0].responseTimes[i] = 15.0
	}
	balancer.handlers[0].responseTimeSum = 15.0 * sampleSize
	balancer.handlers[0].sampleCount = sampleSize

	for i := range sampleSize {
		balancer.handlers[1].responseTimes[i] = 5.0
	}
	balancer.handlers[1].responseTimeSum = 5.0 * sampleSize
	balancer.handlers[1].sampleCount = sampleSize

	// Test the selection logic directly without actual HTTP requests to avoid timing variations.
	counts := map[string]int{"weighted": 0, "normal": 0}
	for range 80 {
		server, err := balancer.nextServer()
		assert.NoError(t, err)
		counts[server.name]++
	}

	total := counts["weighted"] + counts["normal"]
	assert.Equal(t, 80, total)

	// With 3:1 weight ratio, distribution should be ~75%/25% (60/80 and 20/80), ±1 due to initialization.
	assert.InDelta(t, 60, counts["weighted"], 1)
	assert.InDelta(t, 20, counts["normal"], 1)
}
