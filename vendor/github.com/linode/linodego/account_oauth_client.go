package linodego

import (
	"context"
	"encoding/json"
	"fmt"
)

// OAuthClientStatus constants start with OAuthClient and include Linode API Instance Status values
type OAuthClientStatus string

// OAuthClientStatus constants reflect the current status of an OAuth Client
const (
	OAuthClientActive    OAuthClientStatus = "active"
	OAuthClientDisabled  OAuthClientStatus = "disabled"
	OAuthClientSuspended OAuthClientStatus = "suspended"
)

// OAuthClient represents a OAuthClient object
type OAuthClient struct {
	// The unique ID of this OAuth Client.
	ID string `json:"id"`

	// The location a successful log in from https://login.linode.com should be redirected to for this client. The receiver of this redirect should be ready to accept an OAuth exchange code and finish the OAuth exchange.
	RedirectURI string `json:"redirect_uri"`

	// The name of this application. This will be presented to users when they are asked to grant it access to their Account.
	Label string `json:"label"`

	// Current status of the OAuth Client, Enum: "active" "disabled" "suspended"
	Status OAuthClientStatus `json:"status"`

	// The OAuth Client secret, used in the OAuth exchange. This is returned as <REDACTED> except when an OAuth Client is created or its secret is reset. This is a secret, and should not be shared or disclosed publicly.
	Secret string `json:"secret"`

	// If this OAuth Client is public or private.
	Public bool `json:"public"`

	// The URL where this client's thumbnail may be viewed, or nil if this client does not have a thumbnail set.
	ThumbnailURL *string `json:"thumbnail_url"`
}

// OAuthClientCreateOptions fields are those accepted by CreateOAuthClient
type OAuthClientCreateOptions struct {
	// The location a successful log in from https://login.linode.com should be redirected to for this client. The receiver of this redirect should be ready to accept an OAuth exchange code and finish the OAuth exchange.
	RedirectURI string `json:"redirect_uri"`

	// The name of this application. This will be presented to users when they are asked to grant it access to their Account.
	Label string `json:"label"`

	// If this OAuth Client is public or private.
	Public bool `json:"public"`
}

// OAuthClientUpdateOptions fields are those accepted by UpdateOAuthClient
type OAuthClientUpdateOptions struct {
	// The location a successful log in from https://login.linode.com should be redirected to for this client. The receiver of this redirect should be ready to accept an OAuth exchange code and finish the OAuth exchange.
	RedirectURI string `json:"redirect_uri"`

	// The name of this application. This will be presented to users when they are asked to grant it access to their Account.
	Label string `json:"label"`

	// If this OAuth Client is public or private.
	Public bool `json:"public"`
}

// GetCreateOptions converts a OAuthClient to OAuthClientCreateOptions for use in CreateOAuthClient
func (i OAuthClient) GetCreateOptions() (o OAuthClientCreateOptions) {
	o.RedirectURI = i.RedirectURI
	o.Label = i.Label
	o.Public = i.Public
	return
}

// GetUpdateOptions converts a OAuthClient to OAuthClientUpdateOptions for use in UpdateOAuthClient
func (i OAuthClient) GetUpdateOptions() (o OAuthClientUpdateOptions) {
	o.RedirectURI = i.RedirectURI
	o.Label = i.Label
	o.Public = i.Public
	return
}

// OAuthClientsPagedResponse represents a paginated OAuthClient API response
type OAuthClientsPagedResponse struct {
	*PageOptions
	Data []OAuthClient `json:"data"`
}

// endpoint gets the endpoint URL for OAuthClient
func (OAuthClientsPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.OAuthClients.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends OAuthClients when processing paginated OAuthClient responses
func (resp *OAuthClientsPagedResponse) appendData(r *OAuthClientsPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListOAuthClients lists OAuthClients
func (c *Client) ListOAuthClients(ctx context.Context, opts *ListOptions) ([]OAuthClient, error) {
	response := OAuthClientsPagedResponse{}
	err := c.listHelper(ctx, &response, opts)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// GetOAuthClient gets the OAuthClient with the provided ID
func (c *Client) GetOAuthClient(ctx context.Context, id string) (*OAuthClient, error) {
	e, err := c.OAuthClients.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%s", e, id)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&OAuthClient{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*OAuthClient), nil
}

// CreateOAuthClient creates an OAuthClient
func (c *Client) CreateOAuthClient(ctx context.Context, createOpts OAuthClientCreateOptions) (*OAuthClient, error) {
	var body string
	e, err := c.OAuthClients.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&OAuthClient{})

	if bodyData, err := json.Marshal(createOpts); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*OAuthClient), nil
}

// UpdateOAuthClient updates the OAuthClient with the specified id
func (c *Client) UpdateOAuthClient(ctx context.Context, id string, updateOpts OAuthClientUpdateOptions) (*OAuthClient, error) {
	var body string
	e, err := c.OAuthClients.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%s", e, id)

	req := c.R(ctx).SetResult(&OAuthClient{})

	if bodyData, err := json.Marshal(updateOpts); err == nil {
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
	return r.Result().(*OAuthClient), nil
}

// DeleteOAuthClient deletes the OAuthClient with the specified id
func (c *Client) DeleteOAuthClient(ctx context.Context, id string) error {
	e, err := c.OAuthClients.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%s", e, id)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))
	return err
}
