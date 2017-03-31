package plugin

import (
	"net/http"
)

// Middleware defines functions that should be implemented by a Middleware plugin
type Middleware interface {
	Handler
}

// Handler handler is an interface that objects can implement to be registered to serve as middleware
type Handler interface {
	ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}
