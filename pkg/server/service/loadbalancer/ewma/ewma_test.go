package ewma

import (
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

// TestAddServerIdempotency ensures that calling AddServer multiple times with the same
// server name does not create duplicate handlers or EWMA states.
func TestAddServerIdempotency(t *testing.T) {
	b := New(nil, false)
	name := "dup"

	for i := 0; i < 3; i++ {
		b.AddServer(name, dummyHTTPHandler, dynamic.Server{})
	}

	var count int
	for _, h := range b.handlers {
		if h.name == name {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 handler named %q, got %d", name, count)
	}

	registryMu.RLock()
	defer registryMu.RUnlock()
	if len(ewmaRegistry) != 1 {
		t.Errorf("expected exactly 1 entry in ewmaRegistry, got %d", len(ewmaRegistry))
	}
	if _, ok := ewmaRegistry[name]; !ok {
		t.Errorf("expected ewmaRegistry to contain key %q", name)
	}
}

// TestBalancerPrefersFastOverSlow verifies that the balancer prefers the fast server.
func TestBalancerPrefersFastOverSlow(t *testing.T) {
	b := New(nil, false)
	b.rand = fixedRand(4234)

	fast := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	b.AddServer("fast", fast, dynamic.Server{})
	b.AddServer("slow", slow, dynamic.Server{})

	for _, h := range b.handlers {
		b.updateEWMA(h, 50*time.Millisecond)
	}

	var wg sync.WaitGroup
	var fastCnt, slowCnt int32

	numRequests := 2000
	concurrency := 100
	wg.Add(numRequests)
	semaphore := make(chan struct{}, concurrency)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			srv, err := b.nextServer()
			if err != nil {
				return
			}
			if srv.name == "fast" {
				atomic.AddInt32(&fastCnt, 1)
				now := time.Now()
				time.Sleep(10 * time.Millisecond)
				b.updateEWMA(srv, time.Since(now))
			} else {
				atomic.AddInt32(&slowCnt, 1)
				now := time.Now()
				time.Sleep(50 * time.Millisecond)
				b.updateEWMA(srv, time.Since(now))
			}
		}()
	}
	wg.Wait()

	fastPct := float64(atomic.LoadInt32(&fastCnt)) / float64(numRequests) * 100
	slowPct := float64(atomic.LoadInt32(&slowCnt)) / float64(numRequests) * 100

	if fastPct <= slowPct {
		t.Errorf("expected fast server to be chosen more often: fast=%.1f%%, slow=%.1f%%", fastPct, slowPct)
	}
}

// TestSyncSlowStart verifies that Sync method implements slow-start correctly.
func TestSyncSlowStart(t *testing.T) {
	b := New(nil, false)
	b.AddServer("s1", nil, dynamic.Server{})
	b.AddServer("s2", nil, dynamic.Server{})
	b.AddServer("s3", nil, dynamic.Server{})

	now := time.Now()
	for _, h := range b.handlers {
		switch h.name {
		case "s1":
			h.state.ewmaValue.Store(math.Float64bits(10.0))
			h.state.lastUpdated.Store(uint64(now.Add(-1 * time.Second).UnixNano()))
		case "s2":
			h.state.ewmaValue.Store(math.Float64bits(20.0))
			h.state.lastUpdated.Store(uint64(now.Add(-2 * time.Second).UnixNano()))
		case "s3":
			h.state.ewmaValue.Store(math.Float64bits(15.0))
			h.state.lastUpdated.Store(uint64(now.Add(-2 * time.Second).UnixNano()))
		}
	}

	for _, h := range b.handlers {
		got := math.Float64frombits(h.state.ewmaValue.Load())
		var want float64
		switch h.name {
		case "s1":
			want = 10.0
		case "s2":
			want = 20.0
		case "s3":
			want = 15.0
		}
		if got != want {
			t.Errorf("handler %s: EWMA = %v; want %v", h.name, got, want)
		}
	}
}

