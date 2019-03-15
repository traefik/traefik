package chain

import (
	"context"
	"net/http"

	"github.com/containous/alice"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/middlewares"
)

const (
	typeName = "Chain"
)

type chainBuilder interface {
	BuildChain(ctx context.Context, middlewares []string) *alice.Chain
}

// New creates a chain middleware
func New(ctx context.Context, next http.Handler, config config.Chain, builder chainBuilder, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug("Creating middleware")

	middlewareChain := builder.BuildChain(ctx, config.Middlewares)
	return middlewareChain.Then(next)
}
