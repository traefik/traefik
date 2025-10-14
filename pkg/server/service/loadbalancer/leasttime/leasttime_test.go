package leasttime

import (
	"context"
	"net/http"
	"net/http/httptest"
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

// responseRecorder tracks which servers handled requests
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

// TestBalancer tests basic server addition and least-time selection
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
	for range 20 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// With least-time and equal response times, distribution should be fairly balanced
	// Both servers should get some traffic
	assert.Greater(t, recorder.save["first"], 0)
	assert.Greater(t, recorder.save["second"], 0)
	assert.Equal(t, 20, recorder.save["first"]+recorder.save["second"])
}

// TestBalancerNoService tests behavior when no servers are configured
func TestBalancerNoService(t *testing.T) {
	balancer := New(nil, false)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

// TestBalancerOneServerZeroWeight tests that zero-weight servers are ignored
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

	assert.Equal(t, 3, recorder.save["first"])
	assert.Equal(t, 0, recorder.save["second"]) // zero-weight server should not be added
}

// TestBalancerNoServiceUp tests behavior when all servers are marked down
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

// TestBalancerOneServerDown tests that down servers are excluded from selection
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

// TestBalancerDownThenUp tests server status transitions
func TestBalancerDownThenUp(t *testing.T) {
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
	// Both servers should get some traffic
	assert.Greater(t, recorder.save["first"], 0)
	assert.Greater(t, recorder.save["second"], 0)
	assert.Equal(t, 20, recorder.save["first"]+recorder.save["second"])
}

// TestBalancerPropagate tests status propagation to parent balancers
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

	// Test: Set all children of balancer1 to down, should propagate to top
	balancer1.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "first", false)
	balancer1.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "second", false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 4 {
		topBalancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Only balancer2 should receive traffic
	assert.Equal(t, 0, recorder.save["first"])
	assert.Equal(t, 0, recorder.save["second"])
	assert.Greater(t, recorder.save["third"]+recorder.save["fourth"], 0)
}

// TestBalancerAllServersZeroWeight tests that all zero-weight servers result in no available server
func TestBalancerAllServersZeroWeight(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), pointer(0), false)
	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), pointer(0), false)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

// TestBalancerFenced tests that fenced servers are excluded from selection
func TestBalancerFenced(t *testing.T) {
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

	// Only first server should receive traffic
	assert.Equal(t, 3, recorder.save["first"])
	assert.Equal(t, 0, recorder.save["second"])
}

// TestBalancerRegisterStatusUpdaterWithoutHealthCheck tests error when registering updater without health check
func TestBalancerRegisterStatusUpdaterWithoutHealthCheck(t *testing.T) {
	balancer := New(nil, false)

	err := balancer.RegisterStatusUpdater(func(up bool) {})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "healthCheck not enabled")
}

// TestBalancerAddServerMethod tests the AddServer method with dynamic.Server
func TestBalancerAddServerMethod(t *testing.T) {
	balancer := New(nil, false)

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "test")
		rw.WriteHeader(http.StatusOK)
	})

	server := dynamic.Server{
		Weight: pointer(2),
		Fenced: false,
	}

	balancer.AddServer("test", handler, server)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 1, recorder.save["test"])
}

// TestBalancerSticky tests sticky session support
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

	// First request should set cookie
	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	firstServer := recorder.Header().Get("server")
	assert.NotEmpty(t, firstServer)

	// Extract cookie from first response
	cookies := recorder.Result().Cookies()
	assert.NotEmpty(t, cookies)

	// Second request with cookie should hit same server
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	recorder2 := httptest.NewRecorder()
	balancer.ServeHTTP(recorder2, req)
	secondServer := recorder2.Header().Get("server")

	assert.Equal(t, firstServer, secondServer)
}

