package stripprefix

import (
	"context"
	"net/http"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const (
	// ForwardedPrefixHeader is the default header to set prefix.
	ForwardedPrefixHeader = "X-Forwarded-Prefix"
	typeName              = "StripPrefix"
)

// stripPrefix is a middleware used to strip prefix from an URL request.
type stripPrefix struct {
	next     http.Handler
	prefixes []string
	name     string

	// Deprecated: Must be removed (breaking), the default behavior must be forceSlash=false
	forceSlash bool
}

// New creates a new strip prefix middleware.
func New(ctx context.Context, next http.Handler, config dynamic.StripPrefix, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	if config.ForceSlash != nil {
		logger.Warn().Msgf("`ForceSlash` option is deprecated, please remove any usage of this option.")
	}
	// Handle default value (here because of deprecation and the removal of setDefault).
	forceSlash := config.ForceSlash != nil && *config.ForceSlash

	return &stripPrefix{
		prefixes:   config.Prefixes,
		next:       next,
		name:       name,
		forceSlash: forceSlash,
	}, nil
}

func (s *stripPrefix) GetTracingInformation() (string, string) {
	return s.name, typeName
}

func (s *stripPrefix) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	for _, prefix := range s.prefixes {
		if strings.HasPrefix(req.URL.Path, prefix) {
			req.URL.Path = s.getPrefixStripped(req.URL.Path, prefix)
			if req.URL.RawPath != "" {
				req.URL.RawPath = s.getPrefixStripped(req.URL.RawPath, prefix)
			}
			s.serveRequest(rw, req, strings.TrimSpace(prefix))
			return
		}
	}
	s.next.ServeHTTP(rw, req)
}

func (s *stripPrefix) serveRequest(rw http.ResponseWriter, req *http.Request, prefix string) {
	req.Header.Add(ForwardedPrefixHeader, prefix)
	req.RequestURI = req.URL.RequestURI()
	s.next.ServeHTTP(rw, req)
}

func (s *stripPrefix) getPrefixStripped(urlPath, prefix string) string {
	if s.forceSlash {
		// Only for compatibility reason with the previous behavior,
		// but the previous behavior is wrong.
		// This needs to be removed in the next breaking version.
		return "/" + strings.TrimPrefix(strings.TrimPrefix(urlPath, prefix), "/")
	}

	return ensureLeadingSlash(strings.TrimPrefix(urlPath, prefix))
}

func ensureLeadingSlash(str string) string {
	if str == "" {
		return str
	}

	if str[0] == '/' {
		return str
	}

	return "/" + str
}
