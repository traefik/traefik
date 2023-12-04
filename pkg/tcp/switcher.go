package tcp

import (
	"context"

	"github.com/traefik/traefik/v2/pkg/safe"
)

// HandlerSwitcher is a TCP handler switcher.
type HandlerSwitcher struct {
	router safe.Safe
}

// ServeTCP forwards the TCP connection to the current active handler.
func (s *HandlerSwitcher) ServeTCP(ctx context.Context, conn WriteCloser) {
	handler := s.router.Get()
	h, ok := handler.(Handler)
	if ok {
		h.ServeTCP(ctx, conn)
	} else {
		conn.Close()
	}
}

// Switch sets the new TCP handler to use for new connections.
func (s *HandlerSwitcher) Switch(handler Handler) {
	s.router.Set(handler)
}
