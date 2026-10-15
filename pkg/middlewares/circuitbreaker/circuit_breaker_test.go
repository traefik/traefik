package circuitbreaker

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
	"github.com/vulcand/oxy/v2/cbreaker"
)

// TestOxyCircuitBreaker_requiresWrapExternal documents why the circuit breaker handler is built with middlewares.WrapExternal.
// oxy's circuit breaker asserts http.Flusher and http.Hijacker on the ResponseWriter it receives instead of using http.ResponseController,
// so Flush and Hijack do not reach a ResponseWriter which only implements Unwrap, as Traefik's wrappers do.
// The ResponseWriter it gives to the next handler does not implement Unwrap.
// If these cases fail after bumping vulcand/oxy, the library changed
// and WrapExternal may no longer be needed in New.
func TestOxyCircuitBreaker_requiresWrapExternal(t *testing.T) {
	testCases := []struct {
		desc           string
		unwrapOnly     bool
		hijack         bool
		expectFlushed  bool
		expectHijacked bool
	}{
		{
			desc:          "Flush with a ResponseWriter implementing Flush",
			expectFlushed: true,
		},
		{
			desc:       "Flush with a ResponseWriter which only implements Unwrap",
			unwrapOnly: true,
		},
		{
			desc:           "Hijack with a ResponseWriter implementing Hijack",
			hijack:         true,
			expectHijacked: true,
		},
		{
			desc:       "Hijack with a ResponseWriter which only implements Unwrap",
			unwrapOnly: true,
			hijack:     true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var nextSupport testhelpers.ResponseWriterSupport
			handler, err := cbreaker.New(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				nextSupport = testhelpers.SupportOf(rw)

				rc := http.NewResponseController(rw)

				// The outcome is checked on the underlying recorder.
				if test.hijack {
					_, _, _ = rc.Hijack()
					return
				}

				_ = rc.Flush()
			}), "NetworkErrorRatio() > 0.5")
			require.NoError(t, err)

			recorder := testhelpers.NewHijackableRecorder()
			var rw http.ResponseWriter = recorder
			if test.unwrapOnly {
				rw = &testhelpers.UnwrapOnlyResponseWriter{ResponseWriter: recorder}
			}

			handler.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/", nil))

			assert.Equal(t, test.expectFlushed, recorder.Flushed)
			assert.Equal(t, test.expectHijacked, recorder.Hijacked)
			assert.Equal(t, testhelpers.ResponseWriterSupport{Flush: true, Hijack: true}, nextSupport)
		})
	}
}
