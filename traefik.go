package main

import (
	fmtlog "log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/emilevauge/traefik/middlewares"
	"github.com/gorilla/mux"
	"github.com/mailgun/manners"
	"github.com/mailgun/oxy/cbreaker"
	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/roundrobin"
	"github.com/thoas/stats"
	"github.com/unrolled/render"
	"gopkg.in/alecthomas/kingpin.v2"
	"runtime"
)

var (
	globalConfigFile     = kingpin.Arg("conf", "Main configration file.").Default("traefik.toml").String()
	currentConfiguration = new(Configuration)
	metrics              = stats.New()
	oxyLogger            = &OxyLogger{}
	templatesRenderer    = render.New(render.Options{
		Directory:  "templates",
		Asset:      Asset,
		AssetNames: AssetNames,
	})
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	kingpin.Parse()
	fmtlog.SetFlags(fmtlog.Lshortfile | fmtlog.LstdFlags)
	var srv *manners.GracefulServer
	var configurationRouter *mux.Router
	var configurationChan = make(chan *Configuration, 10)
	defer close(configurationChan)
	var sigs = make(chan os.Signal, 1)
	defer close(sigs)
	var stopChan = make(chan bool)
	defer close(stopChan)
	var providers = []Provider{}
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// load global configuration
	globalConfiguration := LoadFileConfig(*globalConfigFile)

	loggerMiddleware := middlewares.NewLogger(globalConfiguration.AccessLogsFile)
	defer loggerMiddleware.Close()

	// logging
	level, err := log.ParseLevel(strings.ToLower(globalConfiguration.LogLevel))
	if err != nil {
		log.Fatal("Error getting level", err)
	}
	log.SetLevel(level)

	if len(globalConfiguration.TraefikLogsFile) > 0 {
		fi, err := os.OpenFile(globalConfiguration.TraefikLogsFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		defer fi.Close()
		if err != nil {
			log.Fatal("Error opening file", err)
		} else {
			log.SetOutput(fi)
			log.SetFormatter(&log.TextFormatter{DisableColors: true, FullTimestamp: true, DisableSorting: true})
		}
	} else {
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableSorting: true})
	}

	configurationRouter = LoadDefaultConfig(globalConfiguration)

	// listen new configurations from providers
	go func() {
		for {
			configuration := <-configurationChan
			log.Infof("Configuration receveived %+v", configuration)
			if configuration == nil {
				log.Info("Skipping empty configuration")
			} else if reflect.DeepEqual(currentConfiguration, configuration) {
				log.Info("Skipping same configuration")
			} else {
				newConfigurationRouter, err := LoadConfig(configuration, globalConfiguration)
				if err == nil {
					currentConfiguration = configuration
					configurationRouter = newConfigurationRouter
					oldServer := srv
					newsrv := prepareServer(configurationRouter, globalConfiguration, oldServer, loggerMiddleware, metrics)
					go startServer(newsrv, globalConfiguration)
					srv = newsrv
					time.Sleep(2 * time.Second)
					if oldServer != nil {
						log.Info("Stopping old server")
						oldServer.Close()
					}
				} else {
					log.Error("Error loading new configuration, aborted ", err)
				}
			}
		}
	}()

	// configure providers
	if globalConfiguration.Docker != nil {
		providers = append(providers, globalConfiguration.Docker)
	}
	if globalConfiguration.Marathon != nil {
		providers = append(providers, globalConfiguration.Marathon)
	}
	if globalConfiguration.File != nil {
		if len(globalConfiguration.File.Filename) == 0 {
			// no filename, setting to global config file
			globalConfiguration.File.Filename = *globalConfigFile
		}
		providers = append(providers, globalConfiguration.File)
	}
	if globalConfiguration.Web != nil {
		providers = append(providers, globalConfiguration.Web)
	}
	if globalConfiguration.Consul != nil {
		providers = append(providers, globalConfiguration.Consul)
	}

	// start providers
	for _, provider := range providers {
		log.Infof("Starting provider %v %+v", reflect.TypeOf(provider), provider)
		currentProvider := provider
		go func() {
			currentProvider.Provide(configurationChan)
		}()
	}

	go func() {
		sig := <-sigs
		log.Infof("I have to go... %+v", sig)
		log.Info("Stopping server")
		srv.Close()
		stopChan <- true
	}()

	//negroni.Use(middlewares.NewCircuitBreaker(oxyLogger))
	//negroni.Use(middlewares.NewRoutes(configurationRouter))
	srv = prepareServer(configurationRouter, globalConfiguration, nil, loggerMiddleware, metrics)
	go startServer(srv, globalConfiguration)

	<-stopChan
	log.Info("Shutting down")
}

