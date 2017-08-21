package servicefabric

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

type ApplicationItem struct {
	HealthState string `json:"HealthState"`
	ID          string `json:"Id"`
	Name        string `json:"Name"`
	Parameters  []*struct {
		Key   string `json:"Key"`
		Value string `json:"Value"`
	} `json:"Parameters"`
	Status      string `json:"Status"`
	TypeName    string `json:"TypeName"`
	TypeVersion string `json:"TypeVersion"`
}

type applicationsData struct {
	ContinuationToken *string           `json:"ContinuationToken"`
	Items             []ApplicationItem `json:"Items"`
}

type servicesData struct {
	ContinuationToken *string `json:"ContinuationToken"`
	Items             []*struct {
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
	ContinuationToken *string `json:"ContinuationToken"`
	Items             []*struct {
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

type replicasData struct {
	ContinuationToken *string `json:"ContinuationToken"`
	Items             []*struct {
		Address                      string `json:"Address"`
		HealthState                  string `json:"HealthState"`
		LastInBuildDurationInSeconds string `json:"LastInBuildDurationInSeconds"`
		NodeName                     string `json:"NodeName"`
		ReplicaID                    string `json:"ReplicaId"`
		ReplicaRole                  string `json:"ReplicaRole"`
		ReplicaStatus                string `json:"ReplicaStatus"`
		ServiceKind                  string `json:"ServiceKind"`
	} `json:"Items"`
}

type serviceType struct {
	ServiceTypeDescription struct {
		IsStateful           bool   `json:"IsStateful"`
		ServiceTypeName      string `json:"ServiceTypeName"`
		PlacementConstraints string `json:"PlacementConstraints"`
		HasPersistedState    bool   `json:"HasPersistedState"`
		Kind                 string `json:"Kind"`
		Extensions           []struct {
			Key   string `json:"Key"`
			Value string `json:"Value"`
		} `json:"Extensions"`
		LoadMetrics              []interface{} `json:"LoadMetrics"`
		ServicePlacementPolicies []interface{} `json:"ServicePlacementPolicies"`
	} `json:"ServiceTypeDescription"`
	ServiceManifestVersion string `json:"ServiceManifestVersion"`
	ServiceManifestName    string `json:"ServiceManifestName"`
	IsServiceGroup         bool   `json:"IsServiceGroup"`
}

type routeMap struct {
	Routes []struct {
		Rule string `json:"rule"`
	} `json:"routes"`
}

type serviceManifest struct {
	XMLName     xml.Name `xml:"ServiceManifest"`
	Description string   `xml:"Description"`
}

var _ provider.Provider = (*Provider)(nil)

type Provider struct {
	ClusterManagementUrl  string `description:"ServiceFabric cluster management endpoint"`
	UseCertificateAuth    bool   `description:"User certificate auth"`
	ClientCertFilePath    string `description:"Path to cert file"`
	ClientCertKeyFilePath string `description:"Path to cert key file"`
	CACertFilePath        string `description:"Path to CA cert file"`
}

// Method for invoking generic HTTP GET requests against the Service Fabric REST API
func (p *Provider) getHttp(url string) ([]byte, error) {
	log.Debugf("GET: %s", url)
	var client http.Client

	if !p.UseCertificateAuth {
		client = http.Client{}
	} else {
		// Handle client authentication

		// Load client cert from file
		cert, err := tls.LoadX509KeyPair(p.ClientCertFilePath, p.ClientCertKeyFilePath)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		// Load CA cert from file
		caCert, err := ioutil.ReadFile(p.CACertFilePath)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		}
		tlsConfig.BuildNameToCertificate()
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		client = http.Client{Transport: transport}
	}

	resp, err := client.Get(url)

	if err != nil {
		log.Errorf("Cannot connect to servicefabric server %+v on %s", err, url)
		return nil, err
	}

	body, errB := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Errorf("Service fabric responded with error code %s to request %s with body %s", resp.Status, url, body)
		return nil, errors.New("Servicefabric returned error code")
	}

	if errB != nil {
		log.Errorf("Could not get response body from servicefabric server %+v", err)
	}

	return body, nil
}

// Retrieve a list of the registered applications
func (p *Provider) getApplications() (applicationsData, error) {
	body, err := p.getHttp(p.ClusterManagementUrl + "/Applications/?api-version=3.0")

	if err != nil {
		return applicationsData{}, err
	}

	var sfResponse applicationsData
	err = json.Unmarshal(body, &sfResponse)

	if err != nil {
		log.Errorf("Could not deserialise response from servicefabric server %+v", err)
	}

	return sfResponse, nil
}

// Retrieve a list of the registered services for a given application
func (p *Provider) getServices(appName string) (servicesData, error) {
	body, err := p.getHttp(p.ClusterManagementUrl + "/Applications/" + appName + "/$/GetServices?api-version=3.0")

	if err != nil {
		return servicesData{}, err
	}

	sfResponse := servicesData{}
	err = json.Unmarshal(body, &sfResponse)

	if err != nil {
		log.Errorf("Could not deserialise response from servicefabric server %+v on body %s", err, body)
	}

	return sfResponse, nil
}

// Retrieve a list of partitions for a given service
func (p *Provider) getPatitions(appName, serviceName string) (partitionsData, error) {
	body, err := p.getHttp(p.ClusterManagementUrl + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/?api-version=3.0")

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

// Retrieve a list of replicas for a given partiton
func (p *Provider) getReplicas(appName, serviceName, parition string) (replicasData, error) {
	body, err := p.getHttp(p.ClusterManagementUrl + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + parition + "/$/GetReplicas?api-version=3.0")

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

// Gets the default endpont for a given set of endpoints
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

	defaultEndpoint, defaultExists := endpoints[""]

	if !defaultExists {
		return emptyString, fmt.Errorf("No default endpoint")
	}
	return defaultEndpoint, nil
}

// Gets the routing information for a given service
func (p *Provider) getServiceRoutes(appTypeName, appTypeVersion, manifestName string) (routeMap, error) {
	body, err := p.getHttp(p.ClusterManagementUrl + "/ApplicationTypes/" + appTypeName + "/$/GetServiceManifest/?api-version=3.0&ApplicationTypeVersion=" + appTypeVersion + "&ServiceManifestName=" + manifestName)

	if err != nil {
		panic(err)
	}

	var payload map[string]string
	err = json.Unmarshal(body, &payload)

	if err != nil {
		panic(err)
	}

	var manifest serviceManifest
	err = xml.Unmarshal([]byte(payload["Manifest"]), &manifest)

	if err != nil {
		panic(err)
	}

	log.Debugf("Manifest: %s", string(manifest.Description))

	var routes routeMap
	err = json.Unmarshal([]byte(manifest.Description), &routes)

	if err != nil {
		return routes, err
	}

	return routes, nil
}

// Gets the manifest file name associated with a given service
func (p *Provider) getServiceManifestName(appType, appTypeVer, serviceName string) string {
	body, err := p.getHttp(p.ClusterManagementUrl + "/ApplicationTypes/" + appType + "/$/GetServiceTypes?ApplicationTypeVersion=" + appTypeVer + "&api-version=1.0")

	if err != nil {
		panic(err)
	}

	var serviceTypes []serviceType
	err = json.Unmarshal([]byte(body), &serviceTypes)

	if err != nil {
		panic(err)
	}

	var serviceManifestName string
	for _, s := range serviceTypes {
		if s.ServiceTypeDescription.ServiceTypeName == serviceName {
			serviceManifestName = s.ServiceManifestName
			break
		}
	}

	if len(serviceManifestName) <= 0 {
		panic("Service name does not exist!")
	}

	return serviceManifestName
}

// Provide allows the servicefabric provider to provide configurations to traefik
// using the given configuration channel.
func (provider *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {

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

				backends := make(map[string]*types.Backend)
				frontends := make(map[string]*types.Frontend)

				//Todo: Investigate paging requests
				appData, err := provider.getApplications()
				if err != nil {
					log.Error(err)
					return err
				}

				for _, a := range appData.Items {

					services, err := provider.getServices(a.ID)
					if err != nil {
						log.Error(err)
						return err
					}

					for _, s := range services.Items {

						backend := &types.Backend{
							Servers: map[string]types.Server{},
						}

						partitions, err := provider.getPatitions(a.ID, s.ID)
						if err != nil {
							log.Error(err)
							return err
						}
						for _, p := range partitions.Items {

							replicas, err := provider.getReplicas(a.ID, s.ID, p.PartitionInformation.ID)
							if err != nil {
								log.Error(err)

								return err
							}
							for _, r := range replicas.Items {

								defaultEndpoint, err := getDefaultEndpoint(r.Address)

								if err != nil {
									log.Errorf("%s for replica %s in service %s", err, r.ReplicaID, s.Name)
									continue
								}

								backend.Servers[r.ReplicaID] = types.Server{
									URL: defaultEndpoint,
								}

							}
						}
						backends[s.Name] = backend

						manifestName := provider.getServiceManifestName(a.TypeName, a.TypeVersion, s.TypeName)
						routeMap, err := provider.getServiceRoutes(a.TypeName, a.TypeVersion, manifestName)

						if err != nil {
							continue
						}

						routes := make(map[string]types.Route)
						for i, r := range routeMap.Routes {
							routeName := "route" + fmt.Sprint(i)
							route := types.Route{
								Rule: r.Rule,
							}
							routes[routeName] = route
						}

						frontend := types.Frontend{
							Routes: routes,
						}
						frontend.Backend = s.Name
						frontends[s.Name] = &frontend
					}
				}

				configMessage := types.ConfigMessage{
					ProviderName: "servicefabric",
					Configuration: &types.Configuration{
						Backends:  backends,
						Frontends: frontends,
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
