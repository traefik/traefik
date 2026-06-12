package tcp

import (
	"context"
	"net"
)

// Handler is the TCP Handlers interface.
type Handler interface {
	ServeTCP(conn WriteCloser)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as handlers.
type HandlerFunc func(conn WriteCloser)

// ServeTCP serves tcp.
func (f HandlerFunc) ServeTCP(conn WriteCloser) {
	f(conn)
}

// WriteCloser describes a net.Conn with a CloseWrite method.
type WriteCloser interface {
	net.Conn
	// CloseWrite on a network connection, indicates that the issuer of the call
	// has terminated sending on that connection.
	// It corresponds to sending a FIN packet.
	CloseWrite() error
}

// LoadBalancer is the interface that TCP load balancers must implement.
type LoadBalancer interface {
	Handler
	Add(name string, handler Handler, weight *int)
	SetStatus(ctx context.Context, childName string, up bool)
	RegisterStatusUpdater(fn func(up bool)) error
}
