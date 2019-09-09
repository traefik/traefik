package cloudflare

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// WAFPackage represents a WAF package configuration.
type WAFPackage struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	ZoneID        string `json:"zone_id"`
	DetectionMode string `json:"detection_mode"`
	Sensitivity   string `json:"sensitivity"`
	ActionMode    string `json:"action_mode"`
}

// WAFPackagesResponse represents the response from the WAF packages endpoint.
type WAFPackagesResponse struct {
	Response
	Result     []WAFPackage `json:"result"`
	ResultInfo ResultInfo   `json:"result_info"`
}

// WAFRule represents a WAF rule.
type WAFRule struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	PackageID   string `json:"package_id"`
	Group       struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"group"`
	Mode         string   `json:"mode"`
	DefaultMode  string   `json:"default_mode"`
	AllowedModes []string `json:"allowed_modes"`
}

// WAFRulesResponse represents the response from the WAF rules endpoint.
type WAFRulesResponse struct {
	Response
	Result     []WAFRule  `json:"result"`
	ResultInfo ResultInfo `json:"result_info"`
}

// WAFRuleResponse represents the response from the WAF rule endpoint.
type WAFRuleResponse struct {
	Response
	Result     WAFRule    `json:"result"`
	ResultInfo ResultInfo `json:"result_info"`
}

// WAFRuleOptions is a subset of WAFRule, for editable options.
type WAFRuleOptions struct {
	Mode string `json:"mode"`
}

// ListWAFPackages returns a slice of the WAF packages for the given zone.
//
// API Reference: https://api.cloudflare.com/#waf-rule-packages-list-firewall-packages
func (api *API) ListWAFPackages(zoneID string) ([]WAFPackage, error) {
	var p WAFPackagesResponse
	var packages []WAFPackage
	var res []byte
	var err error
	uri := "/zones/" + zoneID + "/firewall/waf/packages"
	res, err = api.makeRequest("GET", uri, nil)
	if err != nil {
		return []WAFPackage{}, errors.Wrap(err, errMakeRequestError)
	}
	err = json.Unmarshal(res, &p)
	if err != nil {
		return []WAFPackage{}, errors.Wrap(err, errUnmarshalError)
	}
	if !p.Success {
		// TODO: Provide an actual error message instead of always returning nil
		return []WAFPackage{}, err
	}
	for pi := range p.Result {
		packages = append(packages, p.Result[pi])
	}
	return packages, nil
}

// ListWAFRules returns a slice of the WAF rules for the given WAF package.
//
// API Reference: https://api.cloudflare.com/#waf-rules-list-rules
func (api *API) ListWAFRules(zoneID, packageID string) ([]WAFRule, error) {
	var rules []WAFRule
	var res []byte
	var err error

	uri := "/zones/" + zoneID + "/firewall/waf/packages/" + packageID + "/rules"
	res, err = api.makeRequest("GET", uri, nil)
	if err != nil {
		return []WAFRule{}, errors.Wrap(err, errMakeRequestError)
	}

	var r WAFRulesResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return []WAFRule{}, errors.Wrap(err, errUnmarshalError)
	}

	if !r.Success {
		// TODO: Provide an actual error message instead of always returning nil
		return []WAFRule{}, err
	}

	for ri := range r.Result {
		rules = append(rules, r.Result[ri])
	}
	return rules, nil
}

// WAFRule returns a WAF rule from the given WAF package.
//
// API Reference: https://api.cloudflare.com/#waf-rules-rule-details
func (api *API) WAFRule(zoneID, packageID, ruleID string) (WAFRule, error) {
	uri := "/zones/" + zoneID + "/firewall/waf/packages/" + packageID + "/rules/" + ruleID
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return WAFRule{}, errors.Wrap(err, errMakeRequestError)
	}

	var r WAFRuleResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return WAFRule{}, errors.Wrap(err, errUnmarshalError)
	}

	return r.Result, nil
}

// UpdateWAFRule lets you update the mode of a WAF Rule.
//
// API Reference: https://api.cloudflare.com/#waf-rules-edit-rule
func (api *API) UpdateWAFRule(zoneID, packageID, ruleID, mode string) (WAFRule, error) {
	opts := WAFRuleOptions{Mode: mode}
	uri := "/zones/" + zoneID + "/firewall/waf/packages/" + packageID + "/rules/" + ruleID
	res, err := api.makeRequest("PATCH", uri, opts)
	if err != nil {
		return WAFRule{}, errors.Wrap(err, errMakeRequestError)
	}

	var r WAFRuleResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return WAFRule{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}
