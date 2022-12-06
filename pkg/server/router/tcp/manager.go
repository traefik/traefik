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
	httpmuxer "github.com/traefik/traefik/v2/pkg/muxer/http"
	tcpmuxer "github.com/traefik/traefik/v2/pkg/muxer/tcp"
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
func (m *Manager) BuildHandlers(rootCtx context.Context, entryPoints []string) map[string]*Router {
	entryPointsRouters := m.getTCPRouters(rootCtx, entryPoints)
	entryPointsRoutersHTTP := m.getHTTPRouters(rootCtx, entryPoints, true)

	entryPointHandlers := make(map[string]*Router)
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

func (m *Manager) buildEntryPointHandler(ctx context.Context, configs map[string]*runtime.TCPRouterInfo, configsHTTP map[string]*runtime.RouterInfo, handlerHTTP, handlerHTTPS http.Handler) (*Router, error) {
	// Build a new Router.
	router, err := NewRouter()
	if err != nil {
		return nil, err
	}

	router.SetHTTPHandler(handlerHTTP)

	// Even though the error is seemingly ignored (aside from logging it),
	// we actually rely later on the fact that a tls config is nil (which happens when an error is returned) to take special steps
	// when assigning a handler to a route.
	defaultTLSConf, err := m.tlsManager.Get(traefiktls.DefaultTLSStoreName, traefiktls.DefaultTLSConfigName)
	if err != nil {
		log.FromContext(ctx).Errorf("Error during the build of the default TLS configuration: %v", err)
	}

	// Keyed by domain. The source of truth for doing SNI checking (domain fronting).
	// As soon as there's (at least) two different tlsOptions found for the same domain,
	// we set the value to the default TLS conf.
	tlsOptionsForHost := map[string]string{}

	// Keyed by domain, then by options reference.
	// The actual source of truth for what TLS options will actually be used for the connection.
	// As opposed to tlsOptionsForHost, it keeps track of all the (different) TLS
	// options that occur for a given host name, so that later on we can set relevant
	// errors and logging for all the routers concerned (i.e. wrongly configured).
	tlsOptionsForHostSNI := map[string]map[string]nameAndConfig{}

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

		domains, err := httpmuxer.ParseDomains(routerHTTPConfig.Rule)
		if err != nil {
			routerErr := fmt.Errorf("invalid rule %s, error: %w", routerHTTPConfig.Rule, err)
			routerHTTPConfig.AddError(routerErr, true)
			logger.Error(routerErr)
			continue
		}

		if len(domains) == 0 {
			// Extra Host(*) rule, for HTTPS routers with no Host rule,
			// and for requests for which the SNI does not match _any_ of the other existing routers Host.
			// This is only about choosing the TLS configuration.
			// The actual routing will be done further on by the HTTPS handler.
			// See examples below.
			router.AddHTTPTLSConfig("*", defaultTLSConf)

			// The server name (from a Host(SNI) rule) is the only parameter (available in HTTP routing rules) on which we can map a TLS config,
			// because it is the only one accessible before decryption (we obtain it during the ClientHello).
			// Therefore, when a router has no Host rule, it does not make any sense to specify some TLS options.
			// Consequently, when it comes to deciding what TLS config will be used,
			// for a request that will match an HTTPS router with no Host rule,
			// the result will depend on the _others_ existing routers (their Host rule, to be precise), and the TLS options associated with them,
			// even though they don't match the incoming request. Consider the following examples:

			//	# conf1
			//	httpRouter1:
			//		rule: PathPrefix("/foo")
			//	# Wherever the request comes from, the TLS config used will be the default one, because of the Host(*) fallback.

			//	# conf2
			//	httpRouter1:
			//		rule: PathPrefix("/foo")
			//
			//	httpRouter2:
			//		rule: Host("foo.com") && PathPrefix("/bar")
			//		tlsoptions: myTLSOptions
			//	# When a request for "/foo" comes, even though it won't be routed by httpRouter2,
			//	# if its SNI is set to foo.com, myTLSOptions will be used for the TLS connection.
			//	# Otherwise, it will fallback to the default TLS config.
			logger.Warnf("No domain found in rule %v, the TLS options applied for this router will depend on the SNI of each request", routerHTTPConfig.Rule)
		}

		// Even though the error is seemingly ignored (aside from logging it),
		// we actually rely later on the fact that a tls config is nil (which happens when an error is returned) to take special steps
		// when assigning a handler to a route.
		tlsConf, tlsConfErr := m.tlsManager.Get(traefiktls.DefaultTLSStoreName, tlsOptionsName)
		if tlsConfErr != nil {
			// Note: we do not call AddError here because we already did so when buildRouterHandler errored for the same reason.
			logger.Error(tlsConfErr)
		}

		for _, domain := range domains {
			// domain is already in lower case thanks to the domain parsing
			if tlsOptionsForHostSNI[domain] == nil {
				tlsOptionsForHostSNI[domain] = make(map[string]nameAndConfig)
			}
			tlsOptionsForHostSNI[domain][tlsOptionsName] = nameAndConfig{
				routerName: routerHTTPName,
				TLSConfig:  tlsConf,
			}

			if name, ok := tlsOptionsForHost[domain]; ok && name != tlsOptionsName {
				// Different tlsOptions on the same domain, so fallback to default
				tlsOptionsForHost[domain] = traefiktls.DefaultTLSConfigName
			} else {
				tlsOptionsForHost[domain] = tlsOptionsName
			}
		}
	}

	sniCheck := snicheck.New(tlsOptionsForHost, handlerHTTPS)

	// Keep in mind that defaultTLSConf might be nil here.
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

			if config == nil {
				// we use nil config as a signal to insert a handler
				// that enforces that TLS connection attempts to the corresponding (broken) router should fail.
				logger.Debugf("Adding special closing route for %s because broken TLS options %s", hostSNI, optionsName)
				router.AddHTTPTLSConfig(hostSNI, nil)
				continue
			}

			logger.Debugf("Adding route for %s with TLS options %s", hostSNI, optionsName)
			router.AddHTTPTLSConfig(hostSNI, config)
			continue
		}

		// multiple tlsConfigs

		routers := make([]string, 0, len(tlsConfigs))
		for _, v := range tlsConfigs {
			configsHTTP[v.routerName].AddError(fmt.Errorf("found different TLS options for routers on the same host %v, so using the default TLS options instead", hostSNI), false)
			routers = append(routers, v.routerName)
		}

		logger.Warnf("Found different TLS options for routers on the same host %v, so using the default TLS options instead for these routers: %#v", hostSNI, routers)
		if defaultTLSConf == nil {
			logger.Debugf("Adding special closing route for %s because broken default TLS options", hostSNI)
		}

		router.AddHTTPTLSConfig(hostSNI, defaultTLSConf)
	}

	m.addTCPHandlers(ctx, configs, router)

	return router, nil
}

