package ping

import (
	"fmt"
	"net/http"

	"github.com/containous/mux"
)

type PingHandler struct {
	EntryPoint string `description:"Ping entrypoint Default: traefik"`
}

func (g PingHandler) AddRoutes(router *mux.Router) {
	router.Methods("GET", "HEAD").Path("/ping").HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		fmt.Fprint(response, "OK")
	})

}
