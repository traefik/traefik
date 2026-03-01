package recovery

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
