package redirect

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
)

const (
	typeSchemeName  = "RedirectScheme"
	uriPattern      = `^(https?:\/\/)?(\[[\w:.]+\]|[\w\._-]+)?(:\d+)?(.*)$`
	xForwardedProto = "X-Forwarded-Proto"
)

// NewRedirectScheme creates a new RedirectScheme middleware.
func NewRedirectScheme(ctx context.Context, next http.Handler, conf dynamic.RedirectScheme, name string) (http.Handler, error) {
	logger := log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeSchemeName))
	logger.Debug("Creating middleware")
	logger.Debugf("Setting up redirection to %s %s", conf.Scheme, conf.Port)

	if len(conf.Scheme) == 0 {
		return nil, errors.New("you must provide a target scheme")
	}

	port := ""
	if len(conf.Port) > 0 && !(conf.Scheme == schemeHTTP && conf.Port == "80" || conf.Scheme == schemeHTTPS && conf.Port == "443") {
		port = ":" + conf.Port
	}

	return newRedirect(next, uriPattern, conf.Scheme+"://${2}"+port+"${4}", conf.Permanent, rawURLScheme, name)
}

func rawURLScheme(req *http.Request) string {
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

	if value := req.Header.Get(xForwardedProto); value != "" {
		scheme = value
	}

	if scheme == schemeHTTP && port == ":80" || scheme == schemeHTTPS && port == ":443" {
		port = ""
	}

	return strings.Join([]string{scheme, "://", host, port, uri}, "")
}
