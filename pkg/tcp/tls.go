package tcp

import (
	"context"
	"crypto/tls"
)

// TLSHandler handles TLS connections.
type TLSHandler struct {
	Next   Handler
	Config *tls.Config
}

// ServeTCP terminates the TLS connection.
func (t *TLSHandler) ServeTCP(ctx context.Context, conn WriteCloser) {
	t.Next.ServeTCP(ctx, tls.Server(conn, t.Config))
}
