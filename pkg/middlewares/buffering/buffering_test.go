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

// TestBuffering_DisableResponseBuffer verifies that setting DisableResponseBuffer:
//   - exposes the original ResponseWriter (notably http.Flusher) to the downstream handler,
//     which is required to forward streaming responses such as Server-Sent Events,
//     gRPC streaming, or long-poll watch endpoints; and
//   - bypasses MaxResponseBodyBytes enforcement.
func TestBuffering_DisableResponseBuffer(t *testing.T) {
	payload := make([]byte, math.MaxInt8)
	_, _ = rand.Read(payload)

	testCases := []struct {
		desc                 string
		config               dynamic.Buffering
		expectHandlerFlusher bool
		expectedCode         int
	}{
		{
			desc:                 "default buffers the response; handler sees wrapped ResponseWriter without Flusher",
			config:               dynamic.Buffering{MaxResponseBodyBytes: int64(len(payload) - 1)},
			expectHandlerFlusher: false,
			expectedCode:         http.StatusInternalServerError,
		},
		{
			desc:                 "DisableResponseBuffer passes the original ResponseWriter through and skips MaxResponseBodyBytes",
			config:               dynamic.Buffering{DisableResponseBuffer: true, MaxResponseBodyBytes: int64(len(payload) - 1)},
			expectHandlerFlusher: true,
			expectedCode:         http.StatusOK,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			var gotFlusher bool
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				_, gotFlusher = rw.(http.Flusher)
				rw.WriteHeader(http.StatusOK)
				_, _ = rw.Write(payload)
			})

			buffMiddleware, err := New(t.Context(), next, test.config, "foo")
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "http://localhost", http.NoBody)

			recorder := httptest.NewRecorder()
			buffMiddleware.ServeHTTP(recorder, req)

			assert.Equal(t, test.expectHandlerFlusher, gotFlusher)
			assert.Equal(t, test.expectedCode, recorder.Code)
		})
	}
}

// TestBuffering_DisableRequestBuffer verifies that setting DisableRequestBuffer bypasses
// request body buffering entirely, including MaxRequestBodyBytes enforcement. This matches
// the semantics of nginx's proxy-request-buffering: off.
func TestBuffering_DisableRequestBuffer(t *testing.T) {
	payload := make([]byte, math.MaxInt8)
	_, _ = rand.Read(payload)

	testCases := []struct {
		desc         string
		config       dynamic.Buffering
		expectedCode int
	}{
		{
			desc:         "default enforces MaxRequestBodyBytes",
			config:       dynamic.Buffering{MaxRequestBodyBytes: int64(len(payload) - 1)},
			expectedCode: http.StatusRequestEntityTooLarge,
		},
		{
			desc:         "DisableRequestBuffer bypasses MaxRequestBodyBytes enforcement",
			config:       dynamic.Buffering{DisableRequestBuffer: true, MaxRequestBodyBytes: int64(len(payload) - 1)},
			expectedCode: http.StatusOK,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
				_, _ = rw.Write([]byte("ok"))
			})

			buffMiddleware, err := New(t.Context(), next, test.config, "foo")
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "http://localhost", bytes.NewBuffer(payload))

			recorder := httptest.NewRecorder()
			buffMiddleware.ServeHTTP(recorder, req)

			assert.Equal(t, test.expectedCode, recorder.Code)
		})
	}
}
