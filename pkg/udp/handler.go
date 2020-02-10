package udp

// Handler is the UDP counterpart of the usual HTTP handler.
type Handler interface {
	ServeUDP(conn *Conn)
}

// The HandlerFunc type is an adapter to allow the use of ordinary functions as handlers.
type HandlerFunc func(conn *Conn)

// ServeUDP implements the Handler interface for UDP.
func (f HandlerFunc) ServeUDP(conn *Conn) {
	f(conn)
}
