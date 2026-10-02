package ingressnginx

import (
	"context"
	"net/http"
)

type originalPathContextKey struct{}

// WithOriginalPath stores the request path before an ingress-nginx rewrite.
// It preserves the first value when more than one rewrite middleware is used.
func WithOriginalPath(ctx context.Context, path string) context.Context {
	if _, ok := ctx.Value(originalPathContextKey{}).(string); ok {
		return ctx
	}

	return context.WithValue(ctx, originalPathContextKey{}, path)
}

// OriginalPath returns the path before an ingress-nginx rewrite, if one ran.
func OriginalPath(req *http.Request) string {
	if path, ok := req.Context().Value(originalPathContextKey{}).(string); ok {
		return path
	}

	return req.URL.Path
}
