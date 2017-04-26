package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/containous/mux"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/healthcheck"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/middlewares/accesslog"
	"github.com/containous/traefik/middlewares/common"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/streamrail/concurrent-map"
	"github.com/vulcand/oxy/cbreaker"
	"github.com/vulcand/oxy/connlimit"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
	"github.com/vulcand/oxy/utils"
)

var oxyLogger = &OxyLogger{}

// Server is the reverse-proxy/load-balancer engine
type Server struct {
	serverEntryPoints          serverEntryPoints
	configurationChan          chan types.ConfigMessage
	configurationValidatedChan chan types.ConfigMessage
	signals                    chan os.Signal
	stopChan                   chan bool
	providers                  []provider.Provider
	currentConfigurations      safe.Safe
	globalConfiguration        GlobalConfiguration
	accessLogFile              *os.File
	routinesPool               *safe.Pool
	leadership                 *cluster.Leadership
}

type serverEntryPoints map[string]*serverEntryPoint

type serverEntryPoint struct {
	httpServer *http.Server
	httpRouter *middlewares.HandlerSwitcher
}

type serverRoute struct {
	route         *mux.Route
	stripPrefixes []string
	addPrefix     string
}

// NewServer returns an initialized Server.
func NewServer(globalConfiguration GlobalConfiguration) *Server {
	server := new(Server)

	server.serverEntryPoints = make(map[string]*serverEntryPoint)
	server.configurationChan = make(chan types.ConfigMessage, 100)
	server.configurationValidatedChan = make(chan types.ConfigMessage, 100)
	server.signals = make(chan os.Signal, 1)
	server.stopChan = make(chan bool, 1)
	server.providers = []provider.Provider{}
	signal.Notify(server.signals, syscall.SIGINT, syscall.SIGTERM)
	currentConfigurations := make(configs)
	server.currentConfigurations.Set(currentConfigurations)
	server.globalConfiguration = globalConfiguration
	server.routinesPool = safe.NewPool(context.Background())
	if globalConfiguration.Cluster != nil {
		// leadership creation if cluster mode
		server.leadership = cluster.NewLeadership(server.routinesPool.Ctx(), globalConfiguration.Cluster)
	}

	if server.globalConfiguration.AccessLogsFile != "" {
		var err error
		server.accessLogFile, err = os.OpenFile(server.globalConfiguration.AccessLogsFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Error("Error opening file", err)
		}
	}

	return server
}

// Start starts the server.
func (server *Server) Start() {
	server.startHTTPServers()
	server.startLeadership()
	server.routinesPool.Go(func(stop chan bool) {
		server.listenProviders(stop)
	})
	server.routinesPool.Go(func(stop chan bool) {
		server.listenConfigurations(stop)
	})
	server.configureProviders()
	server.startProviders()
	go server.listenSignals()
}

// Wait blocks until server is shutted down.
func (server *Server) Wait() {
	<-server.stopChan
}

// Stop stops the server
func (server *Server) Stop() {
	defer log.Info("Server stopped")
	var wg sync.WaitGroup
	for sepn, sep := range server.serverEntryPoints {
		wg.Add(1)
		go func(serverEntryPointName string, serverEntryPoint *serverEntryPoint) {
			defer wg.Done()
			graceTimeOut := time.Duration(server.globalConfiguration.GraceTimeOut)
			ctx, cancel := context.WithTimeout(context.Background(), graceTimeOut)
			log.Debugf("Waiting %s seconds before killing connections on entrypoint %s...", graceTimeOut, serverEntryPointName)
			if err := serverEntryPoint.httpServer.Shutdown(ctx); err != nil {
				log.Debugf("Wait is over due to: %s", err)
				serverEntryPoint.httpServer.Close()
			}
			cancel()
			log.Debugf("Entrypoint %s closed", serverEntryPointName)

			if m, ok := serverEntryPoint.httpServer.Handler.(common.Middleware); ok {
				m.Close()
			}

		}(sepn, sep)
	}
	wg.Wait()
	server.stopChan <- true
}

// Close destroys the server
func (server *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(server.globalConfiguration.GraceTimeOut))
	go func(ctx context.Context) {
		<-ctx.Done()
		if ctx.Err() == context.Canceled {
			return
		} else if ctx.Err() == context.DeadlineExceeded {
			log.Warnf("Timeout while stopping traefik, killing instance ✝")
			os.Exit(1)
		}
	}(ctx)
	server.stopLeadership()
	server.routinesPool.Cleanup()
	close(server.configurationChan)
	close(server.configurationValidatedChan)
	signal.Stop(server.signals)
	close(server.signals)
	close(server.stopChan)
	cancel()
}

