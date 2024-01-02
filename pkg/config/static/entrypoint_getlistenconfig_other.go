//go:build !(linux || freebsd || openbsd || darwin)

package static

import "net"

// GetListenConfig returns an empty net.ListenConfig. It ignores the reuse port
// field of the entry point.
func (ep EntryPoint) GetListenConfig() net.ListenConfig {
	return net.ListenConfig{}
}
