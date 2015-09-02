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
	"encoding/json"
	"reflect"
	"net/url"
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
	Servers map[string]Server
	Routes map[string]Route
}

var srv *graceful.Server
var userRouter *mux.Router
var config = new(Config)

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

func LoadConfig() *mux.Router{
	if metadata, err := toml.DecodeFile("tortuous.toml", config); err != nil {
		fmt.Println(err)
		return nil
	}else{
		fmt.Printf("Loaded config: %+v\n", metadata )
	}

	router := mux.NewRouter()
	for routeName, route := range config.Routes {
		fmt.Println("Creating route", routeName)
		fwd, _ := forward.New()
		newRoute:= router.NewRoute()
		for ruleName, rule := range route.Rules{
			fmt.Println("Creating rule", ruleName)
			newRoutes := Invoke(newRoute, rule.Category, rule.Value)
			newRoute = newRoutes[0].Interface().(*mux.Route)
		}
		for _, backendName := range route.Backends {
			fmt.Println("Creating backend", backendName)
			lb, _ := roundrobin.New(fwd)
			for serverName, server := range config.Backends[backendName].Servers {
				fmt.Println("Creating server", serverName)
				url, _ := url.Parse(server.Url)
				lb.UpsertServer(url)
			}
			newRoute.Handler(lb)
		}
	}
	return router
}

func ReloadConfigHandler(rw http.ResponseWriter, r *http.Request) {
	userRouter = LoadConfig()
}

func RestartHandler(rw http.ResponseWriter, r *http.Request) {
	srv.Stop(10 * time.Second)
}

func GetConfigHandler(rw http.ResponseWriter, r *http.Request) {
	if jsonRes, err := json.Marshal(config); err != nil {
		fmt.Println(err)
		return
	}else{
		fmt.Fprintf(rw, "%s", jsonRes)
	}
}

func Invoke(any interface{}, name string, args... interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	return reflect.ValueOf(any).MethodByName(name).Call(inputs)
}

