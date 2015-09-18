package main

import (
	"encoding/json"
	"github.com/gocraft/web"
	"github.com/thoas/stats"
	"net/http"
)

var Stats = stats.New()

type Context struct {
}

func (c *Context) Root(rw web.ResponseWriter, req *web.Request) {
	rw.Header().Set("Content-Type", "application/json")
	rw.Write([]byte("{\"hello\": \"world\"}"))
}

func (c *Context) RetrieveStats(rw web.ResponseWriter, req *web.Request) {
	rw.Header().Set("Content-Type", "application/json")

	stats := Stats.Data()

	b, _ := json.Marshal(stats)

	rw.Write(b)
}

func (c *Context) Recording(rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
	beginning, recorder := Stats.Begin(rw)

	next(recorder, req)

	Stats.End(beginning, recorder)
}

func main() {
	router := web.New(Context{}). // Create your router
					Middleware(web.LoggerMiddleware).       // Use some included middleware
					Middleware(web.ShowErrorsMiddleware).   // ...
					Middleware((*Context).Recording).       // Your own middleware!
					Get("/", (*Context).Root).              // Add a route
					Get("/stats", (*Context).RetrieveStats) // Add a route
	http.ListenAndServe("localhost:3001", router) // Start the server!
}
