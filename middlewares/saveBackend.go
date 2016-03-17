package middlewares

import (
	"net/http"
	"strings"
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
	backendName := (*r.URL).String()
	saveNameForLogger(r, loggerBackend, strings.TrimPrefix(backendName, "backend-"))
	saveBackend.next.ServeHTTP(rw, r)
}
