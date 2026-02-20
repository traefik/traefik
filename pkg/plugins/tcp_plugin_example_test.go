package plugins_test

// This file demonstrates how to create a TCP-compatible plugin
// that works with both HTTP and TCP protocols.

/*
Example plugin structure:

pkg/
  myplugin/
    plugin.go
    .traefik.yml

// plugin.go
package myplugin

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"github.com/traefik/traefik/v3/pkg/tcp"
)

// Config holds the plugin configuration.
type Config struct {
	MaxConnections int    `json:"maxConnections,omitempty"`
	Message        string `json:"message,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		MaxConnections: 10,
		Message:        "Plugin middleware",
	}
}

// New creates a new HTTP middleware instance.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &httpMiddleware{
		next:   next,
		name:   name,
		config: config,
	}, nil
}

// NewTCP creates a new TCP middleware instance.
// This is optional - only include it if your plugin supports TCP.
func NewTCP(ctx context.Context, next tcp.Handler, config *Config, name string) (tcp.Handler, error) {
	return &tcpMiddleware{
		next:   next,
		name:   name,
		config: config,
	}, nil
}

type httpMiddleware struct {
	next   http.Handler
	name   string
	config *Config
}

func (m *httpMiddleware) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// HTTP-specific logic
	// Can access headers, body, cookies, etc.
	fmt.Printf("[%s] HTTP request from %s\n", m.name, req.RemoteAddr)
	m.next.ServeHTTP(rw, req)
}

type tcpMiddleware struct {
	next   tcp.Handler
	name   string
	config *Config
}

func (m *tcpMiddleware) ServeTCP(conn tcp.WriteCloser) {
	// TCP-specific logic
	// Can only access connection-level data
	addr := conn.RemoteAddr().String()
	ip, _, _ := net.SplitHostPort(addr)
	fmt.Printf("[%s] TCP connection from %s\n", m.name, ip)
	m.next.ServeTCP(conn)
}

// .traefik.yml
displayName: My TCP-Compatible Plugin
type: middleware
import: github.com/example/my-plugin
summary: A plugin that works with both HTTP and TCP
testData:
  maxConnections: 5
  message: "Test message"
supportsTCP: true  # This indicates the plugin supports TCP

// Usage in Traefik configuration:

// For HTTP:
http:
  routers:
    my-router:
      rule: "Host(`example.com`)"
      service: my-service
      middlewares:
        - my-plugin
  middlewares:
    my-plugin:
      plugin:
        myplugin:
          maxConnections: 20
          message: "Hello from HTTP"

// For TCP:
tcp:
  routers:
    my-tcp-router:
      rule: "HostSNI(`example.com`)"
      service: my-tcp-service
      middlewares:
        - my-tcp-plugin
  middlewares:
    my-tcp-plugin:
      plugin:
        myplugin:
          maxConnections: 10
          message: "Hello from TCP"

*/
