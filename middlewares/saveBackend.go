package middlewares

import (
	"net/http"
)

type SaveBackend struct {
	next http.Handler
}

func NewSaveBackend(next http.Handler) *SaveBackend {
	return &SaveBackend{next}
}

func (saveBackend *SaveBackend) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	saveNameForLogger(r, loggerBackend, (*r.URL).String())
	saveBackend.next.ServeHTTP(rw, r)
}
