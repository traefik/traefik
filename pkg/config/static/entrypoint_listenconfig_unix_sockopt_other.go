//go:build !freebsd

package static

import "golang.org/x/sys/unix"

const unixSOREUSEPORT = unix.SO_REUSEPORT
