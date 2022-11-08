package brotli

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	smallTestBody = []byte("aaabbcaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbc")
	bigTestBody   = []byte("aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc aaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbcccaaabbbccc")
)

func Test_Vary(t *testing.T) {
	h := newTestHandler(t, smallTestBody)

	req, _ := http.NewRequest(http.MethodGet, "/whatever", nil)
	req.Header.Set(acceptEncoding, "br")

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, acceptEncoding, rw.Header().Get(vary))
}

func Test_SmallBodyNoCompression(t *testing.T) {
	h := newTestHandler(t, smallTestBody)

	req, _ := http.NewRequest(http.MethodGet, "/whatever", nil)
	req.Header.Set(acceptEncoding, "br")

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	// With less than 1024 bytes the response should not be compressed.

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, "", rw.Header().Get(contentEncoding))
	assert.Equal(t, smallTestBody, rw.Body.Bytes())
}

func Test_AlreadyCompressed(t *testing.T) {
	h := newTestHandler(t, bigTestBody)

	req, _ := http.NewRequest(http.MethodGet, "/compressed", nil)
	req.Header.Set(acceptEncoding, "br")

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assert.Equal(t, bigTestBody, rw.Body.Bytes())
}

func Test_NoBody(t *testing.T) {
	testCases := []struct {
		desc            string
		statusCode      int
		contentEncoding string
		body            []byte
	}{
		{
			desc:            "status no content",
			statusCode:      http.StatusNoContent,
			contentEncoding: "",
			body:            nil,
		},
		{
			desc:            "status not modified",
			statusCode:      http.StatusNotModified,
			contentEncoding: "",
			body:            nil,
		},
		{
			desc:            "status OK with empty body",
			statusCode:      http.StatusOK,
			contentEncoding: "",
			body:            []byte{},
		},
		{
			desc:            "status OK with nil body",
			statusCode:      http.StatusOK,
			contentEncoding: "",
			body:            nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			h := newMiddleware(t, Config{MinSize: 1024})(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(test.statusCode)
				if _, err := rw.Write(test.body); err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(acceptEncoding, "br")

			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			body, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			assert.Equal(t, test.contentEncoding, rw.Header().Get(contentEncoding))
			assert.Equal(t, 0, len(body))
		})
	}
}

func Test_MinSize(t *testing.T) {
	bodySize := 0
	h := newMiddleware(t, Config{MinSize: 128})(http.HandlerFunc(
		func(rw http.ResponseWriter, req *http.Request) {
			for i := 0; i < bodySize; i++ {
				// We make sure to Write at least once less than minSize so that both
				// cases below go through the same algo: i.e. they start buffering
				// because they haven't reached minSize.
				if _, err := rw.Write([]byte{'x'}); err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}
			}
		},
	))

	req, _ := http.NewRequest(http.MethodGet, "/whatever", &bytes.Buffer{})
	req.Header.Add(acceptEncoding, "br")

	// Short response is not compressed
	bodySize = 127
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assert.Equal(t, "", rw.Result().Header.Get(contentEncoding))

	// Long response is compressed
	bodySize = 128
	rw = httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assert.Equal(t, "br", rw.Result().Header.Get(contentEncoding))
}

func Test_WriteHeader(t *testing.T) {
	h := newMiddleware(t, Config{MinSize: 1024})(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// We ensure that the subsequent call to WriteHeader is a noop.
		rw.WriteHeader(http.StatusInternalServerError)
		rw.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(acceptEncoding, "br")

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}

func Test_FlushBeforeWrite(t *testing.T) {
	srv := httptest.NewServer(newMiddleware(t, Config{MinSize: 1024})(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		rw.(http.Flusher).Flush()

		if _, err := rw.Write(bigTestBody); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
	require.NoError(t, err)

	req.Header.Set(acceptEncoding, "br")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "br", res.Header.Get(contentEncoding))

	got, err := io.ReadAll(brotli.NewReader(res.Body))
	require.NoError(t, err)
	assert.Equal(t, bigTestBody, got)
}

func Test_FlushAfterWrite(t *testing.T) {
	b := bigTestBody[:1024]
	srv := httptest.NewServer(newMiddleware(t, Config{MinSize: 1024})(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write(b[0:1]); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		rw.(http.Flusher).Flush()
		for i := range b[1:] {
			if _, err := rw.Write(b[i+1 : i+2]); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
		}
	})))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
	require.NoError(t, err)

	req.Header.Set(acceptEncoding, "br")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "br", res.Header.Get(contentEncoding))

	got, err := io.ReadAll(brotli.NewReader(res.Body))
	require.NoError(t, err)
	assert.Equal(t, b, got)
}

