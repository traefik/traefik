package router

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"slices"
	"sort"
	"strings"

	"github.com/containous/alice"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/middlewares/accesslog"
	"github.com/traefik/traefik/v3/pkg/middlewares/denyrouterrecursion"
	metricsMiddle "github.com/traefik/traefik/v3/pkg/middlewares/metrics"
	"github.com/traefik/traefik/v3/pkg/middlewares/observability"
	"github.com/traefik/traefik/v3/pkg/middlewares/recovery"
	httpmuxer "github.com/traefik/traefik/v3/pkg/muxer/http"
	"github.com/traefik/traefik/v3/pkg/observability/logs"
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

// Router pre-routing validation and hierarchy setup.
// This function:
// - Sets ChildRefs in runtime.RouterInfo for routers that reference another router as a service
// - Detects and reports errors in runtime.RouterInfo for:
//   - router with a service and referenced as a parent by another router
//   - non-root router with TLS config
//   - non-root router with Observability config
//   - cyclic references between routers

// ComputePreRouting computes information about routers that is needed before building the handlers.
// It uses the runtime.RouterInfo.ParentRefs field to set the runtime.RouterInfo.ChildRefs field.
// It also detects potential errors and adds them to the runtime.RouterInfo.Errors field.
func (m *Manager) ComputePreRouting() {
	if m.conf == nil || m.conf.Routers == nil {
		return
	}

	// First pass: populate ChildRefs based on ParentRefs
	for routerName, router := range m.conf.Routers {
		if router.ParentRefs == nil {
			continue
		}

		for _, parentName := range router.ParentRefs {
			if parentRouter, exists := m.conf.Routers[parentName]; exists {
				// Add this router as a child of its parent
				if !slices.Contains(parentRouter.ChildRefs, routerName) {
					parentRouter.ChildRefs = append(parentRouter.ChildRefs, routerName)
				}
			} else {
				router.AddError(fmt.Errorf("parent router %q does not exist", parentName), true)
			}
		}
	}

	// Second pass: detect cyclic references and other validation errors
	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	// Get router names in sorted order for deterministic behavior
	var routerNames []string
	for routerName := range m.conf.Routers {
		routerNames = append(routerNames, routerName)
	}
	sort.Strings(routerNames)

	for _, routerName := range routerNames {
		if !visited[routerName] {
			m.detectCycles(routerName, visited, inStack, make([]string, 0))
		}
	}

	// Third pass: other validation errors
	for _, router := range m.conf.Routers {
		// Check for router with service that is referenced as a parent
		if router.Service != "" && len(router.ChildRefs) > 0 {
			router.AddError(fmt.Errorf("router has both a service and is referenced as a parent by other routers"), true)
		}

		// Check for non-root router with TLS config
		if len(router.ParentRefs) > 0 && router.TLS != nil {
			router.AddError(fmt.Errorf("non-root router cannot have TLS configuration"), true)
		}

		// Check for non-root router with Observability config
		if len(router.ParentRefs) > 0 && router.Observability != nil {
			router.AddError(fmt.Errorf("non-root router cannot have Observability configuration"), true)
		}
	}
}

// detectCycles detects cyclic references in router hierarchy and marks only the router that closes the cycle with an error.
func (m *Manager) detectCycles(routerName string, visited, inStack map[string]bool, path []string) {
	if inStack[routerName] {
		// Found a cycle - mark only the router that closes the loop with an error
		cycleStart := -1
		for i, name := range path {
			if name == routerName {
				cycleStart = i
				break
			}
		}

		if cycleStart >= 0 {
			cyclePath := append(path[cycleStart:], routerName)
			cycleRouters := strings.Join(cyclePath, " -> ")

			// The router that closes the loop is the one before the repeated router in cyclePath
			if len(cyclePath) >= 2 {
				// The router that closes the cycle is the one before the repeated router
				closingRouter := cyclePath[len(cyclePath)-2]
				if router, exists := m.conf.Routers[closingRouter]; exists {
					router.AddError(fmt.Errorf("cyclic reference detected in router hierarchy: %s", cycleRouters), true)
				}
			}
		}
		return
	}

	if visited[routerName] {
		return
	}

	router, exists := m.conf.Routers[routerName]
	if !exists || router.ParentRefs == nil {
		visited[routerName] = true
		return
	}

	visited[routerName] = true
	inStack[routerName] = true
	newPath := append(path, routerName)

	// Sort ParentRefs for deterministic cycle path generation
	sortedParentRefs := make([]string, len(router.ParentRefs))
	copy(sortedParentRefs, router.ParentRefs)
	sort.Strings(sortedParentRefs)

	for _, parentName := range sortedParentRefs {
		m.detectCycles(parentName, visited, inStack, newPath)
	}

	inStack[routerName] = false
}

