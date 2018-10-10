package middlewares

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/containous/traefik/log"
)

// ReplaceQueryRegex is a middleware used to replace the query of a URL request with a regular expression
type ReplaceQueryRegex struct {
	Handler     http.Handler
	Regexp      *regexp.Regexp
	Replacement string
}

// NewReplaceQueryRegexHandler returns a new ReplaceQueryRegex
func NewReplaceQueryRegexHandler(regex string, replacement string, handler http.Handler) http.Handler {
	re, err := regexp.Compile(strings.TrimSpace(regex))
	if err != nil {
		log.Errorf("Error compiling regular expression %s: %s", regex, err)
	}
	return &ReplaceQueryRegex{
		Regexp:      re,
		Replacement: strings.TrimSpace(replacement),
		Handler:     handler,
	}
}

func (s *ReplaceQueryRegex) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.Regexp != nil && s.Regexp.MatchString(r.RequestURI) {
		replacement := s.Regexp.ReplaceAllString(r.RequestURI, s.Replacement)
		path := strings.SplitN(r.RequestURI, "?", 2)[0]
		if replacement != "" {
			path = path + "?" + replacement
		}
		if u, err := r.URL.Parse(path); err != nil {
			log.Errorf("bad replacement %s: %s", replacement, err)
		} else {
			r.URL = u
			r.RequestURI = u.RequestURI()
		}
	}
	s.Handler.ServeHTTP(w, r)
}
