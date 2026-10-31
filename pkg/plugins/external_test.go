package plugins

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unwrapOnlyWriter only exposes the wrapped writer through Unwrap.
type unwrapOnlyWriter struct {
	http.ResponseWriter
}

func (w *unwrapOnlyWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// controllerWriter records the deadlines, full duplex and hijack calls reaching it.
type controllerWriter struct {
	http.ResponseWriter

	readDeadline  time.Time
	writeDeadline time.Time
	fullDuplex    bool
	hijacked      bool
}

func (w *controllerWriter) SetReadDeadline(deadline time.Time) error {
	w.readDeadline = deadline
	return nil
}

func (w *controllerWriter) SetWriteDeadline(deadline time.Time) error {
	w.writeDeadline = deadline
	return nil
}

func (w *controllerWriter) EnableFullDuplex() error {
	w.fullDuplex = true
	return nil
}

func (w *controllerWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	return nil, nil, nil
}

// bareWriter hides every optional interface of the wrapped writer.
type bareWriter struct {
	http.ResponseWriter
}

// flushHijackWriter handles Flush and Hijack itself.
type flushHijackWriter struct {
	http.ResponseWriter
}

func (w *flushHijackWriter) Flush() {}

func (w *flushHijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

func TestEntryWriter_Flush(t *testing.T) {
	recorder := httptest.NewRecorder()

	rw := newEntryWriter(&unwrapOnlyWriter{ResponseWriter: recorder})

	flusher, ok := rw.(http.Flusher)
	require.True(t, ok)

	flusher.Flush()

	assert.True(t, recorder.Flushed)
}

func TestEntryWriter_Hijack(t *testing.T) {
	rw := newEntryWriter(&unwrapOnlyWriter{ResponseWriter: httptest.NewRecorder()})

	hijacker, ok := rw.(http.Hijacker)
	require.True(t, ok)

	_, _, err := hijacker.Hijack()

	assert.ErrorIs(t, err, http.ErrNotSupported)
}

func TestNewEntryWriter_wrapsOnce(t *testing.T) {
	rw := newEntryWriter(httptest.NewRecorder())

	assert.Same(t, rw, newEntryWriter(rw))
}

func TestWrapExternal(t *testing.T) {
	testCases := []struct {
		desc                 string
		externalWriter       func(rw http.ResponseWriter) http.ResponseWriter
		expectExternalCalls  bool
		expectExternalHijack bool
		expectFlushErr       error
	}{
		{
			desc: "external writer implements nothing",
			externalWriter: func(rw http.ResponseWriter) http.ResponseWriter {
				return &bareWriter{ResponseWriter: rw}
			},
			expectFlushErr: http.ErrNotSupported,
		},
		{
			desc: "external writer implements Flush and Hijack",
			externalWriter: func(rw http.ResponseWriter) http.ResponseWriter {
				return &flushHijackWriter{ResponseWriter: rw}
			},
			expectExternalHijack: true,
		},
		{
			desc: "external writer implements deadlines, full duplex and Hijack",
			externalWriter: func(rw http.ResponseWriter) http.ResponseWriter {
				return &controllerWriter{ResponseWriter: rw}
			},
			expectExternalCalls:  true,
			expectExternalHijack: true,
			expectFlushErr:       http.ErrNotSupported,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			deadline := time.Now().Add(time.Minute)

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rc := http.NewResponseController(rw)

				assert.NoError(t, rc.SetReadDeadline(deadline))
				assert.NoError(t, rc.SetWriteDeadline(deadline))
				assert.NoError(t, rc.EnableFullDuplex())

				assert.ErrorIs(t, rc.Flush(), test.expectFlushErr)

				_, _, err := rc.Hijack()
				assert.NoError(t, err)
			})

			var externalWriter http.ResponseWriter
			handler, err := wrapExternal(next, func(next http.Handler) (http.Handler, error) {
				return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					externalWriter = test.externalWriter(rw)
					next.ServeHTTP(externalWriter, req)
				}), nil
			})
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			root := &controllerWriter{ResponseWriter: recorder}
			handler.ServeHTTP(root, httptest.NewRequest(http.MethodGet, "/", nil))

			assert.False(t, recorder.Flushed)
			assert.Equal(t, !test.expectExternalHijack, root.hijacked)

			called := root
			if test.expectExternalCalls {
				called = externalWriter.(*controllerWriter)

				assert.Zero(t, root.readDeadline)
				assert.Zero(t, root.writeDeadline)
				assert.False(t, root.fullDuplex)
			}

			assert.Equal(t, deadline, called.readDeadline)
			assert.Equal(t, deadline, called.writeDeadline)
			assert.True(t, called.fullDuplex)
		})
	}
}

func TestWrapExternal_entryWriterPassedThrough(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, ok := rw.(*entryWriter)
		assert.True(t, ok)
	})

	handler, err := wrapExternal(next, func(next http.Handler) (http.Handler, error) {
		return next, nil
	})
	require.NoError(t, err)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}
