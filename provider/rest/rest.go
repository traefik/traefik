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

// Provider is a provider.Provider implementation that provides the UI
type Provider struct {
	configurationChan     chan<- types.ConfigMessage
	EntryPoint            string `description:"Entrypoint"`
	CurrentConfigurations *safe.Safe
}

var (
	templatesRenderer = render.New(render.Options{
		Directory: "nowhere",
	})
)

func (provider *Provider) AddRoutes(systemRouter *mux.Router) {
	systemRouter.Methods("PUT").Path("/api/providers/{provider}").HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		//todo deprecated
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
			//todo change to rest when we can break
			provider.configurationChan <- types.ConfigMessage{ProviderName: "web", Configuration: configuration}
			provider.getConfigHandler(response, request)
		} else {
			log.Errorf("Error parsing configuration %+v", err)
			http.Error(response, fmt.Sprintf("%+v", err), http.StatusBadRequest)
		}
	})
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, _ types.Constraints) error {
	provider.configurationChan = configurationChan
	return nil
}

func (p *Provider) getConfigHandler(response http.ResponseWriter, request *http.Request) {
	currentConfigurations := p.CurrentConfigurations.Get().(types.Configurations)
	templatesRenderer.JSON(response, http.StatusOK, currentConfigurations)
}
