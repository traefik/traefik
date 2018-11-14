package middlewares

import (
	"context"

	"github.com/containous/traefik/log"
	"github.com/sirupsen/logrus"
)

// GetLogger creates a logger configured with the middleware fields.
func GetLogger(ctx context.Context, middleware string, middlewareType string) logrus.FieldLogger {
	return log.FromContext(ctx).WithField(log.MiddlewareName, middleware).WithField(log.MiddlewareType, middlewareType)
}
