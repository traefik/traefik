//go:build freebsd

package static

import "golang.org/x/sys/unix"

const UNIX_SO_REUSEPORT = unix.SO_REUSEPORT_LB
