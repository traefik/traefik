package main

import (
	"github.com/containous/traefik/plugin"
	"net/http"
)

var _ plugin.Middleware = (*BadLoad)(nil)

type BadLoad struct {
}

func Load() string {
	return "bad load!"
}

func (s *BadLoad) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
}
