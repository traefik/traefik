package upstreamvhost

import (
	"context"
	"errors"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/middlewares/ingressnginx"
)

const typeName = "UpstreamVHost"

type upstreamVHost struct {
	name  string
	next  http.Handler
	vHost string
	vars  map[string]string
}

// New creates a new upstream-vhost middleware that rewrites req.Host from a
// template, resolving NGINX variables at request time.
func New(ctx context.Context, next http.Handler, config dynamic.UpstreamVHost, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	if config.VHost == "" {
		return nil, errors.New("vHost cannot be empty")
	}

	return &upstreamVHost{
		name:  name,
		next:  next,
		vHost: config.VHost,
		vars:  config.Vars,
	}, nil
}

func (u *upstreamVHost) GetTracingInformation() (string, string) {
	return u.name, typeName
}

func (u *upstreamVHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Preserve the original client-facing request host before overwriting
	// req.Host with the upstream target. This allows downstream middlewares
	// (e.g. the snippet/forward-auth middleware) to keep resolving $host, $best_http_host
	// and $server_name to the original hostname instead of the upstream-vhost.
	req = req.WithContext(ingressnginx.WithOriginalHost(req.Context(), req.Host))
	req.Host = ingressnginx.ReplaceVariables(u.vHost, req, nil, u.vars)
	u.next.ServeHTTP(rw, req)
}
