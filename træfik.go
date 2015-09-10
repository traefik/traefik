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
	"log"
	"github.com/BurntSushi/toml"
	"github.com/gorilla/handlers"
)

var srv *graceful.Server
var configurationRouter *mux.Router
var currentConfiguration = new(Configuration)
var configurationChan = make(chan *Configuration)
var providers = []Provider{}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	globalConfigFile := "træfik.toml"
	configurationRouter = LoadDefaultConfig()
	go func() {
		for {
			configuration := <-configurationChan
			log.Println("Configuration receveived", configuration)
			if configuration == nil {
				log.Println("Skipping empty configuration")
			} else if (reflect.DeepEqual(currentConfiguration, configuration)) {
				log.Println("Skipping same configuration")
			} else {
				currentConfiguration = configuration
				configurationRouter = LoadConfig(configuration)
				srv.Stop(10 * time.Second)
				time.Sleep(3 * time.Second)
			}
		}
	}()

	configuration := LoadFileConfig(globalConfigFile)
	if (configuration.Docker != nil) {
		providers = append(providers, configuration.Docker)
	}

	if (configuration.Marathon != nil) {
		providers = append(providers, configuration.Marathon)
	}

	if (configuration.File != nil) {
		if (len(configuration.File.Filename) == 0) {
			// no filename, setting to global config file
			configuration.File.Filename = globalConfigFile
		}
		providers = append(providers, configuration.File)
	}

	if (configuration.Web != nil) {
		providers = append(providers, configuration.Web)
	}

	for _, provider := range providers {
		log.Printf("Starting provider %v %+v\n", reflect.TypeOf(provider), provider)
		currentProvider := provider
		go func() {
			currentProvider.Provide(configurationChan)
		}()
	}

	goAway := false
	go func() {
		sig := <-sigs
		log.Println("I have to go...", sig)
		goAway = true
		srv.Stop(time.Duration(configuration.GraceTimeOut) * time.Second)
	}()

	for {
		if goAway {
			break
		}
		srv = &graceful.Server{
			Timeout:          time.Duration(configuration.GraceTimeOut) * time.Second,
			NoSignalHandling: true,

			Server: &http.Server{
				Addr:    configuration.Port,
				Handler: configurationRouter,
			},
		}

		go func() {
			srv.ListenAndServe()
		}()
		log.Println("Started")
		<-srv.StopChan()
		log.Println("Stopped")
	}
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	renderer.HTML(w, http.StatusNotFound, "notFound", nil)
}

func LoadDefaultConfig() *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(notFoundHandler))
	return router
}

func LoadConfig(configuration *Configuration) *mux.Router {
	router := mux.NewRouter()
	router.NotFoundHandler = handlers.CombinedLoggingHandler(os.Stdout, http.HandlerFunc(notFoundHandler))
	backends := map[string]http.Handler{}
	for routeName, route := range configuration.Routes {
		log.Println("Creating route", routeName)
		fwd, _ := forward.New()
		newRoutes := []*mux.Route{}
		for ruleName, rule := range route.Rules {
			log.Println("Creating rule", ruleName)
			newRouteReflect := Invoke(router.NewRoute(), rule.Category, rule.Value)
			newRoute := newRouteReflect[0].Interface().(*mux.Route)
			newRoutes = append(newRoutes, newRoute)
		}
		if (backends[route.Backend] ==nil) {
			log.Println("Creating backend", route.Backend)
			lb, _ := roundrobin.New(fwd)
			rb, _ := roundrobin.NewRebalancer(lb)
			for serverName, server := range configuration.Backends[route.Backend].Servers {
				log.Println("Creating server", serverName)
				url, _ := url.Parse(server.Url)
				rb.UpsertServer(url, roundrobin.Weight(server.Weight))
			}
			backends[route.Backend]=lb
		}else {
			log.Println("Reusing backend", route.Backend)
		}
		for _, muxRoute := range newRoutes {
			muxRoute.Handler(handlers.CombinedLoggingHandler(os.Stdout, backends[route.Backend]))
			err := muxRoute.GetError()
			if err != nil {
				log.Println("Error building route", err)
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
		log.Fatal("Error reading file:", err)
	}
	log.Printf("Global configuration loaded %+v\n", configuration)
	return configuration
}