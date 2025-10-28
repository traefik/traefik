package testpluginsyscall

import (
	"context"
	"net/http"
	"syscall"
	"unsafe"
)

type Config struct {
	Message string
}

func CreateConfig() *Config {
	return &Config{Message: "syscall plugin"}
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	// Use syscall and unsafe to test they're available
	pid := syscall.Getpid()
	size := unsafe.Sizeof(int(0))

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("X-Test-Plugin", "syscall")
		rw.Header().Set("X-Test-PID", string(rune(pid)))
		rw.Header().Set("X-Test-Size", string(rune(size)))
		next.ServeHTTP(rw, req)
	}), nil
}
