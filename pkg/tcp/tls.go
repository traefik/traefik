package tcp

import (
	"context"
	"crypto/tls"
	"sync/atomic"
)

// TLSConn is a TLS connection that also carries the name of the TLS options used.
type TLSConn struct {
	WriteCloser

	tlsOptionsName atomic.Pointer[string]
}

// TLSOptionsName returns the name of the TLS options applied to the connection.
func (c *TLSConn) TLSOptionsName() string {
	return *c.tlsOptionsName.Load()
}

// SetTLSOptionsName records the TLS options selected during the handshake:
// for ECH connections the effective options are only known after decryption.
func (c *TLSConn) SetTLSOptionsName(name string) {
	c.tlsOptionsName.Store(&name)
}

// TLSHandler handles TLS connections.
type TLSHandler struct {
	Next           Handler
	Config         *tls.Config
	TLSOptionsName string
}

// ServeTCP terminates the TLS connection.
func (t *TLSHandler) ServeTCP(conn WriteCloser) {
	tlsConn := &TLSConn{WriteCloser: conn}
	tlsConn.SetTLSOptionsName(t.TLSOptionsName)

	t.Next.ServeTCP(tls.Server(tlsConn, t.Config))
}

type tlsOptionsNameKey struct{}

func AddTLSOptionsNameInContext(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, tlsOptionsNameKey{}, name)
}

// AddTLSConnInContext stores the connection itself, so that the TLS options
// name is read after the handshake, once ECH may have updated it.
func AddTLSConnInContext(ctx context.Context, conn *TLSConn) context.Context {
	return context.WithValue(ctx, tlsOptionsNameKey{}, conn)
}

func GetTLSOptionsName(ctx context.Context) string {
	switch value := ctx.Value(tlsOptionsNameKey{}).(type) {
	case string:
		return value
	case *TLSConn:
		return value.TLSOptionsName()
	}

	return ""
}
