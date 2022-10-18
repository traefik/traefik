package grpcweb

import (
	"context"
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares"
)

const typeName = "grpc-web"

// New builds a new gRPC web request converter.
func New(ctx context.Context, next http.Handler, config dynamic.GrpcWeb, name string) http.Handler {
	log.FromContext(middlewares.GetLoggerCtx(ctx, name, typeName)).Debug("Creating middleware")

	return grpcweb.WrapHandler(next, grpcweb.WithCorsForRegisteredEndpointsOnly(false), grpcweb.WithOriginFunc(func(origin string) bool {
		for _, originCfg := range config.AllowOrigins {
			if originCfg == "*" || originCfg == origin {
				return true
			}
		}
		return false
	}))
}
