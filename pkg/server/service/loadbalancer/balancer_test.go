package loadbalancer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestBalancer(t *testing.T) {
	balancer := NewWRR(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(3))

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 4; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	assert.Equal(t, 3, recorder.save["first"])
	assert.Equal(t, 1, recorder.save["second"])
}

func TestBalancerNoService(t *testing.T) {
	balancer := NewWRR(nil, false)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

func TestBalancerOneServerZeroWeight(t *testing.T) {
	balancer := NewWRR(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), Int(0))

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 3; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	assert.Equal(t, 3, recorder.save["first"])
}

type key string

const serviceName key = "serviceName"

func TestBalancerNoServiceUp(t *testing.T) {
	balancer := NewWRR(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}), Int(1))

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}), Int(1))

	balancer.SetStatus(context.WithValue(context.Background(), serviceName, "parent"), "first", false)
	balancer.SetStatus(context.WithValue(context.Background(), serviceName, "parent"), "second", false)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

func TestBalancerOneServerDown(t *testing.T) {
	balancer := NewWRR(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}), Int(1))
	balancer.SetStatus(context.WithValue(context.Background(), serviceName, "parent"), "second", false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 3; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	assert.Equal(t, 3, recorder.save["first"])
}

func TestBalancerDownThenUp(t *testing.T) {
	balancer := NewWRR(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))
	balancer.SetStatus(context.WithValue(context.Background(), serviceName, "parent"), "second", false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 3; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Equal(t, 3, recorder.save["first"])

	balancer.SetStatus(context.WithValue(context.Background(), serviceName, "parent"), "second", true)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 2; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Equal(t, 1, recorder.save["first"])
	assert.Equal(t, 1, recorder.save["second"])
}

func TestBalancerPropagate(t *testing.T) {
	balancer1 := NewWRR(nil, true)

	balancer1.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))
	balancer1.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))

	balancer2 := NewWRR(nil, true)
	balancer2.Add("third", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "third")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))
	balancer2.Add("fourth", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "fourth")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))

	topBalancer := NewWRR(nil, true)
	topBalancer.Add("balancer1", balancer1, Int(1))
	_ = balancer1.RegisterStatusUpdater(func(up bool) {
		topBalancer.SetStatus(context.WithValue(context.Background(), serviceName, "top"), "balancer1", up)
		// TODO(mpl): if test gets flaky, add channel or something here to signal that
		// propagation is done, and wait on it before sending request.
	})
	topBalancer.Add("balancer2", balancer2, Int(1))
	_ = balancer2.RegisterStatusUpdater(func(up bool) {
		topBalancer.SetStatus(context.WithValue(context.Background(), serviceName, "top"), "balancer2", up)
	})

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 8; i++ {
		topBalancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Equal(t, 2, recorder.save["first"])
	assert.Equal(t, 2, recorder.save["second"])
	assert.Equal(t, 2, recorder.save["third"])
	assert.Equal(t, 2, recorder.save["fourth"])
	wantStatus := []int{200, 200, 200, 200, 200, 200, 200, 200}
	assert.Equal(t, wantStatus, recorder.status)

	// fourth gets downed, but balancer2 still up since third is still up.
	balancer2.SetStatus(context.WithValue(context.Background(), serviceName, "top"), "fourth", false)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 8; i++ {
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
	balancer2.SetStatus(context.WithValue(context.Background(), serviceName, "top"), "third", false)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 8; i++ {
		topBalancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Equal(t, 4, recorder.save["first"])
	assert.Equal(t, 4, recorder.save["second"])
	assert.Equal(t, 0, recorder.save["third"])
	assert.Equal(t, 0, recorder.save["fourth"])
	wantStatus = []int{200, 200, 200, 200, 200, 200, 200, 200}
	assert.Equal(t, wantStatus, recorder.status)
}

func TestWRRBalancerAllServersZeroWeight(t *testing.T) {
	balancer := NewWRR(nil, false)

	balancer.Add("test", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), Int(0))
	balancer.Add("test2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), Int(0))

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

func TestSticky(t *testing.T) {
	balancers := []*Balancer{
		NewWRR(&dynamic.Sticky{
			Cookie: &dynamic.Cookie{
				Name:     "test",
				Secure:   true,
				HTTPOnly: true,
				SameSite: "none",
				MaxAge:   42,
			},
		}, false),
		NewP2C(&dynamic.Sticky{
			Cookie: &dynamic.Cookie{
				Name:     "test",
				Secure:   true,
				HTTPOnly: true,
				SameSite: "none",
				MaxAge:   42,
			},
		}, false),
	}

	// we need to make sure second is chosen
	balancers[1].strategy.(*strategyPowerOfTwoChoices).rand = &mockRand{vals: []int{1, 0}}

	for _, balancer := range balancers {
		t.Run(balancer.strategy.name(), func(t *testing.T) {

			balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("server", "first")
				rw.WriteHeader(http.StatusOK)
			}), Int(1))

			balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("server", "second")
				rw.WriteHeader(http.StatusOK)
			}), Int(2))

			recorder := &responseRecorder{
				ResponseRecorder: httptest.NewRecorder(),
				save:             map[string]int{},
				cookies:          make(map[string]*http.Cookie),
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for i := 0; i < 3; i++ {
				for _, cookie := range recorder.Result().Cookies() {
					assert.NotContains(t, "test=first", cookie.Value)
					assert.NotContains(t, "test=second", cookie.Value)
					req.AddCookie(cookie)
				}
				recorder.ResponseRecorder = httptest.NewRecorder()

				balancer.ServeHTTP(recorder, req)
			}

			assert.Equal(t, 0, recorder.save["first"])
			assert.Equal(t, 3, recorder.save["second"])
			assert.True(t, recorder.cookies["test"].HttpOnly)
			assert.True(t, recorder.cookies["test"].Secure)
			assert.Equal(t, http.SameSiteNoneMode, recorder.cookies["test"].SameSite)
			assert.Equal(t, 42, recorder.cookies["test"].MaxAge)
		})
	}
}

