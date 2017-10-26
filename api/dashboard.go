package api

import (
	"net/http"

	"github.com/containous/mux"
	"github.com/containous/traefik/autogen"
	"github.com/elazarl/go-bindata-assetfs"
)

type DashboardHandler struct{}

func (g DashboardHandler) AddRoutes(router *mux.Router) {
	// Expose dashboard
	router.Methods("GET").Path("/").HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, "/dashboard/", 302)
	})
	router.Methods("GET").PathPrefix("/dashboard/").
		Handler(http.StripPrefix("/dashboard/", http.FileServer(&assetfs.AssetFS{Asset: autogen.Asset, AssetInfo: autogen.AssetInfo, AssetDir: autogen.AssetDir, Prefix: "static"})))

}
