package replacepathqueryregex

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
	typeName = "ReplacePathQueryRegex"
)

// ReplacePathQueryRegex is a middleware used to replace the path and query of a URL request with a regular expression
type replacePathQueryRegex struct {
	next        http.Handler
	regexp      *regexp.Regexp
	replacement string
	name        string
}

// New creates a new replace path and query regex middleware.
func New(ctx context.Context, next http.Handler, config dynamic.ReplacePathQueryRegex, name string) (http.Handler, error) {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

	exp, err := regexp.Compile(strings.TrimSpace(config.Regex))
	if err != nil {
		return nil, fmt.Errorf("error compiling regular expression %s: %s", config.Regex, err)
	}

	return &replacePathQueryRegex{
		regexp:      exp,
		replacement: strings.TrimSpace(config.Replacement),
		next:        next,
		name:        name,
	}, nil
}

func (rpq *replacePathQueryRegex) GetTracingInformation() (string, ext.SpanKindEnum) {
	return rpq.name, tracing.SpanKindNoneEnum
}

func (rpq *replacePathQueryRegex) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if rpq.regexp != nil && rpq.regexp.MatchString(req.RequestURI) {
		replacement := rpq.regexp.ReplaceAllString(req.RequestURI, rpq.replacement)
		path := strings.SplitN(req.RequestURI, "?", 2)[0]
		if replacement != "" {
			path = path + "?" + replacement
		}
		if u, err := req.URL.Parse(path); err == nil {
			req.URL = u
			req.RequestURI = u.RequestURI()
		}
	}
	rpq.next.ServeHTTP(rw, req)
}
