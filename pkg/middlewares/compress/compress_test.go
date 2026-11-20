package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/testhelpers"
)

const (
	contentEncodingHeader = "Content-Encoding"
	contentTypeHeader     = "Content-Type"
	varyHeader            = "Vary"
)

func TestNegotiation(t *testing.T) {
	testCases := []struct {
		desc            string
		acceptEncHeader string
		expEncoding     string
	}{
		{
			desc:        "no accept header",
			expEncoding: "",
		},
		{
			desc:            "unsupported accept header",
			acceptEncHeader: "notreal",
			expEncoding:     "",
		},
		{
			desc:            "accept any header",
			acceptEncHeader: "*",
			expEncoding:     gzipName,
		},
		{
			desc:            "gzip accept header",
			acceptEncHeader: "gzip",
			expEncoding:     gzipName,
		},
		{
			desc:            "br accept header",
			acceptEncHeader: "br",
			expEncoding:     brotliName,
		},
		{
			desc:            "multi accept header, prefer br",
			acceptEncHeader: "br;q=0.8, gzip;q=0.6",
			expEncoding:     brotliName,
		},
		{
			desc:            "multi accept header, prefer gzip",
			acceptEncHeader: "gzip;q=1.0, br;q=0.8",
			expEncoding:     gzipName,
		},
		{
			desc:            "multi accept header list, prefer br",
			acceptEncHeader: "gzip, br",
			expEncoding:     gzipName,
		},
		{
			desc:            "zstd accept header",
			acceptEncHeader: "zstd",
			expEncoding:     zstdName,
		},
		{
			desc:            "multi accept header, prefer zstd",
			acceptEncHeader: "zstd;q=0.9, br;q=0.8, gzip;q=0.6",
			expEncoding:     zstdName,
		},
		{
			desc:            "multi accept header, prefer brotli",
			acceptEncHeader: "gzip;q=0.8, br;q=1.0, zstd;q=0.7",
			expEncoding:     brotliName,
		},
		{
			desc:            "multi accept header, prefer gzip",
			acceptEncHeader: "gzip;q=1.0, br;q=0.8, zstd;q=0.7",
			expEncoding:     gzipName,
		},
		{
			desc:            "multi accept header list, prefer best",
			acceptEncHeader: "gzip, br, zstd",
			expEncoding:     "gzip",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
			if test.acceptEncHeader != "" {
				req.Header.Add(acceptEncodingHeader, test.acceptEncHeader)
			}

			next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				_, _ = rw.Write(generateBytes(10))
			})
			cfg := dynamic.Compress{
				MinResponseBodyBytes: 1,
				Encodings:            defaultSupportedEncodings,
			}
			handler, err := New(t.Context(), next, cfg, "testing")
			require.NoError(t, err)

			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, req)

			assert.Equal(t, test.expEncoding, rw.Header().Get(contentEncodingHeader))
		})
	}
}

func TestShouldCompressWhenNoContentEncodingHeader(t *testing.T) {
	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Add(acceptEncodingHeader, gzipName)

	baseBody := generateBytes(defaultMinSize)

	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write(baseBody)
		assert.NoError(t, err)
	})
	handler, err := New(t.Context(), next, dynamic.Compress{Encodings: defaultSupportedEncodings}, "testing")
	require.NoError(t, err)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, gzipName, rw.Header().Get(contentEncodingHeader))
	assert.Equal(t, acceptEncodingHeader, rw.Header().Get(varyHeader))

	gr, err := gzip.NewReader(rw.Body)
	require.NoError(t, err)

	got, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, got, baseBody)
}

func TestShouldNotCompressWhenContentEncodingHeader(t *testing.T) {
	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Add(acceptEncodingHeader, gzipName)

	fakeCompressedBody := generateBytes(defaultMinSize)
	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add(contentEncodingHeader, gzipName)
		rw.Header().Add(varyHeader, acceptEncodingHeader)
		_, err := rw.Write(fakeCompressedBody)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
	handler, err := New(t.Context(), next, dynamic.Compress{Encodings: defaultSupportedEncodings}, "testing")
	require.NoError(t, err)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, gzipName, rw.Header().Get(contentEncodingHeader))
	assert.Equal(t, acceptEncodingHeader, rw.Header().Get(varyHeader))

	assert.Equal(t, rw.Body.Bytes(), fakeCompressedBody)
}

func TestShouldNotCompressWhenNoAcceptEncodingHeader(t *testing.T) {
	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)

	fakeBody := generateBytes(defaultMinSize)
	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write(fakeBody)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
	handler, err := New(t.Context(), next, dynamic.Compress{Encodings: defaultSupportedEncodings}, "testing")
	require.NoError(t, err)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Empty(t, rw.Header().Get(contentEncodingHeader))
	assert.Empty(t, rw.Header().Get(varyHeader))
	assert.Equal(t, rw.Body.Bytes(), fakeBody)
}

