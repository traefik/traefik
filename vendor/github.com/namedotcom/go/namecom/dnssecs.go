package namecom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
)

var _ = bytes.MinRead

// ListDNSSECs lists all of the DNSSEC keys registered with the registry.
func (n *NameCom) ListDNSSECs(request *ListDNSSECsRequest) (*ListDNSSECsResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/dnssec", request.DomainName)

	values := url.Values{}

	body, err := n.get(endpoint, values)
	if err != nil {
		return nil, err
	}

	resp := &ListDNSSECsResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetDNSSEC retrieves the details for a key registered with the registry.
func (n *NameCom) GetDNSSEC(request *GetDNSSECRequest) (*DNSSEC, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/dnssec/%s", request.DomainName, request.Digest)

	values := url.Values{}

	body, err := n.get(endpoint, values)
	if err != nil {
		return nil, err
	}

	resp := &DNSSEC{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CreateDNSSEC registers a DNSSEC key with the registry.
func (n *NameCom) CreateDNSSEC(request *DNSSEC) (*DNSSEC, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/dnssec", request.DomainName)

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &DNSSEC{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DeleteDNSSEC removes a DNSSEC key from the registry.
func (n *NameCom) DeleteDNSSEC(request *DeleteDNSSECRequest) (*EmptyResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/dnssec/%s", request.DomainName, request.Digest)

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
