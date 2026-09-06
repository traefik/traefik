package udp

import "net"

// Handler is the UDP counterpart of the usual HTTP handler.
type Handler interface {
	ServeUDP(conn WriteCloser)
}

// The HandlerFunc type is an adapter to allow the use of ordinary functions as handlers.
type HandlerFunc func(conn WriteCloser)

// ServeUDP implements the Handler interface for UDP.
func (f HandlerFunc) ServeUDP(conn WriteCloser) {
	f(conn)
}

// WriteCloser is the minimal interface UDP handlers need, a subset of net.Conn, since udp.Conn is a synthetic per-session type with no deadline semantics to expose.
type WriteCloser interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
	RemoteAddr() net.Addr
}
