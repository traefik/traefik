package redirect

import (
	"context"
	"fmt"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const typeTrailingSlashName = "RedirectTrailingSlash"

func NewRedirectTrailingSlash(ctx context.Context, next http.Handler, conf dynamic.RedirectTrailingSlash, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeTrailingSlashName)
	logger.Debug().Msg("Creating middleware")

	regex, replacement, err := buildTrailingSlashRegex(conf.Mode)
	if err != nil {
		return nil, err
	}

	return newRedirect(next, regex, replacement, conf.Permanent, nil, rawURL, name)
}

func buildTrailingSlashRegex(mode dynamic.TrailingSlashMode) (regex, replacement string, err error) {
	switch mode {
	case dynamic.TrailingSlashAdd:
		return `^(https?://[^/]+/[^?]*[^/])(\?.*)?$`, "${1}/${2}", nil
	case dynamic.TrailingSlashRemove:
		return `^(https?://[^/]+/[^?]*[^/])/(\?.*)?$`, "${1}${2}", nil
	default:
		return "", "", fmt.Errorf("invalid mode: %s, must be 'add' or 'remove'", mode)
	}
}