// TestCleanupRegistryRemovesOldStates verifies that cleanupRegistry removes old states.
func TestCleanupRegistryRemovesOldStates(t *testing.T) {
	name := "old"
	st := &serverState{}
	st.ewmaValue.Store(math.Float64bits(1.0))
	st.lastUpdated.Store(uint64(time.Now().Add(-2 * ewmaStateTTL).UnixNano()))

	registryMu.Lock()
	ewmaRegistry[name] = st
	nonexistent := "fresh"
	eFresh := &serverState{}
	eFresh.ewmaValue.Store(math.Float64bits(1.0))
	eFresh.lastUpdated.Store(uint64(time.Now().UnixNano()))
	ewmaRegistry[nonexistent] = eFresh
	registryMu.Unlock()

	b := New(nil, false)
	b.cleanupRegistry()

	registryMu.RLock()
	defer registryMu.RUnlock()
	if _, ok := ewmaRegistry[name]; ok {
		t.Error("expected old state to be removed by cleanupRegistry")
	}
	if _, ok := ewmaRegistry[nonexistent]; !ok {
		t.Error("expected fresh state to remain after cleanupRegistry")
	}
}

// TestCleanupRegistryConcurrently ensures that cleanupRegistry can safely run
// concurrently with registry updates and that fresh states are never removed.
func TestCleanupRegistryConcurrently(t *testing.T) {
	b := New(nil, false)

	registryMu.Lock()
	ewmaRegistry = make(map[string]*serverState)
	now := time.Now().UnixNano()

	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("old-%d", i)
		st := &serverState{}
		st.ewmaValue.Store(math.Float64bits(1.0))
		st.lastUpdated.Store(uint64(now - 2*ewmaStateTTL.Nanoseconds()))
		ewmaRegistry[name] = st
	}
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("fresh-%d", i)
		st := &serverState{}
		st.ewmaValue.Store(math.Float64bits(1.0))
		st.lastUpdated.Store(uint64(now))
		ewmaRegistry[name] = st
	}
	registryMu.Unlock()

	var wg sync.WaitGroup
	numWorkers := 10
	opsPerWorker := 100

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			names := []string{}
			for j := 0; j < opsPerWorker; j++ {
				b.cleanupRegistry()

				registryMu.RLock()
				for k := range ewmaRegistry {
					names = append(names, k)
				}
				registryMu.RUnlock()
				_ = b.computeAverageEWMA(names)
				names = names[:0]
			}
		}(w)
	}

	wg.Wait()

	registryMu.RLock()
	defer registryMu.RUnlock()
	for i := 0; i < 5; i++ {
		oldName := fmt.Sprintf("old-%d", i)
		if _, ok := ewmaRegistry[oldName]; ok {
			t.Errorf("expected old entry %q to be removed", oldName)
		}
		freshName := fmt.Sprintf("fresh-%d", i)
		if _, ok := ewmaRegistry[freshName]; !ok {
			t.Errorf("expected fresh entry %q to remain", freshName)
		}
	}
}

