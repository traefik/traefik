package stripprefix

import (
	"context"
	"net/http"
	"strings"

	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/middlewares"
	"github.com/containous/traefik/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

type key string

const (
	// ForwardedPrefixHeader is the default header to set prefix.
	ForwardedPrefixHeader = "X-Forwarded-Prefix"
	typeName              = "StripPrefix"
	// StripPrefixKey is the context key for storing the unmodified URL request path
	StripPrefixKey key = "StripPrefix"
)

// stripPrefix is a middleware used to strip prefix from an URL request.
type stripPrefix struct {
	next     http.Handler
	prefixes []string
	name     string
}

// New creates a new strip prefix middleware.
func New(ctx context.Context, next http.Handler, config config.StripPrefix, name string) (http.Handler, error) {
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
			req = req.WithContext(context.WithValue(req.Context(), StripPrefixKey, req.URL.Path))
			req.URL.Path = getPrefixStripped(req.URL.Path, prefix)
			if req.URL.RawPath != "" {
				req.URL.RawPath = getPrefixStripped(req.URL.RawPath, prefix)
			}
			s.serveRequest(rw, req, strings.TrimSpace(prefix))
			return
		}
	}
	http.NotFound(rw, req)
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