func TestEmptyAcceptEncoding(t *testing.T) {
	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Add(acceptEncodingHeader, "")

	fakeBody := generateBytes(defaultMinSize)
	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write(fakeBody)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
	handler, err := New(t.Context(), next, dynamic.Compress{Encodings: defaultSupportedEncodings}, "testing")
	require.NoError(t, err)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Empty(t, rw.Header().Get(contentEncodingHeader))
	assert.Empty(t, rw.Header().Get(varyHeader))
	assert.Equal(t, rw.Body.Bytes(), fakeBody)
}

func TestShouldNotCompressWhenIdentityAcceptEncodingHeader(t *testing.T) {
	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Set(acceptEncodingHeader, "identity")

	fakeBody := generateBytes(defaultMinSize)
	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get(acceptEncodingHeader) != "identity" {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		_, err := rw.Write(fakeBody)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
	handler, err := New(t.Context(), next, dynamic.Compress{Encodings: defaultSupportedEncodings}, "testing")
	require.NoError(t, err)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Empty(t, rw.Header().Get(contentEncodingHeader))
	assert.Empty(t, rw.Header().Get(varyHeader))
	assert.Equal(t, rw.Body.Bytes(), fakeBody)
}

func TestShouldNotCompressWhenEmptyAcceptEncodingHeader(t *testing.T) {
	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Set(acceptEncodingHeader, "")

	fakeBody := generateBytes(defaultMinSize)
	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Header.Get(acceptEncodingHeader) != "" {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		_, err := rw.Write(fakeBody)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
	handler, err := New(t.Context(), next, dynamic.Compress{Encodings: defaultSupportedEncodings}, "testing")
	require.NoError(t, err)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Empty(t, rw.Header().Get(contentEncodingHeader))
	assert.Empty(t, rw.Header().Get(varyHeader))
	assert.Equal(t, rw.Body.Bytes(), fakeBody)
}

func TestShouldNotCompressHeadRequest(t *testing.T) {
	req := testhelpers.MustNewRequest(http.MethodHead, "http://localhost", nil)
	req.Header.Add(acceptEncodingHeader, gzipName)

	fakeBody := generateBytes(defaultMinSize)
	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_, err := rw.Write(fakeBody)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
	handler, err := New(t.Context(), next, dynamic.Compress{Encodings: defaultSupportedEncodings}, "testing")
	require.NoError(t, err)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Empty(t, rw.Header().Get(contentEncodingHeader))
	assert.Empty(t, rw.Header().Get(varyHeader))
	assert.Equal(t, rw.Body.Bytes(), fakeBody)
}

func TestShouldNotCompressWhenSpecificContentType(t *testing.T) {
	baseBody := generateBytes(defaultMinSize)

	testCases := []struct {
		desc            string
		conf            dynamic.Compress
		reqContentType  string
		respContentType string
	}{
		{
			desc: "Exclude Request Content-Type",
			conf: dynamic.Compress{
				Encodings:            defaultSupportedEncodings,
				ExcludedContentTypes: []string{"text/event-stream"},
			},
			reqContentType: "text/event-stream",
		},
		{
			desc: "Exclude Response Content-Type",
			conf: dynamic.Compress{
				Encodings:            defaultSupportedEncodings,
				ExcludedContentTypes: []string{"text/event-stream"},
			},
			respContentType: "text/event-stream",
		},
		{
			desc: "Include Response Content-Type",
			conf: dynamic.Compress{
				Encodings:            defaultSupportedEncodings,
				IncludedContentTypes: []string{"text/plain"},
			},
			respContentType: "text/html",
		},
		{
			desc: "Ignoring application/grpc with exclude option",
			conf: dynamic.Compress{
				Encodings:            defaultSupportedEncodings,
				ExcludedContentTypes: []string{"application/json"},
			},
			reqContentType: "application/grpc",
		},
		{
			desc: "Ignoring application/grpc with include option",
			conf: dynamic.Compress{
				Encodings:            defaultSupportedEncodings,
				IncludedContentTypes: []string{"application/json"},
			},
			reqContentType: "application/grpc",
		},
		{
			desc: "Ignoring application/grpc with no option",
			conf: dynamic.Compress{
				Encodings: defaultSupportedEncodings,
			},
			reqContentType: "application/grpc",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
			req.Header.Add(acceptEncodingHeader, gzipName)
			if test.reqContentType != "" {
				req.Header.Add(contentTypeHeader, test.reqContentType)
			}

			next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				if len(test.respContentType) > 0 {
					rw.Header().Set(contentTypeHeader, test.respContentType)
				}

				_, err := rw.Write(baseBody)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}
			})

			handler, err := New(t.Context(), next, test.conf, "test")
			require.NoError(t, err)

			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, req)

			assert.Empty(t, rw.Header().Get(acceptEncodingHeader))
			assert.Empty(t, rw.Header().Get(contentEncodingHeader))
			assert.Equal(t, rw.Body.Bytes(), baseBody)
		})
	}
}

