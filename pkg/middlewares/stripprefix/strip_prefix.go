package stripprefix

import (
	"context"
	"net/http"
	"strings"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

const (
	// ForwardedPrefixHeader is the default header to set prefix.
	ForwardedPrefixHeader = "X-Forwarded-Prefix"
	typeName              = "StripPrefix"
)

// stripPrefix is a middleware used to strip prefix from an URL request.
type stripPrefix struct {
	next       http.Handler
	prefixes   []string
	forceSlash bool // TODO Must be removed (breaking), the default behavior must be forceSlash=false
	name       string
}

// New creates a new strip prefix middleware.
func New(ctx context.Context, next http.Handler, config dynamic.StripPrefix, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")
	return &stripPrefix{
		prefixes:   config.Prefixes,
		forceSlash: config.ForceSlash,
		next:       next,
		name:       name,
	}, nil
}

func (s *stripPrefix) GetTracingInformation() (string, ext.SpanKindEnum) {
	return s.name, tracing.SpanKindNoneEnum
}

func (s *stripPrefix) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, prefix := range s.prefixes {
		if strings.HasPrefix(req.URL.Path, prefix) {
			req.URL.Path = s.getPrefixStripped(req.URL.Path, prefix)
			if req.URL.RawPath != "" {
				req.URL.RawPath = s.getPrefixStripped(req.URL.RawPath, prefix)
			}
			s.serveRequest(rw, req, strings.TrimSpace(prefix))
			return
		}
	}
	s.next.ServeHTTP(rw, req)
}

func (s *stripPrefix) serveRequest(rw http.ResponseWriter, req *http.Request, prefix string) {
	req.Header.Add(ForwardedPrefixHeader, prefix)
	req.RequestURI = req.URL.RequestURI()
	s.next.ServeHTTP(rw, req)
}

func (s *stripPrefix) getPrefixStripped(urlPath, prefix string) string {
	if s.forceSlash {
		// Only for compatibility reason with the previous behavior,
		// but the previous behavior is wrong.
		// This needs to be removed in the next breaking version.
		return "/" + strings.TrimPrefix(strings.TrimPrefix(urlPath, prefix), "/")
	}

	return ensureLeadingSlash(strings.TrimPrefix(urlPath, prefix))
}

func ensureLeadingSlash(str string) string {
	if str == "" {
		return str
	}

	if str[0] == '/' {
		return str
	}

	return "/" + str
}
