package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v2/pkg/config/runtime"
)

type udpRouterRepresentation struct {
	*runtime.UDPRouterInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
}

func newUDPRouterRepresentation(name string, rt *runtime.UDPRouterInfo) udpRouterRepresentation {
	return udpRouterRepresentation{
		UDPRouterInfo: rt,
		Name:          name,
		Provider:      getProviderName(name),
	}
}

type udpServiceRepresentation struct {
	*runtime.UDPServiceInfo
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
	Type     string `json:"type,omitempty"`
}

func newUDPServiceRepresentation(name string, si *runtime.UDPServiceInfo) udpServiceRepresentation {
	return udpServiceRepresentation{
		UDPServiceInfo: si,
		Name:           name,
		Provider:       getProviderName(name),
		Type:           strings.ToLower(extractType(si.UDPService)),
	}
}

func (h Handler) getUDPRouters(rw http.ResponseWriter, request *http.Request) {
	results := make([]udpRouterRepresentation, 0, len(h.runtimeConfiguration.UDPRouters))

	criterion := newSearchCriterion(request.URL.Query())

	for name, rt := range h.runtimeConfiguration.UDPRouters {
		if keepUDPRouter(name, rt, criterion) {
			results = append(results, newUDPRouterRepresentation(name, rt))
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

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

func (h Handler) getUDPRouter(rw http.ResponseWriter, request *http.Request) {
	routerID := mux.Vars(request)["routerID"]

	rw.Header().Set("Content-Type", "application/json")

	router, ok := h.runtimeConfiguration.UDPRouters[routerID]
	if !ok {
		writeError(rw, fmt.Sprintf("router not found: %s", routerID), http.StatusNotFound)
		return
	}

	result := newUDPRouterRepresentation(routerID, router)

	err := json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) getUDPServices(rw http.ResponseWriter, request *http.Request) {
	results := make([]udpServiceRepresentation, 0, len(h.runtimeConfiguration.UDPServices))

	criterion := newSearchCriterion(request.URL.Query())

	for name, si := range h.runtimeConfiguration.UDPServices {
		if keepUDPService(name, si, criterion) {
			results = append(results, newUDPServiceRepresentation(name, si))
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

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

func (h Handler) getUDPService(rw http.ResponseWriter, request *http.Request) {
	serviceID := mux.Vars(request)["serviceID"]

	rw.Header().Set("Content-Type", "application/json")

	service, ok := h.runtimeConfiguration.UDPServices[serviceID]
	if !ok {
		writeError(rw, fmt.Sprintf("service not found: %s", serviceID), http.StatusNotFound)
		return
	}

	result := newUDPServiceRepresentation(serviceID, service)

	err := json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.Ctx(request.Context()).Error().Err(err).Send()
		writeError(rw, err.Error(), http.StatusInternalServerError)
	}
}

func keepUDPRouter(name string, item *runtime.UDPRouterInfo, criterion *searchCriterion) bool {
	if criterion == nil {
		return true
	}

	return criterion.withStatus(item.Status) && criterion.searchIn(name)
}

func keepUDPService(name string, item *runtime.UDPServiceInfo, criterion *searchCriterion) bool {
	if criterion == nil {
		return true
	}

	return criterion.withStatus(item.Status) && criterion.searchIn(name)
}
