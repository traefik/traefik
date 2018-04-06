package namecom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
)

var _ = bytes.MinRead

// ListDomains returns all domains in the account. It omits some information that can be retrieved from GetDomain.
func (n *NameCom) ListDomains(request *ListDomainsRequest) (*ListDomainsResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains")

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

	resp := &ListDomainsResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetDomain returns details about a specific domain
func (n *NameCom) GetDomain(request *GetDomainRequest) (*Domain, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s", request.DomainName)

	values := url.Values{}

	body, err := n.get(endpoint, values)
	if err != nil {
		return nil, err
	}

	resp := &Domain{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CreateDomain purchases a new domain. Domains that are not regularly priced require the purchase_price field to be specified.
func (n *NameCom) CreateDomain(request *CreateDomainRequest) (*CreateDomainResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains")

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &CreateDomainResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// EnableAutorenew enables the domain to be automatically renewed when it gets close to expiring.
func (n *NameCom) EnableAutorenew(request *EnableAutorenewForDomainRequest) (*Domain, error) {
	endpoint := fmt.Sprintf("/v4/domains")

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &Domain{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DisableAutorenew disables automatic renewals, thus requiring the domain to be renewed manually.
func (n *NameCom) DisableAutorenew(request *DisableAutorenewForDomainRequest) (*Domain, error) {
	endpoint := fmt.Sprintf("/v4/domains")

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &Domain{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// RenewDomain will renew a domain. Purchase_price is required if the renewal is not regularly priced.
func (n *NameCom) RenewDomain(request *RenewDomainRequest) (*RenewDomainResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains")

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &RenewDomainResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetAuthCodeForDomain returns the Transfer Authorization Code for the domain.
func (n *NameCom) GetAuthCodeForDomain(request *AuthCodeRequest) (*AuthCodeResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains")

	values := url.Values{}
	if request.DomainName != "" {
		values.Set("domainName", request.DomainName)
	}

	body, err := n.get(endpoint, values)
	if err != nil {
		return nil, err
	}

	resp := &AuthCodeResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// PurchasePrivacy will add Whois Privacy protection to a domain or will an renew existing subscription.
func (n *NameCom) PurchasePrivacy(request *PrivacyRequest) (*PrivacyResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains")

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &PrivacyResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// SetNameservers will set the nameservers for the Domain.
func (n *NameCom) SetNameservers(request *SetNameserversRequest) (*Domain, error) {
	endpoint := fmt.Sprintf("/v4/domains/%s:setNameservers", request.DomainName)

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &Domain{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// SetContacts will set the contacts for the Domain.
func (n *NameCom) SetContacts(request *SetContactsRequest) (*Domain, error) {
	endpoint := fmt.Sprintf("/v4/domains")

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &Domain{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// LockDomain will lock a domain so that it cannot be transfered to another registrar.
func (n *NameCom) LockDomain(request *LockDomainRequest) (*Domain, error) {
	endpoint := fmt.Sprintf("/v4/domains")

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &Domain{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// UnlockDomain will unlock a domain so that it can be transfered to another registrar.
func (n *NameCom) UnlockDomain(request *UnlockDomainRequest) (*Domain, error) {
	endpoint := fmt.Sprintf("/v4/domains")

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &Domain{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CheckAvailability will check a list of domains to see if they are purchaseable. A Maximum of 50 domains can be specified.
func (n *NameCom) CheckAvailability(request *AvailabilityRequest) (*SearchResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains:checkAvailability")

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &SearchResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Search will perform a search for specified keywords.
func (n *NameCom) Search(request *SearchRequest) (*SearchResponse, error) {
	endpoint := fmt.Sprintf("/v4/domains:search")

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &SearchResponse{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// SearchStream will return JSON encoded SearchResults as they are recieved from the registry. The SearchResults are separated by newlines. This can allow clients to react to results before the search is fully completed.
func (n *NameCom) SearchStream(request *SearchRequest) (*SearchResult, error) {
	endpoint := fmt.Sprintf("/v4/domains:searchStream")

	post := &bytes.Buffer{}
	err := json.NewEncoder(post).Encode(request)
	if err != nil {
		return nil, err
	}

	body, err := n.post(endpoint, post)
	if err != nil {
		return nil, err
	}

	resp := &SearchResult{}

	err = json.NewDecoder(body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
