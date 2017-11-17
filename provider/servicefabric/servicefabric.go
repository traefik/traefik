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

const keyPrefix = "traefik."
const traefikExtensionName = "Traefik"

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

// Provide allows the servicefabric provider to provide configurations to traefik
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
	// provider.Constraints = append(provider.Constraints, constraints...)

	p.updateConfig(configurationChan, pool, sfClient, time.Second*10)
	return nil
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

				configuration, err := p.GetConfiguration("templates/servicefabric.tmpl", sfFuncMap, templateObjects)

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
	results := []ServiceItemExtended{}
	apps, err := sfClient.GetApplications()
	if err != nil {
		return nil, err
	}
	for _, app := range apps.Items {
		services, err := sfClient.GetServices(app.ID)
		if err != nil {
			return nil, err
		}

		for _, service := range services.Items {
			item := ServiceItemExtended{
				ServiceItem: service,
				Application: app,
				Labels:      make(map[string]string),
			}

			addLabelsFromServiceExtension(sfClient, service.TypeName, &app, &item)
			addLabelsFromPropertyManager(sfClient, &item)

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
							if len(partitionExt.Replicas) > 0 {
								partitionExt.HasReplicas = true
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
							if len(partitionExt.Instances) > 0 {
								partitionExt.HasInstances = true
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
	value, _ := service.Labels[key]
	return value
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

func (p *Provider) getServicesWithLabel(services []ServiceItemExtended, key string) []ServiceItemExtended {
	srvWithLabel := []ServiceItemExtended{}
	for _, service := range services {
		if p.hasServiceLabel(service, key) {
			srvWithLabel = append(srvWithLabel, service)
		}
	}
	return srvWithLabel
}

func (p *Provider) getServicesWithLabelValue(services []ServiceItemExtended, key, expectedValue string) []ServiceItemExtended {
	srvWithLabel := []ServiceItemExtended{}
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
	primaryString := "Primary"
	if data.ReplicaRole == primaryString {
		return true
	}
	return false
}

func (p *Provider) doesAppParamContain(app sfsdk.ApplicationItem, key, shouldContain string) bool {
	value := p.getApplicationParameter(app, key)
	if strings.Contains(value, shouldContain) {
		return true
	}
	return false
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
	if instanceData.ReplicaStatus == "Ready" || instanceData.HealthState != "Error" {
		return true
	}
	return false
}

func hasHTTPEndpoint(instanceData sfsdk.ReplicaItemBase) bool {
	_, err := getDefaultEndpoint(instanceData.Address)
	return err == nil
}
func decodeEndpointData(endpointData string) (map[string]string, error) {
	var endpointsMap map[string]map[string]string

	if endpointData == "" {
		return nil, errors.New("Endpoint data is empty")
	}

	err := json.Unmarshal([]byte(endpointData), &endpointsMap)
	if err != nil {
		return nil, err
	}
	endpoints, endpointsExist := endpointsMap["Endpoints"]
	if !endpointsExist {
		return nil, errors.New("Endpoint doesn't exist in endpoint data")
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
		return "", errors.New("No default endpoint found")
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
		return "", errors.New("Endpoint doesn't exist")
	}
	return endpoint, nil
}

// Add labels from service manifest extensions
func addLabelsFromServiceExtension(sfClient sfsdk.Client, serviceType string, app *sfsdk.ApplicationItem, service *ServiceItemExtended) error {
	extensionData := ServiceExtensionLabels{}
	err := sfClient.GetServiceExtension(app.TypeName, app.TypeVersion, serviceType, traefikExtensionName, &extensionData)

	if err != nil {
		log.Error(err)
		return err
	}

	if extensionData.Label != nil {
		for _, label := range extensionData.Label {
			if strings.HasPrefix(label.Key, keyPrefix) {
				labelKey := strings.Replace(label.Key, keyPrefix, "", -1)
				log.Debugf("Extension label found for %s with key %s and value %s", service.ID, label.Key, label.Value)
				service.Labels[labelKey] = label.Value
			}
		}
	} else {
		log.Debugf("No Extension found for %s", service.ID)
	}

	return nil
}

// Override labels with runtime values from properties store
func addLabelsFromPropertyManager(sfClient sfsdk.Client, service *ServiceItemExtended) {
	exists, labels, err := sfClient.GetProperties(service.ID)
	if err != nil {
		log.Error(err)
	} else {
		if !exists {
			log.Debugf("Service %s doesn't have any property overrides in PropertyManager", service.ID)
		} else {
			for k, v := range labels {
				if strings.HasPrefix(k, keyPrefix) {
					labelKey := strings.Replace(k, keyPrefix, "", -1)
					log.Debugf("Override label found for %s with key %s and value %s", service.ID, labelKey, v)
					service.Labels[labelKey] = v
				}
			}
		}
	}
}
