package servicefabric

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/flaeg"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	appinsights "github.com/jjcollinge/logrus-appinsights"
	sf "github.com/jjcollinge/servicefabric"
)

var _ provider.Provider = (*Provider)(nil)

const traefikServiceFabricExtensionKey = "Traefik"

const (
	kindStateful  = "Stateful"
	kindStateless = "Stateless"
)

// Provider holds for configuration for the provider
type Provider struct {
	provider.BaseProvider `mapstructure:",squash"`
	ClusterManagementURL  string           `description:"Service Fabric API endpoint"`
	APIVersion            string           `description:"Service Fabric API version" export:"true"`
	RefreshSeconds        flaeg.Duration   `description:"Polling interval (in seconds)" export:"true"`
	TLS                   *types.ClientTLS `description:"Enable TLS support" export:"true"`
	AppInsightsClientName string           `description:"The client name, Identifies the cloud instance"`
	AppInsightsKey        string           `description:"Application Insights Instrumentation Key"`
	AppInsightsBatchSize  int              `description:"Number of trace lines per batch, optional"`
	AppInsightsInterval   flaeg.Duration   `description:"The interval for sending data to Application Insights, optional"`
	sfClient              sfClient
}

// Init the provider
func (p *Provider) Init(constraints types.Constraints) error {
	err := p.BaseProvider.Init(constraints)
	if err != nil {
		return err
	}

	if p.APIVersion == "" {
		p.APIVersion = sf.DefaultAPIVersion
	}

	tlsConfig, err := p.TLS.CreateTLSConfig()
	if err != nil {
		return err
	}

	p.sfClient, err = sf.NewClient(http.DefaultClient, p.ClusterManagementURL, p.APIVersion, tlsConfig)
	if err != nil {
		return err
	}

	if p.RefreshSeconds <= 0 {
		p.RefreshSeconds = flaeg.Duration(10 * time.Second)
	}

	if p.AppInsightsClientName != "" && p.AppInsightsKey != "" {
		if p.AppInsightsBatchSize == 0 {
			p.AppInsightsBatchSize = 10
		}
		if p.AppInsightsInterval == 0 {
			p.AppInsightsInterval = flaeg.Duration(5 * time.Second)
		}
		createAppInsightsHook(p.AppInsightsClientName, p.AppInsightsKey, p.AppInsightsBatchSize, p.AppInsightsInterval)
	}
	return nil
}

// Provide allows the ServiceFabric provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
	return p.updateConfig(configurationChan, pool, time.Duration(p.RefreshSeconds))
}

func (p *Provider) updateConfig(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, pollInterval time.Duration) error {
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

				configuration, err := p.getConfiguration()
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

func (p *Provider) getConfiguration() (*types.Configuration, error) {
	services, err := getClusterServices(p.sfClient)
	if err != nil {
		return nil, err
	}

	return p.buildConfiguration(services)
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

					switch {
					case isStateful(item):
						partitionExt.Replicas = getValidReplicas(sfClient, app, service, partition)
					case isStateless(item):
						partitionExt.Instances = getValidInstances(sfClient, app, service, partition)
					default:
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

func isHealthy(instanceData *sf.ReplicaItemBase) bool {
	return instanceData != nil && (instanceData.ReplicaStatus == "Ready" && instanceData.HealthState != "Error")
}

func hasHTTPEndpoint(instanceData *sf.ReplicaItemBase) bool {
	_, err := getReplicaDefaultEndpoint(instanceData)
	return err == nil
}

func getReplicaDefaultEndpoint(replicaData *sf.ReplicaItemBase) (string, error) {
	endpoints, err := decodeEndpointData(replicaData.Address)
	if err != nil {
		return "", err
	}

	var defaultHTTPEndpoint string
	for _, v := range endpoints {
		if strings.Contains(v, "http") {
			defaultHTTPEndpoint = v
			break
		}
	}

	if len(defaultHTTPEndpoint) == 0 {
		return "", errors.New("no default endpoint found")
	}
	return defaultHTTPEndpoint, nil
}

func decodeEndpointData(endpointData string) (map[string]string, error) {
	var endpointsMap map[string]map[string]string

	if endpointData == "" {
		return nil, errors.New("endpoint data is empty")
	}

	err := json.Unmarshal([]byte(endpointData), &endpointsMap)
	if err != nil {
		return nil, err
	}

	endpoints, endpointsExist := endpointsMap["Endpoints"]
	if !endpointsExist {
		return nil, errors.New("endpoint doesn't exist in endpoint data")
	}

	return endpoints, nil
}

func isStateful(service ServiceItemExtended) bool {
	return service.ServiceKind == kindStateful
}

func isStateless(service ServiceItemExtended) bool {
	return service.ServiceKind == kindStateless
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

func createAppInsightsHook(appInsightsClientName string, instrumentationKey string, maxBatchSize int, interval flaeg.Duration) {
	hook, err := appinsights.New(appInsightsClientName, appinsights.Config{
		InstrumentationKey: instrumentationKey,
		MaxBatchSize:       maxBatchSize,            // optional
		MaxBatchInterval:   time.Duration(interval), // optional
	})
	if err != nil || hook == nil {
		panic(err)
	}

	// ignore fields
	hook.AddIgnore("private")
	log.AddHook(hook)
}