func TestShouldCompressWhenSpecificContentType(t *testing.T) {
	baseBody := generateBytes(defaultMinSize)

	testCases := []struct {
		desc            string
		conf            dynamic.Compress
		respContentType string
	}{
		{
			desc: "Include Response Content-Type",
			conf: dynamic.Compress{
				Encodings:            defaultSupportedEncodings,
				IncludedContentTypes: []string{"text/html"},
			},
			respContentType: "text/html",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
			req.Header.Add(acceptEncodingHeader, gzipName)

			next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Set(contentTypeHeader, test.respContentType)

				if _, err := rw.Write(baseBody); err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}
			})

			handler, err := New(t.Context(), next, test.conf, "test")
			require.NoError(t, err)

			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, req)

			assert.Equal(t, gzipName, rw.Header().Get(contentEncodingHeader))
			assert.Equal(t, acceptEncodingHeader, rw.Header().Get(varyHeader))
			assert.NotEqual(t, rw.Body.Bytes(), baseBody)
		})
	}
}

func TestIntegrationShouldNotCompress(t *testing.T) {
	fakeCompressedBody := generateBytes(100000)

	testCases := []struct {
		name               string
		handler            http.Handler
		expectedStatusCode int
	}{
		{
			name: "when content already compressed",
			handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Add(contentEncodingHeader, gzipName)
				rw.Header().Add(varyHeader, acceptEncodingHeader)
				_, err := rw.Write(fakeCompressedBody)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}
			}),
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "when content already compressed and status code Created",
			handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Add(contentEncodingHeader, gzipName)
				rw.Header().Add(varyHeader, acceptEncodingHeader)
				rw.WriteHeader(http.StatusCreated)
				_, err := rw.Write(fakeCompressedBody)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}
			}),
			expectedStatusCode: http.StatusCreated,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			compress, err := New(t.Context(), test.handler, dynamic.Compress{Encodings: defaultSupportedEncodings}, "testing")
			require.NoError(t, err)

			ts := httptest.NewServer(compress)
			defer ts.Close()

			req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
			req.Header.Add(acceptEncodingHeader, gzipName)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatusCode, resp.StatusCode)

			assert.Equal(t, gzipName, resp.Header.Get(contentEncodingHeader))
			assert.Equal(t, acceptEncodingHeader, resp.Header.Get(varyHeader))

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, fakeCompressedBody, body)
		})
	}
}

func TestShouldWriteHeaderWhenFlush(t *testing.T) {
	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add(contentEncodingHeader, gzipName)
		rw.Header().Add(varyHeader, acceptEncodingHeader)
		rw.WriteHeader(http.StatusUnauthorized)
		assert.NoError(t, http.NewResponseController(rw).Flush())
		_, err := rw.Write([]byte("short"))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
	handler, err := New(t.Context(), next, dynamic.Compress{Encodings: defaultSupportedEncodings}, "testing")
	require.NoError(t, err)

	ts := httptest.NewServer(handler)
	defer ts.Close()

	req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Add(acceptEncodingHeader, gzipName)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	assert.Equal(t, gzipName, resp.Header.Get(contentEncodingHeader))
	assert.Equal(t, acceptEncodingHeader, resp.Header.Get(varyHeader))
}

func TestFlushBeforeWriteSendsHeaders(t *testing.T) {
	testCases := []struct {
		desc             string
		encoding         string
		contentType      string
		expectedEncoding string
	}{
		{
			desc:             "gzip, compressible content type",
			encoding:         gzipName,
			contentType:      "application/json",
			expectedEncoding: gzipName,
		},
		{
			desc:        "gzip, excluded content type",
			encoding:    gzipName,
			contentType: "text/event-stream",
		},
		{
			desc:             "brotli, compressible content type",
			encoding:         brotliName,
			contentType:      "application/json",
			expectedEncoding: brotliName,
		},
		{
			desc:        "brotli, excluded content type",
			encoding:    brotliName,
			contentType: "text/event-stream",
		},
		{
			desc:             "zstd, compressible content type",
			encoding:         zstdName,
			contentType:      "application/json",
			expectedEncoding: zstdName,
		},
		{
			desc:        "zstd, excluded content type",
			encoding:    zstdName,
			contentType: "text/event-stream",
		},
	}

	body := []byte(`{"type":"ADDED"}`)

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			// The handler only writes once the client got the headers,
			// like a Kubernetes watch that has no event to send yet.
			headersReceived := make(chan struct{})

			next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				rw.Header().Set(contentTypeHeader, test.contentType)
				rw.WriteHeader(http.StatusOK)
				assert.NoError(t, http.NewResponseController(rw).Flush())

				select {
				case <-headersReceived:
				case <-req.Context().Done():
					return
				}

				_, err := rw.Write(body)
				assert.NoError(t, err)
			})

			cfg := dynamic.Compress{
				Encodings:            defaultSupportedEncodings,
				ExcludedContentTypes: []string{"text/event-stream"},
			}
			handler, err := New(t.Context(), next, cfg, "testing")
			require.NoError(t, err)

			server := httptest.NewServer(handler)
			t.Cleanup(server.Close)

			req := testhelpers.MustNewRequest(http.MethodGet, server.URL, nil)
			req.Header.Set(acceptEncodingHeader, test.encoding)

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)

			defer resp.Body.Close()

			close(headersReceived)

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, test.expectedEncoding, resp.Header.Get(contentEncodingHeader))

			got, err := io.ReadAll(newDecoder(t, test.expectedEncoding, resp.Body))
			require.NoError(t, err)
			assert.Equal(t, body, got)
		})
	}
}