func (server *Server) startLeadership() {
	if server.leadership != nil {
		server.leadership.Participate(server.routinesPool)
		// server.leadership.AddGoCtx(func(ctx context.Context) {
		// 	log.Debugf("Started test routine")
		// 	<-ctx.Done()
		// 	log.Debugf("Stopped test routine")
		// })
	}
}

func (server *Server) stopLeadership() {
	if server.leadership != nil {
		server.leadership.Stop()
	}
}

func (server *Server) startHTTPServers() {
	server.serverEntryPoints = server.buildEntryPoints(server.globalConfiguration)

	for newServerEntryPointName, newServerEntryPoint := range server.serverEntryPoints {
		// build a filter chain of: log, metrics (wrapping prometheus), stats, authenticator, compress, router
		// (working right to left)

		var chain http.Handler = newServerEntryPoint.httpRouter

		if server.globalConfiguration.EntryPoints[newServerEntryPointName].Compress {
			chain = middlewares.NewCompress(chain)
		}

		if server.globalConfiguration.EntryPoints[newServerEntryPointName].Auth != nil {
			var err error
			authConfig := server.globalConfiguration.EntryPoints[newServerEntryPointName].Auth
			chain, err = middlewares.NewAuthenticator(authConfig, chain)
			if err != nil {
				log.Fatal("Error starting server: ", err)
			}
		}

		web := server.globalConfiguration.Web
		if web != nil && web.Statistics != nil {
			chain = middlewares.NewStatsRecorder(web.Statistics.RecentErrors, chain)
		}

		if web != nil && web.Metrics != nil && web.Metrics.Prometheus != nil {
			p := middlewares.NewPrometheus(newServerEntryPointName, web.Metrics.Prometheus)
			chain = middlewares.NewMetricsWrapper(p, chain)
		}

		chain = common.NewAdapter(metrics, chain)

		chain = middlewares.NewLogger(server.accessLogFile, chain)

		chain = accesslog.NewLogHandler(chain)

		entryPoint := server.globalConfiguration.EntryPoints[newServerEntryPointName]
		newsrv, err := server.prepareServer(newServerEntryPointName, entryPoint, chain, newServerEntryPoint.httpRouter)
		if err != nil {
			log.Fatal("Error preparing server: ", err)
		}

		serverEntryPoint := server.serverEntryPoints[newServerEntryPointName]
		serverEntryPoint.httpServer = newsrv
		go server.startServer(serverEntryPoint.httpServer, server.globalConfiguration)
	}
}

func (server *Server) listenProviders(stop chan bool) {
	lastReceivedConfiguration := safe.New(time.Unix(0, 0))
	lastConfigs := cmap.New()
	for {
		select {
		case <-stop:
			return
		case configMsg, ok := <-server.configurationChan:
			if !ok {
				return
			}
			server.defaultConfigurationValues(configMsg.Configuration)
			currentConfigurations := server.currentConfigurations.Get().(configs)
			jsonConf, _ := json.Marshal(configMsg.Configuration)
			log.Debugf("Configuration received from provider %s: %s", configMsg.ProviderName, string(jsonConf))
			if configMsg.Configuration == nil || configMsg.Configuration.Backends == nil && configMsg.Configuration.Frontends == nil {
				log.Infof("Skipping empty Configuration for provider %s", configMsg.ProviderName)
			} else if reflect.DeepEqual(currentConfigurations[configMsg.ProviderName], configMsg.Configuration) {
				log.Infof("Skipping same configuration for provider %s", configMsg.ProviderName)
			} else {
				lastConfigs.Set(configMsg.ProviderName, &configMsg)
				lastReceivedConfigurationValue := lastReceivedConfiguration.Get().(time.Time)
				providersThrottleDuration := time.Duration(server.globalConfiguration.ProvidersThrottleDuration)
				if time.Now().After(lastReceivedConfigurationValue.Add(providersThrottleDuration)) {
					log.Debugf("Last %s config received more than %s, OK", configMsg.ProviderName, server.globalConfiguration.ProvidersThrottleDuration.String())
					// last config received more than n s ago
					server.configurationValidatedChan <- configMsg
				} else {
					log.Debugf("Last %s config received less than %s, waiting...", configMsg.ProviderName, server.globalConfiguration.ProvidersThrottleDuration.String())
					safe.Go(func() {
						<-time.After(providersThrottleDuration)
						lastReceivedConfigurationValue := lastReceivedConfiguration.Get().(time.Time)
						if time.Now().After(lastReceivedConfigurationValue.Add(time.Duration(providersThrottleDuration))) {
							log.Debugf("Waited for %s config, OK", configMsg.ProviderName)
							if lastConfig, ok := lastConfigs.Get(configMsg.ProviderName); ok {
								server.configurationValidatedChan <- *lastConfig.(*types.ConfigMessage)
							}
						}
					})
				}
				lastReceivedConfiguration.Set(time.Now())
			}
		}
	}
}

