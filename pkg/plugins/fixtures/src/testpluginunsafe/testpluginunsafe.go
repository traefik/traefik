package testpluginunsafe

import (
	"context"
	"net/http"
	"unsafe"
)

type Config struct {
	Message string
}

func CreateConfig() *Config {
	return &Config{Message: "unsafe only plugin"}
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	// Use ONLY unsafe to test it's available
	size := unsafe.Sizeof(int(0))

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("X-Test-Plugin", "unsafe-only")
		rw.Header().Set("X-Test-Unsafe-Size", string(rune(size)))
		next.ServeHTTP(rw, req)
	}), nil
}
