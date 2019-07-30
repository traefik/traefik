package cloudflare

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// FirewallRule is the struct of the firewall rule.
type FirewallRule struct {
	ID          string      `json:"id,omitempty"`
	Paused      bool        `json:"paused"`
	Description string      `json:"description"`
	Action      string      `json:"action"`
	Priority    interface{} `json:"priority"`
	Filter      Filter      `json:"filter"`
	CreatedOn   time.Time   `json:"created_on,omitempty"`
	ModifiedOn  time.Time   `json:"modified_on,omitempty"`
}

// FirewallRulesDetailResponse is the API response for the firewall
// rules.
type FirewallRulesDetailResponse struct {
	Result     []FirewallRule `json:"result"`
	ResultInfo `json:"result_info"`
	Response
}

// FirewallRuleResponse is the API response that is returned
// for requesting a single firewall rule on a zone.
type FirewallRuleResponse struct {
	Result     FirewallRule `json:"result"`
	ResultInfo `json:"result_info"`
	Response
}

// FirewallRules returns all firewall rules.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/get/#get-all-rules
func (api *API) FirewallRules(zoneID string, pageOpts PaginationOptions) ([]FirewallRule, error) {
	uri := fmt.Sprintf("/zones/%s/firewall/rules", zoneID)
	v := url.Values{}

	if pageOpts.PerPage > 0 {
		v.Set("per_page", strconv.Itoa(pageOpts.PerPage))
	}

	if pageOpts.Page > 0 {
		v.Set("page", strconv.Itoa(pageOpts.Page))
	}

	if len(v) > 0 {
		uri = uri + "?" + v.Encode()
	}

	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return []FirewallRule{}, errors.Wrap(err, errMakeRequestError)
	}

	var firewallDetailResponse FirewallRulesDetailResponse
	err = json.Unmarshal(res, &firewallDetailResponse)
	if err != nil {
		return []FirewallRule{}, errors.Wrap(err, errUnmarshalError)
	}

	return firewallDetailResponse.Result, nil
}

// FirewallRule returns a single firewall rule based on the ID.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/get/#get-by-rule-id
func (api *API) FirewallRule(zoneID, firewallRuleID string) (FirewallRule, error) {
	uri := fmt.Sprintf("/zones/%s/firewall/rules/%s", zoneID, firewallRuleID)

	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return FirewallRule{}, errors.Wrap(err, errMakeRequestError)
	}

	var firewallRuleResponse FirewallRuleResponse
	err = json.Unmarshal(res, &firewallRuleResponse)
	if err != nil {
		return FirewallRule{}, errors.Wrap(err, errUnmarshalError)
	}

	return firewallRuleResponse.Result, nil
}

// CreateFirewallRules creates new firewall rules.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/post/
func (api *API) CreateFirewallRules(zoneID string, firewallRules []FirewallRule) ([]FirewallRule, error) {
	uri := fmt.Sprintf("/zones/%s/firewall/rules", zoneID)

	res, err := api.makeRequest("POST", uri, firewallRules)
	if err != nil {
		return []FirewallRule{}, errors.Wrap(err, errMakeRequestError)
	}

	var firewallRulesDetailResponse FirewallRulesDetailResponse
	err = json.Unmarshal(res, &firewallRulesDetailResponse)
	if err != nil {
		return []FirewallRule{}, errors.Wrap(err, errUnmarshalError)
	}

	return firewallRulesDetailResponse.Result, nil
}

// UpdateFirewallRule updates a single firewall rule.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/put/#update-a-single-rule
func (api *API) UpdateFirewallRule(zoneID string, firewallRule FirewallRule) (FirewallRule, error) {
	if firewallRule.ID == "" {
		return FirewallRule{}, errors.Errorf("firewall rule ID cannot be empty")
	}

	uri := fmt.Sprintf("/zones/%s/firewall/rules/%s", zoneID, firewallRule.ID)

	res, err := api.makeRequest("PUT", uri, firewallRule)
	if err != nil {
		return FirewallRule{}, errors.Wrap(err, errMakeRequestError)
	}

	var firewallRuleResponse FirewallRuleResponse
	err = json.Unmarshal(res, &firewallRuleResponse)
	if err != nil {
		return FirewallRule{}, errors.Wrap(err, errUnmarshalError)
	}

	return firewallRuleResponse.Result, nil
}

// UpdateFirewallRules updates a single firewall rule.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/put/#update-multiple-rules
func (api *API) UpdateFirewallRules(zoneID string, firewallRules []FirewallRule) ([]FirewallRule, error) {
	for _, firewallRule := range firewallRules {
		if firewallRule.ID == "" {
			return []FirewallRule{}, errors.Errorf("firewall ID cannot be empty")
		}
	}

	uri := fmt.Sprintf("/zones/%s/firewall/rules", zoneID)

	res, err := api.makeRequest("PUT", uri, firewallRules)
	if err != nil {
		return []FirewallRule{}, errors.Wrap(err, errMakeRequestError)
	}

	var firewallRulesDetailResponse FirewallRulesDetailResponse
	err = json.Unmarshal(res, &firewallRulesDetailResponse)
	if err != nil {
		return []FirewallRule{}, errors.Wrap(err, errUnmarshalError)
	}

	return firewallRulesDetailResponse.Result, nil
}

// DeleteFirewallRule updates a single firewall rule.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/delete/#delete-a-single-rule
func (api *API) DeleteFirewallRule(zoneID, firewallRuleID string) error {
	if firewallRuleID == "" {
		return errors.Errorf("firewall rule ID cannot be empty")
	}

	uri := fmt.Sprintf("/zones/%s/firewall/rules/%s", zoneID, firewallRuleID)

	_, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}

	return nil
}

// DeleteFirewallRules updates a single firewall rule.
//
// API reference: https://developers.cloudflare.com/firewall/api/cf-firewall-rules/delete/#delete-multiple-rules
func (api *API) DeleteFirewallRules(zoneID string, firewallRuleIDs []string) error {
	ids := strings.Join(firewallRuleIDs, ",")
	uri := fmt.Sprintf("/zones/%s/firewall/rules?id=%s", zoneID, ids)

	_, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}

	return nil
}
