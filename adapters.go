/*
Copyright
*/
package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/mailgun/oxy/utils"
	"net/http"
)

type OxyLogger struct {
}

func (oxylogger *OxyLogger) Infof(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func (oxylogger *OxyLogger) Warningf(format string, args ...interface{}) {
	log.Warningf(format, args...)
}

func (oxylogger *OxyLogger) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

type ErrorHandler struct {
}

func (e *ErrorHandler) ServeHTTP(w http.ResponseWriter, req *http.Request, err error) {
	log.Error("server error ", err.Error())
	utils.DefaultHandler.ServeHTTP(w, req, err)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
	//templatesRenderer.HTML(w, http.StatusNotFound, "notFound", nil)
}

func LoadDefaultConfig(gloablConfiguration *GlobalConfiguration) *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	return router
}
