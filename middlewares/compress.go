package middlewares

import (
	"github.com/NYTimes/gziphandler"
	"net/http"
)

// Compress is a middleware that allows redirections
type Compress struct {
}

// ServerHTTP is a function used by negroni
func (c *Compress) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	newGzipHandler := gziphandler.GzipHandler(next)
	newGzipHandler.ServeHTTP(rw, r)
}
