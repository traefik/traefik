package wrr

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/stretchr/testify/assert"
)

func Int(v int) *int { return &v }

type responseRecorder struct {
	*httptest.ResponseRecorder
	save map[string]int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.save[r.Header().Get("server")]++
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
