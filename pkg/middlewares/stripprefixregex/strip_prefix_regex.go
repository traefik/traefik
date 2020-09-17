package stripprefixregex

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/middlewares/stripprefix"
	"github.com/traefik/traefik/v2/pkg/tracing"
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
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

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

			req.URL.Path = ensureLeadingSlash(strings.Replace(req.URL.Path, prefix, "", 1))
			if req.URL.RawPath != "" {
				req.URL.RawPath = ensureLeadingSlash(req.URL.RawPath[len(prefix):])
			}

			req.RequestURI = req.URL.RequestURI()
			s.next.ServeHTTP(rw, req)
			return
		}
	}

	s.next.ServeHTTP(rw, req)
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
