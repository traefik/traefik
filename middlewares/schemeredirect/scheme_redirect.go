package schemeredirect

import (
	"context"
	"net/http"

	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/middlewares/baseredirect"
	"github.com/pkg/errors"

	"github.com/containous/traefik/config"
)

const (
	typeName            = "SchemeRedirect"
	schemeRedirectRegex = `^(https?:\/\/)?([\w\._-]+)(:\d+)?(.*)$`
)

// New creates a new schemeredirect middleware.
func New(ctx context.Context, next http.Handler, conf config.SchemeRedirect, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug("Creating middleware")
	logger.Debugf("Setting up redirection to %s %s", conf.Scheme, conf.Port)

	if len(conf.Scheme) == 0 {
		return nil, errors.New("you must provide a target scheme")
	}

	port := ""
	if (len(conf.Port) > 0) && !((conf.Scheme == "http" && conf.Port == "80") || (conf.Scheme == "https" && conf.Port == "443")) {
		port = ":" + conf.Port
	}

	c := config.BaseRedirect{
		Regex:       schemeRedirectRegex,
		Replacement: conf.Scheme + "://${2}" + port + "${4}",
		Permanent:   conf.Permanent,
	}

	return baseredirect.New(ctx, next, c, name)
}
