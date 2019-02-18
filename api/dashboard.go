package api

import (
	"net/http"

	"github.com/containous/mux"
	"github.com/containous/traefik/log"
	assetfs "github.com/elazarl/go-bindata-assetfs"
)

// DashboardHandler expose dashboard routes
type DashboardHandler struct {
	Assets *assetfs.AssetFS
}

// AddRoutes add dashboard routes on a router
func (g DashboardHandler) AddRoutes(router *mux.Router) {
	if g.Assets == nil {
		log.Error("No assets for dashboard")
		return
	}

	// Expose dashboard
	router.Methods(http.MethodGet).
		Path("/").
		HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			http.Redirect(response, request, request.Header.Get("X-Forwarded-Prefix")+"/dashboard/", 302)
		})

	router.Methods(http.MethodGet).
		Path("/dashboard/status").
		HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			http.Redirect(response, request, "/dashboard/", 302)
		})

	router.Methods(http.MethodGet).
		PathPrefix("/dashboard/").
		Handler(http.StripPrefix("/dashboard/", http.FileServer(g.Assets)))
}