func Test_FlushAfterWriteNil(t *testing.T) {
	srv := httptest.NewServer(newMiddleware(t, Config{MinSize: 1024})(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)
		if _, err := rw.Write(nil); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}

		rw.(http.Flusher).Flush()
	})))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
	require.NoError(t, err)

	req.Header.Set(acceptEncoding, "br")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "", res.Header.Get(contentEncoding))

	got, err := io.ReadAll(brotli.NewReader(res.Body))
	require.NoError(t, err)
	assert.Len(t, got, 0)
}

func Test_FlushAfterAllWrites(t *testing.T) {
	b := bigTestBody[:1050]
	srv := httptest.NewServer(newMiddleware(t, Config{MinSize: 1024})(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		for i := range b {
			if _, err := rw.Write(b[i : i+1]); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
		}
		rw.(http.Flusher).Flush()
	})))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
	require.NoError(t, err)

	req.Header.Set(acceptEncoding, "br")

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "br", res.Header.Get(contentEncoding))

	got, err := io.ReadAll(brotli.NewReader(res.Body))
	require.NoError(t, err)
	assert.Equal(t, b, got)
}

func Test_ExcludedContentTypes(t *testing.T) {
	testCases := []struct {
		desc                 string
		contentType          string
		excludedContentTypes []string
		expCompression       bool
	}{
		{
			desc:                 "Always compress when content types are empty",
			contentType:          "",
			excludedContentTypes: []string{},
			expCompression:       true,
		},
		{
			desc:                 "MIME match",
			contentType:          "application/json",
			excludedContentTypes: []string{"application/json"},
			expCompression:       false,
		},
		{
			desc:                 "MIME no match",
			contentType:          "text/xml",
			excludedContentTypes: []string{"application/json"},
			expCompression:       true,
		},
		{
			desc:                 "MIME match with no other directive ignores non-MIME directives",
			contentType:          "application/json; charset=utf-8",
			excludedContentTypes: []string{"application/json"},
			expCompression:       false,
		},
		{
			desc:                 "MIME match with other directives requires all directives be equal, different charset",
			contentType:          "application/json; charset=ascii",
			excludedContentTypes: []string{"application/json; charset=utf-8"},
			expCompression:       true,
		},
		{
			desc:                 "MIME match with other directives requires all directives be equal, same charset",
			contentType:          "application/json; charset=utf-8",
			excludedContentTypes: []string{"application/json; charset=utf-8"},
			expCompression:       false,
		},
		{
			desc:                 "MIME match with other directives requires all directives be equal, missing charset",
			contentType:          "application/json",
			excludedContentTypes: []string{"application/json; charset=ascii"},
			expCompression:       true,
		},
		{
			desc:                 "MIME match case insensitive",
			contentType:          "Application/Json",
			excludedContentTypes: []string{"application/json"},
			expCompression:       false,
		},
		{
			desc:                 "MIME match ignore whitespace",
			contentType:          "application/json;charset=utf-8",
			excludedContentTypes: []string{"application/json;            charset=utf-8"},
			expCompression:       false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg := Config{
				MinSize:              1024,
				ExcludedContentTypes: test.excludedContentTypes,
			}
			h := newMiddleware(t, cfg)(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set(contentType, test.contentType)

				rw.WriteHeader(http.StatusOK)
				if _, err := rw.Write(bigTestBody); err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}
			}))

			req, _ := http.NewRequest(http.MethodGet, "/whatever", nil)
			req.Header.Set(acceptEncoding, "br")

			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusOK, rw.Code)

			if test.expCompression {
				assert.Equal(t, "br", rw.Header().Get(contentEncoding))

				got, err := io.ReadAll(brotli.NewReader(rw.Body))
				assert.Nil(t, err)
				assert.Equal(t, bigTestBody, got)
			} else {
				assert.NotEqual(t, "br", rw.Header().Get("Content-Encoding"))

				got, err := io.ReadAll(rw.Body)
				assert.Nil(t, err)
				assert.Equal(t, bigTestBody, got)
			}
		})
	}
}

