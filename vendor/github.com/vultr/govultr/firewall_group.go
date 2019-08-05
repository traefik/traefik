package govultr

import (
	"context"
	"net/http"
	"net/url"
)

// FirewallGroupService is the interface to interact with the firewall group endpoints on the Vultr API
// Link: https://www.vultr.com/api/#firewall
type FirewallGroupService interface {
	Create(ctx context.Context, description string) (*FirewallGroup, error)
	Delete(ctx context.Context, groupID string) error
	List(ctx context.Context) ([]FirewallGroup, error)
	Get(ctx context.Context, groupID string) (*FirewallGroup, error)
	ChangeDescription(ctx context.Context, groupID, description string) error
}

// FireWallGroupServiceHandler handles interaction with the firewall group methods for the Vultr API
type FireWallGroupServiceHandler struct {
	client *Client
}

// FirewallGroup represents a Vultr firewall group
type FirewallGroup struct {
	FirewallGroupID string `json:"FIREWALLGROUPID"`
	Description     string `json:"description"`
	DateCreated     string `json:"date_created"`
	DateModified    string `json:"date_modified"`
	InstanceCount   int    `json:"instance_count"`
	RuleCount       int    `json:"rule_count"`
	MaxRuleCount    int    `json:"max_rule_count"`
}

// Create will create a new firewall group on your Vultr account
func (f *FireWallGroupServiceHandler) Create(ctx context.Context, description string) (*FirewallGroup, error) {

	uri := "/v1/firewall/group_create"

	values := url.Values{
		"description": {description},
	}

	req, err := f.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	firewall := new(FirewallGroup)
	err = f.client.DoWithContext(ctx, req, firewall)

	if err != nil {
		return nil, err
	}

	return firewall, nil
}

// Delete will delete a firewall group from your Vultr account
func (f *FireWallGroupServiceHandler) Delete(ctx context.Context, groupID string) error {

	uri := "/v1/firewall/group_delete"

	values := url.Values{
		"FIREWALLGROUPID": {groupID},
	}

	req, err := f.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = f.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// List will return a list of  all firewall groups on your Vultr account
func (f *FireWallGroupServiceHandler) List(ctx context.Context) ([]FirewallGroup, error) {

	uri := "/v1/firewall/group_list"

	req, err := f.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	var firewallGroupMap map[string]FirewallGroup
	err = f.client.DoWithContext(ctx, req, &firewallGroupMap)

	if err != nil {
		return nil, err
	}

	var firewallGroup []FirewallGroup
	for _, f := range firewallGroupMap {
		firewallGroup = append(firewallGroup, f)
	}

	return firewallGroup, nil
}

// Get will return a firewall group based on provided groupID from your Vultr account
func (f *FireWallGroupServiceHandler) Get(ctx context.Context, groupID string) (*FirewallGroup, error) {

	uri := "/v1/firewall/group_list"

	req, err := f.client.NewRequest(ctx, http.MethodGet, uri, nil)

	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("FIREWALLGROUPID", groupID)
	req.URL.RawQuery = q.Encode()

	var firewallGroupMap map[string]FirewallGroup
	err = f.client.DoWithContext(ctx, req, &firewallGroupMap)

	if err != nil {
		return nil, err
	}

	firewallGroup := new(FirewallGroup)
	for _, f := range firewallGroupMap {
		firewallGroup = &f
	}

	return firewallGroup, nil
}

// ChangeDescription will change the description of a firewall group
func (f *FireWallGroupServiceHandler) ChangeDescription(ctx context.Context, groupID, description string) error {

	uri := "/v1/firewall/group_set_description"

	values := url.Values{
		"FIREWALLGROUPID": {groupID},
		"description":     {description},
	}

	req, err := f.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = f.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}
