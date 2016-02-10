/*
Copyright
*/
package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/emilevauge/traefik/middlewares"
	"github.com/emilevauge/traefik/provider"
	"github.com/emilevauge/traefik/types"
	"github.com/gorilla/mux"
	"github.com/mailgun/manners"
	"github.com/mailgun/oxy/cbreaker"
	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/roundrobin"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"regexp"
	"sync"
	"syscall"
	"time"
)

var oxyLogger = &OxyLogger{}

// Server is the reverse-proxy/load-balancer engine
type Server struct {
	serverEntryPoints          map[string]serverEntryPoint
	configurationChan          chan types.ConfigMessage
	configurationValidatedChan chan types.ConfigMessage
	signals                    chan os.Signal
	stopChan                   chan bool
	providers                  []provider.Provider
	serverLock                 sync.Mutex
	currentConfigurations      configs
	globalConfiguration        GlobalConfiguration
	loggerMiddleware           *middlewares.Logger
}

type serverEntryPoint struct {
	httpServer *manners.GracefulServer
	httpRouter *mux.Router
}

// NewServer returns an initialized Server.
func NewServer(globalConfiguration GlobalConfiguration) *Server {
	server := new(Server)

	server.serverEntryPoints = make(map[string]serverEntryPoint)
	server.configurationChan = make(chan types.ConfigMessage, 10)
	server.configurationValidatedChan = make(chan types.ConfigMessage, 10)
	server.signals = make(chan os.Signal, 1)
	server.stopChan = make(chan bool)
	server.providers = []provider.Provider{}
	signal.Notify(server.signals, syscall.SIGINT, syscall.SIGTERM)
	server.currentConfigurations = make(configs)
	server.globalConfiguration = globalConfiguration
	server.loggerMiddleware = middlewares.NewLogger(globalConfiguration.AccessLogsFile)

	return server
}

// Start starts the server and blocks until server is shutted down.
func (server *Server) Start() {
	go server.listenProviders()
	go server.listenConfigurations()
	server.configureProviders()
	server.startProviders()
	go server.listenSignals()
	<-server.stopChan
}

// Stop stops the server
func (server *Server) Stop() {
	for _, serverEntryPoint := range server.serverEntryPoints {
		serverEntryPoint.httpServer.Close()
	}
	server.stopChan <- true
}

// Close destroys the server
func (server *Server) Close() {
	close(server.configurationChan)
	close(server.configurationValidatedChan)
	close(server.signals)
	close(server.stopChan)
	server.loggerMiddleware.Close()
}

func (server *Server) listenProviders() {
	lastReceivedConfiguration := time.Unix(0, 0)
	lastConfigs := make(map[string]*types.ConfigMessage)
	for {
		configMsg := <-server.configurationChan
		jsonConf, _ := json.Marshal(configMsg.Configuration)
		log.Debugf("Configuration receveived from provider %s: %s", configMsg.ProviderName, string(jsonConf))
		lastConfigs[configMsg.ProviderName] = &configMsg
		if time.Now().After(lastReceivedConfiguration.Add(time.Duration(server.globalConfiguration.ProvidersThrottleDuration))) {
			log.Debugf("Last %s config received more than %s, OK", configMsg.ProviderName, server.globalConfiguration.ProvidersThrottleDuration)
			// last config received more than n s ago
			server.configurationValidatedChan <- configMsg
		} else {
			log.Debugf("Last %s config received less than %s, waiting...", configMsg.ProviderName, server.globalConfiguration.ProvidersThrottleDuration)
			go func() {
				<-time.After(server.globalConfiguration.ProvidersThrottleDuration)
				if time.Now().After(lastReceivedConfiguration.Add(time.Duration(server.globalConfiguration.ProvidersThrottleDuration))) {
					log.Debugf("Waited for %s config, OK", configMsg.ProviderName)
					server.configurationValidatedChan <- *lastConfigs[configMsg.ProviderName]
				}
			}()
		}
		lastReceivedConfiguration = time.Now()
	}
}

