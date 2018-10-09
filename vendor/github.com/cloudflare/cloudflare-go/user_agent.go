package cloudflare

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

// UserAgentRule represents a User-Agent Block. These rules can be used to
// challenge, block or whitelist specific User-Agents for a given zone.
type UserAgentRule struct {
	ID            string              `json:"id"`
	Description   string              `json:"description"`
	Mode          string              `json:"mode"`
	Configuration UserAgentRuleConfig `json:"configuration"`
	Paused        bool                `json:"paused"`
}

// UserAgentRuleConfig represents a Zone Lockdown config, which comprises
// a Target ("ip" or "ip_range") and a Value (an IP address or IP+mask,
// respectively.)
type UserAgentRuleConfig ZoneLockdownConfig

// UserAgentRuleResponse represents a response from the Zone Lockdown endpoint.
type UserAgentRuleResponse struct {
	Result UserAgentRule `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// UserAgentRuleListResponse represents a response from the List Zone Lockdown endpoint.
type UserAgentRuleListResponse struct {
	Result []UserAgentRule `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// CreateUserAgentRule creates a User-Agent Block rule for the given zone ID.
//
// API reference: https://api.cloudflare.com/#user-agent-blocking-rules-create-a-useragent-rule
func (api *API) CreateUserAgentRule(zoneID string, ld UserAgentRule) (*UserAgentRuleResponse, error) {
	switch ld.Mode {
	case "block", "challenge", "js_challenge", "whitelist":
		break
	default:
		return nil, errors.New(`the User-Agent Block rule mode must be one of "block", "challenge", "js_challenge", "whitelist"`)
	}

	uri := "/zones/" + zoneID + "/firewall/ua_rules"
	res, err := api.makeRequest("POST", uri, ld)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &UserAgentRuleResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}

// UpdateUserAgentRule updates a User-Agent Block rule (based on the ID) for the given zone ID.
//
// API reference: https://api.cloudflare.com/#user-agent-blocking-rules-update-useragent-rule
func (api *API) UpdateUserAgentRule(zoneID string, id string, ld UserAgentRule) (*UserAgentRuleResponse, error) {
	uri := "/zones/" + zoneID + "/firewall/ua_rules/" + id
	res, err := api.makeRequest("PUT", uri, ld)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &UserAgentRuleResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}

// DeleteUserAgentRule deletes a User-Agent Block rule (based on the ID) for the given zone ID.
//
// API reference: https://api.cloudflare.com/#user-agent-blocking-rules-delete-useragent-rule
func (api *API) DeleteUserAgentRule(zoneID string, id string) (*UserAgentRuleResponse, error) {
	uri := "/zones/" + zoneID + "/firewall/ua_rules/" + id
	res, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &UserAgentRuleResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}

// UserAgentRule retrieves a User-Agent Block rule (based on the ID) for the given zone ID.
//
// API reference: https://api.cloudflare.com/#user-agent-blocking-rules-useragent-rule-details
func (api *API) UserAgentRule(zoneID string, id string) (*UserAgentRuleResponse, error) {
	uri := "/zones/" + zoneID + "/firewall/ua_rules/" + id
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &UserAgentRuleResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}

// ListUserAgentRules retrieves a list of User-Agent Block rules for a given zone ID by page number.
//
// API reference: https://api.cloudflare.com/#user-agent-blocking-rules-list-useragent-rules
func (api *API) ListUserAgentRules(zoneID string, page int) (*UserAgentRuleListResponse, error) {
	v := url.Values{}
	if page <= 0 {
		page = 1
	}

	v.Set("page", strconv.Itoa(page))
	v.Set("per_page", strconv.Itoa(100))
	query := "?" + v.Encode()

	uri := "/zones/" + zoneID + "/firewall/ua_rules" + query
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &UserAgentRuleListResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}
