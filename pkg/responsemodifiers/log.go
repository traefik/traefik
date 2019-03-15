package responsemodifiers

import (
	"context"

	"github.com/containous/traefik/pkg/log"
	"github.com/sirupsen/logrus"
)

// getLogger creates a logger configured with the middleware fields.
func getLogger(ctx context.Context, middleware string, middlewareType string) logrus.FieldLogger {
	return log.FromContext(ctx).WithField(log.MiddlewareName, middleware).WithField(log.MiddlewareType, middlewareType)
}
