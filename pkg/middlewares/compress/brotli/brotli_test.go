package brotli

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/stretchr/testify/assert"
	"github.com/traefik/traefik/v2/pkg/testhelpers"
)

func generateBytes(length int) []byte {
	var value []byte
	for i := 0; i < length; i++ {
		value = append(value, 0x61+byte(i))
	}
	return value
}

func TestBWriter(t *testing.T) {
	type test struct {
		name        string
		minSize     int
		writeData   []byte
		compression bool
	}
	testCases := []test{
		{
			name:        "data less than min - no compression",
			minSize:     100,
			writeData:   generateBytes(10),
			compression: false,
		},
		{
			name:        "data more than min - compression",
			minSize:     100,
			writeData:   generateBytes(100),
			compression: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			trw := httptest.NewRecorder()
			bw := &bWriter{
				Writer: brotli.NewWriterOptions(trw, brotli.WriterOptions{
					Quality: 6,
				}),
				rw:      trw,
				minSize: testCase.minSize,
			}
			_, err := bw.Write(testCase.writeData)
			assert.Nil(t, err)

			if testCase.compression {
				assert.Less(t, len(trw.Body.Bytes()), len(testCase.writeData))
			} else {
				assert.Equal(t, len(testCase.writeData), len(trw.Body.Bytes()))
			}
		})
	}

	trw := httptest.NewRecorder()
	bw := &bWriter{
		rw:      trw,
		minSize: 100,
	}
	bw.WriteHeader(http.StatusOK)
	assert.Equal(t, http.StatusOK, trw.Code)

	trw.Header().Set("traefik", "rocks")
	assert.Equal(t, "rocks", bw.Header().Get("traefik"))
}

func TestWithCompressionLevel(t *testing.T) {
	type testCompressionLevel struct {
		name         string
		compression  int
		expectedComp int
	}
	testCases := []testCompressionLevel{
		{
			name:         "bad level",
			compression:  -1,
			expectedComp: brotli.DefaultCompression,
		},
		{
			name:         "good level",
			compression:  1,
			expectedComp: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := &config{}
			fn := WithCompressionLevel(testCase.compression)
			fn(cfg)
			assert.Equal(t, testCase.expectedComp, cfg.compression)
		})
	}
}

func TestWithMinSize(t *testing.T) {
	type testCompressionLevel struct {
		name         string
		size         int
		expectedSize int
	}
	testCases := []testCompressionLevel{
		{
			name:         "bad level",
			size:         -1,
			expectedSize: DefaultMinSize,
		},
		{
			name:         "good level",
			size:         1,
			expectedSize: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			cfg := &config{}
			fn := WithMinSize(testCase.size)
			fn(cfg)
			assert.Equal(t, testCase.expectedSize, cfg.minSize)
		})
	}
}

func TestNewMiddleware(t *testing.T) {
	type test struct {
		name        string
		writeData   []byte
		expCompress bool
		expEncoding string
	}
	testCases := []test{
		{
			name:        "big request",
			expCompress: true,
			expEncoding: "br",
			writeData:   generateBytes(DefaultMinSize),
		},
		{
			name:        "small request",
			expCompress: false,
			expEncoding: "identity",
			writeData:   generateBytes(DefaultMinSize - 1),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := testhelpers.MustNewRequest(http.MethodGet, "http://localhost", nil)

			next := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				_, err := rw.Write(testCase.writeData)
				assert.NoError(t, err)
			})

			rw := httptest.NewRecorder()
			NewMiddleware(WithMinSize(DefaultMinSize))(next).ServeHTTP(rw, req)

			if testCase.expCompress {
				assert.Equal(t, "Accept-Encoding", rw.Header().Get("Vary"))

				assert.Less(t, len(rw.Body.Bytes()), len(testCase.writeData), "expected a compressed body got %v", rw.Body.Bytes())
			} else {
				assert.Equal(t, testCase.writeData, rw.Body.Bytes())
			}

			assert.Equal(t, testCase.expEncoding, rw.Header().Get("Content-Encoding"))
		})
	}
}

func TestAcceptsBr(t *testing.T) {
	type test struct {
		name     string
		encoding string
		accepted bool
	}
	testCases := []test{
		{
			name:     "simple br accept",
			encoding: "br",
			accepted: true,
		},
		{
			name:     "br accept with quality",
			encoding: "br;q=1.0",
			accepted: true,
		},
		{
			name:     "br accept with quality multiple",
			encoding: "gzip;1.0, br;q=0.8",
			accepted: true,
		},
		{
			name:     "any accept with quality multiple",
			encoding: "gzip;q=0.8, *;q=0.1",
			accepted: true,
		},
		{
			name:     "any accept",
			encoding: "*",
			accepted: true,
		},
		{
			name:     "gzip accept",
			encoding: "gzip",
			accepted: false,
		},
		{
			name:     "gzip accept multiple",
			encoding: "gzip, identity",
			accepted: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.accepted, AcceptsBr(testCase.encoding))
		})
	}
}

func BenchmarkAcceptsBr(b *testing.B) {

	b.Run("AcceptBR", func(b *testing.B) {
		b.ReportAllocs()
		AcceptsBr("gzip;q=0.8, *;q=0.1")
	})

	b.Run("AcceptBR 2", func(b *testing.B) {
		b.ReportAllocs()
		AcceptsBr("gzip;q=0.8, br;q=0.1")
	})

	b.Run("Contains", func(b *testing.B) {
		b.ReportAllocs()
		if strings.Contains("gzip;q=0.8, *;q=0.1", "*") ||
			strings.Contains("gzip;q=0.8, *;q=0.1", "br") {

		}
	})

	b.Run("Contains 2", func(b *testing.B) {
		b.ReportAllocs()
		if strings.Contains("gzip;q=0.8, br;q=0.1", "*") ||
			strings.Contains("gzip;q=0.8, br;q=0.1", "br") {

		}
	})
}
