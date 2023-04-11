//go:build linux || freebsd || openbsd || darwin

package static

import (
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

// GetListenConfig returns a net.ListenConfig for the address field of the entry
// point based on its reuse port field.
func (ep EntryPoint) GetListenConfig() net.ListenConfig {
	if !ep.ReusePort {
		return net.ListenConfig{}
	}
	return net.ListenConfig{
		Control: func(_, _ string, c syscall.RawConn) error {
			var err error
			if controlErr := c.Control(func(fd uintptr) {
				err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
				if err != nil {
					return
				}
				err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix_SO_REUSEPORT, 1)
				if err != nil {
					return
				}
			}); controlErr != nil {
				return controlErr
			}
			return err
		},
	}
}
