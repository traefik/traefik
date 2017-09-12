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
	endpoint   string    `description:"Service Fabric cluster management endpoint"`
	restClient webClient `description:"Reusable HTTP client"`
	apiVersion string    `description:"Service Fabric API version"`
}

type webClient interface {
	Get(url string) (resp *http.Response, err error)
	SetTransport(transport *http.Transport)
}

type httpWebClient struct {
	client http.Client
}

func (c *httpWebClient) Get(url string) (resp *http.Response, err error) {
	return c.client.Get(url)
}

func (c *httpWebClient) SetTransport(transport *http.Transport) {
	c.client.Transport = transport
}

// NewClient returns a new Provider client that can query the
// Service Fabric management API externally or internally
func NewClient(webClient webClient, endpoint, apiVersion, clientCertFilePath, clientCertKeyFilePath, caCertFilePath string) (Client, error) {
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

		webClient.SetTransport(transport)
		client.restClient = webClient
	} else {
		client.restClient = webClient
	}
	return client, nil
}

// GetApplications returns all the registered applications
// within the Service Fabric cluster.
func (c *clientImpl) GetApplications() (*ApplicationsData, error) {
	var aggregateAppData ApplicationsData
	var continueToken string
	for {
		var url string
		if continueToken == "" {
			url = c.endpoint + "/Applications/?api-version=" + c.apiVersion
		} else {
			url = c.endpoint + "/Applications/?api-version=" + c.apiVersion + "&continue=" + continueToken
		}
		res, err := getHTTP(c.restClient, url)
		if err != nil {
			return &ApplicationsData{}, err
		}
		var appData ApplicationsData
		err = json.Unmarshal(res, &appData)
		if err != nil {
			log.Errorf("Could not deserialise JSON response: %+v", err)
		}
		aggregateAppData.Items = append(aggregateAppData.Items, appData.Items...)
		continueToken = getString(appData.ContinuationToken)
		if continueToken == "" {
			break
		}
	}
	return &aggregateAppData, nil
}

// GetServices returns all the services associated
// with a Service Fabric application.
func (c *clientImpl) GetServices(appName string) (*ServicesData, error) {
	var aggregateServicesData ServicesData
	var continueToken string
	for {
		var url string
		if continueToken == "" {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices?api-version=" + c.apiVersion
		} else {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices?api-version=" + c.apiVersion + "&continue=" + continueToken
		}
		res, err := getHTTP(c.restClient, url)
		if err != nil {
			return &ServicesData{}, err
		}
		var servicesData ServicesData
		err = json.Unmarshal(res, &servicesData)
		if err != nil {
			log.Errorf("Could not deserialise JSON response: %+v", err)
		}
		aggregateServicesData.Items = append(aggregateServicesData.Items, servicesData.Items...)
		continueToken = getString(servicesData.ContinuationToken)
		if continueToken == "" {
			break
		}
	}
	return &aggregateServicesData, nil
}

// GetPartitions returns all the partitions associated
// with a Service Fabric service.
func (c *clientImpl) GetPartitions(appName, serviceName string) (*PartitionsData, error) {
	var aggregatePartitionsData PartitionsData
	var continueToken string
	for {
		var url string
		if continueToken == "" {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/?api-version=" + c.apiVersion
		} else {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/?api-version=" + c.apiVersion + "&continue=" + continueToken
		}
		res, err := getHTTP(c.restClient, url)
		if err != nil {
			return &PartitionsData{}, err
		}
		var partitionsData PartitionsData
		err = json.Unmarshal(res, &partitionsData)
		if err != nil {
			log.Errorf("Could not deserialise JSON response: %+v", err)
		}
		aggregatePartitionsData.Items = append(aggregatePartitionsData.Items, partitionsData.Items...)
		continueToken = getString(partitionsData.ContinuationToken)
		if continueToken == "" {
			break
		}
	}
	return &aggregatePartitionsData, nil
}

// GetInstances returns all the instances associated
// with a stateless Service Fabric partition.
func (c *clientImpl) GetInstances(appName, serviceName, partitionName string) (*InstancesData, error) {
	var aggregateInstancesData InstancesData
	var continueToken string
	for {
		var url string
		if continueToken == "" {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + partitionName + "/$/GetReplicas?api-version=" + c.apiVersion
		} else {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + partitionName + "/$/GetReplicas?api-version=" + c.apiVersion + "&continue=" + continueToken
		}
		res, err := getHTTP(c.restClient, url)
		if err != nil {
			return &InstancesData{}, err
		}
		var instancesData InstancesData
		err = json.Unmarshal(res, &instancesData)
		if err != nil {
			log.Errorf("Could not deserialise JSON response: %+v", err)
		}
		aggregateInstancesData.Items = append(aggregateInstancesData.Items, instancesData.Items...)
		continueToken = getString(instancesData.ContinuationToken)
		if continueToken == "" {
			break
		}
	}
	return &aggregateInstancesData, nil
}

// GetReplicas returns all the replicas associated
// with a stateful Service Fabric partition.
func (c *clientImpl) GetReplicas(appName, serviceName, partitionName string) (*ReplicasData, error) {
	var aggregateReplicasData ReplicasData
	var continueToken string
	for {
		var url string
		if continueToken == "" {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + partitionName + "/$/GetReplicas?api-version=" + c.apiVersion
		} else {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + partitionName + "/$/GetReplicas?api-version=" + c.apiVersion + "&continue=" + continueToken
		}
		res, err := getHTTP(c.restClient, url)
		if err != nil {
			return &ReplicasData{}, err
		}
		var replicasData ReplicasData
		err = json.Unmarshal(res, &replicasData)
		if err != nil {
			log.Errorf("Could not deserialise JSON response: %+v", err)
		}
		aggregateReplicasData.Items = append(aggregateReplicasData.Items, replicasData.Items...)
		continueToken = getString(replicasData.ContinuationToken)
		if continueToken == "" {
			break
		}
	}
	return &aggregateReplicasData, nil
}

func getHTTP(http webClient, url string) ([]byte, error) {
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

func getString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}
