package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/containous/mux"
	"github.com/containous/traefik/pkg/config/dynamic"
	"github.com/containous/traefik/pkg/config/static"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/types"
	"github.com/containous/traefik/pkg/version"
	assetfs "github.com/elazarl/go-bindata-assetfs"
)

const (
	defaultPerPage = 100
	defaultPage    = 1
)

const nextPageHeader = "X-Next-Page"

type serviceInfoRepresentation struct {
	*dynamic.ServiceInfo
	ServerStatus map[string]string `json:"serverStatus,omitempty"`
}

// RunTimeRepresentation is the configuration information exposed by the API handler.
type RunTimeRepresentation struct {
	Routers     map[string]*dynamic.RouterInfo        `json:"routers,omitempty"`
	Middlewares map[string]*dynamic.MiddlewareInfo    `json:"middlewares,omitempty"`
	Services    map[string]*serviceInfoRepresentation `json:"services,omitempty"`
	TCPRouters  map[string]*dynamic.TCPRouterInfo     `json:"tcpRouters,omitempty"`
	TCPServices map[string]*dynamic.TCPServiceInfo    `json:"tcpServices,omitempty"`
}

type pageInfo struct {
	startIndex int
	endIndex   int
	nextPage   int
}

// Handler serves the configuration and status of Traefik on API endpoints.
type Handler struct {
	dashboard bool
	debug     bool
	// runtimeConfiguration is the data set used to create all the data representations exposed by the API.
	runtimeConfiguration *dynamic.RuntimeConfiguration
	staticConfig         static.Configuration
	statistics           *types.Statistics
	// stats                *thoasstats.Stats // FIXME stats
	// StatsRecorder         *middlewares.StatsRecorder // FIXME stats
	dashboardAssets *assetfs.AssetFS
}

// New returns a Handler defined by staticConfig, and if provided, by runtimeConfig.
// It finishes populating the information provided in the runtimeConfig.
func New(staticConfig static.Configuration, runtimeConfig *dynamic.RuntimeConfiguration) *Handler {
	rConfig := runtimeConfig
	if rConfig == nil {
		rConfig = &dynamic.RuntimeConfiguration{}
	}

	return &Handler{
		dashboard:            staticConfig.API.Dashboard,
		statistics:           staticConfig.API.Statistics,
		dashboardAssets:      staticConfig.API.DashboardAssets,
		runtimeConfiguration: rConfig,
		staticConfig:         staticConfig,
		debug:                staticConfig.API.Debug,
	}
}

// Append add api routes on a router
func (h Handler) Append(router *mux.Router) {
	if h.debug {
		DebugHandler{}.Append(router)
	}

	router.Methods(http.MethodGet).Path("/api/rawdata").HandlerFunc(h.getRuntimeConfiguration)
	router.Methods(http.MethodGet).Path("/api/overview").HandlerFunc(h.getOverview)

	router.Methods(http.MethodGet).Path("/api/entrypoints").HandlerFunc(h.getEntryPoints)
	router.Methods(http.MethodGet).Path("/api/entrypoints/{entryPointID}").HandlerFunc(h.getEntryPoint)

	router.Methods(http.MethodGet).Path("/api/http/routers").HandlerFunc(h.getRouters)
	router.Methods(http.MethodGet).Path("/api/http/routers/{routerID}").HandlerFunc(h.getRouter)
	router.Methods(http.MethodGet).Path("/api/http/services").HandlerFunc(h.getServices)
	router.Methods(http.MethodGet).Path("/api/http/services/{serviceID}").HandlerFunc(h.getService)
	router.Methods(http.MethodGet).Path("/api/http/middlewares").HandlerFunc(h.getMiddlewares)
	router.Methods(http.MethodGet).Path("/api/http/middlewares/{middlewareID}").HandlerFunc(h.getMiddleware)

	router.Methods(http.MethodGet).Path("/api/tcp/routers").HandlerFunc(h.getTCPRouters)
	router.Methods(http.MethodGet).Path("/api/tcp/routers/{routerID}").HandlerFunc(h.getTCPRouter)
	router.Methods(http.MethodGet).Path("/api/tcp/services").HandlerFunc(h.getTCPServices)
	router.Methods(http.MethodGet).Path("/api/tcp/services/{serviceID}").HandlerFunc(h.getTCPService)

	// FIXME stats
	// health route
	// router.Methods(http.MethodGet).Path("/health").HandlerFunc(p.getHealthHandler)

	version.Handler{}.Append(router)

	if h.dashboard {
		DashboardHandler{Assets: h.dashboardAssets}.Append(router)
	}
}

func (h Handler) getRuntimeConfiguration(rw http.ResponseWriter, request *http.Request) {
	siRepr := make(map[string]*serviceInfoRepresentation, len(h.runtimeConfiguration.Services))
	for k, v := range h.runtimeConfiguration.Services {
		siRepr[k] = &serviceInfoRepresentation{
			ServiceInfo:  v,
			ServerStatus: v.GetAllStatus(),
		}
	}

	result := RunTimeRepresentation{
		Routers:     h.runtimeConfiguration.Routers,
		Middlewares: h.runtimeConfiguration.Middlewares,
		Services:    siRepr,
		TCPRouters:  h.runtimeConfiguration.TCPRouters,
		TCPServices: h.runtimeConfiguration.TCPServices,
	}

	rw.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func pagination(request *http.Request, max int) (pageInfo, error) {
	perPage, err := getIntParam(request, "per_page", defaultPerPage)
	if err != nil {
		return pageInfo{}, err
	}

	page, err := getIntParam(request, "page", defaultPage)
	if err != nil {
		return pageInfo{}, err
	}

	startIndex := (page - 1) * perPage
	if startIndex != 0 && startIndex >= max {
		return pageInfo{}, fmt.Errorf("invalid request: page: %d, per_page: %d", page, perPage)
	}

	endIndex := startIndex + perPage
	if endIndex >= max {
		endIndex = max
	}

	nextPage := 1
	if page*perPage < max {
		nextPage = page + 1
	}

	return pageInfo{startIndex: startIndex, endIndex: endIndex, nextPage: nextPage}, nil
}

func getIntParam(request *http.Request, key string, defaultValue int) (int, error) {
	raw := request.URL.Query().Get(key)
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid request: %s: %d", key, value)
	}
	return value, nil
}

func getProviderName(id string) string {
	return strings.SplitN(id, "@", 2)[1]
}
