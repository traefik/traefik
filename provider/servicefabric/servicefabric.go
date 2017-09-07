package servicefabric

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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

type Provider struct {
	provider.BaseProvider `mapstructure:",squash"`
	ClusterManagementURL  string `description:"Service Fabric API endpoint"`
	APIVersion            string `description:"Service Fabric API version"`
	ClientCertFilePath    string `description:"Path to cert file"`
	ClientCertKeyFilePath string `description:"Path to cert key file"`
	CACertFilePath        string `description:"Path to CA cert file"`
	FontendConfigEnv      string `description:"Environment varaible containing path to frontend config file"`
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
			var lastConfigUpdate types.ConfigMessage

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

				frontends, err := loadFrontendConfigFile(provider.FontendConfigEnv)
				if err != nil {
					log.Error(err)
				}

				backends := make(map[string]*types.Backend)

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
						backend := &types.Backend{
							Servers: map[string]types.Server{},
						}
						partitions, err := sfClient.GetPartitions(app.ID, service.ID)
						if err != nil {
							log.Error(err)
							return err
						}
						for _, partition := range partitions.Items {
							if partition.ServiceKind == "Stateful" {
								replicas, err := sfClient.GetReplicas(app.ID, service.ID, partition.PartitionInformation.ID)
								if err != nil {
									log.Error(err)
									return err
								}
								for _, replica := range replicas.Items {
									defaultEndpoint, err := getDefaultEndpoint(replica.Address)
									if err != nil {
										log.Errorf("%s for replica %s in service %s", err, replica.ReplicaID, service.Name)
										// Service may not have a HTTP endpoint so ignore
										continue
									}
									backend.Servers[replica.ReplicaID] = types.Server{
										URL: defaultEndpoint,
									}
								}
							} else if partition.ServiceKind == "Stateless" {
								instances, err := sfClient.GetInstances(app.ID, service.ID, partition.PartitionInformation.ID)
								if err != nil {
									log.Error(err)
									return err
								}
								for _, instance := range instances.Items {
									defaultEndpoint, err := getDefaultEndpoint(instance.Address)
									if err != nil {
										log.Errorf("%s for instance %s in service %s", err, instance.InstanceID, service.Name)
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

					}
				}
				configMessage := types.ConfigMessage{
					ProviderName: "servicefabric",
					Configuration: &types.Configuration{
						Backends:  backends,
						Frontends: *frontends,
					},
				}
				if !reflect.DeepEqual(lastConfigUpdate, configMessage) {
					log.Info("New configuration for service fabric:", configMessage)
					configurationChan <- configMessage
					lastConfigUpdate = configMessage
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

func getDefaultEndpoint(endpointData string) (string, error) {
	var endpointsMap map[string]map[string]string
	var emptyString string
	err := json.Unmarshal([]byte(endpointData), &endpointsMap)
	if err != nil {
		log.Error(err)
		return emptyString, errors.New("Failed to deserialize endpoints")
	}
	endpoints, endpointsExist := endpointsMap["Endpoints"]
	if !endpointsExist {
		return emptyString, fmt.Errorf("No endpoints")
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
		return emptyString, fmt.Errorf("No default HTTP endpoint")
	}
	return strings.Replace(defaultHTTPEndpoint, "localhost", "10.211.55.3", -1), nil
}

func loadFrontendConfigFile(envVar string) (*map[string]*types.Frontend, error) {
	directory := os.Getenv(envVar)
	files, _ := filepath.Glob(directory + "/*Config*/**.toml")

	var mostRecentFile os.FileInfo
	var configFilePath string
	for _, file := range files {
		fileInfo, _ := os.Stat(file)
		if mostRecentFile == nil || fileInfo.ModTime().After(mostRecentFile.ModTime()) {
			mostRecentFile = fileInfo
			configFilePath = file
		}
	}

	configuration := new(types.Configuration)
	if _, err := toml.DecodeFile(configFilePath, configuration); err != nil {
		return nil, fmt.Errorf("error reading configuration file: %s", err)
	}

	log.Info("Loading fontend config from:", configFilePath)
	log.Info(configuration.Frontends)

	return &configuration.Frontends, nil
}
