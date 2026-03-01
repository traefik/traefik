package headermodifier

import (
	"context"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const responseHeaderModifierTypeName = "ResponseHeaderModifier"

// requestHeaderModifier is a middleware used to modify the headers of an HTTP response.
type responseHeaderModifier struct {
	next http.Handler
	name string

	set    map[string]string
	add    map[string]string
	remove []string
}

// NewResponseHeaderModifier creates a new response header modifier middleware.
func NewResponseHeaderModifier(ctx context.Context, next http.Handler, config dynamic.HeaderModifier, name string) http.Handler {
	logger := middlewares.GetLogger(ctx, name, responseHeaderModifierTypeName)
	logger.Debug().Msg("Creating middleware")

	return &responseHeaderModifier{
		next:   next,
		name:   name,
		set:    config.Set,
		add:    config.Add,
		remove: config.Remove,
	}
}

func (r *responseHeaderModifier) GetTracingInformation() (string, string) {
	return r.name, responseHeaderModifierTypeName
}

func (r *responseHeaderModifier) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.next.ServeHTTP(middlewares.NewResponseModifier(rw, req, r.modifyResponseHeaders), req)
}

func (r *responseHeaderModifier) modifyResponseHeaders(res *http.Response) error {
	for headerName, headerValue := range r.set {
		res.Header.Set(headerName, headerValue)
	}

	for headerName, headerValue := range r.add {
		res.Header.Add(headerName, headerValue)
	}

	for _, headerName := range r.remove {
		res.Header.Del(headerName)
	}

	return nil
}
