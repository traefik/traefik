package middlewares

import (
	"net/http"
	"strings"
)

// StripPrefix is a middleware used to strip prefix from an URL request
type StripPrefix struct {
	Handler http.Handler
	Prefix  string
}

func (s *StripPrefix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p := strings.TrimPrefix(r.URL.Path, s.Prefix); len(p) < len(r.URL.Path) {
		r.URL.Path = p
		r.RequestURI = r.URL.RequestURI()
		s.Handler.ServeHTTP(w, r)
	} else {
		http.NotFound(w, r)
	}
}
