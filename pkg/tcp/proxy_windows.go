//go:build windows

package tcp

import (
	"errors"
	"net"
	"syscall"
)

// isReadConnResetError reports whether err is a connection reset error during a read operation.
// The read error may be wrapped in another net.OpError (e.g. "readfrom" or "writeto"),
// as io.Copy relies on (*net.TCPConn).ReadFrom or (*net.TCPConn).WriteTo when available.
func isReadConnResetError(err error) bool {
	for {
		oerr, ok := errors.AsType[*net.OpError](err)
		if !ok {
			return false
		}

		if oerr.Op == "read" {
			return errors.Is(oerr.Err, syscall.WSAECONNRESET)
		}

		err = oerr.Err
	}
}
