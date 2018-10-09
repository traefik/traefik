package router

import (
	"context"

	"github.com/containous/traefik/config/static"
	"github.com/containous/traefik/provider/acme"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/server/middleware"
	"github.com/containous/traefik/types"
)

// NewRouteAppenderFactory Creates a new RouteAppenderFactory
func NewRouteAppenderFactory(staticConfiguration static.Configuration, entryPointName string, acmeProvider *acme.Provider) *RouteAppenderFactory {
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
	acmeProvider        *acme.Provider
}

// NewAppender Creates a new RouteAppender
func (r *RouteAppenderFactory) NewAppender(ctx context.Context, middlewaresBuilder *middleware.Builder, currentConfiguration *safe.Safe) types.RouteAppender {
	aggregator := NewRouteAppenderAggregator(ctx, middlewaresBuilder, r.staticConfiguration, r.entryPointName, currentConfiguration)

	if r.acmeProvider != nil && r.acmeProvider.HTTPChallenge != nil && r.acmeProvider.HTTPChallenge.EntryPoint == r.entryPointName {
		aggregator.AddAppender(r.acmeProvider)
	}

	return aggregator
}
