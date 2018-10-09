package cloudflare

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// AccessRule represents a firewall access rule.
type AccessRule struct {
	ID            string                  `json:"id,omitempty"`
	Notes         string                  `json:"notes,omitempty"`
	AllowedModes  []string                `json:"allowed_modes,omitempty"`
	Mode          string                  `json:"mode,omitempty"`
	Configuration AccessRuleConfiguration `json:"configuration,omitempty"`
	Scope         AccessRuleScope         `json:"scope,omitempty"`
	CreatedOn     time.Time               `json:"created_on,omitempty"`
	ModifiedOn    time.Time               `json:"modified_on,omitempty"`
}

// AccessRuleConfiguration represents the configuration of a firewall
// access rule.
type AccessRuleConfiguration struct {
	Target string `json:"target,omitempty"`
	Value  string `json:"value,omitempty"`
}

// AccessRuleScope represents the scope of a firewall access rule.
type AccessRuleScope struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
}

// AccessRuleResponse represents the response from the firewall access
// rule endpoint.
type AccessRuleResponse struct {
	Result AccessRule `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// AccessRuleListResponse represents the response from the list access rules
// endpoint.
type AccessRuleListResponse struct {
	Result []AccessRule `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// ListUserAccessRules returns a slice of access rules for the logged-in user.
//
// This takes an AccessRule to allow filtering of the results returned.
//
// API reference: https://api.cloudflare.com/#user-level-firewall-access-rule-list-access-rules
func (api *API) ListUserAccessRules(accessRule AccessRule, page int) (*AccessRuleListResponse, error) {
	return api.listAccessRules("/user", accessRule, page)
}

// CreateUserAccessRule creates a firewall access rule for the logged-in user.
//
// API reference: https://api.cloudflare.com/#user-level-firewall-access-rule-create-access-rule
func (api *API) CreateUserAccessRule(accessRule AccessRule) (*AccessRuleResponse, error) {
	return api.createAccessRule("/user", accessRule)
}

// UpdateUserAccessRule updates a single access rule for the logged-in user &
// given access rule identifier.
//
// API reference: https://api.cloudflare.com/#user-level-firewall-access-rule-update-access-rule
func (api *API) UpdateUserAccessRule(accessRuleID string, accessRule AccessRule) (*AccessRuleResponse, error) {
	return api.updateAccessRule("/user", accessRuleID, accessRule)
}

// DeleteUserAccessRule deletes a single access rule for the logged-in user and
// access rule identifiers.
//
// API reference: https://api.cloudflare.com/#user-level-firewall-access-rule-update-access-rule
func (api *API) DeleteUserAccessRule(accessRuleID string) (*AccessRuleResponse, error) {
	return api.deleteAccessRule("/user", accessRuleID)
}

// ListZoneAccessRules returns a slice of access rules for the given zone
// identifier.
//
// This takes an AccessRule to allow filtering of the results returned.
//
// API reference: https://api.cloudflare.com/#firewall-access-rule-for-a-zone-list-access-rules
func (api *API) ListZoneAccessRules(zoneID string, accessRule AccessRule, page int) (*AccessRuleListResponse, error) {
	return api.listAccessRules("/zones/"+zoneID, accessRule, page)
}

// CreateZoneAccessRule creates a firewall access rule for the given zone
// identifier.
//
// API reference: https://api.cloudflare.com/#firewall-access-rule-for-a-zone-create-access-rule
func (api *API) CreateZoneAccessRule(zoneID string, accessRule AccessRule) (*AccessRuleResponse, error) {
	return api.createAccessRule("/zones/"+zoneID, accessRule)
}

// UpdateZoneAccessRule updates a single access rule for the given zone &
// access rule identifiers.
//
// API reference: https://api.cloudflare.com/#firewall-access-rule-for-a-zone-update-access-rule
func (api *API) UpdateZoneAccessRule(zoneID, accessRuleID string, accessRule AccessRule) (*AccessRuleResponse, error) {
	return api.updateAccessRule("/zones/"+zoneID, accessRuleID, accessRule)
}

// DeleteZoneAccessRule deletes a single access rule for the given zone and
// access rule identifiers.
//
// API reference: https://api.cloudflare.com/#firewall-access-rule-for-a-zone-delete-access-rule
func (api *API) DeleteZoneAccessRule(zoneID, accessRuleID string) (*AccessRuleResponse, error) {
	return api.deleteAccessRule("/zones/"+zoneID, accessRuleID)
}

// ListOrganizationAccessRules returns a slice of access rules for the given
// organization identifier.
//
// This takes an AccessRule to allow filtering of the results returned.
//
// API reference: https://api.cloudflare.com/#organization-level-firewall-access-rule-list-access-rules
func (api *API) ListOrganizationAccessRules(organizationID string, accessRule AccessRule, page int) (*AccessRuleListResponse, error) {
	return api.listAccessRules("/organizations/"+organizationID, accessRule, page)
}

// CreateOrganizationAccessRule creates a firewall access rule for the given
// organization identifier.
//
// API reference: https://api.cloudflare.com/#organization-level-firewall-access-rule-create-access-rule
func (api *API) CreateOrganizationAccessRule(organizationID string, accessRule AccessRule) (*AccessRuleResponse, error) {
	return api.createAccessRule("/organizations/"+organizationID, accessRule)
}

// UpdateOrganizationAccessRule updates a single access rule for the given
// organization & access rule identifiers.
//
// API reference: https://api.cloudflare.com/#organization-level-firewall-access-rule-update-access-rule
func (api *API) UpdateOrganizationAccessRule(organizationID, accessRuleID string, accessRule AccessRule) (*AccessRuleResponse, error) {
	return api.updateAccessRule("/organizations/"+organizationID, accessRuleID, accessRule)
}

// DeleteOrganizationAccessRule deletes a single access rule for the given
// organization and access rule identifiers.
//
// API reference: https://api.cloudflare.com/#organization-level-firewall-access-rule-delete-access-rule
func (api *API) DeleteOrganizationAccessRule(organizationID, accessRuleID string) (*AccessRuleResponse, error) {
	return api.deleteAccessRule("/organizations/"+organizationID, accessRuleID)
}

func (api *API) listAccessRules(prefix string, accessRule AccessRule, page int) (*AccessRuleListResponse, error) {
	// Construct a query string
	v := url.Values{}
	if page <= 0 {
		page = 1
	}
	v.Set("page", strconv.Itoa(page))
	// Request as many rules as possible per page - API max is 100
	v.Set("per_page", "100")
	if accessRule.Notes != "" {
		v.Set("notes", accessRule.Notes)
	}
	if accessRule.Mode != "" {
		v.Set("mode", accessRule.Mode)
	}
	if accessRule.Scope.Type != "" {
		v.Set("scope_type", accessRule.Scope.Type)
	}
	if accessRule.Configuration.Value != "" {
		v.Set("configuration_value", accessRule.Configuration.Value)
	}
	if accessRule.Configuration.Target != "" {
		v.Set("configuration_target", accessRule.Configuration.Target)
	}
	v.Set("page", strconv.Itoa(page))
	query := "?" + v.Encode()

	uri := prefix + "/firewall/access_rules/rules" + query
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &AccessRuleListResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}
	return response, nil
}

func (api *API) createAccessRule(prefix string, accessRule AccessRule) (*AccessRuleResponse, error) {
	uri := prefix + "/firewall/access_rules/rules"
	res, err := api.makeRequest("POST", uri, accessRule)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &AccessRuleResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}

func (api *API) updateAccessRule(prefix, accessRuleID string, accessRule AccessRule) (*AccessRuleResponse, error) {
	uri := prefix + "/firewall/access_rules/rules/" + accessRuleID
	res, err := api.makeRequest("PATCH", uri, accessRule)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &AccessRuleResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}
	return response, nil
}

func (api *API) deleteAccessRule(prefix, accessRuleID string) (*AccessRuleResponse, error) {
	uri := prefix + "/firewall/access_rules/rules/" + accessRuleID
	res, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &AccessRuleResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response, nil
}
