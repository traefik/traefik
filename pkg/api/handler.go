package api

import (
	"io"
	"net/http"

	"github.com/containous/mux"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/config/static"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/types"
	"github.com/containous/traefik/pkg/version"
	assetfs "github.com/elazarl/go-bindata-assetfs"
	thoasstats "github.com/thoas/stats"
	"github.com/unrolled/render"
)

// Handler serves the configuration and status of Traefik on API endpoints.
type Handler struct {
	EntryPoint string
	Dashboard  bool
	Debug      bool
	// runtimeConfiguration is the data set used to create all the data representations exposed by the API.
	runtimeConfiguration *config.RuntimeConfiguration
	Statistics           *types.Statistics
	Stats                *thoasstats.Stats
	// StatsRecorder         *middlewares.StatsRecorder // FIXME stats
	DashboardAssets *assetfs.AssetFS
}

// New returns a Handler defined by staticConfig, and if provided, by runtimeConfig.
// It finishes populating the information provided in the runtimeConfig.
func New(staticConfig static.Configuration, runtimeConfig *config.RuntimeConfiguration) *Handler {
	rConfig := runtimeConfig
	if rConfig == nil {
		rConfig = &config.RuntimeConfiguration{}
	}
	h := Handler{
		EntryPoint:           staticConfig.API.EntryPoint,
		Dashboard:            staticConfig.API.Dashboard,
		Statistics:           staticConfig.API.Statistics,
		DashboardAssets:      staticConfig.API.DashboardAssets,
		runtimeConfiguration: rConfig,
		Debug:                staticConfig.Global.Debug,
	}
	return &h
}

var templateRenderer jsonRenderer = render.New(render.Options{Directory: "nowhere"})

type jsonRenderer interface {
	JSON(w io.Writer, status int, v interface{}) error
}

// Append add api routes on a router
func (h Handler) Append(router *mux.Router) {
	if h.Debug {
		DebugHandler{}.Append(router)
	}

	router.Methods(http.MethodGet).Path("/api/rawdata").HandlerFunc(h.getRuntimeConfiguration)

	// FIXME stats
	// health route
	//router.Methods(http.MethodGet).Path("/health").HandlerFunc(p.getHealthHandler)

	version.Handler{}.Append(router)

	if h.Dashboard {
		DashboardHandler{Assets: h.DashboardAssets}.Append(router)
	}
}

type serviceInfoRepresentation struct {
	*config.ServiceInfo `json:",omitempty"`
	ServerStatus        map[string]string `json:"serverStatus,omitempty"`
}

// RunTimeRepresentation is the configuration information exposed by the API handler.
type RunTimeRepresentation struct {
	Routers     map[string]*config.RouterInfo         `json:"routers,omitempty"`
	Middlewares map[string]*config.MiddlewareInfo     `json:"middlewares,omitempty"`
	Services    map[string]*serviceInfoRepresentation `json:"services,omitempty"`
	TCPRouters  map[string]*config.TCPRouterInfo      `json:"tcpRouters,omitempty"`
	TCPServices map[string]*config.TCPServiceInfo     `json:"tcpServices,omitempty"`
}

func (h Handler) getRuntimeConfiguration(rw http.ResponseWriter, request *http.Request) {
	siRepr := make(map[string]*serviceInfoRepresentation, len(h.runtimeConfiguration.Services))
	for k, v := range h.runtimeConfiguration.Services {
		siRepr[k] = &serviceInfoRepresentation{
			ServiceInfo:  v,
			ServerStatus: v.GetAllStatus(),
		}
	}
	rtRepr := RunTimeRepresentation{
		Routers:     h.runtimeConfiguration.Routers,
		Middlewares: h.runtimeConfiguration.Middlewares,
		Services:    siRepr,
		TCPRouters:  h.runtimeConfiguration.TCPRouters,
		TCPServices: h.runtimeConfiguration.TCPServices,
	}
	err := templateRenderer.JSON(rw, http.StatusOK, rtRepr)
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}
