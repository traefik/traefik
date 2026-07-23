package service

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
)

// connectHandler defers the payload of a CONNECT request until the backend has accepted the tunnel.
type connectHandler struct {
	next http.Handler
}

// newConnectHandler wraps next with the CONNECT payload deferral behavior.
func newConnectHandler(next http.Handler) http.Handler {
	return &connectHandler{next: next}
}

func (h *connectHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodConnect {
		h.next.ServeHTTP(rw, req)
		return
	}

	// Tunneling is only supported for clients speaking HTTP/2 and above.
	if req.ProtoMajor == 1 {
		rw.WriteHeader(http.StatusNotImplemented)
		return
	}

	// Nothing to defer for a CONNECT without body or with a fixed Content-Length.
	if req.ContentLength >= 0 || req.Body == nil || req.Body == http.NoBody {
		h.next.ServeHTTP(rw, req)
		return
	}

	bodyDeferrer := newBodyDeferrer(req.Context().Done(), req.Body)
	req.Body = bodyDeferrer

	h.next.ServeHTTP(&connectResponseWriter{ResponseWriter: rw, bodyDeferrer: bodyDeferrer}, req)
}

// bodyDeferrer holds the body of a CONNECT request until the backend has accepted the tunnel.
type bodyDeferrer struct {
	body             io.ReadCloser
	doneCh           <-chan struct{}
	releaseCh        chan struct{}
	closeReleaseOnce func()
}

func newBodyDeferrer(doneCh <-chan struct{}, body io.ReadCloser) *bodyDeferrer {
	releaseCh := make(chan struct{})

	return &bodyDeferrer{
		body:      body,
		doneCh:    doneCh,
		releaseCh: releaseCh,
		closeReleaseOnce: sync.OnceFunc(func() {
			close(releaseCh)
		}),
	}
}

func (bd *bodyDeferrer) Read(p []byte) (n int, err error) {
	select {
	case <-bd.doneCh:
		return 0, errors.New("request context canceled")
	case <-bd.releaseCh:
	}
	return bd.body.Read(p)
}

// Close closes the underlying request body before releasing the deferred request body read operation.
func (bd *bodyDeferrer) Close() error {
	defer bd.closeReleaseOnce()
	return bd.body.Close()
}

// Release forwards the deferred request body to the backend and copy the subsequent bytes to the tunnel.
// As described in https://datatracker.ietf.org/doc/html/rfc9931#name-requirements-for-http-conne we must wait for a 2xx (Successful)
// response before forwarding any tunnel data.
func (bd *bodyDeferrer) Release() {
	bd.closeReleaseOnce()
}

// connectResponseWriter releases the deferred CONNECT payload as soon as the backend's response status is known,
// then behaves as the wrapped ResponseWriter.
type connectResponseWriter struct {
	http.ResponseWriter

	bodyDeferrer *bodyDeferrer
}

func (w *connectResponseWriter) WriteHeader(statusCode int) {
	// The tunnel was refused, so the request body never becomes tunnel data and must not reach the backend.
	if statusCode/100 != 2 {
		_ = w.bodyDeferrer.Close()
	} else {
		w.bodyDeferrer.Release()
	}

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *connectResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *connectResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}

	return nil, nil, fmt.Errorf("not a hijacker: %T", w.ResponseWriter)
}

func (w *connectResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
