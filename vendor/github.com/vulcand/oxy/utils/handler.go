package utils

import (
	"context"
	"io"
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// StatusClientClosedRequest non-standard HTTP status code for client disconnection
const StatusClientClosedRequest = 499

// StatusClientClosedRequestText non-standard HTTP status for client disconnection
const StatusClientClosedRequestText = "Client Closed Request"

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
	} else if err == context.Canceled {
		statusCode = StatusClientClosedRequest
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(statusText(statusCode)))
	log.Debugf("'%d %s' caused by: %v", statusCode, statusText(statusCode), err)
}

func statusText(statusCode int) string {
	if statusCode == StatusClientClosedRequest {
		return StatusClientClosedRequestText
	}
	return http.StatusText(statusCode)
}

// ErrorHandlerFunc error handler function type
type ErrorHandlerFunc func(http.ResponseWriter, *http.Request, error)

// ServeHTTP calls f(w, r).
func (f ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, err error) {
	f(w, r, err)
}
