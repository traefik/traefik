package redirect

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const typeName = "RequestRedirect"

type redirect struct {
	name string
	next http.Handler

	scheme     *string
	hostname   *string
	port       *string
	path       *string
	pathPrefix *string
	statusCode int
}

// NewRequestRedirect creates a redirect middleware.
func NewRequestRedirect(ctx context.Context, next http.Handler, conf dynamic.RequestRedirect, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	statusCode := conf.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusFound
	}

	// Comply with HTTPRequestRedirectFilter.StatusCode
	if statusCode != http.StatusMovedPermanently &&
		statusCode != http.StatusFound &&
		statusCode != http.StatusSeeOther &&
		statusCode != http.StatusTemporaryRedirect &&
		statusCode != http.StatusPermanentRedirect {
		return nil, fmt.Errorf("unsupported status code: %d", statusCode)
	}

	return redirect{
		name:       name,
		next:       next,
		scheme:     conf.Scheme,
		hostname:   conf.Hostname,
		port:       conf.Port,
		path:       conf.Path,
		pathPrefix: conf.PathPrefix,
		statusCode: statusCode,
	}, nil
}

func (r redirect) GetTracingInformation() (string, string) {
	return r.name, typeName
}

func (r redirect) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	logger := middlewares.GetLogger(req.Context(), r.name, typeName)

	redirectURL := *req.URL
	redirectURL.Host = req.Host

	// req.URL.Scheme is always empty for server requests (Go net/http, RFC 7230 section 5.3).
	// Per the Gateway API spec, when no scheme is configured the request scheme must be used.
	// https://github.com/kubernetes-sigs/gateway-api/blob/v1.4.0/apis/v1/httproute_types.go#L1194-L1195
	redirectURL.Scheme = "http"
	if r.scheme != nil {
		redirectURL.Scheme = *r.scheme
	} else if req.TLS != nil {
		redirectURL.Scheme = "https"
	}

	host := redirectURL.Hostname()
	if r.hostname != nil {
		host = *r.hostname
	}

	port := redirectURL.Port()
	if r.port != nil {
		port = *r.port
	}

	if port != "" {
		host = net.JoinHostPort(host, port)
	}
	redirectURL.Host = host

	if r.path != nil && r.pathPrefix == nil {
		redirectURL.Path = *r.path
	}

	if r.path != nil && r.pathPrefix != nil {
		rawPath := req.URL.EscapedPath()
		// Per the Gateway API spec, a trailing slash on the prefix match value is
		// ignored, so "/foo/" and "/foo" must be stripped identically.
		tail := strings.TrimPrefix(rawPath, strings.TrimSuffix(*r.pathPrefix, "/"))

		newURL := (&url.URL{Path: *r.path}).JoinPath(tail)

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

		redirectURL.Path = newURL.Path
		redirectURL.RawPath = newURL.RawPath
	}

	rw.Header().Set("Location", redirectURL.String())

	rw.WriteHeader(r.statusCode)
	if _, err := rw.Write([]byte(http.StatusText(r.statusCode))); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}
