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
	"github.com/mitchellh/mapstructure"
	rancher "github.com/rancher/go-rancher/v2"
)

const (
	labelRancherStackServiceName = "io.rancher.stack_service.name"
	hostNetwork                  = "host"
)

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

func (p *Provider) apiProvide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {

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
			stacks, err := listRancherStacks(rancherClient)
			if err != nil {
				return err
			}
			services, err := listRancherServices(rancherClient)
			if err != nil {
				return err
			}
			container, err := listRancherContainer(rancherClient)
			if err != nil {
				return err
			}

			var rancherData = parseAPISourcedRancherData(stacks, services, container)

			configuration := p.buildConfiguration(rancherData)
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
							checkAPI, errAPI := rancherClient.ApiKey.List(withoutPagination)

							if errAPI != nil {
								log.Errorf("Cannot establish connection: %+v, Rancher API return: %+v; Skipping refresh Data from Rancher API.", errAPI, checkAPI)
								continue
							}
							log.Debugf("Refreshing new Data from Rancher API")
							stacks, err = listRancherStacks(rancherClient)
							if err != nil {
								continue
							}
							services, err = listRancherServices(rancherClient)
							if err != nil {
								continue
							}
							container, err = listRancherContainer(rancherClient)
							if err != nil {
								continue
							}

							rancherData := parseAPISourcedRancherData(stacks, services, container)

							configuration := p.buildConfiguration(rancherData)
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

func listRancherStacks(client *rancher.RancherClient) ([]*rancher.Stack, error) {

	var stackList []*rancher.Stack

	stacks, err := client.Stack.List(withoutPagination)

	if err != nil {
		log.Errorf("Cannot get Provider Stacks %+v", err)
	}

	for k := range stacks.Data {
		stackList = append(stackList, &stacks.Data[k])
	}

	return stackList, err
}

func listRancherServices(client *rancher.RancherClient) ([]*rancher.Service, error) {

	var servicesList []*rancher.Service

	services, err := client.Service.List(withoutPagination)

	if err != nil {
		log.Errorf("Cannot get Provider Services %+v", err)
	}

	for k := range services.Data {
		servicesList = append(servicesList, &services.Data[k])
	}

	return servicesList, err
}

func listRancherContainer(client *rancher.RancherClient) ([]*rancher.Container, error) {

	var containerList []*rancher.Container

	container, err := client.Container.List(withoutPagination)

	if err != nil {
		log.Errorf("Cannot get Provider Services %+v", err)
		return containerList, err
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

	return containerList, err
}

func parseAPISourcedRancherData(stacks []*rancher.Stack, services []*rancher.Service, containers []*rancher.Container) []rancherData {
	var rancherDataList []rancherData

	for _, stack := range stacks {

		for _, service := range services {

			if service.StackId != stack.Id {
				continue
			}

			rData := rancherData{
				Name:       service.Name + "/" + stack.Name,
				Health:     service.HealthState,
				State:      service.State,
				Labels:     make(map[string]string),
				Containers: []string{},
			}

			if service.LaunchConfig == nil || service.LaunchConfig.Labels == nil {
				log.Warnf("Rancher Service Labels are missing. Stack: %s, service: %s", stack.Name, service.Name)
			} else {
				for key, value := range service.LaunchConfig.Labels {
					rData.Labels[key] = value.(string)
				}
			}

			for _, container := range containers {
				if container.Labels[labelRancherStackServiceName] == stack.Name+"/"+service.Name &&
					containerFilter(container.Name, container.HealthState, container.State) {

					if container.NetworkMode == hostNetwork {
						var endpoints []*rancher.PublicEndpoint
						err := mapstructure.Decode(service.PublicEndpoints, &endpoints)

						if err != nil {
							log.Errorf("Failed to decode PublicEndpoint: %v", err)
							continue
						}

						if len(endpoints) > 0 {
							rData.Containers = append(rData.Containers, endpoints[0].IpAddress)
						}
					} else {
						rData.Containers = append(rData.Containers, container.PrimaryIpAddress)
					}
				}
			}
			rancherDataList = append(rancherDataList, rData)
		}
	}

	return rancherDataList
}
