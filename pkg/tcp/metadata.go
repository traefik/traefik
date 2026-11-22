package tcp

import "context"

// MetadataContextKey is the context key of the per-connection metadata map.
// The key is a plain string so Yaegi plugins, which cannot import Traefik types,
// read and update the same map with ctx.Value("metadata").
const MetadataContextKey = "metadata"

// WithMetadata returns a context that carries a mutable metadata map.
// An existing map stored under MetadataContextKey is left in place:
// callers and the access log must share that object.
func WithMetadata(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if Metadata(ctx) != nil {
		return ctx
	}
	return context.WithValue(ctx, MetadataContextKey, map[string]string{})
}

// Metadata returns the per-connection metadata map, or nil when the context has none.
func Metadata(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}
	metadata, _ := ctx.Value(MetadataContextKey).(map[string]string)
	return metadata
}

// MarkReached marks the connection as having reached the TCP service.
// The access log treats a connection that never reaches the service as rejected.
func MarkReached(next Handler) Handler {
	return HandlerFunc(func(ctx context.Context, conn WriteCloser) {
		if marker, ok := conn.(interface{ MarkBackend() }); ok {
			marker.MarkBackend()
		}
		next.ServeTCP(ctx, conn)
	})
}

// SetRejection records why a middleware closed the connection before the service.
// Connections that do not implement the recorder ignore the call.
func SetRejection(conn WriteCloser, err error) {
	if err == nil || conn == nil {
		return
	}
	if setter, ok := conn.(interface{ SetRejection(error) }); ok {
		setter.SetRejection(err)
	}
}
