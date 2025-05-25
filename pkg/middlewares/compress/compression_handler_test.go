package compress

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	smallTestBody = []byte("aaabbc" + strings.Repeat("aaabbbccc", 9) + "aaabbbc")
	bigTestBody   = []byte(strings.Repeat(strings.Repeat("aaabbbccc", 66)+" ", 6) + strings.Repeat("aaabbbccc", 66))
)

func Test_Vary(t *testing.T) {
	testCases := []struct {
		desc           string
		h              http.Handler
		acceptEncoding string
	}{
		{
			desc:           "brotli",
			h:              newTestBrotliHandler(t, smallTestBody),
			acceptEncoding: "br",
		},
		{
			desc:           "zstd",
			h:              newTestZstandardHandler(t, smallTestBody),
			acceptEncoding: "zstd",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequest(http.MethodGet, "/whatever", nil)
			req.Header.Set(acceptEncoding, test.acceptEncoding)

			rw := httptest.NewRecorder()
			test.h.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusAccepted, rw.Code)
			assert.Equal(t, acceptEncoding, rw.Header().Get(vary))
		})
	}
}

func Test_SmallBodyNoCompression(t *testing.T) {
	testCases := []struct {
		desc           string
		h              http.Handler
		acceptEncoding string
	}{
		{
			desc:           "brotli",
			h:              newTestBrotliHandler(t, smallTestBody),
			acceptEncoding: "br",
		},
		{
			desc:           "zstd",
			h:              newTestZstandardHandler(t, smallTestBody),
			acceptEncoding: "zstd",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequest(http.MethodGet, "/whatever", nil)
			req.Header.Set(acceptEncoding, test.acceptEncoding)

			rw := httptest.NewRecorder()
			test.h.ServeHTTP(rw, req)

			// With less than 1024 bytes the response should not be compressed.
			assert.Equal(t, http.StatusAccepted, rw.Code)
			assert.Empty(t, rw.Header().Get(contentEncoding))
			assert.Equal(t, smallTestBody, rw.Body.Bytes())
		})
	}
}

func Test_AlreadyCompressed(t *testing.T) {
	testCases := []struct {
		desc           string
		h              http.Handler
		acceptEncoding string
	}{
		{
			desc:           "brotli",
			h:              newTestBrotliHandler(t, bigTestBody),
			acceptEncoding: "br",
		},
		{
			desc:           "zstd",
			h:              newTestZstandardHandler(t, bigTestBody),
			acceptEncoding: "zstd",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req, _ := http.NewRequest(http.MethodGet, "/compressed", nil)
			req.Header.Set(acceptEncoding, test.acceptEncoding)

			rw := httptest.NewRecorder()
			test.h.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusAccepted, rw.Code)
			assert.Equal(t, bigTestBody, rw.Body.Bytes())
		})
	}
}

func Test_NoBody(t *testing.T) {
	testCases := []struct {
		desc       string
		statusCode int
		body       []byte
	}{
		{
			desc:       "status no content",
			statusCode: http.StatusNoContent,
			body:       nil,
		},
		{
			desc:       "status not modified",
			statusCode: http.StatusNotModified,
			body:       nil,
		},
		{
			desc:       "status OK with empty body",
			statusCode: http.StatusOK,
			body:       []byte{},
		},
		{
			desc:       "status OK with nil body",
			statusCode: http.StatusOK,
			body:       nil,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(test.statusCode)

				_, err := rw.Write(test.body)
				require.NoError(t, err)
			})

			h := mustNewCompressionHandler(t, Config{MinSize: 1024}, zstdName, next)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(acceptEncoding, "zstd")

			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			body, err := io.ReadAll(rw.Body)
			require.NoError(t, err)

			assert.Empty(t, rw.Header().Get(contentEncoding))
			assert.Empty(t, body)
		})
	}
}

