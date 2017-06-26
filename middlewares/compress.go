package middlewares

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
)

const (
	contentEncodingHeader = "Content-Encoding"
)

// Compress is a middleware that allows redirection
type Compress struct{}

// ServerHTTP is a function used by Negroni
func (c *Compress) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if isEncoded(r.Header) {
		next.ServeHTTP(rw, r)
	} else {
		newGzipHandler := gziphandler.GzipHandler(next)
		newGzipHandler.ServeHTTP(rw, r)
	}
}

func isEncoded(headers http.Header) bool {
	header := headers.Get(contentEncodingHeader)
	// According to https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding,
	// content is not encoded if the header 'Content-Encoding' is empty or equals to 'identity'.
	return header != "" && header != "identity"
}
