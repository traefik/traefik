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

// New creates recovery middleware.
func New(ctx context.Context, next http.Handler) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, middlewareName, typeName)).Debug("Creating middleware")

	return &recovery{
		next: next,
	}, nil
}

func (re *recovery) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer recoverFunc(rw, req)
	re.next.ServeHTTP(rw, req)
}

func recoverFunc(rw http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		logger := log.FromContext(middlewares.GetLoggerCtx(r.Context(), middlewareName, typeName))
		if !shouldLogPanic(err) {
			logger.Debugf("Request has been aborted [%s - %s]: %v", r.RemoteAddr, r.URL, err)
			return
		}

		logger.Errorf("Recovered from panic in HTTP handler [%s - %s]: %+v", r.RemoteAddr, r.URL, err)
		const size = 64 << 10
		buf := make([]byte, size)
		buf = buf[:runtime.Stack(buf, false)]
		logger.Errorf("Stack: %s", buf)

		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// https://github.com/golang/go/blob/a0d6420d8be2ae7164797051ec74fa2a2df466a1/src/net/http/server.go#L1761-L1775
// https://github.com/golang/go/blob/c33153f7b416c03983324b3e8f869ce1116d84bc/src/net/http/httputil/reverseproxy.go#L284
func shouldLogPanic(panicValue interface{}) bool {
	//nolint:errorlint // false-positive because panicValue is an interface.
	return panicValue != nil && panicValue != http.ErrAbortHandler
}
