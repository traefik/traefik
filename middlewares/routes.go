/*
Copyright
*/
package middlewares

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Routes struct {
	router *mux.Router
}

func NewRoutes(router *mux.Router) *Routes {
	return &Routes{router}
}

func (router *Routes) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	routeMatch := mux.RouteMatch{}
	if router.router.Match(r, &routeMatch) {
		json, _ := json.Marshal(routeMatch.Handler)
		log.Println("Request match route ", json)
	}
	next(rw, r)
}
