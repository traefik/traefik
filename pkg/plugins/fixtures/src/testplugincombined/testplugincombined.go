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
	// SetMetadata sets protocol-specific metadata (HTTP: headers, TCP: context)
	SetMetadata(key, value string)
	// Proceed passes to the next handler
	Proceed()
}

// contextKey is the type for context keys to avoid collisions.
type contextKey string

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
	// For HTTP: set as response header
	h.rw.Header().Set(key, value)
}

func (h *HTTPConnection) Proceed() {
	h.next.ServeHTTP(h.rw, h.req)
}

// TCPHandler is the interface that TCP handlers must implement.
// Uses only stdlib types (net.Conn) to avoid yaegi boundary issues.
type TCPHandler interface {
	ServeTCP(ctx context.Context, conn net.Conn, closeWrite func() error)
}

// NewTCP creates a new TCP middleware instance.
// Parameters AND return use only stdlib types!
// Returns a callback function instead of an interface to avoid yaegi wrapping
func NewTCP(
	ctx context.Context,
	next func(ctx context.Context, conn net.Conn, closeWrite func() error),
	config *Config,
	name string,
) (func(ctx context.Context, conn net.Conn, closeWrite func() error), error) {
	handler := &tcpHandler{
		middleware: &Middleware{config: config},
		nextFunc:   next,
	}

	// Return a function closure that calls the handler
	return func(ctx context.Context, conn net.Conn, closeWrite func() error) {
		handler.ServeTCP(ctx, conn, closeWrite)
	}, nil
}

// tcpHandler wraps the unified middleware for TCP.
type tcpHandler struct {
	middleware *Middleware
	nextFunc   func(ctx context.Context, conn net.Conn, closeWrite func() error)
}

func (t *tcpHandler) ServeTCP(ctx context.Context, conn net.Conn, closeWrite func() error) {
	// Wrap TCP connection in unified Connection
	wrappedConn := &TCPConnection{
		conn:       conn,
		closeWrite: closeWrite,
		nextFunc:   t.nextFunc,
		ctx:        ctx,
	}

	// Use unified middleware logic
	t.middleware.Serve(wrappedConn)
}

// TCPConnection implements Connection for TCP connections.
type TCPConnection struct {
	conn       net.Conn
	closeWrite func() error
	nextFunc   func(ctx context.Context, conn net.Conn, closeWrite func() error)
	ctx        context.Context
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
	// For TCP: store in context metadata map
	// Try to get existing metadata map from context using string key
	// (Traefik uses both string and typed keys, so we try string first)
	metadata, ok := t.ctx.Value("metadata").(map[string]string)
	if !ok {
		// Try with typed key (contextKey("metadata"))
		type contextKey string
		metadata, ok = t.ctx.Value(contextKey("metadata")).(map[string]string)
		if !ok {
			// Create new metadata map if it doesn't exist
			// Use string key for consistency with postgres.go
			metadata = make(map[string]string)
			t.ctx = context.WithValue(t.ctx, "metadata", metadata)
		}
	}
	// Modify the map in place (maps are reference types)
	metadata[key] = value
	// Note: We don't need to update t.ctx here because we're modifying the map in place
	// But we DO need to make sure t.ctx has the map when we call nextFunc
}

func (t *TCPConnection) Proceed() {
	// Call next handler callback with stdlib types only
	if t.nextFunc != nil {
		t.nextFunc(t.ctx, t.conn, t.closeWrite)
	}
}

// GetContext returns the context for testing metadata.
func (t *TCPConnection) GetContext() context.Context {
	return t.ctx
}

// Test globals to track execution
var (
	TestLastServeTCPCalled   bool
	TestLastServeTCPIP       string
	TestLastServeTCPRejected bool
)

// TestNoOpTCPHandler creates a no-op TCP handler for testing.
func TestNoOpTCPHandler() TCPHandler {
	return &noOpTCPHandler{}
}

type noOpTCPHandler struct{}

func (n *noOpTCPHandler) ServeTCP(ctx context.Context, conn net.Conn, closeWrite func() error) {
	TestLastServeTCPCalled = true
	TestLastServeTCPIP = conn.RemoteAddr().String()
	// No-op - just records that it was called
}
