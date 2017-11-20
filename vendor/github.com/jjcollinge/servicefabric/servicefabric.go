// Package servicefabric is an opinionated Service Fabric client written in Golang
package servicefabric

import (
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cloudflare/cfssl/log"
)

// Client is an interface for Service Fabric client's
// to implement. This is purposely a subset of the total
// Service Fabric API surface.
type Client interface {
	GetApplications() (*ApplicationItemsPage, error)
	GetServices(appName string) (*ServiceItemsPage, error)
	GetPartitions(appName, serviceName string) (*PartitionItemsPage, error)
	GetReplicas(appName, serviceName, partitionName string) (*ReplicaItemsPage, error)
	GetInstances(appName, serviceName, partitionName string) (*InstanceItemsPage, error)
	GetServiceExtension(appType, applicationVersion, serviceTypeName, extensionKey string, response interface{}) error
	GetProperties(name string) (bool, map[string]string, error)
	GetServiceLabels(service *ServiceItem, app *ApplicationItem, prefix string) (map[string]string, error)
}

type clientImpl struct {
	endpoint    string     `description:"Service Fabric cluster management endpoint"`
	http        HTTPClient `description:"Reusable HTTP client"`
	apiVersion  string     `description:"Service Fabric API version"`
	clusterName string     `description:"Service Fabric cluster name"`
}

const defaultAPIVersion string = "3.0"

// NewClient returns a new provider client that can query the
// Service Fabric management API externally or internally
func NewClient(httpClient HTTPClient, endpoint, apiVersion, clientCertFilePath, clientCertKeyFilePath string, insecureSkipVerify bool) (Client, error) {
	if endpoint == "" {
		return nil, errors.New("endpoint missing for client configuration")
	}
	if apiVersion == "" {
		apiVersion = defaultAPIVersion
	}

	client := &clientImpl{
		endpoint:   endpoint,
		apiVersion: apiVersion,
	}

	if clientCertFilePath != "" && clientCertKeyFilePath != "" {
		cert, err := tls.LoadX509KeyPair(clientCertFilePath, clientCertKeyFilePath)
		if err != nil {
			return nil, fmt.Errorf("unable to load X509 key pair %v", err)
		}
		tlsConfig := &tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: insecureSkipVerify,
			Renegotiation:      tls.RenegotiateFreelyAsClient,
		}
		tlsConfig.BuildNameToCertificate()
		transport := &http.Transport{TLSClientConfig: tlsConfig}
		httpClient.Transport(transport)
		client.http = httpClient
	} else {
		client.http = httpClient
	}
	return client, nil
}

// GetApplications returns all the registered applications
// within the Service Fabric cluster.
func (c *clientImpl) GetApplications() (*ApplicationItemsPage, error) {
	var aggregateAppItemsPages ApplicationItemsPage
	var continueToken string
	for {
		var url string
		if continueToken == "" {
			url = c.endpoint + "/Applications/?api-version=" + c.apiVersion
		} else {
			url = c.endpoint + "/Applications/?api-version=" + c.apiVersion + "&continue=" + continueToken
		}
		res, err := getHTTP(c.http, url)
		if err != nil {
			return &ApplicationItemsPage{}, err
		}
		var appItemsPage ApplicationItemsPage
		err = json.Unmarshal(res, &appItemsPage)
		if err != nil {
			return nil, fmt.Errorf("could not deserialise JSON response: %+v", err)
		}
		aggregateAppItemsPages.Items = append(aggregateAppItemsPages.Items, appItemsPage.Items...)
		continueToken = getString(appItemsPage.ContinuationToken)
		if continueToken == "" {
			break
		}
	}
	return &aggregateAppItemsPages, nil
}

// GetServices returns all the services associated
// with a Service Fabric application.
func (c *clientImpl) GetServices(appName string) (*ServiceItemsPage, error) {
	var aggregateServiceItemsPages ServiceItemsPage
	var continueToken string
	for {
		var url string
		if continueToken == "" {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices?api-version=" + c.apiVersion
		} else {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices?api-version=" + c.apiVersion + "&continue=" + continueToken
		}
		res, err := getHTTP(c.http, url)
		if err != nil {
			return &ServiceItemsPage{}, err
		}
		var servicesItemsPage ServiceItemsPage
		err = json.Unmarshal(res, &servicesItemsPage)
		if err != nil {
			return nil, fmt.Errorf("could not deserialise JSON response: %+v", err)
		}
		aggregateServiceItemsPages.Items = append(aggregateServiceItemsPages.Items, servicesItemsPage.Items...)
		continueToken = getString(servicesItemsPage.ContinuationToken)
		if continueToken == "" {
			break
		}
	}
	return &aggregateServiceItemsPages, nil
}

