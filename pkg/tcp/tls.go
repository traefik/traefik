package tcp

import (
	"context"
	"crypto/tls"
)

// TLSConn is a TLS connection that also carries the name of the TLS config used.
type TLSConn struct {
	WriteCloser

	TLSOptionsName string
}

// TLSHandler handles TLS connections.
type TLSHandler struct {
	Next           Handler
	Config         *tls.Config
	TLSOptionsName string
}

// ServeTCP terminates the TLS connection.
func (t *TLSHandler) ServeTCP(conn WriteCloser) {
	t.Next.ServeTCP(tls.Server(TLSConn{WriteCloser: conn, TLSOptionsName: t.TLSOptionsName}, t.Config))
}

type tlsOptionsNameKey struct{}

func AddTLSOptionsNameInContext(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, tlsOptionsNameKey{}, name)
}

func GetTLSOptionsName(ctx context.Context) string {
	if name, ok := ctx.Value(tlsOptionsNameKey{}).(string); ok {
		return name
	}

	return ""
}

type connCloserKey struct{}

// AddConnCloserInContext stores a function able to forcefully close the underlying connection.
// This is needed for transports, such as HTTP/3, where a "Connection: close" response header
// has no effect, so terminating a stale connection requires driving the transport directly.
func AddConnCloserInContext(ctx context.Context, closeConn func()) context.Context {
	return context.WithValue(ctx, connCloserKey{}, closeConn)
}

// CloseConn forcefully closes the connection carried by the given context, if any.
// It is a no-op for transports (such as plain TCP/TLS) that don't register a closer,
// since setting the "Connection: close" response header is sufficient there.
func CloseConn(ctx context.Context) {
	if closeConn, ok := ctx.Value(connCloserKey{}).(func()); ok {
		closeConn()
	}
}
