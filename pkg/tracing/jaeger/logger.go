package jaeger

import (
	"github.com/sirupsen/logrus"
	"github.com/traefik/traefik/v2/pkg/log"
)

// jaegerLogger is an implementation of the Logger interface that delegates to traefik log.
type jaegerLogger struct {
	logger logrus.FieldLogger
}

func newJaegerLogger() *jaegerLogger {
	return &jaegerLogger{
		logger: log.WithoutContext().WithField(log.TracingProviderName, "jaeger"),
	}
}

func (l *jaegerLogger) Error(msg string) {
	l.logger.Errorf("Tracing jaeger error: %s", msg)
}

// Infof logs a message at debug priority.
func (l *jaegerLogger) Infof(msg string, args ...interface{}) {
	l.logger.Debugf(msg, args...)
}
