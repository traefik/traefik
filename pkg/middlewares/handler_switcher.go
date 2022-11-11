package middlewares

import (
	"net/http"

	"github.com/traefik/traefik/v2/pkg/safe"
)

// HTTPHandlerSwitcher allows hot switching of http.ServeMux.
type HTTPHandlerSwitcher struct {
	handler *safe.Atomic[http.Handler]
}

// NewHandlerSwitcher builds a new instance of HTTPHandlerSwitcher.
func NewHandlerSwitcher(newHandler http.Handler) (hs *HTTPHandlerSwitcher) {
	return &HTTPHandlerSwitcher{
		handler: safe.NewAtomic[http.Handler](newHandler),
	}
}

func (h *HTTPHandlerSwitcher) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	h.handler.Get().ServeHTTP(rw, req)
}

// GetHandler returns the current http.ServeMux.
func (h *HTTPHandlerSwitcher) GetHandler() (newHandler http.Handler) {
	return h.handler.Get()
}

// UpdateHandler safely updates the current http.ServeMux with a new one.
func (h *HTTPHandlerSwitcher) UpdateHandler(newHandler http.Handler) {
	h.handler.Set(newHandler)
}
