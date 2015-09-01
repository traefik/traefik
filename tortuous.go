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
)

type Backend struct {
	Name string
	Servers []string
}

type Server struct {
	Name string
	Url string
}

type Rule struct {
	Category string
	Value string
}

type Route struct {
	Name string
	Rules map[string]Rule
}

type Config struct {
	Backends []Backend
	Servers []Server
	Routes []Route
}

var srv *graceful.Server

func main() {

	var config Config
	if _, err := toml.DecodeFile("tortuous.toml", &config); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v\n", config )

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	systemRouter := mux.NewRouter()
	systemRouter.Methods("POST").Path("/").HandlerFunc(ReloadHandler)
	systemRouter.Methods("GET").Path("/").HandlerFunc(GetPidHandler)
	go http.ListenAndServe(":8000", systemRouter)



	fwd, _ := forward.New()
	lb, _ := roundrobin.New(fwd)

	lb.UpsertServer(testutils.ParseURI("http://172.17.0.2:80"))
	lb.UpsertServer(testutils.ParseURI("http://172.17.0.3:80"))

	userRouter := mux.NewRouter()
	userRouter.Host("test.zenika.fr").Handler(lb)

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

func ReloadHandler(rw http.ResponseWriter, r *http.Request) {
	srv.Stop(10 * time.Second)
}

func GetPidHandler(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "%d", os.Getpid())
}
