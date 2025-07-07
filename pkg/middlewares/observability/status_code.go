package observability

import (
	"bufio"
	"net"
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

// Hijack hijacks the connection.
func (s *statusCodeRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return s.ResponseWriter.(http.Hijacker).Hijack()
}

// Flush sends any buffered data to the client.
func (s *statusCodeRecorder) Flush() {
	if flusher, ok := s.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
