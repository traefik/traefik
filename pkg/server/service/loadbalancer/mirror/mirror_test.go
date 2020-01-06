package mirror

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMirroringOn100(t *testing.T) {
	var countMirror1, countMirror2 int32
	var wg sync.WaitGroup
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		wg.Add(1)
		rw.WriteHeader(http.StatusOK)
	})
	mirror := New(handler)
	mirror.routineTestHook = func() {
		wg.Done()
	}
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

	wg.Wait()

	val1 := atomic.LoadInt32(&countMirror1)
	val2 := atomic.LoadInt32(&countMirror2)
	assert.Equal(t, 10, int(val1))
	assert.Equal(t, 50, int(val2))
}

func TestMirroringOn10Z(t *testing.T) {
	var countMirror1, countMirror2 int32
	var wg sync.WaitGroup
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		wg.Add(1)
		rw.WriteHeader(http.StatusOK)
	})
	mirror := New(handler)
	mirror.routineTestHook = func() {
		wg.Done()
	}
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

	wg.Wait()

	val1 := atomic.LoadInt32(&countMirror1)
	val2 := atomic.LoadInt32(&countMirror2)
	assert.Equal(t, 1, int(val1))
	assert.Equal(t, 5, int(val2))
}

func TestInvalidPercent(t *testing.T) {
	var wg sync.WaitGroup
	mirror := New(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		wg.Add(1)
	}))
	mirror.routineTestHook = func() {
		wg.Done()
	}
	err := mirror.AddMirror(nil, -1)
	assert.Error(t, err)

	err = mirror.AddMirror(nil, 101)
	assert.Error(t, err)

	err = mirror.AddMirror(nil, 100)
	assert.NoError(t, err)

	err = mirror.AddMirror(nil, 0)
	assert.NoError(t, err)

	wg.Wait()
}

func TestHijack(t *testing.T) {
	var wg sync.WaitGroup
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		wg.Add(1)
		rw.WriteHeader(http.StatusOK)
	})
	mirror := New(handler)
	mirror.routineTestHook = func() {
		wg.Done()
	}

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

	wg.Wait()

	assert.Equal(t, true, mirrorRequest)
}

func TestFlush(t *testing.T) {
	var wg sync.WaitGroup
	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		wg.Add(1)
		rw.WriteHeader(http.StatusOK)
	})
	mirror := New(handler)
	mirror.routineTestHook = func() {
		wg.Done()
	}

	var mirrorRequest bool
	err := mirror.AddMirror(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		hijacker, ok := rw.(http.Flusher)
		assert.Equal(t, true, ok)

		hijacker.Flush()

		mirrorRequest = true
	}), 100)
	assert.NoError(t, err)

	mirror.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	wg.Wait()

	assert.Equal(t, true, mirrorRequest)
}
