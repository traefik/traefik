package server

import (
	"context"
	"io"
	"net"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares"
)

// StatusClientClosedRequest non-standard HTTP status code for client disconnection
const StatusClientClosedRequest = 499

// StatusClientClosedRequestText non-standard HTTP status for client disconnection
const StatusClientClosedRequestText = "Client Closed Request"

// RecordingErrorHandler is an error handler, implementing the vulcand/oxy
// error handler interface, which is recording network errors by using the netErrorRecorder.
// In addition it sets a proper HTTP status code and body, depending on the type of error occurred.
type RecordingErrorHandler struct {
	netErrorRecorder middlewares.NetErrorRecorder
}

// NewRecordingErrorHandler creates and returns a new instance of RecordingErrorHandler.
func NewRecordingErrorHandler(recorder middlewares.NetErrorRecorder) *RecordingErrorHandler {
	return &RecordingErrorHandler{recorder}
}

func (eh *RecordingErrorHandler) ServeHTTP(w http.ResponseWriter, req *http.Request, err error) {
	statusCode := http.StatusInternalServerError

	if e, ok := err.(net.Error); ok {
		eh.netErrorRecorder.Record(req.Context())
		if e.Timeout() {
			statusCode = http.StatusGatewayTimeout
		} else {
			statusCode = http.StatusBadGateway
		}
	} else if err == io.EOF {
		eh.netErrorRecorder.Record(req.Context())
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
