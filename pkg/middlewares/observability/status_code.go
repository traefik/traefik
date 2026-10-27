package observability

import (
	"net/http"
)

// newStatusCodeRecorder returns an initialized statusCodeRecoder.
func newStatusCodeRecorder(rw http.ResponseWriter, status int) *statusCodeRecorder {
	return &statusCodeRecorder{rw, status}
}

type statusCodeRecorder struct {
	http.ResponseWriter

	status int
}

// WriteHeader captures the status code for later retrieval.
func (s *statusCodeRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

// Status get response status.
func (s *statusCodeRecorder) Status() int {
	return s.status
}

// Unwrap exposes the underlying ResponseWriter for http.ResponseController.
func (s *statusCodeRecorder) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}
