package servicefabric

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	sf "github.com/jjcollinge/servicefabric"
)

var _ provider.Provider = (*Provider)(nil)

const traefikLabelPrefix = "traefik"

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

				services, err := getClusterServices(sfClient)
				if err != nil {
					return err
				}

				templateObjects := struct {
					Services []ServiceItemExtended
				}{
					services,
				}

				var sfFuncMap = template.FuncMap{
					"isPrimary":                       isPrimary,
					"getDefaultEndpoint":              p.getDefaultEndpoint,
					"getNamedEndpoint":                p.getNamedEndpoint,
					"getApplicationParameter":         p.getApplicationParameter,
					"doesAppParamContain":             p.doesAppParamContain,
					"hasServiceLabel":                 hasServiceLabel,
					"getServiceLabelValue":            getServiceLabelValue,
					"getServiceLabelValueWithDefault": getServiceLabelValueWithDefault,
					"getServiceLabelsWithPrefix":      getServiceLabelsWithPrefix,
					"getServicesWithLabelValueMap":    getServicesWithLabelValueMap,
					"getServicesWithLabelValue":       getServicesWithLabelValue,
				}

				configuration, err := p.GetConfiguration(tmpl, sfFuncMap, templateObjects)

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

func (p Provider) doesAppParamContain(app sf.ApplicationItem, key, shouldContain string) bool {
	value := p.getApplicationParameter(app, key)
	return strings.Contains(value, shouldContain)
}

func (p Provider) getApplicationParameter(app sf.ApplicationItem, key string) string {
	for _, param := range app.Parameters {
		if param.Key == key {
			return param.Value
		}
	}
	log.Errorf("Parameter %s doesn't exist in app %s", key, app.Name)
	return ""
}

func (p Provider) getDefaultEndpoint(instance replicaInstance) string {
	id, data := instance.GetReplicaData()
	endpoint, err := getDefaultEndpoint(data.Address)
	if err != nil {
		log.Warnf("No default endpoint for replica %s in service %s endpointData: %s", id, data.Address)
		return ""
	}
	return endpoint
}

func (p Provider) getNamedEndpoint(instance replicaInstance, endpointName string) string {
	id, data := instance.GetReplicaData()
	endpoint, err := getNamedEndpoint(data.Address, endpointName)
	if err != nil {
		log.Warnf("No names endpoint of %s for replica %s in endpointData: %s", endpointName, id, data.Address)
		return ""
	}
	return endpoint
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

			if labels, err := sfClient.GetServiceLabels(&service, &app, traefikLabelPrefix); err != nil {
				log.Error(err)
			} else {
				item.Labels = labels
			}

			if partitions, err := sfClient.GetPartitions(app.ID, service.ID); err != nil {
				log.Error(err)
			} else {
				for _, partition := range partitions.Items {
					partitionExt := PartitionItemExtended{PartitionItem: partition}

					if partition.ServiceKind == "Stateful" {
						partitionExt.Replicas = getValidReplicas(sfClient, app, service, partition)
					} else if partition.ServiceKind == "Stateless" {
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

func hasServiceLabel(service ServiceItemExtended, key string) bool {
	_, exists := service.Labels[key]
	return exists
}

func getServiceLabelValue(service ServiceItemExtended, key string) string {
	return service.Labels[key]
}

func getServicesWithLabelValueMap(services []ServiceItemExtended, key string) map[string][]ServiceItemExtended {
	result := map[string][]ServiceItemExtended{}
	for _, service := range services {
		if value, exists := service.Labels[key]; exists {
			if matchingServices, hasKeyAlready := result[value]; hasKeyAlready {
				result[value] = append(matchingServices, service)
			} else {
				result[value] = []ServiceItemExtended{service}
			}
		}
	}
	return result
}

func getServicesWithLabelValue(services []ServiceItemExtended, key, expectedValue string) []ServiceItemExtended {
	var srvWithLabel []ServiceItemExtended
	for _, service := range services {
		value, exists := service.Labels[key]
		if exists && value == expectedValue {
			srvWithLabel = append(srvWithLabel, service)
		}
	}
	return srvWithLabel
}

func getServiceLabelValueWithDefault(service ServiceItemExtended, key, defaultValue string) string {
	value, exists := service.Labels[key]

	if !exists {
		return defaultValue
	}
	return value
}

func getServiceLabelsWithPrefix(service ServiceItemExtended, prefix string) map[string]string {
	results := make(map[string]string)
	for k, v := range service.Labels {
		if strings.HasPrefix(k, prefix) {
			results[k] = v
		}
	}
	return results
}

func isPrimary(instance replicaInstance) bool {
	_, data := instance.GetReplicaData()
	return data.ReplicaRole == "Primary"
}

func isHealthy(instanceData *sf.ReplicaItemBase) bool {
	return instanceData != nil && (instanceData.ReplicaStatus == "Ready" || instanceData.HealthState != "Error")
}

func hasHTTPEndpoint(instanceData *sf.ReplicaItemBase) bool {
	_, err := getDefaultEndpoint(instanceData.Address)
	return err == nil
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

func getDefaultEndpoint(endpointData string) (string, error) {
	endpoints, err := decodeEndpointData(endpointData)
	if err != nil {
		return "", err
	}

	var defaultHTTPEndpointExists bool
	var defaultHTTPEndpoint string
	for _, v := range endpoints {
		if strings.Contains(v, "http") {
			defaultHTTPEndpoint = v
			defaultHTTPEndpointExists = true
			break
		}
	}

	if !defaultHTTPEndpointExists {
		return "", errors.New("no default endpoint found")
	}
	return defaultHTTPEndpoint, nil
}

func getNamedEndpoint(endpointData string, endpointName string) (string, error) {
	endpoints, err := decodeEndpointData(endpointData)
	if err != nil {
		return "", err
	}

	endpoint, exists := endpoints[endpointName]
	if !exists {
		return "", errors.New("endpoint doesn't exist")
	}
	return endpoint, nil
}
