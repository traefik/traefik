package router

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/containous/alice"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslog"
	"github.com/traefik/traefik/v3/pkg/middlewares/denyrouterrecursion"
	metricsMiddle "github.com/traefik/traefik/v3/pkg/middlewares/metrics"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/middlewares/recovery"
	httpmuxer "github.com/traefik/traefik/v3/pkg/muxer/http"
	"github.com/traefik/traefik/v3/pkg/server/middleware"
	"github.com/traefik/traefik/v3/pkg/server/provider"
	"github.com/traefik/traefik/v3/pkg/tls"
)

const maxUserPriority = math.MaxInt - 1000

type middlewareBuilder interface {
	BuildChain(ctx context.Context, names []string) *alice.Chain
}

type serviceManager interface {
	BuildHTTP(rootCtx context.Context, serviceName string) (http.Handler, error)
	LaunchHealthCheck(ctx context.Context)
}

// Manager A route/router manager.
type Manager struct {
	routerHandlers     map[string]http.Handler
	serviceManager     serviceManager
	observabilityMgr   *middleware.ObservabilityMgr
	middlewaresBuilder middlewareBuilder
	conf               *runtime.Configuration
	tlsManager         *tls.Manager
	parser             httpmuxer.SyntaxParser
}

// NewManager creates a new Manager.
func NewManager(conf *runtime.Configuration, serviceManager serviceManager, middlewaresBuilder middlewareBuilder, observabilityMgr *middleware.ObservabilityMgr, tlsManager *tls.Manager, parser httpmuxer.SyntaxParser) *Manager {
	return &Manager{
		routerHandlers:     make(map[string]http.Handler),
		serviceManager:     serviceManager,
		observabilityMgr:   observabilityMgr,
		middlewaresBuilder: middlewaresBuilder,
		conf:               conf,
		tlsManager:         tlsManager,
		parser:             parser,
	}
}

func (m *Manager) getHTTPRouters(ctx context.Context, entryPoints []string, tls bool) map[string]map[string]*runtime.RouterInfo {
	if m.conf != nil {
		return m.conf.GetRoutersByEntryPoints(ctx, entryPoints, tls)
	}

	return make(map[string]map[string]*runtime.RouterInfo)
}

// BuildHandlers Builds handler for all entry points.
func (m *Manager) BuildHandlers(rootCtx context.Context, entryPoints []string, tls bool) map[string]http.Handler {
	entryPointHandlers := make(map[string]http.Handler)

	for entryPointName, routers := range m.getHTTPRouters(rootCtx, entryPoints, tls) {
		logger := log.Ctx(rootCtx).With().Str(logs.EntryPointName, entryPointName).Logger()
		ctx := logger.WithContext(rootCtx)

		handler, err := m.buildEntryPointHandler(ctx, entryPointName, routers)
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}

		entryPointHandlers[entryPointName] = handler
	}

	// Create default handlers.
	for _, entryPointName := range entryPoints {
		logger := log.Ctx(rootCtx).With().Str(logs.EntryPointName, entryPointName).Logger()
		ctx := logger.WithContext(rootCtx)

		handler, ok := entryPointHandlers[entryPointName]
		if ok || handler != nil {
			continue
		}

		defaultHandler, err := m.observabilityMgr.BuildEPChain(ctx, entryPointName, "", nil).Then(BuildDefaultHTTPRouter())
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}
		entryPointHandlers[entryPointName] = defaultHandler
	}

	return entryPointHandlers
}

func (m *Manager) buildEntryPointHandler(ctx context.Context, entryPointName string, configs map[string]*runtime.RouterInfo) (http.Handler, error) {
	muxer := httpmuxer.NewMuxer(m.parser)

	defaultHandler, err := m.observabilityMgr.BuildEPChain(ctx, entryPointName, "", nil).Then(http.NotFoundHandler())
	if err != nil {
		return nil, err
	}

	muxer.SetDefaultHandler(defaultHandler)

	for routerName, routerConfig := range configs {
		logger := log.Ctx(ctx).With().Str(logs.RouterName, routerName).Logger()
		ctxRouter := logger.WithContext(provider.AddInContext(ctx, routerName))

		if routerConfig.Priority == 0 {
			routerConfig.Priority = httpmuxer.GetRulePriority(routerConfig.Rule)
		}

		if routerConfig.Priority > maxUserPriority && !strings.HasSuffix(routerName, "@internal") {
			err = fmt.Errorf("the router priority %d exceeds the max user-defined priority %d", routerConfig.Priority, maxUserPriority)
			routerConfig.AddError(err, true)
			logger.Error().Err(err).Send()
			continue
		}

		handler, err := m.buildRouterHandler(ctxRouter, routerName, routerConfig)
		if err != nil {
			routerConfig.AddError(err, true)
			logger.Error().Err(err).Send()
			continue
		}

		observabilityChain := m.observabilityMgr.BuildEPChain(ctx, entryPointName, routerConfig.Service, routerConfig.Observability)
		handler, err = observabilityChain.Then(handler)
		if err != nil {
			routerConfig.AddError(err, true)
			logger.Error().Err(err).Send()
			continue
		}

		if err = muxer.AddRoute(routerConfig.Rule, routerConfig.RuleSyntax, routerConfig.Priority, handler); err != nil {
			routerConfig.AddError(err, true)
			logger.Error().Err(err).Send()
			continue
		}
	}

	chain := alice.New()
	chain = chain.Append(func(next http.Handler) (http.Handler, error) {
		return recovery.New(ctx, next)
	})

	return chain.Then(muxer)
}

