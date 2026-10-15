package recovery

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecoverHandler(t *testing.T) {
	tests := []struct {
		desc        string
		panicErr    error
		headersSent bool
	}{
		{
			desc:        "headers sent and custom panic error",
			panicErr:    errors.New("foo"),
			headersSent: true,
		},
		{
			desc:        "headers sent and error abort handler",
			panicErr:    http.ErrAbortHandler,
			headersSent: true,
		},
		{
			desc:     "custom panic error",
			panicErr: errors.New("foo"),
		},
		{
			desc:     "error abort handler",
			panicErr: http.ErrAbortHandler,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			fn := func(rw http.ResponseWriter, req *http.Request) {
				if test.headersSent {
					rw.WriteHeader(http.StatusTeapot)
				}
				panic(test.panicErr)
			}
			recovery, err := New(t.Context(), http.HandlerFunc(fn))
			require.NoError(t, err)

			server := httptest.NewServer(recovery)
			t.Cleanup(server.Close)

			res, err := http.Get(server.URL)
			if test.headersSent {
				require.Nil(t, res)
				assert.ErrorIs(t, err, io.EOF)
			} else {
				require.NoError(t, err)
				assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
			}
		})
	}
}

// TestPanicMessageAndStackAreOneEvent guards against a regression where the panic
// message and its stack trace were written as two separate log events. On
// shutdown a panic can race process exit; the second write (the stack) is then
// lost, leaving an undiagnosable "Recovered from panic" with no trace, as in
// #13693. Emitting both in a single event keeps the trace with the message.
func TestPanicMessageAndStackAreOneEvent(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.ErrorLevel)
	ctx := logger.WithContext(context.Background())

	recovery, err := New(ctx, http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic(errors.New("boom"))
	}))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/foo", nil).WithContext(ctx)
	recovery.ServeHTTP(httptest.NewRecorder(), req)

	// The panic message and its stack trace must live in the same log event.
	var panicked bool
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if !strings.Contains(line, "Recovered from panic") {
			continue
		}

		panicked = true
		assert.Contains(t, line, "Stack:", "the stack trace must be in the same event as the panic message")
	}
	assert.True(t, panicked, "expected a panic log event to be written")
}
