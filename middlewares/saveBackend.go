package middlewares

/*
Middleware saveBackend sends the backend name to the logger.
*/

import (
	"net/http"
)

// SaveBackend holds the next handler
type SaveBackend struct {
	next http.Handler
}

// NewSaveBackend creates a SaveBackend
func NewSaveBackend(next http.Handler) *SaveBackend {
	return &SaveBackend{next}
}

func (sb *SaveBackend) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	saveBackendNameForLogger(r, (*r.URL).String())
	sb.next.ServeHTTP(rw, r)
}
