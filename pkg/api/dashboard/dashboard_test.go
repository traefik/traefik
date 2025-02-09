package dashboard

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_ContentSecurityPolicy(t *testing.T) {
	testCases := []struct {
		desc     string
		handler  Handler
		expected int
	}{
		{
			desc: "OK",
			handler: Handler{
				assets: fstest.MapFS{"foobar.html": &fstest.MapFile{
					Mode:    0o755,
					ModTime: time.Now(),
				}},
			},
			expected: http.StatusOK,
		},
		{
			desc: "Not found",
			handler: Handler{
				assets: fstest.MapFS{},
			},
			expected: http.StatusNotFound,
		},
		{
			desc: "Internal server error",
			handler: Handler{
				assets: errorFS{},
			},
			expected: http.StatusInternalServerError,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/foobar.html", nil)

			rw := httptest.NewRecorder()

			test.handler.ServeHTTP(rw, req)

			assert.Equal(t, test.expected, rw.Code)
			assert.Equal(t, "frame-src 'self' https://traefik.io https://*.traefik.io;", rw.Result().Header.Get("Content-Security-Policy"))
		})
	}
}

type errorFS struct{}

func (e errorFS) Open(name string) (fs.File, error) {
	return nil, errors.New("oops")
}
