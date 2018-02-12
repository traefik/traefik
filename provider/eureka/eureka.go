package eureka

import (
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

// Provider holds configuration of the Provider provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash" export:"true"`
	Endpoint              string `description:"Eureka server endpoint"`
	Delay                 string `description:"Override default configuration time between refresh" export:"true"`
}

// Provide allows the eureka provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, _ types.Constraints) error {

	operation := func() error {
		configuration, err := p.buildConfiguration()
		if err != nil {
			log.Errorf("Failed to build configuration for Provider, error: %s", err)
			return err
		}

		configurationChan <- types.ConfigMessage{
			ProviderName:  "eureka",
			Configuration: configuration,
		}

		var delay time.Duration
		if len(p.Delay) > 0 {
			var err error
			delay, err = time.ParseDuration(p.Delay)
			if err != nil {
				log.Errorf("Failed to parse delay for Provider, error: %s", err)
				return err
			}
		} else {
			delay = time.Second * 30
		}

		ticker := time.NewTicker(delay)
		safe.Go(func() {
			for t := range ticker.C {

				log.Debugf("Refreshing Provider %s", t.String())

				configuration, err := p.buildConfiguration()
				if err != nil {
					log.Errorf("Failed to refresh Provider configuration, error: %s", err)
					continue
				}

				configurationChan <- types.ConfigMessage{
					ProviderName:  "eureka",
					Configuration: configuration,
				}
			}
		})
		return nil
	}

	notify := func(err error, time time.Duration) {
		log.Errorf("Provider connection error %+v, retrying in %s", err, time)
	}
	err := backoff.RetryNotify(operation, job.NewBackOff(backoff.NewExponentialBackOff()), notify)
	if err != nil {
		log.Errorf("Cannot connect to Provider server %+v", err)
		return err
	}
	return nil
}
