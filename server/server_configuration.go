package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/containous/flaeg/parse"
	"github.com/containous/mux"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/healthcheck"
	"github.com/containous/traefik/hostresolver"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/metrics"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/middlewares/pipelining"
	"github.com/containous/traefik/rules"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/tls/generate"
	"github.com/containous/traefik/types"
	"github.com/eapache/channels"
	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
	"github.com/vulcand/oxy/forward"
)

// loadConfiguration manages dynamically frontends, backends and TLS configurations
func (s *Server) loadConfiguration(configMsg types.ConfigMessage) {
	currentConfigurations := s.currentConfigurations.Get().(types.Configurations)

	// Copy configurations to new map so we don't change current if LoadConfig fails
	newConfigurations := make(types.Configurations)
	for k, v := range currentConfigurations {
		newConfigurations[k] = v
	}
	newConfigurations[configMsg.ProviderName] = configMsg.Configuration

	s.metricsRegistry.ConfigReloadsCounter().Add(1)

	newServerEntryPoints := s.loadConfig(newConfigurations, s.globalConfiguration)

	s.metricsRegistry.LastConfigReloadSuccessGauge().Set(float64(time.Now().Unix()))

	for newServerEntryPointName, newServerEntryPoint := range newServerEntryPoints {
		s.serverEntryPoints[newServerEntryPointName].httpRouter.UpdateHandler(newServerEntryPoint.httpRouter.GetHandler())

		if s.entryPoints[newServerEntryPointName].Configuration.TLS == nil {
			if newServerEntryPoint.certs.ContainsCertificates() {
				log.Debugf("Certificates not added to non-TLS entryPoint %s.", newServerEntryPointName)
			}
		} else {
			s.serverEntryPoints[newServerEntryPointName].certs.DynamicCerts.Set(newServerEntryPoint.certs.DynamicCerts.Get())
			s.serverEntryPoints[newServerEntryPointName].certs.ResetCache()
		}
		log.Infof("Server configuration reloaded on %s", s.serverEntryPoints[newServerEntryPointName].httpServer.Addr)
	}

	s.currentConfigurations.Set(newConfigurations)

	for _, listener := range s.configurationListeners {
		listener(*configMsg.Configuration)
	}

	s.postLoadConfiguration()
}

// loadConfig returns a new gorilla.mux Route from the specified global configuration and the dynamic
// provider configurations.
func (s *Server) loadConfig(configurations types.Configurations, globalConfiguration configuration.GlobalConfiguration) map[string]*serverEntryPoint {

	serverEntryPoints := s.buildServerEntryPoints()

	backendsHandlers := map[string]http.Handler{}
	backendsHealthCheck := map[string]*healthcheck.BackendConfig{}

	var postConfigs []handlerPostConfig

	for providerName, config := range configurations {
		frontendNames := sortedFrontendNamesForConfig(config)

		for _, frontendName := range frontendNames {
			frontendPostConfigs, err := s.loadFrontendConfig(providerName, frontendName, config,
				serverEntryPoints,
				backendsHandlers, backendsHealthCheck)
			if err != nil {
				log.Errorf("%v. Skipping frontend %s...", err, frontendName)
			}

			if len(frontendPostConfigs) > 0 {
				postConfigs = append(postConfigs, frontendPostConfigs...)
			}
		}
	}

	for _, postConfig := range postConfigs {
		err := postConfig(backendsHandlers)
		if err != nil {
			log.Errorf("middleware post configuration error: %v", err)
		}
	}

	healthcheck.GetHealthCheck(s.metricsRegistry).SetBackendsConfiguration(s.routinesPool.Ctx(), backendsHealthCheck)

	// Get new certificates list sorted per entrypoints
	// Update certificates
	entryPointsCertificates := s.loadHTTPSConfiguration(configurations, globalConfiguration.DefaultEntryPoints)

	// Sort routes and update certificates
	for serverEntryPointName, serverEntryPoint := range serverEntryPoints {
		serverEntryPoint.httpRouter.GetHandler().SortRoutes()
		if _, exists := entryPointsCertificates[serverEntryPointName]; exists {
			serverEntryPoint.certs.DynamicCerts.Set(entryPointsCertificates[serverEntryPointName])
		}
	}

	return serverEntryPoints
}

