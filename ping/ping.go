package ping

import (
	"fmt"
	"github.com/containous/mux"
	"net/http"
)

type PingHandler struct {
	EntryPoint string `description:"Ping entrypoint Default: traefik"`
}

func (g PingHandler) AddRoutes(router *mux.Router) {
	router.Methods("GET", "HEAD").Path("/ping").HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		fmt.Fprint(response, "OK")
	})

}
