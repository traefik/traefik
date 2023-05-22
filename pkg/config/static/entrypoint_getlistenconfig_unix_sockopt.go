//go:build linux || openbsd || darwin

package static

import "golang.org/x/sys/unix"

const unix_SO_REUSEPORT = unix.SO_REUSEPORT