// TestConcurrency_LongRunningCleanup ensures that cleanupRegistry can safely run
// concurrently with nextServer and updateEWMA, even if registryMu is held for a long time,
// without causing deadlocks.
func TestConcurrency_LongRunningCleanup(t *testing.T) {
	b := New(nil, false)
	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("h%d", i)
		b.AddServer(name, dummyHTTPHandler, dynamic.Server{})
	}
	for _, h := range b.handlers {
		b.updateEWMA(h, 50*time.Millisecond)
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	done := make(chan struct{})

	go func() {
		if _, err := b.nextServer(); err != nil {
			t.Errorf("nextServer unexpectedly errored: %v", err)
		}
		for _, h := range b.handlers {
			b.updateEWMA(h, 10*time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-done:
		// test ok
	case <-time.After(1 * time.Second):
		t.Fatal("operations deadlocked during long-running cleanup")
	}
}

// TestFilterStatusAndFenced ensures that down or fenced handlers are skipped.
func TestFilterStatusAndFenced(t *testing.T) {
	b := New(nil, false)

	b.AddServer("h1", nil, dynamic.Server{})
	b.AddServer("h2", nil, dynamic.Server{Fenced: true})
	b.AddServer("h3", nil, dynamic.Server{})

	b.SetStatus(context.Background(), "h3", false)

	if h, err := b.nextServer(); err == nil {
		if h.name != "h1" {
			t.Fatal("should only get h1 server")
		}
	}

	b.SetStatus(context.Background(), "h3", true)

	successes := 0
	for attempts := 0; attempts < 20; attempts++ {
		srv, err := b.nextServer()
		if err != nil {
			continue
		}
		successes++
		if srv.name == "h2" {
			t.Errorf("fenced handler h2 should not be selected")
		}
	}
	if successes == 0 {
		t.Fatal("expected at least one successful nextServer after bringing h3 up")
	}
}

// TestStatusUpdater ensures RegisterStatusUpdater and SetStatus invoke callbacks.
func TestStatusUpdater(t *testing.T) {
	b := New(nil, true)
	b.SetStatus(context.Background(), "x1", false)
	b.SetStatus(context.Background(), "x2", true)
	b.SetStatus(context.Background(), "x3", true)
	b.SetStatus(context.Background(), "x4", false)

	if len(b.status) < 2 {
		t.Errorf("expected 2 callbacks, got %v", b.status)
	}
}

// TestStickySession verifies that sticky sessions route to the same server.
func TestStickySession(t *testing.T) {
	cookieCfg := &dynamic.Sticky{Cookie: &dynamic.Cookie{Name: "sessionID"}}
	b := New(cookieCfg, false)

	h1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("h1")) })
	h2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("h2")) })
	b.AddServer("h1", h1, dynamic.Server{})
	b.AddServer("h2", h2, dynamic.Server{})

	rw1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/", nil)
	b.ServeHTTP(rw1, req1)
	resp1 := rw1.Result()
	body1, _ := io.ReadAll(resp1.Body)

	cookies := resp1.Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected sticky cookie to be set")
	}
	cookie := cookies[0]

	rw2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.AddCookie(cookie)
	b.ServeHTTP(rw2, req2)
	body2, _ := io.ReadAll(rw2.Result().Body)

	if string(body2) != string(body1) {
		t.Errorf("expected response from same server '%s', got '%s'", string(body1), string(body2))
	}
}

// TestStickySessionWithMultipleServers ensures that when the originally sticky
// server goes down, subsequent requests with the old cookie are routed via P2C
// and the sticky cookie is updated to the new server.
func TestStickySessionWithMultipleServers(t *testing.T) {
	cookieCfg := &dynamic.Sticky{Cookie: &dynamic.Cookie{Name: "sessionID"}}
	b := New(cookieCfg, false)

	h1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("h1"))
	})
	h2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("h2"))
	})

	b.AddServer("h1", h1, dynamic.Server{})
	b.AddServer("h2", h2, dynamic.Server{})

	rw1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/", nil)
	b.ServeHTTP(rw1, req1)
	resp1 := rw1.Result()
	body1, _ := io.ReadAll(resp1.Body)
	if string(body1) != "h1" {
		t.Fatalf("expected first response from h1, got %q", body1)
	}
	cookies1 := resp1.Cookies()
	if len(cookies1) == 0 {
		t.Fatal("expected sticky cookie to be set on first response")
	}
	originalCookie := cookies1[0]

	b.SetStatus(context.Background(), "h1", false)

	rw2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.AddCookie(originalCookie)
	b.ServeHTTP(rw2, req2)
	resp2 := rw2.Result()
	body2, _ := io.ReadAll(resp2.Body)
	if string(body2) == "h1" {
		t.Errorf("expected fallback from h1, but got %q", body2)
	}
	if string(body2) != "h2" {
		t.Errorf("expected fallback to h2, got %q", body2)
	}
	cookies2 := resp2.Cookies()
	if len(cookies2) == 0 {
		t.Fatal("expected sticky cookie to be rewritten on fallback")
	}
	newCookie := cookies2[0]
	if newCookie.Value == originalCookie.Value {
		t.Errorf("expected cookie value to change after fallback, still %q", newCookie.Value)
	}
}

