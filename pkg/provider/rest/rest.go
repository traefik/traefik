package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/log"
	"github.com/containous/traefik/v2/pkg/provider"
	"github.com/containous/traefik/v2/pkg/safe"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

var _ provider.Provider = (*Provider)(nil)

// Provider is a provider.Provider implementation that provides a Rest API.
type Provider struct {
	Insecure          bool `description:"Activate REST Provider on an insecure entryPoints named traefik." json:"insecure,omitempty" toml:"insecure,omitempty" yaml:"insecure,omitempty" export:"true"`
	configurationChan chan<- dynamic.Message
}

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
}

var templatesRenderer = render.New(render.Options{Directory: "nowhere"})

// Init the provider.
func (p *Provider) Init() error {
	return nil
}

// Handler creates an http.Handler for the Rest API
func (p *Provider) Handler() http.Handler {
	router := mux.NewRouter()
	p.Append(router)
	return router
}

// Append add rest provider routes on a router.
func (p *Provider) Append(systemRouter *mux.Router) {
	systemRouter.
		Methods(http.MethodPut).
		Path("/api/providers/{provider}").
		HandlerFunc(func(response http.ResponseWriter, request *http.Request) {

			vars := mux.Vars(request)
			if vars["provider"] != "rest" {
				response.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(response, "Only 'rest' provider can be updated through the REST API")
				return
			}

			configuration := new(dynamic.Configuration)
			body, _ := ioutil.ReadAll(request.Body)

			if err := json.Unmarshal(body, configuration); err != nil {
				log.WithoutContext().Errorf("Error parsing configuration %+v", err)
				http.Error(response, fmt.Sprintf("%+v", err), http.StatusBadRequest)
				return
			}

			p.configurationChan <- dynamic.Message{ProviderName: "rest", Configuration: configuration}
			if err := templatesRenderer.JSON(response, http.StatusOK, configuration); err != nil {
				log.WithoutContext().Error(err)
			}
		})
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	p.configurationChan = configurationChan
	return nil
}
