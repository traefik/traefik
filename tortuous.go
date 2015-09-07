package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/mux"
	"github.com/leekchan/gtf"
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
)

var srv *graceful.Server
var userRouter *mux.Router
var renderer = render.New()
var currentService = new(Service)
var serviceChan = make(chan Service)
var providers = []Provider{}

func main() {
	providers = append(providers, new(DockerProvider))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	systemRouter := mux.NewRouter()
	systemRouter.Methods("POST").Path("/reload").HandlerFunc(ReloadConfigHandler)
	systemRouter.Methods("GET").Path("/").HandlerFunc(GetConfigHandler)
	go http.ListenAndServe(":8000", systemRouter)

	go func() {
		for {
			service := <-serviceChan
			fmt.Println("Service receveived", service)
			currentService = &service
			userRouter = LoadConfig(service)
			srv.Stop(10 * time.Second)
		}
	}()

	for _, provider := range providers {
		provider.Provide(serviceChan)
	}

	goAway := false
	go func() {
		sig := <-sigs
		fmt.Println("I have to go...", sig)
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
				Handler: userRouter,
			},
		}

		go srv.ListenAndServe()
		fmt.Println("Started")
		<-srv.StopChan()
		fmt.Println("Stopped")
	}
}

func LoadDockerConfig(client *docker.Client, service Service) {
	containerList, _ := client.ListContainers(docker.ListContainersOptions{})
	containersInspected := []docker.Container{}
	for _, container := range containerList {
		containerInspected, _ := client.InspectContainer(container.ID)
		containersInspected = append(containersInspected, *containerInspected)
	}
	containers := struct {
		Containers []docker.Container
	}{
		containersInspected,
	}
	tmpl, err := gtf.New("docker.tmpl").ParseFiles("docker.tmpl")
	if err != nil {
		panic(err)
	}

	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, containers)
	if err != nil {
		panic(err)
	}

	fmt.Println(buffer.String())

	if _, err := toml.Decode(buffer.String(), service); err != nil {
		fmt.Println(err)
		return
	}
}

func LoadFileConfig(service Service) {
	if _, err := toml.DecodeFile("tortuous.toml", service); err != nil {
		fmt.Println(err)
		return
	}
}

func LoadConfig(service Service) *mux.Router {
	/*endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)
	dockerEvents := make(chan *docker.APIEvents)
	LoadDockerConfig(client)
	client.AddEventListener(dockerEvents)
	go func() {
		for {
			event := <-dockerEvents
			fmt.Println("Event receveived", event)
		}
	}()*/
	//LoadFileConfig()

	router := mux.NewRouter()
	for routeName, route := range service.Routes {
		fmt.Println("Creating route", routeName)
		fwd, _ := forward.New()
		newRoutes := []*mux.Route{}
		for ruleName, rule := range route.Rules {
			fmt.Println("Creating rule", ruleName)
			newRouteReflect := Invoke(router.NewRoute(), rule.Category, rule.Value)
			newRoute := newRouteReflect[0].Interface().(*mux.Route)
			newRoutes = append(newRoutes, newRoute)
		}
		for _, backendName := range route.Backends {
			fmt.Println("Creating backend", backendName)
			lb, _ := roundrobin.New(fwd)
			rb, _ := roundrobin.NewRebalancer(lb)
			for serverName, server := range service.Backends[backendName].Servers {
				fmt.Println("Creating server", serverName)
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
	userRouter = LoadConfig(*currentService)
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
