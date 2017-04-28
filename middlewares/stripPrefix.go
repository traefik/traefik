package middlewares

import (
	"net/http"
	"strings"

	"github.com/containous/traefik/middlewares/common"
)

// StripPrefix is a middleware used to strip prefix from an URL request
type StripPrefix struct {
	common.BasicMiddleware
	Prefixes []string
}

var _ common.Middleware = &StripPrefix{}

func (s *StripPrefix) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, prefix := range s.Prefixes {
		if p := strings.TrimPrefix(r.URL.Path, strings.TrimSpace(prefix)); len(p) < len(r.URL.Path) {
			r.URL.Path = p
			r.RequestURI = r.URL.RequestURI()
			s.Next().ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}
