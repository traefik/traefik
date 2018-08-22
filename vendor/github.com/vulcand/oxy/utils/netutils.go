package utils

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"reflect"

	log "github.com/sirupsen/logrus"
)

// ProxyWriter calls recorder, used to debug logs
type ProxyWriter struct {
	w      http.ResponseWriter
	code   int
	length int64

	log *log.Logger
}

// NewProxyWriter creates a new ProxyWriter
func NewProxyWriter(w http.ResponseWriter) *ProxyWriter {
	return NewProxyWriterWithLogger(w, log.StandardLogger())
}

// NewProxyWriterWithLogger creates a new ProxyWriter
func NewProxyWriterWithLogger(w http.ResponseWriter, l *log.Logger) *ProxyWriter {
	return &ProxyWriter{
		w:   w,
		log: l,
	}
}

// StatusCode gets status code
func (p *ProxyWriter) StatusCode() int {
	if p.code == 0 {
		// per contract standard lib will set this to http.StatusOK if not set
		// by user, here we avoid the confusion by mirroring this logic
		return http.StatusOK
	}
	return p.code
}

// GetLength gets content length
func (p *ProxyWriter) GetLength() int64 {
	return p.length
}

// Header gets response header
func (p *ProxyWriter) Header() http.Header {
	return p.w.Header()
}

func (p *ProxyWriter) Write(buf []byte) (int, error) {
	p.length = p.length + int64(len(buf))
	return p.w.Write(buf)
}

// WriteHeader writes status code
func (p *ProxyWriter) WriteHeader(code int) {
	p.code = code
	p.w.WriteHeader(code)
}

// Flush flush the writer
func (p *ProxyWriter) Flush() {
	if f, ok := p.w.(http.Flusher); ok {
		f.Flush()
	}
}

// CloseNotify returns a channel that receives at most a single value (true)
// when the client connection has gone away.
func (p *ProxyWriter) CloseNotify() <-chan bool {
	if cn, ok := p.w.(http.CloseNotifier); ok {
		return cn.CloseNotify()
	}
	p.log.Debugf("Upstream ResponseWriter of type %v does not implement http.CloseNotifier. Returning dummy channel.", reflect.TypeOf(p.w))
	return make(<-chan bool)
}

// Hijack lets the caller take over the connection.
func (p *ProxyWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hi, ok := p.w.(http.Hijacker); ok {
		return hi.Hijack()
	}
	p.log.Debugf("Upstream ResponseWriter of type %v does not implement http.Hijacker. Returning dummy channel.", reflect.TypeOf(p.w))
	return nil, nil, fmt.Errorf("the response writer that was wrapped in this proxy, does not implement http.Hijacker. It is of type: %v", reflect.TypeOf(p.w))
}

// NewBufferWriter creates a new BufferWriter
func NewBufferWriter(w io.WriteCloser) *BufferWriter {
	return &BufferWriter{
		W: w,
		H: make(http.Header),
	}
}

// BufferWriter buffer writer
type BufferWriter struct {
	H    http.Header
	Code int
	W    io.WriteCloser
}

// Close close the writer
func (b *BufferWriter) Close() error {
	return b.W.Close()
}

// Header gets response header
func (b *BufferWriter) Header() http.Header {
	return b.H
}

func (b *BufferWriter) Write(buf []byte) (int, error) {
	return b.W.Write(buf)
}

// WriteHeader writes status code
func (b *BufferWriter) WriteHeader(code int) {
	b.Code = code
}

// CloseNotify returns a channel that receives at most a single value (true)
// when the client connection has gone away.
func (b *BufferWriter) CloseNotify() <-chan bool {
	if cn, ok := b.W.(http.CloseNotifier); ok {
		return cn.CloseNotify()
	}
	log.Warningf("Upstream ResponseWriter of type %v does not implement http.CloseNotifier. Returning dummy channel.", reflect.TypeOf(b.W))
	return make(<-chan bool)
}

// Hijack lets the caller take over the connection.
func (b *BufferWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hi, ok := b.W.(http.Hijacker); ok {
		return hi.Hijack()
	}
	log.Debugf("Upstream ResponseWriter of type %v does not implement http.Hijacker. Returning dummy channel.", reflect.TypeOf(b.W))
	return nil, nil, fmt.Errorf("the response writer that was wrapped in this proxy, does not implement http.Hijacker. It is of type: %v", reflect.TypeOf(b.W))
}

type nopWriteCloser struct {
	io.Writer
}

func (*nopWriteCloser) Close() error { return nil }

// NopWriteCloser returns a WriteCloser with a no-op Close method wrapping
// the provided Writer w.
func NopWriteCloser(w io.Writer) io.WriteCloser {
	return &nopWriteCloser{Writer: w}
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
func CopyHeaders(dst http.Header, src http.Header) {
	for k, vv := range src {
		dst[k] = append(dst[k], vv...)
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
