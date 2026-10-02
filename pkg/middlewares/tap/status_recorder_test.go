package tap

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusRecorder_status(t *testing.T) {
	testCases := []struct {
		desc           string
		write          func(rw http.ResponseWriter)
		expectedStatus int
	}{
		{
			desc:  "nothing written",
			write: func(http.ResponseWriter) {},
		},
		{
			desc: "status written",
			write: func(rw http.ResponseWriter) {
				rw.WriteHeader(http.StatusAccepted)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			desc: "body written without a status",
			write: func(rw http.ResponseWriter) {
				_, _ = rw.Write([]byte("ok"))
			},
			expectedStatus: http.StatusOK,
		},
		{
			desc: "body written after a status",
			write: func(rw http.ResponseWriter) {
				rw.WriteHeader(http.StatusServiceUnavailable)
				_, _ = rw.Write([]byte("unavailable"))
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			desc: "status written twice",
			write: func(rw http.ResponseWriter) {
				rw.WriteHeader(http.StatusAccepted)
				rw.WriteHeader(http.StatusServiceUnavailable)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			desc: "status written after a body",
			write: func(rw http.ResponseWriter) {
				_, _ = rw.Write([]byte("ok"))
				rw.WriteHeader(http.StatusServiceUnavailable)
			},
			expectedStatus: http.StatusOK,
		},
		{
			// A tap service accepting the record only once it is read must not have it taken
			// for accepted on its interim response.
			desc: "informational status written before a final one",
			write: func(rw http.ResponseWriter) {
				rw.WriteHeader(http.StatusContinue)
				rw.WriteHeader(http.StatusEarlyHints)
				rw.WriteHeader(http.StatusServiceUnavailable)
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			desc: "informational status only",
			write: func(rw http.ResponseWriter) {
				rw.WriteHeader(http.StatusContinue)
			},
		},
		{
			desc: "informational status written before a body",
			write: func(rw http.ResponseWriter) {
				rw.WriteHeader(http.StatusEarlyHints)
				_, _ = rw.Write([]byte("ok"))
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			recorder := &statusRecorder{header: make(http.Header)}
			test.write(recorder)

			assert.Equal(t, test.expectedStatus, recorder.status)
		})
	}
}

func TestStatusRecorder_discardsBody(t *testing.T) {
	recorder := &statusRecorder{header: make(http.Header)}

	n, err := recorder.Write([]byte("pong"))
	require.NoError(t, err)

	// The whole body is reported as written, so that the tap service does not fail on it.
	assert.Equal(t, 4, n)
}

func TestStatusRecorder_header(t *testing.T) {
	header := make(http.Header)
	recorder := &statusRecorder{header: header}

	recorder.Header().Set("X-Tap", "yes")

	assert.Equal(t, "yes", header.Get("X-Tap"))
}

func TestStatusRecorder_flush(t *testing.T) {
	recorder := &statusRecorder{header: make(http.Header)}

	recorder.Flush()

	assert.Zero(t, recorder.status)
}

func TestStatusRecorder_hijack(t *testing.T) {
	recorder := &statusRecorder{header: make(http.Header)}

	conn, rw, err := recorder.Hijack()
	require.Error(t, err)

	assert.Nil(t, conn)
	assert.Nil(t, rw)
}