func (s *Server) loadFrontendConfig(
	providerName string, frontendName string, config *types.Configuration,
	serverEntryPoints map[string]*serverEntryPoint,
	backendsHandlers map[string]http.Handler, backendsHealthCheck map[string]*healthcheck.BackendConfig,
) ([]handlerPostConfig, error) {

	frontend := config.Frontends[frontendName]
	hostResolver := buildHostResolver(s.globalConfiguration)

	if len(frontend.EntryPoints) == 0 {
		return nil, fmt.Errorf("no entrypoint defined for frontend %s", frontendName)
	}

	backend := config.Backends[frontend.Backend]
	if backend == nil {
		return nil, fmt.Errorf("undefined backend '%s' for frontend %s", frontend.Backend, frontendName)
	}

	frontendHash, err := frontend.Hash()
	if err != nil {
		return nil, fmt.Errorf("error calculating hash value for frontend %s: %v", frontendName, err)
	}

	var postConfigs []handlerPostConfig

	for _, entryPointName := range frontend.EntryPoints {
		log.Debugf("Wiring frontend %s to entryPoint %s", frontendName, entryPointName)

		entryPoint := s.entryPoints[entryPointName].Configuration

		if backendsHandlers[entryPointName+providerName+frontendHash] == nil {
			log.Debugf("Creating backend %s", frontend.Backend)

			handlers, responseModifier, postConfig, err := s.buildMiddlewares(frontendName, frontend, config.Backends, entryPointName, entryPoint, providerName)
			if err != nil {
				return nil, err
			}

			if postConfig != nil {
				postConfigs = append(postConfigs, postConfig)
			}

			fwd, err := s.buildForwarder(entryPointName, entryPoint, frontendName, frontend, responseModifier, backend)
			if err != nil {
				return nil, fmt.Errorf("failed to create the forwarder for frontend %s: %v", frontendName, err)
			}

			lb, healthCheckConfig, err := s.buildBalancerMiddlewares(frontendName, frontend, backend, fwd)
			if err != nil {
				return nil, err
			}

			// Handler used by error pages
			if backendsHandlers[entryPointName+providerName+frontend.Backend] == nil {
				backendsHandlers[entryPointName+providerName+frontend.Backend] = lb
			}

			if healthCheckConfig != nil {
				backendsHealthCheck[entryPointName+providerName+frontendHash] = healthCheckConfig
			}

			n := negroni.New()

			for _, handler := range handlers {
				n.Use(handler)
			}

			n.UseHandler(lb)

			backendsHandlers[entryPointName+providerName+frontendHash] = n
		} else {
			log.Debugf("Reusing backend %s [%s - %s - %s - %s]",
				frontend.Backend, entryPointName, providerName, frontendName, frontendHash)
		}

		serverRoute, err := buildServerRoute(serverEntryPoints[entryPointName], frontendName, frontend, hostResolver)
		if err != nil {
			return nil, err
		}

		handler := buildMatcherMiddlewares(serverRoute, backendsHandlers[entryPointName+providerName+frontendHash])
		serverRoute.Route.Handler(handler)

		err = serverRoute.Route.GetError()
		if err != nil {
			// FIXME error management
			log.Errorf("Error building route: %s", err)
		}
	}

	return postConfigs, nil
}