// TestBalancerStickyFallback tests that sticky sessions fallback to least-time when sticky server is down
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

	// Make initial request to establish sticky session with server1
	recorder1 := httptest.NewRecorder()
	balancer.ServeHTTP(recorder1, httptest.NewRequest(http.MethodGet, "/", nil))
	firstServer := recorder1.Header().Get("server")
	assert.NotEmpty(t, firstServer)

	// Extract cookie from first response
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

	// Establish sticky session with any server
	recorder1 := httptest.NewRecorder()
	balancer.ServeHTTP(recorder1, httptest.NewRequest(http.MethodGet, "/", nil))
	stickyServer := recorder1.Header().Get("server")
	assert.NotEmpty(t, stickyServer)

	cookies := recorder1.Result().Cookies()
	assert.NotEmpty(t, cookies)

	// Fence the sticky server (simulate graceful shutdown)
	balancer.handlersMu.Lock()
	balancer.fenced[stickyServer] = struct{}{}
	balancer.handlersMu.Unlock()

	// Existing sticky session should STILL work (graceful draining)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	recorder2 := httptest.NewRecorder()
	balancer.ServeHTTP(recorder2, req)
	assert.Equal(t, stickyServer, recorder2.Header().Get("server"))

	// But NEW requests should NOT go to the fenced server
	recorder3 := httptest.NewRecorder()
	balancer.ServeHTTP(recorder3, httptest.NewRequest(http.MethodGet, "/", nil))
	newServer := recorder3.Header().Get("server")
	assert.NotEqual(t, stickyServer, newServer)
	assert.NotEmpty(t, newServer)
}

// TestRingBufferBasic tests basic ring buffer functionality with few samples
func TestRingBufferBasic(t *testing.T) {
	handler := &namedHandler{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}),
		name:    "test",
		weight:  1,
	}

	// Test cold start - no samples
	avg := handler.getAvgResponseTime()
	assert.Equal(t, 0.0, avg)

	// Add one sample
	handler.updateResponseTime(10 * time.Millisecond)
	avg = handler.getAvgResponseTime()
	assert.Equal(t, 10.0, avg)

	// Add more samples
	handler.updateResponseTime(20 * time.Millisecond)
	handler.updateResponseTime(30 * time.Millisecond)
	avg = handler.getAvgResponseTime()
	assert.Equal(t, 20.0, avg) // (10 + 20 + 30) / 3 = 20
}

// TestRingBufferWraparound tests ring buffer behavior when it wraps around
func TestRingBufferWraparound(t *testing.T) {
	handler := &namedHandler{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}),
		name:    "test",
		weight:  1,
	}

	// Fill the buffer with 100 samples of 10ms each
	for i := 0; i < sampleSize; i++ {
		handler.updateResponseTime(10 * time.Millisecond)
	}
	avg := handler.getAvgResponseTime()
	assert.Equal(t, 10.0, avg)

	// Add one more sample (should replace oldest)
	handler.updateResponseTime(20 * time.Millisecond)
	avg = handler.getAvgResponseTime()
	// Sum: 99*10 + 1*20 = 1010, avg = 1010/100 = 10.1
	assert.Equal(t, 10.1, avg)

	// Add 10 more samples of 30ms
	for i := 0; i < 10; i++ {
		handler.updateResponseTime(30 * time.Millisecond)
	}
	avg = handler.getAvgResponseTime()
	// Sum: 89*10 + 1*20 + 10*30 = 890 + 20 + 300 = 1210, avg = 1210/100 = 12.1
	assert.Equal(t, 12.1, avg)
}

// TestRingBufferLarge tests ring buffer with many samples (> 100)
func TestRingBufferLarge(t *testing.T) {
	handler := &namedHandler{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}),
		name:    "test",
		weight:  1,
	}

	// Add 150 samples
	for i := 0; i < 150; i++ {
		handler.updateResponseTime(time.Duration(i+1) * time.Millisecond)
	}

	// Should only track last 100 samples: 51, 52, ..., 150
	// Sum = (51 + 150) * 100 / 2 = 10050
	// Avg = 10050 / 100 = 100.5
	avg := handler.getAvgResponseTime()
	assert.Equal(t, 100.5, avg)
}

