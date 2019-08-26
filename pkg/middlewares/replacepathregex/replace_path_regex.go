package replacepathregex

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/middlewares"
	"github.com/containous/traefik/v2/pkg/middlewares/replacepath"
	"github.com/containous/traefik/v2/pkg/tracing"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	typeName = "ReplacePathRegex"
)

// ReplacePathRegex is a middleware used to replace the path of a URL request with a regular expression.
type replacePathRegex struct {
	next        http.Handler
	regexp      *regexp.Regexp
	replacement string
	name        string
}

// New creates a new replace path regex middleware.
func New(ctx context.Context, next http.Handler, config dynamic.ReplacePathRegex, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug("Creating middleware")

	exp, err := regexp.Compile(strings.TrimSpace(config.Regex))
	if err != nil {
		return nil, fmt.Errorf("error compiling regular expression %s: %s", config.Regex, err)
	}

	return &replacePathRegex{
		regexp:      exp,
		replacement: strings.TrimSpace(config.Replacement),
		next:        next,
		name:        name,
	}, nil
}

func (rp *replacePathRegex) GetTracingInformation() (string, ext.SpanKindEnum) {
	return rp.name, tracing.SpanKindNoneEnum
}

func (rp *replacePathRegex) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if rp.regexp != nil && len(rp.replacement) > 0 && rp.regexp.MatchString(req.URL.Path) {
		req.Header.Add(replacepath.ReplacedPathHeader, req.URL.Path)
		req.URL.Path = rp.regexp.ReplaceAllString(req.URL.Path, rp.replacement)
		req.RequestURI = req.URL.RequestURI()
	}
	rp.next.ServeHTTP(rw, req)
}
