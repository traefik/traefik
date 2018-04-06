package namecom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
)

var _ = bytes.MinRead

// ListVanityNameservers lists all nameservers registered with the registry. It omits the IP addresses from the response. Those can be found from calling GetVanityNameserver.
func (n *NameCom) ListVanityNameservers(request *ListVanityNameserversRequest) (*ListVanityNameserversResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/vanity_nameservers", request.DomainName)

	values := url.Values{}
	if request.PerPage != 0 {
		values.Set("perPage", fmt.Sprintf("%d", request.PerPage))
	}
	if request.Page != 0 {
		values.Set("page", fmt.Sprintf("%d", request.Page))
	}

	body, err := n.get(endpoint, values)
	if err != nil {
		return nil, err
	}

	resp := &ListVanityNameserversResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetVanityNameserver gets the details for a vanity nameserver registered with the registry.
func (n *NameCom) GetVanityNameserver(request *GetVanityNameserverRequest) (*VanityNameserver, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/vanity_nameservers/%s", request.DomainName, request.Hostname)

	values := url.Values{}

	body, err := n.get(endpoint, values)
	if err != nil {
		return nil, err
	}

	resp := &VanityNameserver{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CreateVanityNameserver registers a nameserver with the registry.
func (n *NameCom) CreateVanityNameserver(request *VanityNameserver) (*VanityNameserver, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/vanity_nameservers", request.DomainName)

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &VanityNameserver{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// UpdateVanityNameserver allows you to update the glue record IP addresses at the registry.
func (n *NameCom) UpdateVanityNameserver(request *VanityNameserver) (*VanityNameserver, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/vanity_nameservers/%s", request.DomainName, request.Hostname)

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.put(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &VanityNameserver{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DeleteVanityNameserver unregisteres the nameserver at the registry. This might fail if the registry believes the nameserver is in use.
func (n *NameCom) DeleteVanityNameserver(request *DeleteVanityNameserverRequest) (*EmptyResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/vanity_nameservers/%s", request.DomainName, request.Hostname)

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.delete(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &EmptyResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
