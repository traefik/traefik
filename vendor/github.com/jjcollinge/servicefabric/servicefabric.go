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
)

// DefaultAPIVersion is a default Service Fabric REST API version
const DefaultAPIVersion = "3.0"

// Client for Service Fabric.
// This is purposely a subset of the total Service Fabric API surface.
type Client struct {
	// endpoint Service Fabric cluster management endpoint
	endpoint string
	// apiVersion Service Fabric API version
	apiVersion string
	// httpClient HTTP client
	httpClient *http.Client
}

// NewClient returns a new provider client that can query the
// Service Fabric management API externally or internally
func NewClient(httpClient *http.Client, endpoint, apiVersion string, tlsConfig *tls.Config) (*Client, error) {
	if endpoint == "" {
		return nil, errors.New("endpoint missing for httpClient configuration")
	}
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}

	if tlsConfig != nil {
		tlsConfig.Renegotiation = tls.RenegotiateFreelyAsClient
		tlsConfig.BuildNameToCertificate()
		httpClient.Transport = &http.Transport{TLSClientConfig: tlsConfig}
	}

	return &Client{
		endpoint:   endpoint,
		apiVersion: apiVersion,
		httpClient: httpClient,
	}, nil
}

// GetApplications returns all the registered applications
// within the Service Fabric cluster.
func (c Client) GetApplications() (*ApplicationItemsPage, error) {
	var aggregateAppItemsPages ApplicationItemsPage
	var continueToken string
	for {
		res, err := c.getHTTP("Applications/", withContinue(continueToken))
		if err != nil {
			return nil, err
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
func (c Client) GetServices(appName string) (*ServiceItemsPage, error) {
	var aggregateServiceItemsPages ServiceItemsPage
	var continueToken string
	for {
		res, err := c.getHTTP("Applications/"+appName+"/$/GetServices", withContinue(continueToken))
		if err != nil {
			return nil, err
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
func (c Client) GetPartitions(appName, serviceName string) (*PartitionItemsPage, error) {
	var aggregatePartitionItemsPages PartitionItemsPage
	var continueToken string
	for {
		basePath := "Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/"
		res, err := c.getHTTP(basePath, withContinue(continueToken))
		if err != nil {
			return nil, err
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
func (c Client) GetInstances(appName, serviceName, partitionName string) (*InstanceItemsPage, error) {
	var aggregateInstanceItemsPages InstanceItemsPage
	var continueToken string
	for {
		basePath := "Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + partitionName + "/$/GetReplicas"
		res, err := c.getHTTP(basePath, withContinue(continueToken))
		if err != nil {
			return nil, err
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
func (c Client) GetReplicas(appName, serviceName, partitionName string) (*ReplicaItemsPage, error) {
	var aggregateReplicaItemsPages ReplicaItemsPage
	var continueToken string
	for {
		basePath := "Applications/" + appName + "/$/GetServices/" + serviceName + "/$/GetPartitions/" + partitionName + "/$/GetReplicas"
		res, err := c.getHTTP(basePath, withContinue(continueToken))
		if err != nil {
			return nil, err
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

// GetServiceExtension returns all the extensions specified
// in a Service's manifest file. If the XML schema does not
// map to the provided interface, the default type interface will
// be returned.
func (c Client) GetServiceExtension(appType, applicationVersion, serviceTypeName, extensionKey string, response interface{}) error {
	res, err := c.getHTTP("ApplicationTypes/"+appType+"/$/GetServiceTypes", withParam("ApplicationTypeVersion", applicationVersion))
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
					err = xml.Unmarshal([]byte(extension.Value), &response)
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

// GetServiceExtensionMap returns all the extension xml specified
// in a Service's manifest file into (which must conform to ServiceExtensionLabels)
// a map[string]string
func (c Client) GetServiceExtensionMap(service *ServiceItem, app *ApplicationItem, extensionKey string) (map[string]string, error) {
	extensionData := ServiceExtensionLabels{}
	err := c.GetServiceExtension(app.TypeName, app.TypeVersion, service.TypeName, extensionKey, &extensionData)
	if err != nil {
		return nil, err
	}

	labels := map[string]string{}
	if extensionData.Label != nil {
		for _, label := range extensionData.Label {
			labels[label.Key] = label.Value
		}
	}

	return labels, nil
}

// GetProperties uses the Property Manager API to retrieve
// string properties from a name as a dictionary
// Property name is the path to the properties you would like to list.
// for example a serviceID
func (c Client) GetProperties(name string) (bool, map[string]string, error) {
	nameExists, err := c.nameExists(name)
	if err != nil {
		return false, nil, err
	}

	if !nameExists {
		return false, nil, nil
	}

	properties := make(map[string]string)

	var continueToken string
	for {
		res, err := c.getHTTP("Names/"+name+"/$/GetProperties", withContinue(continueToken), withParam("IncludeValues", "true"))
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

		continueToken = propertiesListPage.ContinuationToken
		if continueToken == "" {
			break
		}
	}

	return true, properties, nil
}

// GetServiceLabels add labels from service manifest extensions and properties manager
// expects extension xml in <Label key="key">value</Label>
//
// Deprecated: Use GetProperties and GetServiceExtensionMap instead.
func (c Client) GetServiceLabels(service *ServiceItem, app *ApplicationItem, prefix string) (map[string]string, error) {
	extensionData := ServiceExtensionLabels{}
	err := c.GetServiceExtension(app.TypeName, app.TypeVersion, service.TypeName, prefix, &extensionData)
	if err != nil {
		return nil, err
	}

	prefixPeriod := prefix + "."

	labels := map[string]string{}
	if extensionData.Label != nil {
		for _, label := range extensionData.Label {
			if strings.HasPrefix(label.Key, prefixPeriod) {
				labelKey := strings.Replace(label.Key, prefixPeriod, "", -1)
				labels[labelKey] = label.Value
			}
		}
	}

	exists, properties, err := c.GetProperties(service.ID)
	if err != nil {
		return nil, err
	}

	if exists {
		for k, v := range properties {
			if strings.HasPrefix(k, prefixPeriod) {
				labelKey := strings.Replace(k, prefixPeriod, "", -1)
				labels[labelKey] = v
			}
		}
	}

	return labels, nil
}

func (c Client) nameExists(propertyName string) (bool, error) {
	res, err := c.getHTTPRaw("Names/" + propertyName)
	// Get http will return error for any non 200 response code.
	if err != nil {
		return false, err
	}

	return res.StatusCode == http.StatusOK, nil
}

func (c Client) getHTTP(basePath string, paramsFuncs ...queryParamsFunc) ([]byte, error) {
	if c.httpClient == nil {
		return nil, errors.New("invalid http client provided")
	}

	url := c.getURL(basePath, paramsFuncs...)
	res, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Service Fabric server %+v on %s", err, url)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Service Fabric responded with error code %s to request %s with body %v", res.Status, url, res.Body)
	}

	if res.Body == nil {
		return nil, errors.New("empty response body from Service Fabric")
	}
	defer res.Body.Close()

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body from Service Fabric response %+v", readErr)
	}
	return body, nil
}

func (c Client) getHTTPRaw(basePath string) (*http.Response, error) {
	if c.httpClient == nil {
		return nil, fmt.Errorf("invalid http client provided")
	}

	url := c.getURL(basePath)

	res, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Service Fabric server %+v on %s", err, url)
	}
	return res, nil
}

func (c Client) getURL(basePath string, paramsFuncs ...queryParamsFunc) string {
	params := []string{"api-version=" + c.apiVersion}

	for _, paramsFunc := range paramsFuncs {
		params = paramsFunc(params)
	}

	return fmt.Sprintf("%s/%s?%s", c.endpoint, basePath, strings.Join(params, "&"))
}

func getString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}
