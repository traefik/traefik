package gziphandler

import (
	"bufio"
	"net"
	"net/http"
)

const (
	contentEncodingHeader = "Content-Encoding"
)

// ----------

// http.ResponseWriter
// http.Hijacker
type GzipWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)
	Hijack() (net.Conn, *bufio.ReadWriter, error)
	Close() error
	SetResponseWriter(http.ResponseWriter)
	setIndex(int)
	setMinSize(int)
	setContentTypes([]string)
}

func (w *GzipResponseWriter) SetResponseWriter(rw http.ResponseWriter) {
	w.ResponseWriter = rw
}

func (w *GzipResponseWriter) setIndex(index int) {
	w.index = index
}

func (w *GzipResponseWriter) setMinSize(minSize int) {
	w.minSize = minSize
}

func (w *GzipResponseWriter) setContentTypes(contentTypes []string) {
	w.contentTypes = contentTypes
}

// --------

type GzipResponseWriterWrapper struct {
	GzipResponseWriter
}

func (g *GzipResponseWriterWrapper) Write(b []byte) (int, error) {
	if g.gw == nil && isEncoded(g.Header()) {
		if g.code != 0 {
			g.ResponseWriter.WriteHeader(g.code)
		}
		return g.ResponseWriter.Write(b)
	}
	return g.GzipResponseWriter.Write(b)
}

func isEncoded(headers http.Header) bool {
	header := headers.Get(contentEncodingHeader)
	// According to https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding,
	// content is not encoded if the header 'Content-Encoding' is empty or equals to 'identity'.
	return header != "" && header != "identity"
}
