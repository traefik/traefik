package linodego

import (
	"context"
	"fmt"
)

// IPv6RangesPagedResponse represents a paginated IPv6Range API response
type IPv6RangesPagedResponse struct {
	*PageOptions
	Data []IPv6Range `json:"data"`
}

// endpoint gets the endpoint URL for IPv6Range
func (IPv6RangesPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.IPv6Ranges.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

// appendData appends IPv6Ranges when processing paginated IPv6Range responses
func (resp *IPv6RangesPagedResponse) appendData(r *IPv6RangesPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListIPv6Ranges lists IPv6Ranges
func (c *Client) ListIPv6Ranges(ctx context.Context, opts *ListOptions) ([]IPv6Range, error) {
	response := IPv6RangesPagedResponse{}
	err := c.listHelper(ctx, &response, opts)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// GetIPv6Range gets the template with the provided ID
func (c *Client) GetIPv6Range(ctx context.Context, id string) (*IPv6Range, error) {
	e, err := c.IPv6Ranges.Endpoint()
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
