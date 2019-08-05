package cloudflare

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// AccessPolicy defines a policy for allowing or disallowing access to
// one or more Access applications.
type AccessPolicy struct {
	ID         string     `json:"id,omitempty"`
	Precedence int        `json:"precedence"`
	Decision   string     `json:"decision"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
	Name       string     `json:"name"`

	// The include policy works like an OR logical operator. The user must
	// satisfy one of the rules.
	Include []interface{} `json:"include"`

	// The exclude policy works like a NOT logical operator. The user must
	// not satisfy all of the rules in exclude.
	Exclude []interface{} `json:"exclude"`

	// The require policy works like a AND logical operator. The user must
	// satisfy all of the rules in require.
	Require []interface{} `json:"require"`
}

// AccessPolicyEmail is used for managing access based on the email.
// For example, restrict access to users with the email addresses
// `test@example.com` or `someone@example.com`.
type AccessPolicyEmail struct {
	Email struct {
		Email string `json:"email"`
	} `json:"email"`
}

// AccessPolicyEmailDomain is used for managing access based on an email
// domain domain such as `example.com` instead of individual addresses.
type AccessPolicyEmailDomain struct {
	EmailDomain struct {
		Domain string `json:"domain"`
	} `json:"email_domain"`
}

// AccessPolicyIP is used for managing access based in the IP. It
// accepts individual IPs or CIDRs.
type AccessPolicyIP struct {
	IP struct {
		IP string `json:"ip"`
	} `json:"ip"`
}

// AccessPolicyEveryone is used for managing access to everyone.
type AccessPolicyEveryone struct {
	Everyone struct{} `json:"everyone"`
}

// AccessPolicyAccessGroup is used for managing access based on an
// access group.
type AccessPolicyAccessGroup struct {
	Group struct {
		ID string `json:"id"`
	} `json:"group"`
}

// AccessPolicyListResponse represents the response from the list
// access polciies endpoint.
type AccessPolicyListResponse struct {
	Result []AccessPolicy `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// AccessPolicyDetailResponse is the API response, containing a single
// access policy.
type AccessPolicyDetailResponse struct {
	Success  bool         `json:"success"`
	Errors   []string     `json:"errors"`
	Messages []string     `json:"messages"`
	Result   AccessPolicy `json:"result"`
}

// AccessPolicies returns all access policies for an access application.
//
// API reference: https://api.cloudflare.com/#access-policy-list-access-policies
func (api *API) AccessPolicies(zoneID, applicationID string, pageOpts PaginationOptions) ([]AccessPolicy, ResultInfo, error) {
	v := url.Values{}
	if pageOpts.PerPage > 0 {
		v.Set("per_page", strconv.Itoa(pageOpts.PerPage))
	}
	if pageOpts.Page > 0 {
		v.Set("page", strconv.Itoa(pageOpts.Page))
	}

	uri := fmt.Sprintf(
		"/zones/%s/access/apps/%s/policies",
		zoneID,
		applicationID,
	)

	if len(v) > 0 {
		uri = uri + "?" + v.Encode()
	}

	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return []AccessPolicy{}, ResultInfo{}, errors.Wrap(err, errMakeRequestError)
	}

	var accessPolicyListResponse AccessPolicyListResponse
	err = json.Unmarshal(res, &accessPolicyListResponse)
	if err != nil {
		return []AccessPolicy{}, ResultInfo{}, errors.Wrap(err, errUnmarshalError)
	}

	return accessPolicyListResponse.Result, accessPolicyListResponse.ResultInfo, nil
}

// AccessPolicy returns a single policy based on the policy ID.
//
// API reference: https://api.cloudflare.com/#access-policy-access-policy-details
func (api *API) AccessPolicy(zoneID, applicationID, policyID string) (AccessPolicy, error) {
	uri := fmt.Sprintf(
		"/zones/%s/access/apps/%s/policies/%s",
		zoneID,
		applicationID,
		policyID,
	)

	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return AccessPolicy{}, errors.Wrap(err, errMakeRequestError)
	}

	var accessPolicyDetailResponse AccessPolicyDetailResponse
	err = json.Unmarshal(res, &accessPolicyDetailResponse)
	if err != nil {
		return AccessPolicy{}, errors.Wrap(err, errUnmarshalError)
	}

	return accessPolicyDetailResponse.Result, nil
}

// CreateAccessPolicy creates a new access policy.
//
// API reference: https://api.cloudflare.com/#access-policy-create-access-policy
func (api *API) CreateAccessPolicy(zoneID, applicationID string, accessPolicy AccessPolicy) (AccessPolicy, error) {
	uri := fmt.Sprintf(
		"/zones/%s/access/apps/%s/policies",
		zoneID,
		applicationID,
	)

	res, err := api.makeRequest("POST", uri, accessPolicy)
	if err != nil {
		return AccessPolicy{}, errors.Wrap(err, errMakeRequestError)
	}

	var accessPolicyDetailResponse AccessPolicyDetailResponse
	err = json.Unmarshal(res, &accessPolicyDetailResponse)
	if err != nil {
		return AccessPolicy{}, errors.Wrap(err, errUnmarshalError)
	}

	return accessPolicyDetailResponse.Result, nil
}

// UpdateAccessPolicy updates an existing access policy.
//
// API reference: https://api.cloudflare.com/#access-policy-update-access-policy
func (api *API) UpdateAccessPolicy(zoneID, applicationID string, accessPolicy AccessPolicy) (AccessPolicy, error) {
	if accessPolicy.ID == "" {
		return AccessPolicy{}, errors.Errorf("access policy ID cannot be empty")
	}
	uri := fmt.Sprintf(
		"/zones/%s/access/apps/%s/policies/%s",
		zoneID,
		applicationID,
		accessPolicy.ID,
	)

	res, err := api.makeRequest("PUT", uri, accessPolicy)
	if err != nil {
		return AccessPolicy{}, errors.Wrap(err, errMakeRequestError)
	}

	var accessPolicyDetailResponse AccessPolicyDetailResponse
	err = json.Unmarshal(res, &accessPolicyDetailResponse)
	if err != nil {
		return AccessPolicy{}, errors.Wrap(err, errUnmarshalError)
	}

	return accessPolicyDetailResponse.Result, nil
}

// DeleteAccessPolicy deletes an access policy.
//
// API reference: https://api.cloudflare.com/#access-policy-update-access-policy
func (api *API) DeleteAccessPolicy(zoneID, applicationID, accessPolicyID string) error {
	uri := fmt.Sprintf(
		"/zones/%s/access/apps/%s/policies/%s",
		zoneID,
		applicationID,
		accessPolicyID,
	)

	_, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}

	return nil
}
