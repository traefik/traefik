package middlewares

import (
	"compress/gzip"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/containous/traefik/log"
)

// Compress is a middleware that allows redirection
type Compress struct{}

// ServerHTTP is a function used by Negroni
func (c *Compress) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	gzipHandler(next).ServeHTTP(rw, r)
}

func gzipHandler(h http.Handler) http.Handler {
	wrapper, err := gziphandler.GzipHandlerWithOpts(
		&gziphandler.GzipResponseWriterWrapper{},
		gziphandler.CompressionLevel(gzip.DefaultCompression),
		gziphandler.MinSize(gziphandler.DefaultMinSize))
	if err != nil {
		log.Error(err)
	}
	return wrapper(h)
}
