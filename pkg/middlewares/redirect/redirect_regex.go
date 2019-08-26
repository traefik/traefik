package redirect

import (
	"context"
	"net/http"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/middlewares"
)

const (
	typeRegexName = "RedirectRegex"
)

// NewRedirectRegex creates a redirect middleware.
func NewRedirectRegex(ctx context.Context, next http.Handler, conf dynamic.RedirectRegex, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeRegexName)
	logger.Debug("Creating middleware")
	logger.Debugf("Setting up redirection from %s to %s", conf.Regex, conf.Replacement)

	return newRedirect(ctx, next, conf.Regex, conf.Replacement, conf.Permanent, name)
}
