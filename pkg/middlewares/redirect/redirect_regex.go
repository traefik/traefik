package redirect

import (
	"context"
	"net/http"

	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
)

const (
	typeRegexName = "RedirectRegex"
)

// NewRedirectRegex creates a redirect middleware.
func NewRedirectRegex(ctx context.Context, next http.Handler, conf dynamic.RedirectRegex, name string) (http.Handler, error) {
	logger := log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeRegexName))
	logger.Debug("Creating middleware")
	logger.Debugf("Setting up redirection from %s to %s", conf.Regex, conf.Replacement)

	return newRedirect(next, conf.Regex, conf.Replacement, conf.Permanent, name)
}
