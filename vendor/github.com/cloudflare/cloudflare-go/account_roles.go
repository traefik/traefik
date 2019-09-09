package cloudflare

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// AccountRole defines the roles that a member can have attached.
type AccountRole struct {
	ID          string                           `json:"id"`
	Name        string                           `json:"name"`
	Description string                           `json:"description"`
	Permissions map[string]AccountRolePermission `json:"permissions"`
}

// AccountRolePermission is the shared structure for all permissions
// that can be assigned to a member.
type AccountRolePermission struct {
	Read bool `json:"read"`
	Edit bool `json:"edit"`
}

// AccountRolesListResponse represents the list response from the
// account roles.
type AccountRolesListResponse struct {
	Result []AccountRole `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// AccountRoleDetailResponse is the API response, containing a single
// account role.
type AccountRoleDetailResponse struct {
	Success  bool        `json:"success"`
	Errors   []string    `json:"errors"`
	Messages []string    `json:"messages"`
	Result   AccountRole `json:"result"`
}

// AccountRoles returns all roles of an account.
//
// API reference: https://api.cloudflare.com/#account-roles-list-roles
func (api *API) AccountRoles(accountID string) ([]AccountRole, error) {
	uri := "/accounts/" + accountID + "/roles"

	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return []AccountRole{}, errors.Wrap(err, errMakeRequestError)
	}

	var accountRolesListResponse AccountRolesListResponse
	err = json.Unmarshal(res, &accountRolesListResponse)
	if err != nil {
		return []AccountRole{}, errors.Wrap(err, errUnmarshalError)
	}

	return accountRolesListResponse.Result, nil
}

// AccountRole returns the details of a single account role.
//
// API reference: https://api.cloudflare.com/#account-roles-role-details
func (api *API) AccountRole(accountID string, roleID string) (AccountRole, error) {
	uri := fmt.Sprintf("/accounts/%s/roles/%s", accountID, roleID)

	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return AccountRole{}, errors.Wrap(err, errMakeRequestError)
	}

	var accountRole AccountRoleDetailResponse
	err = json.Unmarshal(res, &accountRole)
	if err != nil {
		return AccountRole{}, errors.Wrap(err, errUnmarshalError)
	}

	return accountRole.Result, nil
}
