package loadbalancer

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusRecordingResponseWriter_IsServerError(t *testing.T) {
	rec := httptest.NewRecorder()
	srw := NewStatusRecordingResponseWriter(rec)

	_, err := srw.Write([]byte("ok"))
	require.NoError(t, err)
	assert.False(t, srw.IsServerError())

	srw.WriteHeader(http.StatusBadGateway)
	assert.True(t, srw.IsServerError())
}

func TestStatusRecordingResponseWriter_HijackPassthrough(t *testing.T) {
	t.Run("returns error when hijack unsupported", func(t *testing.T) {
		srw := NewStatusRecordingResponseWriter(httptest.NewRecorder())
		_, _, err := srw.Hijack()
		require.Error(t, err)
	})

	t.Run("delegates hijack when supported", func(t *testing.T) {
		expectedConn, serverConn := net.Pipe()
		t.Cleanup(func() {
			_ = expectedConn.Close()
			_ = serverConn.Close()
		})

		rw := &hijackWriter{conn: expectedConn, rw: bufio.NewReadWriter(bufio.NewReader(serverConn), bufio.NewWriter(serverConn))}
		srw := NewStatusRecordingResponseWriter(rw)

		conn, brw, err := srw.Hijack()
		require.NoError(t, err)
		assert.Equal(t, expectedConn, conn)
		assert.Equal(t, rw.rw, brw)
	})
}

type hijackWriter struct {
	http.ResponseWriter

	conn net.Conn
	rw   *bufio.ReadWriter
}

func (h *hijackWriter) Header() http.Header {
	return make(http.Header)
}

func (h *hijackWriter) Write(_ []byte) (int, error) {
	return 0, nil
}

func (h *hijackWriter) WriteHeader(_ int) {}

func (h *hijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return h.conn, h.rw, nil
}
