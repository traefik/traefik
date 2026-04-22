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
	typeName         = "RewriteTarget"
	xForwardedPrefix = "X-Forwarded-Prefix"
)

// This regex is used to remove the capture groups when no regex is provided,
// as the replacement string may contain capture group references like $1, $2, etc.
// that should be ignored in this case.
var replacementRegex = regexp.MustCompile(`\$[0-9]+`)

// rewriteTarget is a middleware used to replace the path of a URL request.
type rewriteTarget struct {
	name             string
	next             http.Handler
	regexp           *regexp.Regexp
	replacement      string
	xForwardedPrefix string
	// absoluteURLRedirect is true when the replacement template is an absolute URL,
	// indicating the operator explicitly configured a redirect to an external destination.
	absoluteURLRedirect bool
}

// New creates a new rewrite target middleware.
func New(ctx context.Context, next http.Handler, config dynamic.RewriteTarget, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	if config.Replacement == "" {
		return nil, errors.New("replacement cannot be empty")
	}

	var absoluteURLRedirect bool
	if parsed, err := url.Parse(config.Replacement); err == nil && parsed.Scheme != "" {
		absoluteURLRedirect = true
	}

	var re *regexp.Regexp
	if config.Regex != "" {
		var err error
		re, err = regexp.Compile("(?i)" + strings.TrimSpace(config.Regex)) // regex on ingress-nginx are case-insensitive.
		if err != nil {
			return nil, fmt.Errorf("compiling regular expression %s: %w", config.Regex, err)
		}

		// When rewrite-target is a full URL and there's no capture group,
		// append .* to match the entire path and avoid leaking unmatched suffix.
		// See https://github.com/traefik/traefik/issues/12931
		if re.NumSubexp() == 0 && absoluteURLRedirect {
			re, err = regexp.Compile("(?i)" + strings.TrimSpace(config.Regex) + ".*")
			if err != nil {
				return nil, fmt.Errorf("compiling regular expression for absolute URL %s: %w", config.Regex, err)
			}
		}
	}

	return &rewriteTarget{
		name:                name,
		next:                next,
		regexp:              re,
		replacement:         strings.TrimSpace(config.Replacement),
		xForwardedPrefix:    config.XForwardedPrefix,
		absoluteURLRedirect: absoluteURLRedirect,
	}, nil
}

func (rt *rewriteTarget) GetTracingInformation() (string, string) {
	return rt.name, typeName
}

func (rt *rewriteTarget) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), rt.name, typeName)

	currentPath := req.URL.RawPath
	if currentPath == "" {
		currentPath = req.URL.EscapedPath()
	}

	// regexp doex not match, nothing to rewrite.
	if rt.regexp != nil && !rt.regexp.MatchString(currentPath) {
		rt.next.ServeHTTP(rw, req)
		return
	}

	var newTarget string
	if rt.regexp != nil {
		newTarget = rt.regexp.ReplaceAllString(currentPath, rt.replacement)
	} else {
		newTarget = replacementRegex.ReplaceAllString(rt.replacement, "")
	}

	// Same as ingress-nginx when the new target is empty we reply with an error.
	if newTarget == "" {
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Only issue a 302 redirect if the replacement template itself is an absolute URL.
	// Prevent user-controlled capture group content from injecting an absolute URL redirect.
	if rt.absoluteURLRedirect {
		newTargetURL, err := url.Parse(newTarget)
		if err != nil {
			http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		// Carry the incoming query string through the redirect so an
		// absolute-URL rewrite behaves like ingress-nginx, which preserves
		// query parameters on rewrite-target redirects. Only fill it in
		// when the rewrite itself didn't supply one, matching the nginx
		// precedence of rewrite-set query over request query.
		if newTargetURL.RawQuery == "" && req.URL.RawQuery != "" {
			newTargetURL.RawQuery = req.URL.RawQuery
		}

		http.Redirect(rw, req, newTargetURL.String(), http.StatusFound)
		return
	}

	if rt.xForwardedPrefix != "" {
		prefix := rt.xForwardedPrefix
		if rt.regexp != nil {
			prefix = rt.regexp.ReplaceAllString(currentPath, rt.xForwardedPrefix)
		}
		req.Header.Set(xForwardedPrefix, prefix)
	}

	// As replacement can introduce escaped characters
	// Path must remain an unescaped version of RawPath
	// Doesn't handle multiple times encoded replacement (`/` => `%2F` => `%252F` => ...)
	req.URL.RawPath = newTarget

	var err error
	req.URL.Path, err = url.PathUnescape(req.URL.RawPath)
	if err != nil {
		logger.Error().Msgf("Unable to unescape URL RawPath %q: %v", req.URL.RawPath, err)
		observability.SetStatusErrorf(req.Context(), "Unable to unescape URL RawPath %q: %v", req.URL.RawPath, err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	req.RequestURI = req.URL.RequestURI()

	rt.next.ServeHTTP(rw, req)
}
