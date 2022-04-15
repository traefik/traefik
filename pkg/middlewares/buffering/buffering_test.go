package buffering

import (
	"bytes"
	"context"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
)

func TestBuffering(t *testing.T) {
	bigPayload := make([]byte, math.MaxInt8)
	rand.Read(bigPayload)

	testCases := []struct {
		desc         string
		config       dynamic.Buffering
		body         []byte
		expectedCode int
	}{
		{
			desc:         "Unlimited response and request body size",
			config:       dynamic.Buffering{},
			body:         []byte("FOOBAR"),
			expectedCode: http.StatusOK,
		},
		{
			desc:         "Unlimited response and request body size, with big payload",
			config:       dynamic.Buffering{},
			body:         bigPayload,
			expectedCode: http.StatusOK,
		},
		{
			desc: "Limited request body size",
			config: dynamic.Buffering{
				MaxRequestBodyBytes: 1,
			},
			body:         []byte("FOOBAR"),
			expectedCode: http.StatusRequestEntityTooLarge,
		},
		{
			desc: "Limited request body size, with big payload",
			config: dynamic.Buffering{
				MaxRequestBodyBytes: 1,
			},
			body:         bigPayload,
			expectedCode: http.StatusRequestEntityTooLarge,
		},
		{
			desc: "Limited response body size",
			config: dynamic.Buffering{
				MaxResponseBodyBytes: 1,
			},
			body:         []byte("FOOBAR"),
			expectedCode: http.StatusInternalServerError,
		},
		{
			desc: "Limited response body size, with big payload",
			config: dynamic.Buffering{
				MaxResponseBodyBytes: 1,
			},
			body:         bigPayload,
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
				rw.Write(test.body)
			})

			bufHandler, err := New(context.Background(), next, test.config, "foo")
			require.Nil(t, err)

			req, err := http.NewRequest(http.MethodPost, "", bytes.NewBuffer(test.body))
			require.Nil(t, err)

			recorder := httptest.NewRecorder()
			bufHandler.ServeHTTP(recorder, req)

			assert.Equal(t, test.expectedCode, recorder.Code)
		})
	}
}
