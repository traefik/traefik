// Package location provides a client for Locations.
package location

import (
	"encoding/xml"

	"github.com/Azure/azure-sdk-for-go/management"
)

const (
	azureLocationListURL = "locations"
	errParamNotSpecified = "Parameter %s is not specified."
)

//NewClient is used to instantiate a new LocationClient from an Azure client
func NewClient(client management.Client) LocationClient {
	return LocationClient{client: client}
}

func (c LocationClient) ListLocations() (ListLocationsResponse, error) {
	var l ListLocationsResponse

	response, err := c.client.SendAzureGetRequest(azureLocationListURL)
	if err != nil {
		return l, err
	}

	err = xml.Unmarshal(response, &l)
	return l, err
}
