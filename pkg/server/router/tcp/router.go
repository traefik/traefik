package tcp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	"github.com/traefik/traefik/v2/pkg/config/runtime"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/middlewares/snicheck"
	"github.com/traefik/traefik/v2/pkg/rules"
	"github.com/traefik/traefik/v2/pkg/server/provider"
	tcpservice "github.com/traefik/traefik/v2/pkg/server/service/tcp"
	"github.com/traefik/traefik/v2/pkg/tcp"
	traefiktls "github.com/traefik/traefik/v2/pkg/tls"
)

type middlewareBuilder interface {
	BuildChain(ctx context.Context, names []string) *tcp.Chain
}

// NewManager Creates a new Manager.
func NewManager(conf *runtime.Configuration,
	serviceManager *tcpservice.Manager,
	middlewaresBuilder middlewareBuilder,
	httpHandlers map[string]http.Handler,
	httpsHandlers map[string]http.Handler,
	tlsManager *traefiktls.Manager,
) *Manager {
	return &Manager{
		serviceManager:     serviceManager,
		middlewaresBuilder: middlewaresBuilder,
		httpHandlers:       httpHandlers,
		httpsHandlers:      httpsHandlers,
		tlsManager:         tlsManager,
		conf:               conf,
	}
}

// Manager is a route/router manager.
type Manager struct {
	serviceManager     *tcpservice.Manager
	middlewaresBuilder middlewareBuilder
	httpHandlers       map[string]http.Handler
	httpsHandlers      map[string]http.Handler
	tlsManager         *traefiktls.Manager
	conf               *runtime.Configuration
}

func (m *Manager) getTCPRouters(ctx context.Context, entryPoints []string) map[string]map[string]*runtime.TCPRouterInfo {
	if m.conf != nil {
		return m.conf.GetTCPRoutersByEntryPoints(ctx, entryPoints)
	}

	return make(map[string]map[string]*runtime.TCPRouterInfo)
}

func (m *Manager) getHTTPRouters(ctx context.Context, entryPoints []string, tls bool) map[string]map[string]*runtime.RouterInfo {
	if m.conf != nil {
		return m.conf.GetRoutersByEntryPoints(ctx, entryPoints, tls)
	}

	return make(map[string]map[string]*runtime.RouterInfo)
}

// BuildHandlers builds the handlers for the given entrypoints.
func (m *Manager) BuildHandlers(rootCtx context.Context, entryPoints []string) map[string]*tcp.Router {
	entryPointsRouters := m.getTCPRouters(rootCtx, entryPoints)
	entryPointsRoutersHTTP := m.getHTTPRouters(rootCtx, entryPoints, true)

	entryPointHandlers := make(map[string]*tcp.Router)
	for _, entryPointName := range entryPoints {
		entryPointName := entryPointName

		routers := entryPointsRouters[entryPointName]

		ctx := log.With(rootCtx, log.Str(log.EntryPointName, entryPointName))

		handler, err := m.buildEntryPointHandler(ctx, routers, entryPointsRoutersHTTP[entryPointName], m.httpHandlers[entryPointName], m.httpsHandlers[entryPointName])
		if err != nil {
			log.FromContext(ctx).Error(err)
			continue
		}
		entryPointHandlers[entryPointName] = handler
	}
	return entryPointHandlers
}

type nameAndConfig struct {
	routerName string // just so we have it as additional information when logging
	TLSConfig  *tls.Config
}

