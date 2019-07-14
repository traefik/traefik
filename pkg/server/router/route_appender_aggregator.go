package router

import (
	"context"

	"github.com/containous/alice"
	"github.com/containous/mux"
	"github.com/containous/traefik/pkg/api"
	"github.com/containous/traefik/pkg/config/runtime"
	"github.com/containous/traefik/pkg/config/static"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/metrics"
	"github.com/containous/traefik/pkg/types"
)

// chainBuilder The contract of the middleware builder
type chainBuilder interface {
	BuildChain(ctx context.Context, middlewares []string) *alice.Chain
}

// NewRouteAppenderAggregator Creates a new RouteAppenderAggregator
func NewRouteAppenderAggregator(ctx context.Context, chainBuilder chainBuilder, conf static.Configuration,
	entryPointName string, runtimeConfiguration *runtime.Configuration) *RouteAppenderAggregator {
	aggregator := &RouteAppenderAggregator{}

	if conf.Providers != nil && conf.Providers.Rest != nil {
		aggregator.AddAppender(conf.Providers.Rest)
	}

	if conf.API != nil && conf.API.EntryPoint == entryPointName {
		chain := chainBuilder.BuildChain(ctx, conf.API.Middlewares)
		aggregator.AddAppender(&WithMiddleware{
			appender:          api.New(conf, runtimeConfiguration),
			routerMiddlewares: chain,
		})
	}

	if conf.Ping != nil && conf.Ping.EntryPoint == entryPointName {
		chain := chainBuilder.BuildChain(ctx, conf.Ping.Middlewares)
		aggregator.AddAppender(&WithMiddleware{
			appender:          conf.Ping,
			routerMiddlewares: chain,
		})
	}

	if conf.Metrics != nil && conf.Metrics.Prometheus != nil && conf.Metrics.Prometheus.EntryPoint == entryPointName {
		chain := chainBuilder.BuildChain(ctx, conf.Metrics.Prometheus.Middlewares)
		aggregator.AddAppender(&WithMiddleware{
			appender:          metrics.PrometheusHandler{},
			routerMiddlewares: chain,
		})
	}

	return aggregator
}

// RouteAppenderAggregator RouteAppender that aggregate other RouteAppender
type RouteAppenderAggregator struct {
	appenders []types.RouteAppender
}

// Append Adds routes to the router
func (r *RouteAppenderAggregator) Append(systemRouter *mux.Router) {
	for _, router := range r.appenders {
		router.Append(systemRouter)
	}
}

// AddAppender adds a router in the aggregator
func (r *RouteAppenderAggregator) AddAppender(router types.RouteAppender) {
	r.appenders = append(r.appenders, router)
}

// WithMiddleware router with internal middleware
type WithMiddleware struct {
	appender          types.RouteAppender
	routerMiddlewares *alice.Chain
}

// Append Adds routes to the router
func (wm *WithMiddleware) Append(systemRouter *mux.Router) {
	realRouter := systemRouter.PathPrefix("/").Subrouter()

	wm.appender.Append(realRouter)

	if err := realRouter.Walk(wrapRoute(wm.routerMiddlewares)); err != nil {
		log.WithoutContext().Error(err)
	}
}

// wrapRoute with middlewares
func wrapRoute(middlewares *alice.Chain) func(*mux.Route, *mux.Router, []*mux.Route) error {
	return func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		handler, err := middlewares.Then(route.GetHandler())
		if err != nil {
			return err
		}

		route.Handler(handler)
		return nil
	}
}
