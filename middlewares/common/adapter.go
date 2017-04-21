package common

import "net/http"

// Adapter wraps any Negroni-like handler as Traefik middleware.
type Adapter struct {
	BasicMiddleware
	wrapped negroni
}

type negroni interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}

// NewAdapter wraps any Negroni-like handler as part of a Traefik middleware chain.
func NewAdapter(handler negroni, next http.Handler) Middleware {
	return &Adapter{NewMiddleware(next), handler}
}

func (a *Adapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nextFunc := a.Next().ServeHTTP
	a.wrapped.ServeHTTP(w, r, nextFunc)
}