// TestComputeEWMA verifies computeEWMA in boundary conditions.
func TestComputeEWMA(t *testing.T) {
	τ := 100.0 // decay constant
	prev := 50.0
	latency := 20.0

	// dt = 0 → result == prevEWMA
	if got := computeEWMA(prev, latency, 0, τ); got != prev {
		t.Errorf("dt=0: expected %v, got %v", prev, got)
	}

	// very large dt → result → latency
	largeDT := τ * 1e6
	if got := computeEWMA(prev, latency, largeDT, τ); math.Abs(got-latency) > 1e-6 {
		t.Errorf("large dt: expected ≈%v, got %v", latency, got)
	}

	// latency = 0 → exponential decay toward 0
	if got := computeEWMA(prev, 0, τ, τ); got >= prev {
		t.Errorf("latency=0: expected <%v, got %v", prev, got)
	}
}

// TestUpdateEWMAFirstMeasurement verifies updateEWMA sets EWMA == observed latency when state is fresh.
func TestUpdateEWMAFirstMeasurement(t *testing.T) {
	b := New(nil, false)
	h := &namedHandler{name: "x", state: &serverState{}}

	h.state.lastUpdated.Store(0)
	d := 123 * time.Millisecond
	b.updateEWMA(h, d)

	got := math.Float64frombits(h.state.ewmaValue.Load())
	want := float64(d.Seconds())
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("first update: expected EWMA=%v, got %v", want, got)
	}
}

// TestComputeAverageEWMA verifies that computeAverageEWMA returns the correct average of existing
// EWMA values in the registry, and falls back to initialEwma when no states are present.
func TestComputeAverageEWMA(t *testing.T) {
	b := New(nil, false)

	registryMu.Lock()
	ewmaRegistry = map[string]*serverState{}
	now := time.Now().UnixNano()
	stA := &serverState{}
	stA.ewmaValue.Store(math.Float64bits(10.0))
	stA.lastUpdated.Store(uint64(now))
	stB := &serverState{}
	stB.ewmaValue.Store(math.Float64bits(20.0))
	stB.lastUpdated.Store(uint64(now))
	ewmaRegistry["a"] = stA
	ewmaRegistry["b"] = stB
	registryMu.Unlock()

	avg := b.computeAverageEWMA([]string{"a", "b"})
	if avg != 15.0 {
		t.Errorf("expected average EWMA 15.0, got %v", avg)
	}

	registryMu.Lock()
	ewmaRegistry = map[string]*serverState{}
	registryMu.Unlock()

	avgEmpty := b.computeAverageEWMA([]string{"a", "b"})
	if avgEmpty != b.initialEwmaSeconds {
		t.Errorf("expected initial EWMA %v when registry empty, got %v", b.initialEwmaSeconds, avgEmpty)
	}
}

// TestConcurrentAccess ensures no data races when AddServer, nextServer and updateEWMA
// are called concurrently.
func TestConcurrentAccess(t *testing.T) {
	b := New(nil, false)

	for i := 0; i < 3; i++ {
		name := fmt.Sprintf("init%d", i)
		b.AddServer(name, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), dynamic.Server{})
	}

	for _, h := range b.handlers {
		b.updateEWMA(h, 50*time.Millisecond)
	}

	var wg sync.WaitGroup
	numWorkers := 50
	opsPerWorker := 500

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			rnd := rand.New(rand.NewSource(int64(id)))
			for j := 0; j < opsPerWorker; j++ {
				srv, err := b.nextServer()
				if err == nil {
					delay := time.Duration(rnd.Intn(100)) * time.Millisecond
					b.updateEWMA(srv, delay)
				}
				if j%100 == 0 {
					name := fmt.Sprintf("dyn-%d-%d", id, j)
					b.AddServer(name, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), dynamic.Server{})
				}
			}
		}(i)
	}

	wg.Wait()
}

