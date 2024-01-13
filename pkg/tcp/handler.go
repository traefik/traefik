package tcp

import (
	"context"
	"net"
)

// Handler is the TCP Handlers interface.
type Handler interface {
	ServeTCP(ctx context.Context, conn WriteCloser)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as handlers.
type HandlerFunc func(ctx context.Context, conn WriteCloser)

// ServeTCP serves tcp.
func (f HandlerFunc) ServeTCP(ctx context.Context, conn WriteCloser) {
	f(ctx, conn)
}

// WriteCloser describes a net.Conn with a CloseWrite method.
type WriteCloser interface {
	net.Conn
	// CloseWrite on a network connection, indicates that the issuer of the call
	// has terminated sending on that connection.
	// It corresponds to sending a FIN packet.
	CloseWrite() error
}
