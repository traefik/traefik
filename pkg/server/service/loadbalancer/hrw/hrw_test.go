package hrw

import (
	"context"
	"encoding/binary"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// genIPAddress generate randomly an IP address as a string.
func genIPAddress() string {
	buf := make([]byte, 4)

	ip := rand.Uint32()

	binary.LittleEndian.PutUint32(buf, ip)
	ipStr := net.IP(buf)
	return ipStr.String()
}

// initStatusArray initialize an array filled with status value for test assertions.
func initStatusArray(size int, value int) []int {
	status := make([]int, 0, size)
	for i := 1; i <= size; i++ {
		status = append(status, value)
	}
	return status
}

// Tests evaluate load balancing of single and multiple clients.
// Due to the randomness of IP Adresses, repartition between services is not perfect
// The tests validate repartition using a margin of 10% of the number of requests

func TestBalancer(t *testing.T) {
	balancer := New(false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(4), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), Int(1), false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for range 100 {
		req.RemoteAddr = genIPAddress()
		balancer.ServeHTTP(recorder, req)
	}
	assert.InDelta(t, 80, recorder.save["first"], 10)
	assert.InDelta(t, 20, recorder.save["second"], 10)
}

func TestBalancerNoService(t *testing.T) {
	balancer := New(false)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

func TestBalancerOneServerZeroWeight(t *testing.T) {
	balancer := New(false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), Int(0), false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 3 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	assert.Equal(t, 3, recorder.save["first"])
}

type key string

const serviceName key = "serviceName"

func TestBalancerNoServiceUp(t *testing.T) {
	balancer := New(false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}), Int(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}), Int(1), false)

	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "first", false)
	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "second", false)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

func TestBalancerOneServerDown(t *testing.T) {
	balancer := New(false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	}), Int(1), false)
	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "second", false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 3 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}

	assert.Equal(t, 3, recorder.save["first"])
}

func TestBalancerDownThenUp(t *testing.T) {
	balancer := New(false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), Int(1), false)
	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "second", false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	for range 3 {
		balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	assert.Equal(t, 3, recorder.save["first"])

	balancer.SetStatus(context.WithValue(t.Context(), serviceName, "parent"), "second", true)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for range 100 {
		req.RemoteAddr = genIPAddress()
		balancer.ServeHTTP(recorder, req)
	}
	assert.InDelta(t, 50, recorder.save["first"], 10)
	assert.InDelta(t, 50, recorder.save["second"], 10)
}

func TestBalancerPropagate(t *testing.T) {
	balancer1 := New(true)

	balancer1.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(1), false)
	balancer1.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), Int(1), false)

	balancer2 := New(true)
	balancer2.Add("third", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "third")
		rw.WriteHeader(http.StatusOK)
	}), Int(1), false)
	balancer2.Add("fourth", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "fourth")
		rw.WriteHeader(http.StatusOK)
	}), Int(1), false)

	topBalancer := New(true)
	topBalancer.Add("balancer1", balancer1, Int(1), false)
	_ = balancer1.RegisterStatusUpdater(func(up bool) {
		topBalancer.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "balancer1", up)
		// TODO(mpl): if test gets flaky, add channel or something here to signal that
		// propagation is done, and wait on it before sending request.
	})
	topBalancer.Add("balancer2", balancer2, Int(1), false)
	_ = balancer2.RegisterStatusUpdater(func(up bool) {
		topBalancer.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "balancer2", up)
	})

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for range 100 {
		req.RemoteAddr = genIPAddress()
		topBalancer.ServeHTTP(recorder, req)
	}
	assert.InDelta(t, 25, recorder.save["first"], 10)
	assert.InDelta(t, 25, recorder.save["second"], 10)
	assert.InDelta(t, 25, recorder.save["third"], 10)
	assert.InDelta(t, 25, recorder.save["fourth"], 10)
	wantStatus := initStatusArray(100, 200)
	assert.Equal(t, wantStatus, recorder.status)

	// fourth gets downed, but balancer2 still up since third is still up.
	balancer2.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "fourth", false)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	for range 100 {
		req.RemoteAddr = genIPAddress()
		topBalancer.ServeHTTP(recorder, req)
	}
	assert.InDelta(t, 25, recorder.save["first"], 10)
	assert.InDelta(t, 25, recorder.save["second"], 10)
	assert.InDelta(t, 50, recorder.save["third"], 10)
	assert.InDelta(t, 0, recorder.save["fourth"], 0)
	wantStatus = initStatusArray(100, 200)
	assert.Equal(t, wantStatus, recorder.status)

	// third gets downed, and the propagation triggers balancer2 to be marked as
	// down as well for topBalancer.
	balancer2.SetStatus(context.WithValue(t.Context(), serviceName, "top"), "third", false)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	for range 100 {
		req.RemoteAddr = genIPAddress()
		topBalancer.ServeHTTP(recorder, req)
	}
	assert.InDelta(t, 50, recorder.save["first"], 10)
	assert.InDelta(t, 50, recorder.save["second"], 10)
	assert.InDelta(t, 0, recorder.save["third"], 0)
	assert.InDelta(t, 0, recorder.save["fourth"], 0)
	wantStatus = initStatusArray(100, 200)
	assert.Equal(t, wantStatus, recorder.status)
}

func TestBalancerAllServersZeroWeight(t *testing.T) {
	balancer := New(false)

	balancer.Add("test", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), Int(0), false)
	balancer.Add("test2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), Int(0), false)

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

func TestSticky(t *testing.T) {
	balancer := New(false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(1), false)

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), Int(2), false)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = genIPAddress()
	for range 10 {
		for _, cookie := range recorder.Result().Cookies() {
			req.AddCookie(cookie)
		}
		recorder.ResponseRecorder = httptest.NewRecorder()

		balancer.ServeHTTP(recorder, req)
	}

	assert.True(t, recorder.save["first"] == 0 || recorder.save["first"] == 10)
	assert.True(t, recorder.save["second"] == 0 || recorder.save["second"] == 10)
	// from one IP, the choice between server must be the same for the 10 requests
	// weight does not impose what would be chosen from 1 client
}

func Int(v int) *int { return &v }

type responseRecorder struct {
	*httptest.ResponseRecorder
	save     map[string]int
	sequence []string
	status   []int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.save[r.Header().Get("server")]++
	r.sequence = append(r.sequence, r.Header().Get("server"))
	r.status = append(r.status, statusCode)
	r.ResponseRecorder.WriteHeader(statusCode)
}