func TestEmptyBody(t *testing.T) {
	testCases := []struct {
		desc        string
		handler     func(t *testing.T, rw http.ResponseWriter)
		compressed  bool
		statusCode  int
		contentType string
	}{
		{
			desc:       "nothing written",
			handler:    func(*testing.T, http.ResponseWriter) {},
			statusCode: http.StatusOK,
		},
		{
			desc: "nil write",
			handler: func(t *testing.T, rw http.ResponseWriter) {
				t.Helper()

				_, err := rw.Write(nil)
				assert.NoError(t, err)
			},
			statusCode: http.StatusOK,
		},
		{
			desc: "empty write",
			handler: func(t *testing.T, rw http.ResponseWriter) {
				t.Helper()

				_, err := rw.Write([]byte{})
				assert.NoError(t, err)
			},
			statusCode: http.StatusOK,
		},
		{
			desc: "flush",
			handler: func(t *testing.T, rw http.ResponseWriter) {
				t.Helper()

				assert.NoError(t, http.NewResponseController(rw).Flush())
			},
			compressed: true,
			statusCode: http.StatusOK,
		},
		{
			desc: "flush twice",
			handler: func(t *testing.T, rw http.ResponseWriter) {
				t.Helper()

				assert.NoError(t, http.NewResponseController(rw).Flush())
				assert.NoError(t, http.NewResponseController(rw).Flush())
			},
			compressed: true,
			statusCode: http.StatusOK,
		},
		{
			desc: "nil write then flush",
			handler: func(t *testing.T, rw http.ResponseWriter) {
				t.Helper()

				_, err := rw.Write(nil)
				assert.NoError(t, err)
				assert.NoError(t, http.NewResponseController(rw).Flush())
			},
			compressed: true,
			statusCode: http.StatusOK,
		},
		{
			desc: "flush then nil write",
			handler: func(t *testing.T, rw http.ResponseWriter) {
				t.Helper()

				assert.NoError(t, http.NewResponseController(rw).Flush())
				_, err := rw.Write(nil)
				assert.NoError(t, err)
			},
			compressed: true,
			statusCode: http.StatusOK,
		},
		{
			desc: "flush with excluded content type",
			handler: func(t *testing.T, rw http.ResponseWriter) {
				t.Helper()

				rw.Header().Set(contentTypeHeader, "text/event-stream")
				assert.NoError(t, http.NewResponseController(rw).Flush())
			},
			statusCode:  http.StatusOK,
			contentType: "text/event-stream",
		},
		{
			desc: "flush with Content-Length below the minimum size",
			handler: func(t *testing.T, rw http.ResponseWriter) {
				t.Helper()

				rw.Header().Set("Content-Length", "0")
				assert.NoError(t, http.NewResponseController(rw).Flush())
			},
			statusCode: http.StatusOK,
		},
		{
			desc: "no content then flush",
			handler: func(t *testing.T, rw http.ResponseWriter) {
				t.Helper()

				rw.WriteHeader(http.StatusNoContent)
				assert.NoError(t, http.NewResponseController(rw).Flush())
			},
			statusCode: http.StatusNoContent,
		},
		{
			desc: "not modified then flush",
			handler: func(t *testing.T, rw http.ResponseWriter) {
				t.Helper()

				rw.WriteHeader(http.StatusNotModified)
				assert.NoError(t, http.NewResponseController(rw).Flush())
			},
			statusCode: http.StatusNotModified,
		},
		{
			desc: "no content with a body then flush",
			handler: func(t *testing.T, rw http.ResponseWriter) {
				t.Helper()

				rw.WriteHeader(http.StatusNoContent)
				// The server rejects this body, what matters is that it is not announced as compressed.
				_, _ = rw.Write([]byte("body"))
				assert.NoError(t, http.NewResponseController(rw).Flush())
			},
			statusCode: http.StatusNoContent,
		},
		{
			desc: "not modified with a body then flush",
			handler: func(t *testing.T, rw http.ResponseWriter) {
				t.Helper()

				rw.WriteHeader(http.StatusNotModified)
				_, _ = rw.Write([]byte("body"))
				assert.NoError(t, http.NewResponseController(rw).Flush())
			},
			statusCode: http.StatusNotModified,
		},
	}

	for _, encoding := range []string{gzipName, brotliName, zstdName} {
		for _, test := range testCases {
			t.Run(encoding+", "+test.desc, func(t *testing.T) {
				t.Parallel()

				next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
					test.handler(t, rw)
				})

				cfg := dynamic.Compress{
					Encodings:            defaultSupportedEncodings,
					ExcludedContentTypes: []string{"text/event-stream"},
				}
				handler, err := New(t.Context(), next, cfg, "testing")
				require.NoError(t, err)

				server := httptest.NewServer(handler)
				t.Cleanup(server.Close)

				req := testhelpers.MustNewRequest(http.MethodGet, server.URL, nil)
				req.Header.Set(acceptEncodingHeader, encoding)

				resp, err := http.DefaultClient.Do(req)
				require.NoError(t, err)

				defer resp.Body.Close()

				assert.Equal(t, test.statusCode, resp.StatusCode)
				assert.Equal(t, test.contentType, resp.Header.Get(contentTypeHeader))

				expectedEncoding := ""
				if test.compressed {
					expectedEncoding = encoding
				}
				assert.Equal(t, expectedEncoding, resp.Header.Get(contentEncodingHeader))

				// Reading the whole body checks that a compressed stream is complete and valid.
				got, err := io.ReadAll(newDecoder(t, expectedEncoding, resp.Body))
				require.NoError(t, err)
				assert.Empty(t, got)
			})
		}
	}
}

