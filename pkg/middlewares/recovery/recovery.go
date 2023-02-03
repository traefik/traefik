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
	defer recoverFunc(rw, req)
	re.next.ServeHTTP(rw, req)
}

func recoverFunc(rw http.ResponseWriter, req *http.Request) {
	if err := recover(); err != nil {
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

		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// https://github.com/golang/go/blob/a0d6420d8be2ae7164797051ec74fa2a2df466a1/src/net/http/server.go#L1761-L1775
// https://github.com/golang/go/blob/c33153f7b416c03983324b3e8f869ce1116d84bc/src/net/http/httputil/reverseproxy.go#L284
func shouldLogPanic(panicValue interface{}) bool {
	//nolint:errorlint // false-positive because panicValue is an interface.
	return panicValue != nil && panicValue != http.ErrAbortHandler
}
