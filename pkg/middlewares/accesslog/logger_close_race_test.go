package accesslog

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	"github.com/containous/alice"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	otypes "github.com/traefik/traefik/v3/pkg/observability/types"
)

// http.Server.Shutdown neither closes nor waits for hijacked connections, so a
// proxied WebSocket handler can still unwind after Close has closed
// logHandlerChan. With bufferingSize > 0 the deferred send in ServeHTTP then
// ran against a closed channel and panicked with "send on closed channel",
// taking the process down during a restart.
//
// See https://github.com/traefik/traefik/issues/13693
func TestHandler_ServeHTTPConcurrentWithClose(t *testing.T) {
	config := &otypes.AccessLog{
		FilePath:      filepath.Join(t.TempDir(), "traefik.log"),
		Format:        CommonFormat,
		BufferingSize: 100,
	}

	logHandler, err := NewHandler(t.Context(), config)
	require.NoError(t, err)

	chain := alice.New()
	chain = chain.Append(capture.Wrap)
	chain = chain.Append(func(next http.Handler) (http.Handler, error) {
		return observability.WithObservabilityHandler(next, observability.Observability{
			AccessLogsEnabled: true,
		}), nil
	})
	chain = chain.Append(logHandler.AliceConstructor())

	handler, err := chain.Then(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))
	require.NoError(t, err)

	// Handlers that are still unwinding while Close runs, standing in for the
	// hijacked connections that Shutdown leaves behind.
	const requests = 200

	var start, done sync.WaitGroup
	start.Add(1)
	done.Add(requests)

	for range requests {
		go func() {
			defer done.Done()
			start.Wait()
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://localhost", nil))
		}()
	}

	closed := make(chan error, 1)
	go func() {
		start.Wait()
		closed <- logHandler.Close()
	}()

	start.Done()

	done.Wait()
	require.NoError(t, <-closed)

	// Close must stay safe to call again once the handler is shut down.
	require.NoError(t, logHandler.Close())
}
