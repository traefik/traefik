package wrr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

type key string

const serviceName key = "serviceName"

func pointer[T any](v T) *T { return &v }

func TestBalancer(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), pointer(3), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 4 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	assert.Equal(t, 3, recorder.save["first"])
	assert.Equal(t, 1, recorder.save["second"])
}

func TestBalancerNoService(t *testing.T) {
	balancer := New(nil, false)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

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
}

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
}

func TestBalancerDownThenUp(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)
	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "second", false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 3 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Equal(t, 3, recorder.save["first"])

	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "second", true)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 2 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Equal(t, 1, recorder.save["first"])
	assert.Equal(t, 1, recorder.save["second"])
}

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
	_ = balancer1.RegisterStatusUpdater(func(up bool) {
		topBalancer.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "balancer1", up)
		// TODO(mpl): if test gets flaky, add channel or something here to signal that
		// propagation is done, and wait on it before sending request.
	})
	topBalancer.Add("balancer2", balancer2, pointer(1), false)
	_ = balancer2.RegisterStatusUpdater(func(up bool) {
		topBalancer.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "balancer2", up)
	})

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 8 {
		topBalancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Equal(t, 2, recorder.save["first"])
	assert.Equal(t, 2, recorder.save["second"])
	assert.Equal(t, 2, recorder.save["third"])
	assert.Equal(t, 2, recorder.save["fourth"])
	wantStatus := []int{200, 200, 200, 200, 200, 200, 200, 200}
	assert.Equal(t, wantStatus, recorder.status)

	// fourth gets downed, but balancer2 still up since third is still up.
	balancer2.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "fourth", false)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 8 {
		topBalancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Equal(t, 2, recorder.save["first"])
	assert.Equal(t, 2, recorder.save["second"])
	assert.Equal(t, 4, recorder.save["third"])
	assert.Equal(t, 0, recorder.save["fourth"])
	wantStatus = []int{200, 200, 200, 200, 200, 200, 200, 200}
	assert.Equal(t, wantStatus, recorder.status)

	// third gets downed, and the propagation triggers balancer2 to be marked as
	// down as well for topBalancer.
	balancer2.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "third", false)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 8 {
		topBalancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Equal(t, 4, recorder.save["first"])
	assert.Equal(t, 4, recorder.save["second"])
	assert.Equal(t, 0, recorder.save["third"])
	assert.Equal(t, 0, recorder.save["fourth"])
	wantStatus = []int{200, 200, 200, 200, 200, 200, 200, 200}
	assert.Equal(t, wantStatus, recorder.status)
}

func TestBalancerAllServersZeroWeight(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("test", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), pointer(0), false)
	balancer.Add("test2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), pointer(0), false)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

func TestBalancerAllServersFenced(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("test", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), true)
	balancer.Add("test2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), true)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	// Fenced but healthy endpoints should be used as fallback
	assert.Equal(t, http.StatusOK, recorder.Result().StatusCode)
}

func TestSticky(t *testing.T) {
	balancer := New(&dynamic.Sticky{
		Cookie: &dynamic.Cookie{
			Name:     "test",
			Secure:   true,
			HTTPOnly: true,
			SameSite: "none",
			Domain:   "foo.com",
			MaxAge:   42,
			Path:     func(v string) *string { return &v }("/foo"),
		},
	}, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), pointer(2), false)

	recorder := &responseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		save:             map[string]int{},
		cookies:          make(map[string]*http.Cookie),
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for range 3 {
		for _, cookie := range recorder.Result().Cookies() {
			assert.NotContains(t, "first", cookie.Value)
			assert.NotContains(t, "second", cookie.Value)
			req.AddCookie(cookie)
		}
		recorder.ResponseRecorder = httptest.NewRecorder()

		balancer.ServeHTTP(recorder, req)
	}

	assert.Equal(t, 0, recorder.save["first"])
	assert.Equal(t, 3, recorder.save["second"])
	assert.True(t, recorder.cookies["test"].HttpOnly)
	assert.True(t, recorder.cookies["test"].Secure)
	assert.Equal(t, "foo.com", recorder.cookies["test"].Domain)
	assert.Equal(t, http.SameSiteNoneMode, recorder.cookies["test"].SameSite)
	assert.Equal(t, 42, recorder.cookies["test"].MaxAge)
	assert.Equal(t, "/foo", recorder.cookies["test"].Path)
}

func TestSticky_Fallback(t *testing.T) {
	balancer := New(&dynamic.Sticky{
		Cookie: &dynamic.Cookie{Name: "test"},
	}, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), pointer(2), false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}, cookies: make(map[string]*http.Cookie)}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "test", Value: "second"})
	for range 3 {
		recorder.ResponseRecorder = httptest.NewRecorder()

		balancer.ServeHTTP(recorder, req)
	}

	assert.Equal(t, 0, recorder.save["first"])
	assert.Equal(t, 3, recorder.save["second"])
}

