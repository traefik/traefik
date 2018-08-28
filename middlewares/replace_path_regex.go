package middlewares

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/containous/traefik/log"
)

// ReplacePathRegex is a middleware used to replace the path of a URL request with a regular expression
type ReplacePathRegex struct {
	Handler     http.Handler
	Regexp      *regexp.Regexp
	Replacement string
}

// NewReplacePathRegexHandler returns a new ReplacePathRegex
func NewReplacePathRegexHandler(regex string, replacement string, handler http.Handler) http.Handler {
	exp, err := regexp.Compile(strings.TrimSpace(regex))
	if err != nil {
		log.Errorf("Error compiling regular expression %s: %s", regex, err)
	}
	return &ReplacePathRegex{
		Regexp:      exp,
		Replacement: strings.TrimSpace(replacement),
		Handler:     handler,
	}
}

func (s *ReplacePathRegex) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Regexp != nil && len(s.Replacement) > 0 && s.Regexp.MatchString(r.URL.Path) {
		r = r.WithContext(context.WithValue(r.Context(), ReplacePathKey, r.URL.Path))
		r.Header.Add(ReplacedPathHeader, r.URL.Path)
		r.URL.Path = s.Regexp.ReplaceAllString(r.URL.Path, s.Replacement)
		r.RequestURI = r.URL.RequestURI()
	}
	s.Handler.ServeHTTP(w, r)
}