// TestInflightCounter tests inflight request tracking
func TestInflightCounter(t *testing.T) {
	balancer := New(nil, false)

	var inflightAtRequest atomic.Int64

	balancer.Add("test", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Capture inflight count during request handling
		// Note: We need to access the handler through the balancer
		rw.Header().Set("server", "test")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Check that inflight count is 0 initially
	balancer.handlersMu.RLock()
	handler := balancer.handlers[0]
	balancer.handlersMu.RUnlock()
	assert.Equal(t, int64(0), handler.inflightCount.Load())

	// Make a request
	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	// After request completes, inflight should be back to 0
	assert.Equal(t, int64(0), handler.inflightCount.Load())

	// Store inflight count during request
	inflightAtRequest.Store(0)
	balancer.handlers[0].Handler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		inflightAtRequest.Store(balancer.handlers[0].inflightCount.Load())
		rw.WriteHeader(http.StatusOK)
	})

	balancer.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	// During request, inflight should have been 1
	assert.Equal(t, int64(1), inflightAtRequest.Load())
}

// TestConcurrentResponseTimeUpdates tests thread safety of response time updates
func TestConcurrentResponseTimeUpdates(t *testing.T) {
	handler := &namedHandler{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}),
		name:    "test",
		weight:  1,
	}

	// Concurrently update response times
	var wg sync.WaitGroup
	numGoroutines := 10
	updatesPerGoroutine := 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < updatesPerGoroutine; j++ {
				handler.updateResponseTime(time.Duration(id+1) * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Should have exactly 100 samples (buffer size)
	avg := handler.getAvgResponseTime()
	assert.Greater(t, avg, 0.0)

	// Verify sample count doesn't exceed buffer size
	handler.mu.RLock()
	assert.LessOrEqual(t, handler.sampleCount, sampleSize)
	handler.mu.RUnlock()
}

// TestConcurrentInflightTracking tests thread safety of inflight counter
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

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handler.inflightCount.Add(1)
			defer handler.inflightCount.Add(-1)

			// Track maximum inflight count
			current := handler.inflightCount.Load()
			for {
				max := maxInflight.Load()
				if current <= max || maxInflight.CompareAndSwap(max, current) {
					break
				}
			}

			time.Sleep(1 * time.Millisecond)
		}()
	}

	wg.Wait()

	// All requests completed, inflight should be 0
	assert.Equal(t, int64(0), handler.inflightCount.Load())
	// Max inflight should be > 1 (concurrent requests)
	assert.Greater(t, maxInflight.Load(), int64(1))
}

