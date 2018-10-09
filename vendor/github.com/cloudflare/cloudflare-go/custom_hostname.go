package cloudflare

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

// CustomHostnameSSL represents the SSL section in a given custom hostname.
type CustomHostnameSSL struct {
	Status      string `json:"status,omitempty"`
	Method      string `json:"method,omitempty"`
	Type        string `json:"type,omitempty"`
	CnameTarget string `json:"cname_target,omitempty"`
	CnameName   string `json:"cname_name,omitempty"`
}

// CustomMetadata defines custom metadata for the hostname. This requires logic to be implemented by Cloudflare to act on the data provided.
type CustomMetadata map[string]interface{}

// CustomHostname represents a custom hostname in a zone.
type CustomHostname struct {
	ID             string            `json:"id,omitempty"`
	Hostname       string            `json:"hostname,omitempty"`
	SSL            CustomHostnameSSL `json:"ssl,omitempty"`
	CustomMetadata CustomMetadata    `json:"custom_metadata,omitempty"`
}

// CustomHostNameResponse represents a response from the Custom Hostnames endpoints.
type CustomHostnameResponse struct {
	Result CustomHostname `json:"result"`
	Response
}

// CustomHostnameListResponse represents a response from the Custom Hostnames endpoints.
type CustomHostnameListResponse struct {
	Result []CustomHostname `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// Modify SSL configuration for the given custom hostname in the given zone.
//
// API reference: https://api.cloudflare.com/#custom-hostname-for-a-zone-update-custom-hostname-configuration
func (api *API) UpdateCustomHostnameSSL(zoneID string, customHostnameID string, ssl CustomHostnameSSL) (CustomHostname, error) {
	return CustomHostname{}, errors.New("Not implemented")
}

// Delete a custom hostname (and any issued SSL certificates)
//
// API reference: https://api.cloudflare.com/#custom-hostname-for-a-zone-delete-a-custom-hostname-and-any-issued-ssl-certificates-
func (api *API) DeleteCustomHostname(zoneID string, customHostnameID string) error {
	uri := "/zones/" + zoneID + "/custom_hostnames/" + customHostnameID
	res, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}

	var response *CustomHostnameResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		return errors.Wrap(err, errUnmarshalError)
	}

	return nil
}

// CreateCustomHostname creates a new custom hostname and requests that an SSL certificate be issued for it.
//
// API reference: https://api.cloudflare.com/#custom-hostname-for-a-zone-create-custom-hostname
func (api *API) CreateCustomHostname(zoneID string, ch CustomHostname) (*CustomHostnameResponse, error) {
	uri := "/zones/" + zoneID + "/custom_hostnames"
	res, err := api.makeRequest("POST", uri, ch)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	var response *CustomHostnameResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}

// CustomHostnames fetches custom hostnames for the given zone,
// by applying filter.Hostname if not empty and scoping the result to page'th 50 items.
//
// The returned ResultInfo can be used to implement pagination.
//
// API reference: https://api.cloudflare.com/#custom-hostname-for-a-zone-list-custom-hostnames
func (api *API) CustomHostnames(zoneID string, page int, filter CustomHostname) ([]CustomHostname, ResultInfo, error) {
	v := url.Values{}
	v.Set("per_page", "50")
	v.Set("page", strconv.Itoa(page))
	if filter.Hostname != "" {
		v.Set("hostname", filter.Hostname)
	}
	query := "?" + v.Encode()

	uri := "/zones/" + zoneID + "/custom_hostnames" + query
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return []CustomHostname{}, ResultInfo{}, errors.Wrap(err, errMakeRequestError)
	}
	var customHostnameListResponse CustomHostnameListResponse
	err = json.Unmarshal(res, &customHostnameListResponse)
	if err != nil {
		return []CustomHostname{}, ResultInfo{}, errors.Wrap(err, errMakeRequestError)
	}

	return customHostnameListResponse.Result, customHostnameListResponse.ResultInfo, nil
}

// CustomHostname inspects the given custom hostname in the given zone.
//
// API reference: https://api.cloudflare.com/#custom-hostname-for-a-zone-custom-hostname-configuration-details
func (api *API) CustomHostname(zoneID string, customHostnameID string) (CustomHostname, error) {
	uri := "/zones/" + zoneID + "/custom_hostnames/" + customHostnameID
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return CustomHostname{}, errors.Wrap(err, errMakeRequestError)
	}

	var response CustomHostnameResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		return CustomHostname{}, errors.Wrap(err, errUnmarshalError)
	}

	return response.Result, nil
}

// CustomHostnameIDByName retrieves the ID for the given hostname in the given zone.
func (api *API) CustomHostnameIDByName(zoneID string, hostname string) (string, error) {
	customHostnames, _, err := api.CustomHostnames(zoneID, 1, CustomHostname{Hostname: hostname})
	if err != nil {
		return "", errors.Wrap(err, "CustomHostnames command failed")
	}
	for _, ch := range customHostnames {
		if ch.Hostname == hostname {
			return ch.ID, nil
		}
	}
	return "", errors.New("CustomHostname could not be found")
}
