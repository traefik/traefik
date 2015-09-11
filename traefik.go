package main

import (
	"github.com/gorilla/mux"
	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/roundrobin"
	"github.com/tylerb/graceful"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
	"github.com/op/go-logging"
	"github.com/BurntSushi/toml"
	"github.com/gorilla/handlers"
)

var currentConfiguration = new(Configuration)
var log = logging.MustGetLogger("traefik")

func main() {
	var srv *graceful.Server
	var configurationRouter *mux.Router
	var configurationChan = make(chan *Configuration)
	var providers = []Provider{}
	var format = logging.MustStringFormatter("%{color}%{time:15:04:05.000} %{shortfile:20.20s} %{level:8.8s} %{id:03x} ▶%{color:reset} %{message}")
	var sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// load global configuration
	globalConfigFile := "traefik.toml"
	gloablConfiguration := LoadFileConfig(globalConfigFile)

	// logging
	backends := []logging.Backend{}
	level, err := logging.LogLevel(gloablConfiguration.LogLevel)
	if err != nil {
		log.Fatal("Error getting level", err)
	}

	if (len(gloablConfiguration.TraefikLogsFile) > 0 ) {
		fi, err := os.OpenFile(gloablConfiguration.TraefikLogsFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Error opening file", err)
		}else {
			logBackend := logging.NewLogBackend(fi, "", 0)
			logBackendFormatter := logging.NewBackendFormatter(logBackend, logging.GlogFormatter)
			logBackendLeveled := logging.AddModuleLevel(logBackend)
			logBackendLeveled.SetLevel(level, "")
			backends = append(backends, logBackendFormatter)
		}
	}
	if (gloablConfiguration.TraefikLogsStdout) {
		logBackend := logging.NewLogBackend(os.Stdout, "", 0)
		logBackendFormatter := logging.NewBackendFormatter(logBackend, format)
		logBackendLeveled := logging.AddModuleLevel(logBackend)
		logBackendLeveled.SetLevel(level, "")
		backends = append(backends, logBackendFormatter)
	}
	logging.SetBackend(backends...)


	configurationRouter = LoadDefaultConfig(gloablConfiguration)

	// listen new configurations from providers
	go func() {
		for {
			configuration := <-configurationChan
			log.Info("Configuration receveived %+v", configuration)
			if configuration == nil {
				log.Info("Skipping empty configuration")
			} else if (reflect.DeepEqual(currentConfiguration, configuration)) {
				log.Info("Skipping same configuration")
			} else {
				currentConfiguration = configuration
				configurationRouter = LoadConfig(configuration, gloablConfiguration)
				srv.Stop(10 * time.Second)
				time.Sleep(3 * time.Second)
			}
		}
	}()

	// configure providers
	if (gloablConfiguration.Docker != nil) {
		providers = append(providers, gloablConfiguration.Docker)
	}
	if (gloablConfiguration.Marathon != nil) {
		providers = append(providers, gloablConfiguration.Marathon)
	}
	if (gloablConfiguration.File != nil) {
		if (len(gloablConfiguration.File.Filename) == 0) {
			// no filename, setting to global config file
			gloablConfiguration.File.Filename = globalConfigFile
		}
		providers = append(providers, gloablConfiguration.File)
	}
	if (gloablConfiguration.Web != nil) {
		providers = append(providers, gloablConfiguration.Web)
	}

	// start providers
	for _, provider := range providers {
		log.Notice("Starting provider %v %+v", reflect.TypeOf(provider), provider)
		currentProvider := provider
		go func() {
			currentProvider.Provide(configurationChan)
		}()
	}

	goAway := false
	go func() {
		sig := <-sigs
		log.Notice("I have to go... %+v", sig)
		goAway = true
		srv.Stop(time.Duration(gloablConfiguration.GraceTimeOut) * time.Second)
	}()

	for {
		if goAway {
			break
		}
		srv = &graceful.Server{
			Timeout: time.Duration(gloablConfiguration.GraceTimeOut) * time.Second,
			NoSignalHandling: true,

			Server: &http.Server{
				Addr:    gloablConfiguration.Port,
				Handler: configurationRouter,
			},
		}

		go func() {
			if (len(gloablConfiguration.CertFile) > 0 && len(gloablConfiguration.KeyFile) > 0) {
				srv.ListenAndServeTLS(gloablConfiguration.CertFile, gloablConfiguration.KeyFile)
			} else {
				srv.ListenAndServe()
			}
		}()
		log.Notice("Started")
		<-srv.StopChan()
		log.Notice("Stopped")
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	renderer.HTML(w, http.StatusNotFound, "notFound", nil)
}

func LoadDefaultConfig(gloablConfiguration *GlobalConfiguration) *mux.Router {
	router := mux.NewRouter()
	if (len(gloablConfiguration.AccessLogsFile) > 0 ) {
		fi, err := os.OpenFile(gloablConfiguration.AccessLogsFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Error opening file", err)
		}
		router.NotFoundHandler = handlers.CombinedLoggingHandler(fi, http.HandlerFunc(notFoundHandler))
	}else {
		router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	}
	return router
}

func LoadConfig(configuration *Configuration, gloablConfiguration *GlobalConfiguration) *mux.Router {
	router := mux.NewRouter()
	if (len(gloablConfiguration.AccessLogsFile) > 0 ) {
		fi, err := os.OpenFile(gloablConfiguration.AccessLogsFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Error opening file", err)
		}
		router.NotFoundHandler = handlers.CombinedLoggingHandler(fi, http.HandlerFunc(notFoundHandler))
	}else {
		router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	}
	backends := map[string]http.Handler{}
	for routeName, route := range configuration.Routes {
		log.Debug("Creating route %s", routeName)
		fwd, _ := forward.New()
		newRoutes := []*mux.Route{}
		for ruleName, rule := range route.Rules {
			log.Debug("Creating rule %s", ruleName)
			newRouteReflect := Invoke(router.NewRoute(), rule.Category, rule.Value)
			newRoute := newRouteReflect[0].Interface().(*mux.Route)
			newRoutes = append(newRoutes, newRoute)
		}
		if (backends[route.Backend] ==nil) {
			log.Debug("Creating backend %s", route.Backend)
			lb, _ := roundrobin.New(fwd)
			rb, _ := roundrobin.NewRebalancer(lb)
			for serverName, server := range configuration.Backends[route.Backend].Servers {
				log.Debug("Creating server %s", serverName)
				url, _ := url.Parse(server.Url)
				rb.UpsertServer(url, roundrobin.Weight(server.Weight))
			}
			backends[route.Backend]=lb
		}else {
			log.Debug("Reusing backend", route.Backend)
		}
		for _, muxRoute := range newRoutes {
			if (len(gloablConfiguration.AccessLogsFile) > 0 ) {
				fi, err := os.OpenFile(gloablConfiguration.AccessLogsFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
				if err != nil {
					log.Fatal("Error opening file", err)
				}
				muxRoute.Handler(handlers.CombinedLoggingHandler(fi, backends[route.Backend]))
			}else {
				muxRoute.Handler(backends[route.Backend])
			}
			err := muxRoute.GetError()
			if err != nil {
				log.Error("Error building route", err)
			}
		}
	}
	return router
}

func Invoke(any interface{}, name string, args ...interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	return reflect.ValueOf(any).MethodByName(name).Call(inputs)
}

func LoadFileConfig(file string) *GlobalConfiguration {
	configuration := NewGlobalConfiguration()
	if _, err := toml.DecodeFile(file, configuration); err != nil {
		log.Fatal("Error reading file", err)
	}
	log.Debug("Global configuration loaded %+v", configuration)
	return configuration
}