package middlewares

import (
	"net/http"
	"regexp"
	"strings"
)

// ReplacePathRegex is a middleware used to replace the path of a URL request with a regular expression
type ReplacePathRegex struct {
	Handler     http.Handler
	Regexp      *regexp.Regexp
	Replacement string
}

// NewReplacePathRegexHandler returns a new instance of ReplacePathRegex
func NewReplacePathRegexHandler(regex string, replacement string, handler http.Handler) http.Handler {
	return &ReplacePathRegex{
		Regexp:      regexp.MustCompile(strings.TrimSpace(regex)),
		Replacement: strings.TrimSpace(replacement),
		Handler:     handler,
	}
}

func (s *ReplacePathRegex) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Regexp != nil && len(s.Replacement) > 0 && s.Regexp.MatchString(r.URL.Path) {
		r.Header.Add(ReplacedPathHeader, r.URL.Path)
		r.URL.Path = s.Regexp.ReplaceAllString(r.URL.Path, s.Replacement)
		r.RequestURI = r.URL.RequestURI()
	}
	s.Handler.ServeHTTP(w, r)
}
