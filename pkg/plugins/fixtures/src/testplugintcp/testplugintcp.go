package testplugintcp

import (
	"context"
	"net"
	"net/http"
)

// Config holds the plugin configuration.
type Config struct {
	HeaderValue string
	IPPrefix    string
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		HeaderValue: "tcp-plugin",
		IPPrefix:    "test-",
	}
}

// New creates a new HTTP middleware instance.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("X-Test-Plugin-TCP", config.HeaderValue)
		next.ServeHTTP(rw, req)
	}), nil
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
	return &tcpMiddleware{
		next:     next,
		config:   config,
		ipPrefix: config.IPPrefix,
	}, nil
}

type tcpMiddleware struct {
	next     TCPHandler
	config   *Config
	ipPrefix string
}

func (m *tcpMiddleware) ServeTCP(conn TCPWriteCloser) {
	// Simple test: if IP starts with "127", proceed, otherwise close
	addr := conn.RemoteAddr().String()
	host, _, _ := net.SplitHostPort(addr)

	// For testing: block IPs starting with "192"
	if len(host) >= 3 && host[:3] == "192" {
		conn.Close()
		return
	}

	m.next.ServeTCP(conn)
}
