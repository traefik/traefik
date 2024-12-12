package api

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/config/static"
	"github.com/traefik/traefik/v3/pkg/version"
)

type apiError struct {
	Message string `json:"message"`
}

func writeError(rw http.ResponseWriter, msg string, code int) {
	data, err := json.Marshal(apiError{Message: msg})
	if err != nil {
		http.Error(rw, msg, code)
		return
	}

	http.Error(rw, string(data), code)
}

type serviceInfoRepresentation struct {
	*runtime.ServiceInfo
	ServerStatus map[string]string `json:"serverStatus,omitempty"`
}

// RunTimeRepresentation is the configuration information exposed by the API handler.
type RunTimeRepresentation struct {
	Routers        map[string]*runtime.RouterInfo        `json:"routers,omitempty"`
	Middlewares    map[string]*runtime.MiddlewareInfo    `json:"middlewares,omitempty"`
	Services       map[string]*serviceInfoRepresentation `json:"services,omitempty"`
	TCPRouters     map[string]*runtime.TCPRouterInfo     `json:"tcpRouters,omitempty"`
	TCPMiddlewares map[string]*runtime.TCPMiddlewareInfo `json:"tcpMiddlewares,omitempty"`
	TCPServices    map[string]*runtime.TCPServiceInfo    `json:"tcpServices,omitempty"`
	UDPRouters     map[string]*runtime.UDPRouterInfo     `json:"udpRouters,omitempty"`
	UDPServices    map[string]*runtime.UDPServiceInfo    `json:"udpServices,omitempty"`
}

// Handler serves the configuration and status of Traefik on API endpoints.
type Handler struct {
	staticConfig static.Configuration

	// runtimeConfiguration is the data set used to create all the data representations exposed by the API.
	runtimeConfiguration *runtime.Configuration
}

// NewBuilder returns a http.Handler builder based on runtime.Configuration.
func NewBuilder(staticConfig static.Configuration) func(*runtime.Configuration) http.Handler {
	return func(configuration *runtime.Configuration) http.Handler {
		return New(staticConfig, configuration).createRouter()
	}
}

// New returns a Handler defined by staticConfig, and if provided, by runtimeConfig.
// It finishes populating the information provided in the runtimeConfig.
func New(staticConfig static.Configuration, runtimeConfig *runtime.Configuration) *Handler {
	rConfig := runtimeConfig
	if rConfig == nil {
		rConfig = &runtime.Configuration{}
	}

	return &Handler{
		runtimeConfiguration: rConfig,
		staticConfig:         staticConfig,
	}
}

// createRouter creates API routes and router.
func (h Handler) createRouter() *mux.Router {
	router := mux.NewRouter().UseEncodedPath()

	apiRouter := router.PathPrefix(h.staticConfig.API.BasePath).Subrouter().UseEncodedPath()

	if h.staticConfig.API.Debug {
		DebugHandler{}.Append(apiRouter)
	}

	apiRouter.Methods(http.MethodGet).Path("/api/rawdata").HandlerFunc(h.getRuntimeConfiguration)

	// Experimental endpoint
	apiRouter.Methods(http.MethodGet).Path("/api/overview").HandlerFunc(h.getOverview)

	apiRouter.Methods(http.MethodGet).Path("/api/support-dump").HandlerFunc(h.getSupportDump)

	apiRouter.Methods(http.MethodGet).Path("/api/entrypoints").HandlerFunc(h.getEntryPoints)
	apiRouter.Methods(http.MethodGet).Path("/api/entrypoints/{entryPointID}").HandlerFunc(h.getEntryPoint)

	apiRouter.Methods(http.MethodGet).Path("/api/http/routers").HandlerFunc(h.getRouters)
	apiRouter.Methods(http.MethodGet).Path("/api/http/routers/{routerID}").HandlerFunc(h.getRouter)
	apiRouter.Methods(http.MethodGet).Path("/api/http/services").HandlerFunc(h.getServices)
	apiRouter.Methods(http.MethodGet).Path("/api/http/services/{serviceID}").HandlerFunc(h.getService)
	apiRouter.Methods(http.MethodGet).Path("/api/http/middlewares").HandlerFunc(h.getMiddlewares)
	apiRouter.Methods(http.MethodGet).Path("/api/http/middlewares/{middlewareID}").HandlerFunc(h.getMiddleware)

	apiRouter.Methods(http.MethodGet).Path("/api/tcp/routers").HandlerFunc(h.getTCPRouters)
	apiRouter.Methods(http.MethodGet).Path("/api/tcp/routers/{routerID}").HandlerFunc(h.getTCPRouter)
	apiRouter.Methods(http.MethodGet).Path("/api/tcp/services").HandlerFunc(h.getTCPServices)
	apiRouter.Methods(http.MethodGet).Path("/api/tcp/services/{serviceID}").HandlerFunc(h.getTCPService)
	apiRouter.Methods(http.MethodGet).Path("/api/tcp/middlewares").HandlerFunc(h.getTCPMiddlewares)
	apiRouter.Methods(http.MethodGet).Path("/api/tcp/middlewares/{middlewareID}").HandlerFunc(h.getTCPMiddleware)

	apiRouter.Methods(http.MethodGet).Path("/api/udp/routers").HandlerFunc(h.getUDPRouters)
	apiRouter.Methods(http.MethodGet).Path("/api/udp/routers/{routerID}").HandlerFunc(h.getUDPRouter)
	apiRouter.Methods(http.MethodGet).Path("/api/udp/services").HandlerFunc(h.getUDPServices)
	apiRouter.Methods(http.MethodGet).Path("/api/udp/services/{serviceID}").HandlerFunc(h.getUDPService)

	version.Handler{}.Append(apiRouter)

	return router
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
		Routers:        h.runtimeConfiguration.Routers,
		Middlewares:    h.runtimeConfiguration.Middlewares,
		Services:       siRepr,
		TCPRouters:     h.runtimeConfiguration.TCPRouters,
		TCPMiddlewares: h.runtimeConfiguration.TCPMiddlewares,
		TCPServices:    h.runtimeConfiguration.TCPServices,
		UDPRouters:     h.runtimeConfiguration.UDPRouters,
		UDPServices:    h.runtimeConfiguration.UDPServices,
	}

	rw.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func getProviderName(id string) string {
	return strings.SplitN(id, "@", 2)[1]
}

func extractType(element interface{}) string {
	v := reflect.ValueOf(element).Elem()
	for i := range v.NumField() {
		field := v.Field(i)

		if field.Kind() == reflect.Map && field.Type().Elem() == reflect.TypeOf(dynamic.PluginConf{}) {
			if keys := field.MapKeys(); len(keys) == 1 {
				return keys[0].String()
			}
		}

		if field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct {
			if !field.IsNil() {
				return v.Type().Field(i).Name
			}
		}
	}
	return ""
}
