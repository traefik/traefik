package middlewares

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/containous/mux"
)

// Routes holds the gorilla mux routes (for the API & co).
type Routes struct {
	router *mux.Router
}

// NewRoutes return a Routes based on the given router.
func NewRoutes(router *mux.Router) *Routes {
	return &Routes{router}
}

func (router *Routes) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	routeMatch := mux.RouteMatch{}
	if router.router.Match(r, &routeMatch) {
		rt, _ := json.Marshal(routeMatch.Handler)
		log.Println("Request match route ", rt)
	}
	next(rw, r)
}
