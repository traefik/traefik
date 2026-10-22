package requestmirror

import (
	"context"
	"fmt"
	"net/http"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
	"github.com/traefik/traefik/v3/pkg/safe"
	"github.com/traefik/traefik/v3/pkg/server/service/loadbalancer/mirror"
)

const typeName = "RequestMirror"

type serviceBuilder interface {
	BuildHTTP(ctx context.Context, serviceName string) (http.Handler, error)
}

type requestMirror struct {
	name    string
	handler http.Handler
}

// NewRequestMirror creates a request mirror middleware.
func NewRequestMirror(ctx context.Context, next http.Handler, conf dynamic.RequestMirror, serviceBuilder serviceBuilder, pool *safe.Pool, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	mirrorHandler, err := serviceBuilder.BuildHTTP(ctx, conf.Service)
	if err != nil {
		return nil, fmt.Errorf("building mirror service %q: %w", conf.Service, err)
	}

	// Mirroring sends a distinct clone of the request to next and to the mirror,
	// so the filters declared after this one, as per the Gateway API filter ordering,
	// do not alter the mirrored request.
	m := mirror.New(next, pool, true, dynamic.MirroringDefaultMaxBodySize, nil)
	if err := m.AddMirror(mirrorHandler, conf.Percent); err != nil {
		return nil, err
	}

	return &requestMirror{
		name:    name,
		handler: m,
	}, nil
}

func (r *requestMirror) GetTracingInformation() (string, string) {
	return r.name, typeName
}

func (r *requestMirror) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(rw, req)
}