func TestCompressionWriterReuse(t *testing.T) {
	bigBody := generateBytes(10 * defaultMinSize)
	smallBody := generateBytes(10)

	// Each response leaves its compression writer in a different state before it goes back to the pool.
	responses := []struct {
		body       []byte
		flush      bool
		compressed bool
	}{
		{body: bigBody, compressed: true},
		{flush: true, compressed: true},
		{body: smallBody},
		{body: smallBody, flush: true, compressed: true},
		{body: bigBody, flush: true, compressed: true},
		{body: bigBody, compressed: true},
	}

	for _, encoding := range []string{gzipName, brotliName, zstdName} {
		t.Run(encoding, func(t *testing.T) {
			t.Parallel()

			var current int
			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				response := responses[current]

				_, err := rw.Write(response.body)
				assert.NoError(t, err)

				if response.flush {
					assert.NoError(t, http.NewResponseController(rw).Flush())
				}
			})

			handler, err := New(t.Context(), next, dynamic.Compress{Encodings: defaultSupportedEncodings}, "testing")
			require.NoError(t, err)

			for i, response := range responses {
				current = i

				req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
				req.Header.Set(acceptEncodingHeader, encoding)

				rw := httptest.NewRecorder()
				handler.ServeHTTP(rw, req)

				expectedEncoding := ""
				if response.compressed {
					expectedEncoding = encoding
				}
				assert.Equal(t, expectedEncoding, rw.Header().Get(contentEncodingHeader))

				got, err := io.ReadAll(newDecoder(t, expectedEncoding, rw.Body))
				require.NoError(t, err)
				// Compared as strings, as a nil body and an empty read are both empty.
				assert.Equal(t, string(response.body), string(got))
			}
		})
	}
}

func TestDefaultEncoding(t *testing.T) {
	testCases := []struct {
		desc             string
		defaultEncoding  string
		acceptEncoding   string
		expectedEncoding string
	}{
		{
			desc: "no accept header, no default encoding",
		},
		{
			desc:             "no accept header, gzip default encoding",
			defaultEncoding:  gzipName,
			expectedEncoding: gzipName,
		},
		{
			desc:             "no accept header, brotli default encoding",
			defaultEncoding:  brotliName,
			expectedEncoding: brotliName,
		},
		{
			desc:             "no accept header, zstd default encoding",
			defaultEncoding:  zstdName,
			expectedEncoding: zstdName,
		},
		{
			desc:             "accept any header, no default encoding",
			acceptEncoding:   "*",
			expectedEncoding: gzipName,
		},
		{
			desc:             "accept any header, gzip default encoding",
			defaultEncoding:  gzipName,
			acceptEncoding:   "*",
			expectedEncoding: gzipName,
		},
		{
			desc:             "accept any header, brotli default encoding",
			defaultEncoding:  brotliName,
			acceptEncoding:   "*",
			expectedEncoding: brotliName,
		},
		{
			desc:             "accept any header, zstd default encoding",
			defaultEncoding:  zstdName,
			acceptEncoding:   "*",
			expectedEncoding: zstdName,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				_, err := rw.Write(bigTestBody)
				assert.NoError(t, err)
			})

			cfg := dynamic.Compress{
				Encodings:       defaultSupportedEncodings,
				DefaultEncoding: test.defaultEncoding,
			}
			handler, err := New(t.Context(), next, cfg, "testing")
			require.NoError(t, err)

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
			if test.acceptEncoding != "" {
				req.Header.Set(acceptEncodingHeader, test.acceptEncoding)
			}

			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, req)

			assert.Equal(t, test.expectedEncoding, rw.Header().Get(contentEncodingHeader))

			got, err := io.ReadAll(newDecoder(t, test.expectedEncoding, rw.Body))
			require.NoError(t, err)
			assert.Equal(t, bigTestBody, got)
		})
	}
}

