package tcp

import "context"

// PluginMetadataKey is the key a Yaegi plugin passes to ctx.Value to get the connection metadata map.
// Plugins cannot import Traefik types, so the plugin bridge answers this string key
// instead of storing the map under it, which keeps Traefik's own context keys collision-free.
const PluginMetadataKey = "metadata"

type metadataKey struct{}

// WithMetadata returns a context that carries a mutable metadata map.
// An existing map is left in place: callers and the access log must share that object.
func WithMetadata(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if Metadata(ctx) != nil {
		return ctx
	}
	return context.WithValue(ctx, metadataKey{}, map[string]string{})
}

// Metadata returns the per-connection metadata map, or nil when the context has none.
func Metadata(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}
	metadata, _ := ctx.Value(metadataKey{}).(map[string]string)
	return metadata
}

// MarkBackend records that the connection reached a backend server.
// The access log treats a connection that never reaches a backend as rejected.
// Connections that do not implement the recorder ignore the call.
func MarkBackend(conn WriteCloser) {
	if conn == nil {
		return
	}
	if marker, ok := conn.(interface{ MarkBackend() }); ok {
		marker.MarkBackend()
	}
}

// SetRejection records why the connection was closed before reaching a backend server.
// Connections that do not implement the recorder ignore the call.
func SetRejection(conn WriteCloser, err error) {
	if err == nil || conn == nil {
		return
	}
	if setter, ok := conn.(interface{ SetRejection(err error) }); ok {
		setter.SetRejection(err)
	}
}
