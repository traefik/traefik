package testhelpers

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
)

// UnwrapOnlyResponseWriter only exposes the wrapped ResponseWriter through Unwrap,
// like Traefik's ResponseWriter wrappers which do not implement Flush and Hijack.
type UnwrapOnlyResponseWriter struct {
	http.ResponseWriter
}

// Unwrap returns the wrapped ResponseWriter.
func (w *UnwrapOnlyResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// HijackableRecorder is an httptest.ResponseRecorder which records Hijack calls.
type HijackableRecorder struct {
	*httptest.ResponseRecorder

	Hijacked bool
}

// NewHijackableRecorder returns an initialized HijackableRecorder.
func NewHijackableRecorder() *HijackableRecorder {
	return &HijackableRecorder{ResponseRecorder: httptest.NewRecorder()}
}

// Hijack records the call, the returned connection is nil.
func (r *HijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	r.Hijacked = true
	return nil, nil, nil
}

// ResponseWriterSupport lists the optional ResponseWriter methods used to reach the underlying connection.
type ResponseWriterSupport struct {
	Flush  bool
	Hijack bool
	Unwrap bool
}

// SupportOf returns the optional methods implemented by the given ResponseWriter.
func SupportOf(rw http.ResponseWriter) ResponseWriterSupport {
	_, flush := rw.(http.Flusher)
	_, hijack := rw.(http.Hijacker)
	_, unwrap := rw.(interface{ Unwrap() http.ResponseWriter })

	return ResponseWriterSupport{Flush: flush, Hijack: hijack, Unwrap: unwrap}
}
