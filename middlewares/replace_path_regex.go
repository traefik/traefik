package middlewares

import (
	"net/http"
	"regexp"
)

// ReplacePathRegex is a middleware used to replace the path of a URL request with a regular expression
type ReplacePathRegex struct {
	Handler http.Handler
	Regexp  *regexp.Regexp
	Repl    string
}

func (s *ReplacePathRegex) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Regexp != nil && len(s.Repl) > 0 {
		r.Header.Add(ReplacedPathHeader, r.URL.Path)
		r.URL.Path = s.Regexp.ReplaceAllString(r.URL.Path, s.Repl)
		r.RequestURI = r.URL.RequestURI()
	}
	s.Handler.ServeHTTP(w, r)
}
