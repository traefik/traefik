//go:build linux || freebsd || openbsd || darwin

package static

import (
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

// GetListenConfig returns a net.ListenConfig for the address configuration of
// the entry point based on its reusePort configuration.
func (ep EntryPoint) GetListenConfig() net.ListenConfig {
	if !ep.ReusePort {
		return net.ListenConfig{}
	}
	return net.ListenConfig{
		Control: func(_, _ string, c syscall.RawConn) error {
			var setSockOptErr error
			err := c.Control(func(fd uintptr) {
				setSockOptErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
				if setSockOptErr != nil {
					return
				}
				setSockOptErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unixSOREUSEPORT, 1)
				if setSockOptErr != nil {
					return
				}
			})
			if err != nil {
				return fmt.Errorf("control: %w", err)
			}
			if setSockOptErr != nil {
				return fmt.Errorf("setsockopt: %w", setSockOptErr)
			}
			return nil
		},
	}
}