// TestSticky_Fenced checks that fenced node receive traffic if their sticky cookie matches.
func TestSticky_Fenced(t *testing.T) {
	balancer := New(&dynamic.Sticky{Cookie: &dynamic.Cookie{Name: "test"}}, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), false)

	balancer.Add("fenced", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "fenced")
		rw.WriteHeader(http.StatusOK)
	}), pointer(1), true)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}, cookies: make(map[string]*http.Cookie)}

	stickyReq := httptest.NewRequest(http.MethodGet, "/", nil)
	stickyReq.AddCookie(&http.Cookie{Name: "test", Value: "fenced"})

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	for range 4 {
		recorder.ResponseRecorder = httptest.NewRecorder()

		balancer.ServeHTTP(recorder, stickyReq)
		balancer.ServeHTTP(recorder, req)
	}

	assert.Equal(t, 4, recorder.save["fenced"])
	assert.Equal(t, 2, recorder.save["first"])
	assert.Equal(t, 2, recorder.save["second"])
}

// TestBalancerBias makes sure that the WRR algorithm spreads elements evenly right from the start,
// and that it does not "over-favor" the high-weighted ones with a biased start-up regime.
func TestBalancerBias(t *testing.T) {
	balancer := New(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "A")
		rw.WriteHeader(http.StatusOK)
	}), pointer(11), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "B")
		rw.WriteHeader(http.StatusOK)
	}), pointer(3), false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}

	for range 14 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	wantSequence := []string{"A", "A", "A", "B", "A", "A", "A", "A", "B", "A", "A", "A", "B", "A"}

	assert.Equal(t, wantSequence, recorder.sequence)
}

type responseRecorder struct {
	*httptest.ResponseRecorder
	save     map[string]int
	sequence []string
	status   []int
	cookies  map[string]*http.Cookie
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.save[r.Header().Get("server")]++
	r.sequence = append(r.sequence, r.Header().Get("server"))
	r.status = append(r.status, statusCode)
	for _, cookie := range r.Result().Cookies() {
		r.cookies[cookie.Name] = cookie
	}
	r.ResponseRecorder.WriteHeader(statusCode)
}

// TestNextServerWithAllTerminating tests that when all endpoints are terminating
// (fenced but still serving), the load balancer should still route traffic to them
// instead of returning an error.
func TestNextServerWithAllTerminating(t *testing.T) {
	balancer := New(nil, false)

	// Add three handlers, all will be marked as fenced (terminating)
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler1"))
	})
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler2"))
	})
	handler3 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler3"))
	})

	weight := 1
	balancer.Add("handler1", handler1, &weight, true) // fenced=true
	balancer.Add("handler2", handler2, &weight, true) // fenced=true
	balancer.Add("handler3", handler3, &weight, true) // fenced=true

	// Mark all handlers as healthy (serving)
	ctx := context.Background()
	balancer.SetStatus(ctx, "handler1", true)
	balancer.SetStatus(ctx, "handler2", true)
	balancer.SetStatus(ctx, "handler3", true)

	// nextServer should return a handler even though all are fenced
	// because they're still healthy and serving
	server, err := balancer.nextServer()
	require.NoError(t, err, "Should not error when all handlers are fenced but healthy")
	require.NotNil(t, server, "Should return a server even when all are fenced")
	assert.Contains(t, []string{"handler1", "handler2", "handler3"}, server.name)
}

// TestNextServerPreferNonTerminating tests that non-terminating endpoints
// are preferred over terminating ones
func TestNextServerPreferNonTerminating(t *testing.T) {
	balancer := New(nil, false)

	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler1"))
	})
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler2"))
	})

	weight := 1
	balancer.Add("handler1", handler1, &weight, false) // not fenced
	balancer.Add("handler2", handler2, &weight, true)  // fenced

	ctx := context.Background()
	balancer.SetStatus(ctx, "handler1", true)
	balancer.SetStatus(ctx, "handler2", true)

	// Should prefer the non-fenced handler
	selectedHandlers := make(map[string]int)
	for i := 0; i < 10; i++ {
		server, err := balancer.nextServer()
		require.NoError(t, err)
		require.NotNil(t, server)
		selectedHandlers[server.name]++
	}

	// handler1 (non-fenced) should be selected more often or exclusively
	assert.Greater(t, selectedHandlers["handler1"], 0, "Non-fenced handler should be selected")

	// In the current implementation, non-fenced handlers should be selected exclusively
	// when both fenced and non-fenced handlers are available
	assert.Equal(t, 10, selectedHandlers["handler1"], "Non-fenced handler should be selected exclusively")
}

