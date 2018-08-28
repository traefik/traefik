package middlewares

import (
	"context"
	"net/http"

	"github.com/containous/mux"
	"github.com/containous/traefik/log"
)

// StripPrefixRegex is a middleware used to strip prefix from an URL request
type StripPrefixRegex struct {
	Handler http.Handler
	router  *mux.Router
}

// NewStripPrefixRegex builds a new StripPrefixRegex given a handler and prefixes
func NewStripPrefixRegex(handler http.Handler, prefixes []string) *StripPrefixRegex {
	stripPrefix := StripPrefixRegex{Handler: handler, router: mux.NewRouter()}

	for _, prefix := range prefixes {
		stripPrefix.router.PathPrefix(prefix)
	}

	return &stripPrefix
}

func (s *StripPrefixRegex) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var match mux.RouteMatch
	if s.router.Match(r, &match) {
		params := make([]string, 0, len(match.Vars)*2)
		for key, val := range match.Vars {
			params = append(params, key)
			params = append(params, val)
		}

		prefix, err := match.Route.URL(params...)
		if err != nil || len(prefix.Path) > len(r.URL.Path) {
			log.Error("Error in stripPrefix middleware", err)
			return
		}
		rawReqPath := r.URL.Path
		r.URL.Path = r.URL.Path[len(prefix.Path):]
		if r.URL.RawPath != "" {
			r.URL.RawPath = r.URL.RawPath[len(prefix.Path):]
		}
		r = r.WithContext(context.WithValue(r.Context(), StripPrefixKey, rawReqPath))
		r.Header.Add(ForwardedPrefixHeader, prefix.Path)
		r.RequestURI = ensureLeadingSlash(r.URL.RequestURI())
		s.Handler.ServeHTTP(w, r)
		return
	}
	http.NotFound(w, r)
}

// SetHandler sets handler
func (s *StripPrefixRegex) SetHandler(Handler http.Handler) {
	s.Handler = Handler
}
