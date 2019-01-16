package redirect

import (
	"context"
	"net/http"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/middlewares/baseredirect"
)

const (
	typeName = "Redirect"
)

// New creates a redirect middleware.
func New(ctx context.Context, next http.Handler, conf config.Redirect, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug("Creating middleware")
	logger.Debugf("Setting up redirection from %s to %s", conf.Regex, conf.Replacement)

	c := config.BaseRedirect{
		Replacement: conf.Replacement,
		Regex:       conf.Regex,
		Permanent:   conf.Permanent,
	}

	return baseredirect.New(ctx, next, c, name)
}
