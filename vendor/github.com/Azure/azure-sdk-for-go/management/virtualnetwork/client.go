// Package virtualnetwork provides a client for Virtual Networks.
package virtualnetwork

import (
	"encoding/xml"

	"github.com/Azure/azure-sdk-for-go/management"
)

const (
	azureNetworkConfigurationURL = "services/networking/media"
)

// NewClient is used to return new VirtualNetworkClient instance
func NewClient(client management.Client) VirtualNetworkClient {
	return VirtualNetworkClient{client: client}
}

// GetVirtualNetworkConfiguration retreives the current virtual network
// configuration for the currently active subscription. Note that the
// underlying Azure API means that network related operations are not safe
// for running concurrently.
func (c VirtualNetworkClient) GetVirtualNetworkConfiguration() (NetworkConfiguration, error) {
	networkConfiguration := c.NewNetworkConfiguration()
	response, err := c.client.SendAzureGetRequest(azureNetworkConfigurationURL)
	if err != nil {
		return networkConfiguration, err
	}

	err = xml.Unmarshal(response, &networkConfiguration)
	return networkConfiguration, err

}

// SetVirtualNetworkConfiguration configures the virtual networks for the
// currently active subscription according to the NetworkConfiguration given.
// Note that the underlying Azure API means that network related operations
// are not safe for running concurrently.
func (c VirtualNetworkClient) SetVirtualNetworkConfiguration(networkConfiguration NetworkConfiguration) (management.OperationID, error) {
	networkConfiguration.setXMLNamespaces()
	networkConfigurationBytes, err := xml.Marshal(networkConfiguration)
	if err != nil {
		return "", err
	}

	return c.client.SendAzurePutRequest(azureNetworkConfigurationURL, "text/plain", networkConfigurationBytes)
}
