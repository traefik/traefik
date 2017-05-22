package middlewares

import (
	"net/http"
	"strings"
)

const (
	forwardedPrefixHeader = "X-Forwarded-Prefix"
)

// StripPrefix is a middleware used to strip prefix from an URL request
type StripPrefix struct {
	Handler  http.Handler
	Prefixes []string
}

func (s *StripPrefix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, prefix := range s.Prefixes {
		origPrefix := strings.TrimSpace(prefix)
		if origPrefix == r.URL.Path {
			r.URL.Path = "/"
			s.serveRequest(w, r, origPrefix)
			return
		}

		prefix = strings.TrimSuffix(origPrefix, "/") + "/"
		if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
			r.URL.Path = "/" + strings.TrimPrefix(p, "/")
			s.serveRequest(w, r, origPrefix)
			return
		}
	}
	http.NotFound(w, r)
}

func (s *StripPrefix) serveRequest(w http.ResponseWriter, r *http.Request, prefix string) {
	r.Header[forwardedPrefixHeader] = []string{prefix}
	r.RequestURI = r.URL.RequestURI()
	s.Handler.ServeHTTP(w, r)
}

// SetHandler sets handler
func (s *StripPrefix) SetHandler(Handler http.Handler) {
	s.Handler = Handler
}
