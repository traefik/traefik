package observability

import "context"

type routeContextKey struct{}

// WithHTTPRoute stores the matched router rule in the context for downstream consumers.
func WithHTTPRoute(ctx context.Context, route string) context.Context {
	if route == "" {
		return ctx
	}
	return context.WithValue(ctx, routeContextKey{}, route)
}

// HTTPRouteFromContext retrieves the matched router rule from the context when present.
func HTTPRouteFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	route, ok := ctx.Value(routeContextKey{}).(string)
	if !ok || route == "" {
		return "", false
	}

	return route, true
}