func (m *Manager) buildEntryPointHandler(ctx context.Context, configs map[string]*runtime.TCPRouterInfo, configsHTTP map[string]*runtime.RouterInfo, handlerHTTP, handlerHTTPS http.Handler) (*tcp.Router, error) {
	// Build a new Router.
	router := &tcp.Router{}
	router.SetHTTPHandler(handlerHTTP)

	defaultTLSConf, err := m.tlsManager.Get(traefiktls.DefaultTLSStoreName, traefiktls.DefaultTLSConfigName)
	if err != nil {
		log.FromContext(ctx).Errorf("Error during the build of the default TLS configuration: %v", err)
	}

	if len(configsHTTP) > 0 {
		router.AddRouteHTTPTLS("*", defaultTLSConf)
	}

	// Keyed by domain, then by options reference.
	tlsOptionsForHostSNI := map[string]map[string]nameAndConfig{}
	tlsOptionsForHost := map[string]string{}
	for routerHTTPName, routerHTTPConfig := range configsHTTP {
		if routerHTTPConfig.TLS == nil {
			continue
		}

		ctxRouter := log.With(provider.AddInContext(ctx, routerHTTPName), log.Str(log.RouterName, routerHTTPName))
		logger := log.FromContext(ctxRouter)

		tlsOptionsName := traefiktls.DefaultTLSConfigName
		if len(routerHTTPConfig.TLS.Options) > 0 && routerHTTPConfig.TLS.Options != traefiktls.DefaultTLSConfigName {
			tlsOptionsName = provider.GetQualifiedName(ctxRouter, routerHTTPConfig.TLS.Options)
		}

		domains, err := rules.ParseDomains(routerHTTPConfig.Rule)
		if err != nil {
			routerErr := fmt.Errorf("invalid rule %s, error: %w", routerHTTPConfig.Rule, err)
			routerHTTPConfig.AddError(routerErr, true)
			logger.Debug(routerErr)
			continue
		}

		if len(domains) == 0 {
			logger.Warnf("No domain found in rule %v, the TLS options applied for this router will depend on the hostSNI of each request", routerHTTPConfig.Rule)
		}

		for _, domain := range domains {
			tlsConf, err := m.tlsManager.Get(traefiktls.DefaultTLSStoreName, tlsOptionsName)
			if err != nil {
				routerHTTPConfig.AddError(err, true)
				logger.Debug(err)
				continue
			}

			// domain is already in lower case thanks to the domain parsing
			if tlsOptionsForHostSNI[domain] == nil {
				tlsOptionsForHostSNI[domain] = make(map[string]nameAndConfig)
			}
			tlsOptionsForHostSNI[domain][tlsOptionsName] = nameAndConfig{
				routerName: routerHTTPName,
				TLSConfig:  tlsConf,
			}

			if name, ok := tlsOptionsForHost[domain]; ok && name != tlsOptionsName {
				// Different tlsOptions on the same domain fallback to default
				tlsOptionsForHost[domain] = traefiktls.DefaultTLSConfigName
			} else {
				tlsOptionsForHost[domain] = tlsOptionsName
			}
		}
	}

	sniCheck := snicheck.New(tlsOptionsForHost, handlerHTTPS)

	router.SetHTTPSHandler(sniCheck, defaultTLSConf)

	logger := log.FromContext(ctx)
	for hostSNI, tlsConfigs := range tlsOptionsForHostSNI {
		if len(tlsConfigs) == 1 {
			var optionsName string
			var config *tls.Config
			for k, v := range tlsConfigs {
				optionsName = k
				config = v.TLSConfig
				break
			}

			logger.Debugf("Adding route for %s with TLS options %s", hostSNI, optionsName)

			router.AddRouteHTTPTLS(hostSNI, config)
		} else {
			routers := make([]string, 0, len(tlsConfigs))
			for _, v := range tlsConfigs {
				configsHTTP[v.routerName].AddError(fmt.Errorf("found different TLS options for routers on the same host %v, so using the default TLS options instead", hostSNI), false)
				routers = append(routers, v.routerName)
			}

			logger.Warnf("Found different TLS options for routers on the same host %v, so using the default TLS options instead for these routers: %#v", hostSNI, routers)

			router.AddRouteHTTPTLS(hostSNI, defaultTLSConf)
		}
	}

	for routerName, routerConfig := range configs {
		ctxRouter := log.With(provider.AddInContext(ctx, routerName), log.Str(log.RouterName, routerName))
		logger := log.FromContext(ctxRouter)

		if routerConfig.Service == "" {
			err := errors.New("the service is missing on the router")
			routerConfig.AddError(err, true)
			logger.Error(err)
			continue
		}

		if routerConfig.Rule == "" {
			err := errors.New("router has no rule")
			routerConfig.AddError(err, true)
			logger.Error(err)
			continue
		}

		handler, err := m.buildTCPHandler(ctxRouter, routerConfig)
		if err != nil {
			routerConfig.AddError(err, true)
			logger.Error(err)
			continue
		}

		domains, err := rules.ParseHostSNI(routerConfig.Rule)
		if err != nil {
			routerErr := fmt.Errorf("unknown rule %s", routerConfig.Rule)
			routerConfig.AddError(routerErr, true)
			logger.Error(routerErr)
			continue
		}

		for _, domain := range domains {
			logger.Debugf("Adding route %s on TCP", domain)
			switch {
			case routerConfig.TLS != nil:
				if !rules.IsASCII(domain) {
					asciiError := fmt.Errorf("invalid domain name value %q, non-ASCII characters are not allowed", domain)
					routerConfig.AddError(asciiError, true)
					logger.Debug(asciiError)
					continue
				}

				if routerConfig.TLS.Passthrough {
					route := tcp.NewRoute(handler)
					route.AddMatcher(tcp.NewSNIHost(domain))
					router.AddRoute(route)
					continue
				}

				tlsOptionsName := routerConfig.TLS.Options

				if len(tlsOptionsName) == 0 {
					tlsOptionsName = traefiktls.DefaultTLSConfigName
				}

				if tlsOptionsName != traefiktls.DefaultTLSConfigName {
					tlsOptionsName = provider.GetQualifiedName(ctxRouter, tlsOptionsName)
				}

				tlsConf, err := m.tlsManager.Get(traefiktls.DefaultTLSStoreName, tlsOptionsName)
				if err != nil {
					routerConfig.AddError(err, true)
					logger.Debug(err)
					continue
				}

				route := tcp.NewRoute(&tcp.TLSHandler{Next: handler, Config: tlsConf})
				route.AddMatcher(tcp.NewSNIHost(domain))
				router.AddRoute(route)
			case domain == "*":
				router.SetCatchAllNoTLS(handler)
			default:
				logger.Warn("TCP Router ignored, cannot specify a Host rule without TLS")
			}
		}

		ips, err := rules.ParseClientIP(routerConfig.Rule)
		if err != nil {
			routerErr := fmt.Errorf("unknown rule %s", routerConfig.Rule)
			routerConfig.AddError(routerErr, true)
			logger.Error(routerErr)
			continue
		}
		for _, ip := range ips {
			route := tcp.NewRoute(handler)
			route.AddMatcher(tcp.NewClientIP(ip))
			router.AddRoute(route)
		}
	}

	return router, nil
}

func (m *Manager) buildTCPHandler(ctx context.Context, router *runtime.TCPRouterInfo) (tcp.Handler, error) {
	var qualifiedNames []string
	for _, name := range router.Middlewares {
		qualifiedNames = append(qualifiedNames, provider.GetQualifiedName(ctx, name))
	}
	router.Middlewares = qualifiedNames

	if router.Service == "" {
		return nil, errors.New("the service is missing on the router")
	}

	sHandler, err := m.serviceManager.BuildTCP(ctx, router.Service)
	if err != nil {
		return nil, err
	}

	mHandler := m.middlewaresBuilder.BuildChain(ctx, router.Middlewares)

	return tcp.NewChain().Extend(*mHandler).Then(sHandler)
}
