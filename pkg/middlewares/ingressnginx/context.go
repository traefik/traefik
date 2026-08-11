package ingressnginx

import "context"

type ingressNginxContextKey string

const originalHostKey ingressNginxContextKey = "ingress-nginx-original-host"

// WithOriginalHost returns a context with the original request host saved.
// This is used by the upstream-vhost middleware to preserve the original
// client-facing hostname before overwriting req.Host with the upstream target.
func WithOriginalHost(ctx context.Context, host string) context.Context {
	return context.WithValue(ctx, originalHostKey, host)
}

// OriginalHost returns the original request host saved in the context, if any.
func OriginalHost(ctx context.Context) (string, bool) {
	host, ok := ctx.Value(originalHostKey).(string)
	return host, ok
}