func TestResponseHeaders(t *testing.T) {
	htmlBody := []byte("<!DOCTYPE html><html><body>" + strings.Repeat("aaabbbccc", 200) + "</body></html>")

	testCases := []struct {
		desc            string
		cfg             dynamic.Compress
		headers         map[string]string
		statusCode      int
		body            []byte
		flush           bool
		compressed      bool
		expectedHeaders map[string]string
	}{
		{
			desc:            "partial content is not compressed",
			headers:         map[string]string{contentRange: "bytes 0-6006/10000"},
			statusCode:      http.StatusPartialContent,
			body:            bigTestBody,
			expectedHeaders: map[string]string{contentRange: "bytes 0-6006/10000"},
		},
		{
			desc:            "Accept-Ranges is removed when compressing",
			headers:         map[string]string{acceptRanges: "bytes"},
			body:            bigTestBody,
			compressed:      true,
			expectedHeaders: map[string]string{acceptRanges: ""},
		},
		{
			desc:            "Accept-Ranges is kept when not compressing",
			headers:         map[string]string{acceptRanges: "bytes"},
			body:            smallTestBody,
			expectedHeaders: map[string]string{acceptRanges: "bytes"},
		},
		{
			desc:            "Content-Length above the minimum size is compressed and removed",
			headers:         map[string]string{contentLength: strconv.Itoa(len(bigTestBody))},
			body:            bigTestBody,
			compressed:      true,
			expectedHeaders: map[string]string{contentLength: ""},
		},
		{
			desc:            "Content-Length below the minimum size is not compressed when flushed",
			headers:         map[string]string{contentLength: strconv.Itoa(len(smallTestBody))},
			body:            smallTestBody,
			flush:           true,
			expectedHeaders: map[string]string{contentLength: strconv.Itoa(len(smallTestBody))},
		},
		{
			desc:            "missing Content-Type is sniffed",
			body:            htmlBody,
			compressed:      true,
			expectedHeaders: map[string]string{contentType: "text/html; charset=utf-8"},
		},
		{
			desc:            "sniffed Content-Type is matched against excluded content types",
			cfg:             dynamic.Compress{ExcludedContentTypes: []string{"text/html"}},
			body:            htmlBody,
			expectedHeaders: map[string]string{contentType: "text/html; charset=utf-8"},
		},
		{
			desc:            "No-Gzip-Compression header does not disable compression",
			headers:         map[string]string{"No-Gzip-Compression": "1"},
			body:            bigTestBody,
			compressed:      true,
			expectedHeaders: map[string]string{"No-Gzip-Compression": "1"},
		},
	}

	for _, encoding := range []string{gzipName, brotliName, zstdName} {
		for _, test := range testCases {
			t.Run(encoding+", "+test.desc, func(t *testing.T) {
				t.Parallel()

				next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
					for k, v := range test.headers {
						rw.Header().Set(k, v)
					}
					if test.statusCode != 0 {
						rw.WriteHeader(test.statusCode)
					}

					_, err := rw.Write(test.body)
					assert.NoError(t, err)

					if test.flush {
						assert.NoError(t, http.NewResponseController(rw).Flush())
					}
				})

				cfg := test.cfg
				cfg.Encodings = defaultSupportedEncodings
				handler, err := New(t.Context(), next, cfg, "testing")
				require.NoError(t, err)

				req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
				req.Header.Set(acceptEncodingHeader, encoding)

				rw := httptest.NewRecorder()
				handler.ServeHTTP(rw, req)

				for k, v := range test.expectedHeaders {
					assert.Equal(t, v, rw.Header().Get(k))
				}

				expectedEncoding := ""
				if test.compressed {
					expectedEncoding = encoding
				}
				assert.Equal(t, expectedEncoding, rw.Header().Get(contentEncodingHeader))

				got, err := io.ReadAll(newDecoder(t, expectedEncoding, rw.Body))
				require.NoError(t, err)
				assert.Equal(t, test.body, got)
			})
		}
	}
}

