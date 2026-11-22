package plugins

import (
	"bufio"
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

type entryWriterKey struct{}

// wrapExternal builds, with the given build function, a plugin handler (Yaegi, Wasm), whose code we do not control, around next.
// The external handler receives a ResponseWriter implementing http.Flusher and http.Hijacker,
// as such code asserts them directly instead of using http.ResponseController.
// The ResponseWriter it gives to next may not implement Unwrap:
// it is wrapped to reach the deadlines, full duplex and hijacking of the ResponseWriter given to the external handler,
// when the external ResponseWriter does not support them.
// Flush is left to the external ResponseWriter: a buffering writer must not be flushed past.
func wrapExternal(next http.Handler, build func(next http.Handler) (http.Handler, error)) (http.Handler, error) {
	handler, _, err := wrapExternalWith(next, func(next http.Handler) (http.Handler, struct{}, error) {
		external, err := build(next)
		return external, struct{}{}, err
	})

	return handler, err
}

// wrapExternalWith is wrapExternal for a build function which also returns a value, such as a context modifier, returned as is.
func wrapExternalWith[T any](next http.Handler, build func(next http.Handler) (http.Handler, T, error)) (http.Handler, T, error) {
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

// SetReadDeadline sets the read deadline through the external ResponseWriter, or the entry ResponseWriter if it does not support it.
func (w *exitWriter) SetReadDeadline(deadline time.Time) error {
	if err := http.NewResponseController(w.ResponseWriter).SetReadDeadline(deadline); !errors.Is(err, http.ErrNotSupported) {
		return err
	}

	return http.NewResponseController(w.entry).SetReadDeadline(deadline)
}

// SetWriteDeadline sets the write deadline through the external ResponseWriter, or the entry ResponseWriter if it does not support it.
func (w *exitWriter) SetWriteDeadline(deadline time.Time) error {
	if err := http.NewResponseController(w.ResponseWriter).SetWriteDeadline(deadline); !errors.Is(err, http.ErrNotSupported) {
		return err
	}

	return http.NewResponseController(w.entry).SetWriteDeadline(deadline)
}

// EnableFullDuplex enables full duplex through the external ResponseWriter, or the entry ResponseWriter if it does not support it.
func (w *exitWriter) EnableFullDuplex() error {
	if err := http.NewResponseController(w.ResponseWriter).EnableFullDuplex(); !errors.Is(err, http.ErrNotSupported) {
		return err
	}

	return http.NewResponseController(w.entry).EnableFullDuplex()
}

// Hijack hijacks the connection through the external ResponseWriter, or the entry ResponseWriter if it does not support it.
// Unlike flushing past it, hijacking past the external ResponseWriter cannot corrupt the response:
// the writes it does afterward fail with http.ErrHijacked.
func (w *exitWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if conn, brw, err := http.NewResponseController(w.ResponseWriter).Hijack(); !errors.Is(err, http.ErrNotSupported) {
		return conn, brw, err
	}

	return http.NewResponseController(w.entry).Hijack()
}

// Unwrap exposes the external ResponseWriter, so that http.ResponseController uses its own Flush.
func (w *exitWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
