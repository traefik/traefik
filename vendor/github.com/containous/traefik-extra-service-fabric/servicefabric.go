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
	sfsdk "github.com/jjcollinge/servicefabric"
)

var _ provider.Provider = (*Provider)(nil)

const traefikLabelPrefix = "traefik"

// Provider holds for configuration for the provider
type Provider struct {
	provider.BaseProvider `mapstructure:",squash"`
	ClusterManagementURL  string `description:"Service Fabric API endpoint"`
	APIVersion            string `description:"Service Fabric API version"`
	UseCertificateAuth    bool   `description:"Should use certificate authentication"`
	ClientCertFilePath    string `description:"Path to cert file"`
	ClientCertKeyFilePath string `description:"Path to cert key file"`
	InsecureSkipVerify    bool   `description:"Skip verification of server ca certificate"`
}

// Provide allows the ServiceFabric provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	if p.APIVersion == "" {
		p.APIVersion = "3.0"
	}
	webClient := sfsdk.NewHTTPClient(http.Client{})
	sfClient, err := sfsdk.NewClient(
		webClient,
		p.ClusterManagementURL,
		p.APIVersion,
		p.ClientCertFilePath,
		p.ClientCertKeyFilePath,
		p.InsecureSkipVerify)
	if err != nil {
		return err
	}

	return p.updateConfig(configurationChan, pool, sfClient, time.Second*10)
}

func (p *Provider) updateConfig(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, sfClient sfsdk.Client, pollInterval time.Duration) error {
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

				services, err := p.getClusterServices(sfClient)
				if err != nil {
					return err
				}

				templateObjects := struct {
					Services []ServiceItemExtended
				}{
					services,
				}

				var sfFuncMap = template.FuncMap{
					"isPrimary":                       p.isPrimary,
					"getDefaultEndpoint":              p.getDefaultEndpoint,
					"getNamedEndpoint":                p.getNamedEndpoint,
					"getApplicationParameter":         p.getApplicationParameter,
					"doesAppParamContain":             p.doesAppParamContain,
					"hasServiceLabel":                 p.hasServiceLabel,
					"getServiceLabelValue":            p.getServiceLabelValue,
					"getServiceLabelValueWithDefault": p.getServiceLabelValueWithDefault,
					"getServiceLabelsWithPrefix":      p.getServiceLabelsWithPrefix,
					"getServicesWithLabelValueMap":    p.getServicesWithLabelValueMap,
					"getServicesWithLabelValue":       p.getServicesWithLabelValue,
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
			log.Errorf("Provider connection error: %s; retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to Provider: %s", err)
		}
	})
	return nil
}

func (p *Provider) getClusterServices(sfClient sfsdk.Client) ([]ServiceItemExtended, error) {
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

			labels, err := sfClient.GetServiceLabels(&service, &app, traefikLabelPrefix)

			if err != nil {
				log.Error(err)
			} else {
				item.Labels = labels
			}

			partitions, err := sfClient.GetPartitions(app.ID, service.ID)
			if err != nil {
				log.Error(err)
			} else {
				for _, partition := range partitions.Items {
					partitionExt := PartitionItemExtended{
						PartitionItem: partition,
						Replicas:      []sfsdk.ReplicaItem{},
					}

					if partition.ServiceKind == "Stateful" {
						replicas, err := sfClient.GetReplicas(app.ID, service.ID, partition.PartitionInformation.ID)

						if err != nil {
							log.Error(err)
						} else {
							for _, instance := range replicas.Items {
								if isHealthy(*instance.ReplicaItemBase) && hasHTTPEndpoint(*instance.ReplicaItemBase) {
									partitionExt.Replicas = append(partitionExt.Replicas, instance)
								}
							}
						}
					} else if partition.ServiceKind == "Stateless" {
						instances, err := sfClient.GetInstances(app.ID, service.ID, partition.PartitionInformation.ID)

						if err != nil {
							log.Error(err)
						} else {
							for _, instance := range instances.Items {
								if isHealthy(*instance.ReplicaItemBase) && hasHTTPEndpoint(*instance.ReplicaItemBase) {
									partitionExt.Instances = append(partitionExt.Instances, instance)
								}
							}
						}
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

func (p *Provider) hasServiceLabel(service ServiceItemExtended, key string) bool {
	_, exists := service.Labels[key]
	return exists
}

func (p *Provider) getServiceLabelValue(service ServiceItemExtended, key string) string {
	return service.Labels[key]
}

func (p *Provider) getServicesWithLabelValueMap(services []ServiceItemExtended, key string) map[string][]ServiceItemExtended {
	result := map[string][]ServiceItemExtended{}
	for _, service := range services {
		if value, exists := service.Labels[key]; exists {
			matchingServices, hasKeyAlready := result[value]
			if hasKeyAlready {
				result[value] = append(matchingServices, service)
			} else {
				result[value] = []ServiceItemExtended{service}
			}
		}
	}
	return result
}

func (p *Provider) getServicesWithLabelValue(services []ServiceItemExtended, key, expectedValue string) []ServiceItemExtended {
	var srvWithLabel []ServiceItemExtended
	for _, service := range services {
		value, exists := service.Labels[key]
		if exists && value == expectedValue {
			srvWithLabel = append(srvWithLabel, service)
		}
	}
	return srvWithLabel
}

func (p *Provider) getServiceLabelValueWithDefault(service ServiceItemExtended, key, defaultValue string) string {
	value, exists := service.Labels[key]

	if !exists {
		return defaultValue
	}

	return value
}

func (p *Provider) getServiceLabelsWithPrefix(service ServiceItemExtended, prefix string) map[string]string {
	results := make(map[string]string)
	for k, v := range service.Labels {
		if strings.HasPrefix(k, prefix) {
			results[k] = v
		}
	}
	return results
}

func (p *Provider) isPrimary(instance sfsdk.ReplicaInstance) bool {
	_, data := instance.GetReplicaData()
	return data.ReplicaRole == "Primary"
}

func (p *Provider) doesAppParamContain(app sfsdk.ApplicationItem, key, shouldContain string) bool {
	value := p.getApplicationParameter(app, key)
	return strings.Contains(value, shouldContain)
}

func (p *Provider) getApplicationParameter(app sfsdk.ApplicationItem, key string) string {
	for _, param := range app.Parameters {
		if param.Key == key {
			return param.Value
		}
	}
	log.Errorf("Parameter %s doesn't exist in app %s", key, app.Name)
	return ""
}

func (p *Provider) getDefaultEndpoint(instance sfsdk.ReplicaInstance) string {
	id, data := instance.GetReplicaData()
	endpoint, err := getDefaultEndpoint(data.Address)
	if err != nil {
		log.Warnf("No default endpoint for replica %s in service %s endpointData: %s", id, data.Address)
		return ""
	}
	return endpoint
}

func (p *Provider) getNamedEndpoint(instance sfsdk.ReplicaInstance, endpointName string) string {
	id, data := instance.GetReplicaData()
	endpoint, err := getNamedEndpoint(data.Address, endpointName)
	if err != nil {
		log.Warnf("No names endpoint of %s for replica %s in endpointData: %s", endpointName, id, data.Address)
		return ""
	}
	return endpoint
}

func isHealthy(instanceData sfsdk.ReplicaItemBase) bool {
	return instanceData.ReplicaStatus == "Ready" || instanceData.HealthState != "Error"
}

func hasHTTPEndpoint(instanceData sfsdk.ReplicaItemBase) bool {
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
		isHTTP := strings.Contains(v, "http")
		if isHTTP {
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
