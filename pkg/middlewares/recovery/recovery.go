package recovery

import (
	"context"
	"net/http"
	"runtime"

	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
)

const (
	typeName       = "Recovery"
	middlewareName = "traefik-internal-recovery"
)

type recovery struct {
	next http.Handler
}

type responseWriterWithHeaderStatus struct {
	http.ResponseWriter
	headerWritten bool
}

func (w *responseWriterWithHeaderStatus) WriteHeader(statusCode int) {
	if !w.headerWritten {
		w.headerWritten = true
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func (w *responseWriterWithHeaderStatus) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

func (w *responseWriterWithHeaderStatus) Header() http.Header {
	return w.ResponseWriter.Header()
}

// New creates recovery middleware.
func New(ctx context.Context, next http.Handler) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, middlewareName, typeName)).Debug("Creating middleware")

	return &recovery{
		next: next,
	}, nil
}

func (re *recovery) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	rws := &responseWriterWithHeaderStatus{ResponseWriter: rw}
	defer recoverFunc(rws, req)
	re.next.ServeHTTP(rw, req)
}

func recoverFunc(rws *responseWriterWithHeaderStatus, req *http.Request) {
	if err := recover(); err != nil {
		logger := log.FromContext(middlewares.GetLoggerCtx(req.Context(), middlewareName, typeName))
		if shouldLogPanic(err) {
			logger.Errorf("Recovered from panic in HTTP handler [%s - %s]: %+v", req.RemoteAddr, req.URL, err)
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			logger.Errorf("Stack: %s", buf)
		} else {
			logger.Debugf("Request has been aborted [%s - %s]: %v", req.RemoteAddr, req.URL, err)
		}

		if rws.headerWritten {
			// headers already sent, need to close the connection
			// not sure how to access the underlying connection from the response writer
			// I don't think we can use Hijack here because that would close the TCP socket, what about HTTP/2?
			panic(http.ErrAbortHandler)
		} else {
			// headers not sent, send error response
			http.Error(rws, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
}

// https://github.com/golang/go/blob/a0d6420d8be2ae7164797051ec74fa2a2df466a1/src/net/http/server.go#L1761-L1775
// https://github.com/golang/go/blob/c33153f7b416c03983324b3e8f869ce1116d84bc/src/net/http/httputil/reverseproxy.go#L284
func shouldLogPanic(panicValue interface{}) bool {
	//nolint:errorlint // false-positive because panicValue is an interface.
	return panicValue != nil && panicValue != http.ErrAbortHandler
}
