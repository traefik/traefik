package circuitbreaker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/paerser/types"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
)

func TestNewInvalidExpression(t *testing.T) {
	_, err := New(context.Background(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), dynamic.CircuitBreaker{
		Expression: "not a valid expression",
	}, "test")
	require.Error(t, err)
}

func TestCircuitBreakerOpensAndRecovers(t *testing.T) {
	var failing atomic.Bool
	failing.Store(true)

	next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		if failing.Load() {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(http.StatusOK)
	})

	handler, err := New(context.Background(), next, dynamic.CircuitBreaker{
		Expression:       "ResponseCodeRatio(500, 600, 0, 600) > 0.5",
		CheckPeriod:      types.Duration(10 * time.Millisecond),
		FallbackDuration: types.Duration(100 * time.Millisecond),
		RecoveryDuration: types.Duration(100 * time.Millisecond),
		ResponseCode:     http.StatusServiceUnavailable,
	}, "test")
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	var status int
	for range 100 {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		status = recorder.Code
		if status == http.StatusServiceUnavailable {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	assert.Equal(t, http.StatusServiceUnavailable, status)

	failing.Store(false)
	for range 100 {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		status = recorder.Code
		if status == http.StatusOK {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	assert.Equal(t, http.StatusOK, status)
}