func (s *Server) buildForwarder(entryPointName string, entryPoint *configuration.EntryPoint,
	frontendName string, frontend *types.Frontend,
	responseModifier modifyResponse, backend *types.Backend) (http.Handler, error) {

	roundTripper, err := s.getRoundTripper(entryPointName, frontend.PassTLSCert, entryPoint.TLS)
	if err != nil {
		return nil, fmt.Errorf("failed to create RoundTripper for frontend %s: %v", frontendName, err)
	}

	rewriter, err := NewHeaderRewriter(entryPoint.ForwardedHeaders.TrustedIPs, entryPoint.ForwardedHeaders.Insecure)
	if err != nil {
		return nil, fmt.Errorf("error creating rewriter for frontend %s: %v", frontendName, err)
	}

	var flushInterval parse.Duration
	if backend.ResponseForwarding != nil {
		err := flushInterval.Set(backend.ResponseForwarding.FlushInterval)
		if err != nil {
			return nil, fmt.Errorf("error creating flush interval for frontend %s: %v", frontendName, err)
		}
	}

	var fwd http.Handler
	fwd, err = forward.New(
		forward.Stream(true),
		forward.PassHostHeader(frontend.PassHostHeader),
		forward.RoundTripper(roundTripper),
		forward.Rewriter(rewriter),
		forward.ResponseModifier(responseModifier),
		forward.BufferPool(s.bufferPool),
		forward.StreamingFlushInterval(time.Duration(flushInterval)),
		forward.WebsocketConnectionClosedHook(func(req *http.Request, conn net.Conn) {
			server := req.Context().Value(http.ServerContextKey).(*http.Server)
			if server != nil {
				connState := server.ConnState
				if connState != nil {
					connState(conn, http.StateClosed)
				}
			}
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating forwarder for frontend %s: %v", frontendName, err)
	}

	if s.tracingMiddleware.IsEnabled() {
		tm := s.tracingMiddleware.NewForwarderMiddleware(frontendName, frontend.Backend)

		next := fwd
		fwd = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tm.ServeHTTP(w, r, next.ServeHTTP)
		})
	}

	fwd = pipelining.NewPipelining(fwd)

	return fwd, nil
}

func buildServerRoute(serverEntryPoint *serverEntryPoint, frontendName string, frontend *types.Frontend, hostResolver *hostresolver.Resolver) (*types.ServerRoute, error) {
	serverRoute := &types.ServerRoute{Route: serverEntryPoint.httpRouter.GetHandler().NewRoute().Name(frontendName)}

	priority := 0
	for routeName, route := range frontend.Routes {
		rls := rules.Rules{Route: serverRoute, HostResolver: hostResolver}
		newRoute, err := rls.Parse(route.Rule)
		if err != nil {
			return nil, fmt.Errorf("error creating route for frontend %s: %v", frontendName, err)
		}

		serverRoute.Route = newRoute

		priority += len(route.Rule)
		log.Debugf("Creating route %s %s", routeName, route.Rule)
	}

	if frontend.Priority > 0 {
		serverRoute.Route.Priority(frontend.Priority)
	} else {
		serverRoute.Route.Priority(priority)
	}

	return serverRoute, nil
}

func (s *Server) preLoadConfiguration(configMsg types.ConfigMessage) {
	providersThrottleDuration := time.Duration(s.globalConfiguration.ProvidersThrottleDuration)
	s.defaultConfigurationValues(configMsg.Configuration)
	currentConfigurations := s.currentConfigurations.Get().(types.Configurations)

	if log.GetLevel() == logrus.DebugLevel {
		jsonConf, _ := json.Marshal(configMsg.Configuration)
		log.Debugf("Configuration received from provider %s: %s", configMsg.ProviderName, string(jsonConf))
	}

	if configMsg.Configuration == nil || configMsg.Configuration.Backends == nil && configMsg.Configuration.Frontends == nil && configMsg.Configuration.TLS == nil {
		log.Infof("Skipping empty Configuration for provider %s", configMsg.ProviderName)
		return
	}

	if reflect.DeepEqual(currentConfigurations[configMsg.ProviderName], configMsg.Configuration) {
		log.Infof("Skipping same configuration for provider %s", configMsg.ProviderName)
		return
	}

	providerConfigUpdateCh, ok := s.providerConfigUpdateMap[configMsg.ProviderName]
	if !ok {
		providerConfigUpdateCh = make(chan types.ConfigMessage)
		s.providerConfigUpdateMap[configMsg.ProviderName] = providerConfigUpdateCh
		s.routinesPool.Go(func(stop chan bool) {
			s.throttleProviderConfigReload(providersThrottleDuration, s.configurationValidatedChan, providerConfigUpdateCh, stop)
		})
	}

	providerConfigUpdateCh <- configMsg
}

