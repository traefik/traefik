package recovery

import (
	"context"
	"net/http"
	"runtime"

	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const (
	typeName       = "Recovery"
	middlewareName = "traefik-internal-recovery"
)

type recovery struct {
	next http.Handler
}

// New creates recovery middleware.
func New(ctx context.Context, next http.Handler) (http.Handler, error) {
	middlewares.GetLogger(ctx, middlewareName, typeName).Debug().Msg("Creating middleware")

	return &recovery{
		next: next,
	}, nil
}

func (re *recovery) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	recoveryRW := &responseWriterWrapper{rw: rw}
	defer recoverFunc(recoveryRW, req)

	re.next.ServeHTTP(recoveryRW, req)
}

func recoverFunc(rw *responseWriterWrapper, req *http.Request) {
	if err := recover(); err != nil {
		defer rw.finalizeResponse()

		logger := middlewares.GetLogger(req.Context(), middlewareName, typeName)
		if !shouldLogPanic(err) {
			logger.Debug().Msgf("Request has been aborted [%s - %s]: %v", req.RemoteAddr, req.URL, err)
			return
		}

		logger.Error().Msgf("Recovered from panic in HTTP handler [%s - %s]: %+v", req.RemoteAddr, req.URL, err)
		const size = 64 << 10
		buf := make([]byte, size)
		buf = buf[:runtime.Stack(buf, false)]
		logger.Error().Msgf("Stack: %s", buf)
	}
}

// https://github.com/golang/go/blob/a0d6420d8be2ae7164797051ec74fa2a2df466a1/src/net/http/server.go#L1761-L1775
// https://github.com/golang/go/blob/c33153f7b416c03983324b3e8f869ce1116d84bc/src/net/http/httputil/reverseproxy.go#L284
func shouldLogPanic(panicValue any) bool {
	//nolint:errorlint // false-positive because panicValue is an interface.
	return panicValue != nil && panicValue != http.ErrAbortHandler
}

type responseWriterWrapper struct {
	rw          http.ResponseWriter
	headersSent bool
}

func (r *responseWriterWrapper) Header() http.Header {
	return r.rw.Header()
}

func (r *responseWriterWrapper) Write(bytes []byte) (int, error) {
	r.headersSent = true
	return r.rw.Write(bytes)
}

func (r *responseWriterWrapper) WriteHeader(code int) {
	if r.headersSent {
		return
	}

	// Handling informational headers.
	if code >= 100 && code <= 199 {
		r.rw.WriteHeader(code)
		return
	}

	r.headersSent = true
	r.rw.WriteHeader(code)
}

// Unwrap exposes the underlying ResponseWriter for http.ResponseController.
func (r *responseWriterWrapper) Unwrap() http.ResponseWriter {
	return r.rw
}

func (r *responseWriterWrapper) finalizeResponse() {
	// If headers have been sent this is not possible to respond with an HTTP error,
	// and we let the server abort the response silently thanks to the http.ErrAbortHandler sentinel panic value.
	if r.headersSent {
		panic(http.ErrAbortHandler)
	}

	// The response has not yet started to be written,
	// we can safely return a fresh new error response.
	http.Error(r.rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
