package linodego

import (
	"context"
	"encoding/json"
)

// AccountSettings are the account wide flags or plans that effect new resources
type AccountSettings struct {
	// The default backups enrollment status for all new Linodes for all users on the account.  When enabled, backups are mandatory per instance.
	BackupsEnabled bool `json:"backups_enabled"`

	// Wether or not Linode Managed service is enabled for the account.
	Managed bool `json:"managed"`

	// Wether or not the Network Helper is enabled for all new Linode Instance Configs on the account.
	NetworkHelper bool `json:"network_helper"`

	// A plan name like "longview-3"..."longview-100", or a nil value for to cancel any existing subscription plan.
	LongviewSubscription *string `json:"longview_subscription"`
}

// AccountSettingsUpdateOptions are the updateable account wide flags or plans that effect new resources.
type AccountSettingsUpdateOptions struct {
	// The default backups enrollment status for all new Linodes for all users on the account.  When enabled, backups are mandatory per instance.
	BackupsEnabled *bool `json:"backups_enabled,omitempty"`

	// A plan name like "longview-3"..."longview-100", or a nil value for to cancel any existing subscription plan.
	LongviewSubscription *string `json:"longview_subscription,omitempty"`

	// The default network helper setting for all new Linodes and Linode Configs for all users on the account.
	NetworkHelper *bool `json:"network_helper,omitempty"`
}

// GetAccountSettings gets the account wide flags or plans that effect new resources
func (c *Client) GetAccountSettings(ctx context.Context) (*AccountSettings, error) {
	e, err := c.AccountSettings.Endpoint()
	if err != nil {
		return nil, err
	}
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&AccountSettings{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*AccountSettings), nil
}

// UpdateAccountSettings updates the settings associated with the account
func (c *Client) UpdateAccountSettings(ctx context.Context, settings AccountSettingsUpdateOptions) (*AccountSettings, error) {
	var body string
	e, err := c.AccountSettings.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&AccountSettings{})

	if bodyData, err := json.Marshal(settings); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Put(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*AccountSettings), nil
}