func (s *Server) defaultConfigurationValues(configuration *types.Configuration) {
	if configuration == nil || configuration.Frontends == nil {
		return
	}
	s.configureFrontends(configuration.Frontends)
	configureBackends(configuration.Backends)
}

func (s *Server) configureFrontends(frontends map[string]*types.Frontend) {
	defaultEntrypoints := s.globalConfiguration.DefaultEntryPoints

	for frontendName, frontend := range frontends {
		// default endpoints if not defined in frontends
		if len(frontend.EntryPoints) == 0 {
			frontend.EntryPoints = defaultEntrypoints
		}

		frontendEntryPoints, undefinedEntryPoints := s.filterEntryPoints(frontend.EntryPoints)
		if len(undefinedEntryPoints) > 0 {
			log.Errorf("Undefined entry point(s) '%s' for frontend %s", strings.Join(undefinedEntryPoints, ","), frontendName)
		}

		frontend.EntryPoints = frontendEntryPoints
	}
}

func (s *Server) filterEntryPoints(entryPoints []string) ([]string, []string) {
	var frontendEntryPoints []string
	var undefinedEntryPoints []string

	for _, fepName := range entryPoints {
		var exist bool

		for epName := range s.entryPoints {
			if epName == fepName {
				exist = true
				break
			}
		}

		if exist {
			frontendEntryPoints = append(frontendEntryPoints, fepName)
		} else {
			undefinedEntryPoints = append(undefinedEntryPoints, fepName)
		}
	}

	return frontendEntryPoints, undefinedEntryPoints
}

func configureBackends(backends map[string]*types.Backend) {
	for backendName := range backends {
		backend := backends[backendName]
		if backend.LoadBalancer != nil && backend.LoadBalancer.Sticky {
			log.Warnf("Deprecated configuration found: %s. Please use %s.", "backend.LoadBalancer.Sticky", "backend.LoadBalancer.Stickiness")
		}

		_, err := types.NewLoadBalancerMethod(backend.LoadBalancer)
		if err == nil {
			if backend.LoadBalancer != nil && backend.LoadBalancer.Stickiness == nil && backend.LoadBalancer.Sticky {
				backend.LoadBalancer.Stickiness = &types.Stickiness{
					CookieName: "_TRAEFIK_BACKEND",
				}
			}
		} else {
			log.Debugf("Backend %s: %v", backendName, err)

			var stickiness *types.Stickiness
			if backend.LoadBalancer != nil {
				if backend.LoadBalancer.Stickiness == nil {
					if backend.LoadBalancer.Sticky {
						stickiness = &types.Stickiness{
							CookieName: "_TRAEFIK_BACKEND",
						}
					}
				} else {
					stickiness = backend.LoadBalancer.Stickiness
				}
			}
			backend.LoadBalancer = &types.LoadBalancer{
				Method:     "wrr",
				Stickiness: stickiness,
			}
		}
	}
}

func (s *Server) listenConfigurations(stop chan bool) {
	for {
		select {
		case <-stop:
			return
		case configMsg, ok := <-s.configurationValidatedChan:
			if !ok || configMsg.Configuration == nil {
				return
			}
			s.loadConfiguration(configMsg)
		}
	}
}

