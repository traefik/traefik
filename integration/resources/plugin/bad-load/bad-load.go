package main

import (
	"github.com/containous/traefik/plugin"
	"net/http"
)

var _ plugin.Middleware = (*BadLoad)(nil)

// BadLoad is a plugin that should not load
type BadLoad struct {
}

// Load loads the plugin instance
func Load() string {
	return "bad load!"
}

// ServeHTTP implements the Middleware interface
func (s *BadLoad) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
}
