package loadbalancer

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

// StatusRecordingResponseWriter wraps http.ResponseWriter to capture the HTTP status code
// written by the upstream handler without buffering the response body.
type StatusRecordingResponseWriter struct {
	http.ResponseWriter

	statusCode int
}

// NewStatusRecordingResponseWriter creates a new StatusRecordingResponseWriter.
func NewStatusRecordingResponseWriter(rw http.ResponseWriter) *StatusRecordingResponseWriter {
	return &StatusRecordingResponseWriter{ResponseWriter: rw, statusCode: http.StatusOK}
}

// WriteHeader records the status code and delegates to the underlying ResponseWriter.
func (s *StatusRecordingResponseWriter) WriteHeader(code int) {
	s.statusCode = code
	s.ResponseWriter.WriteHeader(code)
}

// Write records StatusOK when body is written without an explicit status code.
func (s *StatusRecordingResponseWriter) Write(b []byte) (int, error) {
	return s.ResponseWriter.Write(b)
}

// StatusCode returns the recorded HTTP status code.
func (s *StatusRecordingResponseWriter) StatusCode() int {
	return s.statusCode
}

// IsServerError returns true if the recorded status code indicates a 5xx server error.
func (s *StatusRecordingResponseWriter) IsServerError() bool {
	return s.statusCode >= http.StatusInternalServerError
}

// Unwrap returns the wrapped response writer for ResponseController support.
func (s *StatusRecordingResponseWriter) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}

// Flush implements http.Flusher when supported by the wrapped writer.
func (s *StatusRecordingResponseWriter) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements http.Hijacker when supported by the wrapped writer.
func (s *StatusRecordingResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := s.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}

	return nil, nil, fmt.Errorf("not a hijacker: %T", s.ResponseWriter)
}
