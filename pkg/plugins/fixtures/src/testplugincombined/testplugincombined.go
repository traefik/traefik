package testplugincombined

import (
	"context"
	"net"
	"net/http"
)

// Config holds the plugin configuration.
type Config struct {
	// AllowedIPPrefix restricts connections to IPs starting with this prefix
	AllowedIPPrefix string
	// HTTPHeaderValue sets a custom header for HTTP requests
	HTTPHeaderValue string
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		AllowedIPPrefix: "127",
		HTTPHeaderValue: "combined-plugin",
	}
}

// Connection abstracts both HTTP and TCP connections.
type Connection interface {
	// GetRemoteIP returns the client IP address
	GetRemoteIP() string
	// Reject rejects the connection/request
	Reject()
	// SetMetadata sets protocol-specific metadata (e.g., HTTP headers)
	SetMetadata(key, value string)
	// Proceed passes to the next handler
	Proceed()
}

// Middleware contains the core protocol-agnostic logic.
type Middleware struct {
	config *Config
}

// Serve is the unified logic that works for both HTTP and TCP.
func (m *Middleware) Serve(conn Connection) {
	// Extract IP
	ip := conn.GetRemoteIP()

	// Check if IP is allowed
	if m.config.AllowedIPPrefix != "" && len(ip) >= len(m.config.AllowedIPPrefix) {
		if ip[:len(m.config.AllowedIPPrefix)] != m.config.AllowedIPPrefix {
			// Reject connection (HTTP: 403, TCP: close)
			conn.Reject()
			return
		}
	}

	// Set metadata (HTTP: headers, TCP: no-op)
	conn.SetMetadata("X-Combined-Plugin", m.config.HTTPHeaderValue)
	conn.SetMetadata("X-Client-IP", ip)

	// Proceed to next handler
	conn.Proceed()
}

// New creates a new HTTP middleware instance.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &httpHandler{
		middleware: &Middleware{config: config},
		next:       next,
	}, nil
}

// httpHandler wraps the unified middleware for HTTP.
type httpHandler struct {
	middleware *Middleware
	next       http.Handler
}

func (h *httpHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Wrap HTTP request/response in unified Connection
	conn := &HTTPConnection{
		rw:   rw,
		req:  req,
		next: h.next,
	}

	// Use unified middleware logic
	h.middleware.Serve(conn)
}

// HTTPConnection implements Connection for HTTP requests.
type HTTPConnection struct {
	rw   http.ResponseWriter
	req  *http.Request
	next http.Handler
}

func (h *HTTPConnection) GetRemoteIP() string {
	host, _, _ := net.SplitHostPort(h.req.RemoteAddr)
	return host
}

func (h *HTTPConnection) Reject() {
	http.Error(h.rw, "Forbidden", http.StatusForbidden)
}

func (h *HTTPConnection) SetMetadata(key, value string) {
	h.rw.Header().Set(key, value)
}

func (h *HTTPConnection) Proceed() {
	h.next.ServeHTTP(h.rw, h.req)
}

// TCPHandler is the interface that TCP handlers must implement.
type TCPHandler interface {
	ServeTCP(conn TCPWriteCloser)
}

// TCPWriteCloser is the interface for TCP connections.
type TCPWriteCloser interface {
	net.Conn
	CloseWrite() error
}

// NewTCP creates a new TCP middleware instance.
func NewTCP(ctx context.Context, next TCPHandler, config *Config, name string) (TCPHandler, error) {
	return &tcpHandler{
		middleware: &Middleware{config: config},
		next:       next,
	}, nil
}

// tcpHandler wraps the unified middleware for TCP.
type tcpHandler struct {
	middleware *Middleware
	next       TCPHandler
}

func (t *tcpHandler) ServeTCP(conn TCPWriteCloser) {
	// Wrap TCP connection in unified Connection
	wrappedConn := &TCPConnection{
		conn: conn,
		next: t.next,
	}

	// Use unified middleware logic
	t.middleware.Serve(wrappedConn)
}

// TCPConnection implements Connection for TCP connections.
type TCPConnection struct {
	conn TCPWriteCloser
	next TCPHandler
}

func (t *TCPConnection) GetRemoteIP() string {
	addr := t.conn.RemoteAddr().String()
	host, _, _ := net.SplitHostPort(addr)
	return host
}

func (t *TCPConnection) Reject() {
	t.conn.Close()
}

func (t *TCPConnection) SetMetadata(key, value string) {
	// TCP has no headers, so this is a no-op
	// In a real implementation, you might log this
}

func (t *TCPConnection) Proceed() {
	t.next.ServeTCP(t.conn)
}