// GetPartitions returns all the partitions associated
// with a Service Fabric service.
func (c *clientImpl) GetPartitions(appName, serviceName string) (*PartitionItemsPage, error) {
	var aggregatePartitionItemsPages PartitionItemsPage
	var continueToken string
	for {
		var url string
		if continueToken == "" {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/?api-version=" + c.apiVersion
		} else {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/?api-version=" + c.apiVersion + "&continue=" + continueToken
		}
		res, err := getHTTP(c.http, url)
		if err != nil {
			return &PartitionItemsPage{}, err
		}
		var partitionsItemsPage PartitionItemsPage
		err = json.Unmarshal(res, &partitionsItemsPage)
		if err != nil {
			return nil, fmt.Errorf("could not deserialise JSON response: %+v", err)
		}
		aggregatePartitionItemsPages.Items = append(aggregatePartitionItemsPages.Items, partitionsItemsPage.Items...)
		continueToken = getString(partitionsItemsPage.ContinuationToken)
		if continueToken == "" {
			break
		}
	}
	return &aggregatePartitionItemsPages, nil
}

// GetInstances returns all the instances associated
// with a stateless Service Fabric partition.
func (c *clientImpl) GetInstances(appName, serviceName, partitionName string) (*InstanceItemsPage, error) {
	var aggregateInstanceItemsPages InstanceItemsPage
	var continueToken string
	for {
		var url string
		if continueToken == "" {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + partitionName + "/$/GetReplicas?api-version=" + c.apiVersion
		} else {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + partitionName + "/$/GetReplicas?api-version=" + c.apiVersion + "&continue=" + continueToken
		}
		res, err := getHTTP(c.http, url)
		if err != nil {
			return &InstanceItemsPage{}, err
		}
		var instanceItemsPage InstanceItemsPage
		err = json.Unmarshal(res, &instanceItemsPage)
		if err != nil {
			return nil, fmt.Errorf("could not deserialise JSON response: %+v", err)
		}
		aggregateInstanceItemsPages.Items = append(aggregateInstanceItemsPages.Items, instanceItemsPage.Items...)
		continueToken = getString(instanceItemsPage.ContinuationToken)
		if continueToken == "" {
			break
		}
	}
	return &aggregateInstanceItemsPages, nil
}

// GetReplicas returns all the replicas associated
// with a stateful Service Fabric partition.
func (c *clientImpl) GetReplicas(appName, serviceName, partitionName string) (*ReplicaItemsPage, error) {
	var aggregateReplicaItemsPages ReplicaItemsPage
	var continueToken string
	for {
		var url string
		if continueToken == "" {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + partitionName + "/$/GetReplicas?api-version=" + c.apiVersion
		} else {
			url = c.endpoint + "/Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + partitionName + "/$/GetReplicas?api-version=" + c.apiVersion + "&continue=" + continueToken
		}
		res, err := getHTTP(c.http, url)
		if err != nil {
			return &ReplicaItemsPage{}, err
		}
		var replicasItemsPage ReplicaItemsPage
		err = json.Unmarshal(res, &replicasItemsPage)
		if err != nil {
			return nil, fmt.Errorf("could not deserialise JSON response: %+v", err)
		}
		aggregateReplicaItemsPages.Items = append(aggregateReplicaItemsPages.Items, replicasItemsPage.Items...)
		continueToken = getString(replicasItemsPage.ContinuationToken)
		if continueToken == "" {
			break
		}
	}
	return &aggregateReplicaItemsPages, nil
}

