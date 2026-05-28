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