func (server *Server) defaultConfigurationValues(configuration *types.Configuration) {
	if configuration == nil || configuration.Frontends == nil {
		return
	}
	for _, frontend := range configuration.Frontends {
		// default endpoints if not defined in frontends
		if len(frontend.EntryPoints) == 0 {
			frontend.EntryPoints = server.globalConfiguration.DefaultEntryPoints
		}
	}
	for backendName, backend := range configuration.Backends {
		_, err := types.NewLoadBalancerMethod(backend.LoadBalancer)
		if err != nil {
			log.Debugf("Load balancer method '%+v' for backend %s: %v. Using default wrr.", backend.LoadBalancer, backendName, err)
			backend.LoadBalancer = &types.LoadBalancer{Method: "wrr"}
		}
	}
}

func (server *Server) listenConfigurations(stop chan bool) {
	for {
		select {
		case <-stop:
			return
		case configMsg, ok := <-server.configurationValidatedChan:
			if !ok {
				return
			}
			currentConfigurations := server.currentConfigurations.Get().(configs)

			// Copy configurations to new map so we don't change current if LoadConfig fails
			newConfigurations := make(configs)
			for k, v := range currentConfigurations {
				newConfigurations[k] = v
			}
			newConfigurations[configMsg.ProviderName] = configMsg.Configuration

			newServerEntryPoints, err := server.loadConfig(newConfigurations, server.globalConfiguration)
			if err == nil {
				for newServerEntryPointName, newServerEntryPoint := range newServerEntryPoints {
					server.serverEntryPoints[newServerEntryPointName].httpRouter.UpdateHandler(newServerEntryPoint.httpRouter.GetHandler())
					log.Infof("Server configuration reloaded on %s", server.serverEntryPoints[newServerEntryPointName].httpServer.Addr)
				}
				server.currentConfigurations.Set(newConfigurations)
				server.postLoadConfig()
			} else {
				log.Error("Error loading new configuration, aborted ", err)
			}
		}
	}
}

func (server *Server) postLoadConfig() {
	if server.globalConfiguration.ACME == nil {
		return
	}
	if server.leadership != nil && !server.leadership.IsLeader() {
		return
	}
	if server.globalConfiguration.ACME.OnHostRule {
		currentConfigurations := server.currentConfigurations.Get().(configs)
		for _, configuration := range currentConfigurations {
			for _, frontend := range configuration.Frontends {

				// check if one of the frontend entrypoints is configured with TLS
				TLSEnabled := false
				for _, entrypoint := range frontend.EntryPoints {
					if server.globalConfiguration.EntryPoints[entrypoint].TLS != nil {
						TLSEnabled = true
						break
					}
				}

				if TLSEnabled {
					for _, route := range frontend.Routes {
						rules := Rules{}
						domains, err := rules.ParseDomains(route.Rule)
						if err != nil {
							log.Errorf("Error parsing domains: %v", err)
						} else {
							server.globalConfiguration.ACME.LoadCertificateForDomains(domains)
						}
					}
				}
			}
		}
	}
}

