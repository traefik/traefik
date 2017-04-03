package dnsimple

import (
	"fmt"
)

// Delegation represents a list of name servers that correspond to a domain delegation.
type Delegation []string

// delegationResponse represents a response from an API method that returns a delegation struct.
type delegationResponse struct {
	Response
	Data *Delegation `json:"data"`
}

// vanityDelegationResponse represents a response for vanity name server enable and disable operations.
type vanityDelegationResponse struct {
	Response
	Data []VanityNameServer `json:"data"`
}

// GetDomainDelegation gets the current delegated name servers for the domain.
//
// See https://developer.dnsimple.com/v2/registrar/delegation/#get
func (s *RegistrarService) GetDomainDelegation(accountID string, domainName string) (*delegationResponse, error) {
	path := versioned(fmt.Sprintf("/%v/registrar/domains/%v/delegation", accountID, domainName))
	delegationResponse := &delegationResponse{}

	resp, err := s.client.get(path, delegationResponse)
	if err != nil {
		return nil, err
	}

	delegationResponse.HttpResponse = resp
	return delegationResponse, nil
}

// ChangeDomainDelegation updates the delegated name severs for the domain.
//
// See https://developer.dnsimple.com/v2/registrar/delegation/#get
func (s *RegistrarService) ChangeDomainDelegation(accountID string, domainName string, newDelegation *Delegation) (*delegationResponse, error) {
	path := versioned(fmt.Sprintf("/%v/registrar/domains/%v/delegation", accountID, domainName))
	delegationResponse := &delegationResponse{}

	resp, err := s.client.put(path, newDelegation, delegationResponse)
	if err != nil {
		return nil, err
	}

	delegationResponse.HttpResponse = resp
	return delegationResponse, nil
}

// ChangeDomainDelegationToVanity enables vanity name servers for the given domain.
//
// See https://developer.dnsimple.com/v2/registrar/delegation/#delegateToVanity
func (s *RegistrarService) ChangeDomainDelegationToVanity(accountID string, domainName string, newDelegation *Delegation) (*vanityDelegationResponse, error) {
	path := versioned(fmt.Sprintf("/%v/registrar/domains/%v/delegation/vanity", accountID, domainName))
	delegationResponse := &vanityDelegationResponse{}

	resp, err := s.client.put(path, newDelegation, delegationResponse)
	if err != nil {
		return nil, err
	}

	delegationResponse.HttpResponse = resp
	return delegationResponse, nil
}

// ChangeDomainDelegationFromVanity disables vanity name servers for the given domain.
//
// See https://developer.dnsimple.com/v2/registrar/delegation/#dedelegateFromVanity
func (s *RegistrarService) ChangeDomainDelegationFromVanity(accountID string, domainName string) (*vanityDelegationResponse, error) {
	path := versioned(fmt.Sprintf("/%v/registrar/domains/%v/delegation/vanity", accountID, domainName))
	delegationResponse := &vanityDelegationResponse{}

	resp, err := s.client.delete(path, nil, nil)
	if err != nil {
		return nil, err
	}

	delegationResponse.HttpResponse = resp
	return delegationResponse, nil
}
