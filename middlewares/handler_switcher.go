package middlewares

import (
	"net/http"

	"github.com/containous/traefik/safe"
)

// HandlerSwitcher allows hot switching of http.ServeMux
type HandlerSwitcher struct {
	handler *safe.Safe
}

// NewHandlerSwitcher builds a new instance of HandlerSwitcher
func NewHandlerSwitcher(newHandler http.Handler) (hs *HandlerSwitcher) {
	return &HandlerSwitcher{
		handler: safe.New(newHandler),
	}
}

func (h *HandlerSwitcher) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	handlerBackup := h.handler.Get().(http.Handler)
	handlerBackup.ServeHTTP(rw, req)
}

// GetHandler returns the current http.ServeMux
func (h *HandlerSwitcher) GetHandler() (newHandler http.Handler) {
	handler := h.handler.Get().(http.Handler)
	return handler
}

// UpdateHandler safely updates the current http.ServeMux with a new one
func (h *HandlerSwitcher) UpdateHandler(newHandler http.Handler) {
	h.handler.Set(newHandler)
}
