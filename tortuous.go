package main
import (
	"github.com/gorilla/mux"
	"github.com/tylerb/graceful"
	"net/http"
	"fmt"
	"os"
	"github.com/mailgun/oxy/forward"
	"github.com/mailgun/oxy/roundrobin"
	"github.com/mailgun/oxy/testutils"
	"time"
	"net"
	"os/signal"
	"syscall"
	"github.com/BurntSushi/toml"
	"encoding/json"
)

type Backend struct {
	Servers []string
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
		fmt.Println("Started")
		srv = &graceful.Server{
			Timeout: 10 * time.Second,
			NoSignalHandling: true,

			ConnState: func(conn net.Conn, state http.ConnState) {
				fmt.Println( "Connection ", state)
			},

			Server: &http.Server{
				Addr: ":8001",
				Handler: userRouter,
			},
		}

		go srv.ListenAndServe()
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
			newRoute = newRoute.Host(rule.Value)
		}
		for _, backendName := range route.Backends {
			fmt.Println("Creating backend", backendName)
			lb, _ := roundrobin.New(fwd)
			for _, serverName := range config.Backends[backendName].Servers {
				fmt.Println("Creating server", serverName)
				lb.UpsertServer(testutils.ParseURI(config.Servers[serverName].Url))
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
