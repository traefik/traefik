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

// WAFRulesResponse represents the response from the WAF rule endpoint.
type WAFRulesResponse struct {
	Response
	Result     []WAFRule  `json:"result"`
	ResultInfo ResultInfo `json:"result_info"`
}

// ListWAFPackages returns a slice of the WAF packages for the given zone.
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
func (api *API) ListWAFRules(zoneID, packageID string) ([]WAFRule, error) {
	var r WAFRulesResponse
	var rules []WAFRule
	var res []byte
	var err error
	uri := "/zones/" + zoneID + "/firewall/waf/packages/" + packageID + "/rules"
	res, err = api.makeRequest("GET", uri, nil)
	if err != nil {
		return []WAFRule{}, errors.Wrap(err, errMakeRequestError)
	}
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
