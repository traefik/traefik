package upstreamvhost

import (
	"context"
	"errors"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/ingressnginx"
)

const typeName = "UpstreamVhost"

type upstreamVhost struct {
	next  http.Handler
	vhost string
	vars  map[string]string
	name  string
}

// New creates a new upstream-vhost middleware that rewrites req.Host from a
// template, resolving NGINX variables at request time.
func New(ctx context.Context, next http.Handler, config dynamic.UpstreamVhost, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	if config.Vhost == "" {
		return nil, errors.New("vhost cannot be empty")
	}

	return &upstreamVhost{
		next:  next,
		vhost: config.Vhost,
		vars:  config.Vars,
		name:  name,
	}, nil
}

func (u *upstreamVhost) GetTracingInformation() (string, string) {
	return u.name, typeName
}

func (u *upstreamVhost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.Host = ingressnginx.ReplaceVariables(u.vhost, req, nil, u.vars)
	u.next.ServeHTTP(rw, req)
}
