package redirect

import (
	"context"
	"net/http"

	"github.com/containous/traefik/middlewares"
	"github.com/pkg/errors"

	"github.com/containous/traefik/config"
)

const (
	typeSchemeName      = "RedirectScheme"
	schemeRedirectRegex = `^(https?:\/\/)?([\w\._-]+)(:\d+)?(.*)$`
)

// NewRedirectScheme creates a new RedirectScheme middleware.
func NewRedirectScheme(ctx context.Context, next http.Handler, conf config.RedirectScheme, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeSchemeName)
	logger.Debug("Creating middleware")
	logger.Debugf("Setting up redirection to %s %s", conf.Scheme, conf.Port)

	if len(conf.Scheme) == 0 {
		return nil, errors.New("you must provide a target scheme")
	}

	port := ""
	if (len(conf.Port) > 0) && !((conf.Scheme == "http" && conf.Port == "80") || (conf.Scheme == "https" && conf.Port == "443")) {
		port = ":" + conf.Port
	}

	return newRedirect(ctx, next, schemeRedirectRegex, conf.Scheme+"://${2}"+port+"${4}", conf.Permanent, name)
}