func (server *Server) listenConfigurations() {
	for {
		configMsg := <-server.configurationValidatedChan
		if configMsg.Configuration == nil {
			log.Info("Skipping empty Configuration")
		} else if reflect.DeepEqual(server.currentConfigurations[configMsg.ProviderName], configMsg.Configuration) {
			log.Info("Skipping same configuration")
		} else {
			// Copy configurations to new map so we don't change current if LoadConfig fails
			newConfigurations := make(configs)
			for k, v := range server.currentConfigurations {
				newConfigurations[k] = v
			}
			newConfigurations[configMsg.ProviderName] = configMsg.Configuration

			newServerEntryPoints, err := server.loadConfig(newConfigurations, server.globalConfiguration)
			if err == nil {
				server.serverLock.Lock()
				for newServerEntryPointName, newServerEntryPoint := range newServerEntryPoints {
					currentServerEntryPoint := server.serverEntryPoints[newServerEntryPointName]
					server.currentConfigurations = newConfigurations
					currentServerEntryPoint.httpRouter = newServerEntryPoint.httpRouter
					oldServer := currentServerEntryPoint.httpServer
					newsrv, err := server.prepareServer(currentServerEntryPoint.httpRouter, server.globalConfiguration.EntryPoints[newServerEntryPointName], oldServer, server.loggerMiddleware, metrics)
					if err != nil {
						log.Fatal("Error preparing server: ", err)
					}
					go server.startServer(newsrv, server.globalConfiguration)
					currentServerEntryPoint.httpServer = newsrv
					server.serverEntryPoints[newServerEntryPointName] = currentServerEntryPoint
					time.Sleep(1 * time.Second)
					if oldServer != nil {
						log.Info("Stopping old server")
						oldServer.Close()
					}
				}
				server.serverLock.Unlock()
			} else {
				log.Error("Error loading new configuration, aborted ", err)
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
	if server.globalConfiguration.Etcd != nil {
		server.providers = append(server.providers, server.globalConfiguration.Etcd)
	}
	if server.globalConfiguration.Zookeeper != nil {
		server.providers = append(server.providers, server.globalConfiguration.Zookeeper)
	}
	if server.globalConfiguration.Boltdb != nil {
		server.providers = append(server.providers, server.globalConfiguration.Boltdb)
	}
}

func (server *Server) startProviders() {
	// start providers
	for _, provider := range server.providers {
		jsonConf, _ := json.Marshal(provider)
		log.Infof("Starting provider %v %s", reflect.TypeOf(provider), jsonConf)
		currentProvider := provider
		go func() {
			err := currentProvider.Provide(server.configurationChan)
			if err != nil {
				log.Errorf("Error starting provider %s", err)
			}
		}()
	}
}

func (server *Server) listenSignals() {
	sig := <-server.signals
	log.Infof("I have to go... %+v", sig)
	log.Info("Stopping server")
	server.Stop()
}

// creates a TLS config that allows terminating HTTPS for multiple domains using SNI
func (server *Server) createTLSConfig(tlsOption *TLS) (*tls.Config, error) {
	if tlsOption == nil {
		return nil, nil
	}
	if len(tlsOption.Certificates) == 0 {
		return nil, nil
	}

	config := &tls.Config{}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	var err error
	config.Certificates = make([]tls.Certificate, len(tlsOption.Certificates))
	for i, v := range tlsOption.Certificates {
		config.Certificates[i], err = tls.LoadX509KeyPair(v.CertFile, v.KeyFile)
		if err != nil {
			return nil, err
		}
	}
	// BuildNameToCertificate parses the CommonName and SubjectAlternateName fields
	// in each certificate and populates the config.NameToCertificate map.
	config.BuildNameToCertificate()
	return config, nil
}

func (server *Server) startServer(srv *manners.GracefulServer, globalConfiguration GlobalConfiguration) {
	log.Info("Starting server on ", srv.Addr)
	if srv.TLSConfig != nil {
		err := srv.ListenAndServeTLSWithConfig(srv.TLSConfig)
		if err != nil {
			log.Fatal("Error creating server: ", err)
		}
	} else {
		err := srv.ListenAndServe()
		if err != nil {
			log.Fatal("Error creating server: ", err)
		}
	}
	log.Info("Server stopped")
}

func (server *Server) prepareServer(router *mux.Router, entryPoint *EntryPoint, oldServer *manners.GracefulServer, middlewares ...negroni.Handler) (*manners.GracefulServer, error) {
	log.Info("Preparing server")
	// middlewares
	var negroni = negroni.New()
	for _, middleware := range middlewares {
		negroni.Use(middleware)
	}
	negroni.UseHandler(router)
	tlsConfig, err := server.createTLSConfig(entryPoint.TLS)
	if err != nil {
		log.Fatalf("Error creating TLS config %s", err)
		return nil, err
	}

	if oldServer == nil {
		return manners.NewWithServer(
			&http.Server{
				Addr:      entryPoint.Address,
				Handler:   negroni,
				TLSConfig: tlsConfig,
			}), nil
	}
	gracefulServer, err := oldServer.HijackListener(&http.Server{
		Addr:      entryPoint.Address,
		Handler:   negroni,
		TLSConfig: tlsConfig,
	}, tlsConfig)
	if err != nil {
		log.Fatalf("Error hijacking server %s", err)
		return nil, err
	}
	return gracefulServer, nil
}

func (server *Server) buildEntryPoints(globalConfiguration GlobalConfiguration) map[string]serverEntryPoint {
	serverEntryPoints := make(map[string]serverEntryPoint)
	for entryPointName := range globalConfiguration.EntryPoints {
		router := server.buildDefaultHTTPRouter()
		serverEntryPoints[entryPointName] = serverEntryPoint{
			httpRouter: router,
		}
	}
	return serverEntryPoints
}

// LoadConfig returns a new gorilla.mux Route from the specified global configuration and the dynamic
// provider configurations.
func (server *Server) loadConfig(configurations configs, globalConfiguration GlobalConfiguration) (map[string]serverEntryPoint, error) {
	serverEntryPoints := server.buildEntryPoints(globalConfiguration)
	redirectHandlers := make(map[string]http.Handler)

	backends := map[string]http.Handler{}
	for _, configuration := range configurations {
		for frontendName, frontend := range configuration.Frontends {
			log.Debugf("Creating frontend %s", frontendName)
			fwd, _ := forward.New(forward.Logger(oxyLogger), forward.PassHostHeader(frontend.PassHostHeader))
			// default endpoints if not defined in frontends
			if len(frontend.EntryPoints) == 0 {
				frontend.EntryPoints = globalConfiguration.DefaultEntryPoints
			}
			for _, entryPointName := range frontend.EntryPoints {
				log.Debugf("Wiring frontend %s to entryPoint %s", frontendName, entryPointName)
				if _, ok := serverEntryPoints[entryPointName]; !ok {
					return nil, errors.New("Undefined entrypoint: " + entryPointName)
				}
				newRoute := serverEntryPoints[entryPointName].httpRouter.NewRoute().Name(frontendName)
				for routeName, route := range frontend.Routes {
					log.Debugf("Creating route %s %s:%s", routeName, route.Rule, route.Value)
					newRouteReflect, err := invoke(newRoute, route.Rule, route.Value)
					if err != nil {
						return nil, err
					}
					newRoute = newRouteReflect[0].Interface().(*mux.Route)
				}
				entryPoint := globalConfiguration.EntryPoints[entryPointName]
				if entryPoint.Redirect != nil {
					if redirectHandlers[entryPointName] != nil {
						newRoute.Handler(redirectHandlers[entryPointName])
					} else if handler, err := server.loadEntryPointConfig(entryPointName, entryPoint); err != nil {
						return nil, err
					} else {
						newRoute.Handler(handler)
						redirectHandlers[entryPointName] = handler
					}
				} else {
					if backends[frontend.Backend] == nil {
						log.Debugf("Creating backend %s", frontend.Backend)
						var lb http.Handler
						rr, _ := roundrobin.New(fwd)
						if configuration.Backends[frontend.Backend] == nil {
							return nil, errors.New("Undefined backend: " + frontend.Backend)
						}
						lbMethod, err := types.NewLoadBalancerMethod(configuration.Backends[frontend.Backend].LoadBalancer)
						if err != nil {
							configuration.Backends[frontend.Backend].LoadBalancer = &types.LoadBalancer{Method: "wrr"}
						}
						switch lbMethod {
						case types.Drr:
							log.Debugf("Creating load-balancer drr")
							rebalancer, _ := roundrobin.NewRebalancer(rr, roundrobin.RebalancerLogger(oxyLogger))
							lb = rebalancer
							for serverName, server := range configuration.Backends[frontend.Backend].Servers {
								url, err := url.Parse(server.URL)
								if err != nil {
									return nil, err
								}
								log.Debugf("Creating server %s at %s with weight %d", serverName, url.String(), server.Weight)
								rebalancer.UpsertServer(url, roundrobin.Weight(server.Weight))
							}
						case types.Wrr:
							log.Debugf("Creating load-balancer wrr")
							lb = middlewares.NewWebsocketUpgrader(rr)
							for serverName, server := range configuration.Backends[frontend.Backend].Servers {
								url, err := url.Parse(server.URL)
								if err != nil {
									return nil, err
								}
								log.Debugf("Creating server %s at %s with weight %d", serverName, url.String(), server.Weight)
								rr.UpsertServer(url, roundrobin.Weight(server.Weight))
							}
						}
						var negroni = negroni.New()
						if configuration.Backends[frontend.Backend].CircuitBreaker != nil {
							log.Debugf("Creating circuit breaker %s", configuration.Backends[frontend.Backend].CircuitBreaker.Expression)
							negroni.Use(middlewares.NewCircuitBreaker(lb, configuration.Backends[frontend.Backend].CircuitBreaker.Expression, cbreaker.Logger(oxyLogger)))
						} else {
							negroni.UseHandler(lb)
						}
						backends[frontend.Backend] = negroni
					} else {
						log.Debugf("Reusing backend %s", frontend.Backend)
					}
					newRoute.Handler(backends[frontend.Backend])
				}
				err := newRoute.GetError()
				if err != nil {
					log.Errorf("Error building route: %s", err)
				}
			}
		}
	}
	return serverEntryPoints, nil
}

func (server *Server) loadEntryPointConfig(entryPointName string, entryPoint *EntryPoint) (http.Handler, error) {
	regex := entryPoint.Redirect.Regex
	replacement := entryPoint.Redirect.Replacement
	if len(entryPoint.Redirect.EntryPoint) > 0 {
		regex = "^(?:https?:\\/\\/)?([\\da-z\\.-]+)(?::\\d+)(.*)$"
		if server.globalConfiguration.EntryPoints[entryPoint.Redirect.EntryPoint] == nil {
			return nil, errors.New("Unkown entrypoint " + entryPoint.Redirect.EntryPoint)
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
	negroni := negroni.New()
	negroni.Use(rewrite)
	return negroni, nil
}

func (server *Server) buildDefaultHTTPRouter() *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	return router
}
