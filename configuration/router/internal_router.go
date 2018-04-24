package router

import (
	"github.com/containous/mux"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/metrics"
	"github.com/containous/traefik/middlewares"
	mauth "github.com/containous/traefik/middlewares/auth"
	"github.com/containous/traefik/types"
	"github.com/urfave/negroni"
)

// NewInternalRouterAggregator Create a new internalRouterAggregator
func NewInternalRouterAggregator(globalConfiguration configuration.GlobalConfiguration, entryPointName string) *InternalRouterAggregator {
	var serverMiddlewares []negroni.Handler

	if globalConfiguration.EntryPoints[entryPointName].WhiteList != nil {
		ipWhitelistMiddleware, err := middlewares.NewIPWhiteLister(
			globalConfiguration.EntryPoints[entryPointName].WhiteList.SourceRange,
			globalConfiguration.EntryPoints[entryPointName].WhiteList.UseXForwardedFor)
		if err != nil {
			log.Fatalf("Error creating whitelist middleware: %s", err)
		}
		if ipWhitelistMiddleware != nil {
			serverMiddlewares = append(serverMiddlewares, ipWhitelistMiddleware)
		}
	}

	if globalConfiguration.EntryPoints[entryPointName].Auth != nil {
		authMiddleware, err := mauth.NewAuthenticator(globalConfiguration.EntryPoints[entryPointName].Auth, nil)
		if err != nil {
			log.Fatalf("Error creating authenticator middleware: %s", err)
		}
		serverMiddlewares = append(serverMiddlewares, authMiddleware)
	}

	router := InternalRouterAggregator{}
	routerWithPrefix := InternalRouterAggregator{}
	routerWithPrefixAndMiddleware := InternalRouterAggregator{}

	if globalConfiguration.Metrics != nil && globalConfiguration.Metrics.Prometheus != nil && globalConfiguration.Metrics.Prometheus.EntryPoint == entryPointName {
		routerWithPrefixAndMiddleware.AddRouter(metrics.PrometheusHandler{})
	}

	if globalConfiguration.Rest != nil && globalConfiguration.Rest.EntryPoint == entryPointName {
		routerWithPrefixAndMiddleware.AddRouter(globalConfiguration.Rest)
	}

	if globalConfiguration.API != nil && globalConfiguration.API.EntryPoint == entryPointName {
		routerWithPrefixAndMiddleware.AddRouter(globalConfiguration.API)
	}

	if globalConfiguration.Ping != nil && globalConfiguration.Ping.EntryPoint == entryPointName {
		routerWithPrefix.AddRouter(globalConfiguration.Ping)
	}

	if globalConfiguration.ACME != nil && globalConfiguration.ACME.HTTPChallenge != nil && globalConfiguration.ACME.HTTPChallenge.EntryPoint == entryPointName {
		router.AddRouter(globalConfiguration.ACME)
	}

	realRouterWithMiddleware := WithMiddleware{router: &routerWithPrefixAndMiddleware, routerMiddlewares: serverMiddlewares}
	if globalConfiguration.Web != nil && globalConfiguration.Web.Path != "" {
		router.AddRouter(&WithPrefix{PathPrefix: globalConfiguration.Web.Path, Router: &routerWithPrefix})
		router.AddRouter(&WithPrefix{PathPrefix: globalConfiguration.Web.Path, Router: &realRouterWithMiddleware})
	} else {
		router.AddRouter(&routerWithPrefix)
		router.AddRouter(&realRouterWithMiddleware)
	}

	return &router
}

// WithMiddleware router with internal middleware
type WithMiddleware struct {
	router            types.InternalRouter
	routerMiddlewares []negroni.Handler
}

// AddRoutes Add routes to the router
func (wm *WithMiddleware) AddRoutes(systemRouter *mux.Router) {
	realRouter := systemRouter.PathPrefix("/").Subrouter()

	wm.router.AddRoutes(realRouter)

	if len(wm.routerMiddlewares) > 0 {
		realRouter.Walk(wrapRoute(wm.routerMiddlewares))
	}
}

// WithPrefix router which add a prefix
type WithPrefix struct {
	Router     types.InternalRouter
	PathPrefix string
}

// AddRoutes Add routes to the router
func (wp *WithPrefix) AddRoutes(systemRouter *mux.Router) {
	realRouter := systemRouter.PathPrefix("/").Subrouter()
	if wp.PathPrefix != "" {
		realRouter = systemRouter.PathPrefix(wp.PathPrefix).Subrouter()
		realRouter.StrictSlash(true)
		realRouter.SkipClean(true)
	}
	wp.Router.AddRoutes(realRouter)
}

// InternalRouterAggregator InternalRouter that aggregate other internalRouter
type InternalRouterAggregator struct {
	internalRouters []types.InternalRouter
}

// AddRouter add a router in the aggregator
func (r *InternalRouterAggregator) AddRouter(router types.InternalRouter) {
	r.internalRouters = append(r.internalRouters, router)
}

// AddRoutes Add routes to the router
func (r *InternalRouterAggregator) AddRoutes(systemRouter *mux.Router) {
	for _, router := range r.internalRouters {
		router.AddRoutes(systemRouter)
	}
}

// wrapRoute with middlewares
func wrapRoute(middlewares []negroni.Handler) func(*mux.Route, *mux.Router, []*mux.Route) error {
	return func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		middles := append(middlewares, negroni.Wrap(route.GetHandler()))
		route.Handler(negroni.New(middles...))
		return nil
	}
}
