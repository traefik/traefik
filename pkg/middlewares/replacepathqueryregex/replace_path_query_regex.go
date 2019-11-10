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

func (r *replacePathQueryRegex) GetTracingInformation() (string, ext.SpanKindEnum) {
	return r.name, tracing.SpanKindNoneEnum
}

func (r *replacePathQueryRegex) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if r.regexp == nil || !r.regexp.MatchString(req.RequestURI) {
		r.next.ServeHTTP(rw, req)
		return
	}

	replacement := r.regexp.ReplaceAllString(req.RequestURI, r.replacement)
	path := strings.SplitN(req.RequestURI, "?", 2)[0]
	if replacement != "" {
		path = path + "?" + replacement
	}

	u, err := req.URL.Parse(path)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	req.URL = u
	req.RequestURI = u.RequestURI()

	r.next.ServeHTTP(rw, req)
}
