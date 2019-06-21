package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/containous/mux"
	"github.com/containous/traefik/pkg/config"
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
	*config.ServiceInfo
	ServerStatus map[string]string `json:"serverStatus,omitempty"`
}

// RunTimeRepresentation is the configuration information exposed by the API handler.
type RunTimeRepresentation struct {
	Routers     map[string]*config.RouterInfo         `json:"routers,omitempty"`
	Middlewares map[string]*config.MiddlewareInfo     `json:"middlewares,omitempty"`
	Services    map[string]*serviceInfoRepresentation `json:"services,omitempty"`
	TCPRouters  map[string]*config.TCPRouterInfo      `json:"tcpRouters,omitempty"`
	TCPServices map[string]*config.TCPServiceInfo     `json:"tcpServices,omitempty"`
}

type routerRepresentation struct {
	*config.RouterInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
}

type serviceRepresentation struct {
	*config.ServiceInfo
	ServerStatus map[string]string `json:"serverStatus,omitempty"`
	Name         string            `json:"name,omitempty"`
	Provider     string            `json:"provider,omitempty"`
}

type middlewareRepresentation struct {
	*config.MiddlewareInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
}

type tcpRouterRepresentation struct {
	*config.TCPRouterInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
}

type tcpServiceRepresentation struct {
	*config.TCPServiceInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
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
	runtimeConfiguration *config.RuntimeConfiguration
	statistics           *types.Statistics
	// stats                *thoasstats.Stats // FIXME stats
	// StatsRecorder         *middlewares.StatsRecorder // FIXME stats
	dashboardAssets *assetfs.AssetFS
}

// New returns a Handler defined by staticConfig, and if provided, by runtimeConfig.
// It finishes populating the information provided in the runtimeConfig.
func New(staticConfig static.Configuration, runtimeConfig *config.RuntimeConfiguration) *Handler {
	rConfig := runtimeConfig
	if rConfig == nil {
		rConfig = &config.RuntimeConfiguration{}
	}

	return &Handler{
		dashboard:            staticConfig.API.Dashboard,
		statistics:           staticConfig.API.Statistics,
		dashboardAssets:      staticConfig.API.DashboardAssets,
		runtimeConfiguration: rConfig,
		debug:                staticConfig.API.Debug,
	}
}

