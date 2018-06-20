package dnsimple

import (
	"fmt"
)

// DomainPush represents a domain push in DNSimple.
type DomainPush struct {
	ID         int64  `json:"id,omitempty"`
	DomainID   int64  `json:"domain_id,omitempty"`
	ContactID  int64  `json:"contact_id,omitempty"`
	AccountID  int64  `json:"account_id,omitempty"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
	AcceptedAt string `json:"accepted_at,omitempty"`
}

func domainPushPath(accountID string, pushID int64) (path string) {
	path = fmt.Sprintf("/%v/pushes", accountID)
	if pushID != 0 {
		path += fmt.Sprintf("/%v", pushID)
	}
	return
}

// domainPushResponse represents a response from an API method that returns a DomainPush struct.
type domainPushResponse struct {
	Response
	Data *DomainPush `json:"data"`
}

// domainPushesResponse represents a response from an API method that returns a collection of DomainPush struct.
type domainPushesResponse struct {
	Response
	Data []DomainPush `json:"data"`
}

// DomainPushAttributes represent a domain push payload (see initiate).
type DomainPushAttributes struct {
	NewAccountEmail string `json:"new_account_email,omitempty"`
	ContactID       int64  `json:"contact_id,omitempty"`
}

// InitiatePush initiate a new domain push.
//
// See https://developer.dnsimple.com/v2/domains/pushes/#initiate
func (s *DomainsService) InitiatePush(accountID, domainID string, pushAttributes DomainPushAttributes) (*domainPushResponse, error) {
	path := versioned(fmt.Sprintf("/%v/pushes", domainPath(accountID, domainID)))
	pushResponse := &domainPushResponse{}

	resp, err := s.client.post(path, pushAttributes, pushResponse)
	if err != nil {
		return nil, err
	}

	pushResponse.HttpResponse = resp
	return pushResponse, nil
}

// ListPushes lists the pushes for an account.
//
// See https://developer.dnsimple.com/v2/domains/pushes/#list
func (s *DomainsService) ListPushes(accountID string, options *ListOptions) (*domainPushesResponse, error) {
	path := versioned(domainPushPath(accountID, 0))
	pushesResponse := &domainPushesResponse{}

	path, err := addURLQueryOptions(path, options)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.get(path, pushesResponse)
	if err != nil {
		return nil, err
	}

	pushesResponse.HttpResponse = resp
	return pushesResponse, nil
}

// AcceptPush accept a push for a domain.
//
// See https://developer.dnsimple.com/v2/domains/pushes/#accept
func (s *DomainsService) AcceptPush(accountID string, pushID int64, pushAttributes DomainPushAttributes) (*domainPushResponse, error) {
	path := versioned(domainPushPath(accountID, pushID))
	pushResponse := &domainPushResponse{}

	resp, err := s.client.post(path, pushAttributes, nil)
	if err != nil {
		return nil, err
	}

	pushResponse.HttpResponse = resp
	return pushResponse, nil
}

// RejectPush reject a push for a domain.
//
// See https://developer.dnsimple.com/v2/domains/pushes/#reject
func (s *DomainsService) RejectPush(accountID string, pushID int64) (*domainPushResponse, error) {
	path := versioned(domainPushPath(accountID, pushID))
	pushResponse := &domainPushResponse{}

	resp, err := s.client.delete(path, nil, nil)
	if err != nil {
		return nil, err
	}

	pushResponse.HttpResponse = resp
	return pushResponse, nil
}
