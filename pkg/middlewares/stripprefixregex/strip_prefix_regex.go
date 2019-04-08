package stripprefixregex

import (
	"context"
	"net/http"
	"strings"

	"github.com/containous/mux"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/middlewares"
	"github.com/containous/traefik/pkg/middlewares/stripprefix"
	"github.com/containous/traefik/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	typeName = "StripPrefixRegex"
)

// StripPrefixRegex is a middleware used to strip prefix from an URL request.
type stripPrefixRegex struct {
	next   http.Handler
	router *mux.Router
	name   string
}

// New builds a new StripPrefixRegex middleware.
func New(ctx context.Context, next http.Handler, config config.StripPrefixRegex, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug("Creating middleware")

	stripPrefix := stripPrefixRegex{
		next:   next,
		router: mux.NewRouter(),
		name:   name,
	}

	for _, prefix := range config.Regex {
		stripPrefix.router.PathPrefix(prefix)
	}

	return &stripPrefix, nil
}

func (s *stripPrefixRegex) GetTracingInformation() (string, ext.SpanKindEnum) {
	return s.name, tracing.SpanKindNoneEnum
}

func (s *stripPrefixRegex) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var match mux.RouteMatch
	if s.router.Match(req, &match) {
		params := make([]string, 0, len(match.Vars)*2)
		for key, val := range match.Vars {
			params = append(params, key)
			params = append(params, val)
		}

		prefix, err := match.Route.URL(params...)
		if err != nil || len(prefix.Path) > len(req.URL.Path) {
			logger := middlewares.GetLogger(req.Context(), s.name, typeName)
			logger.Error("Error in stripPrefix middleware", err)
			return
		}
		req = req.WithContext(context.WithValue(req.Context(), stripprefix.TypeName, req.URL.Path))
		req.URL.Path = req.URL.Path[len(prefix.Path):]
		if req.URL.RawPath != "" {
			req.URL.RawPath = req.URL.RawPath[len(prefix.Path):]
		}
		req.Header.Add(stripprefix.ForwardedPrefixHeader, prefix.Path)
		req.RequestURI = ensureLeadingSlash(req.URL.RequestURI())

		s.next.ServeHTTP(rw, req)
		return
	}
	http.NotFound(rw, req)
}

func ensureLeadingSlash(str string) string {
	return "/" + strings.TrimPrefix(str, "/")
}
