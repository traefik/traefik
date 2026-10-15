package buffering

import (
	"bytes"
	"crypto/rand"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
	oxybuffer "github.com/vulcand/oxy/v2/buffer"
)

func TestBuffering(t *testing.T) {
	payload := make([]byte, math.MaxInt8)
	_, _ = rand.Read(payload)

	testCases := []struct {
		desc         string
		config       dynamic.Buffering
		body         []byte
		expectedCode int
	}{
		{
			desc:         "Unlimited response and request body size",
			body:         payload,
			expectedCode: http.StatusOK,
		},
		{
			desc: "Limited request body size",
			config: dynamic.Buffering{
				MaxRequestBodyBytes: 1,
			},
			body:         payload,
			expectedCode: http.StatusRequestEntityTooLarge,
		},
		{
			desc: "Limited response body size",
			config: dynamic.Buffering{
				MaxResponseBodyBytes: 1,
			},
			body:         payload,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
				_, err := rw.Write(test.body)
				require.NoError(t, err)
			})

			buffMiddleware, err := New(t.Context(), next, test.config, "foo")
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "http://localhost", bytes.NewBuffer(test.body))

			recorder := httptest.NewRecorder()
			buffMiddleware.ServeHTTP(recorder, req)

			assert.Equal(t, test.expectedCode, recorder.Code)
		})
	}
}

// TestOxyBuffer_requiresWrapExternal documents why the buffering handler is built with middlewares.WrapExternal.
// oxy's buffer asserts http.Hijacker on the ResponseWriter it receives instead of using http.ResponseController,
// so Hijack does not reach a ResponseWriter which only implements Unwrap, as Traefik's wrappers do.
// The ResponseWriter it gives to the next handler implements neither Unwrap nor Flush.
// If these cases fail after bumping vulcand/oxy, the library changed
// and WrapExternal may no longer be needed in New.
func TestOxyBuffer_requiresWrapExternal(t *testing.T) {
	testCases := []struct {
		desc           string
		unwrapOnly     bool
		expectHijacked bool
	}{
		{
			desc:           "ResponseWriter implementing Hijack",
			expectHijacked: true,
		},
		{
			desc:       "ResponseWriter which only implements Unwrap",
			unwrapOnly: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var nextSupport testhelpers.ResponseWriterSupport
			handler, err := oxybuffer.New(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				nextSupport = testhelpers.SupportOf(rw)

				// The outcome is checked on the underlying recorder.
				if _, _, err := http.NewResponseController(rw).Hijack(); err != nil {
					rw.WriteHeader(http.StatusOK)
				}
			}))
			require.NoError(t, err)

			recorder := testhelpers.NewHijackableRecorder()
			var rw http.ResponseWriter = recorder
			if test.unwrapOnly {
				rw = &testhelpers.UnwrapOnlyResponseWriter{ResponseWriter: recorder}
			}

			handler.ServeHTTP(rw, httptest.NewRequest(http.MethodGet, "/", nil))

			assert.Equal(t, test.expectHijacked, recorder.Hijacked)
			assert.Equal(t, testhelpers.ResponseWriterSupport{Hijack: true}, nextSupport)
		})
	}
}
