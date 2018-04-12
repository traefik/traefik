package namecom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
)

var _ = bytes.MinRead

// ListURLForwardings returns a pagenated list of URL forwarding entries for a domain.
func (n *NameCom) ListURLForwardings(request *ListURLForwardingsRequest) (*ListURLForwardingsResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/url/forwarding", request.DomainName)

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

	resp := &ListURLForwardingsResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetURLForwarding returns an URL forwarding entry.
func (n *NameCom) GetURLForwarding(request *GetURLForwardingRequest) (*URLForwarding, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/url/forwarding/%s", request.DomainName, request.Host)

	values := url.Values{}

	body, err := n.get(endpoint, values)
	if err != nil {
		return nil, err
	}

	resp := &URLForwarding{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CreateURLForwarding creates an URL forwarding entry. If this is the first URL forwarding entry, it may modify the A records for the domain accordingly.
func (n *NameCom) CreateURLForwarding(request *URLForwarding) (*URLForwarding, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/url/forwarding", request.DomainName)

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &URLForwarding{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// UpdateURLForwarding updates which URL the host is being forwarded to.
func (n *NameCom) UpdateURLForwarding(request *URLForwarding) (*URLForwarding, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/url/forwarding/%s", request.DomainName, request.Host)

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.put(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &URLForwarding{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DeleteURLForwarding deletes the URL forwarding entry.
func (n *NameCom) DeleteURLForwarding(request *DeleteURLForwardingRequest) (*EmptyResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/url/forwarding/%s", request.DomainName, request.Host)

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
