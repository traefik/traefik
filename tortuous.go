package main
import (
	"github.com/gorilla/mux"
	"github.com/tylerb/graceful"
	"net/http"
	"fmt"
	"os"
	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/roundrobin"
	"time"
	"net"
	"os/signal"
	"syscall"
	"github.com/BurntSushi/toml"
	"reflect"
	"net/url"
	"github.com/fsouza/go-dockerclient"
	"github.com/leekchan/gtf"
	"bytes"
	"github.com/unrolled/render"
)

type Backend struct {
	Servers map[string]Server
}

type Server struct {
	Url string
}

type Rule struct {
	Category string
	Value string
}

type Route struct {
	Backends []string
	Rules map[string]Rule
}

type Config struct {
	Backends map[string]Backend
	Routes map[string]Route
}

var srv *graceful.Server
var userRouter *mux.Router
var config = new(Config)
var renderer = render.New()

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	systemRouter := mux.NewRouter()
	systemRouter.Methods("POST").Path("/restart").HandlerFunc(RestartHandler)
	systemRouter.Methods("POST").Path("/reload").HandlerFunc(ReloadConfigHandler)
	systemRouter.Methods("GET").Path("/").HandlerFunc(GetConfigHandler)
	go http.ListenAndServe(":8000", systemRouter)

	userRouter = LoadConfig()

	goAway := false
	go func() {
		sig := <-sigs
		fmt.Println("I have to go...", sig)
		goAway = true
		srv.Stop(10 * time.Second)
	}()

	for{
		if (goAway){
			break
		}
		srv = &graceful.Server{
			Timeout: 10 * time.Second,
			NoSignalHandling: true,

			ConnState: func(conn net.Conn, state http.ConnState) {
				// conn has a new state
			},

			Server: &http.Server{
				Addr: ":8001",
				Handler: userRouter,
			},
		}

		go srv.ListenAndServe()
		fmt.Println("Started")
		<- srv.StopChan()
		fmt.Println("Stopped")
	}
}

func LoadDockerConfig(){
	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)
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
	if err != nil { panic(err) }

	var buffer bytes.Buffer

	err = tmpl.Execute(&buffer, containers)
	if err != nil { panic(err) }

	fmt.Println(buffer.String())

	if _, err := toml.Decode(buffer.String(), config); err != nil {
		fmt.Println(err)
		return
	}
}

func LoadFileConfig(){
	if _, err := toml.DecodeFile("tortuous.toml", config); err != nil {
		fmt.Println(err)
		return
	}
}


func LoadConfig() *mux.Router{
	//LoadDockerConfig()
	LoadFileConfig()

	router := mux.NewRouter()
	for routeName, route := range config.Routes {
		fmt.Println("Creating route", routeName)
		fwd, _ := forward.New()
		newRoutes:= []*mux.Route{}
		for ruleName, rule := range route.Rules{
			fmt.Println("Creating rule", ruleName)
			newRouteReflect := Invoke(router.NewRoute(), rule.Category, rule.Value)
			newRoute := newRouteReflect[0].Interface().(*mux.Route)
			newRoutes = append(newRoutes, newRoute)
		}
		for _, backendName := range route.Backends {
			fmt.Println("Creating backend", backendName)
			lb, _ := roundrobin.New(fwd)
			rb, _ := roundrobin.NewRebalancer(lb)
			for serverName, server := range config.Backends[backendName].Servers {
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

func ReloadConfigHandler(rw http.ResponseWriter, r *http.Request) {
	userRouter = LoadConfig()
	renderer.JSON(rw, http.StatusOK, map[string]interface{}{"status": "reloaded"})
}

func RestartHandler(rw http.ResponseWriter, r *http.Request) {
	srv.Stop(10 * time.Second)
	renderer.JSON(rw, http.StatusOK, map[string]interface{}{"status": "restarted"})
}

func GetConfigHandler(rw http.ResponseWriter, r *http.Request) {
	renderer.JSON(rw, http.StatusOK, config)
}

func Invoke(any interface{}, name string, args... interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	return reflect.ValueOf(any).MethodByName(name).Call(inputs)
}