func TestSticky_FallBack(t *testing.T) {
	balancers := []*Balancer{
		NewWRR(&dynamic.Sticky{
			Cookie: &dynamic.Cookie{Name: "test"},
		}, false),
		NewP2C(&dynamic.Sticky{
			Cookie: &dynamic.Cookie{Name: "test"},
		}, false),
	}

	for _, balancer := range balancers {
		t.Run(balancer.strategy.name(), func(t *testing.T) {
			balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("server", "first")
				rw.WriteHeader(http.StatusOK)
			}), Int(1))

			balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set("server", "second")
				rw.WriteHeader(http.StatusOK)
			}), Int(2))

			recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.AddCookie(&http.Cookie{Name: "test", Value: "second"})
			for i := 0; i < 3; i++ {
				recorder.ResponseRecorder = httptest.NewRecorder()

				balancer.ServeHTTP(recorder, req)
			}

			assert.Equal(t, 0, recorder.save["first"])
			assert.Equal(t, 3, recorder.save["second"])
		})
	}
}

// TestBalancerBias makes sure that the WRR algorithm spreads elements evenly right from the start,
// and that it does not "over-favor" the high-weighted ones with a biased start-up regime.
func TestBalancerBias(t *testing.T) {
	balancer := NewWRR(nil, false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "A")
		rw.WriteHeader(http.StatusOK)
	}), Int(11))

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "B")
		rw.WriteHeader(http.StatusOK)
	}), Int(3))

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}

	for i := 0; i < 14; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	wantSequence := []string{"A", "A", "A", "B", "A", "A", "A", "A", "B", "A", "A", "A", "B", "A"}

	assert.Equal(t, wantSequence, recorder.sequence)
}

func Int(v int) *int { return &v }

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

func testHandler(name string, weight float64, inflight int) *namedHandler {
	h := &namedHandler{
		name: name,
	}
	h.inflight.Store(int64(inflight))
	h.weight = weight
	h.Handler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("server", name)
		rw.WriteHeader(http.StatusOK)
	})
	return h
}

func TestStrategies(t *testing.T) {
	newStrategies := []func() strategy{
		newStrategyWRR,
		newStrategyP2C,
	}

	for _, s := range newStrategies {
		t.Run(s().name(), testStrategy(s))
	}
}

func testStrategy(newStrategy func() strategy) func(t *testing.T) {
	return func(t *testing.T) {
		t.Run("OneHealthyBackend", testStrategyOneHealthyBackend(newStrategy()))
		t.Run("TwoHealthyBackends", testStrategyTwoHealthyBackends(newStrategy()))
		t.Run("OneHealthyOneUnhealthy", testStrategyOneHealthyOneUnhealthy(newStrategy()))
		t.Run("OneHostDownThenUp", testStrategyOneHostDownThenUp(newStrategy()))
	}
}

func testStrategyOneHealthyBackend(strategy strategy) func(t *testing.T) {
	return func(t *testing.T) {
		strategy.add(testHandler("A", 1, 0))

		healthy := map[string]struct{}{"A": {}}

		recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}

		const requests = 10
		for i := 0; i < requests; i++ {
			strategy.nextServer(healthy).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
		}

		assert.Equal(t, requests, recorder.save["A"], "A should have been hit with all requests")
	}
}

func testStrategyTwoHealthyBackends(strategy strategy) func(t *testing.T) {
	return func(t *testing.T) {
		strategy.add(testHandler("A", 1, 0))
		strategy.add(testHandler("B", 1, 0))

		healthy := map[string]struct{}{"A": {}, "B": {}}

		recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}

		for i := 0; i < 100; i++ {
			strategy.nextServer(healthy).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
		}

		// not all strategies are going to be 50/50, but they shouldn't
		// balance to 100/0 if both are healthy.
		assert.Greater(t, recorder.save["A"], 0, "A should have been hit")
		assert.Greater(t, recorder.save["B"], 0, "B should have been hit")
		t.Logf("strategy %s with two backends has a ratio of %d:%d", strategy.name(), recorder.save["A"], recorder.save["B"])
	}
}

func testStrategyOneHealthyOneUnhealthy(strategy strategy) func(t *testing.T) {
	return func(t *testing.T) {
		strategy.add(testHandler("A", 1, 0))
		strategy.add(testHandler("B", 1, 0))

		strategy.setUp("B", false)
		healthy := map[string]struct{}{"A": {}}

		recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}

		const requests = 100
		for i := 0; i < requests; i++ {
			strategy.nextServer(healthy).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
		}

		assert.Equal(t, requests, recorder.save["A"], "A should have been hit with all requests")
		assert.Equal(t, 0, recorder.save["B"], "B should not have been hit")
	}
}

func testStrategyOneHostDownThenUp(strategy strategy) func(t *testing.T) {
	return func(t *testing.T) {
		strategy.add(testHandler("A", 1, 0))
		strategy.add(testHandler("B", 1, 0))

		strategy.setUp("A", false)
		strategy.setUp("A", true)

		healthy := map[string]struct{}{"A": {}, "B": {}}

		const requests = 100
		recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}

		for i := 0; i < requests; i++ {
			strategy.nextServer(healthy).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
		}

		assert.Greater(t, recorder.save["A"], 0, "A should have been hit")
		assert.Greater(t, recorder.save["B"], 0, "B should have been hit")
	}
}
