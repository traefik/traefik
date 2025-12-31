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
func NewManager(conf *runtime.Configuration,
	serviceManager serviceManager,
	middlewaresBuilder middlewareBuilder,
	observabilityMgr *middleware.ObservabilityMgr,
	tlsManager *tls.Manager,
	parser httpmuxer.SyntaxParser,
) *Manager {
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

		handler, err := m.buildRouterHandler(ctxRouter, entryPointName, routerName, routerConfig)
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

func (m *Manager) buildRouterHandler(ctx context.Context, entryPointName, routerName string, routerConfig *runtime.RouterInfo) (http.Handler, error) {
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

	handler, err := m.buildHTTPHandler(ctx, routerConfig, entryPointName, routerName)
	if err != nil {
		return nil, err
	}

	m.routerHandlers[routerName] = handler
	return handler, nil
}

func (m *Manager) buildHTTPHandler(ctx context.Context, router *runtime.RouterInfo, entryPointName, routerName string) (http.Handler, error) {
	var qualifiedNames []string
	for _, name := range router.Middlewares {
		qualifiedNames = append(qualifiedNames, provider.GetQualifiedName(ctx, name))
	}
	router.Middlewares = qualifiedNames

	chain := alice.New()

	if router.DefaultRule {
		chain = chain.Append(denyrouterrecursion.WrapHandler(routerName))
	}

	var (
		nextHandler http.Handler
		serviceName string
		err         error
	)

	// Check if this router has child routers or a service.
	switch {
	case len(router.ChildRefs) > 0:
		// This router routes to child routers - create a muxer for them
		nextHandler, err = m.buildChildRoutersMuxer(ctx, entryPointName, router.ChildRefs)
		if err != nil {
			return nil, fmt.Errorf("building child routers muxer: %w", err)
		}
		serviceName = fmt.Sprintf("%s-muxer", routerName)
	case router.Service != "":
		// This router routes to a service
		qualifiedService := provider.GetQualifiedName(ctx, router.Service)
		nextHandler, err = m.serviceManager.BuildHTTP(ctx, qualifiedService)
		if err != nil {
			return nil, err
		}
		serviceName = qualifiedService
	default:
		return nil, errors.New("router must have either a service or child routers")
	}

	// Access logs, metrics, and tracing middlewares are idempotent if the associated signal is disabled.
	chain = chain.Append(observability.WrapRouterHandler(ctx, routerName, router.Rule, serviceName))

	metricsHandler := metricsMiddle.RouterMetricsHandler(ctx, m.observabilityMgr.MetricsRegistry(), routerName, serviceName)
	chain = chain.Append(observability.WrapMiddleware(ctx, metricsHandler))

	chain = chain.Append(func(next http.Handler) (http.Handler, error) {
		return accesslog.NewConcatFieldHandler(next, accesslog.RouterName, routerName), nil
	})

	// Here we are adding deny handlers for encoded path characters and fragment.
	// Deny handler are only added for root routers, child routers are protected by their parent router deny handlers.
	if len(router.ParentRefs) == 0 {
		chain = chain.Append(func(next http.Handler) (http.Handler, error) {
			return denyFragment(next), nil
		})
		chain = chain.Append(func(next http.Handler) (http.Handler, error) {
			return denyEncodedPathCharacters(router.DeniedEncodedPathCharacters.Map(), next), nil
		})
	}

	mHandler := m.middlewaresBuilder.BuildChain(ctx, router.Middlewares)

	return chain.Extend(*mHandler).Then(nextHandler)
}

// ParseRouterTree sets up router tree and validates router configuration.
// This function performs the following operations in order:
//
// 1. Populate ChildRefs: Uses ParentRefs to build the parent-child relationship graph
// 2. Root-first traversal: Starting from root routers (no ParentRefs), traverses the tree
// 3. Cycle detection: Detects circular dependencies and removes cyclic links
// 4. Reachability check: Marks routers unreachable from any root as disabled
// 5. Dead-end detection: Marks routers with no service and no children as disabled
// 6. Validation: Checks for configuration errors
//
// Router status is set during this process:
// - Enabled: Reachable routers with valid configuration
// - Disabled: Unreachable, dead-end, or routers with critical errors
// - Warning: Routers with non-critical errors (like cycles)
//
// The function modifies router.Status, router.ChildRefs, and adds errors to router.Err.
func (m *Manager) ParseRouterTree() {
	if m.conf == nil || m.conf.Routers == nil {
		return
	}

	// Populate ChildRefs based on ParentRefs and find root routers.
	var rootRouters []string
	for routerName, router := range m.conf.Routers {
		if len(router.ParentRefs) == 0 {
			rootRouters = append(rootRouters, routerName)
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

		// Check for non-root router with TLS config.
		if router.TLS != nil {
			router.AddError(errors.New("non-root router cannot have TLS configuration"), true)
			continue
		}

		// Check for non-root router with Observability config.
		if router.Observability != nil {
			router.AddError(errors.New("non-root router cannot have Observability configuration"), true)
			continue
		}

		// Check for non-root router with Entrypoint config.
		if len(router.EntryPoints) > 0 {
			router.AddError(errors.New("non-root router cannot have Entrypoints configuration"), true)
			continue
		}
	}
	sort.Strings(rootRouters)

	// Root-first traversal with cycle detection.
	visited := make(map[string]bool)
	currentPath := make(map[string]bool)
	var path []string

	for _, rootName := range rootRouters {
		if !visited[rootName] {
			m.traverse(rootName, visited, currentPath, path)
		}
	}

	for routerName, router := range m.conf.Routers {
		// Set status for all routers based on reachability.
		if !visited[routerName] {
			router.AddError(errors.New("router is not reachable"), true)
			continue
		}

		// Detect dead-end routers (no service + no children) - AFTER cycle handling.
		if router.Service == "" && len(router.ChildRefs) == 0 {
			router.AddError(errors.New("router has no service and no child routers"), true)
			continue
		}

		// Check for router with service that is referenced as a parent.
		if router.Service != "" && len(router.ChildRefs) > 0 {
			router.AddError(errors.New("router has both a service and is referenced as a parent by other routers"), true)
			continue
		}
	}
}

// traverse performs a depth-first traversal starting from root routers,
// detecting cycles and marking visited routers for reachability detection.
func (m *Manager) traverse(routerName string, visited, currentPath map[string]bool, path []string) {
	if currentPath[routerName] {
		// Found a cycle - handle it properly.
		m.handleCycle(routerName, path)
		return
	}

	if visited[routerName] {
		return
	}

	router, exists := m.conf.Routers[routerName]
	// Since the ChildRefs population already guarantees router existence, this check is purely defensive.
	if !exists {
		visited[routerName] = true
		return
	}

	visited[routerName] = true
	currentPath[routerName] = true
	newPath := append(path, routerName)

	// Sort ChildRefs for deterministic behavior.
	sortedChildRefs := make([]string, len(router.ChildRefs))
	copy(sortedChildRefs, router.ChildRefs)
	sort.Strings(sortedChildRefs)

	// Traverse children.
	for _, childName := range sortedChildRefs {
		m.traverse(childName, visited, currentPath, newPath)
	}

	currentPath[routerName] = false
}

// handleCycle handles cycle detection and removes the victim from guilty router's ChildRefs.
func (m *Manager) handleCycle(victimRouter string, path []string) {
	// Find where the cycle starts in the path
	cycleStart := -1
	for i, name := range path {
		if name == victimRouter {
			cycleStart = i
			break
		}
	}

	if cycleStart < 0 {
		return
	}

	// Build the cycle path: from cycle start to current + victim.
	cyclePath := append(path[cycleStart:], victimRouter)
	cycleRouters := strings.Join(cyclePath, " -> ")

	// The guilty router is the last one in the path (the one creating the cycle).
	if len(path) > 0 {
		guiltyRouterName := path[len(path)-1]
		guiltyRouter, exists := m.conf.Routers[guiltyRouterName]
		if !exists {
			return
		}

		// Add cycle error to guilty router.
		guiltyRouter.AddError(fmt.Errorf("cyclic reference detected in router tree: %s", cycleRouters), false)

		// Remove victim from guilty router's ChildRefs.
		for i, childRef := range guiltyRouter.ChildRefs {
			if childRef == victimRouter {
				guiltyRouter.ChildRefs = append(guiltyRouter.ChildRefs[:i], guiltyRouter.ChildRefs[i+1:]...)
				break
			}
		}
	}
}

// buildChildRoutersMuxer creates a muxer for child routers.
func (m *Manager) buildChildRoutersMuxer(ctx context.Context, entryPointName string, childRefs []string) (http.Handler, error) {
	childMuxer := httpmuxer.NewMuxer(m.parser)

	// Set a default handler for the child muxer (404 Not Found).
	childMuxer.SetDefaultHandler(http.NotFoundHandler())

	childCount := 0
	for _, childName := range childRefs {
		childRouter, exists := m.conf.Routers[childName]
		if !exists {
			return nil, fmt.Errorf("child router %q does not exist", childName)
		}

		// Skip child routers with errors.
		if len(childRouter.Err) > 0 {
			continue
		}

		logger := log.Ctx(ctx).With().Str(logs.RouterName, childName).Logger()
		ctxChild := logger.WithContext(provider.AddInContext(ctx, childName))

		// Set priority if not set.
		if childRouter.Priority == 0 {
			childRouter.Priority = httpmuxer.GetRulePriority(childRouter.Rule)
		}

		// Build the child router handler.
		childHandler, err := m.buildRouterHandler(ctxChild, entryPointName, childName, childRouter)
		if err != nil {
			childRouter.AddError(err, true)
			logger.Error().Err(err).Send()
			continue
		}

		// Add the child router to the muxer.
		if err = childMuxer.AddRoute(childRouter.Rule, childRouter.RuleSyntax, childRouter.Priority, childHandler); err != nil {
			childRouter.AddError(err, true)
			logger.Error().Err(err).Send()
			continue
		}

		childCount++
	}

	// Prevent empty muxer.
	if childCount == 0 {
		return nil, fmt.Errorf("no child routers could be added to muxer (%d skipped)", len(childRefs))
	}

	return childMuxer, nil
}
