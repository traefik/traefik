package main

import (
	"github.com/BurntSushi/toml"
	"github.com/codegangsta/negroni"
	"github.com/emilevauge/traefik/middlewares"
	"github.com/gorilla/mux"
	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/roundrobin"
	"github.com/op/go-logging"
	"github.com/thoas/stats"
	"github.com/tylerb/graceful"
	"github.com/unrolled/render"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
	fmtlog "log"
	"github.com/mailgun/oxy/cbreaker"
)

var (
	globalConfigFile     = kingpin.Arg("conf", "Main configration file.").Default("traefik.toml").String()
	currentConfiguration = new(Configuration)
	metrics              = stats.New()
	log                  = logging.MustGetLogger("traefik")
	oxyLogger			 = &OxyLogger{}
	templatesRenderer    = render.New(render.Options{
		Directory:  "templates",
		Asset:      Asset,
		AssetNames: AssetNames,
	})
)

func main() {
	kingpin.Parse()
	fmtlog.SetFlags(fmtlog.Lshortfile | fmtlog.LstdFlags)
	var srv *graceful.Server
	var configurationRouter *mux.Router
	var configurationChan = make(chan *Configuration, 10)
	defer close(configurationChan)
	var providers = []Provider{}
	var format = logging.MustStringFormatter("%{color}%{time:15:04:05.000} %{shortfile:20.20s} %{level:8.8s} %{id:03x} ▶%{color:reset} %{message}")
	var sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// load global configuration
	gloablConfiguration := LoadFileConfig(*globalConfigFile)

	loggerMiddleware := middlewares.NewLogger(gloablConfiguration.AccessLogsFile)
	defer loggerMiddleware.Close()

	// logging
	backends := []logging.Backend{}
	level, err := logging.LogLevel(gloablConfiguration.LogLevel)
	if err != nil {
		log.Fatal("Error getting level", err)
	}

	if len(gloablConfiguration.TraefikLogsFile) > 0 {
		fi, err := os.OpenFile(gloablConfiguration.TraefikLogsFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		defer fi.Close()
		if err != nil {
			log.Fatal("Error opening file", err)
		} else {
			logBackend := logging.NewLogBackend(fi, "", 0)
			logBackendFormatter := logging.NewBackendFormatter(logBackend, logging.GlogFormatter)
			backends = append(backends, logBackendFormatter)
		}
	}
	if gloablConfiguration.TraefikLogsStdout {
		logBackend := logging.NewLogBackend(os.Stdout, "", 0)
		logBackendFormatter := logging.NewBackendFormatter(logBackend, format)
		backends = append(backends, logBackendFormatter)
	}
	logging.SetBackend(backends...)
	logging.SetLevel(level, "traefik")

	configurationRouter = LoadDefaultConfig(gloablConfiguration)

	// listen new configurations from providers
	go func() {
		for {
			configuration := <-configurationChan
			log.Info("Configuration receveived %+v", configuration)
			if configuration == nil {
				log.Info("Skipping empty configuration")
			} else if reflect.DeepEqual(currentConfiguration, configuration) {
				log.Info("Skipping same configuration")
			} else {
				newConfigurationRouter, err := LoadConfig(configuration, gloablConfiguration)
				if (err == nil ){
					currentConfiguration = configuration
					configurationRouter = newConfigurationRouter
					srv.Stop(time.Duration(gloablConfiguration.GraceTimeOut) * time.Second)
					time.Sleep(3 * time.Second)
				}else{
					log.Error("Error loading new configuration, aborted ", err)
				}
			}
		}
	}()

	// configure providers
	if gloablConfiguration.Docker != nil {
		providers = append(providers, gloablConfiguration.Docker)
	}
	if gloablConfiguration.Marathon != nil {
		providers = append(providers, gloablConfiguration.Marathon)
	}
	if gloablConfiguration.File != nil {
		if len(gloablConfiguration.File.Filename) == 0 {
			// no filename, setting to global config file
			gloablConfiguration.File.Filename = *globalConfigFile
		}
		providers = append(providers, gloablConfiguration.File)
	}
	if gloablConfiguration.Web != nil {
		providers = append(providers, gloablConfiguration.Web)
	}
	if gloablConfiguration.Consul != nil {
		providers = append(providers, gloablConfiguration.Consul)
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

		// middlewares
		var negroni = negroni.New()
		negroni.Use(metrics)
		negroni.Use(loggerMiddleware)
		//negroni.Use(middlewares.NewCircuitBreaker(oxyLogger))
		//negroni.Use(middlewares.NewRoutes(configurationRouter))
		negroni.UseHandler(configurationRouter)

		srv = &graceful.Server{
			Timeout:          time.Duration(gloablConfiguration.GraceTimeOut) * time.Second,
			NoSignalHandling: true,

			Server: &http.Server{
				Addr:    gloablConfiguration.Port,
				Handler: negroni,
			},
		}

		go func() {
			if len(gloablConfiguration.CertFile) > 0 && len(gloablConfiguration.KeyFile) > 0 {
				err := srv.ListenAndServeTLS(gloablConfiguration.CertFile, gloablConfiguration.KeyFile)
				if (err!= nil ){
					log.Fatal("Error creating server: ", err)
				}
			} else {
				err := srv.ListenAndServe()
				if (err!= nil ){
					log.Fatal("Error creating server: ", err)
				}
			}
		}()
		log.Notice("Started")
		<-srv.StopChan()
		log.Notice("Stopped")
	}
}

func LoadConfig(configuration *Configuration, gloablConfiguration *GlobalConfiguration) (*mux.Router, error) {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	backends := map[string]http.Handler{}
	for frontendName, frontend := range configuration.Frontends {
		log.Debug("Creating frontend %s", frontendName)
		fwd, _ := forward.New(forward.Logger(oxyLogger))
		newRoute := router.NewRoute().Name(frontendName)
		for routeName, route := range frontend.Routes {
			log.Debug("Creating route %s %s:%s", routeName, route.Rule, route.Value)
			newRouteReflect := Invoke(newRoute, route.Rule, route.Value)
			newRoute = newRouteReflect[0].Interface().(*mux.Route)
		}
		if backends[frontend.Backend] == nil {
			log.Debug("Creating backend %s", frontend.Backend)
			lb, _ := roundrobin.New(fwd)
			rb, _ := roundrobin.NewRebalancer(lb, roundrobin.RebalancerLogger(oxyLogger))
			for serverName, server := range configuration.Backends[frontend.Backend].Servers {
				if url, err := url.Parse(server.Url); err != nil{
					return nil, err
				}else{
					log.Debug("Creating server %s %s", serverName, url.String())
					rb.UpsertServer(url, roundrobin.Weight(server.Weight))
				}
			}
			backends[frontend.Backend] = rb
		} else {
			log.Debug("Reusing backend %s", frontend.Backend)
		}
//		stream.New(backends[frontend.Backend], stream.Retry("IsNetworkError() && Attempts() <= " + strconv.Itoa(gloablConfiguration.Replay)), stream.Logger(oxyLogger))
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
	log.Debug("Global configuration loaded %+v", configuration)
	return configuration
}