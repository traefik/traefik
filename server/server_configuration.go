package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/containous/mux"
	"github.com/containous/traefik/configuration"
	"github.com/containous/traefik/healthcheck"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/metrics"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/middlewares/errorpages"
	"github.com/containous/traefik/rules"
	"github.com/containous/traefik/safe"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/eapache/channels"
	"github.com/urfave/negroni"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/utils"
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

	newServerEntryPoints, err := s.loadConfig(newConfigurations, s.globalConfiguration)
	if err != nil {
		s.metricsRegistry.ConfigReloadsFailureCounter().Add(1)
		s.metricsRegistry.LastConfigReloadFailureGauge().Set(float64(time.Now().Unix()))
		log.Error("Error loading new configuration, aborted ", err)
		return
	}

	s.metricsRegistry.LastConfigReloadSuccessGauge().Set(float64(time.Now().Unix()))

	for newServerEntryPointName, newServerEntryPoint := range newServerEntryPoints {
		s.serverEntryPoints[newServerEntryPointName].httpRouter.UpdateHandler(newServerEntryPoint.httpRouter.GetHandler())

		if s.entryPoints[newServerEntryPointName].Configuration.TLS == nil {
			if newServerEntryPoint.certs.Get() != nil {
				log.Debugf("Certificates not added to non-TLS entryPoint %s.", newServerEntryPointName)
			}
		} else {
			s.serverEntryPoints[newServerEntryPointName].certs.Set(newServerEntryPoint.certs.Get())
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
func (s *Server) loadConfig(configurations types.Configurations, globalConfiguration configuration.GlobalConfiguration) (map[string]*serverEntryPoint, error) {
	serverEntryPoints := s.buildServerEntryPoints()
	redirectHandlers := make(map[string]negroni.Handler)
	backendsHandlers := map[string]http.Handler{}
	backendsHealthCheck := map[string]*healthcheck.BackendConfig{}
	var errorPageHandlers []*errorpages.Handler

	errorHandler := NewRecordingErrorHandler(middlewares.DefaultNetErrorRecorder{})

	redirectHandlers, err := s.buildEntryPointRedirect()
	if err != nil {
		return nil, err
	}

	for providerName, config := range configurations {
		frontendNames := sortedFrontendNamesForConfig(config)
	frontend:
		for _, frontendName := range frontendNames {
			frontend := config.Frontends[frontendName]

			log.Debugf("Creating frontend %s", frontendName)

			if len(frontend.EntryPoints) == 0 {
				log.Errorf("no entrypoint defined for frontend %s", frontendName)
				log.Errorf("Skipping frontend %s...", frontendName)
				continue frontend
			}

			for _, entryPointName := range frontend.EntryPoints {
				log.Debugf("Wiring frontend %s to entryPoint %s", frontendName, entryPointName)

				entryPoint := s.entryPoints[entryPointName].Configuration

				n := negroni.New()

				if _, exist := redirectHandlers[entryPointName]; exist {
					n.Use(redirectHandlers[entryPointName])
				}

				frontendHash, err := frontend.Hash()
				if err != nil {
					log.Errorf("Error calculating hash value for frontend %s: %v", frontendName, err)
					log.Errorf("Skipping frontend %s...", frontendName)
					continue frontend
				}

				if backendsHandlers[entryPointName+providerName+frontendHash] == nil {
					log.Debugf("Creating backend %s", frontend.Backend)

					if config.Backends[frontend.Backend] == nil {
						log.Errorf("Undefined backend '%s' for frontend %s", frontend.Backend, frontendName)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					headerMiddleware := middlewares.NewHeaderFromStruct(frontend.Headers)
					secureMiddleware := middlewares.NewSecure(frontend.Headers)

					if len(frontend.Errors) > 0 {
						handlers, err := buildErrorPagesMiddleware(frontendName, frontend, config.Backends, entryPointName, providerName)
						if err != nil {
							log.Error(err)
						} else {
							errorPageHandlers = append(errorPageHandlers, handlers...)
							for _, handler := range handlers {
								n.Use(handler)
							}
						}
					}

					if s.metricsRegistry.IsEnabled() {
						n.Use(middlewares.NewBackendMetricsMiddleware(s.metricsRegistry, frontend.Backend))
					}

					ipWhitelistMiddleware, err := buildIPWhiteLister(frontend.WhiteList, frontend.WhitelistSourceRange)
					if err != nil {
						log.Errorf("Error creating IP Whitelister: %s", err)
					} else if ipWhitelistMiddleware != nil {
						n.Use(
							s.tracingMiddleware.NewNegroniHandlerWrapper(
								"IP whitelist",
								s.wrapNegroniHandlerWithAccessLog(ipWhitelistMiddleware, fmt.Sprintf("ipwhitelister for %s", frontendName)),
								false))
						log.Debugf("Configured IP Whitelists: %s", frontend.WhitelistSourceRange)
					}

					if frontend.Redirect != nil && entryPointName != frontend.Redirect.EntryPoint {
						rewrite, err := s.buildRedirectHandler(entryPointName, frontend.Redirect)
						if err != nil {
							log.Errorf("Error creating Frontend Redirect: %v", err)
						} else {
							n.Use(s.wrapNegroniHandlerWithAccessLog(rewrite, fmt.Sprintf("frontend redirect for %s", frontendName)))
							log.Debugf("Frontend %s redirect created", frontendName)
						}
					}

					if headerMiddleware != nil {
						log.Debugf("Adding header middleware for frontend %s", frontendName)
						n.Use(s.tracingMiddleware.NewNegroniHandlerWrapper("Header", headerMiddleware, false))
					}

					if secureMiddleware != nil {
						log.Debugf("Adding secure middleware for frontend %s", frontendName)
						n.UseFunc(secureMiddleware.HandlerFuncWithNextForRequestOnly)
					}

					if len(frontend.BasicAuth) > 0 {
						authMiddleware, err := s.buildBasicAuthMiddleware(frontend.BasicAuth)
						if err != nil {
							log.Errorf("Error creating Auth: %s", err)
						} else {
							n.Use(s.wrapNegroniHandlerWithAccessLog(authMiddleware, fmt.Sprintf("Auth for %s", frontendName)))
						}
					}

					// LB

					var responseModifier = buildModifyResponse(secureMiddleware, headerMiddleware)

					fwd, err := s.buildForwarder(entryPointName, entryPoint, frontendName, frontend, errorHandler, responseModifier)
					if err != nil {
						log.Errorf("Failed to create the forwarder for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					balancer, err := s.buildLoadBalancer(frontendName, frontend.Backend, config.Backends[frontend.Backend], fwd)
					if err != nil {
						log.Errorf("Failed to create the load-balancer for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
					}

					// Health Check
					if hcOpts := parseHealthCheckOptions(balancer, frontend.Backend, config.Backends[frontend.Backend].HealthCheck, s.globalConfiguration.HealthCheck); hcOpts != nil {
						log.Debugf("Setting up backend health check %s", *hcOpts)

						hcOpts.Transport = s.defaultForwardingRoundTripper
						backendsHealthCheck[entryPointName+providerName+frontendHash] = healthcheck.NewBackendConfig(*hcOpts, frontend.Backend)
					}

					// Empty (backend with no servers)
					var lb http.Handler = middlewares.NewEmptyBackendHandler(balancer)

					if frontend.RateLimit != nil && len(frontend.RateLimit.RateSet) > 0 {
						lb, err = buildRateLimiter(lb, frontend.RateLimit)
						if err != nil {
							log.Errorf("Error creating rate limiter: %v", err)
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}

						lb = s.wrapHTTPHandlerWithAccessLog(
							s.tracingMiddleware.NewHTTPHandlerWrapper("Rate limit", lb, false),
							fmt.Sprintf("rate limit for %s", frontendName),
						)
					}

					maxConns := config.Backends[frontend.Backend].MaxConn
					if maxConns != nil && maxConns.Amount != 0 {
						log.Debugf("Creating load-balancer connection limit")

						handler, err := buildMaxConn(lb, maxConns)
						if err != nil {
							log.Error(err)
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						} else {
							lb = s.wrapHTTPHandlerWithAccessLog(handler, fmt.Sprintf("connection limit for %s", frontendName))
						}
					}

					if globalConfiguration.Retry != nil {
						countServers := len(config.Backends[frontend.Backend].Servers)
						lb = s.buildRetryMiddleware(lb, globalConfiguration.Retry, countServers, frontend.Backend)
						lb = s.tracingMiddleware.NewHTTPHandlerWrapper("Retry", lb, false)
					}

					if config.Backends[frontend.Backend].Buffering != nil {
						bufferedLb, err := buildBufferingMiddleware(lb, config.Backends[frontend.Backend].Buffering)
						if err != nil {
							log.Errorf("Error setting up buffering middleware: %s", err)
						} else {
							lb = bufferedLb
						}
					}

					if config.Backends[frontend.Backend].CircuitBreaker != nil {
						log.Debugf("Creating circuit breaker %s", config.Backends[frontend.Backend].CircuitBreaker.Expression)
						expression := config.Backends[frontend.Backend].CircuitBreaker.Expression
						circuitBreaker, err := middlewares.NewCircuitBreaker(lb, expression, middlewares.NewCircuitBreakerOptions(expression))
						if err != nil {
							log.Errorf("Error creating circuit breaker: %v", err)
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
						lb = s.tracingMiddleware.NewHTTPHandlerWrapper("Circuit breaker", circuitBreaker, false)
					}

					n.UseHandler(lb)

					backendsHandlers[entryPointName+providerName+frontendHash] = n
				} else {
					log.Debugf("Reusing backend %s", frontend.Backend)
				}

				serverRoute, err := buildServerRoute(serverEntryPoints[entryPointName], frontendName, frontend)
				if err != nil {
					log.Error(err)
					log.Errorf("Skipping frontend %s...", frontendName)
					continue frontend
				}

				s.wireFrontendBackend(serverRoute, backendsHandlers[entryPointName+providerName+frontendHash])

				err = serverRoute.Route.GetError()
				if err != nil {
					log.Errorf("Error building route: %s", err)
				}
			}
		}
	}

	for _, errorPageHandler := range errorPageHandlers {
		if handler, ok := backendsHandlers[errorPageHandler.BackendName]; ok {
			errorPageHandler.PostLoad(handler)
		} else {
			errorPageHandler.PostLoad(nil)
		}
	}

	healthcheck.GetHealthCheck(s.metricsRegistry).SetBackendsConfiguration(s.routinesPool.Ctx(), backendsHealthCheck)

	// Get new certificates list sorted per entrypoints
	// Update certificates
	entryPointsCertificates, err := s.loadHTTPSConfiguration(configurations, globalConfiguration.DefaultEntryPoints)
	// FIXME error management

	// Sort routes and update certificates
	for serverEntryPointName, serverEntryPoint := range serverEntryPoints {
		serverEntryPoint.httpRouter.GetHandler().SortRoutes()
		if _, exists := entryPointsCertificates[serverEntryPointName]; exists {
			serverEntryPoint.certs.Set(entryPointsCertificates[serverEntryPointName])
		}
	}

	return serverEntryPoints, err
}

func (s *Server) buildForwarder(entryPointName string, entryPoint *configuration.EntryPoint,
	frontendName string, frontend *types.Frontend,
	errorHandler utils.ErrorHandler, responseModifier modifyResponse) (http.Handler, error) {

	roundTripper, err := s.getRoundTripper(entryPointName, frontend.PassTLSCert, entryPoint.TLS)
	if err != nil {
		return nil, fmt.Errorf("failed to create RoundTripper for frontend %s: %v", frontendName, err)
	}

	rewriter, err := NewHeaderRewriter(entryPoint.ForwardedHeaders.TrustedIPs, entryPoint.ForwardedHeaders.Insecure)
	if err != nil {
		return nil, fmt.Errorf("error creating rewriter for frontend %s: %v", frontendName, err)
	}

	var fwd http.Handler
	fwd, err = forward.New(
		forward.Stream(true),
		forward.PassHostHeader(frontend.PassHostHeader),
		forward.RoundTripper(roundTripper),
		forward.ErrorHandler(errorHandler),
		forward.Rewriter(rewriter),
		forward.ResponseModifier(responseModifier),
		forward.BufferPool(s.bufferPool),
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

	return fwd, nil
}

func buildServerRoute(serverEntryPoint *serverEntryPoint, frontendName string, frontend *types.Frontend) (*types.ServerRoute, error) {
	serverRoute := &types.ServerRoute{Route: serverEntryPoint.httpRouter.GetHandler().NewRoute().Name(frontendName)}

	priority := 0
	for routeName, route := range frontend.Routes {
		rls := rules.Rules{Route: serverRoute}
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

	jsonConf, _ := json.Marshal(configMsg.Configuration)

	log.Debugf("Configuration received from provider %s: %s", configMsg.ProviderName, string(jsonConf))

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
				publish <- nextConfig.(types.ConfigMessage)
				time.Sleep(throttle)
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

func (s *Server) wireFrontendBackend(serverRoute *types.ServerRoute, handler http.Handler) {
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

	serverRoute.Route.Handler(handler)
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
func (s *Server) loadHTTPSConfiguration(configurations types.Configurations, defaultEntryPoints configuration.DefaultEntryPoints) (map[string]map[string]*tls.Certificate, error) {
	newEPCertificates := make(map[string]map[string]*tls.Certificate)
	// Get all certificates
	for _, config := range configurations {
		if config.TLS != nil && len(config.TLS) > 0 {
			if err := traefiktls.SortTLSPerEntryPoints(config.TLS, newEPCertificates, defaultEntryPoints); err != nil {
				return nil, err
			}
		}
	}
	return newEPCertificates, nil
}

func (s *Server) buildServerEntryPoints() map[string]*serverEntryPoint {
	serverEntryPoints := make(map[string]*serverEntryPoint)
	for entryPointName, entryPoint := range s.entryPoints {
		serverEntryPoints[entryPointName] = &serverEntryPoint{
			httpRouter:       middlewares.NewHandlerSwitcher(s.buildDefaultHTTPRouter()),
			onDemandListener: entryPoint.OnDemandListener,
		}
		if entryPoint.CertificateStore != nil {
			serverEntryPoints[entryPointName].certs = entryPoint.CertificateStore.DynamicCerts
		} else {
			serverEntryPoints[entryPointName].certs = &safe.Safe{}
		}
	}
	return serverEntryPoints
}

func (s *Server) buildDefaultHTTPRouter() *mux.Router {
	rt := mux.NewRouter()
	rt.NotFoundHandler = s.wrapHTTPHandlerWithAccessLog(http.HandlerFunc(notFoundHandler), "backend not found")
	rt.StrictSlash(true)
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

func parseHealthCheckOptions(lb healthcheck.BalancerHandler, backend string, hc *types.HealthCheck, hcConfig *configuration.HealthCheckConfig) *healthcheck.Options {
	if hc == nil || hc.Path == "" || hcConfig == nil {
		return nil
	}

	interval := time.Duration(hcConfig.Interval)
	if hc.Interval != "" {
		intervalOverride, err := time.ParseDuration(hc.Interval)
		switch {
		case err != nil:
			log.Errorf("Illegal health check interval for backend '%s': %s", backend, err)
		case intervalOverride <= 0:
			log.Errorf("Health check interval smaller than zero for backend '%s', backend", backend)
		default:
			interval = intervalOverride
		}
	}

	return &healthcheck.Options{
		Scheme:   hc.Scheme,
		Path:     hc.Path,
		Port:     hc.Port,
		Interval: interval,
		LB:       lb,
		Hostname: hc.Hostname,
		Headers:  hc.Headers,
	}
}
