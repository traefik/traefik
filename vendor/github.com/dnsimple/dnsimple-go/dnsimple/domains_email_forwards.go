package dnsimple

import (
	"fmt"
)

// EmailForward represents an email forward in DNSimple.
type EmailForward struct {
	ID        int64  `json:"id,omitempty"`
	DomainID  int64  `json:"domain_id,omitempty"`
	From      string `json:"from,omitempty"`
	To        string `json:"to,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func emailForwardPath(accountID string, domainIdentifier string, forwardID int64) (path string) {
	path = fmt.Sprintf("%v/email_forwards", domainPath(accountID, domainIdentifier))
	if forwardID != 0 {
		path += fmt.Sprintf("/%v", forwardID)
	}
	return
}

// emailForwardResponse represents a response from an API method that returns an EmailForward struct.
type emailForwardResponse struct {
	Response
	Data *EmailForward `json:"data"`
}

// emailForwardsResponse represents a response from an API method that returns a collection of EmailForward struct.
type emailForwardsResponse struct {
	Response
	Data []EmailForward `json:"data"`
}

// ListEmailForwards lists the email forwards for a domain.
//
// See https://developer.dnsimple.com/v2/domains/email-forwards/#list
func (s *DomainsService) ListEmailForwards(accountID string, domainIdentifier string, options *ListOptions) (*emailForwardsResponse, error) {
	path := versioned(emailForwardPath(accountID, domainIdentifier, 0))
	forwardsResponse := &emailForwardsResponse{}

	path, err := addURLQueryOptions(path, options)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.get(path, forwardsResponse)
	if err != nil {
		return nil, err
	}

	forwardsResponse.HttpResponse = resp
	return forwardsResponse, nil
}

// CreateEmailForward creates a new email forward.
//
// See https://developer.dnsimple.com/v2/domains/email-forwards/#create
func (s *DomainsService) CreateEmailForward(accountID string, domainIdentifier string, forwardAttributes EmailForward) (*emailForwardResponse, error) {
	path := versioned(emailForwardPath(accountID, domainIdentifier, 0))
	forwardResponse := &emailForwardResponse{}

	resp, err := s.client.post(path, forwardAttributes, forwardResponse)
	if err != nil {
		return nil, err
	}

	forwardResponse.HttpResponse = resp
	return forwardResponse, nil
}

// GetEmailForward fetches an email forward.
//
// See https://developer.dnsimple.com/v2/domains/email-forwards/#get
func (s *DomainsService) GetEmailForward(accountID string, domainIdentifier string, forwardID int64) (*emailForwardResponse, error) {
	path := versioned(emailForwardPath(accountID, domainIdentifier, forwardID))
	forwardResponse := &emailForwardResponse{}

	resp, err := s.client.get(path, forwardResponse)
	if err != nil {
		return nil, err
	}

	forwardResponse.HttpResponse = resp
	return forwardResponse, nil
}

// DeleteEmailForward PERMANENTLY deletes an email forward from the domain.
//
// See https://developer.dnsimple.com/v2/domains/email-forwards/#delete
func (s *DomainsService) DeleteEmailForward(accountID string, domainIdentifier string, forwardID int64) (*emailForwardResponse, error) {
	path := versioned(emailForwardPath(accountID, domainIdentifier, forwardID))
	forwardResponse := &emailForwardResponse{}

	resp, err := s.client.delete(path, nil, nil)
	if err != nil {
		return nil, err
	}

	forwardResponse.HttpResponse = resp
	return forwardResponse, nil
}
