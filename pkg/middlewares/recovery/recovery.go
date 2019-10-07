package recovery

import (
	"context"
	"net/http"

	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
)

const (
	typeName = "Recovery"
)

type recovery struct {
	next http.Handler
	name string
}

// New creates recovery middleware.
func New(ctx context.Context, next http.Handler, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

	return &recovery{
		next: next,
		name: name,
	}, nil
}

func (re *recovery) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer recoverFunc(middlewares.GetLoggerCtx(req.Context(), re.name, typeName), rw)
	re.next.ServeHTTP(rw, req)
}

func recoverFunc(ctx context.Context, rw http.ResponseWriter) {
	if err := recover(); err != nil {
		log.FromContext(ctx).Errorf("Recovered from panic in http handler: %+v", err)
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