func Test_FlushExcludedContentTypes(t *testing.T) {
	testCases := []struct {
		desc                 string
		contentType          string
		excludedContentTypes []string
		expCompression       bool
	}{
		{
			desc:                 "Always compress when content types are empty",
			contentType:          "",
			excludedContentTypes: []string{},
			expCompression:       true,
		},
		{
			desc:                 "MIME match",
			contentType:          "application/json",
			excludedContentTypes: []string{"application/json"},
			expCompression:       false,
		},
		{
			desc:                 "MIME no match",
			contentType:          "text/xml",
			excludedContentTypes: []string{"application/json"},
			expCompression:       true,
		},
		{
			desc:                 "MIME match with no other directive ignores non-MIME directives",
			contentType:          "application/json; charset=utf-8",
			excludedContentTypes: []string{"application/json"},
			expCompression:       false,
		},
		{
			desc:                 "MIME match with other directives requires all directives be equal, different charset",
			contentType:          "application/json; charset=ascii",
			excludedContentTypes: []string{"application/json; charset=utf-8"},
			expCompression:       true,
		},
		{
			desc:                 "MIME match with other directives requires all directives be equal, same charset",
			contentType:          "application/json; charset=utf-8",
			excludedContentTypes: []string{"application/json; charset=utf-8"},
			expCompression:       false,
		},
		{
			desc:                 "MIME match with other directives requires all directives be equal, missing charset",
			contentType:          "application/json",
			excludedContentTypes: []string{"application/json; charset=ascii"},
			expCompression:       true,
		},
		{
			desc:                 "MIME match case insensitive",
			contentType:          "Application/Json",
			excludedContentTypes: []string{"application/json"},
			expCompression:       false,
		},
		{
			desc:                 "MIME match ignore whitespace",
			contentType:          "application/json;charset=utf-8",
			excludedContentTypes: []string{"application/json;            charset=utf-8"},
			expCompression:       false,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg := Config{
				MinSize:              1024,
				ExcludedContentTypes: test.excludedContentTypes,
			}
			h := newMiddleware(t, cfg)(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set(contentType, test.contentType)
				rw.WriteHeader(http.StatusOK)

				tb := bigTestBody
				for len(tb) > 0 {
					// Write 100 bytes per run
					// Detection should not be affected (we send 100 bytes)
					toWrite := 100
					if toWrite > len(tb) {
						toWrite = len(tb)
					}

					if _, err := rw.Write(tb[:toWrite]); err != nil {
						http.Error(rw, err.Error(), http.StatusInternalServerError)
					}

					// Flush between each write
					rw.(http.Flusher).Flush()
					tb = tb[toWrite:]
				}
			}))

			req, _ := http.NewRequest(http.MethodGet, "/whatever", nil)
			req.Header.Set(acceptEncoding, "br")

			// This doesn't allow checking flushes, but we validate if content is correct.
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusOK, rw.Code)

			if test.expCompression {
				assert.Equal(t, "br", rw.Header().Get(contentEncoding))

				got, err := io.ReadAll(brotli.NewReader(rw.Body))
				assert.Nil(t, err)
				assert.Equal(t, bigTestBody, got)
			} else {
				assert.NotEqual(t, "br", rw.Header().Get(contentEncoding))

				got, err := io.ReadAll(rw.Body)
				assert.Nil(t, err)
				assert.Equal(t, bigTestBody, got)
			}
		})
	}
}

func newMiddleware(t *testing.T, cfg Config) func(http.Handler) http.HandlerFunc {
	t.Helper()

	m, err := NewMiddleware(cfg)
	require.NoError(t, err)

	return m
}

func newTestHandler(t *testing.T, body []byte) http.Handler {
	t.Helper()

	return newMiddleware(t, Config{MinSize: 1024})(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/compressed" {
				rw.Header().Set("Content-Encoding", "br")
			}
			if _, err := rw.Write(body); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			}
		}),
	)
}
