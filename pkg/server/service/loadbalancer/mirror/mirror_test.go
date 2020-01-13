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
	err := mirror.AddMirror(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&countMirror1, 1)
	}), 10)
	assert.NoError(t, err)

	err = mirror.AddMirror(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&countMirror2, 1)
	}), 50)
	assert.NoError(t, err)

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
	err := mirror.AddMirror(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&countMirror1, 1)
	}), 10)
	assert.NoError(t, err)

	err = mirror.AddMirror(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		atomic.AddInt32(&countMirror2, 1)
	}), 50)
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		mirror.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}

	pool.Stop()

	val1 := atomic.LoadInt32(&countMirror1)
	val2 := atomic.LoadInt32(&countMirror2)
	assert.Equal(t, 1, int(val1))
	assert.Equal(t, 5, int(val2))
}

func TestInvalidPercent(t *testing.T) {
	mirror := New(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {}), safe.NewPool(context.Background()))
	err := mirror.AddMirror(nil, -1)
	assert.Error(t, err)

	err = mirror.AddMirror(nil, 101)
	assert.Error(t, err)

	err = mirror.AddMirror(nil, 100)
	assert.NoError(t, err)

	err = mirror.AddMirror(nil, 0)
	assert.NoError(t, err)
}

func TestHijack(t *testing.T) {
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})
	pool := safe.NewPool(context.Background())
	mirror := New(handler, pool)

	var mirrorRequest bool
	err := mirror.AddMirror(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		hijacker, ok := rw.(http.Hijacker)
		assert.Equal(t, true, ok)

		_, _, err := hijacker.Hijack()
		assert.Error(t, err)
		mirrorRequest = true
	}), 100)
	assert.NoError(t, err)

	mirror.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	pool.Stop()
	assert.Equal(t, true, mirrorRequest)
}

func TestFlush(t *testing.T) {
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})
	pool := safe.NewPool(context.Background())
	mirror := New(handler, pool)

	var mirrorRequest bool
	err := mirror.AddMirror(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		hijacker, ok := rw.(http.Flusher)
		assert.Equal(t, true, ok)

		hijacker.Flush()

		mirrorRequest = true
	}), 100)
	assert.NoError(t, err)

	mirror.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	pool.Stop()
	assert.Equal(t, true, mirrorRequest)
}
