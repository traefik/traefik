package ping

import (
	"fmt"
	"net/http"

	"github.com/containous/mux"
)

//Handler expose ping routes
type Handler struct {
	EntryPoint string `description:"Ping entrypoint Default: traefik"`
}

// AddRoutes add ping routes on a router
func (g Handler) AddRoutes(router *mux.Router) {
	router.Methods("GET", "HEAD").Path("/ping").HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		fmt.Fprint(response, "OK")
	})

}
