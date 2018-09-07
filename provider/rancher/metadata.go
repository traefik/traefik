package rancher

import (
	"context"
	"fmt"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/sirupsen/logrus"

	rancher "github.com/rancher/go-rancher-metadata/metadata"
)

// MetadataConfiguration contains configuration properties specific to
// the Rancher metadata service provider.
type MetadataConfiguration struct {
	IntervalPoll bool   `description:"Poll the Rancher metadata service every 'rancher.refreshseconds' (less accurate)"`
	Prefix       string `description:"Prefix used for accessing the Rancher metadata service"`
}

func (p *Provider) metadataProvide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
	metadataServiceURL := fmt.Sprintf("http://rancher-metadata.rancher.internal/%s", p.Metadata.Prefix)

	safe.Go(func() {
		operation := func() error {
			client, err := rancher.NewClientAndWait(metadataServiceURL)
			if err != nil {
				log.Errorf("Failed to create Rancher metadata service client: %v", err)
				return err
			}

			updateConfiguration := func(version string) {
				log.WithField("metadata_version", version).Debugln("Refreshing configuration from Rancher metadata service")

				stacks, err := client.GetStacks()
				if err != nil {
					log.Errorf("Failed to query Rancher metadata service: %v", err)
					return
				}

				rancherData := parseMetadataSourcedRancherData(stacks)
				configuration := p.buildConfiguration(rancherData)
				configurationChan <- types.ConfigMessage{
					ProviderName:  "rancher",
					Configuration: configuration,
				}
			}
			updateConfiguration("init")

			if p.Watch {
				pool.Go(func(stop chan bool) {
					switch {
					case p.Metadata.IntervalPoll:
						p.intervalPoll(client, updateConfiguration, stop)
					default:
						p.longPoll(client, updateConfiguration, stop)
					}
				})
			}
			return nil
		}

		notify := func(err error, time time.Duration) {
			log.WithFields(logrus.Fields{
				"error":    err,
				"retry_in": time,
			}).Errorln("Rancher metadata service connection error")
		}

		if err := backoff.RetryNotify(operation, job.NewBackOff(backoff.NewExponentialBackOff()), notify); err != nil {
			log.WithField("endpoint", metadataServiceURL).Errorln("Cannot connect to Rancher metadata service")
		}
	})

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
				log.WithField("error", err).Errorln("Failed to read Rancher metadata service version")
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
	// favoured for the most accurate configuration updates.
	safe.Go(func() {
		client.OnChange(p.RefreshSeconds, updateConfiguration)
	})
	<-stop
}

func parseMetadataSourcedRancherData(stacks []rancher.Stack) (rancherDataList []rancherData) {
	for _, stack := range stacks {
		for _, service := range stack.Services {
			var containerIPAddresses []string
			for _, container := range service.Containers {
				if containerFilter(container.Name, container.HealthState, container.State) {
					containerIPAddresses = append(containerIPAddresses, container.PrimaryIp)
				}
			}

			rancherDataList = append(rancherDataList, rancherData{
				Name:       service.Name + "/" + stack.Name,
				State:      service.State,
				Labels:     service.Labels,
				Containers: containerIPAddresses,
			})
		}
	}
	return rancherDataList
}
