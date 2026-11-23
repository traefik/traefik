package plugins_test

// This file shows how a Yaegi middleware plugin supports TCP routers next to HTTP routers.
// The working fixture is fixtures/src/testplugincombined.

/*
Interpreted code cannot implement Traefik's tcp.Handler, so NewTCP only uses standard library types:
the next handler and the returned handler are plain functions.
A NewTCP with another signature is rejected when the middleware is built.

// plugin.go
package myplugin

import (
	"context"
	"net"
	"net/http"
	"strings"
)

// Config holds the plugin configuration.
type Config struct {
	AllowedPrefix string `json:"allowedPrefix,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{AllowedPrefix: "10."}
}

// New creates the HTTP middleware.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return next, nil
}

// NewTCP creates the TCP middleware. It is optional: without it, the plugin only supports HTTP routers.
func NewTCP(
	ctx context.Context,
	next func(ctx context.Context, conn net.Conn, closeWrite func() error),
	config *Config,
	name string,
) (func(ctx context.Context, conn net.Conn, closeWrite func() error), error) {
	return func(ctx context.Context, conn net.Conn, closeWrite func() error) {
		host, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

		// The access log adds every entry of this map to the connection line.
		if metadata, ok := ctx.Value("metadata").(map[string]string); ok {
			metadata["X-Client-IP"] = host
		}

		if !strings.HasPrefix(host, config.AllowedPrefix) {
			_ = conn.Close()
			return
		}

		next(ctx, conn, closeWrite)
	}, nil
}

// Dynamic configuration of a TCP router using the plugin:
tcp:
  routers:
    my-tcp-router:
      rule: "HostSNI(`*`)"
      service: my-tcp-service
      middlewares:
        - my-tcp-plugin
  middlewares:
    my-tcp-plugin:
      plugin:
        myplugin:
          allowedPrefix: "10."
*/
