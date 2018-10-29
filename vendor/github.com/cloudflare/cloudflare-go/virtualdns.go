package cloudflare

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// VirtualDNS represents a Virtual DNS configuration.
type VirtualDNS struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	OriginIPs            []string `json:"origin_ips"`
	VirtualDNSIPs        []string `json:"virtual_dns_ips"`
	MinimumCacheTTL      uint     `json:"minimum_cache_ttl"`
	MaximumCacheTTL      uint     `json:"maximum_cache_ttl"`
	DeprecateAnyRequests bool     `json:"deprecate_any_requests"`
	ModifiedOn           string   `json:"modified_on"`
}

// VirtualDNSResponse represents a Virtual DNS response.
type VirtualDNSResponse struct {
	Response
	Result *VirtualDNS `json:"result"`
}

// VirtualDNSListResponse represents an array of Virtual DNS responses.
type VirtualDNSListResponse struct {
	Response
	Result []*VirtualDNS `json:"result"`
}

// CreateVirtualDNS creates a new Virtual DNS cluster.
//
// API reference: https://api.cloudflare.com/#virtual-dns-users--create-a-virtual-dns-cluster
func (api *API) CreateVirtualDNS(v *VirtualDNS) (*VirtualDNS, error) {
	res, err := api.makeRequest("POST", "/user/virtual_dns", v)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &VirtualDNSResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response.Result, nil
}

// VirtualDNS fetches a single virtual DNS cluster.
//
// API reference: https://api.cloudflare.com/#virtual-dns-users--get-a-virtual-dns-cluster
func (api *API) VirtualDNS(virtualDNSID string) (*VirtualDNS, error) {
	uri := "/user/virtual_dns/" + virtualDNSID
	res, err := api.makeRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &VirtualDNSResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response.Result, nil
}

// ListVirtualDNS lists the virtual DNS clusters associated with an account.
//
// API reference: https://api.cloudflare.com/#virtual-dns-users--get-virtual-dns-clusters
func (api *API) ListVirtualDNS() ([]*VirtualDNS, error) {
	res, err := api.makeRequest("GET", "/user/virtual_dns", nil)
	if err != nil {
		return nil, errors.Wrap(err, errMakeRequestError)
	}

	response := &VirtualDNSListResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return nil, errors.Wrap(err, errUnmarshalError)
	}

	return response.Result, nil
}

// UpdateVirtualDNS updates a Virtual DNS cluster.
//
// API reference: https://api.cloudflare.com/#virtual-dns-users--modify-a-virtual-dns-cluster
func (api *API) UpdateVirtualDNS(virtualDNSID string, vv VirtualDNS) error {
	uri := "/user/virtual_dns/" + virtualDNSID
	res, err := api.makeRequest("PUT", uri, vv)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}

	response := &VirtualDNSResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return errors.Wrap(err, errUnmarshalError)
	}

	return nil
}

// DeleteVirtualDNS deletes a Virtual DNS cluster. Note that this cannot be
// undone, and will stop all traffic to that cluster.
//
// API reference: https://api.cloudflare.com/#virtual-dns-users--delete-a-virtual-dns-cluster
func (api *API) DeleteVirtualDNS(virtualDNSID string) error {
	uri := "/user/virtual_dns/" + virtualDNSID
	res, err := api.makeRequest("DELETE", uri, nil)
	if err != nil {
		return errors.Wrap(err, errMakeRequestError)
	}

	response := &VirtualDNSResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		return errors.Wrap(err, errUnmarshalError)
	}

	return nil
}
