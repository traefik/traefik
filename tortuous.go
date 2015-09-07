package main

import (
	"github.com/gorilla/mux"
	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/roundrobin"
	"github.com/tylerb/graceful"
	"github.com/unrolled/render"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"
	"log"
)

var srv *graceful.Server
var systemRouter *mux.Router
var renderer = render.New()
var currentService = new(Service)
var serviceChan = make(chan *Service)
var providers = []Provider{}

func main() {
	//providers = append(providers, new(DockerProvider))
	providers = append(providers, new(FileProvider))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	systemRouter := mux.NewRouter()
	systemRouter.Methods("POST").Path("/reload").HandlerFunc(ReloadConfigHandler)
	systemRouter.Methods("GET").Path("/").HandlerFunc(GetConfigHandler)
	go http.ListenAndServe(":8000", systemRouter)

	go func() {
		for {
			service := <-serviceChan
			log.Println("Service receveived", service)
			if service == nil {
				log.Println("Skipping nil service")
			} else if(reflect.DeepEqual(currentService, service)){
				log.Println("Skipping same service")
			} else{
				currentService = service
				systemRouter = LoadConfig(service)
				srv.Stop(10 * time.Second)
				time.Sleep(3 * time.Second)
			}
		}
	}()

	go func() {
		for _, provider := range providers {
			provider.Provide(serviceChan)
		}
	}()

	goAway := false
	go func() {
		sig := <-sigs
		log.Println("I have to go...", sig)
		goAway = true
		srv.Stop(10 * time.Second)
	}()

	for {
		if goAway {
			break
		}
		srv = &graceful.Server{
			Timeout:          10 * time.Second,
			NoSignalHandling: true,

			ConnState: func(conn net.Conn, state http.ConnState) {
				// conn has a new state
			},

			Server: &http.Server{
				Addr:    ":8001",
				Handler: systemRouter,
			},
		}

		go srv.ListenAndServe()
		log.Println("Started")
		<-srv.StopChan()
		log.Println("Stopped")
	}
}

func LoadConfig(service *Service) *mux.Router {
	router := mux.NewRouter()
	for routeName, route := range service.Routes {
		log.Println("Creating route", routeName)
		fwd, _ := forward.New()
		newRoutes := []*mux.Route{}
		for ruleName, rule := range route.Rules {
			log.Println("Creating rule", ruleName)
			newRouteReflect := Invoke(router.NewRoute(), rule.Category, rule.Value)
			newRoute := newRouteReflect[0].Interface().(*mux.Route)
			newRoutes = append(newRoutes, newRoute)
		}
		for _, backendName := range route.Backends {
			log.Println("Creating backend", backendName)
			lb, _ := roundrobin.New(fwd)
			rb, _ := roundrobin.NewRebalancer(lb)
			for serverName, server := range service.Backends[backendName].Servers {
				log.Println("Creating server", serverName)
				url, _ := url.Parse(server.Url)
				rb.UpsertServer(url)
			}
			for _, route := range newRoutes {
				route.Handler(lb)
			}
		}
	}
	return router
}

func DeployService() {
	systemRouter = LoadConfig(currentService)
}

func ReloadConfigHandler(rw http.ResponseWriter, r *http.Request) {
	DeployService()
	srv.Stop(10 * time.Second)
	renderer.JSON(rw, http.StatusOK, map[string]interface{}{"status": "reloaded"})
}

func RestartHandler(rw http.ResponseWriter, r *http.Request) {
	renderer.JSON(rw, http.StatusOK, map[string]interface{}{"status": "restarted"})
}

func GetConfigHandler(rw http.ResponseWriter, r *http.Request) {
	renderer.JSON(rw, http.StatusOK, currentService)
}

func Invoke(any interface{}, name string, args ...interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	return reflect.ValueOf(any).MethodByName(name).Call(inputs)
}
