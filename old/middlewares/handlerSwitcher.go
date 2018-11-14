package middlewares

import (
	"net/http"

	"github.com/containous/mux"
	"github.com/containous/traefik/safe"
)

// HandlerSwitcher allows hot switching of http.ServeMux
type HandlerSwitcher struct {
	handler *safe.Safe
}

// NewHandlerSwitcher builds a new instance of HandlerSwitcher
func NewHandlerSwitcher(newHandler *mux.Router) (hs *HandlerSwitcher) {
	return &HandlerSwitcher{
		handler: safe.New(newHandler),
	}
}

func (hs *HandlerSwitcher) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	handlerBackup := hs.handler.Get().(*mux.Router)
	handlerBackup.ServeHTTP(rw, r)
}

// GetHandler returns the current http.ServeMux
func (hs *HandlerSwitcher) GetHandler() (newHandler *mux.Router) {
	handler := hs.handler.Get().(*mux.Router)
	return handler
}

// UpdateHandler safely updates the current http.ServeMux with a new one
func (hs *HandlerSwitcher) UpdateHandler(newHandler *mux.Router) {
	hs.handler.Set(newHandler)
}
