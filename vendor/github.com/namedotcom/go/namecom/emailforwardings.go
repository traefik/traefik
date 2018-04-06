package namecom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
)

var _ = bytes.MinRead

// ListEmailForwardings returns a pagenated list of email forwarding entries for a domain.
func (n *NameCom) ListEmailForwardings(request *ListEmailForwardingsRequest) (*ListEmailForwardingsResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/email/forwarding", request.DomainName)

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

	resp := &ListEmailForwardingsResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetEmailForwarding returns an email forwarding entry.
func (n *NameCom) GetEmailForwarding(request *GetEmailForwardingRequest) (*EmailForwarding, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/email/forwarding/%s", request.DomainName, request.EmailBox)

	values := url.Values{}

	body, err := n.get(endpoint, values)
	if err != nil {
		return nil, err
	}

	resp := &EmailForwarding{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CreateEmailForwarding creates an email forwarding entry. If this is the first email forwarding entry, it may modify the MX records for the domain accordingly.
func (n *NameCom) CreateEmailForwarding(request *EmailForwarding) (*EmailForwarding, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/email/forwarding", request.DomainName)

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &EmailForwarding{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// UpdateEmailForwarding updates which email address the email is being forwarded to.
func (n *NameCom) UpdateEmailForwarding(request *EmailForwarding) (*EmailForwarding, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/email/forwarding/%s", request.DomainName, request.EmailBox)

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.put(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &EmailForwarding{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DeleteEmailForwarding deletes the email forwarding entry.
func (n *NameCom) DeleteEmailForwarding(request *DeleteEmailForwardingRequest) (*EmptyResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s/email/forwarding/%s", request.DomainName, request.EmailBox)

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
