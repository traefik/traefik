//go:build linux || openbsd || darwin

package server

import "golang.org/x/sys/unix"

const unixSOREUSEPORT = unix.SO_REUSEPORT
