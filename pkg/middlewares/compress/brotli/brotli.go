package brotli

import (
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
)

type bWriter struct {
	rw http.ResponseWriter
	*brotli.Writer

	minSize int
}

func (b *bWriter) Header() http.Header {
	return b.rw.Header()
}

func (b *bWriter) WriteHeader(statusCode int) {
	b.rw.WriteHeader(statusCode)
}

func (b *bWriter) Write(p []byte) (n int, err error) {
	if len(p) < b.minSize {
		b.rw.Header().Del("Vary")
		b.rw.Header().Set("Content-Encoding", "identity")
		return b.rw.Write(p)
	}

	return b.Writer.Write(p)
}

type config struct {
	compression int
	minSize     int
}

type option func(c *config)

// WithCompressionLevel allows setting the compression level for brotli.
// 0 for speed, 11 for compression.
func WithCompressionLevel(compression int) option {
	return func(c *config) {
		c.compression = brotli.DefaultCompression
		if compression >= brotli.BestSpeed && compression <= brotli.BestCompression {
			c.compression = compression
		}
	}
}

// WithMinSize allows setting the minimum size to compress a response.
// Default is 1024 bytes.
func WithMinSize(minSize int) option {
	return func(c *config) {
		c.minSize = 1024
		if minSize >= 0 {
			c.minSize = minSize
		}
	}
}

// NewMiddleware returns a new brotli compressing middleware.
func NewMiddleware(opts ...option) func(http.Handler) http.HandlerFunc {
	cfg := config{
		compression: brotli.DefaultCompression,
	}

	for _, o := range opts {
		o(&cfg)
	}

	return func(h http.Handler) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodHead {
				h.ServeHTTP(rw, r)
				return
			}

			if !acceptsBr(r.Header.Get("Accept-Encoding")) {
				h.ServeHTTP(rw, r)
				return
			}

			rw.Header().Add("Vary", "Accept-Encoding")
			rw.Header().Set("Content-Encoding", "br")
			bw := &bWriter{
				Writer: brotli.NewWriterOptions(rw, brotli.WriterOptions{
					Quality: cfg.compression,
				}),
				rw:      rw,
				minSize: cfg.minSize,
			}

			defer bw.Close()

			h.ServeHTTP(bw, r)
		}
	}
}

// acceptsBr is a naive method of checking if "br" was set as an accepted encoding
func acceptsBr(acceptEncoding string) bool {
	for _, v := range strings.Split(acceptEncoding, ",") {
		for i, e := range strings.Split(v, ";") {
			if i == 0 && strings.Contains(e, "br") {
				return true
			}
		}
	}

	return false
}
