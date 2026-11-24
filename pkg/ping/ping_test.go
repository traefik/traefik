package ping

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerServeHTTP(t *testing.T) {
	handler := &Handler{TerminatingStatusCode: http.StatusServiceUnavailable}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ping", nil))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, http.StatusText(http.StatusOK), recorder.Body.String())
}

func TestHandlerWithContext(t *testing.T) {
	handler := &Handler{TerminatingStatusCode: http.StatusServiceUnavailable}
	ctx, cancel := context.WithCancel(context.Background())
	handler.WithContext(ctx)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ping", nil))
	assert.Equal(t, http.StatusOK, recorder.Code)

	cancel()

	require.Eventually(t, func() bool {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ping", nil))

		return recorder.Code == http.StatusServiceUnavailable
	}, time.Second, time.Millisecond)

	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ping", nil))
	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	assert.Equal(t, http.StatusText(http.StatusServiceUnavailable), recorder.Body.String())
}

// TestHandlerTerminatingRace exercises requests racing with the context
// cancellation that sets the terminating flag. Run with -race to detect the
// unsynchronized access to the flag.
func TestHandlerTerminatingRace(t *testing.T) {
	handler := &Handler{TerminatingStatusCode: http.StatusServiceUnavailable}
	ctx, cancel := context.WithCancel(context.Background())
	handler.WithContext(ctx)

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range 100 {
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/ping", nil))
			}
		}()
	}

	cancel()

	wg.Wait()
}
