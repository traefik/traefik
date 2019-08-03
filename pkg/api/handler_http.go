package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/containous/traefik/v2/pkg/config/runtime"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/gorilla/mux"
)

type routerRepresentation struct {
	*runtime.RouterInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
}

type serviceRepresentation struct {
	*runtime.ServiceInfo
	ServerStatus map[string]string `json:"serverStatus,omitempty"`
	Name         string            `json:"name,omitempty"`
	Provider     string            `json:"provider,omitempty"`
}

type middlewareRepresentation struct {
	*runtime.MiddlewareInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
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

	rw.Header().Set("Content-Type", "application/json")
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

	rw.Header().Set("Content-Type", "application/json")

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

	rw.Header().Set("Content-Type", "application/json")
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

	rw.Header().Add("Content-Type", "application/json")

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

	rw.Header().Set("Content-Type", "application/json")
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

	rw.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}
