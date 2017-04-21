package middlewares

import (
	"net/http"

	"github.com/containous/traefik/middlewares/common"
	"github.com/vulcand/vulcand/plugin/rewrite"
)

// Rewrite is a middleware that allows redirections
type Rewrite struct {
	common.BasicMiddleware
}

var _ common.Middleware = &Rewrite{}

// NewRewrite creates a Rewrite middleware. The regular expressions are compiled only once.
func NewRewrite(regex, replacement string, redirect bool, next http.Handler) (common.Middleware, error) {
	rewriter, err := rewrite.NewRewrite(regex, replacement, false, redirect)
	if err != nil {
		return nil, err
	}
	handler, err := rewriter.NewHandler(next)
	if err != nil {
		return nil, err
	}
	return &Rewrite{common.NewMiddleware(handler)}, nil
}

func (rewrite *Rewrite) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	rewrite.Next().ServeHTTP(rw, r)
}
