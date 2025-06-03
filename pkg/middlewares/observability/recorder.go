package observability

import (
	"bufio"
	"net"
	"net/http"
)

// newRecorder returns an initialized statusCodeRecoder.
func newRecorder(rw http.ResponseWriter, status int, spanName string) *recorder {
	return &recorder{rw, status, spanName}
}

type recorder struct {
	http.ResponseWriter
	status   int
	spanName string
}

// WriteHeader captures the status code for later retrieval.
func (s *recorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

// Status get response status.
func (s *recorder) Status() int {
	return s.status
}

// SetSpanName updates the span name for later retrieval
func (s *recorder) SetSpanName(spanName string) {
	s.spanName = spanName
}

// SpanName retrieves the span name.
func (s *recorder) SpanName() string {
	return s.spanName
}

// Hijack hijacks the connection.
func (s *recorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return s.ResponseWriter.(http.Hijacker).Hijack()
}

// Flush sends any buffered data to the client.
func (s *recorder) Flush() {
	if flusher, ok := s.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
