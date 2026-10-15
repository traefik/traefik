//go:build unix

package runtimeconfig

import "golang.org/x/sys/unix"

const nonBlockingFileFlag = unix.O_NONBLOCK
