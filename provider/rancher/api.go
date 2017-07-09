package rancher

import (
	"context"
	"os"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	rancher "github.com/rancher/go-rancher/client"
)

const labelRancheStackServiceName = "io.rancher.stack_service.name"

var withoutPagination *rancher.ListOpts

// APIConfiguration contains configuration properties specific to the Rancher
// API provider.
type APIConfiguration struct {
	Endpoint  string `description:"Rancher server API HTTP(S) endpoint"`
	AccessKey string `description:"Rancher server API access key"`
	SecretKey string `description:"Rancher server API secret key"`
}

func init() {
	withoutPagination = &rancher.ListOpts{
		Filters: map[string]interface{}{"limit": 0},
	}
}

func (p *Provider) createClient() (*rancher.RancherClient, error) {
	rancherURL := getenv("CATTLE_URL", p.API.Endpoint)
	accessKey := getenv("CATTLE_ACCESS_KEY", p.API.AccessKey)
	secretKey := getenv("CATTLE_SECRET_KEY", p.API.SecretKey)

	return rancher.NewRancherClient(&rancher.ClientOpts{
		Url:       rancherURL,
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func (p *Provider) apiProvide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	p.Constraints = append(p.Constraints, constraints...)

	if p.API == nil {
		p.API = &APIConfiguration{}
	}

	safe.Go(func() {
		operation := func() error {
			rancherClient, err := p.createClient()

			if err != nil {
				log.Errorf("Failed to create a client for rancher, error: %s", err)
				return err
			}

			ctx := context.Background()
			var environments = listRancherEnvironments(rancherClient)
			var services = listRancherServices(rancherClient)
			var container = listRancherContainer(rancherClient)

			var rancherData = parseAPISourcedRancherData(environments, services, container)

			configuration := p.loadRancherConfig(rancherData)
			configurationChan <- types.ConfigMessage{
				ProviderName:  "rancher",
				Configuration: configuration,
			}

			if p.Watch {
				_, cancel := context.WithCancel(ctx)
				ticker := time.NewTicker(time.Second * time.Duration(p.RefreshSeconds))
				pool.Go(func(stop chan bool) {
					for {
						select {
						case <-ticker.C:

							log.Debugf("Refreshing new Data from Provider API")
							var environments = listRancherEnvironments(rancherClient)
							var services = listRancherServices(rancherClient)
							var container = listRancherContainer(rancherClient)

							rancherData := parseAPISourcedRancherData(environments, services, container)

							configuration := p.loadRancherConfig(rancherData)
							if configuration != nil {
								configurationChan <- types.ConfigMessage{
									ProviderName:  "rancher",
									Configuration: configuration,
								}
							}
						case <-stop:
							ticker.Stop()
							cancel()
							return
						}
					}
				})
			}

			return nil
		}
		notify := func(err error, time time.Duration) {
			log.Errorf("Provider connection error %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(operation, job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to Provider Endpoint %+v", err)
		}
	})

	return nil
}

func listRancherEnvironments(client *rancher.RancherClient) []*rancher.Project {

	// Rancher Environment in frontend UI is actually project in API
	// https://forums.rancher.com/t/api-key-for-all-environments/279/9

	var environmentList = []*rancher.Project{}

	environments, err := client.Project.List(nil)

	if err != nil {
		log.Errorf("Cannot get Rancher Environments %+v", err)
	}

	for k := range environments.Data {
		environmentList = append(environmentList, &environments.Data[k])
	}

	return environmentList
}

func listRancherServices(client *rancher.RancherClient) []*rancher.Service {

	var servicesList = []*rancher.Service{}

	services, err := client.Service.List(withoutPagination)

	if err != nil {
		log.Errorf("Cannot get Provider Services %+v", err)
	}

	for k := range services.Data {
		servicesList = append(servicesList, &services.Data[k])
	}

	return servicesList
}

func listRancherContainer(client *rancher.RancherClient) []*rancher.Container {

	containerList := []*rancher.Container{}

	container, err := client.Container.List(withoutPagination)

	if err != nil {
		log.Errorf("Cannot get Provider Services %+v", err)
	}

	valid := true

	for valid {
		for k := range container.Data {
			containerList = append(containerList, &container.Data[k])
		}

		container, err = container.Next()

		if err != nil {
			break
		}

		if container == nil || len(container.Data) == 0 {
			valid = false
		}
	}

	return containerList
}

func parseAPISourcedRancherData(environments []*rancher.Project, services []*rancher.Service, containers []*rancher.Container) []rancherData {
	var rancherDataList []rancherData

	for _, environment := range environments {

		for _, service := range services {
			if service.EnvironmentId != environment.Id {
				continue
			}

			rancherData := rancherData{
				Name:       environment.Name + "/" + service.Name,
				Health:     service.HealthState,
				State:      service.State,
				Labels:     make(map[string]string),
				Containers: []string{},
			}

			if service.LaunchConfig == nil || service.LaunchConfig.Labels == nil {
				log.Warnf("Rancher Service Labels are missing. Environment: %s, service: %s", environment.Name, service.Name)
			} else {
				for key, value := range service.LaunchConfig.Labels {
					rancherData.Labels[key] = value.(string)
				}
			}

			for _, container := range containers {
				if container.Labels[labelRancheStackServiceName] == rancherData.Name &&
					containerFilter(container.Name, container.HealthState, container.State) {
					rancherData.Containers = append(rancherData.Containers, container.PrimaryIpAddress)
				}
			}
			rancherDataList = append(rancherDataList, rancherData)
		}
	}

	return rancherDataList
}
