package servicefabric

import (
	"net/http"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	sf "github.com/jjcollinge/servicefabric"
)

var _ provider.Provider = (*Provider)(nil)

const traefikServiceFabricExtensionKey = "Traefik"

// Provider holds for configuration for the provider
type Provider struct {
	provider.BaseProvider `mapstructure:",squash"`
	ClusterManagementURL  string           `description:"Service Fabric API endpoint"`
	APIVersion            string           `description:"Service Fabric API version" export:"true"`
	RefreshSeconds        int              `description:"Polling interval (in seconds)" export:"true"`
	TLS                   *types.ClientTLS `description:"Enable TLS support" export:"true"`
}

// Provide allows the ServiceFabric provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	if p.APIVersion == "" {
		p.APIVersion = sf.DefaultAPIVersion
	}

	tlsConfig, err := p.TLS.CreateTLSConfig()
	if err != nil {
		return err
	}

	sfClient, err := sf.NewClient(http.DefaultClient, p.ClusterManagementURL, p.APIVersion, tlsConfig)
	if err != nil {
		return err
	}

	if p.RefreshSeconds <= 0 {
		p.RefreshSeconds = 10
	}

	return p.updateConfig(configurationChan, pool, sfClient, time.Duration(p.RefreshSeconds)*time.Second)
}

func (p *Provider) updateConfig(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, sfClient sfClient, pollInterval time.Duration) error {
	pool.Go(func(stop chan bool) {
		operation := func() error {
			ticker := time.NewTicker(pollInterval)
			for range ticker.C {
				select {
				case shouldStop := <-stop:
					if shouldStop {
						ticker.Stop()
						return nil
					}
				default:
					log.Info("Checking service fabric config")
				}

				configuration, err := p.buildConfiguration(sfClient)
				if err != nil {
					return err
				}

				configurationChan <- types.ConfigMessage{
					ProviderName:  "servicefabric",
					Configuration: configuration,
				}
			}
			return nil
		}

		notify := func(err error, time time.Duration) {
			log.Errorf("Provider connection error: %v; retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to Provider: %v", err)
		}
	})
	return nil
}

func getClusterServices(sfClient sfClient) ([]ServiceItemExtended, error) {
	apps, err := sfClient.GetApplications()
	if err != nil {
		return nil, err
	}

	var results []ServiceItemExtended
	for _, app := range apps.Items {
		services, err := sfClient.GetServices(app.ID)
		if err != nil {
			return nil, err
		}

		for _, service := range services.Items {
			item := ServiceItemExtended{
				ServiceItem: service,
				Application: app,
			}

			if labels, err := getLabels(sfClient, &service, &app); err != nil {
				log.Error(err)
			} else {
				item.Labels = labels
			}

			if partitions, err := sfClient.GetPartitions(app.ID, service.ID); err != nil {
				log.Error(err)
			} else {
				for _, partition := range partitions.Items {
					partitionExt := PartitionItemExtended{PartitionItem: partition}

					if isStateful(item) {
						partitionExt.Replicas = getValidReplicas(sfClient, app, service, partition)
					} else if isStateless(item) {
						partitionExt.Instances = getValidInstances(sfClient, app, service, partition)
					} else {
						log.Errorf("Unsupported service kind %s in service %s", partition.ServiceKind, service.Name)
						continue
					}

					item.Partitions = append(item.Partitions, partitionExt)
				}
			}

			results = append(results, item)
		}
	}

	return results, nil
}

func getValidReplicas(sfClient sfClient, app sf.ApplicationItem, service sf.ServiceItem, partition sf.PartitionItem) []sf.ReplicaItem {
	var validReplicas []sf.ReplicaItem

	if replicas, err := sfClient.GetReplicas(app.ID, service.ID, partition.PartitionInformation.ID); err != nil {
		log.Error(err)
	} else {
		for _, instance := range replicas.Items {
			if isHealthy(instance.ReplicaItemBase) && hasHTTPEndpoint(instance.ReplicaItemBase) {
				validReplicas = append(validReplicas, instance)
			}
		}
	}
	return validReplicas
}

func getValidInstances(sfClient sfClient, app sf.ApplicationItem, service sf.ServiceItem, partition sf.PartitionItem) []sf.InstanceItem {
	var validInstances []sf.InstanceItem

	if instances, err := sfClient.GetInstances(app.ID, service.ID, partition.PartitionInformation.ID); err != nil {
		log.Error(err)
	} else {
		for _, instance := range instances.Items {
			if isHealthy(instance.ReplicaItemBase) && hasHTTPEndpoint(instance.ReplicaItemBase) {
				validInstances = append(validInstances, instance)
			}
		}
	}
	return validInstances
}

func isPrimary(instance replicaInstance) bool {
	_, data := instance.GetReplicaData()
	return data.ReplicaRole == "Primary"
}

func isHealthy(instanceData *sf.ReplicaItemBase) bool {
	return instanceData != nil && (instanceData.ReplicaStatus == "Ready" && instanceData.HealthState != "Error")
}

func hasHTTPEndpoint(instanceData *sf.ReplicaItemBase) bool {
	_, err := getReplicaDefaultEndpoint(instanceData)
	return err == nil
}

// Return a set of labels from the Extension and Property manager
// Allow Extension labels to disable importing labels from the property manager.
func getLabels(sfClient sfClient, service *sf.ServiceItem, app *sf.ApplicationItem) (map[string]string, error) {
	labels, err := sfClient.GetServiceExtensionMap(service, app, traefikServiceFabricExtensionKey)
	if err != nil {
		log.Errorf("Error retrieving serviceExtensionMap: %v", err)
		return nil, err
	}

	if label.GetBoolValue(labels, traefikSFEnableLabelOverrides, traefikSFEnableLabelOverridesDefault) {
		if exists, properties, err := sfClient.GetProperties(service.ID); err == nil && exists {
			for key, value := range properties {
				labels[key] = value
			}
		}
	}
	return labels, nil
}
