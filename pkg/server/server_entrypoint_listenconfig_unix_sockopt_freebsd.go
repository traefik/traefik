//go:build freebsd

package server

import "golang.org/x/sys/unix"

const unixSOREUSEPORT = unix.SO_REUSEPORT_LB