// TestTTFBMeasurement tests TTFB measurement accuracy
func TestTTFBMeasurement(t *testing.T) {
	balancer := New(nil, false)

	// Add server with known delay
	delay := 50 * time.Millisecond
	balancer.Add("slow", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(delay)
		rw.Header().Set("server", "slow")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Make multiple requests to build average
	for i := 0; i < 5; i++ {
		recorder := httptest.NewRecorder()
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Check that average response time is approximately the delay
	balancer.handlersMu.RLock()
	handler := balancer.handlers[0]
	balancer.handlersMu.RUnlock()

	avg := handler.getAvgResponseTime()
	// Allow 20ms tolerance for timing variations
	assert.InDelta(t, float64(delay.Milliseconds()), avg, 20.0)
}

// TestTTFBMeasurementMultipleServers tests TTFB tracking with different server speeds
func TestTTFBMeasurementMultipleServers(t *testing.T) {
	balancer := New(nil, false)

	// Add fast server
	balancer.Add("fast", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(10 * time.Millisecond)
		rw.Header().Set("server", "fast")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Add slow server
	balancer.Add("slow", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(100 * time.Millisecond)
		rw.Header().Set("server", "slow")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Make requests (round-robin will alternate)
	for i := 0; i < 10; i++ {
		recorder := httptest.NewRecorder()
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Check that fast server has lower average than slow server
	balancer.handlersMu.RLock()
	fastHandler := balancer.handlers[0]
	slowHandler := balancer.handlers[1]
	balancer.handlersMu.RUnlock()

	fastAvg := fastHandler.getAvgResponseTime()
	slowAvg := slowHandler.getAvgResponseTime()

	assert.Greater(t, slowAvg, fastAvg)
	assert.InDelta(t, 10.0, fastAvg, 20.0)
	assert.InDelta(t, 100.0, slowAvg, 20.0)
}

// TestResponseTrackerWriteHeader tests that WriteHeader is called correctly
func TestResponseTrackerWriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	startTime := time.Now()
	headerTime := startTime
	tracker := &responseTracker{
		ResponseWriter: recorder,
		headerTime:     &headerTime,
		headerWritten:  false,
	}

	// Simulate delay before writing header
	time.Sleep(10 * time.Millisecond)
	tracker.WriteHeader(http.StatusOK)

	// Check that header was written
	assert.True(t, tracker.headerWritten)
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Check that headerTime was updated
	assert.True(t, headerTime.After(startTime))
}

// TestZeroSamplesReturnsZero tests that getAvgResponseTime returns 0 when no samples
func TestZeroSamplesReturnsZero(t *testing.T) {
	handler := &namedHandler{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}),
		name:    "test",
		weight:  1,
	}

	avg := handler.getAvgResponseTime()
	assert.Equal(t, 0.0, avg)
}

// TestScoreCalculationWithWeights tests that weights are properly considered in score calculation
func TestScoreCalculationWithWeights(t *testing.T) {
	balancer := New(nil, false)

	// Add two servers with same response time but different weights
	// Server with higher weight should be preferred
	balancer.Add("weighted", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(50 * time.Millisecond)
		rw.Header().Set("server", "weighted")
		rw.WriteHeader(http.StatusOK)
	}), pointer(3), false) // Weight 3

	balancer.Add("normal", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(50 * time.Millisecond)
		rw.Header().Set("server", "normal")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false) // Weight 1

	// Make requests to build up response time averages
	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 10; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// After warmup, weighted server should get more requests
	// Score for weighted: (50 × (1 + 0)) / 3 = 16.67
	// Score for normal: (50 × (1 + 0)) / 1 = 50
	// Weighted server has lower score and should be preferred
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 30; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Weighted server should get significantly more traffic
	assert.Greater(t, recorder.save["weighted"], recorder.save["normal"])
	// Should be roughly 3:1 ratio or better for weighted server
	assert.Greater(t, recorder.save["weighted"], 20)
}

