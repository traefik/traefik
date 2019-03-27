package stats

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	// Status returns the status code of the response or 0 if the response has not been written.
	Status() int
	// Written returns whether or not the ResponseWriter has been written.
	Written() bool
	// Size returns the size of the response body.
	Size() int
	// Before allows for a function to be called before the ResponseWriter has been written to. This is
	// useful for setting headers or any other operations that must happen before a response has been written.
	Before(func(ResponseWriter))
}

type beforeFunc func(ResponseWriter)

type recorderResponseWriter struct {
	http.ResponseWriter
	status      int
	size        int
	beforeFuncs []beforeFunc
	written bool
}

func NewRecorderResponseWriter(w http.ResponseWriter, statusCode int) ResponseWriter {
	return &recorderResponseWriter{ResponseWriter: w, status: statusCode}
}

func (r *recorderResponseWriter) WriteHeader(code int) {
	r.written = true
	r.ResponseWriter.WriteHeader(code)
	r.status = code
}

func (r *recorderResponseWriter) Flush() {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if ok {
		flusher.Flush()
	}
}

func (r *recorderResponseWriter) Status() int {
	return r.status
}

func (r *recorderResponseWriter) Write(b []byte) (int, error) {
	if !r.Written() {
		// The status will be StatusOK if WriteHeader has not been called yet
		r.WriteHeader(http.StatusOK)
	}
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

// Proxy method to Status to add support for gocraft
func (r *recorderResponseWriter) StatusCode() int {
	return r.Status()
}

func (r *recorderResponseWriter) Size() int {
	return r.size
}

func (r *recorderResponseWriter) Written() bool {
	return r.StatusCode() != 0
}

func (r *recorderResponseWriter) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (r *recorderResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if !r.written {
		r.status = 0
	}
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (r *recorderResponseWriter) Before(before func(ResponseWriter)) {
	r.beforeFuncs = append(r.beforeFuncs, before)
}