func Test_MinSize(t *testing.T) {
	cfg := Config{
		MinSize: 128,
	}

	var bodySize int

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		for range bodySize {
			// We make sure to Write at least once less than minSize so that both
			// cases below go through the same algo: i.e. they start buffering
			// because they haven't reached minSize.
			_, err := rw.Write([]byte{'x'})
			require.NoError(t, err)
		}
	})

	h := mustNewCompressionHandler(t, cfg, zstdName, next)

	req, _ := http.NewRequest(http.MethodGet, "/whatever", &bytes.Buffer{})
	req.Header.Add(acceptEncoding, "zstd")

	// Short response is not compressed
	bodySize = cfg.MinSize - 1
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assert.Empty(t, rw.Result().Header.Get(contentEncoding))

	// Long response is compressed
	bodySize = cfg.MinSize
	rw = httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assert.Equal(t, "zstd", rw.Result().Header.Get(contentEncoding))
}

func Test_MultipleWriteHeader(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// We ensure that the subsequent call to WriteHeader is a noop.
		rw.WriteHeader(http.StatusInternalServerError)
		rw.WriteHeader(http.StatusNotFound)
	})

	h := mustNewCompressionHandler(t, Config{MinSize: 1024}, zstdName, next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(acceptEncoding, "zstd")

	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}

func Test_FlushBeforeWrite(t *testing.T) {
	testCases := []struct {
		desc           string
		cfg            Config
		algo           string
		readerBuilder  func(io.Reader) (io.Reader, error)
		acceptEncoding string
	}{
		{
			desc: "brotli",
			cfg:  Config{MinSize: 1024, MiddlewareName: "Test"},
			algo: brotliName,
			readerBuilder: func(reader io.Reader) (io.Reader, error) {
				return brotli.NewReader(reader), nil
			},
			acceptEncoding: "br",
		},
		{
			desc: "zstd",
			cfg:  Config{MinSize: 1024, MiddlewareName: "Test"},
			algo: zstdName,
			readerBuilder: func(reader io.Reader) (io.Reader, error) {
				return zstd.NewReader(reader)
			},
			acceptEncoding: "zstd",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)
				rw.(http.Flusher).Flush()

				_, err := rw.Write(bigTestBody)
				require.NoError(t, err)
			})

			srv := httptest.NewServer(mustNewCompressionHandler(t, test.cfg, test.algo, next))
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
			require.NoError(t, err)

			req.Header.Set(acceptEncoding, test.acceptEncoding)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer res.Body.Close()

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, test.acceptEncoding, res.Header.Get(contentEncoding))

			reader, err := test.readerBuilder(res.Body)
			require.NoError(t, err)

			got, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, bigTestBody, got)
		})
	}
}

func Test_FlushAfterWrite(t *testing.T) {
	testCases := []struct {
		desc           string
		cfg            Config
		algo           string
		readerBuilder  func(io.Reader) (io.Reader, error)
		acceptEncoding string
	}{
		{
			desc: "brotli",
			cfg:  Config{MinSize: 1024, MiddlewareName: "Test"},
			algo: brotliName,
			readerBuilder: func(reader io.Reader) (io.Reader, error) {
				return brotli.NewReader(reader), nil
			},
			acceptEncoding: "br",
		},
		{
			desc: "zstd",
			cfg:  Config{MinSize: 1024, MiddlewareName: "Test"},
			algo: zstdName,
			readerBuilder: func(reader io.Reader) (io.Reader, error) {
				return zstd.NewReader(reader)
			},
			acceptEncoding: "zstd",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)

				_, err := rw.Write(bigTestBody[0:1])
				require.NoError(t, err)

				rw.(http.Flusher).Flush()
				for _, b := range bigTestBody[1:] {
					_, err := rw.Write([]byte{b})
					require.NoError(t, err)
				}
			})

			srv := httptest.NewServer(mustNewCompressionHandler(t, test.cfg, test.algo, next))
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
			require.NoError(t, err)

			req.Header.Set(acceptEncoding, test.acceptEncoding)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer res.Body.Close()

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, test.acceptEncoding, res.Header.Get(contentEncoding))

			reader, err := test.readerBuilder(res.Body)
			require.NoError(t, err)

			got, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, bigTestBody, got)
		})
	}
}

