package server

import (
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

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
	//templatesRenderer.HTML(w, http.StatusNotFound, "notFound", nil)
}
