package metrics

import (
	"github.com/thoas/stats"
	"github.com/unrolled/render"
	"net/http"
)

// Metrics holds Traefik aggregated stats
var Metrics = stats.New()

// GetHealthHandler expose Metrics data on /health
func GetHealthHandler(templatesRenderer *render.Render) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		templatesRenderer.JSON(response, http.StatusOK, Metrics.Data())
	}
}