func Test_FlushAfterWriteNil(t *testing.T) {
	testCases := []struct {
		desc           string
		cfg            Config
		algo           string
		readerBuilder  func(io.Reader) (io.Reader, error)
		acceptEncoding string
	}{
		{
			desc: "brotli",
			cfg:  Config{MinSize: 1024, MiddlewareName: "Test"},
			algo: brotliName,
			readerBuilder: func(reader io.Reader) (io.Reader, error) {
				return brotli.NewReader(reader), nil
			},
			acceptEncoding: "br",
		},
		{
			desc: "zstd",
			cfg:  Config{MinSize: 1024, MiddlewareName: "Test"},
			algo: zstdName,
			readerBuilder: func(reader io.Reader) (io.Reader, error) {
				return zstd.NewReader(reader)
			},
			acceptEncoding: "zstd",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)

				_, err := rw.Write(nil)
				require.NoError(t, err)

				rw.(http.Flusher).Flush()
			})

			srv := httptest.NewServer(mustNewCompressionHandler(t, test.cfg, test.algo, next))
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
			require.NoError(t, err)

			req.Header.Set(acceptEncoding, test.acceptEncoding)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer res.Body.Close()

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Empty(t, res.Header.Get(contentEncoding))

			reader, err := test.readerBuilder(res.Body)
			require.NoError(t, err)

			got, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Empty(t, got)
		})
	}
}

func Test_FlushAfterAllWrites(t *testing.T) {
	testCases := []struct {
		desc           string
		cfg            Config
		algo           string
		readerBuilder  func(io.Reader) (io.Reader, error)
		acceptEncoding string
	}{
		{
			desc: "brotli",
			cfg:  Config{MinSize: 1024, MiddlewareName: "Test"},
			algo: brotliName,
			readerBuilder: func(reader io.Reader) (io.Reader, error) {
				return brotli.NewReader(reader), nil
			},
			acceptEncoding: "br",
		},
		{
			desc: "zstd",
			cfg:  Config{MinSize: 1024, MiddlewareName: "Test"},
			algo: zstdName,
			readerBuilder: func(reader io.Reader) (io.Reader, error) {
				return zstd.NewReader(reader)
			},
			acceptEncoding: "zstd",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				for i := range bigTestBody {
					_, err := rw.Write(bigTestBody[i : i+1])
					require.NoError(t, err)
				}
				rw.(http.Flusher).Flush()
			})

			srv := httptest.NewServer(mustNewCompressionHandler(t, test.cfg, test.algo, next))
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
			require.NoError(t, err)

			req.Header.Set(acceptEncoding, test.acceptEncoding)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer res.Body.Close()

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, test.acceptEncoding, res.Header.Get(contentEncoding))

			reader, err := test.readerBuilder(res.Body)
			require.NoError(t, err)

			got, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, bigTestBody, got)
		})
	}
}

func Test_FlushForceCompress(t *testing.T) {
	testCases := []struct {
		desc           string
		cfg            Config
		algo           string
		readerBuilder  func(io.Reader) (io.Reader, error)
		acceptEncoding string
	}{
		{
			desc: "brotli",
			cfg:  Config{MinSize: 1024, MiddlewareName: "Test"},
			algo: brotliName,
			readerBuilder: func(reader io.Reader) (io.Reader, error) {
				return brotli.NewReader(reader), nil
			},
			acceptEncoding: "br",
		},
		{
			desc: "zstd",
			cfg:  Config{MinSize: 1024, MiddlewareName: "Test"},
			algo: zstdName,
			readerBuilder: func(reader io.Reader) (io.Reader, error) {
				return zstd.NewReader(reader)
			},
			acceptEncoding: "zstd",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.WriteHeader(http.StatusOK)

				_, err := rw.Write(smallTestBody)
				require.NoError(t, err)

				rw.(http.Flusher).Flush()
			})

			srv := httptest.NewServer(mustNewCompressionHandler(t, test.cfg, test.algo, next))
			defer srv.Close()

			req, err := http.NewRequest(http.MethodGet, srv.URL, http.NoBody)
			require.NoError(t, err)

			req.Header.Set(acceptEncoding, test.acceptEncoding)

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			defer res.Body.Close()

			assert.Equal(t, http.StatusOK, res.StatusCode)
			assert.Equal(t, test.acceptEncoding, res.Header.Get(contentEncoding))

			reader, err := test.readerBuilder(res.Body)
			require.NoError(t, err)

			got, err := io.ReadAll(reader)
			require.NoError(t, err)
			assert.Equal(t, smallTestBody, got)
		})
	}
}

