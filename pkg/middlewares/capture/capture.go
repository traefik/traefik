// Package capture is a middleware that captures requests/responses size, and status.
//
// For another middleware to get those attributes of a request/response, this middleware
// should be added before in the middleware chain.
//
//	handler, _ := NewHandler()
//	chain := alice.New().
//	     Append(WrapHandler(handler)).
//	     Append(myOtherMiddleware).
//	     then(...)
//
// As this middleware stores those data in the request's context, the data can
// be retrieved at anytime after the ServerHTTP.
//
//	func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.Handler) {
//		capt, err := capture.FromContext(req.Context())
//		if err != nil {
//		...
//		}
//
//		fmt.Println(capt.Status())
//		fmt.Println(capt.ResponseSize())
//		fmt.Println(capt.RequestSize())
//	}
package capture

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/containous/alice"
	"github.com/traefik/traefik/v2/pkg/middlewares"
)

type key string

const capturedData key = "capturedData"

// Handler will store each request data to its context.
type Handler struct{}

// WrapHandler wraps capture handler into an Alice Constructor.
func WrapHandler(handler *Handler) alice.Constructor {
	return func(next http.Handler) (http.Handler, error) {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			handler.ServeHTTP(rw, req, next)
		}), nil
	}
}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request, next http.Handler) {
	c := Capture{}
	if req.Body != nil {
		readCounter := &readCounter{source: req.Body}
		c.rr = readCounter
		req.Body = readCounter
	}
	responseWriter := newResponseWriter(rw)
	c.rw = responseWriter
	ctx := context.WithValue(req.Context(), capturedData, &c)
	next.ServeHTTP(responseWriter, req.WithContext(ctx))
}

// Capture is the object populated by the capture middleware,
// allowing to gather information about the request and response.
type Capture struct {
	rr *readCounter
	rw responseWriter
}

// FromContext returns the Capture value found in ctx, or an empty Capture otherwise.
func FromContext(ctx context.Context) (*Capture, error) {
	c := ctx.Value(capturedData)
	if c == nil {
		return nil, errors.New("value not found")
	}
	capt, ok := c.(*Capture)
	if !ok {
		return nil, errors.New("value stored in Context is not a *Capture")
	}
	return capt, nil
}

func (c Capture) ResponseSize() int64 {
	return c.rw.Size()
}

func (c Capture) StatusCode() int {
	return c.rw.Status()
}

// RequestSize returns the size of the request's body if it applies,
// zero otherwise.
func (c Capture) RequestSize() int64 {
	if c.rr == nil {
		return 0
	}
	return c.rr.size
}

type readCounter struct {
	// source ReadCloser from where the request body is read.
	source io.ReadCloser
	// size is total the number of bytes read.
	size int64
}

func (r *readCounter) Read(p []byte) (int, error) {
	n, err := r.source.Read(p)
	r.size += int64(n)
	return n, err
}

func (r *readCounter) Close() error {
	return r.source.Close()
}

var _ middlewares.Stateful = &responseWriterWithCloseNotify{}

type responseWriter interface {
	http.ResponseWriter
	Size() int64
	Status() int
}

func newResponseWriter(rw http.ResponseWriter) responseWriter {
	capt := &captureResponseWriter{rw: rw}
	if _, ok := rw.(http.CloseNotifier); !ok {
		return capt
	}

	return &responseWriterWithCloseNotify{capt}
}

// captureResponseWriter is a wrapper of type http.ResponseWriter
// that tracks response status and size.
type captureResponseWriter struct {
	rw     http.ResponseWriter
	status int
	size   int64
}

func (crw *captureResponseWriter) Header() http.Header {
	return crw.rw.Header()
}

func (crw *captureResponseWriter) Size() int64 {
	return crw.size
}

func (crw *captureResponseWriter) Status() int {
	return crw.status
}

func (crw *captureResponseWriter) Write(b []byte) (int, error) {
	if crw.status == 0 {
		crw.status = http.StatusOK
	}

	size, err := crw.rw.Write(b)
	crw.size += int64(size)

	return size, err
}

func (crw *captureResponseWriter) WriteHeader(s int) {
	crw.rw.WriteHeader(s)
	crw.status = s
}

func (crw *captureResponseWriter) Flush() {
	if f, ok := crw.rw.(http.Flusher); ok {
		f.Flush()
	}
}

func (crw *captureResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := crw.rw.(http.Hijacker); ok {
		return h.Hijack()
	}

	return nil, nil, fmt.Errorf("not a hijacker: %T", crw.rw)
}

type responseWriterWithCloseNotify struct {
	*captureResponseWriter
}

// CloseNotify returns a channel that receives at most a
// single value (true) when the client connection has gone away.
func (r *responseWriterWithCloseNotify) CloseNotify() <-chan bool {
	return r.rw.(http.CloseNotifier).CloseNotify()
}
