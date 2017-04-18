package eureka

import (
	"io/ioutil"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/ArthurHlt/go-eureka-client/eureka"
	log "github.com/Sirupsen/logrus"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

// Provider holds configuration of the Provider provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash"`
	Endpoint              string
	Delay                 string
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
		go func() {
			for t := range ticker.C {

				log.Debug("Refreshing Provider " + t.String())

				configuration, err := p.buildConfiguration()
				if err != nil {
					log.Errorf("Failed to refresh Provider configuration, error: %s", err)
					return
				}

				configurationChan <- types.ConfigMessage{
					ProviderName:  "eureka",
					Configuration: configuration,
				}
			}
		}()
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

// Build the configuration from Provider server
func (p *Provider) buildConfiguration() (*types.Configuration, error) {
	var EurekaFuncMap = template.FuncMap{
		"getPort":       p.getPort,
		"getProtocol":   p.getProtocol,
		"getWeight":     p.getWeight,
		"getInstanceID": p.getInstanceID,
	}

	eureka.GetLogger().SetOutput(ioutil.Discard)

	client := eureka.NewClient([]string{
		p.Endpoint,
	})

	applications, err := client.GetApplications()
	if err != nil {
		return nil, err
	}

	templateObjects := struct {
		Applications []eureka.Application
	}{
		applications.Applications,
	}

	configuration, err := p.GetConfiguration("templates/eureka.tmpl", EurekaFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration, nil
}

func (p *Provider) getPort(instance eureka.InstanceInfo) string {
	if instance.SecurePort.Enabled {
		return strconv.Itoa(instance.SecurePort.Port)
	}
	return strconv.Itoa(instance.Port.Port)
}

func (p *Provider) getProtocol(instance eureka.InstanceInfo) string {
	if instance.SecurePort.Enabled {
		return "https"
	}
	return "http"
}

func (p *Provider) getWeight(instance eureka.InstanceInfo) string {
	if val, ok := instance.Metadata.Map["traefik.weight"]; ok {
		return val
	}
	return "0"
}

func (p *Provider) getInstanceID(instance eureka.InstanceInfo) string {
	if val, ok := instance.Metadata.Map["traefik.backend.id"]; ok {
		return val
	}
	return strings.Replace(instance.IpAddr, ".", "-", -1) + "-" + p.getPort(instance)
}
