package chain

import (
	"context"
	"net/http"

	"github.com/containous/alice"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const (
	typeName = "Chain"
)

type chainBuilder interface {
	BuildChain(ctx context.Context, middlewares []string) *alice.Chain
}

// New creates a chain middleware.
func New(ctx context.Context, next http.Handler, config dynamic.Chain, builder chainBuilder, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	middlewareChain := builder.BuildChain(ctx, config.Middlewares)
	return middlewareChain.Then(next)
}
