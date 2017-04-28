package middlewares

import (
	"net/http"
	"regexp"
)

const (
	forwardedPrefixHeader = "X-Forwarded-Prefix"
)

// StripPrefix is a middleware used to strip prefix from an URL request
type StripPrefix struct {
	Handler  http.Handler
	Prefixes []*regexp.Regexp
}

func (s *StripPrefix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	originalPath := r.URL.Path
	for _, prefix := range s.Prefixes {
		index := prefix.FindStringIndex(originalPath)
		if index != nil && index[0]!=index[1] {
			r.URL.Path = originalPath[index[1]:]
			r.Header[forwardedPrefixHeader] = []string{ originalPath[index[0]:index[1]] }
			r.RequestURI = r.URL.RequestURI()
			s.Handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

// SetHandler sets handler
func (s *StripPrefix) SetHandler(Handler http.Handler) {
	s.Handler = Handler
}
