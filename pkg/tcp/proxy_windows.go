//go:build windows

package tcp

import "syscall"

const errConnReset = syscall.WSAECONNRESET
