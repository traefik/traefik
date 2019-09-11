package stripprefix

import (
	"context"
	"net/http"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	// ForwardedPrefixHeader is the default header to set prefix.
	ForwardedPrefixHeader = "X-Forwarded-Prefix"
	typeName              = "StripPrefix"
)

// stripPrefix is a middleware used to strip prefix from an URL request.
type stripPrefix struct {
	next     http.Handler
	prefixes []string
	name     string
}

// New creates a new strip prefix middleware.
func New(ctx context.Context, next http.Handler, config dynamic.StripPrefix, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug("Creating middleware")
	return &stripPrefix{
		prefixes: config.Prefixes,
		next:     next,
		name:     name,
	}, nil
}

func (s *stripPrefix) GetTracingInformation() (string, ext.SpanKindEnum) {
	return s.name, tracing.SpanKindNoneEnum
}

func (s *stripPrefix) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, prefix := range s.prefixes {
		if strings.HasPrefix(req.URL.Path, prefix) {
			req.URL.Path = getPrefixStripped(req.URL.Path, prefix)
			if req.URL.RawPath != "" {
				req.URL.RawPath = getPrefixStripped(req.URL.RawPath, prefix)
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

func getPrefixStripped(s, prefix string) string {
	return ensureLeadingSlash(strings.TrimPrefix(s, prefix))
}

func ensureLeadingSlash(str string) string {
	return "/" + strings.TrimPrefix(str, "/")
}
