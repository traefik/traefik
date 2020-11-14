package redirect

import (
	"context"
	"errors"
	"net/http"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
)

const (
	typeSchemeName      = "RedirectScheme"
	schemeRedirectRegex = `^(https?:\/\/)?(\[[\w:.]+\]|[\w\._-]+)?(:\d+)?(.*)$`
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
	if len(conf.Port) > 0 && !(conf.Scheme == "http" && conf.Port == "80" || conf.Scheme == "https" && conf.Port == "443") {
		port = ":" + conf.Port
	}

	return newRedirect(next, schemeRedirectRegex, conf.Scheme+"://${2}"+port+"${4}", conf.Permanent, name)
}
