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

// requestMirror sends a copy of the requests to a mirror service, and ignores its responses.
type requestMirror struct {
	http.Handler

	name string
}

// NewRequestMirror creates a request mirror middleware.
func NewRequestMirror(ctx context.Context, next http.Handler, conf dynamic.RequestMirror, serviceBuilder serviceBuilder, pool *safe.Pool, name string) (http.Handler, error) {
	logger := middlewares.GetLogger(ctx, name, typeName)
	logger.Debug().Msg("Creating middleware")

	mirrorHandler, err := serviceBuilder.BuildHTTP(ctx, conf.Service)
	if err != nil {
		return nil, fmt.Errorf("building mirror service %q: %w", conf.Service, err)
	}

	// The body is streamed rather than buffered, as the Gateway API does not allow to bound the buffered body size,
	// and the mirrored request is cloned before calling next, as the filters declared after this one must not alter it.
	handler := mirror.New(next, pool, true, -1, true, nil)
	if err := handler.AddMirror(mirrorHandler, conf.Numerator, conf.Denominator); err != nil {
		return nil, fmt.Errorf("adding mirror service %q: %w", conf.Service, err)
	}

	return &requestMirror{Handler: handler, name: name}, nil
}

func (r *requestMirror) GetTracingInformation() (string, string) {
	return r.name, typeName
}
