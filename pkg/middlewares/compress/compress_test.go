package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"net/textproto"
	"testing"

	"github.com/klauspost/compress/gzhttp"
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
			// In this test, the default encodings are defaulted to gzip, brotli, and zstd,
			// which make gzip the default encoding, and will be selected.
			// However, the klauspost/compress gzhttp handler does not compress when Accept-Encoding: * is set.
			// Until klauspost/compress gzhttp package supports the asterisk,
			// we will not support it when selecting the gzip encoding.
			desc:            "accept any header",
			acceptEncHeader: "*",
			expEncoding:     "",
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
			desc:            "multi accept header list, prefer gzip",
			acceptEncHeader: "gzip, br, zstd",
			expEncoding:     gzipName,
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

	baseBody := generateBytes(gzhttp.DefaultMinSize)

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

	fakeCompressedBody := generateBytes(gzhttp.DefaultMinSize)
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

	fakeBody := generateBytes(gzhttp.DefaultMinSize)
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

	fakeBody := generateBytes(gzhttp.DefaultMinSize)
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

	fakeBody := generateBytes(gzhttp.DefaultMinSize)
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

	fakeBody := generateBytes(gzhttp.DefaultMinSize)
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

	fakeBody := generateBytes(gzhttp.DefaultMinSize)
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
	baseBody := generateBytes(gzhttp.DefaultMinSize)

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
	baseBody := generateBytes(gzhttp.DefaultMinSize)

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
		rw.(http.Flusher).Flush()
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

func generateBytes(length int) []byte {
	var value []byte
	for i := range length {
		value = append(value, 0x61+byte(i))
	}
	return value
}
