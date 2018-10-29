package linodego

import (
	"context"
	"fmt"
)

// IPAddressesPagedResponse represents a paginated IPAddress API response
type IPAddressesPagedResponse struct {
	*PageOptions
	Data []InstanceIP `json:"data"`
}

// endpoint gets the endpoint URL for IPAddress
func (IPAddressesPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.IPAddresses.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends IPAddresses when processing paginated InstanceIPAddress responses
func (resp *IPAddressesPagedResponse) appendData(r *IPAddressesPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListIPAddresses lists IPAddresses
func (c *Client) ListIPAddresses(ctx context.Context, opts *ListOptions) ([]InstanceIP, error) {
	response := IPAddressesPagedResponse{}
	err := c.listHelper(ctx, &response, opts)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// GetIPAddress gets the template with the provided ID
func (c *Client) GetIPAddress(ctx context.Context, id string) (*InstanceIP, error) {
	e, err := c.IPAddresses.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%s", e, id)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&InstanceIP{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*InstanceIP), nil
}