func (server *Server) configureProviders() {
	// configure providers
	if server.globalConfiguration.Docker != nil {
		server.providers = append(server.providers, server.globalConfiguration.Docker)
	}
	if server.globalConfiguration.Marathon != nil {
		server.providers = append(server.providers, server.globalConfiguration.Marathon)
	}
	if server.globalConfiguration.File != nil {
		server.providers = append(server.providers, server.globalConfiguration.File)
	}
	if server.globalConfiguration.Web != nil {
		server.globalConfiguration.Web.server = server
		server.providers = append(server.providers, server.globalConfiguration.Web)
	}
	if server.globalConfiguration.Consul != nil {
		server.providers = append(server.providers, server.globalConfiguration.Consul)
	}
	if server.globalConfiguration.ConsulCatalog != nil {
		server.providers = append(server.providers, server.globalConfiguration.ConsulCatalog)
	}
	if server.globalConfiguration.Etcd != nil {
		server.providers = append(server.providers, server.globalConfiguration.Etcd)
	}
	if server.globalConfiguration.Zookeeper != nil {
		server.providers = append(server.providers, server.globalConfiguration.Zookeeper)
	}
	if server.globalConfiguration.Boltdb != nil {
		server.providers = append(server.providers, server.globalConfiguration.Boltdb)
	}
	if server.globalConfiguration.Kubernetes != nil {
		server.providers = append(server.providers, server.globalConfiguration.Kubernetes)
	}
	if server.globalConfiguration.Mesos != nil {
		server.providers = append(server.providers, server.globalConfiguration.Mesos)
	}
	if server.globalConfiguration.Eureka != nil {
		server.providers = append(server.providers, server.globalConfiguration.Eureka)
	}
	if server.globalConfiguration.ECS != nil {
		server.providers = append(server.providers, server.globalConfiguration.ECS)
	}
	if server.globalConfiguration.Rancher != nil {
		server.providers = append(server.providers, server.globalConfiguration.Rancher)
	}
	if server.globalConfiguration.DynamoDB != nil {
		server.providers = append(server.providers, server.globalConfiguration.DynamoDB)
	}
}

func (server *Server) startProviders() {
	// start providers
	for _, provider := range server.providers {
		providerType := reflect.TypeOf(provider)
		jsonConf, _ := json.Marshal(provider)
		log.Infof("Starting provider %v %s", providerType, jsonConf)
		currentProvider := provider
		safe.Go(func() {
			err := currentProvider.Provide(server.configurationChan, server.routinesPool, server.globalConfiguration.Constraints)
			if err != nil {
				log.Errorf("Error starting provider %v: %s", providerType, err)
			}
		})
	}
}

func (server *Server) listenSignals() {
	sig := <-server.signals
	log.Infof("I have to go... %+v", sig)
	log.Info("Stopping server")
	server.Stop()
}

