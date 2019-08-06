package api

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/containous/mux"
	"github.com/containous/traefik/pkg/config/static"
	"github.com/containous/traefik/pkg/log"
)

type entryPointRepresentation struct {
	*static.EntryPoint
	Name string `json:"name,omitempty"`
}

func (h Handler) getEntryPoints(rw http.ResponseWriter, request *http.Request) {
	results := make([]entryPointRepresentation, 0, len(h.staticConfig.EntryPoints))

	for name, ep := range h.staticConfig.EntryPoints {
		results = append(results, entryPointRepresentation{
			EntryPoint: ep,
			Name:       name,
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

func (h Handler) getEntryPoint(rw http.ResponseWriter, request *http.Request) {
	entryPointID := mux.Vars(request)["entryPointID"]

	ep, ok := h.staticConfig.EntryPoints[entryPointID]
	if !ok {
		http.NotFound(rw, request)
		return
	}

	result := entryPointRepresentation{
		EntryPoint: ep,
		Name:       entryPointID,
	}

	rw.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(rw).Encode(result)
	if err != nil {
		log.FromContext(request.Context()).Error(err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}
