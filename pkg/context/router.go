package context

import "context"

type contextKey int

const (
	routerKey contextKey = iota
	serviceKey
)

// AddRouterInContext adds the router name to the context.
func AddRouterInContext(ctx context.Context, routerName string) context.Context {
	return context.WithValue(ctx, routerKey, routerName)
}

// GetRouterName gets the router name from the context.
func GetRouterName(ctx context.Context) (string, bool) {
	routerName, ok := ctx.Value(routerKey).(string)
	return routerName, ok
}

// AddServiceInContext adds the service name to the context.
func AddServiceInContext(ctx context.Context, serviceName string) context.Context {
	return context.WithValue(ctx, serviceKey, serviceName)
}

// GetServiceName gets the service name from the context.
func GetServiceName(ctx context.Context) (string, bool) {
	serviceName, ok := ctx.Value(serviceKey).(string)
	return serviceName, ok
}
