package testpluginundeclared

import (
	"context"
	"net/http"

	"testpluginundeclared/internal/mem"
)

type Config struct {
	Message string
}

func CreateConfig() *Config {
	return &Config{Message: "undeclared plugin"}
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	_ = mem.WordSize()

	return next, nil
}
