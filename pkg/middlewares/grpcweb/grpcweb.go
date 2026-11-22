package grpcweb

import (
	"context"
	"net/http"
	"strings"

	"github.com/traefik/grpc-web/go/grpcweb"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const typeName = "GRPCWeb"

const (
	// emptyTrailerFrame is the frame header of a gRPC-Web trailer frame with a
	// zero-length payload: the MSB of the first byte marks a trailer frame and
	// the next four bytes encode the payload length.
	emptyTrailerFrame = "\x80\x00\x00\x00\x00"
	// emptyTrailerFrameBase64 is the base64 encoding of emptyTrailerFrame, as
	// written for grpc-web-text responses.
	emptyTrailerFrameBase64 = "gAAAAAA="
	// maxHeldWrite is the size under which a write is held back by the guard,
	// which only needs to cover the empty trailer frame of the conversion and
	// its base64-encoded writes.
	maxHeldWrite = len(emptyTrailerFrameBase64)
	// maxPendingBytes bounds the bytes held back by the guard; it only exists
	// as a safety net against pathological writers of many small writes.
	maxPendingBytes = 1024
)

// New builds a new gRPC web request converter.
func New(ctx context.Context, next http.Handler, config dynamic.GrpcWeb, name string) http.Handler {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	return &grpcWeb{
		handler: grpcweb.WrapHandler(next, grpcweb.WithCorsForRegisteredEndpointsOnly(false), grpcweb.WithOriginFunc(func(origin string) bool {
			for _, originCfg := range config.AllowOrigins {
				if originCfg == "*" || originCfg == origin {
					return true
				}
			}
			return false
		})),
	}
}

type grpcWeb struct {
	handler http.Handler
}

func (g *grpcWeb) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if needsTrailerFrameGuard(req) {
		guard := &emptyTrailerFrameGuard{wrapped: rw}
		// Ensures bytes held back are not lost if the conversion does not flush.
		defer guard.forwardPending()

		g.handler.ServeHTTP(guard, req)
		return
	}

	g.handler.ServeHTTP(rw, req)
}

// needsTrailerFrameGuard mirrors the gRPC-Web request detection of
// github.com/traefik/grpc-web: the guard is only useful for requests that the
// conversion rewrites, while passthrough responses must stay untouched.
func needsTrailerFrameGuard(req *http.Request) bool {
	return req.Method == http.MethodPost && strings.HasPrefix(req.Header.Get("Content-Type"), "application/grpc-web")
}

// emptyTrailerFrameGuard is an http.ResponseWriter wrapper dropping the empty
// trailer frame that github.com/traefik/grpc-web appends to trailers-only
// upstream responses (https://github.com/traefik/traefik/issues/13842).
//
// The conversion always writes its trailer frame right before the wrapped
// handler returns, so the frame is the only content of the last flush cycle;
// small writes are held back until the next write or flush, and dropped when
// they turn out to be exactly that empty frame. A trailer frame without
// payload carries no information, and PROTOCOL-WEB prescribes headers only for
// trailers-only responses, which is what every other gRPC-Web proxy emits.
type emptyTrailerFrameGuard struct {
	wrapped http.ResponseWriter

	// pending holds the small writes not yet forwarded to wrapped.
	pending []byte
}

func (g *emptyTrailerFrameGuard) Header() http.Header {
	return g.wrapped.Header()
}

func (g *emptyTrailerFrameGuard) WriteHeader(statusCode int) {
	g.wrapped.WriteHeader(statusCode)
}

func (g *emptyTrailerFrameGuard) Write(p []byte) (int, error) {
	if len(p) == 0 {
		// Zero-length writes carry nothing (net/http ignores them as well), but
		// they must not evict the held back empty trailer frame.
		return 0, nil
	}

	if len(p) > maxHeldWrite || len(g.pending)+len(p) > maxPendingBytes {
		g.forwardPending()
		return g.wrapped.Write(p)
	}

	g.pending = append(g.pending, p...)
	return len(p), nil
}

// Flush implements http.Flusher.
func (g *emptyTrailerFrameGuard) Flush() {
	_ = g.FlushError()
}

// FlushError flushes the wrapped writer and returns its error, for http.ResponseController.
func (g *emptyTrailerFrameGuard) FlushError() error {
	if g.forwardPending() {
		// The held back bytes were the empty trailer frame; skipping the flush
		// lets net/http finalize the response as headers only, instead of
		// announcing a chunked body that stays empty.
		return nil
	}

	return http.NewResponseController(g.wrapped).Flush()
}

// Unwrap returns the wrapped http.ResponseWriter, for http.ResponseController
// and for the websocket upgrade, which follows the Unwrap chain to hijack the
// underlying connection.
func (g *emptyTrailerFrameGuard) Unwrap() http.ResponseWriter {
	return g.wrapped
}

// forwardPending writes the held back bytes to the wrapped writer, unless they
// are exactly an empty trailer frame, in which case they are dropped. It
// reports whether bytes were dropped.
func (g *emptyTrailerFrameGuard) forwardPending() bool {
	if len(g.pending) == 0 {
		return false
	}

	dropped := string(g.pending) == emptyTrailerFrame || string(g.pending) == emptyTrailerFrameBase64
	if !dropped {
		_, _ = g.wrapped.Write(g.pending)
	}
	g.pending = g.pending[:0]

	return dropped
}