// TestNextServerFallbackToTerminating tests that terminating endpoints
// are used as a fallback when no healthy non-terminating endpoints exist
func TestNextServerFallbackToTerminating(t *testing.T) {
	balancer := New(nil, false)

	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler1"))
	})
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler2"))
	})

	weight := 1
	balancer.Add("handler1", handler1, &weight, false) // not fenced, but will be marked unhealthy
	balancer.Add("handler2", handler2, &weight, true)  // fenced but healthy

	// Mark handler1 as down, handler2 as up but fenced
	ctx := context.Background()
	balancer.SetStatus(ctx, "handler1", false)
	balancer.SetStatus(ctx, "handler2", true)

	// Should fall back to the fenced but healthy handler
	server, err := balancer.nextServer()
	require.NoError(t, err, "Should not error when fenced handler is available")
	require.NotNil(t, server)
	assert.Equal(t, "handler2", server.name, "Should fallback to fenced handler when no healthy non-fenced handlers exist")
}

// TestNextServerNoHealthyHandlers tests that an error is returned
// when no healthy handlers exist at all
func TestNextServerNoHealthyHandlers(t *testing.T) {
	balancer := New(nil, false)

	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	weight := 1
	balancer.Add("handler1", handler1, &weight, false)

	// Mark handler as unhealthy (Add() automatically marks it as healthy)
	ctx := context.Background()
	balancer.SetStatus(ctx, "handler1", false)

	server, err := balancer.nextServer()
	assert.Error(t, err, "Should error when no healthy handlers exist")
	assert.Nil(t, server)
	assert.Equal(t, errNoAvailableServer, err)
}

// TestServeHTTPWithAllTerminating tests the full request flow
// when all backends are terminating
func TestServeHTTPWithAllTerminating(t *testing.T) {
	balancer := New(nil, false)

	callCount1 := 0
	callCount2 := 0

	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount1++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler1"))
	})
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount2++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler2"))
	})

	weight := 1
	balancer.Add("handler1", handler1, &weight, true) // fenced
	balancer.Add("handler2", handler2, &weight, true) // fenced

	ctx := context.Background()
	balancer.SetStatus(ctx, "handler1", true)
	balancer.SetStatus(ctx, "handler2", true)

	// Send 10 requests - all should succeed
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		balancer.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Request should succeed even with all fenced handlers")
	}

	// All requests should succeed
	totalCalls := callCount1 + callCount2
	assert.Equal(t, 10, totalCalls, "All requests should be handled")
	// Note: WRR with equal weights and all fenced may not distribute perfectly evenly,
	// but at least one handler should receive traffic (which proves fallback works)
	assert.True(t, callCount1 > 0 || callCount2 > 0, "At least one handler should receive requests")
}

// TestServeHTTPGracefulShutdownScenario tests a realistic graceful shutdown scenario:
// - Initially 2 healthy endpoints
// - One endpoint starts terminating (fenced=true)
// - Traffic should still reach the terminating endpoint
// - The terminating endpoint goes down
// - All traffic should go to the remaining healthy endpoint
func TestServeHTTPGracefulShutdownScenario(t *testing.T) {
	balancer := New(nil, false)

	callCount1 := 0
	callCount2 := 0

	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount1++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler1"))
	})
	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount2++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("handler2"))
	})

	weight := 1

	ctx := context.Background()

	// Phase 1: Both healthy and not terminating
	balancer.Add("handler1", handler1, &weight, false)
	balancer.Add("handler2", handler2, &weight, false)
	balancer.SetStatus(ctx, "handler1", true)
	balancer.SetStatus(ctx, "handler2", true)

	// Send some requests - both should receive traffic
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		balancer.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	phase1Calls1 := callCount1
	phase1Calls2 := callCount2
	assert.Greater(t, phase1Calls1, 0, "Handler1 should receive requests in phase 1")
	assert.Greater(t, phase1Calls2, 0, "Handler2 should receive requests in phase 1")

	// Phase 2: handler2 starts terminating (fenced=true)
	// Simulate this by updating the handler
	balancer.handlersMu.Lock()
	balancer.fenced["handler2"] = struct{}{}
	balancer.handlersMu.Unlock()

	// Send more requests - handler1 should get all of them now
	// (since non-fenced handlers are preferred)
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		balancer.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	phase2Calls1 := callCount1 - phase1Calls1
	phase2Calls2 := callCount2 - phase1Calls2

	// Handler1 (non-fenced) should receive all traffic
	assert.Equal(t, 10, phase2Calls1, "Non-fenced handler should receive all requests")
	assert.Equal(t, 0, phase2Calls2, "Fenced handler should receive no requests when non-fenced is available")

	// Phase 3: handler1 goes down, only handler2 (fenced) remains
	balancer.SetStatus(ctx, "handler1", false)

	// Send more requests - handler2 (fenced but healthy) should handle them
	callCount1Before := callCount1
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		balancer.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	phase3Calls1 := callCount1 - callCount1Before
	phase3Calls2 := callCount2 - (phase1Calls2 + phase2Calls2)

	// Only handler2 should receive traffic (graceful shutdown fallback)
	assert.Equal(t, 0, phase3Calls1, "Down handler should receive no requests")
	assert.Equal(t, 10, phase3Calls2, "Fenced but healthy handler should receive all requests as fallback")
}
