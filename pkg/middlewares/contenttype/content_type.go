package contenttype

import (
	"context"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const (
	typeName = "ContentType"
)

// ContentType is a middleware used to activate Content-Type auto-detection.
type contentType struct {
	next http.Handler
	name string
}

// New creates a new handler.
func New(ctx context.Context, next http.Handler, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")
	return &contentType{next: next, name: name}, nil
}

func (c *contentType) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Re-enable auto-detection.
	if ct, ok := rw.Header()["Content-Type"]; ok && ct == nil {
		middlewares.GetLogger(req.Context(), c.name, typeName).
			Debug().Msg("Enable Content-Type auto-detection.")
		delete(rw.Header(), "Content-Type")
	}

	c.next.ServeHTTP(rw, req)
}

func DisableAutoDetection(next http.Handler) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Prevent Content-Type auto-detection.
		if _, ok := rw.Header()["Content-Type"]; !ok {
			rw.Header()["Content-Type"] = nil
		}

		next.ServeHTTP(rw, req)
	}
}
