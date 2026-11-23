package testplugintcpinvalid

import (
	"context"
	"net"
	"net/http"
)

// Config holds the plugin configuration.
type Config struct{}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// New creates a new HTTP middleware instance.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return next, nil
}

// TCPHandler mirrors Traefik's tcp.Handler, which Traefik cannot pass to interpreted code.
type TCPHandler interface {
	ServeTCP(ctx context.Context, conn net.Conn)
}

// NewTCP uses an interface instead of the function types Traefik expects, so Traefik must reject it.
func NewTCP(ctx context.Context, next TCPHandler, config *Config, name string) (TCPHandler, error) {
	return next, nil
}
