package middlewares

import (
	"github.com/NYTimes/gziphandler"
	"net/http"
)

// Rewrite is a middleware that allows redirections
type Compress struct {
}

//
func (_ *Compress) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	newGzipHandler := gziphandler.GzipHandler(next)
	newGzipHandler.ServeHTTP(rw, r)
}
