package wrr

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

func Int(v int) *int { return &v }

type responseRecorder struct {
	*httptest.ResponseRecorder
	save     map[string]int
	sequence []string
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.save[r.Header().Get("server")]++
	r.sequence = append(r.sequence, r.Header().Get("server"))
	r.ResponseRecorder.WriteHeader(statusCode)
}

func TestBalancer(t *testing.T) {
	balancer := New(nil)

	balancer.AddService("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(3))

	balancer.AddService("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
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
	balancer := New(nil)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusInternalServerError, recorder.Result().StatusCode)
}

func TestBalancerOneServerZeroWeight(t *testing.T) {
	balancer := New(nil)

	balancer.AddService("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))

	balancer.AddService("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), Int(0))

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for i := 0; i < 3; i++ {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	assert.Equal(t, 3, recorder.save["first"])
}

func TestBalancerAllServersZeroWeight(t *testing.T) {
	balancer := New(nil)

	balancer.AddService("test", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), Int(0))
	balancer.AddService("test2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), Int(0))

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusInternalServerError, recorder.Result().StatusCode)
}

func TestSticky(t *testing.T) {
	balancer := New(&dynamic.Sticky{
		Cookie: &dynamic.Cookie{Name: "test"},
	})

	balancer.AddService("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))

	balancer.AddService("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), Int(2))

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for i := 0; i < 3; i++ {
		for _, cookie := range recorder.Result().Cookies() {
			req.AddCookie(cookie)
		}
		recorder.ResponseRecorder = httptest.NewRecorder()

		balancer.ServeHTTP(recorder, req)
	}

	assert.Equal(t, 0, recorder.save["first"])
	assert.Equal(t, 3, recorder.save["second"])
}

// TestBalancerBias makes sure that the WRR algorithm spreads elements evenly right from the start,
// and that it does not "over-favor" the high-weighted ones with a biased start-up regime.
func TestBalancerBias(t *testing.T) {
	balancer := New(nil)

	balancer.AddService("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "A")
		rw.WriteHeader(http.StatusOK)
	}), Int(11))

	balancer.AddService("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
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
