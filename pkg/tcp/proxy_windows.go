//go:build windows

package tcp

import (
	"errors"
	"net"
	"syscall"
)

// isReadConnResetError reports whether err is a connection reset error during a read operation.
func isReadConnResetError(err error) bool {
	oerr, ok := errors.AsType[*net.OpError](err)
	return ok && oerr.Op == "read" && errors.Is(err, syscall.WSAECONNRESET)
}
