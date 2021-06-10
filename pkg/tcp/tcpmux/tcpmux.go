package tcpmux

import (
	"github.com/traefik/traefik/v2/pkg/tcp"
)

type TCPMux struct {
	// Routes that will be matched (ordered).
	routes []*Route
}

// NewTCPMux returns a new TCP mux router.
func NewTCPMux() (TCPMux, error) {
	return TCPMux{}, nil
}

func (t *TCPMux) match(conn tcp.WriteCloser) tcp.Handler {
	// For each route, check if match, and return the handler for that route.
	for _, route := range t.routes {
		if route.Match(conn) {
			return route.handler
		}
	}
	return nil
}

// ServeTCP serves TCP traffic on the matching handler.
func (t *TCPMux) ServeTCP(conn tcp.WriteCloser) {
	handler := t.match(conn)

	if handler != nil {
		handler.ServeTCP(conn)
	}
}
