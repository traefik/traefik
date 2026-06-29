package buffering

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
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

func TestBufferingDisabledRequestBuffer(t *testing.T) {
	payload := make([]byte, 128)
	_, _ = rand.Read(payload)

	testCases := []struct {
		desc         string
		config       dynamic.Buffering
		body         []byte
		contentLen   int64 // -1 to strip Content-Length (simulate streaming)
		expectedCode int
	}{
		{
			desc: "No limit passes through",
			config: dynamic.Buffering{
				DisableRequestBuffer: true,
			},
			body:         payload,
			contentLen:   0,
			expectedCode: http.StatusOK,
		},
		{
			desc: "Known Content-Length below limit passes through",
			config: dynamic.Buffering{
				DisableRequestBuffer: true,
				MaxRequestBodyBytes:  int64(len(payload)),
			},
			body:         payload,
			contentLen:   0,
			expectedCode: http.StatusOK,
		},
		{
			desc: "Known Content-Length exceeds limit rejected immediately",
			config: dynamic.Buffering{
				DisableRequestBuffer: true,
				MaxRequestBodyBytes:  1,
			},
			body:         payload,
			contentLen:   0,
			expectedCode: http.StatusRequestEntityTooLarge,
		},
		{
			desc: "Streaming body exceeds limit rejected on read",
			config: dynamic.Buffering{
				DisableRequestBuffer: true,
				MaxRequestBodyBytes:  1,
			},
			body:       payload,
			contentLen: -1, // strip Content-Length to simulate chunked/streaming
			expectedCode: http.StatusRequestEntityTooLarge,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			// next reads the full body so MaxBytesReader can trigger in the streaming case.
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				_, err := io.ReadAll(req.Body)
				if err != nil {
					var maxBytesErr *http.MaxBytesError
					if errors.As(err, &maxBytesErr) {
						http.Error(rw, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
						return
					}
					http.Error(rw, err.Error(), http.StatusInternalServerError)
					return
				}
				rw.WriteHeader(http.StatusOK)
			})

			buffMiddleware, err := New(t.Context(), next, test.config, "foo")
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "http://localhost", bytes.NewBuffer(test.body))
			if test.contentLen == -1 {
				req.ContentLength = -1
			}

			recorder := httptest.NewRecorder()
			buffMiddleware.ServeHTTP(recorder, req)

			assert.Equal(t, test.expectedCode, recorder.Code)
		})
	}
}
