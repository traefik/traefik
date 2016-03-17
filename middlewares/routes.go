package middlewares

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
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
	next(rw, r)
	routeMatch := mux.RouteMatch{}
	if router.router.Match(r, &routeMatch) {
		frontendName := routeMatch.Route.GetName()
		saveNameForLogger(r, loggerFrontend, strings.TrimPrefix(frontendName, "frontend-"))
	}
}