// addTCPHandlers creates the TCP handlers defined in configs, and adds them to router.
func (m *Manager) addTCPHandlers(ctx context.Context, configs map[string]*runtime.TCPRouterInfo, router *Router) {
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

		domains, err := tcpmuxer.ParseHostSNI(routerConfig.Rule)
		if err != nil {
			routerErr := fmt.Errorf("invalid rule: %q , %w", routerConfig.Rule, err)
			routerConfig.AddError(routerErr, true)
			logger.Error(routerErr)
			continue
		}

		// HostSNI Rule, but TLS not set on the router, which is an error.
		// However, we allow the HostSNI(*) exception.
		if len(domains) > 0 && routerConfig.TLS == nil && domains[0] != "*" {
			routerErr := fmt.Errorf("invalid rule: %q , has HostSNI matcher, but no TLS on router", routerConfig.Rule)
			routerConfig.AddError(routerErr, true)
			logger.Error(routerErr)
		}

		var handler tcp.Handler
		if routerConfig.TLS == nil || routerConfig.TLS.Passthrough {
			handler, err = m.buildTCPHandler(ctxRouter, routerConfig)
			if err != nil {
				routerConfig.AddError(err, true)
				logger.Error(err)
				continue
			}
		}

		if routerConfig.TLS == nil {
			logger.Debugf("Adding route for %q", routerConfig.Rule)
			if err := router.AddRoute(routerConfig.Rule, routerConfig.Priority, handler); err != nil {
				routerConfig.AddError(err, true)
				logger.Error(err)
			}
			continue
		}

		if routerConfig.TLS.Passthrough {
			logger.Debugf("Adding Passthrough route for %q", routerConfig.Rule)
			if err := router.muxerTCPTLS.AddRoute(routerConfig.Rule, routerConfig.Priority, handler); err != nil {
				routerConfig.AddError(err, true)
				logger.Error(err)
			}
			continue
		}

		for _, domain := range domains {
			if httpmuxer.IsASCII(domain) {
				continue
			}

			asciiError := fmt.Errorf("invalid domain name value %q, non-ASCII characters are not allowed", domain)
			routerConfig.AddError(asciiError, true)
			logger.Error(asciiError)
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

			logger.Error(err)
			logger.Debugf("Adding special TLS closing route for %q because broken TLS options %s", routerConfig.Rule, tlsOptionsName)

			err = router.muxerTCPTLS.AddRoute(routerConfig.Rule, routerConfig.Priority, &brokenTLSRouter{})
			if err != nil {
				routerConfig.AddError(err, true)
				logger.Error(err)
			}
			continue
		}

		// Now that the Rule is not just about the Host, we could theoretically have a config like:
		//	router1:
		//		rule: HostSNI(foo.com) && ClientIP(IP1)
		//		tlsOption: tlsOne
		//	router2:
		//		rule: HostSNI(foo.com) && ClientIP(IP2)
		//		tlsOption: tlsTwo
		// i.e. same HostSNI but different tlsOptions
		// This is only applicable if the muxer can decide about the routing _before_ telling the client about the tlsConf (i.e. before the TLS HandShake).
		// This seems to be the case so far with the existing matchers (HostSNI, and ClientIP), so it's all good.
		// Otherwise, we would have to do as for HTTPS, i.e. disallow different TLS configs for the same HostSNIs.

		handler, err = m.buildTCPHandler(ctxRouter, routerConfig)
		if err != nil {
			routerConfig.AddError(err, true)
			logger.Error(err)
			continue
		}

		handler = &tcp.TLSHandler{
			Next:   handler,
			Config: tlsConf,
		}

		logger.Debugf("Adding TLS route for %q", routerConfig.Rule)

		err = router.muxerTCPTLS.AddRoute(routerConfig.Rule, routerConfig.Priority, handler)
		if err != nil {
			routerConfig.AddError(err, true)
			logger.Error(err)
		}
	}
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
