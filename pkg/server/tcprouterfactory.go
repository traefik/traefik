package server

import (
	"context"
	"net/http"

	"github.com/containous/traefik/v2/pkg/api"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/containous/traefik/v2/pkg/config/static"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/metrics"
	"github.com/containous/traefik/v2/pkg/responsemodifiers"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/containous/traefik/v2/pkg/server/middleware"
	"github.com/containous/traefik/v2/pkg/server/router"
	routertcp "github.com/containous/traefik/v2/pkg/server/router/tcp"
	"github.com/containous/traefik/v2/pkg/server/service"
	"github.com/containous/traefik/v2/pkg/server/service/tcp"
	tcpCore "github.com/containous/traefik/v2/pkg/tcp"
	"github.com/containous/traefik/v2/pkg/tls"
	"github.com/gorilla/mux"
)

// TCPRouterFactory the factory of TCP routers.
type TCPRouterFactory struct {
	routeAppenderFactories map[string]RouteAppenderFactory

	chainBuilder    *ChainBuilder
	metricsRegistry metrics.Registry
	tlsManager      *tls.Manager

	defaultRoundTripper http.RoundTripper

	api         func(configuration *runtime.Configuration) http.Handler
	restHandler http.Handler

	routinesPool *safe.Pool
}

// NewTCPRouterFactory creates a new TCPRouterFactory
func NewTCPRouterFactory(staticConfiguration static.Configuration, routinesPool *safe.Pool, routeAppenderFactories map[string]RouteAppenderFactory, tlsManager *tls.Manager, metricsRegistry metrics.Registry, chainBuilder *ChainBuilder) *TCPRouterFactory {
	factory := &TCPRouterFactory{
		routeAppenderFactories: routeAppenderFactories,
		tlsManager:             tlsManager,
		metricsRegistry:        metricsRegistry,
		chainBuilder:           chainBuilder,
		defaultRoundTripper:    setupDefaultRoundTripper(staticConfiguration.ServersTransport),
		routinesPool:           routinesPool,
	}

	if staticConfiguration.API != nil {
		factory.api = api.NewBuilder(staticConfiguration)
	}

	if staticConfiguration.Providers != nil && staticConfiguration.Providers.Rest != nil {
		factory.restHandler = staticConfiguration.Providers.Rest.Handler()
	}

	return factory
}

// CreateTCPRouters creates new TCPRouters
func (f *TCPRouterFactory) CreateTCPRouters(conf dynamic.Configuration) map[string]*tcpCore.Router {
	ctx := context.Background()

	var entryPoints []string
	for entryPointName := range f.routeAppenderFactories {
		entryPoints = append(entryPoints, entryPointName)
	}

	rtConf := runtime.NewConfig(conf)

	handlersNonTLS, handlersTLS := f.createHTTPHandlers(ctx, rtConf, entryPoints)

	serviceManager := tcp.NewManager(rtConf)

	routerManager := routertcp.NewManager(rtConf, serviceManager, handlersNonTLS, handlersTLS, f.tlsManager)
	routersTCP := routerManager.BuildHandlers(ctx, entryPoints)

	rtConf.PopulateUsedBy()

	return routersTCP
}

// createHTTPHandlers returns, for the given configuration and entryPoints, the HTTP handlers for non-TLS connections, and for the TLS ones.
// The given configuration must not be nil. its fields will get mutated.
func (f *TCPRouterFactory) createHTTPHandlers(ctx context.Context, configuration *runtime.Configuration, entryPoints []string) (map[string]http.Handler, map[string]http.Handler) {
	svcManager := service.NewManager(configuration.Services, f.defaultRoundTripper, f.metricsRegistry, f.routinesPool)
	serviceManager := service.NewInternalHandlers(f.api, configuration, f.restHandler, svcManager)

	middlewaresBuilder := middleware.NewBuilder(configuration.Middlewares, serviceManager)
	responseModifierFactory := responsemodifiers.NewBuilder(configuration.Middlewares)

	routerManager := router.NewManager(configuration, serviceManager, middlewaresBuilder, responseModifierFactory)

	handlersNonTLS := routerManager.BuildHandlers(ctx, entryPoints, false)
	handlersTLS := routerManager.BuildHandlers(ctx, entryPoints, true)

	routerHandlers := make(map[string]http.Handler)
	for _, entryPointName := range entryPoints {
		internalMuxRouter := mux.NewRouter().SkipClean(true)

		ctx = log.With(ctx, log.Str(log.EntryPointName, entryPointName))

		factory := f.routeAppenderFactories[entryPointName]
		if factory != nil {
			appender := factory.NewAppender(ctx, configuration)
			appender.Append(internalMuxRouter)
		}

		if h, ok := handlersNonTLS[entryPointName]; ok {
			internalMuxRouter.NotFoundHandler = h
		} else {
			internalMuxRouter.NotFoundHandler = buildDefaultHTTPRouter()
		}

		routerHandlers[entryPointName] = internalMuxRouter

		chain := f.chainBuilder.Build(ctx, entryPointName)

		handler, err := chain.Then(internalMuxRouter.NotFoundHandler)
		if err != nil {
			log.FromContext(ctx).Error(err)
			continue
		}
		internalMuxRouter.NotFoundHandler = handler

		handlerTLS, ok := handlersTLS[entryPointName]
		if ok && handlerTLS != nil {
			handlerTLSWithMiddlewares, err := chain.Then(handlerTLS)
			if err != nil {
				log.FromContext(ctx).Error(err)
				continue
			}
			handlersTLS[entryPointName] = handlerTLSWithMiddlewares
		}
	}

	return routerHandlers, handlersTLS
}

func buildDefaultHTTPRouter() *mux.Router {
	rt := mux.NewRouter()
	rt.NotFoundHandler = http.HandlerFunc(http.NotFound)
	rt.SkipClean(true)
	return rt
}

func setupDefaultRoundTripper(conf *static.ServersTransport) http.RoundTripper {
	transport, err := createHTTPTransport(conf)
	if err != nil {
		log.WithoutContext().Errorf("Could not configure HTTP Transport, fallbacking on default transport: %v", err)
		return http.DefaultTransport
	}

	return transport
}