// Append add api routes on a router
func (h Handler) Append(router *mux.Router) {
	if h.debug {
		DebugHandler{}.Append(router)
	}

	router.Methods(http.MethodGet).Path("/api/rawdata").HandlerFunc(h.getRuntimeConfiguration)

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

func (h Handler) getRouters(rw http.ResponseWriter, request *http.Request) {
	results := make([]routerRepresentation, 0, len(h.runtimeConfiguration.Routers))

	for name, rt := range h.runtimeConfiguration.Routers {
		results = append(results, routerRepresentation{
			RouterInfo: rt,
			Name:       name,
			Provider:   getProviderName(name),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	pageInfo, err := pagination(request, len(results))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set(nextPageHeader, strconv.Itoa(pageInfo.nextPage))

	err = json.NewEncoder(rw).Encode(results[pageInfo.startIndex:pageInfo.endIndex])
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getRouter(rw http.ResponseWriter, request *http.Request) {
	routerID := mux.Vars(request)["routerID"]

	router, ok := h.runtimeConfiguration.Routers[routerID]
	if !ok {
		http.NotFound(rw, request)
		return
	}

	result := routerRepresentation{
		RouterInfo: router,
		Name:       routerID,
		Provider:   getProviderName(routerID),
	}

	err := json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getServices(rw http.ResponseWriter, request *http.Request) {
	results := make([]serviceRepresentation, 0, len(h.runtimeConfiguration.Services))

	for name, si := range h.runtimeConfiguration.Services {
		results = append(results, serviceRepresentation{
			ServiceInfo:  si,
			Name:         name,
			Provider:     getProviderName(name),
			ServerStatus: si.GetAllStatus(),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	pageInfo, err := pagination(request, len(results))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set(nextPageHeader, strconv.Itoa(pageInfo.nextPage))

	err = json.NewEncoder(rw).Encode(results[pageInfo.startIndex:pageInfo.endIndex])
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getService(rw http.ResponseWriter, request *http.Request) {
	serviceID := mux.Vars(request)["serviceID"]

	service, ok := h.runtimeConfiguration.Services[serviceID]
	if !ok {
		http.NotFound(rw, request)
		return
	}

	result := serviceRepresentation{
		ServiceInfo:  service,
		Name:         serviceID,
		Provider:     getProviderName(serviceID),
		ServerStatus: service.GetAllStatus(),
	}

	err := json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getMiddlewares(rw http.ResponseWriter, request *http.Request) {
	results := make([]middlewareRepresentation, 0, len(h.runtimeConfiguration.Middlewares))

	for name, mi := range h.runtimeConfiguration.Middlewares {
		results = append(results, middlewareRepresentation{
			MiddlewareInfo: mi,
			Name:           name,
			Provider:       getProviderName(name),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	pageInfo, err := pagination(request, len(results))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set(nextPageHeader, strconv.Itoa(pageInfo.nextPage))

	err = json.NewEncoder(rw).Encode(results[pageInfo.startIndex:pageInfo.endIndex])
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getMiddleware(rw http.ResponseWriter, request *http.Request) {
	middlewareID := mux.Vars(request)["middlewareID"]

	middleware, ok := h.runtimeConfiguration.Middlewares[middlewareID]
	if !ok {
		http.NotFound(rw, request)
		return
	}

	result := middlewareRepresentation{
		MiddlewareInfo: middleware,
		Name:           middlewareID,
		Provider:       getProviderName(middlewareID),
	}

	err := json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getTCPRouters(rw http.ResponseWriter, request *http.Request) {
	results := make([]tcpRouterRepresentation, 0, len(h.runtimeConfiguration.TCPRouters))

	for name, rt := range h.runtimeConfiguration.TCPRouters {
		results = append(results, tcpRouterRepresentation{
			TCPRouterInfo: rt,
			Name:          name,
			Provider:      getProviderName(name),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	pageInfo, err := pagination(request, len(results))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set(nextPageHeader, strconv.Itoa(pageInfo.nextPage))

	err = json.NewEncoder(rw).Encode(results[pageInfo.startIndex:pageInfo.endIndex])
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getTCPRouter(rw http.ResponseWriter, request *http.Request) {
	routerID := mux.Vars(request)["routerID"]

	router, ok := h.runtimeConfiguration.TCPRouters[routerID]
	if !ok {
		http.NotFound(rw, request)
		return
	}

	result := tcpRouterRepresentation{
		TCPRouterInfo: router,
		Name:          routerID,
		Provider:      getProviderName(routerID),
	}

	err := json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getTCPServices(rw http.ResponseWriter, request *http.Request) {
	results := make([]tcpServiceRepresentation, 0, len(h.runtimeConfiguration.TCPServices))

	for name, si := range h.runtimeConfiguration.TCPServices {
		results = append(results, tcpServiceRepresentation{
			TCPServiceInfo: si,
			Name:           name,
			Provider:       getProviderName(name),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	pageInfo, err := pagination(request, len(results))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set(nextPageHeader, strconv.Itoa(pageInfo.nextPage))

	err = json.NewEncoder(rw).Encode(results[pageInfo.startIndex:pageInfo.endIndex])
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getTCPService(rw http.ResponseWriter, request *http.Request) {
	serviceID := mux.Vars(request)["serviceID"]

	service, ok := h.runtimeConfiguration.TCPServices[serviceID]
	if !ok {
		http.NotFound(rw, request)
		return
	}

	result := tcpServiceRepresentation{
		TCPServiceInfo: service,
		Name:           serviceID,
		Provider:       getProviderName(serviceID),
	}

	err := json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
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
