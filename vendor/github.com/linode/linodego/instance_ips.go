package linodego

import (
	"context"
	"encoding/json"
	"fmt"
)

// InstanceIPAddressResponse contains the IPv4 and IPv6 details for an Instance
type InstanceIPAddressResponse struct {
	IPv4 *InstanceIPv4Response `json:"ipv4"`
	IPv6 *InstanceIPv6Response `json:"ipv6"`
}

// InstanceIPv4Response contains the details of all IPv4 addresses associated with an Instance
type InstanceIPv4Response struct {
	Public  []*InstanceIP `json:"public"`
	Private []*InstanceIP `json:"private"`
	Shared  []*InstanceIP `json:"shared"`
}

// InstanceIP represents an Instance IP with additional DNS and networking details
type InstanceIP struct {
	Address    string `json:"address"`
	Gateway    string `json:"gateway"`
	SubnetMask string `json:"subnet_mask"`
	Prefix     int    `json:"prefix"`
	Type       string `json:"type"`
	Public     bool   `json:"public"`
	RDNS       string `json:"rdns"`
	LinodeID   int    `json:"linode_id"`
	Region     string `json:"region"`
}

// InstanceIPv6Response contains the IPv6 addresses and ranges for an Instance
type InstanceIPv6Response struct {
	LinkLocal *InstanceIP  `json:"link_local"`
	SLAAC     *InstanceIP  `json:"slaac"`
	Global    []*IPv6Range `json:"global"`
}

// IPv6Range represents a range of IPv6 addresses routed to a single Linode in a given Region
type IPv6Range struct {
	Range  string `json:"range"`
	Region string `json:"region"`
}

// GetInstanceIPAddresses gets the IPAddresses for a Linode instance
func (c *Client) GetInstanceIPAddresses(ctx context.Context, linodeID int) (*InstanceIPAddressResponse, error) {
	e, err := c.InstanceIPs.endpointWithID(linodeID)
	if err != nil {
		return nil, err
	}
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&InstanceIPAddressResponse{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*InstanceIPAddressResponse), nil
}

// GetInstanceIPAddress gets the IPAddress for a Linode instance matching a supplied IP address
func (c *Client) GetInstanceIPAddress(ctx context.Context, linodeID int, ipaddress string) (*InstanceIP, error) {
	e, err := c.InstanceIPs.endpointWithID(linodeID)
	if err != nil {
		return nil, err
	}
	e = fmt.Sprintf("%s/%s", e, ipaddress)
	r, err := coupleAPIErrors(c.R(ctx).SetResult(&InstanceIP{}).Get(e))
	if err != nil {
		return nil, err
	}
	return r.Result().(*InstanceIP), nil
}

// AddInstanceIPAddress adds a public or private IP to a Linode instance
func (c *Client) AddInstanceIPAddress(ctx context.Context, linodeID int, public bool) (*InstanceIP, error) {
	var body string
	e, err := c.InstanceIPs.endpointWithID(linodeID)
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&InstanceIP{})

	instanceipRequest := struct {
		Type   string `json:"type"`
		Public bool   `json:"public"`
	}{"ipv4", public}

	if bodyData, err := json.Marshal(instanceipRequest); err == nil {
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

	return r.Result().(*InstanceIP), nil
}
