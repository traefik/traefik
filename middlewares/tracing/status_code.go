package tracing

import (
	"bufio"
	"net"
	"net/http"
)

type statusCodeRecoder interface {
	http.ResponseWriter
	GetStatus() int
}

type statusCodeWithoutCloseNotify struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the status code for later retrieval.
func (s *statusCodeWithoutCloseNotify) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

// GetStatus get response status
func (s *statusCodeWithoutCloseNotify) GetStatus() int {
	return s.status
}

// Hijack hijacks the connection
func (s *statusCodeWithoutCloseNotify) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return s.ResponseWriter.(http.Hijacker).Hijack()
}

// Flush sends any buffered data to the client.
func (s *statusCodeWithoutCloseNotify) Flush() {
	s.ResponseWriter.(http.Flusher).Flush()
}

type statusCodeWithCloseNotify struct {
	*statusCodeWithoutCloseNotify
}

func (s *statusCodeWithCloseNotify) CloseNotify() <-chan bool {
	return s.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// newStatusCodeRecoder returns an initialized statusCodeRecoder.
func newStatusCodeRecoder(rw http.ResponseWriter, status int) statusCodeRecoder {
	recorder := &statusCodeWithoutCloseNotify{rw, status}
	if _, ok := rw.(http.CloseNotifier); ok {
		return &statusCodeWithCloseNotify{recorder}
	}
	return recorder
}
