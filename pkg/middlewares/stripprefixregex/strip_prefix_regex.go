package stripprefixregex

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/middlewares/stripprefix"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	typeName = "StripPrefixRegex"
)

// StripPrefixRegex is a middleware used to strip prefix from an URL request.
type stripPrefixRegex struct {
	next        http.Handler
	expressions []*regexp.Regexp
	name        string
}

// New builds a new StripPrefixRegex middleware.
func New(ctx context.Context, next http.Handler, config dynamic.StripPrefixRegex, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug("Creating middleware")

	stripPrefix := stripPrefixRegex{
		next: next,
		name: name,
	}

	for _, exp := range config.Regex {
		reg, err := regexp.Compile(strings.TrimSpace(exp))
		if err != nil {
			return nil, err
		}
		stripPrefix.expressions = append(stripPrefix.expressions, reg)
	}

	return &stripPrefix, nil
}

func (s *stripPrefixRegex) GetTracingInformation() (string, ext.SpanKindEnum) {
	return s.name, tracing.SpanKindNoneEnum
}

func (s *stripPrefixRegex) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, exp := range s.expressions {
		parts := exp.FindStringSubmatch(req.URL.Path)
		if len(parts) > 0 && len(parts[0]) > 0 {
			prefix := parts[0]
			if !strings.HasPrefix(req.URL.Path, prefix) {
				continue
			}

			req.Header.Add(stripprefix.ForwardedPrefixHeader, prefix)

			req.URL.Path = strings.Replace(req.URL.Path, prefix, "", 1)
			if req.URL.RawPath != "" {
				req.URL.RawPath = req.URL.RawPath[len(prefix):]
			}

			req.RequestURI = ensureLeadingSlash(req.URL.RequestURI())
			s.next.ServeHTTP(rw, req)
			return
		}
	}

	s.next.ServeHTTP(rw, req)
}

func ensureLeadingSlash(str string) string {
	return "/" + strings.TrimPrefix(str, "/")
}
