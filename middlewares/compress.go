package middlewares

import (
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/containous/traefik/middlewares/common"
)

// Compress is a middleware that gzips content.
type Compress struct {
	common.BasicMiddleware
}

var _ common.Middleware = &Compress{}

// NewCompress creates a new middleware that gzips content.
func NewCompress(next http.Handler) common.Middleware {
	return &Compress{common.NewMiddleware(gziphandler.GzipHandler(next))}
}

func (c *Compress) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	c.Next().ServeHTTP(rw, r)
}
