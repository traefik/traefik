package metrics

import (
	"bufio"
	"net"
	"net/http"
)

type recorder interface {
	http.ResponseWriter
	http.Flusher
	getCode() int
}

func newResponseRecorder(rw http.ResponseWriter) recorder {
	rec := &responseRecorder{
		ResponseWriter: rw,
		statusCode:     http.StatusOK,
	}
	if _, ok := rw.(http.CloseNotifier); !ok {
		return rec
	}
	return &responseRecorderWithCloseNotify{rec}
}

// responseRecorder captures information from the response and preserves it for
// later analysis.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

type responseRecorderWithCloseNotify struct {
	*responseRecorder
}

// CloseNotify returns a channel that receives at most a
// single value (true) when the client connection has gone away.
func (r *responseRecorderWithCloseNotify) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (r *responseRecorder) getCode() int {
	return r.statusCode
}

// WriteHeader captures the status code for later retrieval.
func (r *responseRecorder) WriteHeader(status int) {
	r.ResponseWriter.WriteHeader(status)
	r.statusCode = status
}

// Hijack hijacks the connection.
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

// Flush sends any buffered data to the client.
func (r *responseRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