// throttleProviderConfigReload throttles the configuration reload speed for a single provider.
// It will immediately publish a new configuration and then only publish the next configuration after the throttle duration.
// Note that in the case it receives N new configs in the timeframe of the throttle duration after publishing,
// it will publish the last of the newly received configurations.
func (s *Server) throttleProviderConfigReload(throttle time.Duration, publish chan<- types.ConfigMessage, in <-chan types.ConfigMessage, stop chan bool) {
	ring := channels.NewRingChannel(1)
	defer ring.Close()

	s.routinesPool.Go(func(stop chan bool) {
		for {
			select {
			case <-stop:
				return
			case nextConfig := <-ring.Out():
				if config, ok := nextConfig.(types.ConfigMessage); ok {
					publish <- config
					time.Sleep(throttle)
				}
			}
		}
	})

	for {
		select {
		case <-stop:
			return
		case nextConfig := <-in:
			ring.In() <- nextConfig
		}
	}
}

func buildMatcherMiddlewares(serverRoute *types.ServerRoute, handler http.Handler) http.Handler {
	// path replace - This needs to always be the very last on the handler chain (first in the order in this function)
	// -- Replacing Path should happen at the very end of the Modifier chain, after all the Matcher+Modifiers ran
	if len(serverRoute.ReplacePath) > 0 {
		handler = &middlewares.ReplacePath{
			Path:    serverRoute.ReplacePath,
			Handler: handler,
		}
	}

	if len(serverRoute.ReplacePathRegex) > 0 {
		sp := strings.Split(serverRoute.ReplacePathRegex, " ")
		if len(sp) == 2 {
			handler = middlewares.NewReplacePathRegexHandler(sp[0], sp[1], handler)
		} else {
			log.Warnf("Invalid syntax for ReplacePathRegex: %s. Separate the regular expression and the replacement by a space.", serverRoute.ReplacePathRegex)
		}
	}

	// add prefix - This needs to always be right before ReplacePath on the chain (second in order in this function)
	// -- Adding Path Prefix should happen after all *Strip Matcher+Modifiers ran, but before Replace (in case it's configured)
	if len(serverRoute.AddPrefix) > 0 {
		handler = &middlewares.AddPrefix{
			Prefix:  serverRoute.AddPrefix,
			Handler: handler,
		}
	}

	// strip prefix
	if len(serverRoute.StripPrefixes) > 0 {
		handler = &middlewares.StripPrefix{
			Prefixes: serverRoute.StripPrefixes,
			Handler:  handler,
		}
	}

	// strip prefix with regex
	if len(serverRoute.StripPrefixesRegex) > 0 {
		handler = middlewares.NewStripPrefixRegex(handler, serverRoute.StripPrefixesRegex)
	}

	return handler
}

func (s *Server) postLoadConfiguration() {
	if s.metricsRegistry.IsEnabled() {
		activeConfig := s.currentConfigurations.Get().(types.Configurations)
		metrics.OnConfigurationUpdate(activeConfig)
	}

	if s.globalConfiguration.ACME == nil || s.leadership == nil || !s.leadership.IsLeader() {
		return
	}

	if s.globalConfiguration.ACME.OnHostRule {
		currentConfigurations := s.currentConfigurations.Get().(types.Configurations)
		for _, config := range currentConfigurations {
			for _, frontend := range config.Frontends {

				// check if one of the frontend entrypoints is configured with TLS
				// and is configured with ACME
				acmeEnabled := false
				for _, entryPoint := range frontend.EntryPoints {
					if s.globalConfiguration.ACME.EntryPoint == entryPoint && s.entryPoints[entryPoint].Configuration.TLS != nil {
						acmeEnabled = true
						break
					}
				}

				if acmeEnabled {
					for _, route := range frontend.Routes {
						rls := rules.Rules{}
						domains, err := rls.ParseDomains(route.Rule)
						if err != nil {
							log.Errorf("Error parsing domains: %v", err)
						} else if len(domains) == 0 {
							log.Debugf("No domain parsed in rule %q", route.Rule)
						} else {
							s.globalConfiguration.ACME.LoadCertificateForDomains(domains)
						}
					}
				}
			}
		}
	}
}

