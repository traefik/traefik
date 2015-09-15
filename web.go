package main

import (
	"encoding/json"
	"fmt"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

type WebProvider struct {
	Address string
}

type Page struct {
	Configuration Configuration
}

func (provider *WebProvider) Provide(configurationChan chan<- *Configuration) {
	systemRouter := mux.NewRouter()
	systemRouter.Methods("GET").Path("/").Handler(http.HandlerFunc(GetHtmlConfigHandler))
	systemRouter.Methods("GET").Path("/metrics").Handler(http.HandlerFunc(GetStatsHandler))
	systemRouter.Methods("GET").Path("/api").Handler(http.HandlerFunc(GetConfigHandler))
	systemRouter.Methods("POST").Path("/api").Handler(http.HandlerFunc(
		func(rw http.ResponseWriter, r *http.Request) {
			configuration := new(Configuration)
			b, _ := ioutil.ReadAll(r.Body)
			err := json.Unmarshal(b, configuration)
			if err == nil {
				configurationChan <- configuration
				GetConfigHandler(rw, r)
			} else {
				log.Error("Error parsing configuration %+v\n", err)
				http.Error(rw, fmt.Sprintf("%+v", err), http.StatusBadRequest)
			}
		}))
	systemRouter.Methods("GET").Path("/api/backends").Handler(http.HandlerFunc(GetBackendsHandler))
	systemRouter.Methods("GET").Path("/api/backends/{id}").Handler(http.HandlerFunc(GetBackendHandler))
	systemRouter.Methods("GET").Path("/api/frontends").Handler(http.HandlerFunc(GetFrontendsHandler))
	systemRouter.Methods("GET").Path("/api/frontends/{id}").Handler(http.HandlerFunc(GetFrontendHandler))
	systemRouter.Methods("GET").PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "static"})))

	go http.ListenAndServe(provider.Address, systemRouter)
}

func GetConfigHandler(rw http.ResponseWriter, r *http.Request) {
	templatesRenderer.JSON(rw, http.StatusOK, currentConfiguration)
}

func GetHtmlConfigHandler(response http.ResponseWriter, request *http.Request) {
	templatesRenderer.HTML(response, http.StatusOK, "configuration", Page{Configuration: *currentConfiguration})
}

func GetStatsHandler(rw http.ResponseWriter, r *http.Request) {
	templatesRenderer.JSON(rw, http.StatusOK, metrics.Data())
}

func GetBackendsHandler(rw http.ResponseWriter, r *http.Request) {
	templatesRenderer.JSON(rw, http.StatusOK, currentConfiguration.Backends)
}

func GetBackendHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	templatesRenderer.JSON(rw, http.StatusOK, currentConfiguration.Backends[id])
}

func GetFrontendsHandler(rw http.ResponseWriter, r *http.Request) {
	templatesRenderer.JSON(rw, http.StatusOK, currentConfiguration.Frontends)
}

func GetFrontendHandler(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	templatesRenderer.JSON(rw, http.StatusOK, currentConfiguration.Frontends[id])
}
