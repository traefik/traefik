package middlewares

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

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

// bareWriter hides every optional interface of the wrapped writer.
type bareWriter struct {
	http.ResponseWriter
}

type flushWriter struct {
	http.ResponseWriter

	flushed bool
}

func (w *flushWriter) Flush() {
	w.flushed = true
}

type hijackWriter struct {
	http.ResponseWriter

	hijacked bool
}

func (w *hijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
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

func TestWrapExternal_Flush(t *testing.T) {
	testCases := []struct {
		desc               string
		externalWriter     func(rw http.ResponseWriter) http.ResponseWriter
		expectEntryFlushed bool
	}{
		{
			desc: "external writer implements Flush",
			externalWriter: func(rw http.ResponseWriter) http.ResponseWriter {
				return &flushWriter{ResponseWriter: rw}
			},
		},
		{
			desc: "external writer implements Unwrap",
			externalWriter: func(rw http.ResponseWriter) http.ResponseWriter {
				return &unwrapOnlyWriter{ResponseWriter: rw}
			},
			expectEntryFlushed: true,
		},
		{
			desc: "external writer implements neither",
			externalWriter: func(rw http.ResponseWriter) http.ResponseWriter {
				return &bareWriter{ResponseWriter: rw}
			},
			expectEntryFlushed: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.NoError(t, http.NewResponseController(rw).Flush())
			})

			var externalWriter http.ResponseWriter
			external := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					externalWriter = test.externalWriter(rw)
					next.ServeHTTP(externalWriter, req)
				})
			}

			recorder := httptest.NewRecorder()
			mustWrapExternal(t, next, external).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

			assert.Equal(t, test.expectEntryFlushed, recorder.Flushed)
			if fw, ok := externalWriter.(*flushWriter); ok {
				assert.True(t, fw.flushed)
			}
		})
	}
}

func TestWrapExternal_Hijack(t *testing.T) {
	testCases := []struct {
		desc                string
		externalWriter      func(rw http.ResponseWriter) http.ResponseWriter
		expectEntryHijacked bool
	}{
		{
			desc: "external writer implements Hijack",
			externalWriter: func(rw http.ResponseWriter) http.ResponseWriter {
				return &hijackWriter{ResponseWriter: rw}
			},
		},
		{
			desc: "external writer implements Unwrap",
			externalWriter: func(rw http.ResponseWriter) http.ResponseWriter {
				return &unwrapOnlyWriter{ResponseWriter: rw}
			},
			expectEntryHijacked: true,
		},
		{
			desc: "external writer implements neither",
			externalWriter: func(rw http.ResponseWriter) http.ResponseWriter {
				return &bareWriter{ResponseWriter: rw}
			},
			expectEntryHijacked: true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				_, _, err := http.NewResponseController(rw).Hijack()
				assert.NoError(t, err)
			})

			var externalWriter http.ResponseWriter
			external := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
					externalWriter = test.externalWriter(rw)
					next.ServeHTTP(externalWriter, req)
				})
			}

			root := &hijackWriter{ResponseWriter: httptest.NewRecorder()}
			mustWrapExternal(t, next, external).ServeHTTP(root, httptest.NewRequest(http.MethodGet, "/", nil))

			assert.Equal(t, test.expectEntryHijacked, root.hijacked)
			if hw, ok := externalWriter.(*hijackWriter); ok {
				assert.True(t, hw.hijacked)
			}
		})
	}
}

func TestWrapExternal_noWrapping(t *testing.T) {
	testCases := []struct {
		desc    string
		handler func(next http.Handler) http.Handler
	}{
		{
			desc: "external handler passes the entry writer",
			handler: func(next http.Handler) http.Handler {
				return mustWrapExternal(t, next, func(next http.Handler) http.Handler { return next })
			},
		},
		{
			desc:    "no entry writer in the context",
			handler: exitHandler,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				_, ok := rw.(*exitWriter)
				assert.False(t, ok)
			})

			test.handler(next).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		})
	}
}

func mustWrapExternal(t *testing.T, next http.Handler, external func(next http.Handler) http.Handler) http.Handler {
	t.Helper()

	handler, err := WrapExternal(next, func(next http.Handler) (http.Handler, error) {
		return external(next), nil
	})
	require.NoError(t, err)

	return handler
}
