//go:build !(linux || freebsd || openbsd || darwin)

package static

import "net"

// GetListenConfig returns an empty net.ListenConfig.
// It ignores the reusePort configuration of the entry point.
func (ep EntryPoint) GetListenConfig() net.ListenConfig {
	return net.ListenConfig{}
}