func (m *Manager) buildRouterHandler(ctx context.Context, routerName string, routerConfig *runtime.RouterInfo) (http.Handler, error) {
	if handler, ok := m.routerHandlers[routerName]; ok {
		return handler, nil
	}

	if routerConfig.TLS != nil {
		// Don't build the router if the TLSOptions configuration is invalid.
		tlsOptionsName := tls.DefaultTLSConfigName
		if len(routerConfig.TLS.Options) > 0 && routerConfig.TLS.Options != tls.DefaultTLSConfigName {
			tlsOptionsName = provider.GetQualifiedName(ctx, routerConfig.TLS.Options)
		}
		if _, err := m.tlsManager.Get(tls.DefaultTLSStoreName, tlsOptionsName); err != nil {
			return nil, fmt.Errorf("building router handler: %w", err)
		}
	}

	handler, err := m.buildHTTPHandler(ctx, routerConfig, routerName)
	if err != nil {
		return nil, err
	}

	// Prevents from enabling observability for internal resources.
	if !m.observabilityMgr.ShouldAddAccessLogs(provider.GetQualifiedName(ctx, routerConfig.Service), routerConfig.Observability) {
		m.routerHandlers[routerName] = handler
		return m.routerHandlers[routerName], nil
	}

	handlerWithAccessLog, err := alice.New(func(next http.Handler) (http.Handler, error) {
		return accesslog.NewFieldHandler(next, accesslog.RouterName, routerName, nil), nil
	}).Then(handler)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Send()
		m.routerHandlers[routerName] = handler
	} else {
		m.routerHandlers[routerName] = handlerWithAccessLog
	}

	return m.routerHandlers[routerName], nil
}

func (m *Manager) buildHTTPHandler(ctx context.Context, router *runtime.RouterInfo, routerName string) (http.Handler, error) {
	var qualifiedNames []string
	for _, name := range router.Middlewares {
		qualifiedNames = append(qualifiedNames, provider.GetQualifiedName(ctx, name))
	}
	router.Middlewares = qualifiedNames

	if router.Service == "" {
		return nil, errors.New("the service is missing on the router")
	}

	sHandler, err := m.serviceManager.BuildHTTP(ctx, router.Service)
	if err != nil {
		return nil, err
	}

	mHandler := m.middlewaresBuilder.BuildChain(ctx, router.Middlewares)

	chain := alice.New()

	if m.observabilityMgr.MetricsRegistry() != nil && m.observabilityMgr.MetricsRegistry().IsRouterEnabled() &&
		m.observabilityMgr.ShouldAddMetrics(provider.GetQualifiedName(ctx, router.Service), router.Observability) {
		chain = chain.Append(metricsMiddle.WrapRouterHandler(ctx, m.observabilityMgr.MetricsRegistry(), routerName, provider.GetQualifiedName(ctx, router.Service)))
	}

	// Prevents from enabling tracing for internal resources.
	if !m.observabilityMgr.ShouldAddTracing(provider.GetQualifiedName(ctx, router.Service), router.Observability) {
		return chain.Extend(*mHandler).Then(sHandler)
	}

	chain = chain.Append(observability.WrapRouterHandler(ctx, routerName, router.Rule, provider.GetQualifiedName(ctx, router.Service)))

	if m.observabilityMgr.MetricsRegistry() != nil && m.observabilityMgr.MetricsRegistry().IsRouterEnabled() {
		metricsHandler := metricsMiddle.WrapRouterHandler(ctx, m.observabilityMgr.MetricsRegistry(), routerName, provider.GetQualifiedName(ctx, router.Service))
		chain = chain.Append(observability.WrapMiddleware(ctx, metricsHandler))
	}

	if router.DefaultRule {
		chain = chain.Append(denyrouterrecursion.WrapHandler(routerName))
	}

	return chain.Extend(*mHandler).Then(sHandler)
}

// BuildDefaultHTTPRouter creates a default HTTP router.
func BuildDefaultHTTPRouter() http.Handler {
	return http.NotFoundHandler()
}