func TestCloseNotifierNotOffered(t *testing.T) {
	for _, encoding := range []string{gzipName, brotliName, zstdName} {
		t.Run(encoding, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
				// Handlers must watch the request context instead.
				_, ok := rw.(http.CloseNotifier)
				assert.False(t, ok)
			})

			handler, err := New(t.Context(), next, dynamic.Compress{Encodings: defaultSupportedEncodings}, "testing")
			require.NoError(t, err)

			server := httptest.NewServer(handler)
			t.Cleanup(server.Close)

			req := testhelpers.MustNewRequest(http.MethodGet, server.URL, nil)
			req.Header.Set(acceptEncodingHeader, encoding)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			require.NoError(t, resp.Body.Close())
		})
	}
}

func TestIntegrationShouldCompress(t *testing.T) {
	fakeBody := generateBytes(100000)

	testCases := []struct {
		name               string
		handler            http.Handler
		expectedStatusCode int
	}{
		{
			name: "when AcceptEncoding header is present",
			handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				_, err := rw.Write(fakeBody)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}
			}),
			expectedStatusCode: http.StatusOK,
		},
		{
			name: "when AcceptEncoding header is present and status code Created",
			handler: http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusCreated)
				_, err := rw.Write(fakeBody)
				if err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}
			}),
			expectedStatusCode: http.StatusCreated,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			compress, err := New(t.Context(), test.handler, dynamic.Compress{Encodings: defaultSupportedEncodings}, "testing")
			require.NoError(t, err)

			ts := httptest.NewServer(compress)
			defer ts.Close()

			req := testhelpers.MustNewRequest(http.MethodGet, ts.URL, nil)
			req.Header.Add(acceptEncodingHeader, gzipName)

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatusCode, resp.StatusCode)

			assert.Equal(t, gzipName, resp.Header.Get(contentEncodingHeader))
			assert.Equal(t, acceptEncodingHeader, resp.Header.Get(varyHeader))

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			if assert.ObjectsAreEqualValues(body, fakeBody) {
				assert.Fail(t, "expected a compressed body", "got %v", body)
			}
		})
	}
}

func TestMinResponseBodyBytes(t *testing.T) {
	fakeBody := generateBytes(100000)

	testCases := []struct {
		name                 string
		minResponseBodyBytes int
		expectedCompression  bool
	}{
		{
			name:                "should compress",
			expectedCompression: true,
		},
		{
			name:                 "should not compress",
			minResponseBodyBytes: 100001,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
			req.Header.Add(acceptEncodingHeader, gzipName)

			next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				if _, err := rw.Write(fakeBody); err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				}
			})
			cfg := dynamic.Compress{
				MinResponseBodyBytes: test.minResponseBodyBytes,
				Encodings:            defaultSupportedEncodings,
			}
			handler, err := New(t.Context(), next, cfg, "testing")
			require.NoError(t, err)

			rw := httptest.NewRecorder()
			handler.ServeHTTP(rw, req)

			if test.expectedCompression {
				assert.Equal(t, gzipName, rw.Header().Get(contentEncodingHeader))
				assert.NotEqual(t, rw.Body.Bytes(), fakeBody)
				return
			}

			assert.Empty(t, rw.Header().Get(contentEncodingHeader))
			assert.Equal(t, rw.Body.Bytes(), fakeBody)
		})
	}
}

func TestZstdWindowSize(t *testing.T) {
	fakeBody := generateBytes(1 << 20)

	next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if _, err := rw.Write(fakeBody); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
	})
	handler, err := New(t.Context(), next, dynamic.Compress{Encodings: defaultSupportedEncodings}, "testing")
	require.NoError(t, err)

	req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)
	req.Header.Add(acceptEncodingHeader, zstdName)

	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	assert.Equal(t, zstdName, rw.Header().Get(contentEncodingHeader))

	// Each in-flight response holds one encoder, and the encoder memory grows with the window size.
	reader, err := zstd.NewReader(rw.Body, zstd.WithDecoderMaxWindow(128<<10))
	require.NoError(t, err)
	t.Cleanup(reader.Close)

	body, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, fakeBody, body)
}

