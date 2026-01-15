package testpluginsafe

import (
	"context"
	"net/http"
)

type Config struct {
	Message string
}

func CreateConfig() *Config {
	return &Config{Message: "safe plugin"}
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("X-Test-Plugin", "safe")
		next.ServeHTTP(rw, req)
	}), nil
}
