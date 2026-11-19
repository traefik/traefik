package ping

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHandler_WithContext_Termination(t *testing.T) {
	handler := &Handler{}
	handler.SetDefaults()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	handler.WithContext(ctx)
	cancel()

	require.Eventually(t, func() bool { return handler.terminating.Load() }, time.Second, time.Millisecond)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "http://127.0.0.1/ping", nil))

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.Equal(t, http.StatusText(http.StatusServiceUnavailable), recorder.Body.String())
}