func startServer(srv *manners.GracefulServer, globalConfiguration *GlobalConfiguration) {
	log.Info("Starting server")
	if len(globalConfiguration.CertFile) > 0 && len(globalConfiguration.KeyFile) > 0 {
		err := srv.ListenAndServeTLS(globalConfiguration.CertFile, globalConfiguration.KeyFile)
		if err != nil {
			netOpError, ok := err.(*net.OpError)
			if ok && netOpError.Err.Error() != "use of closed network connection" {
				log.Fatal("Error creating server: ", err)
			}
		}
	} else {
		err := srv.ListenAndServe()
		if err != nil {
			netOpError, ok := err.(*net.OpError)
			if ok && netOpError.Err.Error() != "use of closed network connection" {
				log.Fatal("Error creating server: ", err)
			}
		}
	}
	log.Info("Server stopped")
}

func prepareServer(router *mux.Router, globalConfiguration *GlobalConfiguration, oldServer *manners.GracefulServer, middlewares ...negroni.Handler) *manners.GracefulServer {
	log.Info("Preparing server")
	// middlewares
	var negroni = negroni.New()
	for _, middleware := range middlewares {
		negroni.Use(middleware)
	}
	negroni.UseHandler(router)

	if oldServer == nil {
		return manners.NewWithServer(
			&http.Server{
				Addr:    globalConfiguration.Port,
				Handler: negroni,
			})
	} else {
		server, err := oldServer.HijackListener(&http.Server{
			Addr:    globalConfiguration.Port,
			Handler: negroni,
		}, nil)
		if err != nil {
			log.Fatalf("Error hijacking server %s", err)
			return nil
		} else {
			return server
		}
	}
}

func LoadConfig(configuration *Configuration, globalConfiguration *GlobalConfiguration) (*mux.Router, error) {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	backends := map[string]http.Handler{}
	for frontendName, frontend := range configuration.Frontends {
		log.Debugf("Creating frontend %s", frontendName)
		fwd, _ := forward.New(forward.Logger(oxyLogger))
		newRoute := router.NewRoute().Name(frontendName)
		for routeName, route := range frontend.Routes {
			log.Debugf("Creating route %s %s:%s", routeName, route.Rule, route.Value)
			newRouteReflect := Invoke(newRoute, route.Rule, route.Value)
			newRoute = newRouteReflect[0].Interface().(*mux.Route)
		}
		if backends[frontend.Backend] == nil {
			log.Debugf("Creating backend %s", frontend.Backend)
			lb, _ := roundrobin.New(fwd)
			rb, _ := roundrobin.NewRebalancer(lb, roundrobin.RebalancerLogger(oxyLogger))
			for serverName, server := range configuration.Backends[frontend.Backend].Servers {
				url, err := url.Parse(server.URL)
				if err != nil {
					return nil, err
				}
				log.Debugf("Creating server %s %s", serverName, url.String())
				rb.UpsertServer(url, roundrobin.Weight(server.Weight))
			}
			backends[frontend.Backend] = rb
		} else {
			log.Debugf("Reusing backend %s", frontend.Backend)
		}
		// stream.New(backends[frontend.Backend], stream.Retry("IsNetworkError() && Attempts() <= " + strconv.Itoa(globalConfiguration.Replay)), stream.Logger(oxyLogger))
		var negroni = negroni.New()
		negroni.Use(middlewares.NewCircuitBreaker(backends[frontend.Backend], cbreaker.Logger(oxyLogger)))
		newRoute.Handler(negroni)
		err := newRoute.GetError()
		if err != nil {
			log.Error("Error building route ", err)
		}
	}
	return router, nil
}

func Invoke(any interface{}, name string, args ...interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	return reflect.ValueOf(any).MethodByName(name).Call(inputs)
}

func LoadFileConfig(file string) *GlobalConfiguration {
	configuration := NewGlobalConfiguration()
	if _, err := toml.DecodeFile(file, configuration); err != nil {
		log.Fatal("Error reading file ", err)
	}
	log.Debugf("Global configuration loaded %+v", configuration)
	return configuration
}
