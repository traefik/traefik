package middlewares

import (
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