// TestScoreCalculationWithInflight tests that inflight requests are considered in score calculation
func TestScoreCalculationWithInflight(t *testing.T) {
	balancer := New(nil, false)

	// We'll manually control the inflight counters to test the score calculation
	// Add two servers with same response time
	balancer.Add("server1", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(10 * time.Millisecond)
		rw.Header().Set("server", "server1")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("server2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(10 * time.Millisecond)
		rw.Header().Set("server", "server2")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Build up response time averages for both servers
	for i := 0; i < 10; i++ {
		recorder := httptest.NewRecorder()
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Now manually set server1 to have high inflight count
	balancer.handlers[0].inflightCount.Store(5)

	// Make requests - they should prefer server2 because:
	// Score for server1: (10 × (1 + 5)) / 1 = 60
	// Score for server2: (10 × (1 + 0)) / 1 = 10
	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 20; i++ {
		// Manually increment/decrement to simulate the ServeHTTP behavior
		// but without actually going through full request flow
		server, _ := balancer.nextServer()
		if server.name == "server1" {
			recorder.save["server1"]++
		} else {
			recorder.save["server2"]++
		}
	}

	// Server2 should get all or most requests
	assert.Greater(t, recorder.save["server2"], 15)

	// Reset the inflight counter
	balancer.handlers[0].inflightCount.Store(0)
}

// TestScoreCalculationColdStart tests that new servers (0ms avg) get fair selection
func TestScoreCalculationColdStart(t *testing.T) {
	balancer := New(nil, false)

	// Add a warm server with established response time
	balancer.Add("warm", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(50 * time.Millisecond)
		rw.Header().Set("server", "warm")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Warm up the first server
	for i := 0; i < 10; i++ {
		recorder := httptest.NewRecorder()
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Now add a cold server (new, no response time data)
	balancer.Add("cold", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(10 * time.Millisecond) // Actually faster
		rw.Header().Set("server", "cold")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Cold server should get selected because:
	// Score for warm: (50 × (1 + 0)) / 1 = 50
	// Score for cold: (0 × (1 + 0)) / 1 = 0
	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 20; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Cold server should get all or most requests initially due to 0ms average
	assert.Greater(t, recorder.save["cold"], 10)

	// After cold server builds up its average, it should continue to get more traffic
	// because it's actually faster (10ms vs 50ms)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 20; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Greater(t, recorder.save["cold"], recorder.save["warm"])
}

// TestFastServerGetsMoreTraffic verifies that servers with lower response times
// receive proportionally more traffic in steady state (after cold start).
// This tests the core selection bias of the least-time algorithm.
func TestFastServerGetsMoreTraffic(t *testing.T) {
	balancer := New(nil, false)

	// Add two servers with different static response times
	balancer.Add("fast", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(20 * time.Millisecond)
		rw.Header().Set("server", "fast")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("slow", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(100 * time.Millisecond)
		rw.Header().Set("server", "slow")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Make requests: cold start phase → steady state convergence
	// Initial requests have random distribution, but algorithm converges
	// to prefer fast server as averages are established
	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 50; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Fast server should get significantly more traffic than slow server
	// Expecting at least 60% of traffic to fast server (30/50 requests)
	assert.Greater(t, recorder.save["fast"], recorder.save["slow"])
	assert.Greater(t, recorder.save["fast"], 30)
}

// TestTrafficShiftsWhenPerformanceDegrades verifies that the load balancer
// adapts to changing server performance by shifting traffic away from degraded servers.
// This tests the adaptive behavior - the core value proposition of least-time load balancing.
func TestTrafficShiftsWhenPerformanceDegrades(t *testing.T) {
	balancer := New(nil, false)

	// Use atomic to dynamically control server1's response time
	server1Delay := atomic.Int64{}
	server1Delay.Store(50) // Start with 50ms

	balancer.Add("server1", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Duration(server1Delay.Load()) * time.Millisecond)
		rw.Header().Set("server", "server1")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("server2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(50 * time.Millisecond) // Static 50ms
		rw.Header().Set("server", "server2")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Pre-fill ring buffers to eliminate cold start effects and ensure deterministic equal performance state
	balancer.handlersMu.RLock()
	for _, h := range balancer.handlers {
		h.mu.Lock()
		for i := 0; i < sampleSize; i++ {
			h.responseTimes[i] = 50.0
		}
		h.responseTimeSum = 50.0 * sampleSize
		h.sampleCount = sampleSize
		h.mu.Unlock()
	}
	balancer.handlersMu.RUnlock()

	// Phase 1: Both servers perform equally (50ms each)
	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 50; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// With equal performance and pre-filled buffers, distribution should be balanced via WRR tie-breaking
	total := recorder.save["server1"] + recorder.save["server2"]
	assert.Equal(t, 50, total)
	assert.Greater(t, recorder.save["server1"], 14) // At least 30% of traffic
	assert.Greater(t, recorder.save["server2"], 14) // At least 30% of traffic

	// Phase 2: server1 degrades (simulating GC pause, CPU spike, or network latency)
	server1Delay.Store(150) // Now 150ms (3x slower)

	// Make more requests to shift the moving average
	// Ring buffer has 100 samples, need significant new samples to shift average
	// server1's average will climb from ~50ms toward 150ms
	recorder2 := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 60; i++ {
		balancer.ServeHTTP(recorder2, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// Traffic should have shifted to server2 (the faster server)
	// server2 should get significantly more traffic (>70%)
	// Score for server1: (~100-150ms × 1) / 1 = 100-150 (as average climbs)
	// Score for server2: (50ms × 1) / 1 = 50
	total2 := recorder2.save["server1"] + recorder2.save["server2"]
	assert.Equal(t, 60, total2)
	assert.Greater(t, recorder2.save["server2"], recorder2.save["server1"])
	assert.Greater(t, recorder2.save["server2"], 40) // At least 66%% of traffic (40/60 requests)
}

// TestMultipleServersWithSameScore tests WRR tie-breaking when multiple servers have identical scores
func TestMultipleServersWithSameScore(t *testing.T) {
	balancer := New(nil, false)

	// Add three servers with identical response times and weights
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

	balancer.Add("server3", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(50 * time.Millisecond)
		rw.Header().Set("server", "server3")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	// Set all servers to identical response times to trigger tie-breaking
	for _, h := range balancer.handlers {
		h.mu.Lock()
		for i := 0; i < sampleSize; i++ {
			h.responseTimes[i] = 50.0
		}
		h.responseTimeSum = 50.0 * sampleSize
		h.sampleCount = sampleSize
		h.mu.Unlock()
	}

	// With all servers having identical scores, WRR tie-breaking should distribute fairly
	// Make enough requests to see distribution
	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 90; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	// All three servers should receive some traffic
	// With WRR and 90 requests, each should get approximately 30 requests
	assert.Greater(t, recorder.save["server1"], 0)
	assert.Greater(t, recorder.save["server2"], 0)
	assert.Greater(t, recorder.save["server3"], 0)

	// Total should be 90
	total := recorder.save["server1"] + recorder.save["server2"] + recorder.save["server3"]
	assert.Equal(t, 90, total)

	// With equal weights and sufficient pre-filled samples, distribution should be relatively balanced
	assert.InDelta(t, 30, recorder.save["server1"], 12)
	assert.InDelta(t, 30, recorder.save["server2"], 12)
	assert.InDelta(t, 30, recorder.save["server3"], 12)
}

// TestWRRTieBreakingWeightedDistribution tests weighted distribution among tied servers
func TestWRRTieBreakingWeightedDistribution(t *testing.T) {
	balancer := New(nil, false)

	// Add two servers with different weights
	// To create equal scores, response times must be proportional to weights
	// Use different sleep times to maintain the proportional response times
	balancer.Add("weighted", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(150 * time.Millisecond) // 3x longer due to 3x weight
		rw.Header().Set("server", "weighted")
		rw.WriteHeader(http.StatusOK)
	}), pointer(3), false) // Weight 3

	balancer.Add("normal", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		time.Sleep(50 * time.Millisecond)
		rw.Header().Set("server", "normal")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false) // Weight 1

	// Manually set response times proportional to weights to create equal scores
	// weighted: score = (150 * 1) / 3 = 50
	// normal: score = (50 * 1) / 1 = 50
	// Both scores are equal, so WRR tie-breaking will apply
	balancer.handlers[0].mu.Lock()
	for i := 0; i < sampleSize; i++ {
		balancer.handlers[0].responseTimes[i] = 150.0
	}
	balancer.handlers[0].responseTimeSum = 150.0 * sampleSize
	balancer.handlers[0].sampleCount = sampleSize
	balancer.handlers[0].mu.Unlock()

	balancer.handlers[1].mu.Lock()
	for i := 0; i < sampleSize; i++ {
		balancer.handlers[1].responseTimes[i] = 50.0
	}
	balancer.handlers[1].responseTimeSum = 50.0 * sampleSize
	balancer.handlers[1].sampleCount = sampleSize
	balancer.handlers[1].mu.Unlock()

	// Make requests - should use WRR for tie-breaking with 3:1 ratio
	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 80; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	total := recorder.save["weighted"] + recorder.save["normal"]
	assert.Equal(t, 80, total)

	// Weighted server should get approximately 75% of traffic (60/80)
	// Normal server should get approximately 25% of traffic (20/80)
	// With full buffer of consistent samples, distribution should follow the 3:1 weight ratio
	assert.Greater(t, recorder.save["weighted"], 45) // At least 56%
	assert.Greater(t, recorder.save["normal"], 10)   // At least 12.5%

	// More precise check: ratio should be approximately 3:1
	// Allow wider tolerance since 80 requests will partially overwrite the buffer
	ratio := float64(recorder.save["weighted"]) / float64(recorder.save["normal"])
	assert.InDelta(t, 3.0, ratio, 1.5) // 3:1 ratio with ±1.5 tolerance
}
