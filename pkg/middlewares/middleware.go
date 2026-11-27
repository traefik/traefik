package middlewares

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
)

type loggersKey struct{}

// loggers is what WithLogger stored: the context logger before any middleware fields were added, and the middleware logger.
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

// WithLogger creates a logger with the middleware fields, and returns a context carrying it.
// Unlike storing the GetLogger result with WithContext,
// the next middleware calling GetLogger or WithLogger on this context replaces the middleware fields instead of repeating them.
func WithLogger(ctx context.Context, middleware, middlewareType string) (context.Context, *zerolog.Logger) {
	base := baseLogger(ctx)
	logger := GetLogger(ctx, middleware, middlewareType)

	ctx = logger.WithContext(ctx)

	return context.WithValue(ctx, loggersKey{}, loggers{base: base, middleware: log.Ctx(ctx)}), logger
}

// baseLogger returns the context logger without the fields added by WithLogger.
// zerolog cannot remove or override a field, so building on the context logger would repeat the middleware fields.
// If the context logger was replaced since WithLogger, it is used as is so that the fields added since are not lost.
func baseLogger(ctx context.Context) *zerolog.Logger {
	l, ok := ctx.Value(loggersKey{}).(loggers)
	if ok && l.middleware == log.Ctx(ctx) {
		return l.base
	}

	return log.Ctx(ctx)
}
