package mirror

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/stretchr/testify/assert"
)

func TestMirroringOn100(t *testing.T) {
	var countMirror1, countMirror2 int32
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})
	pool := safe.NewPool(context.Background())
	mirror := New(handler, pool)
	mirror.AddMirror(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&countMirror1, 1)
	}), 10)
	mirror.AddMirror(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&countMirror2, 1)
	}), 50)

	for i := 0; i < 100; i++ {
		mirror.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}

	pool.Stop()

	val1 := atomic.LoadInt32(&countMirror1)
	val2 := atomic.LoadInt32(&countMirror2)
	assert.Equal(t, 10, int(val1))
	assert.Equal(t, 50, int(val2))
}

func TestMirroringOn10(t *testing.T) {
	var countMirror1, countMirror2 int32
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})
	pool := safe.NewPool(context.Background())
	mirror := New(handler, pool)
	mirror.AddMirror(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&countMirror1, 1)
	}), 10)
	mirror.AddMirror(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&countMirror2, 1)
	}), 50)

	for i := 0; i < 10; i++ {
		mirror.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}

	pool.Stop()

	val1 := atomic.LoadInt32(&countMirror1)
	val2 := atomic.LoadInt32(&countMirror2)
	assert.Equal(t, 1, int(val1))
	assert.Equal(t, 5, int(val2))
}
