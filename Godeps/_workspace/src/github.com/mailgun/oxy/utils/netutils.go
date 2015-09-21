package utils

import (
	"io"
	"net/http"
	"net/url"
)

// ProxyWriter helps to capture response headers and status code
// from the ServeHTTP. It can be safely passed to ServeHTTP handler,
// wrapping the real response writer.
type ProxyWriter struct {
	W    http.ResponseWriter
	Code int
}

func (p *ProxyWriter) StatusCode() int {
	if p.Code == 0 {
		// per contract standard lib will set this to http.StatusOK if not set
		// by user, here we avoid the confusion by mirroring this logic
		return http.StatusOK
	}
	return p.Code
}

func (p *ProxyWriter) Header() http.Header {
	return p.W.Header()
}

func (p *ProxyWriter) Write(buf []byte) (int, error) {
	return p.W.Write(buf)
}

func (p *ProxyWriter) WriteHeader(code int) {
	p.Code = code
	p.W.WriteHeader(code)
}

func (p *ProxyWriter) Flush() {
	if f, ok := p.W.(http.Flusher); ok {
		f.Flush()
	}
}

func NewBufferWriter(w io.WriteCloser) *BufferWriter {
	return &BufferWriter{
		W: w,
		H: make(http.Header),
	}
}

type BufferWriter struct {
	H    http.Header
	Code int
	W    io.WriteCloser
}

func (b *BufferWriter) Close() error {
	return b.W.Close()
}

func (b *BufferWriter) Header() http.Header {
	return b.H
}

func (b *BufferWriter) Write(buf []byte) (int, error) {
	return b.W.Write(buf)
}

// WriteHeader sets rw.Code.
func (b *BufferWriter) WriteHeader(code int) {
	b.Code = code
}

type nopWriteCloser struct {
	io.Writer
}

func (*nopWriteCloser) Close() error { return nil }

// NopCloser returns a WriteCloser with a no-op Close method wrapping
// the provided Writer w.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{w}
}

// CopyURL provides update safe copy by avoiding shallow copying User field
func CopyURL(i *url.URL) *url.URL {
	out := *i
	if i.User != nil {
		out.User = &(*i.User)
	}
	return &out
}

// CopyHeaders copies http headers from source to destination, it
// does not overide, but adds multiple headers
func CopyHeaders(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// HasHeaders determines whether any of the header names is present in the http headers
func HasHeaders(names []string, headers http.Header) bool {
	for _, h := range names {
		if headers.Get(h) != "" {
			return true
		}
	}
	return false
}

// RemoveHeaders removes the header with the given names from the headers map
func RemoveHeaders(headers http.Header, names ...string) {
	for _, h := range names {
		headers.Del(h)
	}
}
