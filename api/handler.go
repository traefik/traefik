package api

import (
	"net/http"

	"github.com/containous/mux"
	"github.com/containous/traefik/middlewares"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/containous/traefik/version"
	"github.com/unrolled/render"

	thoas_stats "github.com/thoas/stats"
)

type Handler struct {
	EntryPoint            string `description:"Entrypoint"`
	Dashboard             bool   `description:"Activate dashboard"`
	Debug                 bool
	CurrentConfigurations *safe.Safe
	Statistics            *types.Statistics `description:"Enable more detailed statistics" export:"true"`
	Stats                 *thoas_stats.Stats
	StatsRecorder         *middlewares.StatsRecorder
}

var (
	templatesRenderer = render.New(render.Options{
		Directory: "nowhere",
	})
)

func (p Handler) AddRoutes(router *mux.Router) {
	if p.Debug {
		DebugHandler{}.AddRoutes(router)
	}

	router.Methods("GET").Path("/api").HandlerFunc(p.getConfigHandler)
	router.Methods("GET").Path("/api/providers").HandlerFunc(p.getConfigHandler)
	router.Methods("GET").Path("/api/providers/{provider}").HandlerFunc(p.getProviderHandler)
	router.Methods("GET").Path("/api/providers/{provider}/backends").HandlerFunc(p.getBackendsHandler)
	router.Methods("GET").Path("/api/providers/{provider}/backends/{backend}").HandlerFunc(p.getBackendHandler)
	router.Methods("GET").Path("/api/providers/{provider}/backends/{backend}/servers").HandlerFunc(p.getServersHandler)
	router.Methods("GET").Path("/api/providers/{provider}/backends/{backend}/servers/{server}").HandlerFunc(p.getServerHandler)
	router.Methods("GET").Path("/api/providers/{provider}/frontends").HandlerFunc(p.getFrontendsHandler)
	router.Methods("GET").Path("/api/providers/{provider}/frontends/{frontend}").HandlerFunc(p.getFrontendHandler)
	router.Methods("GET").Path("/api/providers/{provider}/frontends/{frontend}/routes").HandlerFunc(p.getRoutesHandler)
	router.Methods("GET").Path("/api/providers/{provider}/frontends/{frontend}/routes/{route}").HandlerFunc(p.getRouteHandler)

	// health route
	router.Methods("GET").Path("/health").HandlerFunc(p.getHealthHandler)

	version.VersionHandler{}.AddRoutes(router)

	if p.Dashboard {
		DashboardHandler{}.AddRoutes(router)
	}
}

func getProviderIDFromVars(vars map[string]string) string {
	providerID := vars["provider"]
	//todo deprecated
	if providerID == "rest" {
		providerID = "web"
	}
	return providerID
}

func (p Handler) getConfigHandler(response http.ResponseWriter, request *http.Request) {
	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	templatesRenderer.JSON(response, http.StatusOK, currentConfigurations)
}

func (p Handler) getProviderHandler(response http.ResponseWriter, request *http.Request) {
	providerID := getProviderIDFromVars(mux.Vars(request))

	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		templatesRenderer.JSON(response, http.StatusOK, provider)
	} else {
		http.NotFound(response, request)
	}
}

func (p Handler) getBackendsHandler(response http.ResponseWriter, request *http.Request) {
	providerID := getProviderIDFromVars(mux.Vars(request))

	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	if provider, ok := currentConfigurations[providerID]; ok {
		templatesRenderer.JSON(response, http.StatusOK, provider.Backends)
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
			templatesRenderer.JSON(response, http.StatusOK, backend)
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
			templatesRenderer.JSON(response, http.StatusOK, backend.Servers)
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
				templatesRenderer.JSON(response, http.StatusOK, server)
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
		templatesRenderer.JSON(response, http.StatusOK, provider.Frontends)
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
			templatesRenderer.JSON(response, http.StatusOK, frontend)
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
			templatesRenderer.JSON(response, http.StatusOK, frontend.Routes)
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
				templatesRenderer.JSON(response, http.StatusOK, route)
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
	templatesRenderer.JSON(response, http.StatusOK, health)
}
