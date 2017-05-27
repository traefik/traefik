package server

import (
	"fmt"
	"net/http"

	"github.com/containous/traefik/log"
)

// OxyLogger implements oxy Logger interface with logrus.
type OxyLogger struct {
}

// Infof logs specified string as Debug level in logrus.
func (oxylogger *OxyLogger) Infof(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

// Warningf logs specified string as Warning level in logrus.
func (oxylogger *OxyLogger) Warningf(format string, args ...interface{}) {
	log.Warningf(format, args...)
}

// Errorf logs specified string as Warningf level in logrus.
func (oxylogger *OxyLogger) Errorf(format string, args ...interface{}) {
	log.Warningf(format, args...)
}

func notFoundHandler(contentType string, message string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType+"; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(404)
		fmt.Fprintln(w, message)
	}
}
