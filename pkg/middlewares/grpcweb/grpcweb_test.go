package grpcweb_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares/grpcweb"
)

// emptyTrailerFrame is the gRPC-Web trailer frame header announcing zero-length
// trailers: the MSB of the first byte marks a trailer frame, and the following
// four bytes encode the payload length (0).
var emptyTrailerFrame = []byte{1 << 7, 0, 0, 0, 0}

// trailersOnlyBackend mimics what httputil.ReverseProxy writes for a gRPC
// trailers-only upstream response: the fields of the single HEADERS frame
// (carrying grpc-status with END_STREAM) arrive as plain response headers, and
// no HTTP/2 trailers follow.
func trailersOnlyBackend(status, message string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		h := rw.Header()
		h.Set("Content-Type", "application/grpc")
		h.Set("Grpc-Status", status)
		h.Set("Grpc-Message", message)
		rw.WriteHeader(http.StatusOK)
	})
}

// grpcBackend mimics what httputil.ReverseProxy writes for a regular gRPC
// upstream response: response headers, one length-prefixed message frame in
// the body, then the actual trailers (like ReverseProxy does through the
// http.TrailerPrefix mechanism).
func grpcBackend() http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		h := rw.Header()
		h.Set("Content-Type", "application/grpc")
		rw.WriteHeader(http.StatusOK)

		// One message frame carrying the 5 bytes payload "hello".
		_, _ = rw.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x05, 'h', 'e', 'l', 'l', 'o'})

		rw.Header().Set(http.TrailerPrefix+"Grpc-Status", "0")
	})
}

func TestGRPCWebTrailersOnlyResponse(t *testing.T) {
	testCases := []struct {
		desc           string
		reqContentType string
		reqBody        []byte
	}{
		{
			desc:           "binary format",
			reqContentType: "application/grpc-web+proto",
			reqBody:        []byte{0x00, 0x00, 0x00, 0x00, 0x00}, // One empty message frame.
		},
		{
			desc:           "text format",
			reqContentType: "application/grpc-web-text",
			reqBody:        []byte("AAAAAAA="), // base64 of one empty message frame.
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			handler := grpcweb.New(t.Context(), trailersOnlyBackend("3", "trailers-only response"), dynamic.GrpcWeb{AllowOrigins: []string{"*"}}, "grpc-web")

			req := httptest.NewRequest(http.MethodPost, "/pkg.Service/Method", bytes.NewReader(test.reqBody))
			req.Header.Set("Content-Type", test.reqContentType)

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			// The grpc-status of a trailers-only response is exposed as response headers.
			assert.Equal(t, http.StatusOK, rec.Code)
			assert.Equal(t, "3", rec.Header().Get("Grpc-Status"))
			assert.Equal(t, "trailers-only response", rec.Header().Get("Grpc-Message"))

			// The body must be empty: an empty trailer frame carries no information and
			// PROTOCOL-WEB prescribes headers only for trailers-only responses.
			assert.Empty(t, rec.Body.Bytes())
		})
	}
}

func TestGRPCWebRegularResponse(t *testing.T) {
	handler := grpcweb.New(t.Context(), grpcBackend(), dynamic.GrpcWeb{AllowOrigins: []string{"*"}}, "grpc-web")

	req := httptest.NewRequest(http.MethodPost, "/pkg.Service/Method", bytes.NewReader([]byte{0x00, 0x00, 0x00, 0x00, 0x00}))
	req.Header.Set("Content-Type", "application/grpc-web+proto")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.Bytes()

	// The message frame must be forwarded untouched.
	messageFrame := []byte{0x00, 0x00, 0x00, 0x00, 0x05, 'h', 'e', 'l', 'l', 'o'}
	require.True(t, bytes.HasPrefix(body, messageFrame))

	// The response must still end with a trailer frame carrying the grpc-status.
	trailerFrame := body[len(messageFrame):]
	require.NotEmpty(t, trailerFrame)
	assert.Equal(t, byte(1<<7), trailerFrame[0], "trailer frame should start with the trailer flag")
	assert.Contains(t, string(trailerFrame), "grpc-status: 0")
}

func TestGRPCWebNonGRPCWebRequestIsPassedThrough(t *testing.T) {
	// A body ending with the exact empty trailer frame pattern must survive a
	// passthrough response untouched.
	body := append([]byte("some binary payload"), emptyTrailerFrame...)

	handler := grpcweb.New(t.Context(), http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/octet-stream")
		rw.WriteHeader(http.StatusOK)
		_, err := rw.Write(body)
		require.NoError(t, err)
	}), dynamic.GrpcWeb{AllowOrigins: []string{"*"}}, "grpc-web")

	req := httptest.NewRequest(http.MethodGet, "/file", nil)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, body, rec.Body.Bytes())
}

func TestGRPCWebFlushedMessageStream(t *testing.T) {
	handler := grpcweb.New(t.Context(), http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/grpc")
		rw.WriteHeader(http.StatusOK)

		// Two message frames, flushed separately like a streaming RPC.
		_, err := rw.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x01, 'a'})
		require.NoError(t, err)
		http.NewResponseController(rw).Flush()

		_, err = rw.Write([]byte{0x00, 0x00, 0x00, 0x00, 0x01, 'b'})
		require.NoError(t, err)
		http.NewResponseController(rw).Flush()

		rw.Header().Set(http.TrailerPrefix+"Grpc-Status", "0")
	}), dynamic.GrpcWeb{AllowOrigins: []string{"*"}}, "grpc-web")

	req := httptest.NewRequest(http.MethodPost, "/pkg.Service/Method", bytes.NewReader([]byte{0x00, 0x00, 0x00, 0x00, 0x00}))
	req.Header.Set("Content-Type", "application/grpc-web+proto")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.Bytes()

	// Every message frame must survive, in order, whether or not it was held
	// back before being forwarded.
	assert.Contains(t, string(body), "a")
	assert.Contains(t, string(body), "b")
	assert.True(t, bytes.HasSuffix(body, []byte("grpc-status: 0\r\n")))
}
