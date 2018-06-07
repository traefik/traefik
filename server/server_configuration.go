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
	"github.com/containous/traefik/middlewares/accesslog"
	mauth "github.com/containous/traefik/middlewares/auth"
	"github.com/containous/traefik/middlewares/errorpages"
	"github.com/containous/traefik/rules"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/server/cookie"
	traefiktls "github.com/containous/traefik/tls"
	"github.com/containous/traefik/types"
	"github.com/eapache/channels"
	"github.com/urfave/negroni"
	"github.com/vulcand/oxy/connlimit"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
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
	if err == nil {
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
	} else {
		s.metricsRegistry.ConfigReloadsFailureCounter().Add(1)
		s.metricsRegistry.LastConfigReloadFailureGauge().Set(float64(time.Now().Unix()))
		log.Error("Error loading new configuration, aborted ", err)
	}
}

// loadConfig returns a new gorilla.mux Route from the specified global configuration and the dynamic
// provider configurations.
func (s *Server) loadConfig(configurations types.Configurations, globalConfiguration configuration.GlobalConfiguration) (map[string]*serverEntryPoint, error) {
	serverEntryPoints := s.buildEntryPoints()
	redirectHandlers := make(map[string]negroni.Handler)
	backends := map[string]http.Handler{}
	backendsHealthCheck := map[string]*healthcheck.BackendConfig{}
	var errorPageHandlers []*errorpages.Handler

	errorHandler := NewRecordingErrorHandler(middlewares.DefaultNetErrorRecorder{})

	for providerName, config := range configurations {
		frontendNames := sortedFrontendNamesForConfig(config)
	frontend:
		for _, frontendName := range frontendNames {
			frontend := config.Frontends[frontendName]

			log.Debugf("Creating frontend %s", frontendName)

			var frontendEntryPoints []string
			for _, entryPointName := range frontend.EntryPoints {
				if _, ok := serverEntryPoints[entryPointName]; !ok {
					log.Errorf("Undefined entrypoint '%s' for frontend %s", entryPointName, frontendName)
				} else {
					frontendEntryPoints = append(frontendEntryPoints, entryPointName)
				}
			}
			frontend.EntryPoints = frontendEntryPoints

			if len(frontend.EntryPoints) == 0 {
				log.Errorf("No entrypoint defined for frontend %s", frontendName)
				log.Errorf("Skipping frontend %s...", frontendName)
				continue frontend
			}
			for _, entryPointName := range frontend.EntryPoints {
				log.Debugf("Wiring frontend %s to entryPoint %s", frontendName, entryPointName)

				newServerRoute := &types.ServerRoute{Route: serverEntryPoints[entryPointName].httpRouter.GetHandler().NewRoute().Name(frontendName)}
				for routeName, route := range frontend.Routes {
					err := getRoute(newServerRoute, &route)
					if err != nil {
						log.Errorf("Error creating route for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}
					log.Debugf("Creating route %s %s", routeName, route.Rule)
				}

				entryPoint := s.entryPoints[entryPointName].Configuration
				n := negroni.New()
				if entryPoint.Redirect != nil && entryPointName != entryPoint.Redirect.EntryPoint {
					if redirectHandlers[entryPointName] != nil {
						n.Use(redirectHandlers[entryPointName])
					} else if handler, err := s.buildRedirectHandler(entryPointName, entryPoint.Redirect); err != nil {
						log.Errorf("Error loading entrypoint configuration for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					} else {
						handlerToUse := s.wrapNegroniHandlerWithAccessLog(handler, fmt.Sprintf("entrypoint redirect for %s", frontendName))
						n.Use(handlerToUse)
						redirectHandlers[entryPointName] = handlerToUse
					}
				}

				frontendHash, err := frontend.Hash()
				if err != nil {
					log.Errorf("Error calculating hash value for frontend %s: %v", frontendName, err)
					log.Errorf("Skipping frontend %s...", frontendName)
					continue frontend
				}
				backendCacheKey := entryPointName + providerName + frontendHash
				if backends[backendCacheKey] == nil {
					log.Debugf("Creating backend %s", frontend.Backend)

					roundTripper, err := s.getRoundTripper(entryPointName, globalConfiguration, frontend.PassTLSCert, entryPoint.TLS)
					if err != nil {
						log.Errorf("Failed to create RoundTripper for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					rewriter, err := NewHeaderRewriter(entryPoint.ForwardedHeaders.TrustedIPs, entryPoint.ForwardedHeaders.Insecure)
					if err != nil {
						log.Errorf("Error creating rewriter for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					headerMiddleware := middlewares.NewHeaderFromStruct(frontend.Headers)
					secureMiddleware := middlewares.NewSecure(frontend.Headers)

					var responseModifier = buildModifyResponse(secureMiddleware, headerMiddleware)
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
						log.Errorf("Error creating forwarder for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					if s.tracingMiddleware.IsEnabled() {
						tm := s.tracingMiddleware.NewForwarderMiddleware(frontendName, frontend.Backend)

						next := fwd
						fwd = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							tm.ServeHTTP(w, r, next.ServeHTTP)
						})
					}

					var rr *roundrobin.RoundRobin
					var saveFrontend http.Handler
					if s.accessLoggerMiddleware != nil {
						saveBackend := accesslog.NewSaveBackend(fwd, frontend.Backend)
						saveFrontend = accesslog.NewSaveFrontend(saveBackend, frontendName)
						rr, _ = roundrobin.New(saveFrontend)
					} else {
						rr, _ = roundrobin.New(fwd)
					}

					if config.Backends[frontend.Backend] == nil {
						log.Errorf("Undefined backend '%s' for frontend %s", frontend.Backend, frontendName)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					lbMethod, err := types.NewLoadBalancerMethod(config.Backends[frontend.Backend].LoadBalancer)
					if err != nil {
						log.Errorf("Error loading load balancer method '%+v' for frontend %s: %v", config.Backends[frontend.Backend].LoadBalancer, frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					var sticky *roundrobin.StickySession
					var cookieName string
					if stickiness := config.Backends[frontend.Backend].LoadBalancer.Stickiness; stickiness != nil {
						cookieName = cookie.GetName(stickiness.CookieName, frontend.Backend)
						sticky = roundrobin.NewStickySession(cookieName)
					}

					var lb http.Handler
					switch lbMethod {
					case types.Drr:
						log.Debugf("Creating load-balancer drr")
						rebalancer, _ := roundrobin.NewRebalancer(rr)
						if sticky != nil {
							log.Debugf("Sticky session with cookie %v", cookieName)
							rebalancer, _ = roundrobin.NewRebalancer(rr, roundrobin.RebalancerStickySession(sticky))
						}
						if err := s.configureLBServers(rebalancer, config, frontend); err != nil {
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
						hcOpts := parseHealthCheckOptions(rebalancer, frontend.Backend, config.Backends[frontend.Backend].HealthCheck, globalConfiguration.HealthCheck)
						if hcOpts != nil {
							log.Debugf("Setting up backend health check %s", *hcOpts)
							hcOpts.Transport = s.defaultForwardingRoundTripper
							backendsHealthCheck[backendCacheKey] = healthcheck.NewBackendConfig(*hcOpts, frontend.Backend)
						}
						lb = middlewares.NewEmptyBackendHandler(rebalancer)
					case types.Wrr:
						log.Debugf("Creating load-balancer wrr")
						if sticky != nil {
							log.Debugf("Sticky session with cookie %v", cookieName)
							if s.accessLoggerMiddleware != nil {
								rr, _ = roundrobin.New(saveFrontend, roundrobin.EnableStickySession(sticky))
							} else {
								rr, _ = roundrobin.New(fwd, roundrobin.EnableStickySession(sticky))
							}
						}
						if err := s.configureLBServers(rr, config, frontend); err != nil {
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
						hcOpts := parseHealthCheckOptions(rr, frontend.Backend, config.Backends[frontend.Backend].HealthCheck, globalConfiguration.HealthCheck)
						if hcOpts != nil {
							log.Debugf("Setting up backend health check %s", *hcOpts)
							hcOpts.Transport = s.defaultForwardingRoundTripper
							backendsHealthCheck[backendCacheKey] = healthcheck.NewBackendConfig(*hcOpts, frontend.Backend)
						}
						lb = middlewares.NewEmptyBackendHandler(rr)
					}

					if len(frontend.Errors) > 0 {
						for errorPageName, errorPage := range frontend.Errors {
							if frontend.Backend == errorPage.Backend {
								log.Errorf("Error when creating error page %q for frontend %q: error pages backend %q is the same as backend for the frontend (infinite call risk).",
									errorPageName, frontendName, errorPage.Backend)
							} else if config.Backends[errorPage.Backend] == nil {
								log.Errorf("Error when creating error page %q for frontend %q: the backend %q doesn't exist.",
									errorPageName, frontendName, errorPage.Backend)
							} else {
								errorPagesHandler, err := errorpages.NewHandler(errorPage, entryPointName+providerName+errorPage.Backend)
								if err != nil {
									log.Errorf("Error creating error pages: %v", err)
								} else {
									if errorPageServer, ok := config.Backends[errorPage.Backend].Servers["error"]; ok {
										errorPagesHandler.FallbackURL = errorPageServer.URL
									}

									errorPageHandlers = append(errorPageHandlers, errorPagesHandler)
									n.Use(errorPagesHandler)
								}
							}
						}
					}

					if frontend.RateLimit != nil && len(frontend.RateLimit.RateSet) > 0 {
						lb, err = s.buildRateLimiter(lb, frontend.RateLimit)
						if err != nil {
							log.Errorf("Error creating rate limiter: %v", err)
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
						lb = s.wrapHTTPHandlerWithAccessLog(lb, fmt.Sprintf("rate limit for %s", frontendName))
					}

					maxConns := config.Backends[frontend.Backend].MaxConn
					if maxConns != nil && maxConns.Amount != 0 {
						extractFunc, err := utils.NewExtractor(maxConns.ExtractorFunc)
						if err != nil {
							log.Errorf("Error creating connection limit: %v", err)
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}

						log.Debugf("Creating load-balancer connection limit")

						lb, err = connlimit.New(lb, extractFunc, maxConns.Amount)
						if err != nil {
							log.Errorf("Error creating connection limit: %v", err)
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
						lb = s.wrapHTTPHandlerWithAccessLog(lb, fmt.Sprintf("connection limit for %s", frontendName))
					}

					if globalConfiguration.Retry != nil {
						countServers := len(config.Backends[frontend.Backend].Servers)
						lb = s.buildRetryMiddleware(lb, globalConfiguration, countServers, frontend.Backend)
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
						users := types.Users{}
						for _, user := range frontend.BasicAuth {
							users = append(users, user)
						}

						auth := &types.Auth{}
						auth.Basic = &types.Basic{
							Users: users,
						}
						authMiddleware, err := mauth.NewAuthenticator(auth, s.tracingMiddleware)
						if err != nil {
							log.Errorf("Error creating Auth: %s", err)
						} else {
							n.Use(s.wrapNegroniHandlerWithAccessLog(authMiddleware, fmt.Sprintf("Auth for %s", frontendName)))
						}
					}

					if config.Backends[frontend.Backend].Buffering != nil {
						bufferedLb, err := s.buildBufferingMiddleware(lb, config.Backends[frontend.Backend].Buffering)

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
						n.Use(negroni.Wrap(s.tracingMiddleware.NewHTTPHandlerWrapper("Circuit breaker", circuitBreaker, false)))
					} else {
						n.UseHandler(lb)
					}
					backends[backendCacheKey] = n
				} else {
					log.Debugf("Reusing backend %s", frontend.Backend)
				}
				if frontend.Priority > 0 {
					newServerRoute.Route.Priority(frontend.Priority)
				}
				s.wireFrontendBackend(newServerRoute, backends[backendCacheKey])

				err = newServerRoute.Route.GetError()
				if err != nil {
					log.Errorf("Error building route: %s", err)
				}
			}
		}
	}

	for _, errorPageHandler := range errorPageHandlers {
		if handler, ok := backends[errorPageHandler.BackendName]; ok {
			errorPageHandler.PostLoad(handler)
		} else {
			errorPageHandler.PostLoad(nil)
		}
	}

	healthcheck.GetHealthCheck(s.metricsRegistry).SetBackendsConfiguration(s.routinesPool.Ctx(), backendsHealthCheck)

	// Get new certificates list sorted per entrypoints
	// Update certificates
	entryPointsCertificates, err := s.loadHTTPSConfiguration(configurations, globalConfiguration.DefaultEntryPoints)

	// Sort routes and update certificates
	for serverEntryPointName, serverEntryPoint := range serverEntryPoints {
		serverEntryPoint.httpRouter.GetHandler().SortRoutes()
		if _, exists := entryPointsCertificates[serverEntryPointName]; exists {
			serverEntryPoint.certs.Set(entryPointsCertificates[serverEntryPointName])
		}
	}

	return serverEntryPoints, err
}

func (s *Server) preLoadConfiguration(configMsg types.ConfigMessage) {
	providersThrottleDuration := time.Duration(s.globalConfiguration.ProvidersThrottleDuration)
	s.defaultConfigurationValues(configMsg.Configuration)
	currentConfigurations := s.currentConfigurations.Get().(types.Configurations)
	jsonConf, _ := json.Marshal(configMsg.Configuration)
	log.Debugf("Configuration received from provider %s: %s", configMsg.ProviderName, string(jsonConf))
	if configMsg.Configuration == nil || configMsg.Configuration.Backends == nil && configMsg.Configuration.Frontends == nil && configMsg.Configuration.TLS == nil {
		log.Infof("Skipping empty Configuration for provider %s", configMsg.ProviderName)
	} else if reflect.DeepEqual(currentConfigurations[configMsg.ProviderName], configMsg.Configuration) {
		log.Infof("Skipping same configuration for provider %s", configMsg.ProviderName)
	} else {
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
}

func (s *Server) defaultConfigurationValues(configuration *types.Configuration) {
	if configuration == nil || configuration.Frontends == nil {
		return
	}
	configureFrontends(configuration.Frontends, s.globalConfiguration.DefaultEntryPoints)
	configureBackends(configuration.Backends)
}

func configureFrontends(frontends map[string]*types.Frontend, defaultEntrypoints []string) {
	for _, frontend := range frontends {
		// default endpoints if not defined in frontends
		if len(frontend.EntryPoints) == 0 {
			frontend.EntryPoints = defaultEntrypoints
		}
	}
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
						rules := rules.Rules{}
						domains, err := rules.ParseDomains(route.Rule)
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
	for _, configuration := range configurations {
		if configuration.TLS != nil && len(configuration.TLS) > 0 {
			if err := traefiktls.SortTLSPerEntryPoints(configuration.TLS, newEPCertificates, defaultEntryPoints); err != nil {
				return nil, err
			}
		}
	}
	return newEPCertificates, nil
}

func (s *Server) buildEntryPoints() map[string]*serverEntryPoint {
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

func getRoute(serverRoute *types.ServerRoute, route *types.Route) error {
	rules := rules.Rules{Route: serverRoute}
	newRoute, err := rules.Parse(route.Rule)
	if err != nil {
		return err
	}
	newRoute.Priority(serverRoute.Route.GetPriority() + len(route.Rule))
	serverRoute.Route = newRoute
	return nil
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
