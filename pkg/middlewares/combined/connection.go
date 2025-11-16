package combined

import (
	"net"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/tcp"
)

// Connection abstracts HTTP and TCP connections for protocol-agnostic middlewares.
type Connection interface {
	// RemoteAddr returns the remote address of the connection.
	RemoteAddr() net.Addr
	// Close closes the connection.
	Close()
	// Proceed passes the connection to the next handler.
	Proceed()
}

// HTTPConnection wraps an HTTP request/response.
type HTTPConnection struct {
	rw   http.ResponseWriter
	req  *http.Request
	next http.Handler
}

// NewHTTPConnection creates a Connection from HTTP request/response.
func NewHTTPConnection(rw http.ResponseWriter, req *http.Request, next http.Handler) Connection {
	return &HTTPConnection{
		rw:   rw,
		req:  req,
		next: next,
	}
}

func (h *HTTPConnection) RemoteAddr() net.Addr {
	// Parse RemoteAddr string to net.Addr
	addr, err := net.ResolveTCPAddr("tcp", h.req.RemoteAddr)
	if err != nil {
		// Fallback to a simple wrapper if parsing fails
		return &simpleAddr{addr: h.req.RemoteAddr}
	}
	return addr
}

func (h *HTTPConnection) Close() {
	// For HTTP, we can't really "close" the connection mid-request
	// Just send a 403 Forbidden response
	http.Error(h.rw, "Forbidden", http.StatusForbidden)
}

func (h *HTTPConnection) Proceed() {
	h.next.ServeHTTP(h.rw, h.req)
}

// TCPConnection wraps a TCP connection.
type TCPConnection struct {
	conn tcp.WriteCloser
	next tcp.Handler
}

// NewTCPConnection creates a Connection from a TCP connection.
func NewTCPConnection(conn tcp.WriteCloser, next tcp.Handler) Connection {
	return &TCPConnection{
		conn: conn,
		next: next,
	}
}

func (t *TCPConnection) RemoteAddr() net.Addr {
	return t.conn.RemoteAddr()
}

func (t *TCPConnection) Close() {
	t.conn.Close()
}

func (t *TCPConnection) Proceed() {
	t.next.ServeTCP(t.conn)
}

// simpleAddr is a simple net.Addr implementation for fallback.
type simpleAddr struct {
	addr string
}

func (s *simpleAddr) Network() string {
	return "tcp"
}

func (s *simpleAddr) String() string {
	return s.addr
}
