package redirect

import (
	"context"
	"net/http"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const typeRegexName = "RedirectRegex"

// NewRedirectRegex creates a redirect middleware.
func NewRedirectRegex(ctx context.Context, next http.Handler, conf dynamic.RedirectRegex, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeRegexName)
	logger.Debug().Msg("Creating middleware")
	logger.Debug().Msgf("Setting up redirection from %s to %s", conf.Regex, conf.Replacement)

	return newRedirect(next, conf.Regex, conf.Replacement, conf.Permanent, rawURL, name)
}

func rawURL(req *http.Request) string {
	scheme := schemeHTTP
	host := req.Host
	port := ""
	uri := req.RequestURI

	if match := uriRegexp.FindStringSubmatch(req.RequestURI); len(match) > 0 {
		scheme = match[1]

		if len(match[2]) > 0 {
			host = match[2]
		}

		if len(match[3]) > 0 {
			port = match[3]
		}

		uri = match[4]
	}

	if req.TLS != nil {
		scheme = schemeHTTPS
	}

	return strings.Join([]string{scheme, "://", host, port, uri}, "")
}
