/*
Copyright
*/
package main

import (
	"crypto/tls"
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
	"sync"
	"syscall"
	"time"
)

// Server is the reverse-proxy/load-balancer engine
type Server struct {
	srv                        *manners.GracefulServer
	configurationRouter        *mux.Router
	configurationChan          chan types.ConfigMessage
	configurationChanValidated chan types.ConfigMessage
	sigs                       chan os.Signal
	stopChan                   chan bool
	providers                  []provider.Provider
	serverLock                 sync.Mutex
	currentConfigurations      configs
	globalConfiguration        GlobalConfiguration
	loggerMiddleware           *middlewares.Logger
}

// NewServer returns an initialized Server.
func NewServer(globalConfiguration GlobalConfiguration) *Server {
	server := new(Server)

	server.configurationChan = make(chan types.ConfigMessage, 10)
	server.configurationChanValidated = make(chan types.ConfigMessage, 10)
	server.sigs = make(chan os.Signal, 1)
	server.stopChan = make(chan bool)
	server.providers = []provider.Provider{}
	signal.Notify(server.sigs, syscall.SIGINT, syscall.SIGTERM)
	server.currentConfigurations = make(configs)
	server.globalConfiguration = globalConfiguration
	server.loggerMiddleware = middlewares.NewLogger(globalConfiguration.AccessLogsFile)

	return server
}

// Start starts the server and blocks until server is shutted down.
func (server *Server) Start() {
	server.configurationRouter = LoadDefaultConfig(server.globalConfiguration)
	go server.listenProviders()
	go server.enableRouter()
	server.configureProviders()
	server.startProviders()
	go server.listenSignals()

	var er error
	server.serverLock.Lock()
	server.srv, er = server.prepareServer(server.configurationRouter, server.globalConfiguration, nil, server.loggerMiddleware, metrics)
	if er != nil {
		log.Fatal("Error preparing server: ", er)
	}
	go server.startServer(server.srv, server.globalConfiguration)
	//TODO change that!
	time.Sleep(100 * time.Millisecond)
	server.serverLock.Unlock()

	<-server.stopChan
}

// Stop stops the server
func (server *Server) Stop() {
	server.srv.Close()
	server.stopChan <- true
}

// Close destroys the server
func (server *Server) Close() {
	defer close(server.configurationChan)
	defer close(server.configurationChanValidated)
	defer close(server.sigs)
	defer close(server.stopChan)
	defer server.loggerMiddleware.Close()
}

func (server *Server) listenProviders() {
	lastReceivedConfiguration := time.Unix(0, 0)
	lastConfigs := make(map[string]*types.ConfigMessage)
	for {
		configMsg := <-server.configurationChan
		log.Infof("Configuration receveived from provider %s: %#v", configMsg.ProviderName, configMsg.Configuration)
		lastConfigs[configMsg.ProviderName] = &configMsg
		if time.Now().After(lastReceivedConfiguration.Add(time.Duration(server.globalConfiguration.ProvidersThrottleDuration))) {
			log.Infof("Last %s config received more than %s, OK", configMsg.ProviderName, server.globalConfiguration.ProvidersThrottleDuration)
			// last config received more than n s ago
			server.configurationChanValidated <- configMsg
		} else {
			log.Infof("Last %s config received less than %s, waiting...", configMsg.ProviderName, server.globalConfiguration.ProvidersThrottleDuration)
			go func() {
				<-time.After(server.globalConfiguration.ProvidersThrottleDuration)
				if time.Now().After(lastReceivedConfiguration.Add(time.Duration(server.globalConfiguration.ProvidersThrottleDuration))) {
					log.Infof("Waited for %s config, OK", configMsg.ProviderName)
					server.configurationChanValidated <- *lastConfigs[configMsg.ProviderName]
				}
			}()
		}
		lastReceivedConfiguration = time.Now()
	}
}

