package rancher

import (
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/containous/traefik/pkg/config"
	"github.com/containous/traefik/pkg/job"
	"github.com/containous/traefik/pkg/log"
	"github.com/containous/traefik/pkg/provider"
	"github.com/containous/traefik/pkg/safe"
	rancher "github.com/rancher/go-rancher-metadata/metadata"
)

const (
	// DefaultTemplateRule The default template for the default rule.
	DefaultTemplateRule = "Host(`{{ normalize .Name }}`)"

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
	DefaultRule               string `description:"Default rule"`
	ExposedByDefault          bool   `description:"Expose containers by default" export:"true"`
	EnableServiceHealthFilter bool
	RefreshSeconds            int
	defaultRuleTpl            *template.Template
	IntervalPoll              bool   `description:"Poll the Rancher metadata service every 'rancher.refreshseconds' (less accurate)"`
	Prefix                    string `description:"Prefix used for accessing the Rancher metadata service"`
}

type rancherData struct {
	Name       string
	Labels     map[string]string
	Containers []string
	Health     string
	State      string
	Port       string
	ExtraConf  configuration
}

// Init the provider.
func (p *Provider) Init() error {
	defaultRuleTpl, err := provider.MakeDefaultRuleTemplate(p.DefaultRule, nil)
	if err != nil {
		return fmt.Errorf("error while parsing default rule: %v", err)
	}

	p.defaultRuleTpl = defaultRuleTpl
	return p.BaseProvider.Init()
}

func (p *Provider) createClient() (rancher.Client, error) {
	metadataServiceURL := fmt.Sprintf("http://rancher-metadata.rancher.internal/%s", p.Prefix)
	client, err := rancher.NewClientAndWait(metadataServiceURL)
	if err != nil {
		log.Errorf("Failed to create Rancher metadata service client: %v", err)
		return nil, err
	}

	return client, nil
}

// Provide allows the rancher provider to provide configurations to traefik using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- config.Message, pool *safe.Pool) error {
	ctx := log.With(context.Background(), log.Str(log.ProviderName, "rancher"))
	logger := log.FromContext(ctx)

	operation := func() error {

		client, err := p.createClient()

		if err != nil {
			log.Errorf("Failed to create the metadata client metadata service: %v", err)
			return err
		}

		updateConfiguration := func(version string) {
			stacks, err := client.GetStacks()
			if err != nil {
				log.Errorf("Failed to query Rancher metadata service: %v", err)
				return
			}

			rancherData := p.parseMetadataSourcedRancherData(ctx, stacks)

			fmt.Printf("Received rancherdata %+v", rancherData)

			configuration := p.buildConfiguration(ctx, rancherData)
			configurationChan <- config.Message{
				ProviderName:  "rancher",
				Configuration: configuration,
			}
		}
		updateConfiguration("init")

		if p.Watch {
			pool.Go(func(stop chan bool) {
				switch {
				case p.IntervalPoll:
					p.intervalPoll(client, updateConfiguration, stop)
				default:
					p.longPoll(client, updateConfiguration, stop)
				}
			})
		}

		return nil
	}

	notify := func(err error, time time.Duration) {
		logger.Errorf("Provider connection error %+v, retrying in %s", err, time)
	}
	err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
	if err != nil {
		logger.Errorf("Cannot connect to Provider server: %+v", err)
	}
	return nil
}

func (p *Provider) intervalPoll(client rancher.Client, updateConfiguration func(string), stop chan bool) {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := time.NewTicker(time.Second * time.Duration(p.RefreshSeconds))
	defer ticker.Stop()

	var version string
	for {
		select {
		case <-ticker.C:
			newVersion, err := client.GetVersion()
			if err != nil {
				log.Errorf("Failed to create Rancher metadata service client: %v", err)
			} else if version != newVersion {
				version = newVersion
				updateConfiguration(version)
			}
		case <-stop:
			return
		}
	}
}

func (p *Provider) longPoll(client rancher.Client, updateConfiguration func(string), stop chan bool) {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Holds the connection until there is either a change in the metadata
	// repository or `p.RefreshSeconds` has elapsed. Long polling should be
	// favored for the most accurate configuration updates.
	safe.Go(func() {
		client.OnChange(p.RefreshSeconds, updateConfiguration)
	})
	<-stop
}

func (p *Provider) parseMetadataSourcedRancherData(ctx context.Context, stacks []rancher.Stack) (rancherDataList []rancherData) {
	for _, stack := range stacks {
		for _, service := range stack.Services {

			servicePort := ""
			if len(service.Ports) > 0 {
				servicePort = service.Ports[0]
			}
			for _, port := range service.Ports {
				log.Debugf("Set Port %s", port)
			}

			var containerIPAddresses []string
			for _, container := range service.Containers {
				if containerFilter(container.Name, container.HealthState, container.State) {
					containerIPAddresses = append(containerIPAddresses, container.PrimaryIp)
				}
			}

			service := rancherData{
				Name:       service.Name + "/" + stack.Name,
				State:      service.State,
				Labels:     service.Labels,
				Port:       servicePort,
				Containers: containerIPAddresses,
			}

			extraConf, err := p.getConfiguration(service)
			if err != nil {
				log.FromContext(ctx).Errorf("Skip container %s: %v", service.Name, err)
				continue
			}

			service.ExtraConf = extraConf

			rancherDataList = append(rancherDataList, service)
		}
	}
	return rancherDataList
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
