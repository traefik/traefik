package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// NodeBalancer represents a NodeBalancer object
type NodeBalancer struct {
	CreatedStr string `json:"created"`
	UpdatedStr string `json:"updated"`
	// This NodeBalancer's unique ID.
	ID int `json:"id"`
	// This NodeBalancer's label. These must be unique on your Account.
	Label *string `json:"label"`
	// The Region where this NodeBalancer is located. NodeBalancers only support backends in the same Region.
	Region string `json:"region"`
	// This NodeBalancer's hostname, ending with .nodebalancer.linode.com
	Hostname *string `json:"hostname"`
	// This NodeBalancer's public IPv4 address.
	IPv4 *string `json:"ipv4"`
	// This NodeBalancer's public IPv6 address.
	IPv6 *string `json:"ipv6"`
	// Throttle connections per second (0-20). Set to 0 (zero) to disable throttling.
	ClientConnThrottle int `json:"client_conn_throttle"`
	// Information about the amount of transfer this NodeBalancer has had so far this month.
	Transfer NodeBalancerTransfer `json:"transfer"`

	Created *time.Time `json:"-"`
	Updated *time.Time `json:"-"`
}

// NodeBalancerTransfer contains information about the amount of transfer a NodeBalancer has had in the current month
type NodeBalancerTransfer struct {
	// The total transfer, in MB, used by this NodeBalancer this month.
	Total *float64 `json:"total"`
	// The total inbound transfer, in MB, used for this NodeBalancer this month.
	Out *float64 `json:"out"`
	// The total outbound transfer, in MB, used for this NodeBalancer this month.
	In *float64 `json:"in"`
}

// NodeBalancerCreateOptions are the options permitted for CreateNodeBalancer
type NodeBalancerCreateOptions struct {
	Label              *string                            `json:"label,omitempty"`
	Region             string                             `json:"region,omitempty"`
	ClientConnThrottle *int                               `json:"client_conn_throttle,omitempty"`
	Configs            []*NodeBalancerConfigCreateOptions `json:"configs,omitempty"`
}

// NodeBalancerUpdateOptions are the options permitted for UpdateNodeBalancer
type NodeBalancerUpdateOptions struct {
	Label              *string `json:"label,omitempty"`
	ClientConnThrottle *int    `json:"client_conn_throttle,omitempty"`
}

// GetCreateOptions converts a NodeBalancer to NodeBalancerCreateOptions for use in CreateNodeBalancer
func (i NodeBalancer) GetCreateOptions() NodeBalancerCreateOptions {
	return NodeBalancerCreateOptions{
		Label:              i.Label,
		Region:             i.Region,
		ClientConnThrottle: &i.ClientConnThrottle,
	}
}

// GetUpdateOptions converts a NodeBalancer to NodeBalancerUpdateOptions for use in UpdateNodeBalancer
func (i NodeBalancer) GetUpdateOptions() NodeBalancerUpdateOptions {
	return NodeBalancerUpdateOptions{
		Label:              i.Label,
		ClientConnThrottle: &i.ClientConnThrottle,
	}
}

// NodeBalancersPagedResponse represents a paginated NodeBalancer API response
type NodeBalancersPagedResponse struct {
	*PageOptions
	Data []NodeBalancer `json:"data"`
}

func (NodeBalancersPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.NodeBalancers.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

func (resp *NodeBalancersPagedResponse) appendData(r *NodeBalancersPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListNodeBalancers lists NodeBalancers
func (c *Client) ListNodeBalancers(ctx context.Context, opts *ListOptions) ([]NodeBalancer, error) {
	response := NodeBalancersPagedResponse{}
	err := c.listHelper(ctx, &response, opts)
	if err != nil {
		return nil, err
	}
	return response.Data, nil
}

// fixDates converts JSON timestamps to Go time.Time values
func (i *NodeBalancer) fixDates() *NodeBalancer {
	i.Created, _ = parseDates(i.CreatedStr)
	i.Updated, _ = parseDates(i.UpdatedStr)
	return i
}

// GetNodeBalancer gets the NodeBalancer with the provided ID
func (c *Client) GetNodeBalancer(ctx context.Context, id int) (*NodeBalancer, error) {
	e, err := c.NodeBalancers.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)
	r, err := coupleAPIErrors(c.R(ctx).
		SetResult(&NodeBalancer{}).
		Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*NodeBalancer).fixDates(), nil
}

// CreateNodeBalancer creates a NodeBalancer
func (c *Client) CreateNodeBalancer(ctx context.Context, nodebalancer NodeBalancerCreateOptions) (*NodeBalancer, error) {
	var body string
	e, err := c.NodeBalancers.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&NodeBalancer{})

	if bodyData, err := json.Marshal(nodebalancer); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*NodeBalancer).fixDates(), nil
}

// UpdateNodeBalancer updates the NodeBalancer with the specified id
func (c *Client) UpdateNodeBalancer(ctx context.Context, id int, updateOpts NodeBalancerUpdateOptions) (*NodeBalancer, error) {
	var body string
	e, err := c.NodeBalancers.Endpoint()
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	req := c.R(ctx).SetResult(&NodeBalancer{})

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
	return r.Result().(*NodeBalancer).fixDates(), nil
}

// DeleteNodeBalancer deletes the NodeBalancer with the specified id
func (c *Client) DeleteNodeBalancer(ctx context.Context, id int) error {
	e, err := c.NodeBalancers.Endpoint()
	if err != nil {
		return err
	}
	e = fmt.Sprintf("%s/%d", e, id)

	_, err = coupleAPIErrors(c.R(ctx).Delete(e))

	return err
}
