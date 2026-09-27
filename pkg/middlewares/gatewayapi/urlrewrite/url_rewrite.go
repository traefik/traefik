package urlrewrite

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const (
	typeName = "URLRewrite"
)

type urlRewrite struct {
	name string
	next http.Handler

	hostname   *string
	path       *string
	pathPrefix *string
}

// NewURLRewrite creates a URL rewrite middleware.
func NewURLRewrite(ctx context.Context, next http.Handler, conf dynamic.URLRewrite, name string) http.Handler {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	return urlRewrite{
		name:       name,
		next:       next,
		hostname:   conf.Hostname,
		path:       conf.Path,
		pathPrefix: conf.PathPrefix,
	}
}

func (u urlRewrite) GetTracingInformation() (string, string) {
	return u.name, typeName
}

func (u urlRewrite) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), u.name, typeName)

	if u.path != nil && u.pathPrefix == nil {
		req.URL.Path = *u.path
		req.URL.RawPath = ""
	}
	if u.path != nil && u.pathPrefix != nil {
		rawPath := req.URL.EscapedPath()
		// Per the Gateway API spec, a trailing slash on the prefix match value is
		// ignored, so "/foo/" and "/foo" must be stripped identically.
		tail := strings.TrimPrefix(rawPath, strings.TrimSuffix(*u.pathPrefix, "/"))

		newURL := (&url.URL{Path: *u.path}).JoinPath(tail)

		// JoinPath returns an empty path when both the replacement and the tail
		// are empty, but the Gateway API spec requires the root path in that case.
		if newURL.Path == "" {
			newURL.Path = "/"
		}

		// Stop here if the normalization of the path produces a different path.
		// This should be a no-op, as the prefix and the tail are joined and cleaned segment-wise above,
		// leaving nothing left to reinterpret differently here. Kept as a defense-in-depth guard.
		path := newURL.Path
		newURL = newURL.JoinPath()
		if path != newURL.Path {
			logger.Debug().Msgf("Rejecting request, sanitized path: %q is not equivalent to rewritten path: %q", newURL.Path, path)
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		req.URL.Path = newURL.Path
		req.URL.RawPath = newURL.RawPath
	}

	req.RequestURI = req.URL.RequestURI()

	if u.hostname != nil {
		req.Host = *u.hostname
	}

	u.next.ServeHTTP(rw, req)
}
