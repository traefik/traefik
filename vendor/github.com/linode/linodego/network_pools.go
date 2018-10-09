package linodego

import (
	"context"
	"fmt"
)

// IPv6PoolsPagedResponse represents a paginated IPv6Pool API response
type IPv6PoolsPagedResponse struct {
	*PageOptions
	Data []IPv6Range `json:"data"`
}

// endpoint gets the endpoint URL for IPv6Pool
func (IPv6PoolsPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.IPv6Pools.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends IPv6Pools when processing paginated IPv6Pool responses
func (resp *IPv6PoolsPagedResponse) appendData(r *IPv6PoolsPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListIPv6Pools lists IPv6Pools
func (c *Client) ListIPv6Pools(ctx context.Context, opts *ListOptions) ([]IPv6Range, error) {
	response := IPv6PoolsPagedResponse{}
	err := c.listHelper(ctx, &response, opts)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// GetIPv6Pool gets the template with the provided ID
func (c *Client) GetIPv6Pool(ctx context.Context, id string) (*IPv6Range, error) {
	e, err := c.IPv6Pools.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%s", e, id)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&IPv6Range{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*IPv6Range), nil
}
