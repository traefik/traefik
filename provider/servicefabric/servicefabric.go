package servicefabric

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
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
func (provider *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {
	if provider.APIVersion == "" {
		provider.APIVersion = "3.0"
	}
	sfClient, err := NewClient(provider.ClusterManagementURL,
		provider.APIVersion,
		provider.ClientCertFilePath,
		provider.ClientCertKeyFilePath,
		provider.CACertFilePath)
	if err != nil {
		return err
	}
	provider.Constraints = append(provider.Constraints, constraints...)

	pool.Go(func(stop chan bool) {
		operation := func() error {
			ticker := time.NewTicker(time.Second * 10)
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

				backends := make(map[string]*types.Backend)
				frontends := make(map[string]*types.Frontend)

				configFromFile, err := loadFrontendConfigFile()
				if err != nil {
					log.Error(err)
				} else {
					if configFromFile.Frontends != nil {
						frontends = configFromFile.Frontends
					}
					if configFromFile.Backends != nil {
						backends = configFromFile.Backends
					}
				}

				apps, err := sfClient.GetApplications()
				if err != nil {
					log.Error(err)
					return err
				}
				for _, app := range apps.Items {
					services, err := sfClient.GetServices(app.ID)
					if err != nil {
						log.Error(err)
						return err
					}
					for _, service := range services.Items {
						_, err := addBackendForService(sfClient, app.ID, &service, backends)
						if err != nil {
							log.Error(err)
						}
					}
				}
				configMessage := types.ConfigMessage{
					ProviderName: "servicefabric",
					Configuration: &types.Configuration{
						Backends:  backends,
						Frontends: frontends,
					},
				}

				configurationChan <- configMessage
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

func addBackendForService(sfClient Client, appID string, service *ServiceItem, backends map[string]*types.Backend) (*types.Backend, error) {

	backend, exists := backends[service.Name]
	if !exists {
		backend = &types.Backend{
			Servers: map[string]types.Server{},
			LoadBalancer: &types.LoadBalancer{
				Method: "wrr",
			},
		}
	} else {
		backend.Servers = make(map[string]types.Server)
	}
	partitions, err := sfClient.GetPartitions(appID, service.ID)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	for _, partition := range partitions.Items {
		if partition.ServiceKind == "Stateful" {
			replicas, err := sfClient.GetReplicas(appID, service.ID, partition.PartitionInformation.ID)

			if err != nil {
				log.Error(err)
				return nil, err
			}
			for _, replica := range replicas.Items {
				if replica.ReplicaStatus != "Ready" || replica.HealthState == "Error" {
					log.Infof("Skipping replica %s health: %s replicaStatus: %s in service %s", replica.ReplicaID, replica.HealthState, replica.ReplicaStatus, service.Name)
					continue
				}

				hasEndpoint, defaultEndpoint := getDefaultEndpoint(replica.Address)
				if !hasEndpoint {
					log.Infof("No default endpoint for replica %s in service %s endpointData: %s", replica.ReplicaID, service.Name, replica.Address)
					// Service may not have a HTTP endpoint so ignore
					continue
				}
				backend.Servers[replica.ReplicaID] = types.Server{
					URL: defaultEndpoint,
				}
			}
		} else if partition.ServiceKind == "Stateless" {
			instances, err := sfClient.GetInstances(appID, service.ID, partition.PartitionInformation.ID)
			if err != nil {
				log.Error(err)
				return nil, err
			}
			for _, instance := range instances.Items {
				if instance.ReplicaStatus != "Ready" || instance.HealthState == "Error" {
					log.Infof("Skipping instance %s health: %s replicaStatus: %s in service %s", instance.InstanceID, instance.HealthState, instance.ReplicaStatus, service.Name)
					continue
				}

				hasEndpoint, defaultEndpoint := getDefaultEndpoint(instance.Address)
				if !hasEndpoint {
					log.Infof("No default endpoint for instance %s in service %s endpointData: %s", instance.InstanceID, service.Name, instance.Address)
					// Service may not have a HTTP endpoint so ignore
					continue
				}
				backend.Servers[instance.InstanceID] = types.Server{
					URL: defaultEndpoint,
				}

			}
		} else {
			log.Errorf("Unsupported service kind %s in service %s", partition.ServiceKind, service.Name)
			continue
		}
	}

	// Only setup config for routable services
	if len(backend.Servers) > 0 {
		backends[service.Name] = backend
	} else {
		log.Infof("No routable backends for %s", service.Name)
	}

	return backend, nil
}

func getDefaultEndpoint(endpointData string) (bool, string) {
	var endpointsMap map[string]map[string]string

	if endpointData == "" {
		return false, ""
	}

	err := json.Unmarshal([]byte(endpointData), &endpointsMap)
	if err != nil {
		log.Error(err)
		return false, ""
	}
	endpoints, endpointsExist := endpointsMap["Endpoints"]
	if !endpointsExist {
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
	return true, strings.Replace(defaultHTTPEndpoint, "localhost", "10.211.55.3", -1)
}

func loadFrontendConfigFile() (*types.Configuration, error) {
	dir, _ := os.Getwd()
	glob := dir + "/../*Config*/**.toml"
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
		return nil, fmt.Errorf("Cannot find fontend config with: %s", glob)
	}

	log.Info("Loading fontend config from:", configFilePath)

	configuration := new(types.Configuration)
	if _, err := toml.DecodeFile(configFilePath, configuration); err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s", err)
	}

	log.Info("Loaded: ", configuration.Frontends)

	return configuration, nil
}
