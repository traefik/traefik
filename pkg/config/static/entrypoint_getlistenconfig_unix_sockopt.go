//go:build linux || openbsd || darwin

package static

import "golang.org/x/sys/unix"

const unixSOREUSEPORT = unix.SO_REUSEPORT