// creates a TLS config that allows terminating HTTPS for multiple domains using SNI
func (server *Server) createTLSConfig(entryPointName string, tlsOption *TLS, router *middlewares.HandlerSwitcher) (*tls.Config, error) {
	if tlsOption == nil {
		return nil, nil
	}

	config, err := tlsOption.Certificates.CreateTLSConfig()
	if err != nil {
		return nil, err
	}

	// ensure http2 enabled
	config.NextProtos = []string{"h2", "http/1.1"}

	if len(tlsOption.ClientCAFiles) > 0 {
		pool := x509.NewCertPool()
		for _, caFile := range tlsOption.ClientCAFiles {
			data, err := ioutil.ReadFile(caFile)
			if err != nil {
				return nil, err
			}
			ok := pool.AppendCertsFromPEM(data)
			if !ok {
				return nil, errors.New("invalid certificate(s) in " + caFile)
			}
		}
		config.ClientCAs = pool
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if server.globalConfiguration.ACME != nil {
		if _, ok := server.serverEntryPoints[server.globalConfiguration.ACME.EntryPoint]; ok {
			if entryPointName == server.globalConfiguration.ACME.EntryPoint {
				checkOnDemandDomain := func(domain string) bool {
					routeMatch := &mux.RouteMatch{}
					router := router.GetHandler()
					match := router.Match(&http.Request{URL: &url.URL{}, Host: domain}, routeMatch)
					if match && routeMatch.Route != nil {
						return true
					}
					return false
				}
				if server.leadership == nil {
					err := server.globalConfiguration.ACME.CreateLocalConfig(config, checkOnDemandDomain)
					if err != nil {
						return nil, err
					}
				} else {
					err := server.globalConfiguration.ACME.CreateClusterConfig(server.leadership, config, checkOnDemandDomain)
					if err != nil {
						return nil, err
					}
				}
			}
		} else {
			return nil, errors.New("Unknown entrypoint " + server.globalConfiguration.ACME.EntryPoint + " for ACME configuration")
		}
	}
	if len(config.Certificates) == 0 {
		return nil, errors.New("No certificates found for TLS entrypoint " + entryPointName)
	}
	// BuildNameToCertificate parses the CommonName and SubjectAlternateName fields
	// in each certificate and populates the config.NameToCertificate map.
	config.BuildNameToCertificate()
	//Set the minimum TLS version if set in the config TOML
	if minConst, exists := minVersion[server.globalConfiguration.EntryPoints[entryPointName].TLS.MinVersion]; exists {
		config.PreferServerCipherSuites = true
		config.MinVersion = minConst
	}
	//Set the list of CipherSuites if set in the config TOML
	if server.globalConfiguration.EntryPoints[entryPointName].TLS.CipherSuites != nil {
		//if our list of CipherSuites is defined in the entrypoint config, we can re-initilize the suites list as empty
		config.CipherSuites = make([]uint16, 0)
		for _, cipher := range server.globalConfiguration.EntryPoints[entryPointName].TLS.CipherSuites {
			if cipherConst, exists := cipherSuites[cipher]; exists {
				config.CipherSuites = append(config.CipherSuites, cipherConst)
			} else {
				//CipherSuite listed in the toml does not exist in our listed
				return nil, errors.New("Invalid CipherSuite: " + cipher)
			}
		}
	}
	return config, nil
}

func (server *Server) startServer(srv *http.Server, globalConfiguration GlobalConfiguration) {
	log.Infof("Starting server on %s", srv.Addr)
	var err error
	if srv.TLSConfig != nil {
		err = srv.ListenAndServeTLS("", "")
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil {
		log.Error("Error creating server: ", err)
	}
}

func (server *Server) prepareServer(entryPointName string, entryPoint *EntryPoint, middlewareChain http.Handler, router *middlewares.HandlerSwitcher) (*http.Server, error) {
	log.Infof("Preparing server %s %+v", entryPointName, entryPoint)
	tlsConfig, err := server.createTLSConfig(entryPointName, entryPoint.TLS, router)
	if err != nil {
		log.Errorf("Error creating TLS config: %s", err)
		return nil, err
	}

	return &http.Server{
		Addr:        entryPoint.Address,
		Handler:     middlewareChain,
		TLSConfig:   tlsConfig,
		IdleTimeout: time.Duration(server.globalConfiguration.IdleTimeout),
	}, nil
}

func (server *Server) buildEntryPoints(globalConfiguration GlobalConfiguration) map[string]*serverEntryPoint {
	serverEntryPoints := make(map[string]*serverEntryPoint)
	for entryPointName := range globalConfiguration.EntryPoints {
		router := server.buildDefaultHTTPRouter()
		serverEntryPoints[entryPointName] = &serverEntryPoint{
			httpRouter: middlewares.NewHandlerSwitcher(router),
		}
	}
	return serverEntryPoints
}

// LoadConfig returns a new gorilla.mux Route from the specified global configuration and the dynamic
// provider configurations.
func (server *Server) loadConfig(configurations configs, globalConfiguration GlobalConfiguration) (map[string]*serverEntryPoint, error) {
	serverEntryPoints := server.buildEntryPoints(globalConfiguration)
	redirectHandlers := make(map[string]http.Handler)
	backends := make(map[string]http.Handler)
	backendsHealthcheck := make(map[string]*healthcheck.BackendHealthCheck)

	backend2FrontendMap := map[string]string{}
	for _, configuration := range configurations {
		frontendNames := sortedFrontendNamesForConfig(configuration)
	frontend:
		for _, frontendName := range frontendNames {
			frontend := configuration.Frontends[frontendName]

			log.Debugf("Creating frontend %s", frontendName)

			fwd, err := forward.New(forward.Logger(oxyLogger), forward.PassHostHeader(frontend.PassHostHeader))
			if err != nil {
				log.Errorf("Error creating forwarder for frontend %s: %v", frontendName, err)
				log.Errorf("Skipping frontend %s...", frontendName)
				continue frontend
			}
			if len(frontend.EntryPoints) == 0 {
				log.Errorf("No entrypoint defined for frontend %s, defaultEntryPoints:%s", frontendName, globalConfiguration.DefaultEntryPoints)
				log.Errorf("Skipping frontend %s...", frontendName)
				continue frontend
			}

			for _, entryPointName := range frontend.EntryPoints {
				log.Debugf("Wiring frontend %s to entryPoint %s", frontendName, entryPointName)
				if _, ok := serverEntryPoints[entryPointName]; !ok {
					log.Errorf("Undefined entrypoint '%s' for frontend %s", entryPointName, frontendName)
					log.Errorf("Skipping frontend %s...", frontendName)
					continue frontend
				}
				server.createFrontendHandler(fwd, frontendName, frontend, entryPointName, serverEntryPoints[entryPointName], configuration,
					redirectHandlers, backends, backendsHealthcheck, backend2FrontendMap)
			}
		}
	}

	healthcheck.GetHealthCheck().SetBackendsConfiguration(server.routinesPool.Ctx(), backendsHealthcheck)
	middlewares.SetBackend2FrontendMap(&backend2FrontendMap)

	//sort routes
	for _, serverEntryPoint := range serverEntryPoints {
		serverEntryPoint.httpRouter.GetHandler().SortRoutes()
	}

	return serverEntryPoints, nil
}

func (server *Server) createFrontendHandler(fwd http.Handler,
	frontendName string, frontend *types.Frontend,
	entryPointName string, serverEntryPoint *serverEntryPoint,
	configuration *types.Configuration,
	// in/out parameters
	redirectHandlers map[string]http.Handler,
	backends map[string]http.Handler,
	backendsHealthcheck map[string]*healthcheck.BackendHealthCheck,
	backend2FrontendMap map[string]string) {

	log.Debugf("Wiring frontend %s to entryPoint %s", frontendName, entryPointName)
	newServerRoute := &serverRoute{route: serverEntryPoint.httpRouter.GetHandler().NewRoute().Name(frontendName)}
	for routeName, route := range frontend.Routes {
		err := getRoute(newServerRoute, &route)
		if err != nil {
			log.Errorf("Error creating route for frontend %s: %v", frontendName, err)
			log.Errorf("Skipping frontend %s...", frontendName)
			return
		}
		log.Debugf("Creating route %s %s", routeName, route.Rule)
	}

	entryPoint := server.globalConfiguration.EntryPoints[entryPointName]
	if entryPoint.Redirect != nil {
		if redirectHandlers[entryPointName] != nil {
			newServerRoute.route.Handler(redirectHandlers[entryPointName])
		} else if handler, err := server.loadEntryPointConfig(entryPointName, entryPoint, http.NotFoundHandler()); err != nil {
			log.Errorf("Error loading entrypoint configuration for frontend %s: %v", frontendName, err)
			log.Errorf("Skipping frontend %s...", frontendName)
			return
		} else {
			saveFrontend := accesslog.NewSaveFrontend(handler, frontendName)
			newServerRoute.route.Handler(saveFrontend)
			redirectHandlers[entryPointName] = handler
		}
	} else {
		if backends[frontend.Backend] == nil {
			backend, ok := server.createBackendHandler(fwd, frontendName, frontend.Backend, configuration, backendsHealthcheck, backend2FrontendMap)
			if ok {
				backends[frontend.Backend] = backend
			}
		} else {
			log.Debugf("Reusing backend %s", frontend.Backend)
		}
		if frontend.Priority > 0 {
			newServerRoute.route.Priority(frontend.Priority)
		}
		server.wireFrontendBackend(newServerRoute, backends[frontend.Backend])
	}
	err := newServerRoute.route.GetError()
	if err != nil {
		log.Errorf("Error building route: %s", err)
	}
}

func (server *Server) createBackendHandler(fwd http.Handler, frontendName, backendName string,
	configuration *types.Configuration,
	backendsHealthcheck map[string]*healthcheck.BackendHealthCheck,
	backend2FrontendMap map[string]string) (http.Handler, bool) {

	saveBackend := accesslog.NewSaveBackend(fwd, backendName)
	saveFrontend := accesslog.NewSaveFrontend(saveBackend, frontendName)
	rr, _ := roundrobin.New(saveFrontend)
	if configuration.Backends[backendName] == nil {
		log.Errorf("Undefined backend '%s' for frontend %s", backendName, frontendName)
		log.Errorf("Skipping frontend %s...", frontendName)
		return nil, false
	}

	lbMethod, err := types.NewLoadBalancerMethod(configuration.Backends[backendName].LoadBalancer)
	if err != nil {
		log.Errorf("Error loading load balancer method '%+v' for frontend %s: %v", configuration.Backends[backendName].LoadBalancer, frontendName, err)
		log.Errorf("Skipping frontend %s...", frontendName)
		return nil, false
	}

	stickysession := configuration.Backends[backendName].LoadBalancer.Sticky
	cookiename := "_TRAEFIK_BACKEND"
	var sticky *roundrobin.StickySession

	if stickysession {
		sticky = roundrobin.NewStickySession(cookiename)
	}

	var lb http.Handler
	switch lbMethod {
	case types.Drr:
		log.Debugf("Creating load-balancer drr")
		rebalancer, _ := roundrobin.NewRebalancer(rr, roundrobin.RebalancerLogger(oxyLogger))
		if stickysession {
			log.Debugf("Sticky session with cookie %v", cookiename)
			rebalancer, _ = roundrobin.NewRebalancer(rr, roundrobin.RebalancerLogger(oxyLogger), roundrobin.RebalancerStickySession(sticky))
		}
		lb = rebalancer
		for serverName, server := range configuration.Backends[backendName].Servers {
			url, err := url.Parse(server.URL)
			if err != nil {
				log.Errorf("Error parsing server URL %s: %v", server.URL, err)
				log.Errorf("Skipping frontend %s...", frontendName)
				return nil, false
			}
			backend2FrontendMap[url.String()] = frontendName
			log.Debugf("Creating server %s at %s with weight %d", serverName, url, server.Weight)
			if err := rebalancer.UpsertServer(url, roundrobin.Weight(server.Weight)); err != nil {
				log.Errorf("Error adding server %s to load balancer: %v", server.URL, err)
				log.Errorf("Skipping frontend %s...", frontendName)
				return nil, false
			}
			if configuration.Backends[backendName].HealthCheck != nil {
				var interval time.Duration
				if configuration.Backends[backendName].HealthCheck.Interval != "" {
					interval, err = time.ParseDuration(configuration.Backends[backendName].HealthCheck.Interval)
					if err != nil {
						log.Errorf("Wrong healthcheck interval: %s", err)
						interval = time.Second * 30
					}
				}
				backendsHealthcheck[backendName] = healthcheck.NewBackendHealthCheck(configuration.Backends[backendName].HealthCheck.Path, interval, rebalancer)
			}
		}
	case types.Wrr:
		log.Debugf("Creating load-balancer wrr")
		if stickysession {
			log.Debugf("Sticky session with cookie %v", cookiename)
			rr, _ = roundrobin.New(saveBackend, roundrobin.EnableStickySession(sticky))
		}
		lb = rr
		for serverName, server := range configuration.Backends[backendName].Servers {
			url, err := url.Parse(server.URL)
			if err != nil {
				log.Errorf("Error parsing server URL %s: %v", server.URL, err)
				log.Errorf("Skipping frontend %s...", frontendName)
				return nil, false
			}
			backend2FrontendMap[url.String()] = frontendName
			log.Debugf("Creating server %s at %s with weight %d", serverName, url, server.Weight)
			if err := rr.UpsertServer(url, roundrobin.Weight(server.Weight)); err != nil {
				log.Errorf("Error adding server %s to load balancer: %v", server.URL, err)
				log.Errorf("Skipping frontend %s...", frontendName)
				return nil, false
			}
		}
		if configuration.Backends[backendName].HealthCheck != nil {
			var interval time.Duration
			if configuration.Backends[backendName].HealthCheck.Interval != "" {
				interval, err = time.ParseDuration(configuration.Backends[backendName].HealthCheck.Interval)
				if err != nil {
					log.Errorf("Wrong healthcheck interval: %s", err)
					interval = time.Second * 30
				}
			}
			backendsHealthcheck[backendName] = healthcheck.NewBackendHealthCheck(configuration.Backends[backendName].HealthCheck.Path, interval, rr)
		}
	}

	maxConns := configuration.Backends[backendName].MaxConn
	if maxConns != nil && maxConns.Amount != 0 {
		extractFunc, err := utils.NewExtractor(maxConns.ExtractorFunc)
		if err != nil {
			log.Errorf("Error creating connlimit: %v", err)
			log.Errorf("Skipping frontend %s...", frontendName)
			return nil, false
		}
		log.Debugf("Creating loadd-balancer connlimit")
		lb, err = connlimit.New(lb, extractFunc, maxConns.Amount, connlimit.Logger(oxyLogger))
		if err != nil {
			log.Errorf("Error creating connlimit: %v", err)
			log.Errorf("Skipping frontend %s...", frontendName)
			return nil, false
		}
	}

	// retry ?
	if server.globalConfiguration.Retry != nil {
		retries := len(configuration.Backends[backendName].Servers)
		if server.globalConfiguration.Retry.Attempts > 0 {
			retries = server.globalConfiguration.Retry.Attempts
		}
		lb = middlewares.NewRetry(retries, lb)
		log.Debugf("Creating retries max attempts %d", retries)
	}

	var backendHandler = lb
	if configuration.Backends[backendName].CircuitBreaker != nil {
		expr := configuration.Backends[backendName].CircuitBreaker.Expression
		log.Debugf("Creating circuit breaker %s", expr)
		backendHandler, err = middlewares.NewCircuitBreaker(lb, expr, cbreaker.Logger(oxyLogger))
		if err != nil {
			log.Errorf("Error creating circuit breaker: %v", err)
			log.Errorf("Skipping frontend %s...", frontendName)
			return nil, false
		}
	}

	web := server.globalConfiguration.Web
	if web != nil && web.Metrics != nil && web.Metrics.Prometheus != nil {
		p := middlewares.NewPrometheus(backendName, web.Metrics.Prometheus)
		backendHandler = middlewares.NewMetricsWrapper(p, backendHandler)
	}

	return backendHandler, true
}

func (server *Server) wireFrontendBackend(serverRoute *serverRoute, handler http.Handler) {
	// add prefix
	if len(serverRoute.addPrefix) > 0 {
		handler = &middlewares.AddPrefix{BasicMiddleware: common.NewMiddleware(handler), Prefix: serverRoute.addPrefix}
	}

	// strip prefix
	if len(serverRoute.stripPrefixes) > 0 {
		handler = &middlewares.StripPrefix{BasicMiddleware: common.NewMiddleware(handler), Prefixes: serverRoute.stripPrefixes}
	}

	serverRoute.route.Handler(handler)
}

func (server *Server) loadEntryPointConfig(entryPointName string, entryPoint *EntryPoint, next http.Handler) (http.Handler, error) {
	regex := entryPoint.Redirect.Regex
	replacement := entryPoint.Redirect.Replacement
	if len(entryPoint.Redirect.EntryPoint) > 0 {
		regex = "^(?:https?:\\/\\/)?([\\w\\._-]+)(?::\\d+)?(.*)$"
		if server.globalConfiguration.EntryPoints[entryPoint.Redirect.EntryPoint] == nil {
			return nil, errors.New("Unknown entrypoint " + entryPoint.Redirect.EntryPoint)
		}
		protocol := "http"
		if server.globalConfiguration.EntryPoints[entryPoint.Redirect.EntryPoint].TLS != nil {
			protocol = "https"
		}
		r, _ := regexp.Compile("(:\\d+)")
		match := r.FindStringSubmatch(server.globalConfiguration.EntryPoints[entryPoint.Redirect.EntryPoint].Address)
		if len(match) == 0 {
			return nil, errors.New("Bad Address format: " + server.globalConfiguration.EntryPoints[entryPoint.Redirect.EntryPoint].Address)
		}
		replacement = protocol + "://$1" + match[0] + "$2"
	}
	rewrite, err := middlewares.NewRewrite(regex, replacement, true, next)
	if err != nil {
		return nil, err
	}
	log.Debugf("Creating entryPoint redirect %s -> %s : %s -> %s", entryPointName, entryPoint.Redirect.EntryPoint, regex, replacement)
	return rewrite, nil
}

func (server *Server) buildDefaultHTTPRouter() *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	router.StrictSlash(true)
	router.SkipClean(true)
	return router
}

func getRoute(serverRoute *serverRoute, route *types.Route) error {
	rules := Rules{route: serverRoute}
	newRoute, err := rules.Parse(route.Rule)
	if err != nil {
		return err
	}
	newRoute.Priority(serverRoute.route.GetPriority() + len(route.Rule))
	serverRoute.route = newRoute
	return nil
}

func sortedFrontendNamesForConfig(configuration *types.Configuration) []string {
	keys := []string{}
	for key := range configuration.Frontends {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
