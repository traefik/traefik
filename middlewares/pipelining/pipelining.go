package pipelining

import (
	"bufio"
	"net"
	"net/http"

	"github.com/containous/traefik/middlewares"
)

// Compile time validation that the response recorder implements http interfaces correctly.
var _ middlewares.Stateful = &writerWithCloseNotify{}

// Pipelining returns a middleware
type Pipelining struct {
	next http.Handler
}

// NewPipelining returns a new Pipelining instance
func NewPipelining(next http.Handler) *Pipelining {
	return &Pipelining{
		next: next,
	}
}

func (p *Pipelining) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	writer := &writerWithoutCloseNotify{rw}
	if r.Method == http.MethodPut || r.Method == http.MethodPost {
		p.next.ServeHTTP(&writerWithCloseNotify{writer}, r)
	} else {
		p.next.ServeHTTP(writer, r)
	}

}

type writerWithCloseNotify struct {
	*writerWithoutCloseNotify
}

// CloseNotify returns a channel that receives at most a
// single value (true) when the client connection has gone away.
func (w *writerWithCloseNotify) CloseNotify() <-chan bool {
	return w.W.(http.CloseNotifier).CloseNotify()
}

// writerWithoutCloseNotify helps to disable closeNotify
type writerWithoutCloseNotify struct {
	W http.ResponseWriter
}

// Header returns the response headers.
func (w *writerWithoutCloseNotify) Header() http.Header {
	return w.W.Header()
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *writerWithoutCloseNotify) Write(buf []byte) (int, error) {
	return w.W.Write(buf)
}

// WriteHeader sends an HTTP response header with the provided
// status code.
func (w *writerWithoutCloseNotify) WriteHeader(code int) {
	w.W.WriteHeader(code)
}

// Flush sends any buffered data to the client.
func (w *writerWithoutCloseNotify) Flush() {
	if f, ok := w.W.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack hijacks the connection.
func (w *writerWithoutCloseNotify) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.W.(http.Hijacker).Hijack()
}
