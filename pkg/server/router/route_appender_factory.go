package router

import (
	"context"

	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/containous/traefik/v2/pkg/config/static"
	"github.com/containous/traefik/v2/pkg/provider/acme"
	"github.com/containous/traefik/v2/pkg/server/middleware"
	"github.com/containous/traefik/v2/pkg/types"
)

// NewRouteAppenderFactory Creates a new RouteAppenderFactory
func NewRouteAppenderFactory(staticConfiguration static.Configuration, entryPointName string, acmeProvider []*acme.Provider) *RouteAppenderFactory {
	return &RouteAppenderFactory{
		staticConfiguration: staticConfiguration,
		entryPointName:      entryPointName,
		acmeProvider:        acmeProvider,
	}
}

// RouteAppenderFactory A factory of RouteAppender
type RouteAppenderFactory struct {
	staticConfiguration static.Configuration
	entryPointName      string
	acmeProvider        []*acme.Provider
}

// NewAppender Creates a new RouteAppender
func (r *RouteAppenderFactory) NewAppender(ctx context.Context, middlewaresBuilder *middleware.Builder, runtimeConfiguration *runtime.Configuration) types.RouteAppender {
	aggregator := NewRouteAppenderAggregator(ctx, middlewaresBuilder, r.staticConfiguration, r.entryPointName, runtimeConfiguration)

	for _, p := range r.acmeProvider {
		if p != nil && p.HTTPChallenge != nil && p.HTTPChallenge.EntryPoint == r.entryPointName {
			aggregator.AddAppender(p)
			break
		}
	}

	return aggregator
}
