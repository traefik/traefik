package cloudflare

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

// ZoneLockdown represents a Zone Lockdown rule. A rule only permits access to
// the provided URL pattern(s) from the given IP address(es) or subnet(s).
type ZoneLockdown struct {
	ID             string               `json:"id"`
	Description    string               `json:"description"`
	URLs           []string             `json:"urls"`
	Configurations []ZoneLockdownConfig `json:"configurations"`
	Paused         bool                 `json:"paused"`
}

// ZoneLockdownConfig represents a Zone Lockdown config, which comprises
// a Target ("ip" or "ip_range") and a Value (an IP address or IP+mask,
// respectively.)
type ZoneLockdownConfig struct {
	Target string `json:"target"`
	Value  string `json:"value"`
}

// ZoneLockdownResponse represents a response from the Zone Lockdown endpoint.
type ZoneLockdownResponse struct {
	Result ZoneLockdown `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// ZoneLockdownListResponse represents a response from the List Zone Lockdown
// endpoint.
type ZoneLockdownListResponse struct {
	Result []ZoneLockdown `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// CreateZoneLockdown creates a Zone ZoneLockdown rule for the given zone ID.
//
// API reference: https://api.cloudflare.com/#zone-ZoneLockdown-create-a-ZoneLockdown-rule
func (api *API) CreateZoneLockdown(zoneID string, ld ZoneLockdown) (*ZoneLockdownResponse, error) {
	uri := "/zones/" + zoneID + "/firewall/lockdowns"
	res, err := api.makeRequest("POST", uri, ld)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &ZoneLockdownResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}

// UpdateZoneLockdown updates a Zone ZoneLockdown rule (based on the ID) for the
// given zone ID.
//
// API reference: https://api.cloudflare.com/#zone-ZoneLockdown-update-ZoneLockdown-rule
func (api *API) UpdateZoneLockdown(zoneID string, id string, ld ZoneLockdown) (*ZoneLockdownResponse, error) {
	uri := "/zones/" + zoneID + "/firewall/lockdowns"
	res, err := api.makeRequest("PUT", uri, ld)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &ZoneLockdownResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}

// DeleteZoneLockdown deletes a Zone ZoneLockdown rule (based on the ID) for the
// given zone ID.
//
// API reference: https://api.cloudflare.com/#zone-ZoneLockdown-delete-ZoneLockdown-rule
func (api *API) DeleteZoneLockdown(zoneID string, id string) (*ZoneLockdownResponse, error) {
	uri := "/zones/" + zoneID + "/firewall/lockdowns/" + id
	res, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &ZoneLockdownResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}

// ZoneLockdown retrieves a Zone ZoneLockdown rule (based on the ID) for the
// given zone ID.
//
// API reference: https://api.cloudflare.com/#zone-ZoneLockdown-ZoneLockdown-rule-details
func (api *API) ZoneLockdown(zoneID string, id string) (*ZoneLockdownResponse, error) {
	uri := "/zones/" + zoneID + "/firewall/lockdowns/" + id
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &ZoneLockdownResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}

// ListZoneLockdowns retrieves a list of Zone ZoneLockdown rules for a given
// zone ID by page number.
//
// API reference: https://api.cloudflare.com/#zone-ZoneLockdown-list-ZoneLockdown-rules
func (api *API) ListZoneLockdowns(zoneID string, page int) (*ZoneLockdownListResponse, error) {
	v := url.Values{}
	if page <= 0 {
		page = 1
	}

	v.Set("page", strconv.Itoa(page))
	v.Set("per_page", strconv.Itoa(100))
	query := "?" + v.Encode()

	uri := "/zones/" + zoneID + "/firewall/lockdowns" + query
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &ZoneLockdownListResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}
