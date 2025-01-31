// Package capture is a middleware that captures requests/responses size, and status.
//
// For another middleware to get those attributes of a request/response, this middleware
// should be added before in the middleware chain.
//
//	chain := alice.New().
//	     Append(capture.Wrap).
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

	"github.com/traefik/traefik/v3/pkg/middlewares"
)

type key string

const capturedData key = "capturedData"

// Wrap returns a new handler that inserts a Capture into the given handler for each incoming request.
// It satisfies the alice.Constructor type.
func Wrap(next http.Handler) (http.Handler, error) {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		capt, err := FromContext(req.Context())
		if err != nil {
			c := &Capture{}
			newRW, newReq := c.renew(rw, req)
			next.ServeHTTP(newRW, newReq)
			return
		}

		if capt.NeedsReset(rw) {
			newRW, newReq := capt.renew(rw, req)
			next.ServeHTTP(newRW, newReq)
			return
		}

		next.ServeHTTP(rw, req)
	}), nil
}

// FromContext returns the Capture value found in ctx, or an empty Capture otherwise.
func FromContext(ctx context.Context) (Capture, error) {
	c := ctx.Value(capturedData)
	if c == nil {
		return Capture{}, errors.New("value not found in context")
	}
	capt, ok := c.(*Capture)
	if !ok {
		return Capture{}, errors.New("value stored in context is not a *Capture")
	}
	return *capt, nil
}

// Capture is the object populated by the capture middleware,
// holding probes that allow to gather information about the request and response.
type Capture struct {
	rr  *readCounter
	crw *captureResponseWriter
}

// NeedsReset returns whether the given http.ResponseWriter is the capture's probe.
func (c *Capture) NeedsReset(rw http.ResponseWriter) bool {
	// This comparison is naive.
	return c.crw != rw
}

// Reset returns a new handler that renews the Capture's probes, and inserts
// them when deferring to next.
func (c *Capture) Reset(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		newRW, newReq := c.renew(rw, req)
		next.ServeHTTP(newRW, newReq)
	})
}

func (c *Capture) renew(rw http.ResponseWriter, req *http.Request) (http.ResponseWriter, *http.Request) {
	ctx := context.WithValue(req.Context(), capturedData, c)
	newReq := req.WithContext(ctx)

	if newReq.Body != nil {
		readCounter := &readCounter{source: newReq.Body}
		c.rr = readCounter
		newReq.Body = readCounter
	}
	c.crw = &captureResponseWriter{rw: rw}

	return c.crw, newReq
}

func (c *Capture) ResponseSize() int64 {
	return c.crw.Size()
}

func (c *Capture) StatusCode() int {
	return c.crw.Status()
}

// RequestSize returns the size of the request's body if it applies,
// zero otherwise.
func (c *Capture) RequestSize() int64 {
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

var _ middlewares.Stateful = &captureResponseWriter{}

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
