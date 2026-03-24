package rewritetarget

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
)

const (
	typeName               = "RewriteTarget"
	xForwardedPrefixHeader = "X-Forwarded-Prefix"
)

// RewriteTarget is a middleware used to replace the path of a URL request.
type rewriteTarget struct {
	next             http.Handler
	regexp           *regexp.Regexp
	replacement      string
	xForwardedPrefix string
	name             string
}

// New creates a new rewrite target middleware.
func New(ctx context.Context, next http.Handler, config dynamic.RewriteTarget, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	if config.Replacement == "" {
		return nil, errors.New("replacement cannot be empty")
	}

	mw := &rewriteTarget{
		next:             next,
		replacement:      strings.TrimSpace(config.Replacement),
		xForwardedPrefix: config.XForwardedPrefix,
		name:             name,
	}

	if config.Regex != "" {
		exp, err := regexp.Compile(strings.TrimSpace(config.Regex))
		if err != nil {
			return nil, fmt.Errorf("compiling regular expression %s: %w", config.Regex, err)
		}
		mw.regexp = exp
	}

	return mw, nil
}

func (rt *rewriteTarget) GetTracingInformation() (string, string) {
	return rt.name, typeName
}

func (rt *rewriteTarget) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	currentPath := req.URL.RawPath
	if currentPath == "" {
		currentPath = req.URL.EscapedPath()
	}

	var newTarget string
	if rt.regexp != nil {
		if !rt.regexp.MatchString(currentPath) {
			rt.next.ServeHTTP(rw, req)
			return
		}
		newTarget = rt.regexp.ReplaceAllString(currentPath, rt.replacement)
	} else {
		newTarget = rt.replacement
	}

	// If the replacement resolves to an absolute URL, issue a 302 redirect.
	if parsed, err := url.Parse(newTarget); err == nil && parsed.Scheme != "" {
		http.Redirect(rw, req, newTarget, http.StatusFound)
		return
	}

	req.URL.RawPath = newTarget

	if rt.xForwardedPrefix != "" {
		prefix := rt.xForwardedPrefix
		if rt.regexp != nil {
			prefix = rt.regexp.ReplaceAllString(currentPath, rt.xForwardedPrefix)
		}
		req.Header.Set(xForwardedPrefixHeader, prefix)
	}

	// as replacement can introduce escaped characters
	// Path must remain an unescaped version of RawPath
	// Doesn't handle multiple times encoded replacement (`/` => `%2F` => `%252F` => ...)
	var err error
	req.URL.Path, err = url.PathUnescape(req.URL.RawPath)
	if err != nil {
		middlewares.GetLogger(context.Background(), rt.name, typeName).Error().Msgf("Unable to unescape url raw path %q: %v", req.URL.RawPath, err)
		observability.SetStatusErrorf(req.Context(), "Unable to unescape url raw path %q: %v", req.URL.RawPath, err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	req.RequestURI = req.URL.RequestURI()

	rt.next.ServeHTTP(rw, req)
}
