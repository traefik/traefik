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
