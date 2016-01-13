/*
Copyright
*/
package main

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
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

// Errorf logs specified string as Error level in logrus.
func (oxylogger *OxyLogger) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
	//templatesRenderer.HTML(w, http.StatusNotFound, "notFound", nil)
}

// LoadDefaultConfig returns a default gorrilla.mux router from the specified configuration.
func LoadDefaultConfig(globalConfiguration GlobalConfiguration) *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	return router
}
