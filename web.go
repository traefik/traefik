package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/emilevauge/traefik/autogen"
	"github.com/emilevauge/traefik/types"
	"github.com/gorilla/mux"
	"github.com/thoas/stats"
	"github.com/unrolled/render"
)

var metrics = stats.New()

// WebProvider is a provider.Provider implementation that provides the UI.
// FIXME to be handled another way.
type WebProvider struct {
	Address           string
	CertFile, KeyFile string
	ReadOnly          bool
	server            *Server
}

var (
	templatesRenderer = render.New(render.Options{
		Directory: "nowhere",
	})
)

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *WebProvider) Provide(configurationChan chan<- types.ConfigMessage) error {
	systemRouter := mux.NewRouter()

	// health route
	systemRouter.Methods("GET").Path("/health").HandlerFunc(provider.getHealthHandler)

	// API routes
	systemRouter.Methods("GET").Path("/api").HandlerFunc(provider.getConfigHandler)
	systemRouter.Methods("GET").Path("/api/providers").HandlerFunc(provider.getConfigHandler)
	systemRouter.Methods("GET").Path("/api/providers/{provider}").HandlerFunc(provider.getProviderHandler)
	systemRouter.Methods("PUT").Path("/api/providers/{provider}").HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if provider.ReadOnly {
			response.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(response, "REST API is in read-only mode")
			return
		}
		vars := mux.Vars(request)
		if vars["provider"] != "web" {
			response.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(response, "Only 'web' provider can be updated through the REST API")
			return
		}

		configuration := new(types.Configuration)
		body, _ := ioutil.ReadAll(request.Body)
		err := json.Unmarshal(body, configuration)
		if err == nil {
			configurationChan <- types.ConfigMessage{"web", configuration}
			provider.getConfigHandler(response, request)
		} else {
			log.Errorf("Error parsing configuration %+v", err)
			http.Error(response, fmt.Sprintf("%+v", err), http.StatusBadRequest)
		}
	})
	systemRouter.Methods("GET").Path("/api/providers/{provider}/backends").HandlerFunc(provider.getBackendsHandler)
	systemRouter.Methods("GET").Path("/api/providers/{provider}/backends/{backend}").HandlerFunc(provider.getBackendHandler)
	systemRouter.Methods("GET").Path("/api/providers/{provider}/backends/{backend}/servers").HandlerFunc(provider.getServersHandler)
	systemRouter.Methods("GET").Path("/api/providers/{provider}/backends/{backend}/servers/{server}").HandlerFunc(provider.getServerHandler)
	systemRouter.Methods("GET").Path("/api/providers/{provider}/frontends").HandlerFunc(provider.getFrontendsHandler)
	systemRouter.Methods("GET").Path("/api/providers/{provider}/frontends/{frontend}").HandlerFunc(provider.getFrontendHandler)
	systemRouter.Methods("GET").Path("/api/providers/{provider}/frontends/{frontend}/routes").HandlerFunc(provider.getRoutesHandler)
	systemRouter.Methods("GET").Path("/api/providers/{provider}/frontends/{frontend}/routes/{route}").HandlerFunc(provider.getRouteHandler)

	// Expose dashboard
	systemRouter.Methods("GET").Path("/").HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, "/dashboard/", 302)
	})
	systemRouter.Methods("GET").PathPrefix("/dashboard/").Handler(http.StripPrefix("/dashboard/", http.FileServer(&assetfs.AssetFS{Asset: autogen.Asset, AssetDir: autogen.AssetDir, Prefix: "static"})))

	go func() {
		if len(provider.CertFile) > 0 && len(provider.KeyFile) > 0 {
			err := http.ListenAndServeTLS(provider.Address, provider.CertFile, provider.KeyFile, systemRouter)
			if err != nil {
				log.Fatal("Error creating server: ", err)
			}
		} else {
			err := http.ListenAndServe(provider.Address, systemRouter)
			if err != nil {
				log.Fatal("Error creating server: ", err)
			}
		}
	}()
	return nil
}

func (provider *WebProvider) getHealthHandler(response http.ResponseWriter, request *http.Request) {
	templatesRenderer.JSON(response, http.StatusOK, metrics.Data())
}

func (provider *WebProvider) getConfigHandler(response http.ResponseWriter, request *http.Request) {
	templatesRenderer.JSON(response, http.StatusOK, provider.server.currentConfigurations)
}

func (provider *WebProvider) getProviderHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	if provider, ok := provider.server.currentConfigurations[providerID]; ok {
		templatesRenderer.JSON(response, http.StatusOK, provider)
	} else {
		http.NotFound(response, request)
	}
}

func (provider *WebProvider) getBackendsHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	if provider, ok := provider.server.currentConfigurations[providerID]; ok {
		templatesRenderer.JSON(response, http.StatusOK, provider.Backends)
	} else {
		http.NotFound(response, request)
	}
}

func (provider *WebProvider) getBackendHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	backendID := vars["backend"]
	if provider, ok := provider.server.currentConfigurations[providerID]; ok {
		if backend, ok := provider.Backends[backendID]; ok {
			templatesRenderer.JSON(response, http.StatusOK, backend)
			return
		}
	}
	http.NotFound(response, request)
}

func (provider *WebProvider) getServersHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	backendID := vars["backend"]
	if provider, ok := provider.server.currentConfigurations[providerID]; ok {
		if backend, ok := provider.Backends[backendID]; ok {
			templatesRenderer.JSON(response, http.StatusOK, backend.Servers)
			return
		}
	}
	http.NotFound(response, request)
}

func (provider *WebProvider) getServerHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	backendID := vars["backend"]
	serverID := vars["server"]
	if provider, ok := provider.server.currentConfigurations[providerID]; ok {
		if backend, ok := provider.Backends[backendID]; ok {
			if server, ok := backend.Servers[serverID]; ok {
				templatesRenderer.JSON(response, http.StatusOK, server)
				return
			}
		}
	}
	http.NotFound(response, request)
}

func (provider *WebProvider) getFrontendsHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	if provider, ok := provider.server.currentConfigurations[providerID]; ok {
		templatesRenderer.JSON(response, http.StatusOK, provider.Frontends)
	} else {
		http.NotFound(response, request)
	}
}

func (provider *WebProvider) getFrontendHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	frontendID := vars["frontend"]
	if provider, ok := provider.server.currentConfigurations[providerID]; ok {
		if frontend, ok := provider.Frontends[frontendID]; ok {
			templatesRenderer.JSON(response, http.StatusOK, frontend)
			return
		}
	}
	http.NotFound(response, request)
}

func (provider *WebProvider) getRoutesHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	frontendID := vars["frontend"]
	if provider, ok := provider.server.currentConfigurations[providerID]; ok {
		if frontend, ok := provider.Frontends[frontendID]; ok {
			templatesRenderer.JSON(response, http.StatusOK, frontend.Routes)
			return
		}
	}
	http.NotFound(response, request)
}

func (provider *WebProvider) getRouteHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := vars["provider"]
	frontendID := vars["frontend"]
	routeID := vars["route"]
	if provider, ok := provider.server.currentConfigurations[providerID]; ok {
		if frontend, ok := provider.Frontends[frontendID]; ok {
			if route, ok := frontend.Routes[routeID]; ok {
				templatesRenderer.JSON(response, http.StatusOK, route)
				return
			}
		}
	}
	http.NotFound(response, request)
}
