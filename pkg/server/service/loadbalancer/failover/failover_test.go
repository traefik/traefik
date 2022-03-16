package failover

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

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

func TestFailover(t *testing.T) {
	failover := New(&dynamic.HealthCheck{})

	status := true
	require.NoError(t, failover.RegisterStatusUpdater(func(up bool) {
		status = up
	}))

	failover.SetHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "handler")
		rw.WriteHeader(http.StatusOK)
	}))

	failover.SetFallbackHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "fallback")
		rw.WriteHeader(http.StatusOK)
	}))

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 1, recorder.save["handler"])
	assert.Equal(t, 0, recorder.save["fallback"])
	assert.Equal(t, []int{200}, recorder.status)
	assert.True(t, status)

	failover.SetHandlerStatus(context.Background(), false)

	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 0, recorder.save["handler"])
	assert.Equal(t, 1, recorder.save["fallback"])
	assert.Equal(t, []int{200}, recorder.status)
	assert.True(t, status)

	failover.SetFallbackHandlerStatus(context.Background(), false)

	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 0, recorder.save["handler"])
	assert.Equal(t, 0, recorder.save["fallback"])
	assert.Equal(t, []int{503}, recorder.status)
	assert.False(t, status)
}

func TestFailoverDownThenUp(t *testing.T) {
	failover := New(nil)

	failover.SetHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "handler")
		rw.WriteHeader(http.StatusOK)
	}))

	failover.SetFallbackHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "fallback")
		rw.WriteHeader(http.StatusOK)
	}))

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 1, recorder.save["handler"])
	assert.Equal(t, 0, recorder.save["fallback"])
	assert.Equal(t, []int{200}, recorder.status)

	failover.SetHandlerStatus(context.Background(), false)

	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 0, recorder.save["handler"])
	assert.Equal(t, 1, recorder.save["fallback"])
	assert.Equal(t, []int{200}, recorder.status)

	failover.SetHandlerStatus(context.Background(), true)

	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	failover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 1, recorder.save["handler"])
	assert.Equal(t, 0, recorder.save["fallback"])
	assert.Equal(t, []int{200}, recorder.status)
}

func TestFailoverPropagate(t *testing.T) {
	failover := New(&dynamic.HealthCheck{})
	failover.SetHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "handler")
		rw.WriteHeader(http.StatusOK)
	}))
	failover.SetFallbackHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "fallback")
		rw.WriteHeader(http.StatusOK)
	}))

	topFailover := New(nil)
	topFailover.SetHandler(failover)
	topFailover.SetFallbackHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("server", "topFailover")
		rw.WriteHeader(http.StatusOK)
	}))
	err := failover.RegisterStatusUpdater(func(up bool) {
		topFailover.SetHandlerStatus(context.Background(), up)
	})
	require.NoError(t, err)

	recorder := &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	topFailover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 1, recorder.save["handler"])
	assert.Equal(t, 0, recorder.save["fallback"])
	assert.Equal(t, 0, recorder.save["topFailover"])
	assert.Equal(t, []int{200}, recorder.status)

	failover.SetHandlerStatus(context.Background(), false)

	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	topFailover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 0, recorder.save["handler"])
	assert.Equal(t, 1, recorder.save["fallback"])
	assert.Equal(t, 0, recorder.save["topFailover"])
	assert.Equal(t, []int{200}, recorder.status)

	failover.SetFallbackHandlerStatus(context.Background(), false)

	recorder = &responseRecorder{ResponseRecorder: httptest.NewRecorder(), save: map[string]int{}}
	topFailover.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, 0, recorder.save["handler"])
	assert.Equal(t, 0, recorder.save["fallback"])
	assert.Equal(t, 1, recorder.save["topFailover"])
	assert.Equal(t, []int{200}, recorder.status)
}
