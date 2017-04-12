package middlewares

import (
	"net/http"
	"strings"
)

// AddPrefix is a middleware used to add a prefix to an URL request.
type AddPrefix struct {
	Handler http.Handler
	Prefix  string
}

func (s *AddPrefix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = s.Prefix + r.URL.Path
	r.RequestURI = r.URL.RequestURI()
	s.Handler.ServeHTTP(w, r)
}

// SetHandler sets handler
func (s *AddPrefix) SetHandler(Handler http.Handler) {
	s.Handler = Handler
}

// StripPrefix is a middleware used to strip a prefix from an URL request.
type StripPrefix struct {
	Handler  http.Handler
	Prefixes []string
}

func (s *StripPrefix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, prefix := range s.Prefixes {
		if p := strings.TrimPrefix(r.URL.Path, strings.TrimSpace(prefix)); len(p) < len(r.URL.Path) {
			r.URL.Path = p
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
