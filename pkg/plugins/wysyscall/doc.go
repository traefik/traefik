// Package wysyscall registers the syscall symbols with the wy interpreter.
// wy does not ship them because the syscall API differs between platforms:
// each file holds the API of one GOOS, intersected across the architectures
// Traefik is released for, and the last one the API common to the other
// platforms Go supports syscall on.
package wysyscall

//go:generate go run github.com/darkweak/wy/cmd/wy extract -o . -pkg wysyscall -suffix _linux -goos linux/amd64,linux/386,linux/arm,linux/arm64,linux/ppc64le,linux/s390x,linux/riscv64 syscall
//go:generate go run github.com/darkweak/wy/cmd/wy extract -o . -pkg wysyscall -suffix _darwin -goos darwin/amd64,darwin/arm64 syscall
//go:generate go run github.com/darkweak/wy/cmd/wy extract -o . -pkg wysyscall -suffix _windows -goos windows/amd64,windows/386,windows/arm64 syscall
//go:generate go run github.com/darkweak/wy/cmd/wy extract -o . -pkg wysyscall -suffix _freebsd -goos freebsd/amd64,freebsd/386,freebsd/riscv64 syscall
//go:generate go run github.com/darkweak/wy/cmd/wy extract -o . -pkg wysyscall -suffix _openbsd -goos openbsd/amd64,openbsd/386,openbsd/riscv64 syscall
//go:generate go run github.com/darkweak/wy/cmd/wy extract -o . -pkg wysyscall -build "!linux && !darwin && !windows && !freebsd && !openbsd" -goos netbsd/amd64,netbsd/arm64,dragonfly/amd64,solaris/amd64,illumos/amd64,aix/ppc64 syscall