//// hasCyclicReference detects cyclic references in router hierarchy using DFS.
//func (m *Manager) hasCyclicReference(routerName string, visited map[string]bool) bool {
//	if visited[routerName] {
//		return true // Found a cycle
//	}
//
//	router, exists := m.conf.Routers[routerName]
//	if !exists || router.ParentRefs == nil {
//		return false
//	}
//
//	visited[routerName] = true
//
//	for _, parentName := range router.ParentRefs {
//		if m.hasCyclicReference(parentName, visited) {
//			return true
//		}
//	}
//
//	delete(visited, routerName) // Backtrack
//	return false
//}

func (m *Manager) getHTTPRouters(ctx context.Context, entryPoints []string, tls bool) map[string]map[string]*runtime.RouterInfo {
	if m.conf != nil {
		return m.conf.GetRoutersByEntryPoints(ctx, entryPoints, tls)
	}

	return make(map[string]map[string]*runtime.RouterInfo)
}

// BuildHandlers Builds handler for all entry points.
func (m *Manager) BuildHandlers(rootCtx context.Context, entryPoints []string, tls bool) map[string]http.Handler {
	entryPointHandlers := make(map[string]http.Handler)

	defaultObsConfig := dynamic.RouterObservabilityConfig{}
	defaultObsConfig.SetDefaults()

	for entryPointName, routers := range m.getHTTPRouters(rootCtx, entryPoints, tls) {
		logger := log.Ctx(rootCtx).With().Str(logs.EntryPointName, entryPointName).Logger()
		ctx := logger.WithContext(rootCtx)

		// TODO: Improve this part. Relying on models is a shortcut to get the entrypoint observability configuration. Maybe we should pass down the static configuration.
		// When the entry point has no observability configuration no model is produced,
		// and we need to create the default configuration is this case.
		epObsConfig := defaultObsConfig
		if model, ok := m.conf.Models[entryPointName+"@internal"]; ok && model != nil {
			epObsConfig = model.Observability
		}

		handler, err := m.buildEntryPointHandler(ctx, entryPointName, routers, epObsConfig)
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

		// TODO: Improve this part. Relying on models is a shortcut to get the entrypoint observability configuration. Maybe we should pass down the static configuration.
		// When the entry point has no observability configuration no model is produced,
		// and we need to create the default configuration is this case.
		epObsConfig := defaultObsConfig
		if model, ok := m.conf.Models[entryPointName+"@internal"]; ok && model != nil {
			epObsConfig = model.Observability
		}

		defaultHandler, err := m.observabilityMgr.BuildEPChain(ctx, entryPointName, false, epObsConfig).Then(http.NotFoundHandler())
		if err != nil {
			logger.Error().Err(err).Send()
			continue
		}
		entryPointHandlers[entryPointName] = defaultHandler
	}

	return entryPointHandlers
}

func (m *Manager) buildEntryPointHandler(ctx context.Context, entryPointName string, configs map[string]*runtime.RouterInfo, config dynamic.RouterObservabilityConfig) (http.Handler, error) {
	muxer := httpmuxer.NewMuxer(m.parser)

	defaultHandler, err := m.observabilityMgr.BuildEPChain(ctx, entryPointName, false, config).Then(http.NotFoundHandler())
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

		// Only build handlers for root routers (routers without ParentRefs).
		// Routers with ParentRefs will be built as part of their parent router's muxer.
		if len(routerConfig.ParentRefs) > 0 {
			continue
		}

		handler, err := m.buildRouterHandler(ctxRouter, routerName, routerConfig)
		if err != nil {
			routerConfig.AddError(err, true)
			logger.Error().Err(err).Send()
			continue
		}

		if routerConfig.Observability != nil {
			config = *routerConfig.Observability
		}

		observabilityChain := m.observabilityMgr.BuildEPChain(ctxRouter, entryPointName, strings.HasSuffix(routerConfig.Service, "@internal"), config)
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

	m.routerHandlers[routerName] = handler
	return m.routerHandlers[routerName], nil
}

