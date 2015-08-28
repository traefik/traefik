package main
import (
	"github.com/gorilla/mux"
	"github.com/tylerb/graceful"
	"net/http"
	"fmt"
	"os"
	"github.com/wunderlist/moxy"
	"time"
	"net"
	"os/signal"
	"syscall"
)

var srv *graceful.Server

func main() {
	fmt.Println("Tortuous loading with pid", os.Getpid())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	systemRouter := mux.NewRouter()
	systemRouter.Methods("POST").Path("/").HandlerFunc(ReloadHandler)
	systemRouter.Methods("GET").Path("/").HandlerFunc(GetPidHandler)

	userRouter := mux.NewRouter()
		hosts := []string{"172.17.0.2"}
	filters := []moxy.FilterFunc{}
	proxy := moxy.NewReverseProxy(hosts, filters)
	userRouter.Host("test.zenika.fr").HandlerFunc(proxy.ServeHTTP)

	go http.ListenAndServe(":8000", systemRouter)

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
