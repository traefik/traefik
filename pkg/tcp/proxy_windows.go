//go:build windows
// +build windows

package tcp

import (
	"errors"
	"net"
	"syscall"
)

// isReadConnResetError reports whether err is a connection reset error during a read operation.
func isReadConnResetError(err error) bool {
	var oerr *net.OpError
	return errors.As(err, &oerr) && oerr.Op == "read" && errors.Is(err, syscall.WSAECONNRESET)
}
