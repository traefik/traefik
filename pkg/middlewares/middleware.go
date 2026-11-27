package middlewares

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
)

type loggersKey struct{}

type loggers struct {
	base       *zerolog.Logger
	middleware *zerolog.Logger
}

// GetLogger creates a logger with the middleware fields.
func GetLogger(ctx context.Context, middleware, middlewareType string) *zerolog.Logger {
	logger := baseLogger(ctx).With().
		Str(logs.MiddlewareName, middleware).
		Str(logs.MiddlewareType, middlewareType).
		Logger()

	return &logger
}

// WithLogger is like GetLogger, but returns a context on which later middlewares replace the fields instead of repeating them.
func WithLogger(ctx context.Context, middleware, middlewareType string) (context.Context, *zerolog.Logger) {
	base := baseLogger(ctx)
	logger := GetLogger(ctx, middleware, middlewareType)

	ctx = logger.WithContext(ctx)

	return context.WithValue(ctx, loggersKey{}, loggers{base: base, middleware: log.Ctx(ctx)}), logger
}

// zerolog cannot remove a field, so rebuild from the logger before WithLogger, unless it was replaced since.
func baseLogger(ctx context.Context) *zerolog.Logger {
	l, ok := ctx.Value(loggersKey{}).(loggers)
	if ok && l.middleware == log.Ctx(ctx) {
		return l.base
	}

	return log.Ctx(ctx)
}