// GetServicesExtensions retruns all the extensions specified
// in a Service's manifest file. If the XML schema does not
// map to the provided interface, the default type interface will
// be returned.
func (c *clientImpl) GetServiceExtension(appType, applicationVersion, serviceTypeName, extensionKey string, response interface{}) error {
	url := c.endpoint + "/ApplicationTypes/" + appType + "/$/GetServiceTypes?api-version=" + c.apiVersion + "&ApplicationTypeVersion=" + applicationVersion
	res, err := getHTTP(c.http, url)

	if err != nil {
		return fmt.Errorf("error requesting service extensions: %v", err)
	}

	var serviceTypes []ServiceType
	err = json.Unmarshal(res, &serviceTypes)

	if err != nil {
		return fmt.Errorf("could not deserialise JSON response: %+v", err)
	}

	for _, serviceTypeInfo := range serviceTypes {
		if serviceTypeInfo.ServiceTypeDescription.ServiceTypeName == serviceTypeName {
			for _, extension := range serviceTypeInfo.ServiceTypeDescription.Extensions {
				if strings.EqualFold(extension.Key, extensionKey) {
					xmlData := extension.Value
					err = xml.Unmarshal([]byte(xmlData), &response)
					if err != nil {
						return fmt.Errorf("could not deserialise extension's XML value: %+v", err)
					}
					return nil
				}
			}
		}
	}
	return nil
}

// GetProperties uses the Property Manager API to retreive
// string properties from a name as a dictionary
func (c *clientImpl) GetProperties(name string) (bool, map[string]string, error) {
	nameExists, err := c.nameExists(name)
	if err != nil {
		return false, nil, err
	}
	if !nameExists {
		return false, nil, nil
	}

	properties := make(map[string]string)

	var continueToken string
	baseURL := c.endpoint + "/Names/" + name + "/$/GetProperties?api-version=" + c.apiVersion + "&IncludeValues=true"
	for {
		var url string
		if continueToken == "" {
			url = baseURL
		} else {
			url = baseURL + "&continue=" + continueToken
		}
		res, err := getHTTP(c.http, url)
		if err != nil {
			return false, nil, err
		}
		var propertiesListPage PropertiesListPage
		err = json.Unmarshal(res, &propertiesListPage)
		if err != nil {
			return false, nil, fmt.Errorf("could not deserialise JSON response: %+v", err)
		}

		for _, property := range propertiesListPage.Properties {
			if property.Value.Kind != "String" {
				continue
			}
			properties[property.Name] = property.Value.Data
		}

		if continueToken == "" {
			break
		}
	}

	return true, properties, nil
}

// GetServiceLabels add labels from service manifest extensions and properties manager
// expects extension xml in <Label key="key">value</Label>
func (c *clientImpl) GetServiceLabels(service *ServiceItem, app *ApplicationItem, prefix string) (map[string]string, error) {
	Labels := map[string]string{}
	extensionData := ServiceExtensionLabels{}
	err := c.GetServiceExtension(app.TypeName, app.TypeVersion, service.TypeName, prefix, &extensionData)

	if err != nil {
		return nil, err
	}

	prefixPeriod := prefix + "."

	if extensionData.Label != nil {
		for _, label := range extensionData.Label {
			if strings.HasPrefix(label.Key, prefixPeriod) {
				labelKey := strings.Replace(label.Key, prefixPeriod, "", -1)
				Labels[labelKey] = label.Value
			}
		}
	}

	exists, properties, err := c.GetProperties(service.ID)
	if err != nil {
		log.Error(err)
	} else {
		if exists {
			for k, v := range properties {
				if strings.HasPrefix(k, prefixPeriod) {
					labelKey := strings.Replace(k, prefixPeriod, "", -1)
					Labels[labelKey] = v
				}
			}
		}
	}

	return Labels, nil
}

func (c *clientImpl) nameExists(propertyName string) (bool, error) {
	url := c.endpoint + "/Names/" + propertyName + "?api-version=" + c.apiVersion
	res, err := getHTTPRaw(c.http, url)

	//Get http will return error for any non 200 response code.
	if err != nil {
		return false, fmt.Errorf("Service Fabric responded with error code %s to request %s with body %s", res.Status, url, res.Body)
	}
	if res.StatusCode == 200 {
		return true, nil
	}
	return false, nil
}

func getHTTP(http HTTPClient, url string) ([]byte, error) {
	if http == nil {
		return nil, fmt.Errorf("invalid http client provided")
	}
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Service Fabric server %+v on %s", err, url)
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Service Fabric responded with error code %s to request %s with body %s", res.Status, url, res.Body)
	}
	defer res.Body.Close()
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body from Service Fabric response %+v", readErr)
	}
	return body, nil
}

func getHTTPRaw(http HTTPClient, url string) (*http.Response, error) {
	if http == nil {
		return nil, fmt.Errorf("invalid http client provided")
	}
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Service Fabric server %+v on %s", err, url)
	}
	return res, nil
}

func getString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}
