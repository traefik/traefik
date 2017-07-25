// Package osimage provides a client for Operating System Images.
package osimage

import (
	"encoding/xml"

	"github.com/Azure/azure-sdk-for-go/management"
)

const (
	azureImageListURL    = "services/images"
	errInvalidImage      = "Can not find image %s in specified subscription, please specify another image name."
	errParamNotSpecified = "Parameter %s is not specified."
)

// NewClient is used to instantiate a new OSImageClient from an Azure client.
func NewClient(client management.Client) OSImageClient {
	return OSImageClient{client: client}
}

func (c OSImageClient) ListOSImages() (ListOSImagesResponse, error) {
	var l ListOSImagesResponse

	response, err := c.client.SendAzureGetRequest(azureImageListURL)
	if err != nil {
		return l, err
	}

	err = xml.Unmarshal(response, &l)
	return l, err
}

// AddOSImage adds an operating system image to the image repository that is associated with the specified subscription.
//
// See https://msdn.microsoft.com/en-us/library/azure/jj157192.aspx for details.
func (c OSImageClient) AddOSImage(osi *OSImage) (management.OperationID, error) {
	data, err := xml.Marshal(osi)
	if err != nil {
		return "", err
	}

	return c.client.SendAzurePostRequest(azureImageListURL, data)

}
