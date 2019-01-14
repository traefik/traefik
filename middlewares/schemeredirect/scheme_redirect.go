package schemeredirect

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/containous/traefik/config"
	"github.com/containous/traefik/middlewares/redirect"
)

const (
	schemeRedirectRegex = `^(https?:\/\/)?([\w\._-]+)(:\d+)?(.*)$`
)

// Creates a new schemeredirect middleware
func New(ctx context.Context, next http.Handler, conf config.SchemeRedirect, name string) (http.Handler, error) {
	if len(conf.Scheme) == 0 {
		return nil, errors.New("you must provide a target scheme")
	}

	port := ""
	if (len(conf.Port) > 0) && !((conf.Scheme == "http" && conf.Port == "80") || (conf.Scheme == "https" && conf.Port == "443")) {
		port = ":" + conf.Port
	}

	c := config.Redirect{
		Regex:       schemeRedirectRegex,
		Replacement: conf.Scheme + "://${2}" + port + "${4}",
		Permanent:   conf.Permanent,
	}

	return redirect.New(ctx, next, c, name)
}