func Test_ExcludedContentTypes(t *testing.T) {
	testCases := []struct {
		desc                 string
		contentType          string
		excludedContentTypes []string
		expCompression       bool
	}{
		{
			desc:           "Always compress when content types are empty",
			contentType:    "",
			expCompression: true,
		},
		{
			desc:           "MIME malformed",
			contentType:    "application/json;charset=UTF-8;charset=utf-8",
			expCompression: false,
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
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg := Config{
				MinSize:              1024,
				ExcludedContentTypes: test.excludedContentTypes,
			}

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set(contentType, test.contentType)

				rw.WriteHeader(http.StatusAccepted)

				_, err := rw.Write(bigTestBody)
				require.NoError(t, err)
			})

			h := mustNewCompressionHandler(t, cfg, zstdName, next)

			req, _ := http.NewRequest(http.MethodGet, "/whatever", nil)
			req.Header.Set(acceptEncoding, zstdName)

			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusAccepted, rw.Code)

			if test.expCompression {
				assert.Equal(t, zstdName, rw.Header().Get(contentEncoding))

				reader, err := zstd.NewReader(rw.Body)
				require.NoError(t, err)

				got, err := io.ReadAll(reader)
				assert.NoError(t, err)
				assert.Equal(t, bigTestBody, got)
			} else {
				assert.NotEqual(t, zstdName, rw.Header().Get("Content-Encoding"))

				got, err := io.ReadAll(rw.Body)
				assert.NoError(t, err)
				assert.Equal(t, bigTestBody, got)
			}
		})
	}
}

func Test_IncludedContentTypes(t *testing.T) {
	testCases := []struct {
		desc                 string
		contentType          string
		includedContentTypes []string
		expCompression       bool
	}{
		{
			desc:           "Always compress when content types are empty",
			contentType:    "",
			expCompression: true,
		},
		{
			desc:           "MIME malformed",
			contentType:    "application/json;charset=UTF-8;charset=utf-8",
			expCompression: false,
		},
		{
			desc:                 "MIME match",
			contentType:          "application/json",
			includedContentTypes: []string{"application/json"},
			expCompression:       true,
		},
		{
			desc:                 "MIME no match",
			contentType:          "text/xml",
			includedContentTypes: []string{"application/json"},
			expCompression:       false,
		},
		{
			desc:                 "MIME match with no other directive ignores non-MIME directives",
			contentType:          "application/json; charset=utf-8",
			includedContentTypes: []string{"application/json"},
			expCompression:       true,
		},
		{
			desc:                 "MIME match with other directives requires all directives be equal, different charset",
			contentType:          "application/json; charset=ascii",
			includedContentTypes: []string{"application/json; charset=utf-8"},
			expCompression:       false,
		},
		{
			desc:                 "MIME match with other directives requires all directives be equal, same charset",
			contentType:          "application/json; charset=utf-8",
			includedContentTypes: []string{"application/json; charset=utf-8"},
			expCompression:       true,
		},
		{
			desc:                 "MIME match with other directives requires all directives be equal, missing charset",
			contentType:          "application/json",
			includedContentTypes: []string{"application/json; charset=ascii"},
			expCompression:       false,
		},
		{
			desc:                 "MIME match case insensitive",
			contentType:          "Application/Json",
			includedContentTypes: []string{"application/json"},
			expCompression:       true,
		},
		{
			desc:                 "MIME match ignore whitespace",
			contentType:          "application/json;charset=utf-8",
			includedContentTypes: []string{"application/json;            charset=utf-8"},
			expCompression:       true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg := Config{
				MinSize:              1024,
				IncludedContentTypes: test.includedContentTypes,
			}

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set(contentType, test.contentType)

				rw.WriteHeader(http.StatusAccepted)

				_, err := rw.Write(bigTestBody)
				require.NoError(t, err)
			})

			h := mustNewCompressionHandler(t, cfg, zstdName, next)

			req, _ := http.NewRequest(http.MethodGet, "/whatever", nil)
			req.Header.Set(acceptEncoding, zstdName)

			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusAccepted, rw.Code)

			if test.expCompression {
				assert.Equal(t, zstdName, rw.Header().Get(contentEncoding))

				reader, err := zstd.NewReader(rw.Body)
				require.NoError(t, err)

				got, err := io.ReadAll(reader)
				assert.NoError(t, err)
				assert.Equal(t, bigTestBody, got)
			} else {
				assert.NotEqual(t, zstdName, rw.Header().Get("Content-Encoding"))

				got, err := io.ReadAll(rw.Body)
				assert.NoError(t, err)
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
			desc:           "Always compress when content types are empty",
			contentType:    "",
			expCompression: true,
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
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg := Config{
				MinSize:              1024,
				ExcludedContentTypes: test.excludedContentTypes,
			}

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
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

					_, err := rw.Write(tb[:toWrite])
					require.NoError(t, err)

					// Flush between each write
					rw.(http.Flusher).Flush()
					tb = tb[toWrite:]
				}
			})

			h := mustNewCompressionHandler(t, cfg, zstdName, next)

			req, _ := http.NewRequest(http.MethodGet, "/whatever", nil)
			req.Header.Set(acceptEncoding, zstdName)

			// This doesn't allow checking flushes, but we validate if content is correct.
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusOK, rw.Code)

			if test.expCompression {
				assert.Equal(t, zstdName, rw.Header().Get(contentEncoding))

				reader, err := zstd.NewReader(rw.Body)
				require.NoError(t, err)

				got, err := io.ReadAll(reader)
				assert.NoError(t, err)
				assert.Equal(t, bigTestBody, got)
			} else {
				assert.NotEqual(t, zstdName, rw.Header().Get(contentEncoding))

				got, err := io.ReadAll(rw.Body)
				assert.NoError(t, err)
				assert.Equal(t, bigTestBody, got)
			}
		})
	}
}