// TestStatusFlapping verifies that SetStatus correctly removes and restores
// a server in healthyHandlers.
func TestStatusFlapping(t *testing.T) {
	b := New(nil, false)

	h1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	h2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	b.AddServer("h1", h1, dynamic.Server{})
	b.AddServer("h2", h2, dynamic.Server{})
	for _, h := range b.handlers {
		b.updateEWMA(h, 10*time.Millisecond)
	}

	if len(b.healthyHandlers) != 2 {
		t.Fatalf("expected 2 healthy handlers, got %d", len(b.healthyHandlers))
	}

	b.SetStatus(context.Background(), "h2", false)
	if len(b.healthyHandlers) != 1 || b.healthyHandlers[0].name != "h1" {
		t.Fatalf("after h2 down, expected only h1 healthy, got %v",
			func() []string {
				ns := []string{}
				for _, hn := range b.healthyHandlers {
					ns = append(ns, hn.name)
				}
				return ns
			}())
	}

	b.SetStatus(context.Background(), "h2", true)
	if len(b.healthyHandlers) != 2 {
		t.Fatalf("after h2 up, expected 2 healthy handlers, got %d", len(b.healthyHandlers))
	}
	seen := map[string]bool{}
	for _, nh := range b.healthyHandlers {
		seen[nh.name] = true
	}
	if !seen["h2"] {
		t.Errorf("expected h2 to be healthy after restore, healthyHandlers=%v",
			func() []string {
				ns := []string{}
				for _, hn := range b.healthyHandlers {
					ns = append(ns, hn.name)
				}
				return ns
			}())
	}
}

func TestNextServer_ShufflesHealthyHandlers(t *testing.T) {
	b := &Balancer{
		decayEwmaSeconds:   1,
		initialEwmaSeconds: 0.1,
	}
	b.rand = rand.New(rand.NewSource(42))

	names := []string{"A", "B", "C", "D"}
	for _, n := range names {
		st := &serverState{}
		st.ewmaValue.Store(math.Float64bits(0))
		st.lastUpdated.Store(uint64(time.Now().UnixNano()))
		nh := &namedHandler{name: n, state: st}
		b.handlers = append(b.handlers, nh)
		b.healthyHandlers = append(b.healthyHandlers, nh)
	}

	initial := snapshotNames(b.healthyHandlers)

	for i := 0; i < 10; i++ {
		_, err := b.nextServer()
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
		after := snapshotNames(b.healthyHandlers)
		if !reflect.DeepEqual(after, initial) {
			return
		}
	}

	t.Errorf("healthyHandlers never shuffled; always %v", initial)
}

// dummyHTTPHandler is a no-op HTTP handler for benchmarks
var dummyHTTPHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

// BenchmarkNextServer measures speed of selecting servers
func BenchmarkNextServer(bm *testing.B) {
	bal := prepareBalancerWithRand(100, 123)
	bm.ResetTimer()
	for i := 0; i < bm.N; i++ {
		_, err := bal.nextServer()
		if err != nil {
			bm.Fatalf("nextServer error: %v", err)
		}
	}
}

// BenchmarkUpdateEWMA measures speed of EWMA update
func BenchmarkUpdateEWMA(bm *testing.B) {
	bal := prepareBalancerWithRand(1, 456)
	h := bal.handlers[0]
	bm.ResetTimer()
	for i := 0; i < bm.N; i++ {
		bal.updateEWMA(h, time.Duration((i%100)+1)*time.Millisecond)
	}
}

// fixedRand returns a deterministic rand.Rand for reproducible tests
func fixedRand(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

// prepareBalancerWithRand initializes a balancer with numHandlers and fixed rand source
func prepareBalancerWithRand(numHandlers int, seed int64) *Balancer {
	b := New(nil, false)
	b.rand = fixedRand(seed)
	for i := 0; i < numHandlers; i++ {
		name := fmt.Sprintf("h%04d", i)
		b.AddServer(name, dummyHTTPHandler, dynamic.Server{})
	}
	for _, h := range b.handlers {
		b.updateEWMA(h, time.Duration(seed%50+1)*time.Millisecond)
	}
	return b
}

func snapshotNames(hs []*namedHandler) []string {
	names := make([]string, len(hs))
	for i, h := range hs {
		names[i] = h.name
	}
	return names
}
