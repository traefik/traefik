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

func (rq *replaceQueryRegex) GetTracingInformation() (string, ext.SpanKindEnum) {
	return rq.name, tracing.SpanKindNoneEnum
}

func (rq *replaceQueryRegex) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	splitURI := strings.SplitN(req.RequestURI, "?", 2)
	if len(splitURI) < 2 {
		rq.next.ServeHTTP(rw, req)
		return
	}

	rawPath := splitURI[0]
	rawQuery := splitURI[1]

	if rq.regexp != nil && rq.regexp.MatchString(rawQuery) {
		newQuery := rq.regexp.ReplaceAllString(rawQuery, rq.replacement)
		path := rawPath
		if newQuery != "" {
			path = path + "?" + newQuery
		}
		if u, err := req.URL.Parse(path); err == nil {
			req.URL = u
			req.RequestURI = u.RequestURI()
		}
	}
	rq.next.ServeHTTP(rw, req)
}
