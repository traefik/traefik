package api

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_safePrefix(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected string
	}{
		{
			desc:     "host",
			value:    "https://example.com",
			expected: "",
		},
		{
			desc:     "host with path",
			value:    "https://example.com/foo/bar?test",
			expected: "",
		},
		{
			desc:     "path",
			value:    "/foo/bar",
			expected: "/foo/bar",
		},
		{
			desc:     "path without leading slash",
			value:    "foo/bar",
			expected: "foo/bar",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
			require.NoError(t, err)

			req.Header.Set("X-Forwarded-Prefix", test.value)

			prefix := safePrefix(req)

			assert.Equal(t, test.expected, prefix)
		})
	}
}

func Test_ContentSecurityPolicy(t *testing.T) {
	testCases := []struct {
		desc     string
		handler  DashboardHandler
		expected int
	}{
		{
			desc: "OK",
			handler: DashboardHandler{
				FS: fsMock(func(name string) (fs.File, error) {
					return fileMock{name: name}, nil
				}),
			},
			expected: http.StatusOK,
		},
		{
			desc: "Not found",
			handler: DashboardHandler{
				FS: fsMock(func(name string) (fs.File, error) {
					return nil, fs.ErrNotExist
				}),
			},
			expected: http.StatusNotFound,
		},
		{
			desc: "Internal server error",
			handler: DashboardHandler{
				FS: fsMock(func(name string) (fs.File, error) {
					return nil, errors.New("oops")
				}),
			},
			expected: http.StatusInternalServerError,
		},
	}

	for _, test := range testCases {
		test := test
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

type fsMock func(name string) (fs.File, error)

func (m fsMock) Open(name string) (fs.File, error) {
	return m(name)
}

type fileMock struct {
	name    string
	content []byte
}

func (f fileMock) Stat() (fs.FileInfo, error) {
	return fileInfoMock{name: f.name}, nil
}

func (f fileMock) Read(bytes []byte) (int, error) {
	n := copy(bytes, f.content)
	return n, nil
}

func (f fileMock) Close() error {
	return nil
}

type fileInfoMock struct {
	name string
}

func (f fileInfoMock) Name() string       { return f.name }
func (f fileInfoMock) Size() int64        { return 0 }
func (f fileInfoMock) Mode() fs.FileMode  { return 0 }
func (f fileInfoMock) ModTime() time.Time { return time.Time{} }
func (f fileInfoMock) IsDir() bool        { return false }
func (f fileInfoMock) Sys() interface{}   { return nil }