func (server *Server) enableRouter() {
	for {
		configMsg := <-server.configurationChanValidated
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

			newConfigurationRouter, err := server.loadConfig(newConfigurations, server.globalConfiguration)
			if err == nil {
				server.serverLock.Lock()
				server.currentConfigurations = newConfigurations
				server.configurationRouter = newConfigurationRouter
				oldServer := server.srv
				newsrv, err := server.prepareServer(server.configurationRouter, server.globalConfiguration, oldServer, server.loggerMiddleware, metrics)
				if err != nil {
					log.Fatal("Error preparing server: ", err)
				}
				go server.startServer(newsrv, server.globalConfiguration)
				server.srv = newsrv
				time.Sleep(1 * time.Second)
				if oldServer != nil {
					log.Info("Stopping old server")
					oldServer.Close()
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
		if len(server.globalConfiguration.File.Filename) == 0 {
			// no filename, setting to global config file
			server.globalConfiguration.File.Filename = *globalConfigFile
		}
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
		log.Infof("Starting provider %v %+v", reflect.TypeOf(provider), provider)
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
	sig := <-server.sigs
	log.Infof("I have to go... %+v", sig)
	log.Info("Stopping server")
	server.Stop()
}

// creates a TLS config that allows terminating HTTPS for multiple domains using SNI
func (server *Server) createTLSConfig(certs []Certificate) (*tls.Config, error) {
	if len(certs) == 0 {
		return nil, nil
	}

	config := &tls.Config{}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	var err error
	config.Certificates = make([]tls.Certificate, len(certs))
	for i, v := range certs {
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
	log.Info("Starting server")
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

func (server *Server) prepareServer(router *mux.Router, globalConfiguration GlobalConfiguration, oldServer *manners.GracefulServer, middlewares ...negroni.Handler) (*manners.GracefulServer, error) {
	log.Info("Preparing server")
	// middlewares
	var negroni = negroni.New()
	for _, middleware := range middlewares {
		negroni.Use(middleware)
	}
	negroni.UseHandler(router)
	tlsConfig, err := server.createTLSConfig(globalConfiguration.Certificates)
	if err != nil {
		log.Fatalf("Error creating TLS config %s", err)
		return nil, err
	}

	if oldServer == nil {
		return manners.NewWithServer(
			&http.Server{
				Addr:      globalConfiguration.Port,
				Handler:   negroni,
				TLSConfig: tlsConfig,
			}), nil
	}
	gracefulServer, err := oldServer.HijackListener(&http.Server{
		Addr:      globalConfiguration.Port,
		Handler:   negroni,
		TLSConfig: tlsConfig,
	}, tlsConfig)
	if err != nil {
		log.Fatalf("Error hijacking server %s", err)
		return nil, err
	}
	return gracefulServer, nil
}

// LoadConfig returns a new gorilla.mux Route from the specified global configuration and the dynamic
// provider configurations.
func (server *Server) loadConfig(configurations configs, globalConfiguration GlobalConfiguration) (*mux.Router, error) {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	backends := map[string]http.Handler{}
	for _, configuration := range configurations {
		for frontendName, frontend := range configuration.Frontends {
			log.Debugf("Creating frontend %s", frontendName)
			fwd, _ := forward.New(forward.Logger(oxyLogger), forward.PassHostHeader(frontend.PassHostHeader))
			newRoute := router.NewRoute().Name(frontendName)
			for routeName, route := range frontend.Routes {
				log.Debugf("Creating route %s %s:%s", routeName, route.Rule, route.Value)
				newRouteReflect, err := invoke(newRoute, route.Rule, route.Value)
				if err != nil {
					return nil, err
				}
				newRoute = newRouteReflect[0].Interface().(*mux.Route)
			}
			if backends[frontend.Backend] == nil {
				log.Debugf("Creating backend %s", frontend.Backend)
				var lb http.Handler
				rr, _ := roundrobin.New(fwd)
				if configuration.Backends[frontend.Backend] == nil {
					return nil, errors.New("Backend not found: " + frontend.Backend)
				}
				lbMethod, err := types.NewLoadBalancerMethod(configuration.Backends[frontend.Backend].LoadBalancer)
				if err != nil {
					configuration.Backends[frontend.Backend].LoadBalancer = &types.LoadBalancer{Method: "wrr"}
				}
				switch lbMethod {
				case types.Drr:
					log.Infof("Creating load-balancer drr")
					rebalancer, _ := roundrobin.NewRebalancer(rr, roundrobin.RebalancerLogger(oxyLogger))
					lb = rebalancer
					for serverName, server := range configuration.Backends[frontend.Backend].Servers {
						url, err := url.Parse(server.URL)
						if err != nil {
							return nil, err
						}
						log.Infof("Creating server %s %s", serverName, url.String())
						rebalancer.UpsertServer(url, roundrobin.Weight(server.Weight))
					}
				case types.Wrr:
					log.Infof("Creating load-balancer wrr")
					lb = middlewares.NewWebsocketUpgrader(rr)
					for serverName, server := range configuration.Backends[frontend.Backend].Servers {
						url, err := url.Parse(server.URL)
						if err != nil {
							return nil, err
						}
						log.Infof("Creating server %s %s", serverName, url.String())
						rr.UpsertServer(url, roundrobin.Weight(server.Weight))
					}
				}
				var negroni = negroni.New()
				if configuration.Backends[frontend.Backend].CircuitBreaker != nil {
					log.Infof("Creating circuit breaker %s", configuration.Backends[frontend.Backend].CircuitBreaker.Expression)
					negroni.Use(middlewares.NewCircuitBreaker(lb, configuration.Backends[frontend.Backend].CircuitBreaker.Expression, cbreaker.Logger(oxyLogger)))
				} else {
					negroni.UseHandler(lb)
				}
				backends[frontend.Backend] = negroni
			} else {
				log.Infof("Reusing backend %s", frontend.Backend)
			}
			//		stream.New(backends[frontend.Backend], stream.Retry("IsNetworkError() && Attempts() <= " + strconv.Itoa(globalConfiguration.Replay)), stream.Logger(oxyLogger))

			newRoute.Handler(backends[frontend.Backend])
			err := newRoute.GetError()
			if err != nil {
				log.Errorf("Error building route: %s", err)
			}
		}
	}
	return router, nil
}
