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

	"github.com/codegangsta/negroni"
	"github.com/containous/mux"
	"github.com/containous/traefik/cluster"
	"github.com/containous/traefik/healthcheck"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/middlewares/accesslog"
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
	loggerMiddleware           *middlewares.Logger
	accessLoggerMiddleware     *accesslog.LogHandler
	routinesPool               *safe.Pool
	leadership                 *cluster.Leadership
}

type serverEntryPoints map[string]*serverEntryPoint

type serverEntryPoint struct {
	httpServer *http.Server
	httpRouter *middlewares.HandlerSwitcher
}

type serverRoute struct {
	route              *mux.Route
	stripPrefixes      []string
	stripPrefixesRegex []string
	addPrefix          string
	replacePath        string
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
	server.loggerMiddleware = middlewares.NewLogger(globalConfiguration.AccessLogsFile)
	server.accessLoggerMiddleware = accesslog.NewLogHandler()
	server.routinesPool = safe.NewPool(context.Background())
	if globalConfiguration.Cluster != nil {
		// leadership creation if cluster mode
		server.leadership = cluster.NewLeadership(server.routinesPool.Ctx(), globalConfiguration.Cluster)
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
	server.loggerMiddleware.Close()
	server.accessLoggerMiddleware.Close()
	cancel()
}

func (server *Server) startLeadership() {
	if server.leadership != nil {
		server.leadership.Participate(server.routinesPool)
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
		serverMiddlewares := []negroni.Handler{server.accessLoggerMiddleware, server.loggerMiddleware, metrics}
		if server.globalConfiguration.Web != nil && server.globalConfiguration.Web.Metrics != nil {
			if server.globalConfiguration.Web.Metrics.Prometheus != nil {
				metricsMiddleware := middlewares.NewMetricsWrapper(middlewares.NewPrometheus(newServerEntryPointName, server.globalConfiguration.Web.Metrics.Prometheus))
				serverMiddlewares = append(serverMiddlewares, metricsMiddleware)
			}
		}
		if server.globalConfiguration.Web != nil && server.globalConfiguration.Web.Statistics != nil {
			statsRecorder = middlewares.NewStatsRecorder(server.globalConfiguration.Web.Statistics.RecentErrors)
			serverMiddlewares = append(serverMiddlewares, statsRecorder)
		}
		if server.globalConfiguration.EntryPoints[newServerEntryPointName].Auth != nil {
			authMiddleware, err := middlewares.NewAuthenticator(server.globalConfiguration.EntryPoints[newServerEntryPointName].Auth)
			if err != nil {
				log.Fatal("Error starting server: ", err)
			}
			serverMiddlewares = append(serverMiddlewares, authMiddleware)
		}
		if server.globalConfiguration.EntryPoints[newServerEntryPointName].Compress {
			serverMiddlewares = append(serverMiddlewares, &middlewares.Compress{})
		}
		newsrv, err := server.prepareServer(newServerEntryPointName, newServerEntryPoint.httpRouter, server.globalConfiguration.EntryPoints[newServerEntryPointName], serverMiddlewares...)
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
				// and is configured with ACME
				ACMEEnabled := false
				for _, entrypoint := range frontend.EntryPoints {
					if server.globalConfiguration.ACME.EntryPoint == entrypoint && server.globalConfiguration.EntryPoints[entrypoint].TLS != nil {
						ACMEEnabled = true
						break
					}
				}

				if ACMEEnabled {
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

func (server *Server) prepareServer(entryPointName string, router *middlewares.HandlerSwitcher, entryPoint *EntryPoint, middlewares ...negroni.Handler) (*http.Server, error) {
	log.Infof("Preparing server %s %+v", entryPointName, entryPoint)
	// middlewares
	var negroni = negroni.New()
	for _, middleware := range middlewares {
		negroni.Use(middleware)
	}
	negroni.UseHandler(router)
	tlsConfig, err := server.createTLSConfig(entryPointName, entryPoint.TLS, router)
	if err != nil {
		log.Errorf("Error creating TLS config: %s", err)
		return nil, err
	}

	return &http.Server{
		Addr:        entryPoint.Address,
		Handler:     negroni,
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
	redirectHandlers := make(map[string]negroni.Handler)
	backends := map[string]http.Handler{}
	backendsHealthcheck := map[string]*healthcheck.BackendHealthCheck{}
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

				newServerRoute := &serverRoute{route: serverEntryPoints[entryPointName].httpRouter.GetHandler().NewRoute().Name(frontendName)}
				for routeName, route := range frontend.Routes {
					err := getRoute(newServerRoute, &route)
					if err != nil {
						log.Errorf("Error creating route for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}
					log.Debugf("Creating route %s %s", routeName, route.Rule)
				}

				entryPoint := globalConfiguration.EntryPoints[entryPointName]
				negroni := negroni.New()
				if entryPoint.Redirect != nil {
					if redirectHandlers[entryPointName] != nil {
						negroni.Use(redirectHandlers[entryPointName])
					} else if handler, err := server.loadEntryPointConfig(entryPointName, entryPoint); err != nil {
						log.Errorf("Error loading entrypoint configuration for frontend %s: %v", frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					} else {
						saveFrontend := accesslog.NewSaveNegroniFrontend(handler, frontendName)
						negroni.Use(saveFrontend)
						redirectHandlers[entryPointName] = saveFrontend
					}
				}
				if backends[entryPointName+frontend.Backend] == nil {
					log.Debugf("Creating backend %s", frontend.Backend)
					saveBackend := accesslog.NewSaveBackend(fwd, frontend.Backend)
					saveFrontend := accesslog.NewSaveFrontend(saveBackend, frontendName)
					rr, _ := roundrobin.New(saveFrontend)
					if configuration.Backends[frontend.Backend] == nil {
						log.Errorf("Undefined backend '%s' for frontend %s", frontend.Backend, frontendName)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					lbMethod, err := types.NewLoadBalancerMethod(configuration.Backends[frontend.Backend].LoadBalancer)
					if err != nil {
						log.Errorf("Error loading load balancer method '%+v' for frontend %s: %v", configuration.Backends[frontend.Backend].LoadBalancer, frontendName, err)
						log.Errorf("Skipping frontend %s...", frontendName)
						continue frontend
					}

					stickysession := configuration.Backends[frontend.Backend].LoadBalancer.Sticky
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
						for serverName, server := range configuration.Backends[frontend.Backend].Servers {
							url, err := url.Parse(server.URL)
							if err != nil {
								log.Errorf("Error parsing server URL %s: %v", server.URL, err)
								log.Errorf("Skipping frontend %s...", frontendName)
								continue frontend
							}
							backend2FrontendMap[url.String()] = frontendName
							log.Debugf("Creating server %s at %s with weight %d", serverName, url.String(), server.Weight)
							if err := rebalancer.UpsertServer(url, roundrobin.Weight(server.Weight)); err != nil {
								log.Errorf("Error adding server %s to load balancer: %v", server.URL, err)
								log.Errorf("Skipping frontend %s...", frontendName)
								continue frontend
							}
							hcOpts := parseHealthCheckOptions(rebalancer, frontend.Backend, configuration.Backends[frontend.Backend].HealthCheck, *globalConfiguration.HealthCheck)
							if hcOpts != nil {
								log.Debugf("Setting up backend health check %s", *hcOpts)
								backendsHealthcheck[frontend.Backend] = healthcheck.NewBackendHealthCheck(*hcOpts)
							}
						}
					case types.Wrr:
						log.Debugf("Creating load-balancer wrr")
						if stickysession {
							log.Debugf("Sticky session with cookie %v", cookiename)
							rr, _ = roundrobin.New(saveFrontend, roundrobin.EnableStickySession(sticky))
						}
						lb = rr
						for serverName, server := range configuration.Backends[frontend.Backend].Servers {
							url, err := url.Parse(server.URL)
							if err != nil {
								log.Errorf("Error parsing server URL %s: %v", server.URL, err)
								log.Errorf("Skipping frontend %s...", frontendName)
								continue frontend
							}
							backend2FrontendMap[url.String()] = frontendName
							log.Debugf("Creating server %s at %s with weight %d", serverName, url.String(), server.Weight)
							if err := rr.UpsertServer(url, roundrobin.Weight(server.Weight)); err != nil {
								log.Errorf("Error adding server %s to load balancer: %v", server.URL, err)
								log.Errorf("Skipping frontend %s...", frontendName)
								continue frontend
							}
						}
						hcOpts := parseHealthCheckOptions(rr, frontend.Backend, configuration.Backends[frontend.Backend].HealthCheck, *globalConfiguration.HealthCheck)
						if hcOpts != nil {
							log.Debugf("Setting up backend health check %s", *hcOpts)
							backendsHealthcheck[frontend.Backend] = healthcheck.NewBackendHealthCheck(*hcOpts)
						}
					}
					maxConns := configuration.Backends[frontend.Backend].MaxConn
					if maxConns != nil && maxConns.Amount != 0 {
						extractFunc, err := utils.NewExtractor(maxConns.ExtractorFunc)
						if err != nil {
							log.Errorf("Error creating connlimit: %v", err)
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
						log.Debugf("Creating load-balancer connlimit")
						lb, err = connlimit.New(lb, extractFunc, maxConns.Amount, connlimit.Logger(oxyLogger))
						if err != nil {
							log.Errorf("Error creating connlimit: %v", err)
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
					}
					// retry ?
					if globalConfiguration.Retry != nil {
						retries := len(configuration.Backends[frontend.Backend].Servers)
						if globalConfiguration.Retry.Attempts > 0 {
							retries = globalConfiguration.Retry.Attempts
						}
						lb = middlewares.NewRetry(retries, lb)
						log.Debugf("Creating retries max attempts %d", retries)
					}

					if server.globalConfiguration.Web != nil && server.globalConfiguration.Web.Metrics != nil {
						if server.globalConfiguration.Web.Metrics.Prometheus != nil {
							metricsMiddlewareBackend := middlewares.NewMetricsWrapper(middlewares.NewPrometheus(frontend.Backend, server.globalConfiguration.Web.Metrics.Prometheus))
							negroni.Use(metricsMiddlewareBackend)
						}
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
						authMiddleware, err := middlewares.NewAuthenticator(auth)
						if err != nil {
							log.Fatal("Error creating Auth: ", err)
						}
						negroni.Use(authMiddleware)
					}
					if configuration.Backends[frontend.Backend].CircuitBreaker != nil {
						log.Debugf("Creating circuit breaker %s", configuration.Backends[frontend.Backend].CircuitBreaker.Expression)
						cbreaker, err := middlewares.NewCircuitBreaker(lb, configuration.Backends[frontend.Backend].CircuitBreaker.Expression, cbreaker.Logger(oxyLogger))
						if err != nil {
							log.Errorf("Error creating circuit breaker: %v", err)
							log.Errorf("Skipping frontend %s...", frontendName)
							continue frontend
						}
						negroni.Use(cbreaker)
					} else {
						negroni.UseHandler(lb)
					}
					backends[entryPointName+frontend.Backend] = negroni
				} else {
					log.Debugf("Reusing backend %s", frontend.Backend)
				}
				if frontend.Priority > 0 {
					newServerRoute.route.Priority(frontend.Priority)
				}
				server.wireFrontendBackend(newServerRoute, backends[entryPointName+frontend.Backend])

				err := newServerRoute.route.GetError()
				if err != nil {
					log.Errorf("Error building route: %s", err)
				}
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

func (server *Server) wireFrontendBackend(serverRoute *serverRoute, handler http.Handler) {
	// add prefix
	if len(serverRoute.addPrefix) > 0 {
		handler = &middlewares.AddPrefix{
			Prefix:  serverRoute.addPrefix,
			Handler: handler,
		}
	}

	// strip prefix
	if len(serverRoute.stripPrefixes) > 0 {
		handler = &middlewares.StripPrefix{
			Prefixes: serverRoute.stripPrefixes,
			Handler:  handler,
		}
	}

	// strip prefix with regex
	if len(serverRoute.stripPrefixesRegex) > 0 {
		handler = middlewares.NewStripPrefixRegex(handler, serverRoute.stripPrefixesRegex)
	}

	// path replace
	if len(serverRoute.replacePath) > 0 {
		handler = &middlewares.ReplacePath{
			Path:    serverRoute.replacePath,
			Handler: handler,
		}
	}

	serverRoute.route.Handler(handler)
}

func (server *Server) loadEntryPointConfig(entryPointName string, entryPoint *EntryPoint) (negroni.Handler, error) {
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
	rewrite, err := middlewares.NewRewrite(regex, replacement, true)
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

func parseHealthCheckOptions(lb healthcheck.LoadBalancer, backend string, hc *types.HealthCheck, hcConfig HealthCheckConfig) *healthcheck.Options {
	if hc == nil || hc.Path == "" {
		return nil
	}

	interval := time.Duration(hcConfig.Interval)
	if hc.Interval != "" {
		intervalOverride, err := time.ParseDuration(hc.Interval)
		switch {
		case err != nil:
			log.Errorf("Illegal healthcheck interval for backend '%s': %s", backend, err)
		case intervalOverride <= 0:
			log.Errorf("Healthcheck interval smaller than zero for backend '%s', backend", backend)
		default:
			interval = intervalOverride
		}
	}

	return &healthcheck.Options{
		Path:     hc.Path,
		Interval: interval,
		LB:       lb,
	}
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
