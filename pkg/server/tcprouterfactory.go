package server

import (
	"context"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/containous/traefik/v2/pkg/config/static"
	"github.com/containous/traefik/v2/pkg/responsemodifiers"
	"github.com/containous/traefik/v2/pkg/server/middleware"
	"github.com/containous/traefik/v2/pkg/server/router"
	routertcp "github.com/containous/traefik/v2/pkg/server/router/tcp"
	"github.com/containous/traefik/v2/pkg/server/service"
	"github.com/containous/traefik/v2/pkg/server/service/tcp"
	tcpCore "github.com/containous/traefik/v2/pkg/tcp"
	"github.com/containous/traefik/v2/pkg/tls"
)

// TCPRouterFactory the factory of TCP routers.
type TCPRouterFactory struct {
	entryPoints []string

	managerFactory *service.ManagerFactory

	chainBuilder *middleware.ChainBuilder
	tlsManager   *tls.Manager
}

// NewTCPRouterFactory creates a new TCPRouterFactory
func NewTCPRouterFactory(staticConfiguration static.Configuration, managerFactory *service.ManagerFactory, tlsManager *tls.Manager, chainBuilder *middleware.ChainBuilder) *TCPRouterFactory {
	var entryPoints []string
	for name := range staticConfiguration.EntryPoints {
		entryPoints = append(entryPoints, name)
	}

	return &TCPRouterFactory{
		entryPoints:    entryPoints,
		managerFactory: managerFactory,
		tlsManager:     tlsManager,
		chainBuilder:   chainBuilder,
	}
}

// CreateTCPRouters creates new TCPRouters
func (f *TCPRouterFactory) CreateTCPRouters(conf dynamic.Configuration) map[string]*tcpCore.Router {
	ctx := context.Background()

	rtConf := runtime.NewConfig(conf)

	// HTTP
	serviceManager := f.managerFactory.Build(rtConf)

	middlewaresBuilder := middleware.NewBuilder(rtConf.Middlewares, serviceManager)
	responseModifierFactory := responsemodifiers.NewBuilder(rtConf.Middlewares)

	routerManager := router.NewManager(rtConf, serviceManager, middlewaresBuilder, responseModifierFactory, f.chainBuilder)

	handlersNonTLS := routerManager.BuildHandlers(ctx, f.entryPoints, false)
	handlersTLS := routerManager.BuildHandlers(ctx, f.entryPoints, true)

	// TCP
	svcTCPManager := tcp.NewManager(rtConf)

	rtTCPManager := routertcp.NewManager(rtConf, svcTCPManager, handlersNonTLS, handlersTLS, f.tlsManager)
	routersTCP := rtTCPManager.BuildHandlers(ctx, f.entryPoints)

	rtConf.PopulateUsedBy()

	return routersTCP
}
