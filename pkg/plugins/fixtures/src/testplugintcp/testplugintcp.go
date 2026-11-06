package testplugintcp

import (
	"context"
	"net"
	"net/http"
)

type Config struct {
	Banner string
	Wrap   bool
}

func CreateConfig() *Config {
	return &Config{}
}

func New(_ context.Context, next http.Handler, _ *Config, _ string) (http.Handler, error) {
	return next, nil
}

func NewTCP(_ context.Context, next func(context.Context, net.Conn, func() error), config *Config, _ string) (func(context.Context, net.Conn, func() error), error) {
	return func(ctx context.Context, conn net.Conn, closeWrite func() error) {
		if _, err := conn.Write([]byte(config.Banner)); err != nil {
			_ = conn.Close()
			return
		}

		if config.Wrap {
			next(ctx, &wrappedConn{Conn: conn}, closeWrite)
			return
		}

		next(ctx, conn, closeWrite)
	}, nil
}

type wrappedConn struct {
	net.Conn
}
