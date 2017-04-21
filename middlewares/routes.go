package middlewares

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/containous/mux"
	"github.com/containous/traefik/middlewares/common"
)

// Routes holds the gorilla mux routes (for the API & co).
type Routes struct {
	common.BasicMiddleware
	router *mux.Router
}

var _ common.Middleware = &Routes{}

// NewRoutes return a Routes based on the given router.
func NewRoutes(router *mux.Router) *Routes {
	return &Routes{common.BasicMiddleware{}, router}
}

func (router *Routes) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	routeMatch := mux.RouteMatch{}
	if router.router.Match(r, &routeMatch) {
		js, _ := json.Marshal(routeMatch.Handler)
		log.Println("Request match route ", js)
	}
	router.Next().ServeHTTP(rw, r)
}
