package main

import (
	"github.com/containous/traefik/plugin"
	"net/http"
)

var _ plugin.Middleware = (*NoLoad)(nil)

// NoLoad is a plugin that does not have any Load function
type NoLoad struct {
}

// ServeHTTP implements the Middleware interface
func (s *NoLoad) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
}
