//go:build !(linux || freebsd || openbsd || darwin)

package server

import (
	"net"

	"github.com/traefik/traefik/v3/pkg/config/static"
)

// newListenConfig creates a new net.ListenConfig for the given configuration of
// the entry point.
func newListenConfig(configuration *static.EntryPoint) (lc net.ListenConfig) {
	return
}
