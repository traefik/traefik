package service

import (
	"bufio"
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
	if req.ContentLength > 0 || req.Body == nil {
		h.next.ServeHTTP(rw, req)
		return
	}

	pipeReader, pipeWriter := io.Pipe()
	tunnel := &connectTunnel{in: pipeWriter, data: req.Body}

	// The Transport blocks on the empty pipe, so only the header section reaches the backend until
	// connectResponseWriter releases the payload once the backend accepts the tunnel.
	req.Body = pipeReader

	h.next.ServeHTTP(&connectResponseWriter{ResponseWriter: rw, req: req, tunnel: tunnel}, req)
}

// connectTunnel holds the data of a CONNECT request until the backend has accepted the tunnel.
type connectTunnel struct {
	in          *io.PipeWriter
	releaseOnce sync.Once
	data        io.Reader
}

// Close closes tunnel.
func (t *connectTunnel) Close() {
	_ = t.in.Close()
}

// Release forwards the deferred request body to the backend and copy the subsequent bytes to the tunnel.
// As described in https://datatracker.ietf.org/doc/html/rfc9931#name-requirements-for-http-conne we must wait for a 2xx (Successful)
// response before forwarding any tunnel data.
func (t *connectTunnel) Release(req *http.Request) {
	t.releaseOnce.Do(func() {
		// Forward the tunnel data to the backend in the background: for an established tunnel this copy runs
		// for the lifetime of the tunnel, so release must return to let the response direction be pumped.
		copyDoneCh := make(chan struct{})
		go func() {
			_, err := io.Copy(t.in, t.data)
			_ = t.in.CloseWithError(err)
			close(copyDoneCh)
		}()

		// If the request is canceled, the payload must not reach the backend, so close the pipe to unblock the Transport.
		go func() {
			select {
			case <-req.Context().Done():
				_ = t.in.Close()
			case <-copyDoneCh:
			}
		}()
	})
}

// connectResponseWriter releases the deferred CONNECT payload as soon as the backend's response status is known,
// then behaves as the wrapped ResponseWriter.
type connectResponseWriter struct {
	http.ResponseWriter

	req    *http.Request
	tunnel *connectTunnel
}

func (w *connectResponseWriter) WriteHeader(statusCode int) {
	// The tunnel was refused, so the request body never becomes tunnel data and must not reach the backend.
	if statusCode/100 != 2 {
		w.tunnel.Close()
	} else {
		w.tunnel.Release(w.req)
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
