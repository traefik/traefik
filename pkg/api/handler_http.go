package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/runtime"
	"github.com/traefik/traefik/v3/pkg/tls"
)

type routerRepresentation struct {
	*runtime.RouterInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
}

func newRouterRepresentation(name string, rt *runtime.RouterInfo) routerRepresentation {
	if rt.TLS != nil && rt.TLS.Options == "" {
		rt.TLS.Options = tls.DefaultTLSConfigName
	}

	return routerRepresentation{
		RouterInfo: rt,
		Name:       name,
		Provider:   getProviderName(name),
	}
}

type serviceRepresentation struct {
	*runtime.ServiceInfo
	ServerStatus map[string]string `json:"serverStatus,omitempty"`
	Name         string            `json:"name,omitempty"`
	Provider     string            `json:"provider,omitempty"`
	Type         string            `json:"type,omitempty"`
}

func newServiceRepresentation(name string, si *runtime.ServiceInfo) serviceRepresentation {
	return serviceRepresentation{
		ServiceInfo:  si,
		Name:         name,
		Provider:     getProviderName(name),
		ServerStatus: si.GetAllStatus(),
		Type:         strings.ToLower(extractType(si.Service)),
	}
}

type middlewareRepresentation struct {
	*runtime.MiddlewareInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
	Type     string `json:"type,omitempty"`
}

func newMiddlewareRepresentation(name string, mi *runtime.MiddlewareInfo) middlewareRepresentation {
	return middlewareRepresentation{
		MiddlewareInfo: mi,
		Name:           name,
		Provider:       getProviderName(name),
		Type:           strings.ToLower(extractType(mi.Middleware)),
	}
}

func (h Handler) getRouters(rw http.ResponseWriter, request *http.Request) {
	results := make([]routerRepresentation, 0, len(h.runtimeConfiguration.Routers))

	query := request.URL.Query()
	criterion := newSearchCriterion(query)

	for name, rt := range h.runtimeConfiguration.Routers {
		if keepRouter(name, rt, criterion) {
			results = append(results, newRouterRepresentation(name, rt))
		}
	}

	sortRouters(query, results)

	rw.Header().Set("Content-Type", "application/json")

	pageInfo, err := pagination(request, len(results))
	if err != nil {
		writeError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set(nextPageHeader, strconv.Itoa(pageInfo.nextPage))

	err = json.NewEncoder(rw).Encode(results[pageInfo.startIndex:pageInfo.endIndex])
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getRouter(rw http.ResponseWriter, request *http.Request) {
	scapedRouterID := mux.Vars(request)["routerID"]

	routerID, err := url.PathUnescape(scapedRouterID)
	if err != nil {
		writeError(rw, fmt.Sprintf("unable to decode routerID %q: %s", scapedRouterID, err), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	router, ok := h.runtimeConfiguration.Routers[routerID]
	if !ok {
		writeError(rw, fmt.Sprintf("router not found: %s", routerID), http.StatusNotFound)
		return
	}

	result := newRouterRepresentation(routerID, router)

	err = json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getServices(rw http.ResponseWriter, request *http.Request) {
	results := make([]serviceRepresentation, 0, len(h.runtimeConfiguration.Services))

	query := request.URL.Query()
	criterion := newSearchCriterion(query)

	for name, si := range h.runtimeConfiguration.Services {
		if keepService(name, si, criterion) {
			results = append(results, newServiceRepresentation(name, si))
		}
	}

	sortServices(query, results)

	rw.Header().Set("Content-Type", "application/json")

	pageInfo, err := pagination(request, len(results))
	if err != nil {
		writeError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set(nextPageHeader, strconv.Itoa(pageInfo.nextPage))

	err = json.NewEncoder(rw).Encode(results[pageInfo.startIndex:pageInfo.endIndex])
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getService(rw http.ResponseWriter, request *http.Request) {
	scapedServiceID := mux.Vars(request)["serviceID"]

	serviceID, err := url.PathUnescape(scapedServiceID)
	if err != nil {
		writeError(rw, fmt.Sprintf("unable to decode serviceID %q: %s", scapedServiceID, err), http.StatusBadRequest)
		return
	}

	rw.Header().Add("Content-Type", "application/json")

	service, ok := h.runtimeConfiguration.Services[serviceID]
	if !ok {
		writeError(rw, fmt.Sprintf("service not found: %s", serviceID), http.StatusNotFound)
		return
	}

	result := newServiceRepresentation(serviceID, service)

	err = json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getMiddlewares(rw http.ResponseWriter, request *http.Request) {
	results := make([]middlewareRepresentation, 0, len(h.runtimeConfiguration.Middlewares))

	query := request.URL.Query()
	criterion := newSearchCriterion(query)

	for name, mi := range h.runtimeConfiguration.Middlewares {
		if keepMiddleware(name, mi, criterion) {
			results = append(results, newMiddlewareRepresentation(name, mi))
		}
	}

	sortMiddlewares(query, results)

	rw.Header().Set("Content-Type", "application/json")

	pageInfo, err := pagination(request, len(results))
	if err != nil {
		writeError(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rw.Header().Set(nextPageHeader, strconv.Itoa(pageInfo.nextPage))

	err = json.NewEncoder(rw).Encode(results[pageInfo.startIndex:pageInfo.endIndex])
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getMiddleware(rw http.ResponseWriter, request *http.Request) {
	scapedMiddlewareID := mux.Vars(request)["middlewareID"]

	middlewareID, err := url.PathUnescape(scapedMiddlewareID)
	if err != nil {
		writeError(rw, fmt.Sprintf("unable to decode middlewareID %q: %s", scapedMiddlewareID, err), http.StatusBadRequest)
		return
	}

	rw.Header().Set("Content-Type", "application/json")

	middleware, ok := h.runtimeConfiguration.Middlewares[middlewareID]
	if !ok {
		writeError(rw, fmt.Sprintf("middleware not found: %s", middlewareID), http.StatusNotFound)
		return
	}

	result := newMiddlewareRepresentation(middlewareID, middleware)

	err = json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func keepRouter(name string, item *runtime.RouterInfo, criterion *searchCriterion) bool {
	if criterion == nil {
		return true
	}

	return criterion.withStatus(item.Status) &&
		criterion.searchIn(item.Rule, name) &&
		criterion.filterService(item.Service) &&
		criterion.filterMiddleware(item.Middlewares)
}

func keepService(name string, item *runtime.ServiceInfo, criterion *searchCriterion) bool {
	if criterion == nil {
		return true
	}

	return criterion.withStatus(item.Status) && criterion.searchIn(name)
}

func keepMiddleware(name string, item *runtime.MiddlewareInfo, criterion *searchCriterion) bool {
	if criterion == nil {
		return true
	}

	return criterion.withStatus(item.Status) && criterion.searchIn(name)
}
