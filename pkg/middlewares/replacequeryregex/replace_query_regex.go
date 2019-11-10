package replacequeryregex

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	typeName = "ReplaceQueryRegex"
)

// ReplaceQueryRegex is a middleware used to replace the query of a URL request with a regular expression
type replaceQueryRegex struct {
	next        http.Handler
	regexp      *regexp.Regexp
	replacement string
	name        string
}

// New creates a new replace query regex middleware.
func New(ctx context.Context, next http.Handler, config dynamic.ReplaceQueryRegex, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

	exp, err := regexp.Compile(strings.TrimSpace(config.Regex))
	if err != nil {
		return nil, fmt.Errorf("error compiling regular expression %s: %s", config.Regex, err)
	}

	return &replaceQueryRegex{
		regexp:      exp,
		replacement: strings.TrimSpace(config.Replacement),
		next:        next,
		name:        name,
	}, nil
}

func (r *replaceQueryRegex) GetTracingInformation() (string, ext.SpanKindEnum) {
	return r.name, tracing.SpanKindNoneEnum
}

func (r *replaceQueryRegex) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if r.regexp == nil || !r.regexp.MatchString(req.URL.RawQuery) {
		r.next.ServeHTTP(rw, req)
		return
	}

	req.URL.RawQuery = r.regexp.ReplaceAllString(req.URL.RawQuery, r.replacement)
	req.RequestURI = req.URL.RequestURI()

	r.next.ServeHTTP(rw, req)
}
