package cloudflare

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

// Organization represents a multi-user organization.
type Organization struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Status      string   `json:"status,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Roles       []string `json:"roles,omitempty"`
}

// organizationResponse represents the response from the Organization endpoint.
type organizationResponse struct {
	Response
	Result     []Organization `json:"result"`
	ResultInfo `json:"result_info"`
}

// OrganizationMember has details on a member.
type OrganizationMember struct {
	ID     string             `json:"id,omitempty"`
	Name   string             `json:"name,omitempty"`
	Email  string             `json:"email,omitempty"`
	Status string             `json:"status,omitempty"`
	Roles  []OrganizationRole `json:"roles,omitempty"`
}

// OrganizationInvite has details on an invite.
type OrganizationInvite struct {
	ID                 string             `json:"id,omitempty"`
	InvitedMemberID    string             `json:"invited_member_id,omitempty"`
	InvitedMemberEmail string             `json:"invited_member_email,omitempty"`
	OrganizationID     string             `json:"organization_id,omitempty"`
	OrganizationName   string             `json:"organization_name,omitempty"`
	Roles              []OrganizationRole `json:"roles,omitempty"`
	InvitedBy          string             `json:"invited_by,omitempty"`
	InvitedOn          *time.Time         `json:"invited_on,omitempty"`
	ExpiresOn          *time.Time         `json:"expires_on,omitempty"`
	Status             string             `json:"status,omitempty"`
}

// OrganizationRole has details on a role.
type OrganizationRole struct {
	ID          string   `json:"id,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// OrganizationDetails represents details of an organization.
type OrganizationDetails struct {
	ID      string               `json:"id,omitempty"`
	Name    string               `json:"name,omitempty"`
	Members []OrganizationMember `json:"members"`
	Invites []OrganizationInvite `json:"invites"`
	Roles   []OrganizationRole   `json:"roles,omitempty"`
}

// organizationDetailsResponse represents the response from the OrganizationDetails endpoint.
type organizationDetailsResponse struct {
	Response
	Result OrganizationDetails `json:"result"`
}

// ListOrganizations lists organizations of the logged-in user.
//
// API reference: https://api.cloudflare.com/#user-s-organizations-list-organizations
func (api *API) ListOrganizations() ([]Organization, ResultInfo, error) {
	var r organizationResponse
	res, err := api.makeRequest("GET", "/user/organizations", nil)
	if err != nil {
		return []Organization{}, ResultInfo{}, errors.Wrap(err, errMakeRequestError)
	}

	err = json.Unmarshal(res, &r)
	if err != nil {
		return []Organization{}, ResultInfo{}, errors.Wrap(err, errUnmarshalError)
	}

	return r.Result, r.ResultInfo, nil
}

// OrganizationDetails returns details for the specified organization of the logged-in user.
//
// API reference: https://api.cloudflare.com/#organizations-organization-details
func (api *API) OrganizationDetails(organizationID string) (OrganizationDetails, error) {
	var r organizationDetailsResponse
	uri := "/organizations/" + organizationID
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return OrganizationDetails{}, errors.Wrap(err, errMakeRequestError)
	}

	err = json.Unmarshal(res, &r)
	if err != nil {
		return OrganizationDetails{}, errors.Wrap(err, errUnmarshalError)
	}

	return r.Result, nil
}

// organizationMembersResponse represents the response from the Organization members endpoint.
type organizationMembersResponse struct {
	Response
	Result     []OrganizationMember `json:"result"`
	ResultInfo `json:"result_info"`
}

// OrganizationMembers returns list of members for specified organization of the logged-in user.
//
// API reference: https://api.cloudflare.com/#organization-members-list-members
func (api *API) OrganizationMembers(organizationID string) ([]OrganizationMember, ResultInfo, error) {
	var r organizationMembersResponse
	uri := "/organizations/" + organizationID + "/members"
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return []OrganizationMember{}, ResultInfo{}, errors.Wrap(err, errMakeRequestError)
	}

	err = json.Unmarshal(res, &r)
	if err != nil {
		return []OrganizationMember{}, ResultInfo{}, errors.Wrap(err, errUnmarshalError)
	}

	return r.Result, r.ResultInfo, nil
}

// organizationInvitesResponse represents the response from the Organization invites endpoint.
type organizationInvitesResponse struct {
	Response
	Result     []OrganizationInvite `json:"result"`
	ResultInfo `json:"result_info"`
}

// OrganizationMembers returns list of invites for specified organization of the logged-in user.
//
// API reference: https://api.cloudflare.com/#organization-invites
func (api *API) OrganizationInvites(organizationID string) ([]OrganizationInvite, ResultInfo, error) {
	var r organizationInvitesResponse
	uri := "/organizations/" + organizationID + "/invites"
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return []OrganizationInvite{}, ResultInfo{}, errors.Wrap(err, errMakeRequestError)
	}

	err = json.Unmarshal(res, &r)
	if err != nil {
		return []OrganizationInvite{}, ResultInfo{}, errors.Wrap(err, errUnmarshalError)
	}

	return r.Result, r.ResultInfo, nil
}

// organizationRolesResponse represents the response from the Organization roles endpoint.
type organizationRolesResponse struct {
	Response
	Result     []OrganizationRole `json:"result"`
	ResultInfo `json:"result_info"`
}

// OrganizationRoles returns list of roles for specified organization of the logged-in user.
//
// API reference: https://api.cloudflare.com/#organization-roles-list-roles
func (api *API) OrganizationRoles(organizationID string) ([]OrganizationRole, ResultInfo, error) {
	var r organizationRolesResponse
	uri := "/organizations/" + organizationID + "/roles"
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return []OrganizationRole{}, ResultInfo{}, errors.Wrap(err, errMakeRequestError)
	}

	err = json.Unmarshal(res, &r)
	if err != nil {
		return []OrganizationRole{}, ResultInfo{}, errors.Wrap(err, errUnmarshalError)
	}

	return r.Result, r.ResultInfo, nil
}
