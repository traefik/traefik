package rancher

import (
	"fmt"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

const (
	// Health
	healthy         = "healthy"
	updatingHealthy = "updating-healthy"

	// State
	active          = "active"
	running         = "running"
	upgraded        = "upgraded"
	upgrading       = "upgrading"
	updatingActive  = "updating-active"
	updatingRunning = "updating-running"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	provider.BaseProvider     `mapstructure:",squash" export:"true"`
	APIConfiguration          `mapstructure:",squash" export:"true"` // Provide backwards compatibility
	API                       *APIConfiguration                      `description:"Enable the Rancher API provider" export:"true"`
	Metadata                  *MetadataConfiguration                 `description:"Enable the Rancher metadata service provider" export:"true"`
	Domain                    string                                 `description:"Default domain used"`
	RefreshSeconds            int                                    `description:"Polling interval (in seconds)" export:"true"`
	ExposedByDefault          bool                                   `description:"Expose services by default" export:"true"`
	EnableServiceHealthFilter bool                                   `description:"Filter services with unhealthy states and inactive states" export:"true"`
}

type rancherData struct {
	Name          string
	Labels        map[string]string // List of labels set to container or service
	Containers    []string
	Health        string
	State         string
	SegmentLabels map[string]string
	SegmentName   string
}

func (r rancherData) String() string {
	return fmt.Sprintf("{name:%s, labels:%v, containers: %v, health: %s, state: %s}", r.Name, r.Labels, r.Containers, r.Health, r.State)
}

// Init the provider
func (p *Provider) Init(constraints types.Constraints) error {
	return p.BaseProvider.Init(constraints)
}

// Provide allows either the Rancher API or metadata service provider to
// seed configuration into Traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
	if p.Metadata == nil {
		return p.apiProvide(configurationChan, pool)
	}
	return p.metadataProvide(configurationChan, pool)
}

func containerFilter(name, healthState, state string) bool {
	if healthState != "" && healthState != healthy && healthState != updatingHealthy {
		log.Debugf("Filtering container %s with healthState of %s", name, healthState)
		return false
	}

	if state != "" && state != running && state != updatingRunning && state != upgraded {
		log.Debugf("Filtering container %s with state of %s", name, state)
		return false
	}

	return true
}
