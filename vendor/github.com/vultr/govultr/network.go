package govultr

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

// NetworkService is the interface to interact with the network endpoints on the Vultr API
// Link: https://www.vultr.com/api/#network
type NetworkService interface {
	Create(ctx context.Context, regionID, description, cidrBlock string) (*Network, error)
	Delete(ctx context.Context, networkID string) error
	List(ctx context.Context) ([]Network, error)
}

// NetworkServiceHandler handles interaction with the network methods for the Vultr API
type NetworkServiceHandler struct {
	client *Client
}

// Network represents a Vultr private network
type Network struct {
	NetworkID    string `json:"NETWORKID"`
	RegionID     string `json:"DCID"`
	Description  string `json:"description"`
	V4Subnet     string `json:"v4_subnet"`
	V4SubnetMask int    `json:"v4_subnet_mask"`
	DateCreated  string `json:"date_created"`
}

// Create a new private network. A private network can only be used at the location for which it was created.
func (n *NetworkServiceHandler) Create(ctx context.Context, regionID, description, cidrBlock string) (*Network, error) {

	uri := "/v1/network/create"

	values := url.Values{
		"DCID": {regionID},
	}

	// Optional
	if cidrBlock != "" {
		_, ipNet, err := net.ParseCIDR(cidrBlock)
		if err != nil {
			return nil, err
		}
		if v4Subnet := ipNet.IP.To4(); v4Subnet != nil {
			values.Add("v4_subnet", v4Subnet.String())
		}
		mask, _ := ipNet.Mask.Size()
		values.Add("v4_subnet_mask", strconv.Itoa(mask))
	}

	if description != "" {
		values.Add("description", description)
	}

	req, err := n.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return nil, err
	}

	network := new(Network)
	err = n.client.DoWithContext(ctx, req, network)

	if err != nil {
		return nil, err
	}

	return network, nil
}

// Delete a private network. Before deleting, a network must be disabled from all instances. See https://www.vultr.com/api/#server_private_network_disable
func (n *NetworkServiceHandler) Delete(ctx context.Context, networkID string) error {
	uri := "/v1/network/destroy"

	values := url.Values{
		"NETWORKID": {networkID},
	}

	req, err := n.client.NewRequest(ctx, http.MethodPost, uri, values)

	if err != nil {
		return err
	}

	err = n.client.DoWithContext(ctx, req, nil)

	if err != nil {
		return err
	}

	return nil
}

// List lists all private networks on the current account
func (n *NetworkServiceHandler) List(ctx context.Context) ([]Network, error) {
	uri := "/v1/network/list"

	req, err := n.client.NewRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	var networkMap map[string]Network
	err = n.client.DoWithContext(ctx, req, &networkMap)
	if err != nil {
		return nil, err
	}

	var networks []Network
	for _, network := range networkMap {
		networks = append(networks, network)
	}

	return networks, nil
}
