package api

import (
	"net/http"

	"github.com/containous/mux"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/containous/traefik/version"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	thoas_stats "github.com/thoas/stats"
	"github.com/unrolled/render"
)

// Handler expose api routes
type Handler struct {
	EntryPoint            string `description:"EntryPoint" export:"true"`
	Dashboard             bool   `description:"Activate dashboard" export:"true"`
	Debug                 bool   `export:"true"`
	CurrentConfigurations *safe.Safe
	Statistics            *types.Statistics          `description:"Enable more detailed statistics" export:"true"`
	Stats                 *thoas_stats.Stats         `json:"-"`
	StatsRecorder         *middlewares.StatsRecorder `json:"-"`
	DashboardAssets       *assetfs.AssetFS           `json:"-"`
}

var (
	templatesRenderer = render.New(render.Options{
		Directory: "nowhere",
	})
)

// AddRoutes add api routes on a router
func (p Handler) AddRoutes(router *mux.Router) {
	if p.Debug {
		DebugHandler{}.AddRoutes(router)
	}

	router.Methods(http.MethodGet).Path("/api").HandlerFunc(p.getConfigHandler)
	router.Methods(http.MethodGet).Path("/api/providers").HandlerFunc(p.getConfigHandler)
	router.Methods(http.MethodGet).Path("/api/providers/{provider}").HandlerFunc(p.getProviderHandler)
	router.Methods(http.MethodGet).Path("/api/providers/{provider}/backends").HandlerFunc(p.getBackendsHandler)
	router.Methods(http.MethodGet).Path("/api/providers/{provider}/backends/{backend}").HandlerFunc(p.getBackendHandler)
	router.Methods(http.MethodGet).Path("/api/providers/{provider}/backends/{backend}/servers").HandlerFunc(p.getServersHandler)
	router.Methods(http.MethodGet).Path("/api/providers/{provider}/backends/{backend}/servers/{server}").HandlerFunc(p.getServerHandler)
	router.Methods(http.MethodGet).Path("/api/providers/{provider}/frontends").HandlerFunc(p.getFrontendsHandler)
	router.Methods(http.MethodGet).Path("/api/providers/{provider}/frontends/{frontend}").HandlerFunc(p.getFrontendHandler)
	router.Methods(http.MethodGet).Path("/api/providers/{provider}/frontends/{frontend}/routes").HandlerFunc(p.getRoutesHandler)
	router.Methods(http.MethodGet).Path("/api/providers/{provider}/frontends/{frontend}/routes/{route}").HandlerFunc(p.getRouteHandler)

	// health route
	router.Methods(http.MethodGet).Path("/health").HandlerFunc(p.getHealthHandler)

	version.Handler{}.AddRoutes(router)

	if p.Dashboard {
		DashboardHandler{Assets: p.DashboardAssets}.AddRoutes(router)
	}
}

func getProviderIDFromVars(vars map[string]string) string {
	providerID := vars["provider"]
	// TODO: Deprecated
	if providerID == "rest" {
		providerID = "web"
	}
	return providerID
}

func (p Handler) getConfigHandler(response http.ResponseWriter, request *http.Request) {
	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	err := templatesRenderer.JSON(response, http.StatusOK, currentConfigurations)
	if err != nil {
		log.Error(err)
	}
}

func (p Handler) getProviderHandler(response http.ResponseWriter, request *http.Request) {
	providerID := getProviderIDFromVars(mux.Vars(request))

	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		err := templatesRenderer.JSON(response, http.StatusOK, provider)
		if err != nil {
			log.Error(err)
		}
	} else {
		http.NotFound(response, request)
	}
}

func (p Handler) getBackendsHandler(response http.ResponseWriter, request *http.Request) {
	providerID := getProviderIDFromVars(mux.Vars(request))

	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		err := templatesRenderer.JSON(response, http.StatusOK, provider.Backends)
		if err != nil {
			log.Error(err)
		}
	} else {
		http.NotFound(response, request)
	}
}

func (p Handler) getBackendHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := getProviderIDFromVars(vars)
	backendID := vars["backend"]

	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		if backend, ok := provider.Backends[backendID]; ok {
			err := templatesRenderer.JSON(response, http.StatusOK, backend)
			if err != nil {
				log.Error(err)
			}
			return
		}
	}
	http.NotFound(response, request)
}

func (p Handler) getServersHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := getProviderIDFromVars(vars)
	backendID := vars["backend"]

	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		if backend, ok := provider.Backends[backendID]; ok {
			err := templatesRenderer.JSON(response, http.StatusOK, backend.Servers)
			if err != nil {
				log.Error(err)
			}
			return
		}
	}
	http.NotFound(response, request)
}

func (p Handler) getServerHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := getProviderIDFromVars(vars)
	backendID := vars["backend"]
	serverID := vars["server"]

	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		if backend, ok := provider.Backends[backendID]; ok {
			if server, ok := backend.Servers[serverID]; ok {
				err := templatesRenderer.JSON(response, http.StatusOK, server)
				if err != nil {
					log.Error(err)
				}
				return
			}
		}
	}
	http.NotFound(response, request)
}

func (p Handler) getFrontendsHandler(response http.ResponseWriter, request *http.Request) {
	providerID := getProviderIDFromVars(mux.Vars(request))

	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		err := templatesRenderer.JSON(response, http.StatusOK, provider.Frontends)
		if err != nil {
			log.Error(err)
		}
	} else {
		http.NotFound(response, request)
	}
}

func (p Handler) getFrontendHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := getProviderIDFromVars(vars)
	frontendID := vars["frontend"]

	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		if frontend, ok := provider.Frontends[frontendID]; ok {
			err := templatesRenderer.JSON(response, http.StatusOK, frontend)
			if err != nil {
				log.Error(err)
			}
			return
		}
	}
	http.NotFound(response, request)
}

func (p Handler) getRoutesHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := getProviderIDFromVars(vars)
	frontendID := vars["frontend"]

	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		if frontend, ok := provider.Frontends[frontendID]; ok {
			err := templatesRenderer.JSON(response, http.StatusOK, frontend.Routes)
			if err != nil {
				log.Error(err)
			}
			return
		}
	}
	http.NotFound(response, request)
}

func (p Handler) getRouteHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	providerID := getProviderIDFromVars(vars)
	frontendID := vars["frontend"]
	routeID := vars["route"]

	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		if frontend, ok := provider.Frontends[frontendID]; ok {
			if route, ok := frontend.Routes[routeID]; ok {
				err := templatesRenderer.JSON(response, http.StatusOK, route)
				if err != nil {
					log.Error(err)
				}
				return
			}
		}
	}
	http.NotFound(response, request)
}

// healthResponse combines data returned by thoas/stats with statistics (if
// they are enabled).
type healthResponse struct {
	*thoas_stats.Data
	*middlewares.Stats
}

func (p *Handler) getHealthHandler(response http.ResponseWriter, request *http.Request) {
	health := &healthResponse{Data: p.Stats.Data()}
	if p.StatsRecorder != nil {
		health.Stats = p.StatsRecorder.Data()
	}
	err := templatesRenderer.JSON(response, http.StatusOK, health)
	if err != nil {
		log.Error(err)
	}
}
