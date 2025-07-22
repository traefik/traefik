package server

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/config/static"
	httpmuxer "github.com/traefik/traefik/v3/pkg/muxer/http"
	"github.com/traefik/traefik/v3/pkg/server/middleware"
	tcpmiddleware "github.com/traefik/traefik/v3/pkg/server/middleware/tcp"
	"github.com/traefik/traefik/v3/pkg/server/router"
	tcprouter "github.com/traefik/traefik/v3/pkg/server/router/tcp"
	udprouter "github.com/traefik/traefik/v3/pkg/server/router/udp"
	"github.com/traefik/traefik/v3/pkg/server/service"
	tcpsvc "github.com/traefik/traefik/v3/pkg/server/service/tcp"
	udpsvc "github.com/traefik/traefik/v3/pkg/server/service/udp"
	"github.com/traefik/traefik/v3/pkg/tcp"
	"github.com/traefik/traefik/v3/pkg/tls"
	"github.com/traefik/traefik/v3/pkg/udp"
)

// RouterFactory the factory of TCP/UDP routers.
type RouterFactory struct {
	entryPointsTCP  []string
	entryPointsUDP  []string
	allowACMEByPass map[string]bool

	managerFactory *service.ManagerFactory

	pluginBuilder middleware.PluginsBuilder

	observabilityMgr *middleware.ObservabilityMgr
	tlsManager       *tls.Manager

	dialerManager *tcp.DialerManager

	cancelPrevState func()

	parser httpmuxer.SyntaxParser
}

// NewRouterFactory creates a new RouterFactory.
func NewRouterFactory(staticConfiguration static.Configuration, managerFactory *service.ManagerFactory, tlsManager *tls.Manager,
	observabilityMgr *middleware.ObservabilityMgr, pluginBuilder middleware.PluginsBuilder, dialerManager *tcp.DialerManager,
) (*RouterFactory, error) {
	handlesTLSChallenge := false
	for _, resolver := range staticConfiguration.CertificatesResolvers {
		if resolver.ACME != nil && resolver.ACME.TLSChallenge != nil {
			handlesTLSChallenge = true
			break
		}
	}

	allowACMEByPass := map[string]bool{}
	var entryPointsTCP, entryPointsUDP []string
	for name, ep := range staticConfiguration.EntryPoints {
		allowACMEByPass[name] = ep.AllowACMEByPass || !handlesTLSChallenge

		protocol, err := ep.GetProtocol()
		if err != nil {
			// Should never happen because Traefik should not start if protocol is invalid.
			log.Error().Err(err).Msg("Invalid protocol")
		}

		if protocol == "udp" {
			entryPointsUDP = append(entryPointsUDP, name)
		} else {
			entryPointsTCP = append(entryPointsTCP, name)
		}
	}

	parser, err := httpmuxer.NewSyntaxParser()
	if err != nil {
		return nil, fmt.Errorf("creating parser: %w", err)
	}

	return &RouterFactory{
		entryPointsTCP:   entryPointsTCP,
		entryPointsUDP:   entryPointsUDP,
		managerFactory:   managerFactory,
		observabilityMgr: observabilityMgr,
		tlsManager:       tlsManager,
		pluginBuilder:    pluginBuilder,
		dialerManager:    dialerManager,
		allowACMEByPass:  allowACMEByPass,
		parser:           parser,
	}, nil
}

// CreateRouters creates new TCPRouters and UDPRouters.
func (f *RouterFactory) CreateRouters(rtConf *runtime.Configuration) (map[string]*tcprouter.Router, map[string]udp.Handler) {
	if f.cancelPrevState != nil {
		f.cancelPrevState()
	}

	var ctx context.Context
	ctx, f.cancelPrevState = context.WithCancel(context.Background())

	// HTTP
	serviceManager := f.managerFactory.Build(rtConf)

	middlewaresBuilder := middleware.NewBuilder(rtConf.Middlewares, serviceManager, f.pluginBuilder)

	routerManager := router.NewManager(rtConf, serviceManager, middlewaresBuilder, f.observabilityMgr, f.tlsManager, f.parser)

	handlersNonTLS := routerManager.BuildHandlers(ctx, f.entryPointsTCP, false)
	handlersTLS := routerManager.BuildHandlers(ctx, f.entryPointsTCP, true)

	serviceManager.LaunchHealthCheck(ctx)

	// TCP
	svcTCPManager := tcpsvc.NewManager(rtConf, f.dialerManager)

	middlewaresTCPBuilder := tcpmiddleware.NewBuilder(rtConf.TCPMiddlewares)

	rtTCPManager := tcprouter.NewManager(rtConf, svcTCPManager, middlewaresTCPBuilder, handlersNonTLS, handlersTLS, f.tlsManager)
	routersTCP := rtTCPManager.BuildHandlers(ctx, f.entryPointsTCP)

	for ep, r := range routersTCP {
		if allowACMEByPass, ok := f.allowACMEByPass[ep]; ok && allowACMEByPass {
			r.EnableACMETLSPassthrough()
		}
	}

	// UDP
	svcUDPManager := udpsvc.NewManager(rtConf)
	rtUDPManager := udprouter.NewManager(rtConf, svcUDPManager)
	routersUDP := rtUDPManager.BuildHandlers(ctx, f.entryPointsUDP)

	rtConf.PopulateUsedBy()

	return routersTCP, routersUDP
}
