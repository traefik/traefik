package middlewares

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

func (saveBackend *SaveBackend) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	saveNameForLogger(r, loggerBackend, (*r.URL).String())
	saveBackend.next.ServeHTTP(rw, r)
}
