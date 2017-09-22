package middlewares

import (
	"net/http"
)

// ReplacedPathHeader is the default header to set the old path to
const ReplacedPathHeader = "X-Replaced-Path"

// ReplacePath is a middleware used to replace the path of a URL request
type ReplacePath struct {
	Handler http.Handler
	Path    string
}

func (s *ReplacePath) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Header.Add(ReplacedPathHeader, r.URL.Path)
	r.URL.Path = s.Path
	r.RequestURI = r.URL.RequestURI()
	s.Handler.ServeHTTP(w, r)
}
