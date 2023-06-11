package hrw

import (
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func genIPAddress() string {
	buf := make([]byte, 4)

	ip := rand.Uint32()

	binary.LittleEndian.PutUint32(buf, ip)
	ipStr := net.IP(buf)
	fmt.Printf("%s\n", ipStr)
	return ipStr.String()
}

func TestBalancer(t *testing.T) {
	balancer := New(false)

	balancer.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(4))

	balancer.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for i := 0; i < 100; i++ {
		req.RemoteAddr = genIPAddress()
		balancer.ServeHTTP(recorder, req)
	}
	assert.InDelta(t, 80, recorder.save["first"], 5)
	assert.InDelta(t, 20, recorder.save["second"], 5)
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
	balancer := New(false)

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
	balancer := New(false)

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
	balancer := New(false)

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
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for i := 0; i < 100; i++ {
		req.RemoteAddr = genIPAddress()
		balancer.ServeHTTP(recorder, req)
	}
	assert.InDelta(t, 50, recorder.save["first"], 5)
	assert.InDelta(t, 50, recorder.save["second"], 5)

}

// test not working

func TestBalancerPropagate(t *testing.T) {
	balancer1 := New(true)

	balancer1.Add("first", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "first")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))
	balancer1.Add("second", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "second")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))

	balancer2 := New(true)
	balancer2.Add("third", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "third")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))
	balancer2.Add("fourth", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "fourth")
		rw.WriteHeader(http.StatusOK)
	}), Int(1))

	topBalancer := New(true)
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
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for i := 0; i < 8; i++ {
		req.RemoteAddr = genIPAddress()
		topBalancer.ServeHTTP(recorder, req)
	}
	assert.InDelta(t, 2, recorder.save["first"], 1)
	assert.InDelta(t, 2, recorder.save["second"], 1)
	assert.InDelta(t, 2, recorder.save["third"], 1)
	assert.InDelta(t, 2, recorder.save["fourth"], 1)
	wantStatus := []int{200, 200, 200, 200, 200, 200, 200, 200}
	assert.Equal(t, wantStatus, recorder.status)

	// fourth gets downed, but balancer2 still up since third is still up.
	balancer2.SetStatus(context.WithValue(context.Background(), serviceName, "top"), "fourth", false)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	for i := 0; i < 8; i++ {
		req.RemoteAddr = genIPAddress()
		topBalancer.ServeHTTP(recorder, req)
	}
	assert.InDelta(t, 2, recorder.save["first"], 1)
	assert.InDelta(t, 2, recorder.save["second"], 1)
	assert.InDelta(t, 4, recorder.save["third"], 1)
	assert.InDelta(t, 0, recorder.save["fourth"], 0)
	wantStatus = []int{200, 200, 200, 200, 200, 200, 200, 200}
	assert.Equal(t, wantStatus, recorder.status)

	// third gets downed, and the propagation triggers balancer2 to be marked as
	// down as well for topBalancer.
	balancer2.SetStatus(context.WithValue(context.Background(), serviceName, "top"), "third", false)
	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	// part of test not working
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	for i := 0; i < 8; i++ {
		req.RemoteAddr = genIPAddress()
		topBalancer.ServeHTTP(recorder, req)
	}
	assert.InDelta(t, 4, recorder.save["first"], 1)
	assert.InDelta(t, 4, recorder.save["second"], 1)
	assert.InDelta(t, 0, recorder.save["third"], 0)
	assert.InDelta(t, 0, recorder.save["fourth"], 0)
	wantStatus = []int{200, 200, 200, 200, 200, 200, 200, 200}
	assert.Equal(t, wantStatus, recorder.status)
}

func TestBalancerAllServersZeroWeight(t *testing.T) {
	balancer := New(false)

	balancer.Add("test", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), Int(0))
	balancer.Add("test2", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), Int(0))

	recorder := httptest.NewRecorder()
	balancer.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusServiceUnavailable, recorder.Result().StatusCode)
}

func TestSticky(t *testing.T) {
	balancer := New(false)

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
	req.RemoteAddr = genIPAddress()
	for i := 0; i < 10; i++ {
		for _, cookie := range recorder.Result().Cookies() {
			req.AddCookie(cookie)
		}
		recorder.ResponseRecorder = httptest.NewRecorder()

		balancer.ServeHTTP(recorder, req)
	}

	assert.True(t, recorder.save["first"] == 0 || recorder.save["first"] == 10)
	assert.True(t, recorder.save["second"] == 0 || recorder.save["second"] == 10)
	// from one IP, the choice between server must be the same for the 10 requests
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
