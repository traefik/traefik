package middlewares

import (
	"context"
	"net/http"
)

const (
	// ReplacePathKey is the key within the request context used to
	// store the replaced path
	ReplacePathKey key = "ReplacePath"
	// ReplacedPathHeader is the default header to set the old path to
	ReplacedPathHeader = "X-Replaced-Path"
)

// ReplacePath is a middleware used to replace the path of a URL request
type ReplacePath struct {
	Handler http.Handler
	Path    string
}

func (s *ReplacePath) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r = r.WithContext(context.WithValue(r.Context(), ReplacePathKey, r.URL.Path))
	r.Header.Add(ReplacedPathHeader, r.URL.Path)
	r.URL.Path = s.Path
	r.RequestURI = r.URL.RequestURI()
	s.Handler.ServeHTTP(w, r)
}
