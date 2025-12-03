package grpcweb

import (
	"context"
	"net/http"

	"github.com/baqupio/baqup/v3/pkg/config/dynamic"
	"github.com/baqupio/baqup/v3/pkg/middlewares"
	"github.com/traefik/grpc-web/go/grpcweb"
)

const typeName = "GRPCWeb"

// New builds a new gRPC web request converter.
func New(ctx context.Context, next http.Handler, config dynamic.GrpcWeb, name string) http.Handler {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	return grpcweb.WrapHandler(next, grpcweb.WithCorsForRegisteredEndpointsOnly(false), grpcweb.WithOriginFunc(func(origin string) bool {
		for _, originCfg := range config.AllowOrigins {
			if originCfg == "*" || originCfg == origin {
				return true
			}
		}
		return false
	}))
}
