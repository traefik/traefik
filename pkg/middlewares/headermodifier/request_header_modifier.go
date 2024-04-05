package headermodifier

import (
	"context"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"go.opentelemetry.io/otel/trace"
)

const typeName = "RequestHeaderModifier"

// requestHeaderModifier is a middleware used to modify the headers of an HTTP request.
type requestHeaderModifier struct {
	next http.Handler
	name string

	set    map[string]string
	add    map[string]string
	remove []string
}

// NewRequestHeaderModifier creates a new request header modifier middleware.
func NewRequestHeaderModifier(ctx context.Context, next http.Handler, config dynamic.RequestHeaderModifier, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	return &requestHeaderModifier{
		next:   next,
		name:   name,
		set:    config.Set,
		add:    config.Add,
		remove: config.Remove,
	}, nil
}

func (r *requestHeaderModifier) GetTracingInformation() (string, string, trace.SpanKind) {
	return r.name, typeName, trace.SpanKindUnspecified
}

func (r *requestHeaderModifier) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for headerName, headerValue := range r.set {
		req.Header.Set(headerName, headerValue)
	}

	for headerName, headerValue := range r.add {
		req.Header.Add(headerName, headerValue)
	}

	for _, headerName := range r.remove {
		req.Header.Del(headerName)
	}

	r.next.ServeHTTP(rw, req)
}
