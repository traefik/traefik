package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/containous/mux"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/unrolled/render"
)

// Provider is a provider.Provider implementation that provides a Rest API
type Provider struct {
	configurationChan chan<- types.ConfigMessage
	EntryPoint        string `description:"EntryPoint" export:"true"`
}

var templatesRenderer = render.New(render.Options{Directory: "nowhere"})

// Init the provider
func (p *Provider) Init(_ types.Constraints) error {
	return nil
}

// AddRoutes add rest provider routes on a router
func (p *Provider) AddRoutes(systemRouter *mux.Router) {
	systemRouter.
		Methods(http.MethodPut).
		Path("/api/providers/{provider}").
		HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

			vars := mux.Vars(request)
			// TODO: Deprecated configuration - Need to be removed in the future
			if vars["provider"] != "web" && vars["provider"] != "rest" {
				response.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(response, "Only 'rest' provider can be updated through the REST API")
				return
			} else if vars["provider"] == "web" {
				log.Warn("The provider web is deprecated. Please use /rest instead")
			}

			configuration := new(types.Configuration)
			body, _ := ioutil.ReadAll(request.Body)
			err := json.Unmarshal(body, configuration)
			if err == nil {
				// TODO: Deprecated configuration - Change to `rest` in the future
				p.configurationChan <- types.ConfigMessage{ProviderName: "web", Configuration: configuration}
				err := templatesRenderer.JSON(response, http.StatusOK, configuration)
				if err != nil {
					log.Error(err)
				}
			} else {
				log.Errorf("Error parsing configuration %+v", err)
				http.Error(response, fmt.Sprintf("%+v", err), http.StatusBadRequest)
			}
		})
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
	p.configurationChan = configurationChan
	return nil
}
