package servicefabric

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/containous/traefik/log"
)

// TODO:
// - Mocks and tests
// - Investigate paging
// - Standardise error handling

// Client for the Provider to talk to Service Fabric
type Client interface {
	GetApplications() (*ApplicationsData, error)
	GetServices(appName string) (*ServicesData, error)
	GetPartitions(appName, serviceName string) (*PartitionsData, error)
	GetReplicas(appName, serviceName, partitionName string) (*ReplicasData, error)
	GetInstances(appName, serviceName, partitionName string) (*InstancesData, error)
}

type clientImpl struct {
	endpoint   string      `description:"Service Fabric cluster management endpoint"`
	restClient http.Client `description:"Reusable HTTP client"`
	apiVersion string      `description:"Service Fabric API version"`
}

// NewClient returns a new Provider client that can query the
// Service Fabric management API externally or internally
func NewClient(endpoint, apiVersion, clientCertFilePath, clientCertKeyFilePath, caCertFilePath string) (Client, error) {
	if endpoint == "" {
		return nil, errors.New("endpoint missing for client configuration")
	}

	client := &clientImpl{
		endpoint:   endpoint,
		apiVersion: apiVersion,
	}

	if caCertFilePath != "" {
		cert, err := tls.LoadX509KeyPair(clientCertFilePath, clientCertKeyFilePath)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		caCert, err := ioutil.ReadFile(caCertFilePath)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			RootCAs:            caCertPool,
			InsecureSkipVerify: true,
			Renegotiation:      tls.RenegotiateFreelyAsClient,
		}
		tlsConfig.BuildNameToCertificate()
		transport := &http.Transport{TLSClientConfig: tlsConfig}

		client.restClient = http.Client{Transport: transport}
	} else {
		client.restClient = http.Client{}
	}

	return client, nil
}

// GetApplications returns all the registered applications
// within the Service Fabric cluster.
func (c *clientImpl) GetApplications() (*ApplicationsData, error) {
	url := c.endpoint + "/Applications/?api-version=" + c.apiVersion
	res, err := getHTTP(&c.restClient, url)
	if err != nil {
		return &ApplicationsData{}, err
	}
	var appData ApplicationsData
	err = json.Unmarshal(res, &appData)
	if err != nil {
		log.Errorf("Could not deserialise JSON response: %+v", err)
	}
	return &appData, nil
}

// GetServices returns all the services associated
// with a Service Fabric application.
func (c *clientImpl) GetServices(appName string) (*ServicesData, error) {
	url := c.endpoint + "/Applications/" + appName + "/$/GetServices?api-version=" + c.apiVersion
	res, err := getHTTP(&c.restClient, url)
	if err != nil {
		return &ServicesData{}, err
	}
	var servicesData ServicesData
	err = json.Unmarshal(res, &servicesData)
	if err != nil {
		log.Errorf("Could not deserialise JSON response: %+v", err)
	}
	return &servicesData, nil
}

// GetPartitions returns all the partitions associated
// with a Service Fabric service.
func (c *clientImpl) GetPartitions(appName, serviceName string) (*PartitionsData, error) {
	url := c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/?api-version=" + c.apiVersion
	res, err := getHTTP(&c.restClient, url)
	if err != nil {
		return &PartitionsData{}, err
	}
	var partitionsData PartitionsData
	err = json.Unmarshal(res, &partitionsData)
	if err != nil {
		log.Errorf("Could not deserialise JSON response: %+v", err)
	}
	return &partitionsData, nil
}

// GetInstances returns all the instances associated
// with a stateless Service Fabric partition.
func (c *clientImpl) GetInstances(appName, serviceName, partitionName string) (*InstancesData, error) {
	url := c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + partitionName + "/$/GetReplicas?api-version=" + c.apiVersion
	res, err := getHTTP(&c.restClient, url)
	if err != nil {
		return &InstancesData{}, err
	}
	var instancesData InstancesData
	err = json.Unmarshal(res, &instancesData)
	if err != nil {
		log.Errorf("Could not deserialise JSON response: %+v", err)
	}
	return &instancesData, nil
}

// GetReplicas returns all the replicas associated
// with a stateful Service Fabric partition.
func (c *clientImpl) GetReplicas(appName, serviceName, partitionName string) (*ReplicasData, error) {
	url := c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + partitionName + "/$/GetReplicas?api-version=" + c.apiVersion
	res, err := getHTTP(&c.restClient, url)
	if err != nil {
		return &ReplicasData{}, err
	}
	var replicasData ReplicasData
	err = json.Unmarshal(res, &replicasData)
	if err != nil {
		log.Errorf("Could not deserialise JSON response: %+v", err)
	}
	return &replicasData, nil
}

func getHTTP(http *http.Client, url string) ([]byte, error) {
	if http == nil {
		return nil, fmt.Errorf("Invalid http client provided")
	}
	log.Debugf("GET: %s", url)
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Service Fabric server %+v on %s", err, url)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Service Fabric responded with error code %s to request %s with body %s", res.Status, url, res.Body)
	}
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, fmt.Errorf("Failed to read response body from Service Fabric response %+v", readErr)
	}
	return body, nil
}
