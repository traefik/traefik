package middlewares

import (
	"context"

	"github.com/baqupio/baqup/v3/pkg/observability/logs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// GetLogger creates a logger with the middleware fields.
func GetLogger(ctx context.Context, middleware, middlewareType string) *zerolog.Logger {
	logger := log.Ctx(ctx).With().
		Str(logs.MiddlewareName, middleware).
		Str(logs.MiddlewareType, middlewareType).
		Logger()

	return &logger
}
