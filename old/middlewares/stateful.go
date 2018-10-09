package middlewares

import "net/http"

// Stateful interface groups all http interfaces that must be
// implemented by a stateful middleware (ie: recorders)
type Stateful interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher
	http.CloseNotifier
}
