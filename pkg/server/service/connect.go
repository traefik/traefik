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

	bodyDeferrer := newBodyDeferrer(req.Body)
	req.Body = bodyDeferrer

	h.next.ServeHTTP(&connectResponseWriter{ResponseWriter: rw, req: req, bodyDeferrer: bodyDeferrer}, req)
}

// bodyDeferrer holds the body of a CONNECT request until the backend has accepted the tunnel.
type bodyDeferrer struct {
	*io.PipeReader

	writer      *io.PipeWriter
	body        io.ReadCloser
	releaseOnce sync.Once
}

func newBodyDeferrer(body io.ReadCloser) *bodyDeferrer {
	pipeReader, pipeWriter := io.Pipe()

	return &bodyDeferrer{
		PipeReader: pipeReader,
		writer:     pipeWriter,
		body:       body,
	}
}

// Close the deferrer.
func (bd *bodyDeferrer) Close() error {
	var errs []error
	if err := bd.writer.Close(); err != nil {
		errs = append(errs, fmt.Errorf("closing bodyDeferrer pipe writer: %w", err))
	}
	if err := bd.PipeReader.Close(); err != nil {
		errs = append(errs, fmt.Errorf("closing bodyDeferrer pipe reader: %w", err))
	}
	if err := bd.body.Close(); err != nil {
		errs = append(errs, fmt.Errorf("closing bodyDeferrer request body: %w", err))
	}

	return errors.Join(errs...)
}

// Release forwards the deferred request body to the backend and copy the subsequent bytes to the tunnel.
// As described in https://datatracker.ietf.org/doc/html/rfc9931#name-requirements-for-http-conne we must wait for a 2xx (Successful)
// response before forwarding any tunnel data.
func (bd *bodyDeferrer) Release(req *http.Request) {
	bd.releaseOnce.Do(func() {
		// Forward the tunnel data to the backend in the background: for an established tunnel this copy runs
		// for the lifetime of the tunnel, so release must return to let the response direction be pumped.
		copyDoneCh := make(chan struct{})
		go func() {
			_, err := io.Copy(bd.writer, bd.body)
			_ = bd.writer.CloseWithError(err)
			close(copyDoneCh)
		}()

		// If the request is canceled, the payload must not reach the backend, so close the pipe to unblock the Transport.
		go func() {
			select {
			case <-req.Context().Done():
				bd.Close()
			case <-copyDoneCh:
			}
		}()
	})
}

// connectResponseWriter releases the deferred CONNECT payload as soon as the backend's response status is known,
// then behaves as the wrapped ResponseWriter.
type connectResponseWriter struct {
	http.ResponseWriter

	req          *http.Request
	bodyDeferrer *bodyDeferrer
}

func (w *connectResponseWriter) WriteHeader(statusCode int) {
	// The tunnel was refused, so the request body never becomes tunnel data and must not reach the backend.
	if statusCode/100 != 2 {
		_ = w.bodyDeferrer.Close()
	} else {
		w.bodyDeferrer.Release(w.req)
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
