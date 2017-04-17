package provider

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
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

// Eureka holds configuration of the Eureka provider.
type Eureka struct {
	BaseProvider `mapstructure:",squash"`
	Endpoint     string
	Delay        string
}

// Provide allows the provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Eureka) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, _ types.Constraints) error {

	operation := func() error {
		configuration, err := provider.buildConfiguration()
		if err != nil {
			log.Errorf("Failed to build configuration for Eureka, error: %s", err)
			return err
		}

		configurationChan <- types.ConfigMessage{
			ProviderName:  "eureka",
			Configuration: configuration,
		}

		var delay time.Duration
		if len(provider.Delay) > 0 {
			var err error
			delay, err = time.ParseDuration(provider.Delay)
			if err != nil {
				log.Errorf("Failed to parse delay for Eureka, error: %s", err)
				return err
			}
		} else {
			delay = time.Second * 30
		}

		ticker := time.NewTicker(delay)
		go func() {
			for t := range ticker.C {

				log.Debug("Refreshing Eureka " + t.String())

				configuration, err := provider.buildConfiguration()
				if err != nil {
					log.Errorf("Failed to refresh Eureka configuration, error: %s", err)
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
		log.Errorf("Eureka connection error %+v, retrying in %s", err, time)
	}
	err := backoff.RetryNotify(operation, job.NewBackOff(backoff.NewExponentialBackOff()), notify)
	if err != nil {
		log.Errorf("Cannot connect to Eureka server %+v", err)
		return err
	}
	return nil
}

// Build the configuration from Eureka server
func (provider *Eureka) buildConfiguration() (*types.Configuration, error) {
	var EurekaFuncMap = template.FuncMap{
		"getPort":       provider.getPort,
		"getProtocol":   provider.getProtocol,
		"getWeight":     provider.getWeight,
		"getInstanceID": provider.getInstanceID,
	}

	eureka.GetLogger().SetOutput(ioutil.Discard)

	client := eureka.NewClient([]string{
		provider.Endpoint,
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

	configuration, err := provider.GetConfiguration("templates/eureka.tmpl", EurekaFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}
	return configuration, nil
}

func (provider *Eureka) getPort(instance eureka.InstanceInfo) string {
	if instance.SecurePort.Enabled {
		return strconv.Itoa(instance.SecurePort.Port)
	}
	return strconv.Itoa(instance.Port.Port)
}

func (provider *Eureka) getProtocol(instance eureka.InstanceInfo) string {
	if instance.SecurePort.Enabled {
		return "https"
	}
	return "http"
}

func (provider *Eureka) getWeight(instance eureka.InstanceInfo) string {
	if val, ok := instance.Metadata.Map["traefik.weight"]; ok {
		return val
	}
	return "0"
}

func (provider *Eureka) getInstanceID(instance eureka.InstanceInfo) string {
	if val, ok := instance.Metadata.Map["traefik.backend.id"]; ok {
		return val
	}
	return strings.Replace(instance.IpAddr, ".", "-", -1) + "-" + provider.getPort(instance)
}
