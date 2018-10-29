package cloudflare

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

// PageRuleTarget is the target to evaluate on a request.
//
// Currently Target must always be "url" and Operator must be "matches". Value
// is the URL pattern to match against.
type PageRuleTarget struct {
	Target     string `json:"target"`
	Constraint struct {
		Operator string `json:"operator"`
		Value    string `json:"value"`
	} `json:"constraint"`
}

/*
PageRuleAction is the action to take when the target is matched.

Valid IDs are:

  always_online
  always_use_https
  browser_cache_ttl
  browser_check
  cache_level
  disable_apps
  disable_performance
  disable_railgun
  disable_security
  edge_cache_ttl
  email_obfuscation
  forwarding_url
  ip_geolocation
  mirage
  rocket_loader
  security_level
  server_side_exclude
  smart_errors
  ssl
  waf
*/
type PageRuleAction struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

// PageRuleActions maps API action IDs to human-readable strings.
var PageRuleActions = map[string]string{
	"always_online":       "Always Online",            // Value of type string
	"always_use_https":    "Always Use HTTPS",         // Value of type interface{}
	"browser_cache_ttl":   "Browser Cache TTL",        // Value of type int
	"browser_check":       "Browser Integrity Check",  // Value of type string
	"cache_level":         "Cache Level",              // Value of type string
	"disable_apps":        "Disable Apps",             // Value of type interface{}
	"disable_performance": "Disable Performance",      // Value of type interface{}
	"disable_railgun":     "Disable Railgun",          // Value of type string
	"disable_security":    "Disable Security",         // Value of type interface{}
	"edge_cache_ttl":      "Edge Cache TTL",           // Value of type int
	"email_obfuscation":   "Email Obfuscation",        // Value of type string
	"forwarding_url":      "Forwarding URL",           // Value of type map[string]interface
	"ip_geolocation":      "IP Geolocation Header",    // Value of type string
	"mirage":              "Mirage",                   // Value of type string
	"rocket_loader":       "Rocker Loader",            // Value of type string
	"security_level":      "Security Level",           // Value of type string
	"server_side_exclude": "Server Side Excludes",     // Value of type string
	"smart_errors":        "Smart Errors",             // Value of type string
	"ssl":                 "SSL",                      // Value of type string
	"waf":                 "Web Application Firewall", // Value of type string
}

// PageRule describes a Page Rule.
type PageRule struct {
	ID         string           `json:"id,omitempty"`
	Targets    []PageRuleTarget `json:"targets"`
	Actions    []PageRuleAction `json:"actions"`
	Priority   int              `json:"priority"`
	Status     string           `json:"status"` // can be: active, paused
	ModifiedOn time.Time        `json:"modified_on,omitempty"`
	CreatedOn  time.Time        `json:"created_on,omitempty"`
}

// PageRuleDetailResponse is the API response, containing a single PageRule.
type PageRuleDetailResponse struct {
	Success  bool     `json:"success"`
	Errors   []string `json:"errors"`
	Messages []string `json:"messages"`
	Result   PageRule `json:"result"`
}

// PageRulesResponse is the API response, containing an array of PageRules.
type PageRulesResponse struct {
	Success  bool       `json:"success"`
	Errors   []string   `json:"errors"`
	Messages []string   `json:"messages"`
	Result   []PageRule `json:"result"`
}

// CreatePageRule creates a new Page Rule for a zone.
//
// API reference: https://api.cloudflare.com/#page-rules-for-a-zone-create-a-page-rule
func (api *API) CreatePageRule(zoneID string, rule PageRule) (*PageRule, error) {
	uri := "/zones/" + zoneID + "/pagerules"
	res, err := api.makeRequest("POST", uri, rule)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}
	var r PageRuleDetailResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}
	return &r.Result, nil
}

// ListPageRules returns all Page Rules for a zone.
//
// API reference: https://api.cloudflare.com/#page-rules-for-a-zone-list-page-rules
func (api *API) ListPageRules(zoneID string) ([]PageRule, error) {
	uri := "/zones/" + zoneID + "/pagerules"
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return []PageRule{}, errors.Wrap(err, errMakeRequestError)
	}
	var r PageRulesResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return []PageRule{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// PageRule fetches detail about one Page Rule for a zone.
//
// API reference: https://api.cloudflare.com/#page-rules-for-a-zone-page-rule-details
func (api *API) PageRule(zoneID, ruleID string) (PageRule, error) {
	uri := "/zones/" + zoneID + "/pagerules/" + ruleID
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return PageRule{}, errors.Wrap(err, errMakeRequestError)
	}
	var r PageRuleDetailResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return PageRule{}, errors.Wrap(err, errUnmarshalError)
	}
	return r.Result, nil
}

// ChangePageRule lets you change individual settings for a Page Rule. This is
// in contrast to UpdatePageRule which replaces the entire Page Rule.
//
// API reference: https://api.cloudflare.com/#page-rules-for-a-zone-change-a-page-rule
func (api *API) ChangePageRule(zoneID, ruleID string, rule PageRule) error {
	uri := "/zones/" + zoneID + "/pagerules/" + ruleID
	res, err := api.makeRequest("PATCH", uri, rule)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}
	var r PageRuleDetailResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return errors.Wrap(err, errUnmarshalError)
	}
	return nil
}

// UpdatePageRule lets you replace a Page Rule. This is in contrast to
// ChangePageRule which lets you change individual settings.
//
// API reference: https://api.cloudflare.com/#page-rules-for-a-zone-update-a-page-rule
func (api *API) UpdatePageRule(zoneID, ruleID string, rule PageRule) error {
	uri := "/zones/" + zoneID + "/pagerules/" + ruleID
	res, err := api.makeRequest("PUT", uri, nil)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}
	var r PageRuleDetailResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return errors.Wrap(err, errUnmarshalError)
	}
	return nil
}

// DeletePageRule deletes a Page Rule for a zone.
//
// API reference: https://api.cloudflare.com/#page-rules-for-a-zone-delete-a-page-rule
func (api *API) DeletePageRule(zoneID, ruleID string) error {
	uri := "/zones/" + zoneID + "/pagerules/" + ruleID
	res, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}
	var r PageRuleDetailResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return errors.Wrap(err, errUnmarshalError)
	}
	return nil
}
