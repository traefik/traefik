package brotli

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
)

// DefaultMinSize is the default minimum size until we enable brotli compression.
// 1500 bytes is the MTU size for the internet since that is the largest size allowed at the network layer.
// If you take a file that is 1300 bytes and compress it to 800 bytes, it’s still transmitted in that same 1500 byte packet regardless, so you’ve gained nothing.
// That being the case, you should restrict the gzip compression to files with a size (plus header) greater than a single packet,
// 1024 bytes (1KB) is therefore default.
// From [github.com/klauspost/compress/gzhttp](https://github.com/klauspost/compress/tree/master/gzhttp).
var DefaultMinSize = 1024

type bWriter struct {
	rw http.ResponseWriter
	*brotli.Writer

	minSize int
	written bool
}

func (b *bWriter) Header() http.Header {
	return b.rw.Header()
}

func (b *bWriter) WriteHeader(statusCode int) {
	b.rw.WriteHeader(statusCode)
}

func (b *bWriter) Write(p []byte) (int, error) {
	if len(p) < b.minSize {
		b.rw.Header().Del("Vary")
		b.rw.Header().Set("Content-Encoding", "identity")

		n, err := b.rw.Write(p)
		b.rw.Header().Add("Content-Length", strconv.Itoa(n))
		return n, err
	}

	b.written = true
	return b.Writer.Write(p)
}

// Config is the brotli middleware configuration.
type Config struct {
	// Compression level.
	Compression int
	// MinSize is the minimum size until we enable brotli compression.
	MinSize int
}

// NewMiddleware returns a new brotli compressing middleware.
func NewMiddleware(cfg Config) func(http.Handler) http.HandlerFunc {
	return func(h http.Handler) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Add("Vary", "Accept-Encoding")
			rw.Header().Set("Content-Encoding", "br")
			bw := &bWriter{
				Writer: brotli.NewWriterOptions(rw, brotli.WriterOptions{
					Quality: cfg.Compression,
				}),
				rw:      rw,
				minSize: cfg.MinSize,
			}

			defer func() {
				if bw.written {
					bw.Close()
				}
			}()

			h.ServeHTTP(bw, r)
		}
	}
}

// AcceptsBr is a naive method to check whether brotli is an accepted encoding.
func AcceptsBr(acceptEncoding string) bool {
	for _, v := range strings.Split(acceptEncoding, ",") {
		for i, e := range strings.Split(strings.TrimSpace(v), ";") {
			if i == 0 && (e == "br" || e == "*") {
				return true
			}

			break
		}
	}

	return false
}
