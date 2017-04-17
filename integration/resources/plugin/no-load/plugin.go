package main

import (
	"github.com/containous/traefik/plugin"
	"net/http"
)

var _ plugin.Middleware = (*NoLoad)(nil)

type NoLoad struct {
}

func (s *NoLoad) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
}