// loadHTTPSConfiguration add/delete HTTPS certificate managed dynamically
func (s *Server) loadHTTPSConfiguration(configurations types.Configurations, defaultEntryPoints configuration.DefaultEntryPoints) map[string]map[string]*tls.Certificate {
	newEPCertificates := make(map[string]map[string]*tls.Certificate)
	// Get all certificates
	for _, config := range configurations {
		if config.TLS != nil && len(config.TLS) > 0 {
			traefiktls.SortTLSPerEntryPoints(config.TLS, newEPCertificates, defaultEntryPoints)
		}
	}
	return newEPCertificates
}

func (s *Server) buildServerEntryPoints() map[string]*serverEntryPoint {
	serverEntryPoints := make(map[string]*serverEntryPoint)
	for entryPointName, entryPoint := range s.entryPoints {
		serverEntryPoints[entryPointName] = &serverEntryPoint{
			httpRouter:       middlewares.NewHandlerSwitcher(s.buildDefaultHTTPRouter()),
			onDemandListener: entryPoint.OnDemandListener,
			tlsALPNGetter:    entryPoint.TLSALPNGetter,
		}

		if entryPoint.CertificateStore != nil {
			serverEntryPoints[entryPointName].certs = entryPoint.CertificateStore
		} else {
			serverEntryPoints[entryPointName].certs = traefiktls.NewCertificateStore()
		}

		if entryPoint.Configuration.TLS != nil {
			serverEntryPoints[entryPointName].certs.SniStrict = entryPoint.Configuration.TLS.SniStrict

			if entryPoint.Configuration.TLS.DefaultCertificate != nil {
				cert, err := buildDefaultCertificate(entryPoint.Configuration.TLS.DefaultCertificate)
				if err != nil {
					log.Error(err)
					continue
				}
				serverEntryPoints[entryPointName].certs.DefaultCertificate = cert
			} else {
				cert, err := generate.DefaultCertificate()
				if err != nil {
					log.Errorf("failed to generate default certificate: %v", err)
					continue
				}
				serverEntryPoints[entryPointName].certs.DefaultCertificate = cert
			}
			if len(entryPoint.Configuration.TLS.Certificates) > 0 {
				config, _ := entryPoint.Configuration.TLS.Certificates.CreateTLSConfig(entryPointName)
				certMap := s.buildNameOrIPToCertificate(config.Certificates)
				serverEntryPoints[entryPointName].certs.StaticCerts.Set(certMap)

			}
		}
	}
	return serverEntryPoints
}

func buildDefaultCertificate(defaultCertificate *traefiktls.Certificate) (*tls.Certificate, error) {
	certFile, err := defaultCertificate.CertFile.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to get cert file content: %v", err)
	}

	keyFile, err := defaultCertificate.KeyFile.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to get key file content: %v", err)
	}

	cert, err := tls.X509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load X509 key pair: %v", err)
	}
	return &cert, nil
}

func (s *Server) buildDefaultHTTPRouter() *mux.Router {
	rt := mux.NewRouter()
	rt.NotFoundHandler = s.wrapHTTPHandlerWithAccessLog(http.HandlerFunc(http.NotFound), "backend not found")
	rt.StrictSlash(!s.globalConfiguration.KeepTrailingSlash)
	rt.SkipClean(true)
	return rt
}

func sortedFrontendNamesForConfig(configuration *types.Configuration) []string {
	var keys []string
	for key := range configuration.Frontends {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func buildHostResolver(globalConfig configuration.GlobalConfiguration) *hostresolver.Resolver {
	if globalConfig.HostResolver != nil {
		return &hostresolver.Resolver{
			CnameFlattening: globalConfig.HostResolver.CnameFlattening,
			ResolvConfig:    globalConfig.HostResolver.ResolvConfig,
			ResolvDepth:     globalConfig.HostResolver.ResolvDepth,
		}
	}
	return nil
}
