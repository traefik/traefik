package replacepathregex

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/middlewares"
	"github.com/traefik/traefik/v2/pkg/middlewares/replacepath"
	"github.com/traefik/traefik/v2/pkg/tracing"
)

const typeName = "ReplacePathRegex"

// ReplacePathRegex is a middleware used to replace the path of a URL request with a regular expression.
type replacePathRegex struct {
	next        http.Handler
	regexp      *regexp.Regexp
	replacement string
	name        string
}

// New creates a new replace path regex middleware.
func New(ctx context.Context, next http.Handler, config dynamic.ReplacePathRegex, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	exp, err := regexp.Compile(strings.TrimSpace(config.Regex))
	if err != nil {
		return nil, fmt.Errorf("error compiling regular expression %s: %w", config.Regex, err)
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
	currentPath := req.URL.RawPath
	if currentPath == "" {
		currentPath = req.URL.EscapedPath()
	}

	if rp.regexp != nil && len(rp.replacement) > 0 && rp.regexp.MatchString(currentPath) {
		req.Header.Add(replacepath.ReplacedPathHeader, currentPath)
		req.URL.RawPath = rp.regexp.ReplaceAllString(currentPath, rp.replacement)

		// as replacement can introduce escaped characters
		// Path must remain an unescaped version of RawPath
		// Doesn't handle multiple times encoded replacement (`/` => `%2F` => `%252F` => ...)
		var err error
		req.URL.Path, err = url.PathUnescape(req.URL.RawPath)
		if err != nil {
			middlewares.GetLogger(context.Background(), rp.name, typeName).Error().Err(err).Send()
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		req.RequestURI = req.URL.RequestURI()
	}

	rp.next.ServeHTTP(rw, req)
}
