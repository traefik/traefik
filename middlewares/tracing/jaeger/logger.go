package jaeger

import "github.com/containous/traefik/log"

// jaegerLogger is an implementation of the Logger interface that delegates to traefik log
type jaegerLogger struct{}

func (l *jaegerLogger) Error(msg string) {
	log.Errorf("Tracing jaeger error: %s", msg)
}

// Infof logs a message at debug priority
func (l *jaegerLogger) Infof(msg string, args ...interface{}) {
	log.Debugf(msg, args...)
}
