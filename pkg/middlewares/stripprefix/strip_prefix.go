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
			req.URL.Path = s.getPathStripped(req.URL.Path, prefix)
			if req.URL.RawPath != "" {
				req.URL.RawPath = s.getRawPathStripped(req.URL.RawPath, prefix)
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

func (s *stripPrefix) getPathStripped(urlPath, prefix string) string {
	if s.forceSlash {
		// Only for compatibility reason with the previous behavior,
		// but the previous behavior is wrong.
		// This needs to be removed in the next breaking version.
		return "/" + strings.TrimPrefix(strings.TrimPrefix(urlPath, prefix), "/")
	}

	return ensureLeadingSlash(strings.TrimPrefix(urlPath, prefix))
}

func (s *stripPrefix) getRawPathStripped(rawPath, prefix string) string {
	if s.forceSlash {
		// Only for compatibility reason with the previous behavior,
		// but the previous behavior is wrong.
		// This needs to be removed in the next breaking version.
		return "/" + strings.TrimPrefix(rawPath[encodedPrefixLen(rawPath, prefix):], "/")
	}

	return ensureLeadingSlash(rawPath[encodedPrefixLen(rawPath, prefix):])
}

// encodedPrefixLen returns the number of bytes in rawPath that correspond to
// the decoded prefix, advancing 3 bytes per %XX sequence and 1 byte otherwise.
func encodedPrefixLen(rawPath, decodedPrefix string) int {
	decoded := 0
	i := 0
	for i < len(rawPath) && decoded < len(decodedPrefix) {
		if rawPath[i] == '%' && i+2 < len(rawPath) {
			i += 3
		} else {
			i++
		}
		decoded++
	}
	return i
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
