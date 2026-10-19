package middlewares

import (
	"bufio"
	"net"
	"net/http"
)

// WrapExternal builds, with the given build function, a handler we do not control (third-party middleware, plugin) around next.
// The external handler receives a ResponseWriter implementing http.Flusher and http.Hijacker,
// as such code asserts them directly instead of using http.ResponseController.
// The ResponseWriter it gives to next is passed as is: if it does not implement Flush, Hijack or Unwrap,
// the external handler deliberately breaks the chain (e.g. a buffering writer must not be flushed past).
func WrapExternal(next http.Handler, build func(next http.Handler) (http.Handler, error)) (http.Handler, error) {
	handler, _, err := WrapExternalWith(next, func(next http.Handler) (http.Handler, struct{}, error) {
		external, err := build(next)
		return external, struct{}{}, err
	})

	return handler, err
}

// WrapExternalWith is WrapExternal for a build function which also returns a value, such as a context modifier, returned as is.
func WrapExternalWith[T any](next http.Handler, build func(next http.Handler) (http.Handler, T, error)) (http.Handler, T, error) {
	external, value, err := build(next)
	if err != nil {
		var zero T
		return nil, zero, err
	}

	return entryHandler(external), value, nil
}

func entryHandler(external http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		external.ServeHTTP(newEntryWriter(rw), req)
	})
}

// entryWriter exposes http.Flusher and http.Hijacker backed by http.ResponseController.
type entryWriter struct {
	http.ResponseWriter

	rc *http.ResponseController
}

// newEntryWriter wraps the given ResponseWriter in an entryWriter, unless it already is one.
func newEntryWriter(rw http.ResponseWriter) http.ResponseWriter {
	if ew, ok := rw.(*entryWriter); ok {
		return ew
	}

	return &entryWriter{ResponseWriter: rw, rc: http.NewResponseController(rw)}
}

// Flush sends any buffered data to the client.
func (w *entryWriter) Flush() {
	_ = w.FlushError()
}

// FlushError sends any buffered data to the client, and is preferred over Flush by http.ResponseController.
func (w *entryWriter) FlushError() error {
	return w.rc.Flush()
}

// Hijack hijacks the connection.
func (w *entryWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.rc.Hijack()
}

// Unwrap exposes the underlying ResponseWriter for http.ResponseController.
func (w *entryWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
