package stats

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

type Recorder interface {
	http.ResponseWriter
	Status() int
}

type RecorderResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *RecorderResponseWriter) WriteHeader(code int) {
	r.ResponseWriter.WriteHeader(code)
	r.status = code
}

func (r *RecorderResponseWriter) Flush() {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

func (r *RecorderResponseWriter) Status() int {
	return r.status
}

// Proxy method to Status to add support for gocraft
func (r *RecorderResponseWriter) StatusCode() int {
	return r.Status()
}

func (r *RecorderResponseWriter) Size() int {
	return r.size
}

func (r *RecorderResponseWriter) Written() bool {
	return r.StatusCode() != 0
}

func (r *RecorderResponseWriter) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (r *RecorderResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}
