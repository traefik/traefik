package router

import "net/http"

//This interface captures all routing functionality required by vulcan.
//The routing functionality mainly comes from "github.com/vulcand/route",
type Router interface {

	//Sets the not-found handler (this handler is called when no other handlers/routes in the routing library match
	SetNotFound(http.Handler) error

	//Gets the not-found handler that is currently in use by this router.
	GetNotFound() http.Handler

	//Validates whether this is an acceptable route expression
	IsValid(string) bool

	//Adds a new route->handler combination. The route is a string which provides the routing expression. http.Handler is called when this expression matches a request.
	Handle(string, http.Handler) error

	//Removes a route. The http.Handler associated with it, will be discarded.
	Remove(string) error

	//ServiceHTTP is the http.Handler implementation that allows callers to route their calls to sub-http.Handlers based on route matches.
	ServeHTTP(http.ResponseWriter, *http.Request)
}