// This test is an adapted version of net/http/httputil.Test1xxResponses test.
func Test1xxResponses(t *testing.T) {
	fakeBody := generateBytes(100000)

	testCases := []struct {
		desc     string
		encoding string
	}{
		{
			desc:     "gzip",
			encoding: gzipName,
		},
		{
			desc:     "brotli",
			encoding: brotliName,
		},
		{
			desc:     "zstd",
			encoding: zstdName,
		},
	}
	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				h := w.Header()
				h.Add("Link", "</style.css>; rel=preload; as=style")
				h.Add("Link", "</script.js>; rel=preload; as=script")
				w.WriteHeader(http.StatusEarlyHints)

				h.Add("Link", "</foo.js>; rel=preload; as=script")
				w.WriteHeader(http.StatusProcessing)

				if _, err := w.Write(fakeBody); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			})
			cfg := dynamic.Compress{
				MinResponseBodyBytes: 1024,
				Encodings:            defaultSupportedEncodings,
			}
			compress, err := New(t.Context(), next, cfg, "testing")
			require.NoError(t, err)

			server := httptest.NewServer(compress)
			t.Cleanup(server.Close)
			frontendClient := server.Client()

			checkLinkHeaders := func(t *testing.T, expected, got []string) {
				t.Helper()

				if len(expected) != len(got) {
					t.Errorf("Expected %d link headers; got %d", len(expected), len(got))
				}

				for i := range expected {
					if i >= len(got) {
						t.Errorf("Expected %q link header; got nothing", expected[i])

						continue
					}

					if expected[i] != got[i] {
						t.Errorf("Expected %q link header; got %q", expected[i], got[i])
					}
				}
			}

			var respCounter uint8
			trace := &httptrace.ClientTrace{
				Got1xxResponse: func(code int, header textproto.MIMEHeader) error {
					switch code {
					case http.StatusEarlyHints:
						checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script"}, header["Link"])
					case http.StatusProcessing:
						checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script", "</foo.js>; rel=preload; as=script"}, header["Link"])
					default:
						t.Error("Unexpected 1xx response")
					}

					respCounter++

					return nil
				},
			}
			req, _ := http.NewRequestWithContext(httptrace.WithClientTrace(t.Context(), trace), http.MethodGet, server.URL, nil)
			req.Header.Add(acceptEncodingHeader, test.encoding)

			res, err := frontendClient.Do(req)
			assert.NoError(t, err)

			defer res.Body.Close()

			if respCounter != 2 {
				t.Errorf("Expected 2 1xx responses; got %d", respCounter)
			}
			checkLinkHeaders(t, []string{"</style.css>; rel=preload; as=style", "</script.js>; rel=preload; as=script", "</foo.js>; rel=preload; as=script"}, res.Header["Link"])

			assert.Equal(t, test.encoding, res.Header.Get(contentEncodingHeader))
			body, _ := io.ReadAll(res.Body)
			assert.NotEqual(t, body, fakeBody)
		})
	}
}

func BenchmarkCompressGzip(b *testing.B) {
	runCompressionBenchmark(b, gzipName)
}

func BenchmarkCompressBrotli(b *testing.B) {
	runCompressionBenchmark(b, brotliName)
}

func BenchmarkCompressZstandard(b *testing.B) {
	runCompressionBenchmark(b, zstdName)
}

func runCompressionBenchmark(b *testing.B, algorithm string) {
	b.Helper()

	testCases := []struct {
		name     string
		parallel bool
		size     int
	}{
		{"2k", false, 2048},
		{"20k", false, 20480},
		{"100k", false, 102400},
		{"2k parallel", true, 2048},
		{"20k parallel", true, 20480},
		{"100k parallel", true, 102400},
	}

	for _, test := range testCases {
		b.Run(test.name, func(b *testing.B) {
			baseBody := generateBytes(test.size)

			next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				_, err := rw.Write(baseBody)
				assert.NoError(b, err)
			})
			handler, _ := New(b.Context(), next, dynamic.Compress{}, "testing")

			req, _ := http.NewRequest(http.MethodGet, "/whatever", nil)
			req.Header.Set("Accept-Encoding", algorithm)

			b.ReportAllocs()
			b.SetBytes(int64(test.size))
			if test.parallel {
				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						runBenchmark(b, req, handler, algorithm)
					}
				})
				return
			}

			b.ResetTimer()
			for range b.N {
				runBenchmark(b, req, handler, algorithm)
			}
		})
	}
}

func runBenchmark(b *testing.B, req *http.Request, handler http.Handler, algorithm string) {
	b.Helper()

	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if code := res.Code; code != 200 {
		b.Fatalf("Expected 200 but got %d", code)
	}

	assert.Equal(b, algorithm, res.Header().Get(contentEncodingHeader))
}

func newDecoder(t *testing.T, encoding string, body io.Reader) io.Reader {
	t.Helper()

	switch encoding {
	case gzipName:
		reader, err := gzip.NewReader(body)
		require.NoError(t, err)
		return reader
	case brotliName:
		return brotli.NewReader(body)
	case zstdName:
		reader, err := zstd.NewReader(body)
		require.NoError(t, err)
		t.Cleanup(reader.Close)
		return reader
	default:
		return body
	}
}

func generateBytes(length int) []byte {
	var value []byte
	for i := range length {
		value = append(value, 0x61+byte(i))
	}
	return value
}
