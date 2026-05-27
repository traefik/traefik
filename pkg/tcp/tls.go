package tcp

import (
	"context"
	"crypto/tls"
)

// TLSHandler handles TLS connections.
type TLSHandler struct {
	Next       Handler
	Config     *tls.Config
	ConfigName string
}

// ServeTCP terminates the TLS connection.
func (t *TLSHandler) ServeTCP(conn WriteCloser) {
	t.Next.ServeTCP(tls.Server(TLSConnWithOptionsName{WriteCloser: conn, ConfigName: t.ConfigName}, t.Config))
}

// TLSConnWithOptionsName is a TLS connection that also carries the name of the TLS config used.
type TLSConnWithOptionsName struct {
	WriteCloser

	ConfigName string
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