func Test_FlushIncludedContentTypes(t *testing.T) {
	testCases := []struct {
		desc                 string
		contentType          string
		includedContentTypes []string
		expCompression       bool
	}{
		{
			desc:           "Always compress when content types are empty",
			contentType:    "",
			expCompression: true,
		},
		{
			desc:                 "MIME match",
			contentType:          "application/json",
			includedContentTypes: []string{"application/json"},
			expCompression:       true,
		},
		{
			desc:                 "MIME no match",
			contentType:          "text/xml",
			includedContentTypes: []string{"application/json"},
			expCompression:       false,
		},
		{
			desc:                 "MIME match with no other directive ignores non-MIME directives",
			contentType:          "application/json; charset=utf-8",
			includedContentTypes: []string{"application/json"},
			expCompression:       true,
		},
		{
			desc:                 "MIME match with other directives requires all directives be equal, different charset",
			contentType:          "application/json; charset=ascii",
			includedContentTypes: []string{"application/json; charset=utf-8"},
			expCompression:       false,
		},
		{
			desc:                 "MIME match with other directives requires all directives be equal, same charset",
			contentType:          "application/json; charset=utf-8",
			includedContentTypes: []string{"application/json; charset=utf-8"},
			expCompression:       true,
		},
		{
			desc:                 "MIME match with other directives requires all directives be equal, missing charset",
			contentType:          "application/json",
			includedContentTypes: []string{"application/json; charset=ascii"},
			expCompression:       false,
		},
		{
			desc:                 "MIME match case insensitive",
			contentType:          "Application/Json",
			includedContentTypes: []string{"application/json"},
			expCompression:       true,
		},
		{
			desc:                 "MIME match ignore whitespace",
			contentType:          "application/json;charset=utf-8",
			includedContentTypes: []string{"application/json;            charset=utf-8"},
			expCompression:       true,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cfg := Config{
				MinSize:              1024,
				IncludedContentTypes: test.includedContentTypes,
			}

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
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

					_, err := rw.Write(tb[:toWrite])
					require.NoError(t, err)

					// Flush between each write
					rw.(http.Flusher).Flush()
					tb = tb[toWrite:]
				}
			})

			h := mustNewCompressionHandler(t, cfg, zstdName, next)

			req, _ := http.NewRequest(http.MethodGet, "/whatever", nil)
			req.Header.Set(acceptEncoding, zstdName)

			// This doesn't allow checking flushes, but we validate if content is correct.
			rw := httptest.NewRecorder()
			h.ServeHTTP(rw, req)

			assert.Equal(t, http.StatusOK, rw.Code)

			if test.expCompression {
				assert.Equal(t, zstdName, rw.Header().Get(contentEncoding))

				reader, err := zstd.NewReader(rw.Body)
				require.NoError(t, err)

				got, err := io.ReadAll(reader)
				assert.NoError(t, err)
				assert.Equal(t, bigTestBody, got)
			} else {
				assert.NotEqual(t, zstdName, rw.Header().Get(contentEncoding))

				got, err := io.ReadAll(rw.Body)
				assert.NoError(t, err)
				assert.Equal(t, bigTestBody, got)
			}
		})
	}
}

