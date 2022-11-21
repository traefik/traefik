package redirect

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/middlewares"
)

const (
	typeSchemeName  = "RedirectScheme"
	uriPattern      = `^(https?:\/\/)?(\[[\w:.]+\]|[\w\._-]+)?(:\d+)?(.*)$`
	xForwardedProto = "X-Forwarded-Proto"
)

// NewRedirectScheme creates a new RedirectScheme middleware.
func NewRedirectScheme(ctx context.Context, next http.Handler, conf dynamic.RedirectScheme, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeSchemeName)
	logger.Debug().Msg("Creating middleware")
	logger.Debug().Msgf("Setting up redirection to %s %s", conf.Scheme, conf.Port)

	if len(conf.Scheme) == 0 {
		return nil, errors.New("you must provide a target scheme")
	}

	port := ""
	if len(conf.Port) > 0 && !(conf.Scheme == schemeHTTP && conf.Port == "80" || conf.Scheme == schemeHTTPS && conf.Port == "443") {
		port = ":" + conf.Port
	}

	return newRedirect(next, uriPattern, conf.Scheme+"://${2}"+port+"${4}", conf.Permanent, clientRequestURL, name)
}

func clientRequestURL(req *http.Request) string {
	scheme := schemeHTTP
	host, port, err := net.SplitHostPort(req.Host)
	if err != nil {
		host = req.Host
	} else {
		port = ":" + port
	}
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

	if xProto := req.Header.Get(xForwardedProto); xProto != "" {
		// When the initial request is a connection upgrade request,
		// X-Forwarded-Proto header might have been set by a previous hop to ws(s),
		// even though the actual protocol used so far is HTTP(s).
		// Given that we're in a middleware that is only used in the context of HTTP(s) requests,
		// the only possible valid schemes are one of "http" or "https", so we convert back to them.
		switch {
		case strings.EqualFold(xProto, "ws"):
			scheme = schemeHTTP
		case strings.EqualFold(xProto, "wss"):
			scheme = schemeHTTPS
		default:
			scheme = xProto
		}
	}

	if scheme == schemeHTTP && port == ":80" || scheme == schemeHTTPS && port == ":443" {
		port = ""
	}

	return strings.Join([]string{scheme, "://", host, port, uri}, "")
}
