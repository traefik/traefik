package testplugincombined

import (
	"context"
	"net"
	"net/http"
	"strings"
)

// Config holds the plugin configuration.
type Config struct {
	AllowedIPPrefix string
	HTTPHeaderValue string
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		AllowedIPPrefix: "127",
		HTTPHeaderValue: "combined-plugin",
	}
}

// connection abstracts an HTTP request and a TCP connection, so both protocols share the filtering logic.
type connection interface {
	remoteIP() string
	reject()
	setMetadata(key, value string)
	proceed()
}

type middleware struct {
	config *Config
}

func (m *middleware) serve(conn connection) {
	ip := conn.remoteIP()

	if !strings.HasPrefix(ip, m.config.AllowedIPPrefix) {
		conn.reject()
		return
	}

	conn.setMetadata("X-Combined-Plugin", m.config.HTTPHeaderValue)
	conn.setMetadata("X-Client-IP", ip)

	conn.proceed()
}

// New creates a new HTTP middleware instance.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	mw := &middleware{config: config}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		mw.serve(&httpConnection{rw: rw, req: req, next: next})
	}), nil
}

type httpConnection struct {
	rw   http.ResponseWriter
	req  *http.Request
	next http.Handler
}

func (h *httpConnection) remoteIP() string {
	host, _, _ := net.SplitHostPort(h.req.RemoteAddr)
	return host
}

func (h *httpConnection) reject() {
	http.Error(h.rw, "Forbidden", http.StatusForbidden)
}

func (h *httpConnection) setMetadata(key, value string) {
	h.rw.Header().Set(key, value)
}

func (h *httpConnection) proceed() {
	h.next.ServeHTTP(h.rw, h.req)
}

// NewTCP creates a new TCP middleware instance.
// The handlers are plain functions over standard library types: interpreted code cannot implement Traefik's tcp.Handler.
func NewTCP(
	ctx context.Context,
	next func(ctx context.Context, conn net.Conn, closeWrite func() error),
	config *Config,
	name string,
) (func(ctx context.Context, conn net.Conn, closeWrite func() error), error) {
	mw := &middleware{config: config}

	return func(ctx context.Context, conn net.Conn, closeWrite func() error) {
		mw.serve(&tcpConnection{ctx: ctx, conn: conn, closeWrite: closeWrite, next: next})
	}, nil
}

type tcpConnection struct {
	ctx        context.Context
	conn       net.Conn
	closeWrite func() error
	next       func(ctx context.Context, conn net.Conn, closeWrite func() error)
}

func (t *tcpConnection) remoteIP() string {
	host, _, _ := net.SplitHostPort(t.conn.RemoteAddr().String())
	return host
}

func (t *tcpConnection) reject() {
	_ = t.conn.Close()
}

func (t *tcpConnection) setMetadata(key, value string) {
	// The map is updated in place: the access log reads the same map once the connection ends.
	if metadata, ok := t.ctx.Value("metadata").(map[string]string); ok {
		metadata[key] = value
	}
}

func (t *tcpConnection) proceed() {
	t.next(t.ctx, t.conn, t.closeWrite)
}
