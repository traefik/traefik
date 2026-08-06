package urlrewrite

import (
	"context"
	"net/http"
	"net/url"
	"path"
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

	newPath := req.URL.Path
	if u.path != nil && u.pathPrefix == nil {
		newPath = *u.path
	}
	if u.path != nil && u.pathPrefix != nil {
		// Per the Gateway API spec, a trailing slash on the prefix match value is
		// ignored, so "/foo/" and "/foo" must be stripped identically.
		tail := strings.TrimPrefix(req.URL.Path, strings.TrimSuffix(*u.pathPrefix, "/"))

		// Here we are sanitizing the tail kept after trimming the prefix,
		// as path.Join below silently resolves any ".." or "." segments it contains.
		sanitizedTail := tail
		if sanitizedTail != "" {
			sanitizedTail = (&url.URL{Path: sanitizedTail}).JoinPath().Path
		}

		// Stop here if the normalization of the tail produces a different path,
		// as it would let the rewritten path escape the configured replacement path.
		if tail != sanitizedTail {
			logger.Debug().Msgf("Rejecting request, sanitized path: %q is not equivalent to trimmed path: %q", sanitizedTail, tail)
			http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		newPath = path.Join(*u.path, tail)
		// path.Join returns an empty string when both the replacement and the tail
		// are empty, but the Gateway API spec requires the root path in that case.
		if newPath == "" {
			newPath = "/"
		}

		// add the trailing slash if needed, as path.Join removes trailing slashes.
		if strings.HasSuffix(req.URL.Path, "/") && !strings.HasSuffix(newPath, "/") {
			newPath += "/"
		}
	}

	req.URL.Path = newPath
	req.URL.RawPath = req.URL.EscapedPath()
	req.RequestURI = req.URL.RequestURI()

	if u.hostname != nil {
		req.Host = *u.hostname
	}

	u.next.ServeHTTP(rw, req)
}
