package middlewares

import (
	log "github.com/Sirupsen/logrus"
	"github.com/vulcand/vulcand/plugin/rewrite"
	"net/http"
)

// Rewrite is a middleware that allows redirections
type Rewrite struct {
	rewriter *rewrite.Rewrite
}

// NewRewrite creates a Rewrite middleware
func NewRewrite(regex, replacement string, redirect bool) (*Rewrite, error) {
	rewriter, err := rewrite.NewRewrite(regex, replacement, false, redirect)
	if err != nil {
		return nil, err
	}
	return &Rewrite{rewriter: rewriter}, nil
}

//
func (rewrite *Rewrite) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	handler, err := rewrite.rewriter.NewHandler(next)
	if err != nil {
		log.Error("Error in rewrite middleware ", err)
		return
	}
	handler.ServeHTTP(rw, r)
}