func mustNewCompressionHandler(t *testing.T, cfg Config, algo string, next http.Handler) http.Handler {
	t.Helper()

	var writer NewCompressionWriter
	switch algo {
	case zstdName:
		writer = func(rw http.ResponseWriter) (CompressionWriter, string, error) {
			writer, err := zstd.NewWriter(rw)
			require.NoError(t, err)
			return writer, zstdName, nil
		}
	case brotliName:
		writer = func(rw http.ResponseWriter) (CompressionWriter, string, error) {
			return brotli.NewWriter(rw), brotliName, nil
		}
	default:
		assert.Failf(t, "unknown compression algorithm: %s", algo)
	}

	w, err := NewCompressionHandler(cfg, writer, next)
	require.NoError(t, err)

	return w
}

func newTestBrotliHandler(t *testing.T, body []byte) http.Handler {
	t.Helper()

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/compressed" {
			rw.Header().Set("Content-Encoding", brotliName)
		}

		rw.WriteHeader(http.StatusAccepted)
		_, err := rw.Write(body)
		require.NoError(t, err)
	})

	return mustNewCompressionHandler(t, Config{MinSize: 1024, MiddlewareName: "Compress"}, brotliName, next)
}

func newTestZstandardHandler(t *testing.T, body []byte) http.Handler {
	t.Helper()

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/compressed" {
			rw.Header().Set("Content-Encoding", zstdName)
		}

		rw.WriteHeader(http.StatusAccepted)
		_, err := rw.Write(body)
		require.NoError(t, err)
	})

	return mustNewCompressionHandler(t, Config{MinSize: 1024, MiddlewareName: "Compress"}, zstdName, next)
}

func Test_ParseContentType_equals(t *testing.T) {
	testCases := []struct {
		desc      string
		pct       parsedContentType
		mediaType string
		params    map[string]string
		expect    assert.BoolAssertionFunc
	}{
		{
			desc:   "empty parsed content type",
			expect: assert.True,
		},
		{
			desc: "simple content type",
			pct: parsedContentType{
				mediaType: "plain/text",
			},
			mediaType: "plain/text",
			expect:    assert.True,
		},
		{
			desc: "content type with params",
			pct: parsedContentType{
				mediaType: "plain/text",
				params: map[string]string{
					"charset": "utf8",
				},
			},
			mediaType: "plain/text",
			params: map[string]string{
				"charset": "utf8",
			},
			expect: assert.True,
		},
		{
			desc: "different content type",
			pct: parsedContentType{
				mediaType: "plain/text",
			},
			mediaType: "application/json",
			expect:    assert.False,
		},
		{
			desc: "content type with params",
			pct: parsedContentType{
				mediaType: "plain/text",
				params: map[string]string{
					"charset": "utf8",
				},
			},
			mediaType: "plain/text",
			params: map[string]string{
				"charset": "latin-1",
			},
			expect: assert.False,
		},
		{
			desc: "different number of parameters",
			pct: parsedContentType{
				mediaType: "plain/text",
				params: map[string]string{
					"charset": "utf8",
				},
			},
			mediaType: "plain/text",
			params: map[string]string{
				"charset": "utf8",
				"q":       "0.8",
			},
			expect: assert.False,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			test.expect(t, test.pct.equals(test.mediaType, test.params))
		})
	}
}
