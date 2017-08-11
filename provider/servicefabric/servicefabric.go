package servicefabric

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/containous/traefik/log"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

type applicationsData struct {
	ContinuationToken string `json:"ContinuationToken"`
	Items             []struct {
		HealthState string `json:"HealthState"`
		ID          string `json:"Id"`
		Name        string `json:"Name"`
		Parameters  []struct {
			Key   string `json:"Key"`
			Value string `json:"Value"`
		} `json:"Parameters"`
		Status      string `json:"Status"`
		TypeName    string `json:"TypeName"`
		TypeVersion string `json:"TypeVersion"`
	} `json:"Items"`
}

type servicesData struct {
	ContinuationToken string `json:"ContinuationToken"`
	Items             []struct {
		HasPersistedState bool   `json:"HasPersistedState"`
		HealthState       string `json:"HealthState"`
		ID                string `json:"Id"`
		IsServiceGroup    bool   `json:"IsServiceGroup"`
		ManifestVersion   string `json:"ManifestVersion"`
		Name              string `json:"Name"`
		ServiceKind       string `json:"ServiceKind"`
		ServiceStatus     string `json:"ServiceStatus"`
		TypeName          string `json:"TypeName"`
	} `json:"Items"`
}

type partitionsData struct {
	ContinuationToken string `json:"ContinuationToken"`
	Items             []struct {
		CurrentConfigurationEpoch struct {
			ConfigurationVersion string `json:"ConfigurationVersion"`
			DataLossVersion      string `json:"DataLossVersion"`
		} `json:"CurrentConfigurationEpoch"`
		HealthState          string `json:"HealthState"`
		MinReplicaSetSize    int64  `json:"MinReplicaSetSize"`
		PartitionInformation struct {
			HighKey              string `json:"HighKey"`
			ID                   string `json:"Id"`
			LowKey               string `json:"LowKey"`
			ServicePartitionKind string `json:"ServicePartitionKind"`
		} `json:"PartitionInformation"`
		PartitionStatus      string `json:"PartitionStatus"`
		ServiceKind          string `json:"ServiceKind"`
		TargetReplicaSetSize int64  `json:"TargetReplicaSetSize"`
	} `json:"Items"`
}

type replicasData []struct {
	Address                      string `json:"Address"`
	HealthState                  int64  `json:"HealthState"`
	LastInBuildDurationInSeconds string `json:"LastInBuildDurationInSeconds"`
	NodeName                     string `json:"NodeName"`
	ReplicaID                    string `json:"ReplicaId"`
	ReplicaRole                  int64  `json:"ReplicaRole"`
	ReplicaStatus                int64  `json:"ReplicaStatus"`
	ServiceKind                  int64  `json:"ServiceKind"`
}

// CatalogProvider holds configurations of the Consul catalog provider.
type Provider struct {
	ClusterName          string
	ClusterManagementUrl string
}

func getHttp(url string) ([]byte, error) {
	resp, err := http.Get(url)

	if err != nil {
		log.Errorf("Cannot connect to servicefabric server %+v", err)
		return nil, err
	}

	body, errB := ioutil.ReadAll(resp.Body)
	if errB != nil {
		log.Errorf("Could not get response body from servicefabric server %+v", err)
	}

	return body, errB
}

func (p *Provider) getApplications() (applicationsData, error) {
	body, err := getHttp(p.ClusterManagementUrl + "/Applications/?api-version=3.0")

	if err != nil {
		return applicationsData{}, err
	}

	sfResponse := applicationsData{}
	err = json.Unmarshal(body, &sfResponse)

	if err != nil {
		log.Errorf("Could not deserialise response from servicefabric server %+v", err)
	}

	return sfResponse, nil
}

func (p *Provider) getServices(appName string) (servicesData, error) {
	body, err := getHttp(p.ClusterManagementUrl + "/Application/" + appName + "/$/GetServices?api-version=3.0")

	if err != nil {
		return servicesData{}, err
	}

	sfResponse := servicesData{}
	err = json.Unmarshal(body, &sfResponse)

	if err != nil {
		log.Errorf("Could not deserialise response from servicefabric server %+v", err)
	}

	return sfResponse, nil
}

func (p *Provider) getPatitions(appName, serviceName string) (partitionsData, error) {
	body, err := getHttp(p.ClusterManagementUrl + "/Application/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/?api-version=3.0")

	if err != nil {
		return partitionsData{}, err
	}

	sfResponse := partitionsData{}
	err = json.Unmarshal(body, &sfResponse)

	if err != nil {
		log.Errorf("Could not deserialise response from servicefabric server %+v", err)
	}

	return sfResponse, nil
}

func (p *Provider) getReplicas(appName, serviceName, parition string) (replicasData, error) {
	body, err := getHttp(p.ClusterManagementUrl + "/Application/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + parition + "/$GetReplicas?api-version=3.0")

	if err != nil {
		return replicasData{}, err
	}

	sfResponse := replicasData{}
	err = json.Unmarshal(body, &sfResponse)

	if err != nil {
		log.Errorf("Could not deserialise response from servicefabric server %+v", err)
	}

	return sfResponse, nil
}

//http://10.0.1.109:19080/Applications/Application1/$/GetServices/Application1%2FWeb1/$/GetPartitions/097d54f7-634a-4d16-a814-47d1642af308/$/GetReplicas?api-version=1.0&_cacheToken=1502467318900

// func (p *Provider) getPartitions(appName string) (servicesData, error) {
// 	body, err := getHttp(p.ClusterManagementUrl + "/Application/" + appName + "/$/GetServices?api-version=3.0")

// 	if err != nil {
// 		return servicesData{}, err
// 	}

// 	sfResponse := servicesData{}
// 	err = json.Unmarshal(body, &sfResponse)

// 	if err != nil {
// 		log.Errorf("Could not deserialise response from servicefabric server %+v", err)
// 	}

// 	return sfResponse, nil
// }

// Provide allows the consul catalog provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {

	ticker := time.NewTicker(time.Second * 2)
	pool.Go(func(stop chan bool) {

		for t := range ticker.C {
			log.Info(t)
			select {
			case shouldStop := <-stop:
				if shouldStop {
					ticker.Stop()
					break
				}
			default:
				fmt.Println("Checking service fabric config")
			}

			backends := make(map[string]*types.Backend)

			backends["wensleydale"] = &types.Backend{
				Servers: map[string]types.Server{
					"server1": types.Server{
						URL: "http://bing.com",
					},
				},
			}

			//Todo: Investigate paging requests
			appData, _ := provider.getApplications()

			for _, a := range appData.Items {
				services, _ := provider.getServices(a.Name)

				for _, s := range services.Items {
					backend := &types.Backend{
						Servers: map[string]types.Server{},
					}

					partitions, _ := provider.getPatitions(a.Name, s.Name)
					for _, p := range partitions.Items {
						replicas, _ := provider.getReplicas(a.Name, s.Name, p.PartitionInformation.ID)
						for _, r := range replicas {
							backend.Servers[r.ReplicaID] = types.Server{
								URL: r.Address,
							}

							log.Info(r.Address)
						}
					}
					backends[a.Name+"/"+s.Name] = backend
				}
			}

			configurationChan <- types.ConfigMessage{
				ProviderName: "servicefabric",
				Configuration: &types.Configuration{
					Backends: backends,
				},
			}

		}
	})

	return nil
}