func (m *Manager) buildHTTPHandler(ctx context.Context, router *runtime.RouterInfo, routerName string) (http.Handler, error) {
	var qualifiedNames []string
	for _, name := range router.Middlewares {
		qualifiedNames = append(qualifiedNames, provider.GetQualifiedName(ctx, name))
	}
	router.Middlewares = qualifiedNames

	chain := alice.New()

	if router.DefaultRule {
		chain = chain.Append(denyrouterrecursion.WrapHandler(routerName))
	}

	mHandler := m.middlewaresBuilder.BuildChain(ctx, router.Middlewares)

	var nextHandler http.Handler
	var serviceName string

	// Check if this router has child routers or a service
	if len(router.ChildRefs) > 0 {
		// This router routes to child routers - create a muxer for them
		childMuxer, err := m.buildChildRoutersMuxer(ctx, router.ChildRefs)
		if err != nil {
			return nil, fmt.Errorf("building child routers muxer: %w", err)
		}
		nextHandler = childMuxer
		serviceName = fmt.Sprintf("muxer@%s", routerName)
	} else if router.Service != "" {
		// This router routes to a service
		qualifiedService := provider.GetQualifiedName(ctx, router.Service)
		sHandler, err := m.serviceManager.BuildHTTP(ctx, qualifiedService)
		if err != nil {
			return nil, err
		}
		nextHandler = sHandler
		serviceName = qualifiedService
	} else {
		return nil, errors.New("the router must have either a service or child routers")
	}

	// Access logs, metrics, and tracing middlewares are idempotent if the associated signal is disabled.
	chain = chain.Append(observability.WrapRouterHandler(ctx, routerName, router.Rule, serviceName))
	metricsHandler := metricsMiddle.RouterMetricsHandler(ctx, m.observabilityMgr.MetricsRegistry(), routerName, serviceName)

	chain = chain.Append(observability.WrapMiddleware(ctx, metricsHandler))
	chain = chain.Append(func(next http.Handler) (http.Handler, error) {
		return accesslog.NewFieldHandler(next, accesslog.RouterName, routerName, nil), nil
	})

	return chain.Extend(*mHandler).Then(nextHandler)
}

// buildChildRoutersMuxer creates a muxer for child routers.
func (m *Manager) buildChildRoutersMuxer(ctx context.Context, childRefs []string) (http.Handler, error) {
	childMuxer := httpmuxer.NewMuxer(m.parser)

	// Set a default handler for the child muxer (404 Not Found)
	childMuxer.SetDefaultHandler(http.NotFoundHandler())

	for _, childName := range childRefs {
		childRouter, exists := m.conf.Routers[childName]
		if !exists {
			return nil, fmt.Errorf("child router %q does not exist", childName)
		}

		// Fixme: risk of building an empty muxer if all child routers have errors.
		// Skip child routers with errors
		if len(childRouter.Err) > 0 {
			continue
		}

		logger := log.Ctx(ctx).With().Str(logs.RouterName, childName).Logger()
		ctxChild := logger.WithContext(provider.AddInContext(ctx, childName))

		// Set priority if not set
		if childRouter.Priority == 0 {
			childRouter.Priority = httpmuxer.GetRulePriority(childRouter.Rule)
		}

		// Build the child router handler
		childHandler, err := m.buildRouterHandler(ctxChild, childName, childRouter)
		if err != nil {
			childRouter.AddError(err, true)
			logger.Error().Err(err).Send()
			continue
		}

		// Add the child router to the muxer
		if err = childMuxer.AddRoute(childRouter.Rule, childRouter.RuleSyntax, childRouter.Priority, childHandler); err != nil {
			childRouter.AddError(err, true)
			logger.Error().Err(err).Send()
			continue
		}
	}

	return childMuxer, nil
}
