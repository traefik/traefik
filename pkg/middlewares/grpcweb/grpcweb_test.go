package grpcweb

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/traefik/grpc-web/go/grpcweb"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

// TestGrpcWeb_requiresWrapExternal documents why the gRPC-Web handler is built with middlewares.WrapExternal.
// grpc-web asserts http.Flusher on the ResponseWriter it receives instead of using http.ResponseController,
// so Flush does not reach a ResponseWriter which only implements Unwrap, as Traefik's wrappers do.
// The ResponseWriter it gives to the next handler implements neither Unwrap nor Hijack.
// If these cases fail after bumping traefik/grpc-web, the library changed
// and WrapExternal may no longer be needed in New.
func TestGrpcWeb_requiresWrapExternal(t *testing.T) {
	testCases := []struct {
		desc          string
		unwrapOnly    bool
		expectFlushed bool
	}{
		{
			desc:          "ResponseWriter implementing Flush",
			expectFlushed: true,
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
			handler := grpcweb.WrapHandler(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				nextSupport = testhelpers.SupportOf(rw)

				rw.WriteHeader(http.StatusOK)
				// The outcome is checked on the underlying recorder.
				_ = http.NewResponseController(rw).Flush()
			}), grpcweb.WithCorsForRegisteredEndpointsOnly(false))

			recorder := testhelpers.NewHijackableRecorder()
			var rw http.ResponseWriter = recorder
			if test.unwrapOnly {
				rw = &testhelpers.UnwrapOnlyResponseWriter{ResponseWriter: recorder}
			}

			req := httptest.NewRequest(http.MethodPost, "/", nil)
			req.Header.Set("Content-Type", "application/grpc-web+proto")

			handler.ServeHTTP(rw, req)

			assert.Equal(t, test.expectFlushed, recorder.Flushed)
			assert.Equal(t, testhelpers.ResponseWriterSupport{Flush: true}, nextSupport)
		})
	}
}
