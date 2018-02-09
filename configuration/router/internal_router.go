package router

import (
	"github.com/containous/mux"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/metrics"
	"github.com/containous/traefik/middlewares"
	mauth "github.com/containous/traefik/middlewares/auth"
	"github.com/containous/traefik/types"
	"github.com/urfave/negroni"
)

func NewInternalRouterAggregator(globalConfiguration configuration.GlobalConfiguration, entryPointName string) *InternalRouterAggregator {
	var serverMiddlewares []negroni.Handler

	if globalConfiguration.EntryPoints[entryPointName].WhiteList != nil {
		ipWhitelistMiddleware, _ := middlewares.NewIPWhiteLister(
			globalConfiguration.EntryPoints[entryPointName].WhiteList.SourceRange,
			globalConfiguration.EntryPoints[entryPointName].WhiteList.UseXForwardedFor)
		if ipWhitelistMiddleware != nil {
			serverMiddlewares = append(serverMiddlewares, ipWhitelistMiddleware)
		}
	}

	if globalConfiguration.EntryPoints[entryPointName].Auth != nil {
		authMiddleware, _ := mauth.NewAuthenticator(globalConfiguration.EntryPoints[entryPointName].Auth, nil)
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

	realRouterWithMiddleware := RouterWithMiddleware{router: &routerWithPrefixAndMiddleware, routerMiddlewares: serverMiddlewares}
	if globalConfiguration.Web != nil && globalConfiguration.Web.Path != "" {
		router.AddRouter(&RouterWithPrefix{PathPrefix: globalConfiguration.Web.Path, Router: &routerWithPrefix})
		router.AddRouter(&RouterWithPrefix{PathPrefix: globalConfiguration.Web.Path, Router: &realRouterWithMiddleware})
	} else {
		router.AddRouter(&routerWithPrefix)
		router.AddRouter(&realRouterWithMiddleware)
	}

	return &router
}

type RouterWithMiddleware struct {
	router            types.InternalRouter
	routerMiddlewares []negroni.Handler
}

func (r *RouterWithMiddleware) AddRoutes(systemRouter *mux.Router) {
	realRouter := systemRouter.PathPrefix("/").Subrouter()

	r.router.AddRoutes(realRouter)

	if len(r.routerMiddlewares) > 0 {
		realRouter.Walk(wrapRoute(r.routerMiddlewares))
	}
}

type RouterWithPrefix struct {
	Router     types.InternalRouter
	PathPrefix string
}

func (r *RouterWithPrefix) AddRoutes(systemRouter *mux.Router) {
	realRouter := systemRouter.PathPrefix("/").Subrouter()
	if r.PathPrefix != "" {
		realRouter = systemRouter.PathPrefix(r.PathPrefix).Subrouter()
		realRouter.StrictSlash(true)
		realRouter.SkipClean(true)
	}
	r.Router.AddRoutes(realRouter)
}

type InternalRouterAggregator struct {
	internalRouters []types.InternalRouter
}

func (r *InternalRouterAggregator) AddRouter(router types.InternalRouter) {
	r.internalRouters = append(r.internalRouters, router)
}

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
