package utils

import (
	"io"
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// ErrorHandler error handler
type ErrorHandler interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request, err error)
}

// DefaultHandler default error handler
var DefaultHandler ErrorHandler = &StdHandler{}

// StdHandler Standard error handler
type StdHandler struct{}

func (e *StdHandler) ServeHTTP(w http.ResponseWriter, req *http.Request, err error) {
	statusCode := http.StatusInternalServerError
	if e, ok := err.(net.Error); ok {
		if e.Timeout() {
			statusCode = http.StatusGatewayTimeout
		} else {
			statusCode = http.StatusBadGateway
		}
	} else if err == io.EOF {
		statusCode = http.StatusBadGateway
	}
	w.WriteHeader(statusCode)
	w.Write([]byte(http.StatusText(statusCode)))
	log.Debugf("'%d %s' caused by: %v", statusCode, http.StatusText(statusCode), err)
}

// ErrorHandlerFunc error handler function type
type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error)

// ServeHTTP calls f(w, r).
func (f ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, err error) {
	f(w, r, err)
}
