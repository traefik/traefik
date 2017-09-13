package servicefabric

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds for configuration for the provider
type Provider struct {
	provider.BaseProvider `mapstructure:",squash"`
	ClusterManagementURL  string `description:"Service Fabric API endpoint"`
	APIVersion            string `description:"Service Fabric API version"`
	ClientCertFilePath    string `description:"Path to cert file"`
	ClientCertKeyFilePath string `description:"Path to cert key file"`
	CACertFilePath        string `description:"Path to CA cert file"`
}

// Provide allows the servicefabric provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	if p.APIVersion == "" {
		p.APIVersion = "3.0"
	}
	webClient := &httpWebClient{client: http.Client{}}
	sfClient, err := NewClient(
		webClient,
		p.ClusterManagementURL,
		p.APIVersion,
		p.ClientCertFilePath,
		p.ClientCertKeyFilePath,
		p.CACertFilePath)
	if err != nil {
		return err
	}
	// provider.Constraints = append(provider.Constraints, constraints...)

	p.updateConfig(configurationChan, pool, sfClient, time.Second*10)
	return nil
}

func (p *Provider) updateConfig(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, sfClient Client, pollInterval time.Duration) error {
	pool.Go(func(stop chan bool) {
		operation := func() error {
			ticker := time.NewTicker(pollInterval)
			for _ = range ticker.C {
				select {
				case shouldStop := <-stop:
					if shouldStop {
						ticker.Stop()
						return nil
					}
				default:
					log.Info("Checking service fabric config")
				}

				// SF Note: Deploying a config update could change the location or provide
				// a new version of this file, that is why we query for location each time.
				if p.Filename == "" {
					configFilePath, err := findTemplateFile()
					if err != nil {
						log.Infof("Using default template file: %v", err)
					} else {
						p.Filename = configFilePath
					}
				}

				services, err := p.getClusterServices(sfClient)
				if err != nil {
					log.Error(err)
					return err
				}

				templateObjects := struct {
					Services []ServiceItemExtended
				}{
					services,
				}

				var sfFuncMap = template.FuncMap{
					"isPrimary":               p.isPrimary,
					"isHealthy":               p.isHealthy,
					"hasHTTPEndpoint":         p.hasHTTPEndpoint,
					"getDefaultEndpoint":      p.getDefaultEndpoint,
					"getNamedEndpoint":        p.getNamedEndpoint,
					"getApplicationParameter": p.getApplicationParameter,
					"doesAppParamContain":     p.doesAppParamContain,
				}

				configuration, err := p.GetConfiguration("templates/servicefabric.tmpl", sfFuncMap, templateObjects)

				if err != nil {
					log.Error(err)
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

func (p *Provider) getClusterServices(sfClient Client) ([]ServiceItemExtended, error) {
	results := []ServiceItemExtended{}
	apps, err := sfClient.GetApplications()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, app := range apps.Items {
		services, err := sfClient.GetServices(app.ID)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		for _, service := range services.Items {
			item := ServiceItemExtended{
				ServiceItem: service,
				Application: app,
			}

			partitions, err := sfClient.GetPartitions(app.ID, service.ID)
			if err != nil {
				log.Error(err)
			} else {
				for _, partition := range partitions.Items {
					partitionExt := PartitionItemExtended{
						PartitionItem: partition,
					}

					if partition.ServiceKind == "Stateful" {
						replicas, err := sfClient.GetReplicas(app.ID, service.ID, partition.PartitionInformation.ID)

						if err != nil {
							log.Error(err)
						} else {
							partitionExt.Replicas = replicas.Items
							partitionExt.HasReplicas = true
						}
					} else if partition.ServiceKind == "Stateless" {
						instances, err := sfClient.GetInstances(app.ID, service.ID, partition.PartitionInformation.ID)

						if err != nil {
							log.Error(err)
						} else {
							partitionExt.Instances = instances.Items
							partitionExt.HasInstances = true
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

func (p *Provider) isPrimary(i ReplicaInstance) bool {
	_, data := i.GetReplicaData()
	primaryString := "Primary"
	if data.ReplicaRole == primaryString {
		return true
	}
	return false
}

func (p *Provider) isHealthy(i ReplicaInstance) bool {
	_, data := i.GetReplicaData()
	if data.ReplicaStatus == "Ready" || data.HealthState != "Error" {
		return true
	}
	return false
}

func (p *Provider) doesAppParamContain(a ApplicationItem, key, shouldContain string) bool {
	value := p.getApplicationParameter(a, key)
	if strings.Contains(value, shouldContain) {
		return true
	}
	return false
}

func (p *Provider) getApplicationParameter(a ApplicationItem, k string) string {
	for _, param := range a.Parameters {
		if param.Key == k {
			return param.Value
		}
	}
	log.Errorf("Parameter %s doesn't exist in app %s", k, a.Name)
	return ""
}

func (p *Provider) hasHTTPEndpoint(i ReplicaInstance) bool {
	_, data := i.GetReplicaData()
	exists, _ := getDefaultEndpoint(data.Address)
	if exists {
		return true
	}
	return false
}

func (p *Provider) getDefaultEndpoint(i ReplicaInstance) string {
	id, data := i.GetReplicaData()
	exists, endpoint := getDefaultEndpoint(data.Address)
	if !exists {
		log.Infof("No default endpoint for replica %s in service %s endpointData: %s", id, data.Address)
		return ""
	}
	return endpoint
}

func (p *Provider) getNamedEndpoint(i ReplicaInstance, endpointName string) string {
	id, data := i.GetReplicaData()
	exists, endpoint := getNamedEndpoint(data.Address, endpointName)
	if !exists {
		log.Infof("No names endpoint of %s for replica %s in endpointData: %s", endpointName, id, data.Address)
		return ""
	}
	return endpoint
}

func decodeEndpointData(endpointData string) (bool, map[string]string) {
	var endpointsMap map[string]map[string]string

	if endpointData == "" {
		return false, nil
	}

	err := json.Unmarshal([]byte(endpointData), &endpointsMap)
	if err != nil {
		log.Error(err)
		return false, nil
	}
	endpoints, endpointsExist := endpointsMap["Endpoints"]
	if !endpointsExist {
		return false, nil
	}

	return true, endpoints
}

func getDefaultEndpoint(endpointData string) (bool, string) {
	isValid, endpoints := decodeEndpointData(endpointData)
	if !isValid {
		return false, ""
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
		return false, ""
	}
	return true, defaultHTTPEndpoint
}

func getNamedEndpoint(endpointData string, endpointName string) (bool, string) {
	isValid, endpoints := decodeEndpointData(endpointData)
	if !isValid {
		return false, ""
	}
	endpoint, exists := endpoints[endpointName]
	if !exists {
		return false, ""
	}
	return true, endpoint
}

func findTemplateFile() (string, error) {
	dir, _ := os.Getwd()
	glob := dir + "/../*Config*/**.toml.tmpl"
	files, _ := filepath.Glob(glob)

	var mostRecentFile os.FileInfo
	var configFilePath string
	for _, file := range files {
		fileInfo, _ := os.Stat(file)
		if mostRecentFile == nil || fileInfo.ModTime().After(mostRecentFile.ModTime()) {
			mostRecentFile = fileInfo
			configFilePath = file
		}
	}

	if configFilePath == "" {
		return "", fmt.Errorf("Cannot find template file with glob %s using default", glob)
	}

	log.Info("Found sf template toml from:", configFilePath)

	return configFilePath, nil
}
