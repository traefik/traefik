package accesslog

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"sync"
	"testing"

	"github.com/containous/alice"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/middlewares/capture"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	otypes "github.com/traefik/traefik/v3/pkg/observability/types"
)

// TestHandler_CloseRacesServeHTTP reproduces #13693: http.Server.Shutdown
// does not wait for hijacked connections, so a handler's deferred send to
// logHandlerChan can run after Close has already returned. The race is
// intermittent -- manual reproduction against the unpatched code produced
// clean runs about as often as it produced panics -- so this loops many
// trials instead of relying on a single one to catch a regression.
func TestHandler_CloseRacesServeHTTP(t *testing.T) {
	const trials = 50
	const concurrentRequests = 200

	for trial := range trials {
		t.Run(fmt.Sprintf("trial-%d", trial), func(t *testing.T) {
			t.Parallel()

			logHandler, err := NewHandler(t.Context(), &otypes.AccessLog{
				FilePath:      filepath.Join(t.TempDir(), "access.log"),
				Format:        JSONFormat,
				BufferingSize: 100,
			})
			require.NoError(t, err)

			chain := alice.New()
			chain = chain.Append(capture.Wrap)
			chain = chain.Append(func(next http.Handler) (http.Handler, error) {
				return observability.WithObservabilityHandler(next, observability.Observability{
					AccessLogsEnabled: true,
				}), nil
			})
			chain = chain.Append(logHandler.AliceConstructor())
			handler, err := chain.Then(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				rw.WriteHeader(http.StatusOK)
			}))
			require.NoError(t, err)

			var wg sync.WaitGroup
			for range concurrentRequests {
				wg.Go(func() {
					defer func() {
						if r := recover(); r != nil {
							t.Errorf("ServeHTTP panicked: %v", r)
						}
					}()

					req := &http.Request{
						Header:     map[string][]string{},
						Proto:      testProto,
						Host:       testHostname,
						Method:     testMethod,
						RemoteAddr: fmt.Sprintf("%s:%d", testHostname, testPort+trial),
						URL:        &url.URL{Path: testPath},
					}
					handler.ServeHTTP(httptest.NewRecorder(), req)
				})
			}

			// Close races the in-flight requests above without waiting for them,
			// mirroring how Traefik's force-close after graceTimeOut tears down
			// many hijacked handlers at once -- well after Close has already run
			// for most of them.
			require.NoError(t, logHandler.Close())

			wg.Wait()

			// Close must also be safe to call a second time.
			require.NoError(t, logHandler.Close())
		})
	}
}
