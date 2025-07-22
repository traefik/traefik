package wrr

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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

	balancer.Add("test", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), pointer(1), true)
	balancer.Add("test2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), pointer(1), true)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
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
