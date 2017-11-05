package plugin

import (
	"net/http"
)

// Middleware defines functions that should be implemented by a Middleware plugin
type PluginMiddleware interface {
	PluginHandler
	Stop()
}

// Handler handler is an interface that objects can implement to be registered to serve as middleware
type PluginHandler interface {
	ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}
