package tcp

import (
	"net"
)

// Handler is the TCP Handlers interface
type Handler interface {
	ServeTCP(conn net.Conn)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as handlers.
type HandlerFunc func(conn net.Conn)

// ServeTCP serves tcp
func (f HandlerFunc) ServeTCP(conn net.Conn) {
	f(conn)
}
