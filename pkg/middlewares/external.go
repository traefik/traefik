package middlewares

import (
	"bufio"
	"context"
	"net"
	"net/http"
)

type entryWriterKey struct{}

type flushErrorer interface {
	FlushError() error
}

type unwrapper interface {
	Unwrap() http.ResponseWriter
}

// WrapExternal builds, with the given build function, a handler we do not control (third-party middleware, plugin) around next.
// The external handler receives a ResponseWriter implementing http.Flusher and http.Hijacker,
// as such code asserts them directly instead of using http.ResponseController.
// The ResponseWriter it gives to next may not implement Flush, Hijack or Unwrap:
// in that case the call falls back to the ResponseWriter given to the external handler, so that the chain is not broken.
func WrapExternal(next http.Handler, build func(next http.Handler) (http.Handler, error)) (http.Handler, error) {
	handler, _, err := WrapExternalWith(next, func(next http.Handler) (http.Handler, struct{}, error) {
		external, err := build(next)
		return external, struct{}{}, err
	})

	return handler, err
}

// WrapExternalWith is WrapExternal for a build function which also returns a value, such as a context modifier, returned as is.
func WrapExternalWith[T any](next http.Handler, build func(next http.Handler) (http.Handler, T, error)) (http.Handler, T, error) {
	external, value, err := build(exitHandler(next))
	if err != nil {
		var zero T
		return nil, zero, err
	}

	return entryHandler(external), value, nil
}

func entryHandler(external http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		entry := newEntryWriter(rw)
		external.ServeHTTP(entry, req.WithContext(context.WithValue(req.Context(), entryWriterKey{}, entry)))
	})
}

func exitHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		entry, ok := req.Context().Value(entryWriterKey{}).(http.ResponseWriter)
		if !ok || rw == entry {
			next.ServeHTTP(rw, req)
			return
		}

		next.ServeHTTP(&exitWriter{ResponseWriter: rw, entry: entry}, req)
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

// exitWriter wraps the ResponseWriter created by an external handler.
type exitWriter struct {
	http.ResponseWriter

	// entry is the ResponseWriter given to the external handler.
	entry http.ResponseWriter
}

// Flush sends any buffered data to the client.
func (w *exitWriter) Flush() {
	_ = w.FlushError()
}

// FlushError sends any buffered data to the client, and is preferred over Flush by http.ResponseController.
func (w *exitWriter) FlushError() error {
	target := w.entry
	switch w.ResponseWriter.(type) {
	case flushErrorer, http.Flusher, unwrapper:
		target = w.ResponseWriter
	}

	return http.NewResponseController(target).Flush()
}

// Hijack hijacks the connection.
func (w *exitWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	target := w.entry
	switch w.ResponseWriter.(type) {
	case http.Hijacker, unwrapper:
		target = w.ResponseWriter
	}

	return http.NewResponseController(target).Hijack()
}

// Unwrap exposes the external ResponseWriter when it can itself be unwrapped,
// and the entry ResponseWriter otherwise, for http.ResponseController.
func (w *exitWriter) Unwrap() http.ResponseWriter {
	if _, ok := w.ResponseWriter.(unwrapper); ok {
		return w.ResponseWriter
	}

	return w.entry
}
